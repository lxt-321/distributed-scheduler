package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"dscheduler/model"
)

// HandlerFunc 业务任务处理函数签名
type HandlerFunc func(ctx context.Context, param string) error

// Executor 执行器核心：维护任务处理器、接收调度中心下发并执行、回调结果
type Executor struct {
	AppName  string
	Address  string
	handlers map[string]HandlerFunc
	mu       sync.RWMutex
	adminURL string
}

// NewExecutor 创建执行器
func NewExecutor(appName, address, adminURL string) *Executor {
	return &Executor{
		AppName:  appName,
		Address:  address,
		handlers: make(map[string]HandlerFunc),
		adminURL: adminURL,
	}
}

// RegisterHandler 注册任务处理器
func (e *Executor) RegisterHandler(name string, f HandlerFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[name] = f
}

func (e *Executor) getHandler(name string) (HandlerFunc, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	f, ok := e.handlers[name]
	return f, ok
}

// Run 处理调度中心下发的执行请求
func (e *Executor) Run(c *gin.Context) {
	var p model.TriggerParam
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(200, model.TriggerResult{Code: 500, Msg: "参数错误: " + err.Error()})
		return
	}
	handler, ok := e.getHandler(p.ExecutorHandler)
	if !ok {
		e.callback(p, 500, "未找到任务处理器: "+p.ExecutorHandler)
		c.JSON(200, model.TriggerResult{Code: 500, Msg: "handler not found"})
		return
	}
	// 异步执行，立即响应"已接收"，执行完毕后回调
	go func() {
		timeout := time.Duration(p.ExecutorTimeout) * time.Second
		ctx := context.Background()
		var cancel context.CancelFunc
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		// 分片参数提示（供业务使用）
		if p.ShardingTotal > 0 {
			log.Printf("[executor] 分片任务 %s", formatShard(p.ShardingIndex, p.ShardingTotal))
		}
		if err := handler(ctx, p.ExecutorParams); err != nil {
			e.callback(p, 500, err.Error())
			return
		}
		e.callback(p, 200, "success")
	}()
	c.JSON(200, model.TriggerResult{Code: 200, Msg: "accept"})
}

// callback 执行完成后回调调度中心 /api/callback
func (e *Executor) callback(p model.TriggerParam, code int, msg string) {
	cb := model.CallbackParam{Callbacks: []model.HandleCallbackParam{{
		LogID:       p.LogID,
		LogDateTime: p.LogDateTime,
		HandleCode:  code,
		HandleMsg:   msg,
	}}}
	data, _ := json.Marshal(cb)
	req, _ := http.NewRequest(http.MethodPost, e.adminURL+"/api/callback", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[executor] 回调 admin 失败: %v", err)
		return
	}
	resp.Body.Close()
}

// Beat 心跳探测
func (e *Executor) Beat(c *gin.Context) {
	c.JSON(200, model.TriggerResult{Code: 200, Msg: "pong"})
}

// Kill 终止任务（demo：返回未实现）
func (e *Executor) Kill(c *gin.Context) {
	c.JSON(200, model.TriggerResult{Code: 200, Msg: "not supported in demo"})
}

// Log 查看任务日志（demo：返回空）
func (e *Executor) Log(c *gin.Context) {
	c.JSON(200, gin.H{"code": 200, "msg": "", "content": ""})
}

// RegisterRoutes 注册执行器 HTTP 路由
func (e *Executor) RegisterRoutes(r *gin.Engine) {
	r.POST("/run", e.Run)
	r.POST("/beat", e.Beat)
	r.POST("/kill", e.Kill)
	r.POST("/log", e.Log)
}

func formatShard(index, total int) string {
	return strconv.Itoa(index) + "/" + strconv.Itoa(total)
}
