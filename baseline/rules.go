package baseline

import "fmt"

// GroupRule 表示一个服务器分组规则。
//
// 前四种规则按完整服务组合分组，只改变组间顺序；
// RulePrimaryService 则按最高优先级服务合并分组。
type GroupRule int

const (
	// RuleExactCombination 按完整组合 mask 降序排列。
	RuleExactCombination GroupRule = iota
	// RuleExactServiceCount 按服务数升序、再按 mask 降序排列。
	RuleExactServiceCount
	// RuleExactPrimary 按主服务优先级、服务数、mask 排列。
	RuleExactPrimary
	// RuleExactConvention 按 comboOrderList 中的业务顺序排列。
	RuleExactConvention
	// RulePrimaryService 按最高优先级服务归组，再按 mask 降序排列。
	RulePrimaryService
)

// String 返回规则的稳定名称。
func (r GroupRule) String() string {
	return [...]string{
		"exact-combination",
		"exact-service-count",
		"exact-primary",
		"exact-convention",
		"primary-service",
	}[r]
}

// ParseGroupRule 将配置或请求中的名称解析为规则。
// 空字符串使用 RuleExactCombination；未知名称返回错误，不静默回退。
func ParseGroupRule(value string) (GroupRule, error) {
	switch value {
	case "", "exact-combination":
		return RuleExactCombination, nil
	case "exact-service-count":
		return RuleExactServiceCount, nil
	case "exact-primary":
		return RuleExactPrimary, nil
	case "exact-convention":
		return RuleExactConvention, nil
	case "primary-service":
		return RulePrimaryService, nil
	default:
		return 0, fmt.Errorf("未知分组规则: %q", value)
	}
}
