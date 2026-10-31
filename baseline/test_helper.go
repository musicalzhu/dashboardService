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

// RoleVariations 生成用于验证组内角色排序的测试数据。
//
// 所有服务器属于同一集群，具有相同的服务组合（AS），但服务角色不同：
// Leader(0) < Follower(1) < LSS(2) < SSRole(3) < DSRole(4)。
//
// 服务器 IP 故意设置为与角色顺序相反（Leader 的 IP 最大），
// 用于证明 compareServers 是"按角色排序"而不是"按 IP 排序"。
//
// 预期排序结果（RuleCombination 下，所有服务器归入 AS 组）：
// as-leader < as-follower < as-lss < as-ss < as-ds
func RoleVariations() []Server {
	cluster := Cluster{ClusterID: 1, ClusterRole: PrimaryClusterRole, ClusterName: "cluster-1", State: StateNormal}
	roles := []ServiceRole{Leader, Follower, LSS, SSRole, DSRole}
	servers := make([]Server, len(roles))
	for i, role := range roles {
		servers[i] = Server{
			Name:    "as-" + string(role),
			Cluster: cluster,
			// IP 与角色顺序相反：Leader(.5) → Follower(.4) → ... → DS(.1)
			// 如果排序是按 IP，结果会完全相反——这证明排序是按角色而非 IP
			IP:    fmt.Sprintf("192.168.1.%d", len(roles)-i),
			State: StateNormal,
			Services: []Service{
				{ServiceType: ASType, Role: role, State: StateNormal},
			},
		}
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
