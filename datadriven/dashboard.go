package datadriven

import "dashboardService/baseline"

// Group 是 baseline 中共享的分组结果类型。
type Group = baseline.Group

// RenderDashboard 从数据驱动规则表中选择共享函数，再委托统一集群分组流程。
func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
	impl, ok := ruleImpls[rule]
	if !ok {
		panic("datadriven.RenderDashboard: invalid GroupRule")
	}
	return baseline.GroupBy(servers, impl.keyOf, impl.cmp)
}
