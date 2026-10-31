package strategy

import (
	"fmt"
	"slices"
	"strings"

	"dashboardService/baseline"
)

// 本文件的辅助函数与 baseline 内部实现语义一致，但必须在本包重复实现：
// baseline 的 prepareServer / compareServers / groupName 是未导出的，
// 纯策略模式无法直接复用，这是"每个策略自包含"的必然结果。
//
// POC 诚实说明：
// 这些函数本身不属于"分组骨架"（它们是原子操作），所有策略共享是合理的。
// 真正的重复在 rules.go 中——每个策略的 Group 方法都完整重复了
// "集群分桶 → 组内分桶 → 排序 → 组装"这四步模板。

// prepareServer 克隆 Services 并按服务优先级排序，避免修改调用方切片。
func prepareServer(server Server) Server {
	server.Services = slices.Clone(server.Services)
	slices.SortStableFunc(server.Services, func(a, b Service) int {
		return compareServicePriority(a.ServiceType, b.ServiceType)
	})
	return server
}

// compareServicePriority 按服务类型优先级比较。
func compareServicePriority(a, b ServiceType) int {
	return serviceTypeRank(a) - serviceTypeRank(b)
}

// serviceTypeRank 返回服务类型的优先级（AS=0, TS=1, SS=2, DS=3）。
func serviceTypeRank(t ServiceType) int {
	switch t {
	case baseline.ASType:
		return 0
	case baseline.TSType:
		return 1
	case baseline.SSType:
		return 2
	case baseline.DSType:
		return 3
	default:
		return 4
	}
}

// compareServers 组内服务器排序：状态 → 服务序列 → IP → Name。
func compareServers(a, b Server) int {
	if a.State != b.State {
		if a.State == baseline.StateAbnormal {
			return -1
		}
		return 1
	}
	if c := compareServerServices(a.Services, b.Services); c != 0 {
		return c
	}
	if a.IP != b.IP {
		if a.IP < b.IP {
			return -1
		}
		return 1
	}
	if a.Name != b.Name {
		if a.Name < b.Name {
			return -1
		}
		return 1
	}
	return 0
}

// compareServerServices 按序比较服务器的服务类型和角色。
func compareServerServices(a, b []Service) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if c := compareServicePriority(a[i].ServiceType, b[i].ServiceType); c != 0 {
			return c
		}
		if c := compareRolePriority(a[i].Role, b[i].Role); c != 0 {
			return c
		}
	}
	return len(a) - len(b)
}

// compareRolePriority 按角色优先级比较。
func compareRolePriority(a, b Role) int {
	return roleRank(a) - roleRank(b)
}

// roleRank 返回角色优先级。
func roleRank(r Role) int {
	switch r {
	case baseline.Leader:
		return 0
	case baseline.Follower:
		return 1
	case baseline.LSS:
		return 2
	case baseline.SSRole:
		return 3
	case baseline.DSRole:
		return 4
	default:
		return 5
	}
}

// compareClusters 集群排序：角色 → ClusterID。
func compareClusters(a, b ClusterGroup) int {
	if a.ClusterRole != b.ClusterRole {
		return clusterRoleRank(a.ClusterRole) - clusterRoleRank(b.ClusterRole)
	}
	return a.ClusterID - b.ClusterID
}

// clusterRoleRank 返回集群角色优先级。
func clusterRoleRank(r baseline.ClusterRole) int {
	if r == baseline.PrimaryClusterRole {
		return 0
	}
	return 1
}

// groupName 返回分组键对应的显示名称。
func groupName(key int) string {
	switch key {
	case -1:
		return "未知服务"
	case 0:
		return "无服务"
	}
	if key > 0 && key <= 0b1111 {
		return maskToName(key)
	}
	return fmt.Sprintf("键 %d", key)
}

// maskToName 把 mask 转为服务组合名称（AS+TS+...）。
func maskToName(mask int) string {
	var names []string
	for _, t := range []ServiceType{baseline.ASType, baseline.TSType, baseline.SSType, baseline.DSType} {
		if bitOf(t) != 0 && mask&bitOf(t) != 0 {
			names = append(names, string(t))
		}
	}
	return strings.Join(names, "+")
}

// bitOf 返回服务类型对应的 mask 位。
func bitOf(t ServiceType) int {
	switch t {
	case baseline.ASType:
		return 0b1000
	case baseline.TSType:
		return 0b0100
	case baseline.SSType:
		return 0b0010
	case baseline.DSType:
		return 0b0001
	default:
		return 0
	}
}
