package strategy

// RenderDashboard 是 strategy 方案的便捷入口：按 GroupRule 取策略并执行。
//
// 等价于 ByRule(rule).Group(servers)。
// 提供它是为了让六个 POC 方案的"按规则名渲染"调用形式完全一致：
//
//	baseline.RenderDashboard(servers, rule)    // switch 分发
//	datadriven.RenderDashboard(servers, rule)  // 查表分发
//	options.RenderByRule(servers, rule)        // 取选项列表再应用
//	decorator.RenderDashboard(servers, rule)   // 取 Grouper 后直接调用
//	builder.RenderDashboard(servers, rule)     // 取 Builder 后 Build
//	strategy.RenderDashboard(servers, rule)    // 取 Strategy 后 Group
func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
	return ByRule(rule).Group(servers)
}
