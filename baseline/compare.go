package baseline

import (
	"cmp"
	"math/bits"
	"slices"
)

const (
	roleSS Role = "ss"
	roleDS Role = "ds"
)

// prepareServer 克隆 Services 后按服务优先级排序，避免修改调用方切片。
func prepareServer(server Server) Server {
	server.Services = slices.Clone(server.Services)
	slices.SortStableFunc(server.Services, func(a, b Service) int {
		return cmp.Compare(serviceTypeRank(a.ServiceType), serviceTypeRank(b.ServiceType))
	})
	return server
}

// compareServers 按服务器状态、服务类型/角色序列、IP、Name 排序。
func compareServers(a, b Server) int {
	if c := compareState(a.State, b.State); c != 0 {
		return c
	}
	serviceCount := min(len(a.Services), len(b.Services))
	for i := 0; i < serviceCount; i++ {
		serviceA, serviceB := a.Services[i], b.Services[i]
		if c := cmp.Compare(serviceTypeRank(serviceA.ServiceType), serviceTypeRank(serviceB.ServiceType)); c != 0 {
			return c
		}
		if c := compareRole(serviceA.Role, serviceB.Role); c != 0 {
			return c
		}
	}
	if len(a.Services) != len(b.Services) {
		return cmp.Compare(len(a.Services), len(b.Services))
	}
	if c := cmp.Compare(a.IP, b.IP); c != 0 {
		return c
	}
	return cmp.Compare(a.Name, b.Name)
}

func compareState(a, b State) int {
	if a == b {
		return 0
	}
	rank := func(state State) int {
		switch state {
		case StateAbnormal:
			return 0
		case StateNormal:
			return 1
		default:
			return 2
		}
	}
	if c := cmp.Compare(rank(a), rank(b)); c != 0 {
		return c
	}
	return cmp.Compare(string(a), string(b))
}

func compareRole(a, b Role) int {
	if a == b {
		return 0
	}
	rank := func(role Role) int {
		switch role {
		case Leader:
			return 0
		case Follower:
			return 1
		case LSS:
			return 2
		case roleSS:
			return 3
		case roleDS:
			return 4
		default:
			return 5
		}
	}
	if c := cmp.Compare(rank(a), rank(b)); c != 0 {
		return c
	}
	return cmp.Compare(string(a), string(b))
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
	if c := cmp.Compare(bits.OnesCount(uint(a)), bits.OnesCount(uint(b))); c != 0 {
		return c
	}
	return compareIntDesc(a, b)
}

// CompareByServiceCount 先按服务数量升序，再按 mask 降序比较分组键。
func CompareByServiceCount(a, b int) int {
	return cmpByServiceCount(a, b)
}

// cmpByPrimary 按主服务索引、服务数量、mask 依次排序。
func cmpByPrimary(a, b int) int {
	if c := cmp.Compare(primaryIndex(a), primaryIndex(b)); c != 0 {
		return c
	}
	if c := cmp.Compare(bits.OnesCount(uint(a)), bits.OnesCount(uint(b))); c != 0 {
		return c
	}
	return compareIntDesc(a, b)
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
