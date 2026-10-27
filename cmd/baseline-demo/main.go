package main

import (
	"fmt"
	"strings"

	"dashboardService/baseline"
)

// main 展示五种规则在同一组输入数据上的分组与排序差异。
func main() {
	servers := allServiceCombinations()
	servers = append(servers,
		baseline.Server{
			Name: "unknown-service-node",
			Services: []baseline.Service{
				{ServiceType: baseline.ServiceType("XX"), State: baseline.ServiceNormal},
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

func allServiceCombinations() []baseline.Server {
	serviceTypes := []baseline.ServiceType{baseline.ASType, baseline.TSType, baseline.SSType, baseline.DSType}
	servers := make([]baseline.Server, 0, (1<<len(serviceTypes))-1)
	for mask := 1; mask < 1<<len(serviceTypes); mask++ {
		var services []baseline.Service
		var names []string
		for bit, serviceType := range serviceTypes {
			if mask&(1<<(len(serviceTypes)-1-bit)) == 0 {
				continue
			}
			services = append(services, baseline.Service{
				ServiceType: serviceType,
				State:       baseline.StateNormal,
			})
			names = append(names, string(serviceType))
		}
		servers = append(servers, baseline.Server{
			Name:     strings.Join(names, "+"),
			IP:       "1.1.1.1",
			State:    baseline.StateNormal,
			Services: services,
		})
	}
	return servers
}
