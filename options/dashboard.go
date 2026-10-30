package options

import (
	"fmt"

	"dashboardService/baseline"
)

// RenderDashboard 应用选项组合，委托共享分组流程。
//
// 三种典型调用方式：
//
//	// 1. 使用预定义选项组合（等价于 datadriven 的"选规则"）
//	options.RenderDashboard(servers, options.PrimaryService...)
//
//	// 2. 使用 GroupRule 选规则（与 baseline/datadriven 同形，便于对比）
//	options.RenderByRule(servers, options.RulePrimaryService)
//
//	// 3. 自由组合出"表中不存在的新规则"
//	options.RenderDashboard(servers,
//	    options.WithKeyOf(baseline.PrimaryServiceMask),
//	    options.WithCmp(baseline.CmpByServiceCount),
//	)
//
// 与 baseline（switch）和 datadriven（查表）的根本区别：
// 传入的不是"规则名"，而是"规则的构造过程"——调用方用 Option 组合出配置，
// 而不是从预定义表中挑选。
func RenderDashboard(servers []Server, opts ...Option) []ClusterGroup {
	return baseline.GroupServersByCluster(servers, newRule(opts))
}

// RenderByRule 是"从 GroupRule 渲染"的便捷入口。
//
// 等价于 RenderDashboard(servers, OptionsOfRule(rule)...)。
// 提供它是为了让三个 POC 方案的"按规则名渲染"调用形式完全一致：
//
//	baseline.RenderDashboard(servers, rule)   // switch 分发
//	datadriven.RenderDashboard(servers, rule) // 查表分发
//	options.RenderByRule(servers, rule)       // 取选项列表再应用
//
// 无效规则属于编程错误，会 panic。
func RenderByRule(servers []Server, rule GroupRule) []ClusterGroup {
	opts, ok := ruleOptions[rule]
	if !ok {
		panic(fmt.Sprintf("options.RenderByRule: 无效的 GroupRule %d", int(rule)))
	}
	return RenderDashboard(servers, opts...)
}
