// strategy 方案的演示程序。
//
// 演示：
//  1. 用 RenderDashboard 跑五种预定义规则（与其它五个方案输出一致）；
//  2. 直接调用策略实例（ByRule(rule).Group(servers)）；
//  3. 演示"每个策略自包含"的实际效果——五个策略的 Group 方法各自完整。
package main

import (
	"fmt"

	"dashboardService/baseline"
	"dashboardService/strategy"
)

func main() {
	servers := baseline.AllServiceCombinations()

	// ========== 演示 1：按 GroupRule 渲染（与其它五个方案一致） ==========
	fmt.Println("================ 预定义规则（RenderDashboard） ================")
	for _, rule := range []strategy.GroupRule{
		strategy.RuleExactCombination,
		strategy.RuleExactServiceCount,
		strategy.RuleExactPrimary,
		strategy.RuleExactConvention,
		strategy.RulePrimaryService,
	} {
		fmt.Printf("\n=== %s ===\n", rule)
		baseline.PrintClusters(strategy.RenderDashboard(servers, rule))
	}

	// ========== 演示 2：直接调用策略实例 ==========
	fmt.Println("\n================ 直接调用策略 ================")
	s := strategy.ByRule(strategy.RulePrimaryService)
	fmt.Printf("策略名: %s\n", s.Name())
	baseline.PrintClusters(s.Group(servers))
}
