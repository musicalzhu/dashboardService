package main

import (
	"fmt"

	"dashboardService/baseline"
	"dashboardService/datadriven"
)

// main 展示五种规则在同一组输入数据上的分组与排序差异。
func main() {

	servers := baseline.AllServiceCombinations()

	rules := []datadriven.GroupRule{
		datadriven.RuleExactCombination,
		datadriven.RuleExactServiceCount,
		datadriven.RuleExactPrimary,
		datadriven.RuleExactConvention,
		datadriven.RulePrimaryService,
	}

	for _, rule := range rules {
		fmt.Printf("\n=== %s ===\n", rule)
		baseline.PrintClusters(datadriven.RenderDashboard(servers, rule))
	}
}
