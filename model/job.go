package model

import "time"

// JobGroup 执行器分组（AppName 维度）
type JobGroup struct {
	ID          int64     `json:"id" db:"id"`
	AppName     string    `json:"appName" db:"app_name"`
	Title       string    `json:"title" db:"title"`
	AddressType int       `json:"addressType" db:"address_type"` // 0=自动注册 1=手动录入
	AddressList string    `json:"addressList" db:"address_list"`
	UpdateTime  time.Time `json:"updateTime" db:"update_time"`
}

// JobInfo 任务配置
type JobInfo struct {
	ID                     int64     `json:"id" db:"id"`
	JobGroup               int64     `json:"jobGroup" db:"job_group"`
	JobCron                string    `json:"jobCron" db:"job_cron"`
	JobDesc                string    `json:"jobDesc" db:"job_desc"`
	Author                 string    `json:"author" db:"author"`
	ExecutorRouteStrategy  string    `json:"executorRouteStrategy" db:"executor_route_strategy"`
	ExecutorHandler        string    `json:"executorHandler" db:"executor_handler"`
	ExecutorParam          string    `json:"executorParam" db:"executor_param"`
	ExecutorBlockStrategy  string    `json:"executorBlockStrategy" db:"executor_block_strategy"`
	ExecutorTimeout        int       `json:"executorTimeout" db:"executor_timeout"`
	ExecutorFailRetryCount int       `json:"executorFailRetryCount" db:"executor_fail_retry_count"`
	TriggerStatus          int       `json:"triggerStatus" db:"trigger_status"`
	TriggerLastTime        int64     `json:"triggerLastTime" db:"trigger_last_time"`
	TriggerNextTime        int64     `json:"triggerNextTime" db:"trigger_next_time"`
	UpdateTime             time.Time `json:"updateTime" db:"update_time"`
}

// JobLog 任务执行日志
type JobLog struct {
	ID                     int64     `json:"id" db:"id"`
	JobGroup               int64     `json:"jobGroup" db:"job_group"`
	JobID                  int64     `json:"jobId" db:"job_id"`
	ExecutorAddress        string    `json:"executorAddress" db:"executor_address"`
	ExecutorHandler        string    `json:"executorHandler" db:"executor_handler"`
	ExecutorParam          string    `json:"executorParam" db:"executor_param"`
	ExecutorShardingParam  string    `json:"executorShardingParam" db:"executor_sharding_param"`
	ExecutorFailRetryCount int       `json:"executorFailRetryCount" db:"executor_fail_retry_count"`
	TriggerTime            int64     `json:"triggerTime" db:"trigger_time"`
	TriggerCode            int       `json:"triggerCode" db:"trigger_code"`
	TriggerMsg             string    `json:"triggerMsg" db:"trigger_msg"`
	HandleTime             int64     `json:"handleTime" db:"handle_time"`
	HandleCode             int       `json:"handleCode" db:"handle_code"`
	HandleMsg              string    `json:"handleMsg" db:"handle_msg"`
}
