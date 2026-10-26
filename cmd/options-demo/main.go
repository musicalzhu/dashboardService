package main

import (
	"fmt"

	"dashboardService/baseline"
	"dashboardService/options"
)

// main demonstrates configuring a primary-service grouping with functional options.
func main() {
	grouping, err := options.NewGrouping(
		options.WithKeyFn(baseline.PrimaryServiceMask),
		options.WithCmpFn(baseline.CompareByMaskDesc),
	)
	if err != nil {
		fmt.Println("configure grouping:", err)
		return
	}

	servers := baseline.AllServiceCombinations()
	servers = append(servers,
		baseline.Server{
			Name:     "unknown-service-node",
			Services: []baseline.Service{{ServiceType: baseline.ServiceType("XX"), State: baseline.SvcNormal}},
		},
		baseline.Server{Name: "no-service-node"},
	)
	for _, group := range grouping.Apply(servers) {
		fmt.Printf("%-12s (%d servers):", group.Name, len(group.Servers))
		for _, server := range group.Servers {
			fmt.Printf(" %s", server.Name)
		}
		fmt.Println()
	}
}
