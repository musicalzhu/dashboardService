package builder

// RenderDashboard 是 builder 方案的便捷入口：按 GroupRule 构造 Builder 并 Build。
//
// 等价于 NewFromRule(rule).Build(servers)。
// 提供它是为了让五个 POC 方案的"按规则名渲染"调用形式完全一致：
//
//	baseline.RenderDashboard(servers, rule)    // switch 分发
//	datadriven.RenderDashboard(servers, rule)  // 查表分发
//	options.RenderByRule(servers, rule)        // 取选项列表再应用
//	closure.RenderDashboard(servers, rule)     // 取 Grouper 后调用
//	builder.RenderDashboard(servers, rule)     // 取 Builder 后 Build
//
// 注意：builder 的真正入口是链式调用（New().WithXxx().Build()），
// 不需要 RenderDashboard。提供本函数只是为了让对比时的调用形式一致。
func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
	return NewFromRule(rule).Build(servers)
}
