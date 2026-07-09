# 分布式任务调度平台 · API 文档

> 📍 **基础地址**：`http://127.0.0.1:8080`
>
> 📦 **统一响应格式**：
> ```json
> { "code": 200, "msg": "success", "data": { ... } }
> ```
> - `code = 200`：成功
> - `code = 500`：服务端错误（详见 `msg` 字段）
> - `code = 502`：Bad Gateway（执行器无响应）
> - `code = 503`：Service Unavailable（无在线执行器）

---

## 一、统一响应规范

| code | 含义 | 常见原因 |
|------|------|----------|
| 200 | 成功 | 操作正常完成 |
| 500 | 服务端错误 | 数据库异常 / 代码逻辑错误 |
| 502 | Bad Gateway | 执行器不可达 / 超时 |
| 503 | 无可用执行器 | 执行器全部离线 / appName 不匹配 |

---

## 二、任务管理 API

### 2.1 任务列表
```
GET /api/job/page
```
**Query 参数**：
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| size | int | 否 | 10 | 每页条数 |

**响应示例**：
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "total": 2,
    "list": [
      {
        "id": 1,
        "jobGroup": 1,
        "jobCron": "*/5 * * * * ?",
        "jobDesc": "演示任务",
        "author": "admin",
        "executorRouteStrategy": "FIRST",
        "executorHandler": "demoJob",
        "executorParam": "hello",
        "triggerStatus": 1,
        "triggerLastTime": 1719900000,
        "triggerNextTime": 1719900005
      }
    ]
  }
}
```

---

### 2.2 新增任务
```
POST /api/job/add
Content-Type: application/json
```

**请求体**：
```json
{
  "jobGroup": 1,
  "jobCron": "*/5 * * * * ?",
  "jobDesc": "数据同步任务",
  "author": "lxt",
  "executorRouteStrategy": "ROUND",
  "executorHandler": "demoJob",
  "executorParam": "hello",
  "executorBlockStrategy": "SERIAL_EXECUTION",
  "executorTimeout": 30,
  "executorFailRetryCount": 2,
  "triggerStatus": 0
}
```

**字段说明**：
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| jobGroup | int | ✅ | 执行器分组 ID |
| jobCron | string | ✅ | Cron 表达式（六位，秒级） |
| jobDesc | string | ✅ | 任务描述（中文） |
| author | string | ✅ | 负责人 |
| executorRouteStrategy | string | ✅ | 路由策略，见下方枚举 |
| executorHandler | string | ✅ | 执行器 Handler 名称 |
| executorParam | string | 否 | 任务参数，传递给执行器 |
| executorBlockStrategy | string | 否 | 阻塞策略，默认 SERIAL_EXECUTION |
| executorTimeout | int | 否 | 超时时间（秒），0=不超时 |
| executorFailRetryCount | int | 否 | 失败重试次数，默认 0 |
| triggerStatus | int | 否 | 初始状态：0=停止，1=启动 |

**路由策略枚举值**：
```
FIRST          = 1   // 第一个
LAST           = 2   // 最后一个
ROUND          = 3   // 轮询
RANDOM         = 4   // 随机
CONSISTENT_HASH = 5  // 一致性哈希
LEAST_FREQUENTLY_USED = 6  // 最少使用
SHARDING_BROADCAST = 7     // 分片广播
```

---

### 2.3 更新任务
```
POST /api/job/update
Content-Type: application/json
```
请求体同 `2.2`，需额外携带 `id` 字段。

---

### 2.4 删除任务
```
POST /api/job/remove?id=1
```

---

### 2.5 启动任务（启用调度）
```
POST /api/job/start?id=1
```
> 启动后，调度中心按 Cron 表达式定时触发，无需重复调用。

---

### 2.6 停止任务（暂停调度）
```
POST /api/job/stop?id=1
```

---

### 2.7 手动触发（立即执行一次）
```
POST /api/job/trigger?id=1
```
> 不受 Cron 约束，立即触发一次，常用于测试。调度中心会在触发前加 Redis 分布式锁。

---

## 三、执行器分组 API

### 3.1 分组列表
```
GET /api/group/page
```

**响应示例**：
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "list": [
      {
        "id": 1,
        "appName": "dscheduler-demo",
        "title": "演示执行器",
        "addressType": 0,
        "registryList": ["127.0.0.1:9999"]
      }
    ]
  }
}
```

---

### 3.2 新增执行器组
```
POST /api/group/add
Content-Type: application/json
```

