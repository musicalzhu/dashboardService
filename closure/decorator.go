package closure

import "time"

// ============================================================================
// 装饰器：closure 方案相对其他三个方案的独特能力
// ============================================================================
//
// 为什么只有 closure 能优雅支持：
//   - baseline/datadriven 要给某个规则加缓存，必须在 RenderDashboard 里加 if 判断，
//     污染通用分发逻辑；
//   - options 的 Option 是"配置"，无法承载"包装一个可执行函数"的语义；
//   - closure 的规则本身就是函数，可以被任意装饰器包装，且互不影响。
//
// 实现说明：
//   - Decorator = func(Grouper) Grouper，因为 Decorator 要"接收一个 Grouper，返回包装后的 Grouper"。
//   - 每个 WithXxx 只有一层匿名函数（构造 Decorator），
//     内层"包装 Grouper"的细节被提取为命名函数（xxxGrouper），
//     避免两层匿名函数嵌套带来的阅读困难。
//   - 所有 xxxGrouper 都是纯函数：无隐藏状态，只依赖入参。

// Decorator 是装饰器的通用类型：接收一个 Grouper，返回包装后的 Grouper。
//
// 所有 WithXxx 构造函数都返回 Decorator，便于用 ApplyDecorators 组合。
type Decorator func(Grouper) Grouper

// ============================================================================
// WithLogging：日志装饰器
// ============================================================================

// WithLogging 返回"日志装饰器"：在规则调用前后打日志。
//
// logger 签名与 log.Printf 一致，调用方可注入自己的日志实现。
//
// 类型链：
//
//	WithLogging(logger) → Decorator
//	Decorator(g)        → Grouper
//	Grouper(servers)    → []ClusterGroup
//
// 本函数只负责第一环；后两环由 loggedGrouper 完成。
func WithLogging(logger func(format string, args ...any)) Decorator {
	return func(g Grouper) Grouper {
		return loggedGrouper(g, logger)
	}
}

// loggedGrouper 是 WithLogging 的内层实现：包装 g，在调用前后各打一条日志。
//
// 单独命名它的原因：
//   - 避免 WithLogging 里出现两层匿名函数嵌套，读者无需在脑中展开；
//   - 便于对"日志时机与内容"单独写单元测试；
//   - g 和 logger 都是显式参数，没有隐藏状态，是纯函数。
func loggedGrouper(g Grouper, logger func(format string, args ...any)) Grouper {
	return func(servers []Server) []ClusterGroup {
		logger("grouping start: servers=%d", len(servers))
		result := g(servers)
		totalGroups := 0
		for _, c := range result {
			totalGroups += len(c.Groups)
		}
		logger("grouping done: clusters=%d, groups=%d", len(result), totalGroups)
		return result
	}
}

// ============================================================================
// WithMetrics：指标装饰器
// ============================================================================

// WithMetrics 返回"指标装饰器"：记录每次规则调用的耗时。
//
// 时钟由调用方注入（record 参数），便于测试时替换。
func WithMetrics(record func(duration time.Duration)) Decorator {
	return func(g Grouper) Grouper {
		return timedGrouper(g, record)
	}
}

// timedGrouper 是 WithMetrics 的内层实现：包装 g，记录其执行耗时。
//
// 注意：start 是 timedGrouper 返回的 Grouper 的局部变量，
// 每次调用 Grouper 都是独立的 start，并发安全。
// 不能把 start 提到 timedGrouper 外部，否则会变成共享状态。
func timedGrouper(g Grouper, record func(time.Duration)) Grouper {
	return func(servers []Server) []ClusterGroup {
		start := time.Now()
		result := g(servers)
		record(time.Since(start))
		return result
	}
}

// ============================================================================
// WithRecover：容错装饰器
// ============================================================================

// WithRecover 返回"容错装饰器"：捕获 panic 并转为空结果。
//
// 生产环境中"某个规则的分组出错不能拖垮整个 Dashboard"是合理需求。
func WithRecover(onPanic func(panicValue any)) Decorator {
	return func(g Grouper) Grouper {
		return recoveredGrouper(g, onPanic)
	}
}

// recoveredGrouper 是 WithRecover 的内层实现：用 defer recover 保护 g。
//
// 命名返回值 result 是必需的：
//
//	panic 触发后，函数通过 recover 恢复正常返回流程，
//	但此时原始的 return 语句已被打断；必须由 defer 显式赋值给命名返回值，
//	否则函数会返回零值 nil，而 JSON 序列化 nil 为 null，不符合前端期望。
func recoveredGrouper(g Grouper, onPanic func(any)) Grouper {
	return func(servers []Server) (result []ClusterGroup) {
		defer func() {
			if r := recover(); r != nil {
				if onPanic != nil {
					onPanic(r)
				}
				result = []ClusterGroup{} // 容错时返回非 nil 空切片
			}
		}()
		return g(servers)
	}
}

// ============================================================================
// ApplyDecorators：装饰器链组合
// ============================================================================

// ApplyDecorators 按顺序应用多个装饰器。
//
// 顺序语义：后传入的装饰器在外层。
// 例如：
//
//	ApplyDecorators(g, WithMetrics(...), WithLogging(...))
//	// 等价于 WithLogging(WithMetrics(g))
//	// 日志在最外层：记录"开始→指标记录→结束"
//
// 实现为什么用 for 循环而非递归：
//   - 装饰器数量小（通常 ≤ 5 个），循环比递归更直白；
//   - 没有尾递归优化需求，递归也没有可读性优势。
func ApplyDecorators(g Grouper, decorators ...Decorator) Grouper {
	for _, d := range decorators {
		g = d(g)
	}
	return g
}
