# 分布式任务调度平台

[![Go Version](https://img.shields.io/badge/Go-1.22-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> 🔥 **简历项目** | 腾讯后台开发 · 分布式系统方向
>
> 参考 [XXL-Job](https://github.com/xuxueli/xxl-job) 设计理念，用 Go 从零实现一套轻量级分布式任务调度平台，涵盖**服务注册发现**、**调度引擎**、**执行器管理**、**任务路由**等核心模块。

---

## 🌟 项目亮点（面试重点）

| 维度 | 实现方案 | 面试可展开 |
|------|----------|------------|
| **服务发现** | etcd v3 Lease + Keepalive | 注册中心选型对比（ZooKeeper / Consul / etcd）、TTL 设计与故障感知 |
| **调度引擎** | robfig/cron 定时触发 + Redis SETNX 分布式锁 | 单点调度 → 多调度中心部署、锁粒度与性能权衡 |
| **任务路由** | FIRST / ROUND / RANDOM / SHARDING / FAILOVER 等 7 种策略 | 一致性哈希实现、分片广播的数据分片模型 |
| **高可用** | 执行器心跳保活 + 故障自动剔除 | etcd 租约续期机制、注册信息 TTL 过期策略 |
| **可扩展** | 执行器与调度中心通过 HTTP 解耦 | RESTful API 设计、无状态执行器水平扩展 |

---

## 🏗️ 系统架构

```
                                ┌─────────────────────────────────────────┐
                                │              MySQL 8.0 (3306)            │
                                │   xxl_job_info / xxl_job_log / ...     │
                                └───────────────────┬─────────────────────┘
                                                    │ 持久化
┌──────────────────┐    HTTP 触发     ┌────────────┴────────────────────┐
│  调度中心 Admin   │ ────────────────> │           etcd (2379)           │
│   :8080          │    /api/trigger  │   /dscheduler/executors/{id}    │
│                  │ <─────────────── │   带 TTL 租约，执行器心跳续期     │
│  ┌────────────┐  │   查询执行器列表  └────────────┬────────────────────┘
│  │ 任务 CRUD   │  │                            │
│  │ Cron 调度   │  │                   ┌──────────┴──────────┐
│  │ 日志查询   │  │                   │  执行器集群（可水平扩展）│
│  └────────────┘  │                   │                      │
└──────────────────┘                   │  ┌──────────────┐   │
        │                              │  │ Executor A    │   │
        │                              │  │ :9999 /run    │   │
        │    ┌──────────────┐          │  └──────────────┘   │
        └──> │ Redis (6379)  │          │                      │
             │ 分布式锁 (SETNX)│          │  ┌──────────────┐   │
             └──────────────┘          │  │ Executor B    │   │
                                       │  │ :9999 /run    │   │
                                       └──────────────────────┘
```

**数据流**：
1. 管理员通过 Admin API 创建/更新任务（含 Cron 表达式）
2. 调度中心按 Cron 表达式定时触发，加 Redis 分布式锁防重复
3. 通过 etcd 发现所有在线执行器，根据路由策略选中目标
4. HTTP POST 下发任务执行请求，执行器异步执行后回调 `/api/callback`
5. 执行日志持久化到 MySQL，状态流转：`RUNNING → SUCCESS/FAIL`

---

## 📁 目录结构

```
distributed-scheduler/
├── cmd/
│   ├── admin/main.go          # 调度中心入口（:8080）
│   └── worker/main.go         # 执行器入口（:9999）
├── common/
│   └── response.go           # 统一 HTTP 响应结构
├── config/
│   └── config.go             # YAML 配置加载（支持环境变量覆盖）
├── docs/
│   └── api.md                # API 接口文档
├── executor/
│   ├── client.go             # 执行器 HTTP 客户端（任务下发）
│   ├── handler.go            # 执行器路由处理（/run /beat /kill /log）
│   └── job.go                # 示例任务处理器
├── handler/                   # 调度中心 HTTP Handler
│   ├── job.go                # 任务 CRUD + trigger
│   ├── group.go              # 执行器组管理
│   ├── callback.go           # 执行回调
│   └── log.go                # 日志查询
├── model/
│   ├── job.go                # JobInfo / JobLog 数据结构
│   └── registry.go           # etcd 注册信息结构
├── registry/
│   └── etcd.go               # etcd 客户端（注册/发现/心跳）
├── scheduler/
│   ├── engine.go             # Cron 调度引擎
│   ├── trigger.go            # 任务触发（含路由策略）
│   └── router.go             # 路由策略实现
├── scripts/
│   └── start-all.ps1         # 一键启动脚本
├── sql/
│   └── init.sql              # 数据库建表语句
├── store/
│   ├── mysql.go              # MySQL 连接与查询
│   └── redis.go              # Redis 连接与分布式锁
├── config.example.yaml        # 配置模板（入 Git）
├── config.yaml               # 本地配置（已被 .gitignore）
├── go.mod / go.sum
└── README.md
```

---

## 🚀 快速开始

### 环境依赖

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | ≥ 1.22 | [go.dev/dl](https://go.dev/dl) |
| MySQL | ≥ 8.0 | 端口 3306，需创建 `xxl_job` 数据库 |
| Redis | ≥ 3.0 | 端口 6379，无密码 |
| etcd | ≥ 3.5 | 端口 2379，需开启 gRPC-Gateway |

### 启动步骤

```bash
# 1. 克隆项目
git clone https://github.com/lxt-321/distributed-scheduler.git
cd distributed-scheduler

# 2. 配置 Go 代理（国内网络）
go env -w GOPROXY=https://goproxy.cn,direct

# 3. 拉取依赖
go mod tidy

# 4. 初始化数据库
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS xxl_job CHARACTER SET utf8mb4;"
mysql -uroot -p xxl_job < sql/init.sql

# 5. 复制并编辑配置
cp config.example.yaml config.yaml
# 编辑 config.yaml，修改 mysql.password 为你的密码

# 6. 启动调度中心
go run ./cmd/admin

# 7. 启动执行器（新开终端）
go run ./cmd/worker
```

> 💡 **Linux/macOS 用户**：将 `go env -w` 替换为 `export`，PowerShell 命令替换为对应 shell 语法。

### 验证

| 验证方式 | 地址 |
|----------|------|
| 任务列表页 | http://127.0.0.1:8080/api/job/page |
| 执行日志页 | http://127.0.0.1:8080/api/log/page |
| 手动触发任务 | `POST http://127.0.0.1:8080/api/job/trigger?id=1` |
| 执行器心跳 | `POST http://127.0.0.1:9999/beat` |

启动成功后，Admin 日志每 5 秒输出：
```
[调度中心] 任务 1 触发成功 → 执行器 dscheduler-demo@127.0.0.1:9999
```

---

## 📋 API 概览

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/job/page` | 分页查询任务列表 |
| POST | `/api/job/add` | 新增任务 |
| POST | `/api/job/update` | 更新任务 |
| POST | `/api/job/remove` | 删除任务 |
| POST | `/api/job/start` | 启动任务（启用调度） |
| POST | `/api/job/stop` | 停止任务（暂停调度） |
| POST | `/api/job/trigger` | 手动触发任务 |
| GET | `/api/group/page` | 分页查询执行器组 |
| POST | `/api/group/add` | 新增执行器组 |
| POST | `/api/group/update` | 更新执行器组 |
| POST | `/api/group/remove` | 删除执行器组 |
| POST | `/api/callback` | 执行器执行回调 |
| GET | `/api/log/page` | 分页查询执行日志 |
| GET | `/api/executor/list` | 查询在线执行器列表 |

详细接口规范见 [docs/api.md](docs/api.md)。

---

## 🔧 路由策略

| 策略 | 代码 | 说明 |
|------|------|------|
| FIRST | `1` | 只选择第一个在线执行器 |
| LAST | `2` | 只选择最后一个在线执行器 |
| ROUND | `3` | 轮询选择 |
| RANDOM | `4` | 随机选择 |
| CONSISTENT_HASH | `5` | 一致性哈希（基于任务 ID） |
| LEAST_FREQUENTLY_USED | `6` | 最少使用（记录每个执行器的触发次数） |
| SHARDING_BROADCAST | `7` | **分片广播**：向所有在线执行器下发，携带分片参数 |

**分片广播参数**：

```json
{
  "shardingIndex": 0,   // 当前执行器序号（从 0 开始）
  "shardingTotal": 3    // 总分片数
}
```

业务代码示例（`executor/job.go`）：
```go
// 分片广播时，各执行器根据 shardingIndex 分片处理
if logParam.ShardingIndex == 0 {
    // 处理第 1 个分片的数据
} else if logParam.ShardingIndex == 1 {
    // 处理第 2 个分片的数据
}
```

---

## ⚠️ 已知限制

- 本项目为**开发环境验证版**，面向简历与面试展示
- 生产部署需补充：调度中心高可用（多实例 + 负载均衡）、执行器健康检查深度优化、Cron 表达式支持秒级精度（`* */n` 等）
- etcd 为单节点，开发测试用途

---

## 📚 技术选型参考

| 问题 | 我的选择 | 其他常见方案 | 选型理由 |
|------|----------|-------------|----------|
| 注册中心 | **etcd v3** | ZooKeeper / Consul / Nacos | gRPC 原生支持，Lease 机制天然适合心跳 + 故障剔除，学习曲线适中 |
| 分布式锁 | **Redis SETNX** | etcd TXN / MySQL 悲观锁 | Redis 轻量，大多数团队已有基础设施 |
| 任务存储 | **MySQL** | PostgreSQL / MongoDB | XXL-Job 生态成熟，SQL 统计查询方便 |
| HTTP 框架 | **Gin** | Chi / Echo / 标准库 | 高性能、中间件生态好、面试高频 |
| 定时调度 | **robfig/cron** | gridenter/Cron / 标准库 timer | 表达式规范、社区活跃 |

---

## 📝 面试高频问题 Q&A

**Q：如何保证任务不重复执行？**
> 调度中心在触发任务前，先用 `SETNX` 在 Redis 中抢锁（key = `joblock:{jobId}`），抢到才触发，超时自动释放。多调度中心部署时天然互斥。

**Q：执行器挂了怎么办？**
> etcd 注册信息带 TTL（默认 30s），执行器心跳每 10s 续一次。执行器宕机后 TTL 过期自动从列表中剔除，调度中心下次触发时不会分发到该执行器。

**Q：任务执行超时怎么办？**
> 执行器回调时携带 `handleTimeout`（单位秒），Admin 端可配置超时阈值。超时后任务标记为失败，可配置重试次数。

**Q：分片广播怎么用？**
> 适用于"将 N 万条数据分配到 M 台机器并行处理"场景。设置路由策略为 `SHARDING_BROADCAST`，各执行器收到的 `shardingIndex/shardingTotal` 不同，据此分片取数据。

**Q：为什么用 etcd 而不是 ZooKeeper？**
> etcd 使用 Raft 协议，有成熟 CLI 和 HTTP/gRPC 双接口，Lease 机制对服务发现很友好。ZK 需要编写 Watch 回调，运维成本高。
