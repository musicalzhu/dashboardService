package closure

import "dashboardService/baseline"

// Grouper 是分组规则的一等函数表示。
//
// 每个 Grouper 都是"已冻结 keyOf/cmp 的完整分组函数"——调用即得到分组结果，
// 无需再经过规则查表或配置构造。
//
// 与 datadriven / options 的根本差异：
//   - datadriven：规则存在 map 中，查表后得到 RuleImpl，再传给共享流程；
//   - options：规则由 Option 列表构造出 RuleImpl，再传给共享流程；
//   - closure：规则本身就是可执行函数，调用直接得到结果。
type Grouper func(servers []Server) []ClusterGroup

// makeGrouper 是规则的构造器：把 keyOf/cmp 冻结成一个 Grouper。
//
// 调用方不关心 keyOf/cmp 何时被组合——它们在构造时就被绑定到闭包里，
// 之后不可变。这是 closure 方案的核心机制：
//
//	ExactCombination = makeGrouper(MaskOf, CmpByMask)
//	// ExactCombination 是一个闭包，内部持有 MaskOf 和 CmpByMask 的引用
//
// 每次调用 makeGrouper 都会产生一个独立的闭包——这是规则可被独立装饰的前提。
func makeGrouper(keyOf baseline.KeyFunc, cmp baseline.CompareFunc) Grouper {
	return func(servers []Server) []ClusterGroup {
		return baseline.GroupServersByCluster(servers, baseline.RuleImpl{
			KeyOf: keyOf,
			Cmp:   cmp,
		})
	}
}
