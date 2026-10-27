package baseline

import (
	"cmp"
	"math/bits"
	"slices"
)

// prepareServer 克隆 Services 后按服务优先级排序，避免修改调用方切片。
func prepareServer(server Server) Server {
	server.Services = slices.Clone(server.Services)
	slices.SortStableFunc(server.Services, func(a, b Service) int {
		return cmp.Compare(priorityOf(servicePriority, a.ServiceType), priorityOf(servicePriority, b.ServiceType))
	})
	return server
}

// compareServers 按服务器状态、服务类型/角色序列、IP、Name 排序。
func compareServers(a, b Server) int {
	return cmp.Or(
		compareState(a.State, b.State),
		compareServerServices(a.Services, b.Services),
		cmp.Compare(len(a.Services), len(b.Services)),
		cmp.Compare(a.IP, b.IP),
		cmp.Compare(a.Name, b.Name),
	)
}

// compareServerServices 按序比较服务器的服务类型和角色；相同前缀视为相等。
func compareServerServices(a, b []Service) int {
	for i := 0; i < min(len(a), len(b)); i++ {
		serviceA, serviceB := a[i], b[i]
		if c := cmp.Or(
			cmp.Compare(priorityOf(servicePriority, serviceA.ServiceType), priorityOf(servicePriority, serviceB.ServiceType)),
			compareRole(serviceA.Role, serviceB.Role),
		); c != 0 {
			return c
		}
	}
	return 0
}

func compareState(a, b State) int {
	if a == b {
		return 0
	}
	return cmp.Or(
		cmp.Compare(priorityOf(statePriority, a), priorityOf(statePriority, b)),
		cmp.Compare(string(a), string(b)),
	)
}

func compareRole(a, b ServiceRole) int {
	if a == b {
		return 0
	}
	return cmp.Or(
		cmp.Compare(priorityOf(rolePriority, a), priorityOf(rolePriority, b)),
		cmp.Compare(string(a), string(b)),
	)
}

// cmpByMaskDesc 按组合 mask 降序排列，高优先级组合在前。
func cmpByMaskDesc(a, b int) int {
	return compareIntDesc(a, b)
}

// CompareByMaskDesc 按 mask 降序比较分组键。
func CompareByMaskDesc(a, b int) int {
	return cmpByMaskDesc(a, b)
}

// cmpByServiceCount 按服务数量升序，再按 mask 降序排列。
func cmpByServiceCount(a, b int) int {
	return cmp.Or(
		cmp.Compare(bits.OnesCount(uint(a)), bits.OnesCount(uint(b))),
		compareIntDesc(a, b),
	)
}

// CompareByServiceCount 先按服务数量升序，再按 mask 降序比较分组键。
func CompareByServiceCount(a, b int) int {
	return cmpByServiceCount(a, b)
}

// cmpByPrimary 按主服务索引、服务数量、mask 依次排序。
func cmpByPrimary(a, b int) int {
	return cmp.Or(
		cmp.Compare(primaryIndex(a), primaryIndex(b)),
		cmp.Compare(bits.OnesCount(uint(a)), bits.OnesCount(uint(b))),
		compareIntDesc(a, b),
	)
}

// CompareByPrimary 依次按主服务、服务数量和 mask 比较分组键。
func CompareByPrimary(a, b int) int {
	return cmpByPrimary(a, b)
}

// cmpByConvention 按业务顺序表排序；未登记组合稳定地排在登记组合之后。
func cmpByConvention(a, b int) int {
	rankA := slices.Index(comboOrderList, a)
	rankB := slices.Index(comboOrderList, b)
	switch {
	case rankA >= 0 && rankB >= 0:
		return cmp.Compare(rankA, rankB)
	case rankA < 0 && rankB < 0:
		return compareIntDesc(a, b)
	case rankA < 0:
		return 1
	default:
		return -1
	}
}

func compareIntDesc(a, b int) int {
	return cmp.Compare(b, a)
}

// CompareByConvention 按预先定义的业务顺序比较分组键。
func CompareByConvention(a, b int) int {
	return cmpByConvention(a, b)
}
