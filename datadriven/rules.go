package datadriven

import "dashboardService/baseline"

// GroupRule 复用 baseline 的规则类型，使各 POC 的规则标识保持一致。
type GroupRule = baseline.GroupRule

const (
	RuleExactCombination  = baseline.RuleExactCombination
	RuleExactServiceCount = baseline.RuleExactServiceCount
	RuleExactPrimary      = baseline.RuleExactPrimary
	RuleExactConvention   = baseline.RuleExactConvention
	RulePrimaryService    = baseline.RulePrimaryService
)

type ruleImpl struct {
	keyOf baseline.KeyFunc
	cmp   baseline.CompareFunc
}

// ruleImpls 是本方案唯一的规则分发点；分组、排序和序列化行为由 baseline 实现并测试。
var ruleImpls = map[GroupRule]ruleImpl{
	RuleExactCombination:  {baseline.MaskOf, baseline.CompareByMaskDesc},
	RuleExactServiceCount: {baseline.MaskOf, baseline.CompareByServiceCount},
	RuleExactPrimary:      {baseline.MaskOf, baseline.CompareByPrimary},
	RuleExactConvention:   {baseline.MaskOf, baseline.CompareByConvention},
	RulePrimaryService:    {baseline.PrimaryServiceMask, baseline.CompareByMaskDesc},
}

// ParseGroupRule 调用共享 baseline 校验并解析规则名称。
func ParseGroupRule(value string) (GroupRule, error) {
	return baseline.ParseGroupRule(value)
}
