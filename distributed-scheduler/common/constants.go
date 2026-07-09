package common

// 路由策略
const (
	RouteStrategyFirst            = "FIRST"              // 第一个
	RouteStrategyLast             = "LAST"               // 最后一个
	RouteStrategyRound            = "ROUND"              // 轮询
	RouteStrategyRandom           = "RANDOM"             // 随机
	RouteStrategyShardingBroadcast = "SHARDING_BROADCAST" // 分片广播
	RouteStrategyFailover         = "FAILOVER"           // 故障转移
	RouteStrategyConsistentHash   = "CONSISTENT_HASH"    // 一致性哈希
)

// 阻塞策略（任务上次未结束时的处理方式）
const (
	BlockStrategySerial       = "SERIAL_EXECUTION" // 串行执行
	BlockStrategyDiscardLater = "DISCARD_LATER"    // 丢弃后续调度
	BlockStrategyCoverEarly   = "COVER_EARLY"      // 覆盖之前调度
)

// 触发/执行状态
const (
	TriggerStatusStop  = 0
	TriggerStatusStart = 1

	TriggerCodeSuccess = 200
	TriggerCodeFail    = 500
	HandleCodeSuccess  = 200
	HandleCodeFail     = 500
)
