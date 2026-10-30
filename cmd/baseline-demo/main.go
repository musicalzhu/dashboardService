package main

import (
	"fmt"

	"dashboardService/baseline"
)

// main 展示五种规则在同一组输入数据上的分组与排序差异。
func main() {
	servers := baseline.AllServiceCombinations()
	rules := []baseline.GroupRule{
		baseline.RuleCombination,
		baseline.RuleServiceCount,
		baseline.RulePrimary,
		baseline.RuleConvention,
		baseline.RulePrimaryService,
	}

	for _, rule := range rules {
		fmt.Printf("\n=== %s ===\n", rule)
		baseline.PrintClusters(baseline.RenderDashboard(servers, rule))
	}
}
