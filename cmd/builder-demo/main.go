// builder 方案的演示程序。
//
// 演示五件事：
//  1. 用 RenderDashboard 跑五种预定义规则（与其它四个方案输出一致）；
//  2. 用预定义构造器 + 链式调用（NewExactCombination().Build）；
//  3. 用 New().WithKeyOf().WithCmp() 自由组合（等价于 options 的自由组合）；
//  4. 配置复用：一次配置，多次 Build 不同的 servers 批次（Builder 的独特能力）；
//  5. 默认配置：不调 WithXxx 直接 Build，等价于 RuleExactCombination。
package main

import (
	"fmt"

	"dashboardService/baseline"
	"dashboardService/builder"
)

func main() {
	servers := baseline.AllServiceCombinations()

	// ========== 演示 1：按 GroupRule 渲染（与其它四个方案一致） ==========
	fmt.Println("================ 预定义规则（RenderDashboard） ================")
	for _, rule := range []builder.GroupRule{
		builder.RuleExactCombination,
		builder.RuleExactServiceCount,
		builder.RuleExactPrimary,
		builder.RuleExactConvention,
		builder.RulePrimaryService,
	} {
		fmt.Printf("\n=== %s ===\n", rule)
		baseline.PrintClusters(builder.RenderDashboard(servers, rule))
	}

	// ========== 演示 2：预定义构造器 + 链式调用 ==========
	fmt.Println("\n================ 预定义构造器（NewXxx().Build） ================")
	fmt.Printf("\n=== PrimaryService ===\n")
	baseline.PrintClusters(builder.NewPrimaryService().Build(servers))

	// ========== 演示 3：自由组合（等价于 options 的自由组合） ==========
	// 组合："主服务归组 + 服务数量升序"
	// 这个组合在 datadriven 的 ruleImpls 表中不存在，但 builder 可以构造。
	fmt.Println("\n================ 自由组合：主服务归组 + 数量升序 ================")
	customBuilder := builder.New().
		WithKeyOf(baseline.PrimaryServiceMask).
		WithCmp(baseline.CmpByServiceCount)
	baseline.PrintClusters(customBuilder.Build(servers))

	// ========== 演示 4：配置复用（Builder 的独特能力） ==========
	// 同一个 Builder 用于两个不同的 servers 批次。
	fmt.Println("\n================ 配置复用：同一 Builder 处理多个批次 ================")
	reusable := builder.NewPrimaryService()
	batch1 := servers[:5]
	batch2 := servers[5:]
	fmt.Printf("\n--- 批次 1（前 5 台） ---\n")
	baseline.PrintClusters(reusable.Build(batch1))
	fmt.Printf("\n--- 批次 2（后 10 台） ---\n")
	baseline.PrintClusters(reusable.Build(batch2))

	// ========== 演示 5：默认配置 ==========
	// 不调 WithXxx 直接 Build，等价于 RuleExactCombination。
	fmt.Println("\n================ 默认配置（New().Build） ================")
	baseline.PrintClusters(builder.New().Build(servers))
}
