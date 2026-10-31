package strategy

import "fmt"

// Strategy 是分组策略的接口。
//
// 每个分组规则是一个独立类型，实现 Group 方法。
// Group 方法完整负责：集群分桶 → 组内分桶 → 排序 → 组装。
type Strategy interface {
	// Name 返回策略的规则名，用于日志和 API。
	Name() string
	// Group 执行完整的分组流程。
	Group(servers []Server) []ClusterGroup
}

// strategies 规则枚举到策略实例的映射。
var strategies = map[GroupRule]Strategy{
	RuleExactCombination:  exactCombination{},
	RuleExactServiceCount: exactServiceCount{},
	RuleExactPrimary:      exactPrimary{},
	RuleExactConvention:   exactConvention{},
	RulePrimaryService:    primaryService{},
}

// ByRule 从 GroupRule 枚举取策略实例。
//
// 无效规则是编程错误，会 panic（fail-fast）。
func ByRule(rule GroupRule) Strategy {
	s, ok := strategies[rule]
	if !ok {
		panic(fmt.Sprintf("strategy.ByRule: 无效的 GroupRule %d", int(rule)))
	}
	return s
}
