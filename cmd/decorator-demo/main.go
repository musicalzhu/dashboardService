// decorator 方案的演示程序。
//
// 演示五件事：
//  1. 用 RenderDashboard 跑五种预定义规则（与 baseline/datadriven/options/builder 输出一致）；
//  2. 直接调用预定义 Grouper（decorator 的"规则即函数"特性）；
//  3. 用 Parse 从字符串解析规则；
//  4. 用装饰器只给主服务规则加日志（decorator 的独特能力）；
//  5. 用 ApplyDecorators 组合多个装饰器。
package main

import (
	"fmt"
	"time"

	"dashboardService/baseline"
	"dashboardService/decorator"
)

func main() {
	servers := baseline.AllServiceCombinations()

	// ========== 演示 1：按 GroupRule 渲染（与其它四个方案一致） ==========
	fmt.Println("================ 预定义规则（RenderDashboard） ================")
	for _, rule := range []decorator.GroupRule{
		decorator.RuleExactCombination,
		decorator.RuleExactServiceCount,
		decorator.RuleExactPrimary,
		decorator.RuleExactConvention,
		decorator.RulePrimaryService,
	} {
		fmt.Printf("\n=== %s ===\n", rule)
		baseline.PrintClusters(decorator.RenderDashboard(servers, rule))
	}

	// ========== 演示 2：直接调用 Grouper（decorator 的"规则即函数"特性） ==========
	fmt.Println("\n================ 直接调用 Grouper ================")
	fmt.Printf("\n=== PrimaryService ===\n")
	baseline.PrintClusters(decorator.PrimaryService(servers))

	// ========== 演示 3：从 API 字符串解析 ==========
	fmt.Println("\n================ Parse 规则 ================")
	g, err := decorator.Parse("rule-primary-service")
	if err != nil {
		panic(err)
	}
	baseline.PrintClusters(g(servers))

	// ========== 演示 4：只给主服务规则加日志，其他规则不受影响 ==========
	fmt.Println("\n================ 装饰器：只给 PrimaryService 加日志 ================")
	logFn := func(format string, args ...any) {
		fmt.Printf("[LOG] "+format+"\n", args...)
	}
	primaryWithLog := decorator.WithLogging(logFn)(decorator.PrimaryService)

	fmt.Println("--- 调用带日志的 PrimaryService ---")
	primaryWithLog(servers)

	fmt.Println("--- 调用不带日志的 ExactCombination ---")
	decorator.ExactCombination(servers) // 无日志输出

	// ========== 演示 5：装饰器链（容错 + 指标 + 日志） ==========
	fmt.Println("\n================ 装饰器链：容错 + 指标 + 日志 ================")
	composed := decorator.ApplyDecorators(
		decorator.PrimaryService,
		decorator.WithRecover(func(v any) {
			fmt.Printf("[RECOVER] panic: %v\n", v)
		}),
		decorator.WithMetrics(func(d time.Duration) {
			fmt.Printf("[METRICS] duration=%v\n", d)
		}),
		decorator.WithLogging(logFn),
	)
	composed(servers)
}
