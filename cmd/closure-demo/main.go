package main

import (
	"fmt"

	"dashboardService/baseline"
	"dashboardService/closure"
)

// main 展示将闭包规则作为函数直接调用，无需规则枚举或分发表查询。
func main() {
	servers := baseline.AllServiceCombinations()
	servers = append(servers,
		closure.Server{
			Name:     "unknown-service-node",
			Services: []closure.Service{{ServiceType: closure.ServiceType("XX"), State: closure.SvcNormal}},
		},
		closure.Server{Name: "no-service-node"},
	)

	for _, group := range closure.RulePrimaryService(servers) {
		fmt.Printf("%-20s (%d servers):", group.Name, len(group.Servers))
		for _, server := range group.Servers {
			fmt.Printf(" %s", server.Name)
		}
		fmt.Println()
	}
}
