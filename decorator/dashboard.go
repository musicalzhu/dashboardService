package decorator

// RenderDashboard 是 decorator 方案的便捷入口：按 GroupRule 选规则并执行。
//
// 等价于 ByRule(rule)(servers)。
// 提供它是为了让五个 POC 方案的"按规则名渲染"调用形式完全一致：
//
//	baseline.RenderDashboard(servers, rule)    // switch 分发
//	datadriven.RenderDashboard(servers, rule)  // 查表分发
//	options.RenderByRule(servers, rule)        // 取选项列表再应用
//	decorator.RenderDashboard(servers, rule)   // 取 Grouper 后直接调用
//	builder.RenderDashboard(servers, rule)     // 取 Builder 后 Build
//
// 注意：decorator 的真正入口是直接调用 Grouper —— decorator.PrimaryService(servers)，
// 不需要 RenderDashboard。提供本函数只是为了让对比时的调用形式一致。
func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
	return ByRule(rule)(servers)
}
