package main

import (
	"fmt"

	"dashboardService/baseline"
)

// main 展示五种规则在同一组输入数据上的分组与排序差异。
func main() {
	servers := baseline.AllServiceCombinations()
	servers = append(servers,
		baseline.Server{
			Name: "unknown-service-node",
			Services: []baseline.Service{
				{ServiceType: baseline.ServiceType("XX"), State: baseline.SvcNormal},
			},
		},
		baseline.Server{Name: "no-service-node", IP: "192.168.1.200"},
	)

	rules := []baseline.GroupRule{
		baseline.RuleExactCombination,
		baseline.RuleExactServiceCount,
		baseline.RuleExactPrimary,
		baseline.RuleExactConvention,
		baseline.RulePrimaryService,
	}

	for _, rule := range rules {
		fmt.Printf("\n=== %s ===\n", rule)
		for _, group := range baseline.RenderDashboard(servers, rule) {
			fmt.Printf("%-20s (%d servers):", group.Name, len(group.Servers))
			for _, server := range group.Servers {
				fmt.Printf(" %s", server.Name)
			}
			fmt.Println()
		}
	}
}
