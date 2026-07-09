package scheduler

import (
	"errors"
	"math/rand"

	"dscheduler/common"
)

// RouteResult 路由结果：选中的执行器地址 + 分片参数
type RouteResult struct {
	Addresses     []string // 需要下发的执行器地址
	ShardingTotal int       // 分片总数（仅分片广播 > 0）
	ShardingIndex int       // 当前分片序号
}

// rrCounters 各 appName 的轮询计数器（非并发安全，调用方加锁或原子操作）
var rrCounters = map[string]int{}

// Route 根据路由策略从在线执行器列表中挑选目标
func Route(appName string, strategy string, addresses []string) (RouteResult, error) {
	if len(addresses) == 0 {
		return RouteResult{}, errors.New("暂无可用执行器地址")
	}
	switch strategy {
	case "", common.RouteStrategyFirst:
		return RouteResult{Addresses: addresses[:1]}, nil
	case common.RouteStrategyLast:
		return RouteResult{Addresses: addresses[len(addresses)-1:]}, nil
	case common.RouteStrategyRandom:
		i := rand.Intn(len(addresses))
		return RouteResult{Addresses: addresses[i : i+1]}, nil
	case common.RouteStrategyRound:
		i := rrCounters[appName] % len(addresses)
		rrCounters[appName] = i + 1
		return RouteResult{Addresses: addresses[i : i+1]}, nil
	case common.RouteStrategyShardingBroadcast:
		return RouteResult{Addresses: addresses, ShardingTotal: len(addresses)}, nil
	default:
		return RouteResult{Addresses: addresses[:1]}, nil
	}
}
