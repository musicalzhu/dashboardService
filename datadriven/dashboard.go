package datadriven

import "dashboardService/baseline"

// RenderDashboard 从数据驱动规则表中选择共享函数，再委托统一集群分组流程。
func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
	r, ok := ruleImpls[rule]
	if !ok {
		panic("datadriven.RenderDashboard: invalid GroupRule")
	}
	return baseline.GroupServersByCluster(servers, r)
}
