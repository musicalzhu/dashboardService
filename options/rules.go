package options

import (
	"fmt"

	"dashboardService/baseline"
)

// 预定义选项组合：五个与 datadriven.ruleImpls 表项对等的选项列表。
//
// 每个变量都是 []Option，调用方用 RenderDashboard(servers, ExactCombination...)
// 即可使用对应规则。与 datadriven 的 ruleImpls 表相比：
//   - datadriven 的表项是"值"（RuleImpl），只能选；
//   - options 的选项是"构造过程"（[]Option），可以选，也可以自由组合。
var (
	// ExactCombination 精确组合 + mask 降序。
	// 对应 baseline.RuleCombination / datadriven.RuleExactCombination。
	ExactCombination = []Option{
		WithKeyOf(baseline.MaskOf),
		WithCmp(baseline.CmpByMask),
	}

	// ExactServiceCount 精确组合 + 服务数量升序。
	ExactServiceCount = []Option{
		WithKeyOf(baseline.MaskOf),
		WithCmp(baseline.CmpByServiceCount),
	}

	// ExactPrimary 精确组合 + 主服务优先。
	ExactPrimary = []Option{
		WithKeyOf(baseline.MaskOf),
		WithCmp(baseline.CmpByPrimary),
	}

	// ExactConvention 精确组合 + 业务约定顺序。
	ExactConvention = []Option{
		WithKeyOf(baseline.MaskOf),
		WithCmp(baseline.CmpByConvention),
	}

	// PrimaryService 主服务归组 + mask 降序。
	PrimaryService = []Option{
		WithKeyOf(baseline.PrimaryServiceMask),
		WithCmp(baseline.CmpByMask),
	}
)

// ruleOptions 从 GroupRule 映射到预定义选项列表。
//
// 这个 map 的作用：
//   - 让 options 方案也能通过 GroupRule 选规则，便于与 baseline/datadriven 对比；
//   - 它不是 options 方案的核心——核心是 RenderDashboard 接受任意 Option 组合。
var ruleOptions = map[GroupRule][]Option{
	RuleExactCombination:  ExactCombination,
	RuleExactServiceCount: ExactServiceCount,
	RuleExactPrimary:      ExactPrimary,
	RuleExactConvention:   ExactConvention,
	RulePrimaryService:    PrimaryService,
}

// ParseGroupRule 调用 baseline 校验并解析规则名称。
func ParseGroupRule(value string) (GroupRule, error) {
	return baseline.ParseGroupRule(value)
}

// OptionsOfRule 返回给定 GroupRule 对应的 Option 列表。
//
// 这是规则映射表的"查询入口"，用于：
//   - RenderByRule 的内部实现；
//   - 调用方需要"按规则名取选项"后再自行组合的场景。
func OptionsOfRule(rule GroupRule) ([]Option, error) {
	opts, ok := ruleOptions[rule]
	if !ok {
		return nil, fmt.Errorf("options.OptionsOfRule: 无效的 GroupRule %d", int(rule))
	}
	return opts, nil
}
