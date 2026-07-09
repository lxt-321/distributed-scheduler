# 分布式任务调度平台 API 文档

基础地址：`http://127.0.0.1:8080`
统一响应：`{ "code": 200, "msg": "success", "data": ... }`（200=成功，500=失败）

---

## 一、任务管理

### 1. 任务列表
`GET /api/job/page`

### 2. 新增任务
`POST /api/job/add`
```json
{
  "jobGroup": 1,
  "jobCron": "*/5 * * * * ?",
  "jobDesc": "演示任务",
  "author": "admin",
  "executorRouteStrategy": "FIRST",
  "executorHandler": "demoJob",
  "executorParam": "hello",
  "executorBlockStrategy": "SERIAL_EXECUTION",
  "executorTimeout": 0,
  "executorFailRetryCount": 0,
  "triggerStatus": 1
}
```

### 3. 更新任务
`POST /api/job/update`（同新增，需带 `id`）

### 4. 删除任务
`POST /api/job/remove?id=1`

### 5. 启动任务
`POST /api/job/start?id=1`

### 6. 停止任务
`POST /api/job/stop?id=1`

### 7. 手动触发
`POST /api/job/trigger?id=1`

---

## 二、执行器分组

- `GET /api/group/page` 列表
- `POST /api/group/add` 新增：`{ "appName":"dscheduler-demo", "title":"演示执行器", "addressType":0 }`
- `POST /api/group/update` 更新（带 `id`）
- `POST /api/group/remove?id=1`

---

## 三、执行器回调（内部接口）

`POST /api/callback`
```json
{
  "callBack": [
    { "logId": 12, "logDateTim": 1710000000, "handleCode": 200, "handleMsg": "success" }
  ]
}
```

---

## 四、执行日志

`GET /api/log/page`

---

## 五、执行器对外接口（调度中心调用）

| 方法 | 路径   | 说明           |
|------|--------|----------------|
| POST | /run   | 接收任务下发   |
| POST | /beat  | 心跳探测       |
| POST | /kill  | 终止任务       |
| POST | /log   | 查看任务日志   |

### 触发参数（POST /run）
```json
{
  "jobId": 1,
  "executorHandler": "demoJob",
  "executorParams": "hello",
  "executorBlockStrategy": "SERIAL_EXECUTION",
  "executorTimeout": 0,
  "logId": 12,
  "logDateTime": 1710000000,
  "shardingIndex": 0,
  "shardingTotal": 1
}
```
