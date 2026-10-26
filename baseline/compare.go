package baseline

import (
	"math/bits"
	"slices"
	"strings"
)

// prepareServer 克隆 Services 后按服务优先级排序，避免修改调用方切片。
func prepareServer(server Server) Server {
	server.Services = slices.Clone(server.Services)
	slices.SortFunc(server.Services, func(a, b Service) int {
		return serviceTypeRank(a.ServiceType) - serviceTypeRank(b.ServiceType)
	})
	return server
}

// findService 返回服务器上指定类型的第一个服务；不存在时返回 nil。
func findService(server *Server, serviceType ServiceType) *Service {
	for i := range server.Services {
		if server.Services[i].ServiceType == serviceType {
			return &server.Services[i]
		}
	}
	return nil
}

// compareServiceStatus 令存在的服务优先于缺失服务，并令 normal 状态优先。
func compareServiceStatus(a, b *Service) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	}
	if a.State == b.State {
		return 0
	}
	if a.State == SvcNormal {
		return -1
	}
	return 1
}

// compareServers 在同组内按服务器状态、各服务状态、IP、Name 排序。
func compareServers(a, b Server) int {
	if a.State != b.State {
		if a.State == SrvNormal {
			return -1
		}
		return 1
	}
	for _, serviceType := range allServiceTypes {
		if c := compareServiceStatus(findService(&a, serviceType), findService(&b, serviceType)); c != 0 {
			return c
		}
	}
	if c := strings.Compare(a.IP, b.IP); c != 0 {
		return c
	}
	return strings.Compare(a.Name, b.Name)
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
	if c := bits.OnesCount(uint(a)) - bits.OnesCount(uint(b)); c != 0 {
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
	if c := primaryIndex(a) - primaryIndex(b); c != 0 {
		return c
	}
	if c := bits.OnesCount(uint(a)) - bits.OnesCount(uint(b)); c != 0 {
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
		return rankA - rankB
	case rankA < 0 && rankB < 0:
		return compareIntDesc(a, b)
	case rankA < 0:
		return 1
	default:
		return -1
	}
}

func compareIntDesc(a, b int) int {
	if a > b {
		return -1
	}
	if a < b {
		return 1
	}
	return 0
}

// CompareByConvention 按预先定义的业务顺序比较分组键。
func CompareByConvention(a, b int) int {
	return cmpByConvention(a, b)
}
