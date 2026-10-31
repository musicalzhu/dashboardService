package builder

import (
	"fmt"

	"dashboardService/baseline"
)

// ============================================================================
// 预定义构造器：与 datadriven 的五个规则、options 的五个选项组、closure 的五个 Grouper 一一对等
// ============================================================================
//
// 命名规范：
//   - NewXxx() 而非 Xxx()，避免与包内其它函数冲突（如 NewExactCombination 不与 New 冲突）。
//   - 每个构造器返回的 Builder 已完成全部配置，调用方只需 .Build(servers)。

// NewExactCombination 返回按精确组合 + mask 降序配置的 Builder。
// 对应 baseline.RuleCombination / datadriven.RuleExactCombination。
func NewExactCombination() *Builder {
	return New().
		WithKeyOf(baseline.MaskOf).
		WithCmp(baseline.CmpByMask)
}

// NewExactServiceCount 返回按精确组合 + 服务数量升序配置的 Builder。
func NewExactServiceCount() *Builder {
	return New().
		WithKeyOf(baseline.MaskOf).
		WithCmp(baseline.CmpByServiceCount)
}

// NewExactPrimary 返回按精确组合 + 主服务优先配置的 Builder。
func NewExactPrimary() *Builder {
	return New().
		WithKeyOf(baseline.MaskOf).
		WithCmp(baseline.CmpByPrimary)
}

// NewExactConvention 返回按精确组合 + 业务约定顺序配置的 Builder。
func NewExactConvention() *Builder {
	return New().
		WithKeyOf(baseline.MaskOf).
		WithCmp(baseline.CmpByConvention)
}

// NewPrimaryService 返回按主服务归组 + mask 降序配置的 Builder。
func NewPrimaryService() *Builder {
	return New().
		WithKeyOf(baseline.PrimaryServiceMask).
		WithCmp(baseline.CmpByMask)
}

// NewFromRule 根据 GroupRule 返回对应的预定义 Builder。
//
// 这是 builder 方案与 baseline/datadriven/options/closure 的"规则枚举对齐点"：
//
//	baseline.RenderDashboard(servers, rule)   // switch 分发
//	datadriven.RenderDashboard(servers, rule) // 查表分发
//	options.RenderByRule(servers, rule)       // 取选项列表再应用
//	closure.RenderDashboard(servers, rule)    // 取 Grouper 后调用
//	builder.NewFromRule(rule).Build(servers)  // 取构造器后 Build
//
// 无效规则属于编程错误，会 panic。
func NewFromRule(rule GroupRule) *Builder {
	switch rule {
	case RuleExactCombination:
		return NewExactCombination()
	case RuleExactServiceCount:
		return NewExactServiceCount()
	case RuleExactPrimary:
		return NewExactPrimary()
	case RuleExactConvention:
		return NewExactConvention()
	case RulePrimaryService:
		return NewPrimaryService()
	default:
		panic(fmt.Sprintf("builder.NewFromRule: 无效的 GroupRule %d", int(rule)))
	}
}

// ParseGroupRule 调用 baseline 校验并解析规则名称。
func ParseGroupRule(value string) (GroupRule, error) {
	return baseline.ParseGroupRule(value)
}
