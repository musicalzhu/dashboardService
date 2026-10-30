package baseline

import (
	"fmt"
	"strings"
)

// AllServiceCombinations 生成所有服务类型的组合，确保每个组合都能被分组规则正确处理。
func AllServiceCombinations() []Server {
	serviceTypes := []ServiceType{ASType, TSType, SSType, DSType}
	servers := make([]Server, 0, (1<<len(serviceTypes))-1)
	cluster := Cluster{ClusterID: 1, ClusterRole: PrimaryClusterRole, ClusterName: "cluster-1", State: StateNormal}
	for mask := 1; mask < 1<<len(serviceTypes); mask++ {
		var services []Service
		var names []string
		for bit, serviceType := range serviceTypes {
			if mask&(1<<(len(serviceTypes)-1-bit)) == 0 {
				continue
			}
			services = append(services, Service{ServiceType: serviceType, State: StateNormal})
			names = append(names, string(serviceType))
		}
		servers = append(servers, Server{
			Name:     strings.Join(names, "+"),
			Cluster:  cluster,
			IP:       "1.1.1.1",
			State:    StateNormal,
			Services: services,
		})
	}
	return servers
}

// PrintClusters 按统一格式打印集群分组结果。
func PrintClusters(clusters []ClusterGroup) {
	for _, cluster := range clusters {
		for _, group := range cluster.Groups {
			fmt.Printf("%-20s (%d servers):", group.Name, len(group.Servers))
			for _, server := range group.Servers {
				fmt.Printf(" %s", server.Name)
			}
			fmt.Println()
		}
	}
}
