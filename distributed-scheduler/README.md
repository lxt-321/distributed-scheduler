# 分布式任务调度平台（XXL-Job 风格）

一个基于 Go 的轻量级分布式任务调度平台，参考 XXL-Job 的「调度中心 + 执行器」架构，
用于简历项目展示。支持 **Cron 调度、执行器服务发现（etcd）、多种路由策略、分片广播、
失败重试、分布式锁防重复触发** 等核心能力。

## 架构

```
┌──────────────────────────────┐        ┌──────────────────────────────┐
│      调度中心 (Admin :8080)    │        │   执行器集群 (Executor :9999)  │
│  ┌──────────┐ ┌────────────┐ │  HTTP  │  ┌──────────┐ ┌────────────┐ │
│  │ 任务管理  │ │  调度引擎   │ │ -----> │  │ /run 接收 │ │ 任务处理器  │ │
│  │  CRUD API │ │  Cron 触发  │ │ 触发   │  │  下发执行 │ │ demoJob等  │ │
│  └──────────┘ └─────┬──────┘ │        │  └────┬─────┘ └────────────┘ │
│                     │        │        │       │ 回调 /api/callback    │
│              ┌──────┴──────┐ │ <----- │ ──────┘                      │
│              │ 注册中心发现 │ │        │  注册(etcd租约心跳)           │
│              └─────────────┘ │        │                              │
└──────────────────────────────┘        └──────────────────────────────┘
         │            │            │
      MySQL(任务/日志) Redis(分布式锁) etcd(执行器注册发现)
```

## 技术栈

| 组件   | 技术                      | 用途                         |
|--------|---------------------------|------------------------------|
| 语言   | Go 1.22                   | 全栈实现                     |
| Web    | Gin                       | HTTP API / 执行器接收端      |
| 调度   | robfig/cron/v3            | Cron 表达式解析与定时触发    |
| MySQL  | go-sql-driver/mysql       | 任务配置、执行日志持久化     |
| Redis  | go-redis/v9               | 分布式锁（防重复触发）       |
| etcd   | v3 gRPC-Gateway(HTTP)     | 执行器注册与自动服务发现     |

## 目录结构

```
distributed-scheduler/
├── cmd/
│   ├── admin/   # 调度中心入口
│   └── worker/  # 执行器入口（示例）
├── common/      # 常量、统一响应
├── config/      # 配置加载
├── model/       # 数据模型与 DTO
├── store/       # MySQL / Redis 访问层
├── registry/    # etcd 注册中心客户端
├── scheduler/   # 调度引擎 + 路由策略
├── executor/    # 执行器 SDK + 下发客户端
├── handler/     # 调度中心 HTTP API
├── sql/init.sql # 建表 + 种子数据
├── config.yaml  # 配置文件
└── docs/api.md  # API 文档
```

## 快速开始

### 0. 前置：中间件已在运行
- MySQL 8.0（端口 3306，已设 root 密码）
- Redis 3.0（端口 6379）
- etcd 3.7（端口 2379，需开启 gRPC-Gateway，默认开启）
- Go 1.22.4

### 1. 配置 GOPROXY（国内网络）
```powershell
go env -w GOPROXY=https://goproxy.cn,direct
```

### 2. 初始化数据库
```powershell
mysql -uroot -p < sql/init.sql
```

### 3. 准备配置（密钥不进仓库）
```powershell
copy config.example.yaml config.yaml
```
编辑 `config.yaml` 把 `mysql.password` 改为你的 root 密码；或用环境变量注入（推荐，避免密码入库）：
```powershell
$env:MYSQL_PASSWORD="你的密码"
go run ./cmd/admin
go run ./cmd/worker
```
> 注意：`config.yaml` 已被 `.gitignore` 忽略，本地密钥不会上传 GitHub；仓库中以 `config.example.yaml` 为模板。

### 4. 拉取依赖
```powershell
go mod tidy
```

### 5. 启动调度中心
```powershell
go run ./cmd/admin
```

### 6. 启动执行器（新开终端，仍在项目根目录）
```powershell
go run ./cmd/worker
```

### 7. 验证
- 调度中心日志每 5 秒触发一次 `demoJob`（种子任务）
- 执行器终端打印 `[demoJob] 收到参数: hello`
- 浏览器访问 `http://127.0.0.1:8080/api/log/page` 查看执行日志
- 访问 `http://127.0.0.1:8080/api/job/page` 查看任务
- 手动触发：`POST http://127.0.0.1:8080/api/job/trigger?id=1`

## 核心特性说明

- **服务发现**：执行器启动时向 etcd 写入带 TTL 租约的 key，并定期续租；调度中心通过 etcd 前缀查询实时发现在线执行器，进程退出自动下线。
- **路由策略**：FIRST / LAST / ROUND（轮询）/ RANDOM / SHARDING_BROADCAST（分片广播）/ FAILOVER / CONSISTENT_HASH。
- **分片广播**：一次触发向所有执行器下发，并携带 `shardingIndex / shardingTotal`，业务侧据此做数据分片处理。
- **分布式锁**：基于 Redis `SET NX PX`，多调度中心部署时防止同一任务被重复触发。
- **失败重试**：任务配置 `executorFailRetryCount`，执行器回调 `handleCode=500` 时由调度中心重试。
- **执行回调**：执行器异步执行完毕后回调 `/api/callback` 更新执行日志。

## 简历亮点（面试可讲）

1. 调度与执行分离，执行器可水平扩展。
2. etcd 实现执行器自动注册与故障自动剔除（租约 + 心跳）。
3. Redis 分布式锁保证调度幂等。
4. 分片广播解决海量数据任务的水平拆分。
5. 调度引擎基于 cron 表达式，支持秒级精度。
