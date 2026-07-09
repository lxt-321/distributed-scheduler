package model

// RegistryInfo 执行器注册信息（存储于 etcd）
type RegistryInfo struct {
	AppName string `json:"appName"`
	Address string `json:"address"`
}

// TriggerParam 调度中心下发给执行器的触发参数
type TriggerParam struct {
	JobID                 int64  `json:"jobId"`
	ExecutorHandler       string `json:"executorHandler"`
	ExecutorParams        string `json:"executorParams"`
	ExecutorBlockStrategy string `json:"executorBlockStrategy"`
	ExecutorTimeout       int    `json:"executorTimeout"`
	LogID                 int64  `json:"logId"`
	LogDateTime           int64  `json:"logDateTime"`
	ShardingIndex         int    `json:"shardingIndex"`
	ShardingTotal         int    `json:"shardingTotal"`
}

// TriggerResult 执行器返回给调度中心的触发结果
type TriggerResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// HandleCallbackParam 执行器回调调度中心的执行结果
type HandleCallbackParam struct {
	LogID       int64  `json:"logId"`
	LogDateTime int64  `json:"logDateTim"`
	HandleCode  int    `json:"handleCode"`
	HandleMsg   string `json:"handleMsg"`
}

// CallbackParam 回调请求体
type CallbackParam struct {
	Callbacks []HandleCallbackParam `json:"callBack"`
}
