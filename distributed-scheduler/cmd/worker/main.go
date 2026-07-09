package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"dscheduler/config"
	"dscheduler/executor"
	"dscheduler/registry"
	"github.com/gin-gonic/gin"
)

// 执行器（Executor）入口：注册任务处理器、注册到 etcd、接收调度中心下发
func main() {
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	c := config.Global

	appName := "dscheduler-demo"
	address := "127.0.0.1:9999"
	adminURL := "http://127.0.0.1:" + strconv.Itoa(c.Server.Port)

	ex := executor.NewExecutor(appName, address, adminURL)

	// 注册示例任务处理器
	ex.RegisterHandler("demoJob", func(ctx context.Context, param string) error {
		log.Printf("[demoJob] 收到参数: %s", param)
		time.Sleep(500 * time.Millisecond)
		return nil
	})
	ex.RegisterHandler("shardingJob", func(ctx context.Context, param string) error {
		log.Printf("[shardingJob] 执行分片任务, 参数: %s", param)
		return nil
	})
	ex.RegisterHandler("failJob", func(ctx context.Context, param string) error {
		return fmt.Errorf("模拟任务执行失败")
	})

	// 注册到 etcd（服务发现 + 租约心跳）
	etcd := registry.NewEtcdClient()
	stop := make(chan struct{})
	if err := etcd.RegisterExecutor(c.Admin.ExecutorDiscoverPrefix, appName, address, c.Admin.ExecutorHeartbeatTTL, stop); err != nil {
		log.Fatalf("注册到 etcd 失败: %v", err)
	}

	r := gin.Default()
	ex.RegisterRoutes(r)
	log.Printf("执行器 [%s] 启动 @ %s, admin=%s, 注册前缀=%s", appName, address, adminURL, c.Admin.ExecutorDiscoverPrefix)
	if err := r.Run(":" + strconv.Itoa(9999)); err != nil {
		log.Fatalf("执行器启动失败: %v", err)
	}
}