```json
{
  "appName": "dscheduler-demo",
  "title": "演示执行器",
  "addressType": 0,
  "registryList": ["127.0.0.1:9999"]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| appName | string | ✅ | 执行器唯一标识，需与 Worker 配置一致 |
| title | string | ✅ | 中文名称 |
| addressType | int | ✅ | 0=自动注册（etcd），1=手动地址 |
| registryList | []string | 否 | addressType=1 时填写固定地址 |

---

### 3.3 更新执行器组
```
POST /api/group/update
Content-Type: application/json
```
需携带 `id` 字段。

---

### 3.4 删除执行器组
```
POST /api/group/remove?id=1
```

---

## 四、执行回调（Worker → Admin）

### 4.1 执行回调
```
POST /api/callback
Content-Type: application/json
```

**请求体**：
```json
{
  "callBack": [
    {
      "logId": 12,
      "logDateTim": 1719900000,
      "handleCode": 200,
      "handleMsg": "执行成功"
    }
  ]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| logId | int | ✅ | 日志 ID（由 Admin 下发任务时传入） |
| logDateTim | int | ✅ | 开始执行时间戳（秒） |
| handleCode | int | ✅ | 执行结果码，200=成功，500=失败 |
| handleMsg | string | 否 | 执行结果信息 |

**handleCode 枚举**：
```
200  = SUCCESS         // 执行成功
500  = FAIL           // 执行失败（触发重试逻辑）
502  = TIMEOUT        // 执行超时
503  = KILL           // 被手动终止
```

---

## 五、执行日志 API

### 5.1 日志分页查询
```
GET /api/log/page
```

**Query 参数**：
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| size | int | 否 | 每页条数，默认 20 |

**响应示例**：
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "total": 5,
    "list": [
      {
        "id": 12,
        "jobId": 1,
        "jobName": "演示任务",
        "executorName": "dscheduler-demo",
        "executorAddress": "127.0.0.1:9999",
        "executorShardingParam": "",
        "triggerTime": 1719900000,
        "triggerCode": 200,
        "triggerMsg": "",
        "handleTime": 1719900002,
        "handleCode": 200,
        "handleMsg": "执行成功",
        "glueType": "BEAN",
        "executorHandler": "demoJob",
        "executorParam": "hello"
      }
    ]
  }
}
```

---

## 六、在线执行器 API

### 6.1 执行器列表
```
GET /api/executor/list
```

**响应示例**：
```json
{
  "code": 200,
  "msg": "success",
  "data": [
    {
      "appName": "dscheduler-demo",
      "address": "127.0.0.1:9999",
      "registryTime": 1719900000
    }
  ]
}
```
> 从 etcd 实时查询，返回所有在线（租约未过期）的执行器。

---

## 七、执行器对外接口（Worker 端）

> 以下接口由调度中心主动调用，路径为 Worker 监听地址（默认 `http://127.0.0.1:9999`）。

### 7.1 任务下发
```
POST /run
Content-Type: application/json
```

**请求体**：
```json
{
  "jobId": 1,
  "executorHandler": "demoJob",
  "executorParams": "hello",
  "executorBlockStrategy": "SERIAL_EXECUTION",
  "executorTimeout": 30,
  "logId": 12,
  "logDateTime": 1719900000,
  "shardingIndex": 0,
  "shardingTotal": 1
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| jobId | int | 任务 ID |
| executorHandler | string | 执行的 Handler 名称 |
| executorParams | string | 任务参数 |
| executorBlockStrategy | string | 阻塞策略 |
| executorTimeout | int | 超时秒数 |
| logId | int | 日志 ID（回调时回传） |
| logDateTime | int | 开始时间戳 |
| shardingIndex | int | 分片序号（分片广播时>0） |
| shardingTotal | int | 总分片数 |

**Worker 端需注册的 Handler 名称**（`executor/job.go` 中定义）：
```
demoJob        // 演示任务（默认实现）
```

---

### 7.2 心跳检测
```
POST /beat
Content-Type: application/json
```

**请求体**：
```json
{
  "appName": "dscheduler-demo"
}
```

**响应**：
```json
{ "code": 200, "msg": "success" }
```

---

### 7.3 任务终止
```
POST /kill
Content-Type: application/json
```

**请求体**：
```json
{
  "jobId": 1,
  "executorHandler": "demoJob"
}
```

---

### 7.4 日志查看
```
POST /log
Content-Type: application/json
```

**请求体**：
```json
{
  "logDateTime": 1719900000,
  "logId": 12
}
```
