package closure

import "dashboardService/baseline"

// Grouper 是可直接调用的分组规则。
type Grouper func(servers []Server) []ClusterGroup

// makeGrouper 将分组键和组间比较器闭包为一个可直接调用的规则函数。
func makeGrouper(keyOf baseline.KeyFunc, cmp baseline.CompareFunc) Grouper {
	return func(servers []Server) []ClusterGroup {
		return baseline.GroupBy(servers, keyOf, cmp)
	}
}

// 五种预定义规则直接以闭包形式暴露；调用方可像调用普通函数一样使用。
// 这些导出变量可被重新赋值，使用方应将其视为只读规则入口。
var (
	// RuleExactCombination 按完整服务组合分组，并按 mask 降序排列。
	RuleExactCombination = makeGrouper(baseline.MaskOf, baseline.CompareByMaskDesc)
	// RuleExactServiceCount 按完整组合分组，先按服务数升序、再按 mask 降序排列。
	RuleExactServiceCount = makeGrouper(baseline.MaskOf, baseline.CompareByServiceCount)
	// RuleExactPrimary 按完整组合分组，依次按主服务、服务数和 mask 排列。
	RuleExactPrimary = makeGrouper(baseline.MaskOf, baseline.CompareByPrimary)
	// RuleExactConvention 按完整组合分组，并按业务约定顺序排列。
	RuleExactConvention = makeGrouper(baseline.MaskOf, baseline.CompareByConvention)
	// RulePrimaryService 按最高优先级服务归组，并按 mask 降序排列。
	RulePrimaryService = makeGrouper(baseline.PrimaryServiceMask, baseline.CompareByMaskDesc)
)
