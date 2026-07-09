package scheduler

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"dscheduler/common"
	"dscheduler/config"
	"dscheduler/executor"
	"dscheduler/model"
	"dscheduler/registry"
	"dscheduler/store"
)

// Engine 调度引擎：基于 cron 定时触发任务，并通过 etcd 发现执行器、按路由策略下发
type Engine struct {
	cron    *cron.Cron
	etcd    *registry.EtcdClient
	mu      sync.Mutex
	running map[int64]cron.EntryID
}

// NewEngine 创建调度引擎
func NewEngine(etcd *registry.EtcdClient) *Engine {
	return &Engine{
		cron:    cron.New(cron.WithSeconds()), // 启用 6 段 cron（秒 分 时 日 月 周）
		etcd:    etcd,
		running: make(map[int64]cron.EntryID),
	}
}

// Start 启动调度：加载所有运行中的任务并注册到 cron
func (e *Engine) Start() {
	e.cron.Start()
	jobs, err := store.LoadTriggerJobs()
	if err != nil {
		log.Printf("[scheduler] 加载任务失败: %v", err)
		return
	}
	for _, j := range jobs {
		if err := e.AddJob(j); err != nil {
			log.Printf("[scheduler] 任务 %d 注册失败: %v", j.ID, err)
		}
	}
	log.Printf("[scheduler] 调度引擎启动，已加载 %d 个运行中的任务", len(jobs))
}

// Stop 停止调度
func (e *Engine) Stop() {
	ctx := e.cron.Stop()
	<-ctx.Done()
}

// AddJob 动态添加/重启任务（启动时与新增/启动 API 调用）
func (e *Engine) AddJob(job model.JobInfo) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if id, ok := e.running[job.ID]; ok {
		e.cron.Remove(id)
		delete(e.running, job.ID)
	}
	if job.TriggerStatus != common.TriggerStatusStart {
		return nil
	}
	id, err := e.cron.AddFunc(job.JobCron, func() { e.trigger(job) })
	if err != nil {
		return err
	}
	e.running[job.ID] = id
	return nil
}

// RemoveJob 停止并移除任务
func (e *Engine) RemoveJob(jobID int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if id, ok := e.running[jobID]; ok {
		e.cron.Remove(id)
		delete(e.running, jobID)
	}
}

// TriggerJob 手动触发（API 调用），异步执行避免阻塞 HTTP
func (e *Engine) TriggerJob(jobID int64) error {
	job, err := store.GetJob(jobID)
	if err != nil {
		return err
	}
	go e.trigger(job)
	return nil
}

// trigger 核心触发流程：取锁 -> 发现执行器 -> 路由 -> 下发 -> 记日志
func (e *Engine) trigger(job model.JobInfo) {
	ctx := context.Background()

	// 1) 分布式锁：多调度中心部署时防止同一任务被重复触发
	lockKey := "dscheduler:trigger:lock:" + strconv.FormatInt(job.ID, 10)
	ok, err := store.TryLock(ctx, lockKey, 30*time.Second)
	if err != nil {
		log.Printf("[scheduler] 任务 %d 取锁失败: %v", job.ID, err)
		return
	}
	if !ok {
		return // 已被其他调度中心占用
	}
	defer store.RDB.Del(ctx, lockKey)

	// 2) 反查执行器 AppName
	appName, err := store.GetAppNameByGroup(job.JobGroup)
	if err != nil {
		recordTriggerFail(job, "分组不存在: "+err.Error())
		return
	}

	// 3) 通过 etcd 发现在线执行器
	addresses, err := e.etcd.Discover(config.Global.Admin.ExecutorDiscoverPrefix, appName)
	if err != nil || len(addresses) == 0 {
		recordTriggerFail(job, "未发现可用执行器")
		return
	}

	// 4) 路由策略
	route, err := Route(appName, job.ExecutorRouteStrategy, addresses)
	if err != nil {
		recordTriggerFail(job, err.Error())
		return
	}

	// 5) 下发到执行器（分片广播时并发下发多个分片）
	for idx, addr := range route.Addresses {
		shardTotal := 1
		shardIndex := 0
		if route.ShardingTotal > 0 {
			shardTotal = route.ShardingTotal
			shardIndex = idx
		}
		logID := store.CreateLog(job, addr)
		param := model.TriggerParam{
			JobID:                 job.ID,
			ExecutorHandler:       job.ExecutorHandler,
			ExecutorParams:        job.ExecutorParam,
			ExecutorBlockStrategy: job.ExecutorBlockStrategy,
			ExecutorTimeout:       job.ExecutorTimeout,
			LogID:                 logID,
			LogDateTime:           time.Now().Unix(),
			ShardingIndex:         shardIndex,
			ShardingTotal:         shardTotal,
		}
		go func(addr string, param model.TriggerParam) {
			res, err := executor.RunJob(addr, param)
			if err != nil {
				store.UpdateLogTrigger(logID, common.TriggerCodeFail, "下发失败: "+err.Error())
				return
			}
			if res.Code != common.CodeSuccess {
				store.UpdateLogTrigger(logID, common.TriggerCodeFail, res.Msg)
				return
			}
			store.UpdateLogTrigger(logID, common.TriggerCodeSuccess, "触发成功")
		}(addr, param)
	}
}

func recordTriggerFail(job model.JobInfo, msg string) {
	log.Printf("[scheduler] 任务 %d 触发失败: %s", job.ID, msg)
	store.UpdateLogTrigger(store.CreateLog(job, ""), common.TriggerCodeFail, msg)
}
