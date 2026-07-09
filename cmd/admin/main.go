package main

import (
	"log"
	"strconv"

	"dscheduler/config"
	"dscheduler/handler"
	"dscheduler/registry"
	"dscheduler/scheduler"
	"dscheduler/store"
	"github.com/gin-gonic/gin"
)

// 调度中心（Admin）入口
func main() {
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if err := store.InitMySQL(); err != nil {
		log.Fatalf("MySQL 初始化失败: %v", err)
	}
	if err := store.InitRedis(); err != nil {
		log.Fatalf("Redis 初始化失败: %v", err)
	}

	etcd := registry.NewEtcdClient()
	engine := scheduler.NewEngine(etcd)
	engine.Start()
	defer engine.Stop()

	r := gin.Default()
	h := handler.NewHandler(engine)
	h.RegisterRoutes(r)

	port := config.Global.Server.Port
	if port == 0 {
		port = 8080
	}
	log.Printf("调度中心启动，监听 :%d", port)
	if err := r.Run(":" + strconv.Itoa(port)); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
