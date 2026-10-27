package main

import (
	"fmt"

	"dashboardService/datadriven"
)

// main 构造演示输入，并逐一展示五种规则如何改变分组或组间顺序。
//
// 演示输入含全部 15 种非空服务组合、一台未知类型服务服务器和一台无服务服务器；
// 因此运行结果也能直观看到两种特殊分组都被保留。
func main() {
	servers := []datadriven.Server{
		{Name: "as-node", Services: []datadriven.Service{{ServiceType: datadriven.AS, State: datadriven.SvcNormal}}},
		{Name: "ts-ss-node", Services: []datadriven.Service{
			{ServiceType: datadriven.TS, State: datadriven.SvcNormal},
			{ServiceType: datadriven.SS, State: datadriven.SvcNormal},
		}},
		{Name: "unknown-service-node", Services: []datadriven.Service{
			{ServiceType: datadriven.ServiceType("XX"), State: datadriven.SvcNormal},
		}},
		{Name: "no-service-node", IP: "192.168.1.200"},
	}

	rules := []datadriven.GroupRule{
		datadriven.RuleExactCombination,
		datadriven.RuleExactServiceCount,
		datadriven.RuleExactPrimary,
		datadriven.RuleExactConvention,
		datadriven.RulePrimaryService,
	}

	for _, rule := range rules {
		fmt.Printf("\n=== %s ===\n", rule)
		for _, cluster := range datadriven.RenderDashboard(servers, rule) {
			for _, group := range cluster.Groups {
				fmt.Printf("%-20s (%d servers):", group.Name, len(group.Servers))
				for _, server := range group.Servers {
					fmt.Printf(" %s", server.Name)
				}
				fmt.Println()
			}
		}
	}
}
