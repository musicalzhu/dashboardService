package main

import (
	"fmt"

	"dashboardService/closure"
)

// main 展示将闭包规则作为函数直接调用，无需规则枚举或分发表查询。
func main() {
	servers := []closure.Server{
		{Name: "as-node", Services: []closure.Service{{ServiceType: closure.AS, State: closure.SvcNormal}}},
		{Name: "ts-node", Services: []closure.Service{{ServiceType: closure.TS, State: closure.SvcNormal}}},
		{Name: "unknown-service-node", Services: []closure.Service{
			{ServiceType: closure.ServiceType("XX"), State: closure.SvcNormal},
		}},
		{Name: "no-service-node"},
	}

	for _, cluster := range closure.RulePrimaryService(servers) {
		for _, group := range cluster.Groups {
			fmt.Printf("%-20s (%d servers):", group.Name, len(group.Servers))
			for _, server := range group.Servers {
				fmt.Printf(" %s", server.Name)
			}
			fmt.Println()
		}
	}
}
