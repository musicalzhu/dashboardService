// options 方案的演示程序。
//
// 演示三件事：
//  1. 用 RenderByRule 跑五种预定义规则（与 baseline/datadriven 输出一致）；
//  2. 用 RenderDashboard + 预定义选项组合跑同样的规则（等价形式）；
//  3. 用 RenderDashboard + 自由组合构造"表中不存在的规则"（options 的独特能力）。
package main

import (
	"fmt"

	"dashboardService/baseline"
	"dashboardService/options"
)

func main() {
	servers := baseline.AllServiceCombinations()

	// ---------- 演示 1：用 GroupRule 选规则，与 baseline/datadriven 输出一致 ----------
	fmt.Println("================ 预定义规则（RenderByRule） ================")
	for _, rule := range []options.GroupRule{
		options.RuleExactCombination,
		options.RuleExactServiceCount,
		options.RuleExactPrimary,
		options.RuleExactConvention,
		options.RulePrimaryService,
	} {
		fmt.Printf("\n=== %s ===\n", rule)
		baseline.PrintClusters(options.RenderByRule(servers, rule))
	}

	// ---------- 演示 2：用预定义选项组合，等价形式 ----------
	fmt.Println("\n================ 预定义选项组合（RenderDashboard + opts...） ================")
	fmt.Printf("\n=== PrimaryService ===\n")
	baseline.PrintClusters(options.RenderDashboard(servers, options.PrimaryService...))

	// ---------- 演示 3：自由组合出"表中不存在的规则" ----------
	// 组合："主服务归组 + 服务数量升序"
	// 这个组合在 datadriven 的 ruleImpls 表中不存在，但 options 可以构造。
	fmt.Println("\n================ 自由组合：主服务归组 + 数量升序 ================")
	baseline.PrintClusters(options.RenderDashboard(servers,
		options.WithKeyOf(baseline.PrimaryServiceMask),
		options.WithCmp(baseline.CmpByServiceCount),
	))
}
