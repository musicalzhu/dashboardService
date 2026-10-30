package datadriven

import "dashboardService/baseline"

// ruleImpls 是本方案唯一的规则分发点；分组、排序和序列化行为由 baseline 实现并测试。
var ruleImpls = map[GroupRule]baseline.RuleImpl{
	RuleExactCombination:  {KeyOf: baseline.MaskOf, Cmp: baseline.CmpByMask},
	RuleExactServiceCount: {KeyOf: baseline.MaskOf, Cmp: baseline.CmpByServiceCount},
	RuleExactPrimary:      {KeyOf: baseline.MaskOf, Cmp: baseline.CmpByPrimary},
	RuleExactConvention:   {KeyOf: baseline.MaskOf, Cmp: baseline.CmpByConvention},
	RulePrimaryService:    {KeyOf: baseline.PrimaryServiceMask, Cmp: baseline.CmpByMask},
}

// ParseGroupRule 调用共享 baseline 校验并解析规则名称。
func ParseGroupRule(value string) (GroupRule, error) {
	return baseline.ParseGroupRule(value)
}
