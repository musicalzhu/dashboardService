package decorator

import (
	"fmt"
	"slices"

	"dashboardService/baseline"
)

// 五个预定义规则：每个都是构造时冻结了 keyOf/cmp 的 Grouper。
//
// 与 datadriven.ruleImpls、options 预定义选项组、builder 预定义构造器一一对等：
//
//	datadriven: RuleImpl{KeyOf: MaskOf, Cmp: CmpByMask}
//	options:    []Option{WithKeyOf(MaskOf), WithCmp(CmpByMask)}
//	builder:    New().WithKeyOf(MaskOf).WithCmp(CmpByMask)
//	decorator:  makeGrouper(MaskOf, CmpByMask)
//
// 四者最终传给 GroupServersByCluster 的配置相同，输出完全一致。
var (
	// ExactCombination 精确组合 + mask 降序。
	ExactCombination = makeGrouper(baseline.MaskOf, baseline.CmpByMask)

	// ExactServiceCount 精确组合 + 服务数量升序。
	ExactServiceCount = makeGrouper(baseline.MaskOf, baseline.CmpByServiceCount)

	// ExactPrimary 精确组合 + 主服务优先。
	ExactPrimary = makeGrouper(baseline.MaskOf, baseline.CmpByPrimary)

	// ExactConvention 精确组合 + 业务约定顺序。
	ExactConvention = makeGrouper(baseline.MaskOf, baseline.CmpByConvention)

	// PrimaryService 主服务归组 + mask 降序。
	PrimaryService = makeGrouper(baseline.PrimaryServiceMask, baseline.CmpByMask)
)

// ruleGroupers 从 GroupRule 枚举到 Grouper 的映射。
//
// 提供它是为了让"按 GroupRule 选规则"的调用形式与 baseline/datadriven 一致。
var ruleGroupers = map[GroupRule]Grouper{
	RuleExactCombination:  ExactCombination,
	RuleExactServiceCount: ExactServiceCount,
	RuleExactPrimary:      ExactPrimary,
	RuleExactConvention:   ExactConvention,
	RulePrimaryService:    PrimaryService,
}

// allGroupers 规则名到 Grouper 的映射。
//
// 为什么需要它：Grouper 是函数值，无法从字符串直接调用，也无法枚举。
// API 需要"字符串 → 规则"和"列出所有规则"时，必须维护这个 map。
//
// 这个 map 的存在，说明 decorator 方案要支持完整 API，
// 就必须有类似 datadriven 的规则表——这是函数值"不可枚举"弱点的补偿。
// POC 评审时应诚实指出：decorator 不是"零表的方案"，
// 只是"表不是核心，函数和装饰器才是核心"。
var allGroupers = map[string]Grouper{
	"rule-combination":     ExactCombination,
	"rule-service-count":   ExactServiceCount,
	"rule-primary":         ExactPrimary,
	"rule-convention":      ExactConvention,
	"rule-primary-service": PrimaryService,
}

// ByRule 从 GroupRule 枚举取 Grouper。
//
// 无效规则属于编程错误，会 panic（fail-fast，与 baseline/datadriven 一致）。
func ByRule(rule GroupRule) Grouper {
	g, ok := ruleGroupers[rule]
	if !ok {
		panic(fmt.Sprintf("decorator.ByRule: 无效的 GroupRule %d", int(rule)))
	}
	return g
}

// Parse 从 API 参数字符串解析规则。
//
// 空字符串返回默认规则 ExactCombination，便于前端省略参数。
// 未知字符串返回错误，不静默回退。
func Parse(value string) (Grouper, error) {
	if value == "" {
		return ExactCombination, nil
	}
	g, ok := allGroupers[value]
	if !ok {
		return nil, fmt.Errorf("decorator.Parse: 未知分组规则 %q", value)
	}
	return g, nil
}

// Names 返回所有规则名称（稳定顺序）。
//
// 用于 API 文档、前端下拉框等需要"列出所有规则"的场景。
// 枚举能力是 decorator 相对"纯函数值"的额外设施，必须显式维护。
func Names() []string {
	names := make([]string, 0, len(allGroupers))
	for name := range allGroupers {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
