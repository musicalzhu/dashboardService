package purestrategy

import (
	"cmp"
	"fmt"
	"math/bits"
	"slices"
	"strings"
)

const (
	unknownKey   = -1
	noServiceKey = -2
)

var serviceBits = map[ServiceType]int{
	AS: 8,
	TS: 4,
	SS: 2,
	DS: 1,
}

var conventionOrder = []int{
	8, 10, 9, 11,
	12, 14, 13, 15,
	4, 6, 5, 7,
	2, 3,
	1,
}

type exactCombination struct{}
type exactServiceCount struct{}
type exactPrimary struct{}
type exactConvention struct{}
type primaryService struct{}

func (exactCombination) Group(servers []Server) []Group {
	prepare := func(server Server) Server {
		server.Services = slices.Clone(server.Services)
		slices.SortStableFunc(server.Services, func(a, b Service) int {
			return cmp.Compare(serviceTypeRank(a.ServiceType), serviceTypeRank(b.ServiceType))
		})
		return server
	}
	byKey := make(map[int][]Server)
	for _, input := range servers {
		server := prepare(input)
		key, special := groupingKey(server)
		if special {
			byKey[key] = append(byKey[key], server)
			continue
		}
		mask := serviceMask(server)
		byKey[mask] = append(byKey[mask], server)
	}

	keys := make([]int, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b int) int {
		return compareSpecialKeys(a, b, func(a, b int) int { return cmp.Compare(b, a) })
	})

	groups := make([]Group, 0, len(keys))
	for _, key := range keys {
		slices.SortFunc(byKey[key], compareServers)
		groups = append(groups, Group{Key: key, Name: groupName(key), Servers: byKey[key]})
	}
	return groups
}

func (exactServiceCount) Group(servers []Server) []Group {
	prepare := func(server Server) Server {
		server.Services = slices.Clone(server.Services)
		slices.SortStableFunc(server.Services, func(a, b Service) int {
			return cmp.Compare(serviceTypeRank(a.ServiceType), serviceTypeRank(b.ServiceType))
		})
		return server
	}
	byKey := make(map[int][]Server)
	for _, input := range servers {
		server := prepare(input)
		key, special := groupingKey(server)
		if special {
			byKey[key] = append(byKey[key], server)
			continue
		}
		mask := serviceMask(server)
		byKey[mask] = append(byKey[mask], server)
	}

	keys := make([]int, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b int) int {
		return compareSpecialKeys(a, b, func(a, b int) int {
			return cmp.Or(cmp.Compare(bits.OnesCount(uint(a)), bits.OnesCount(uint(b))), cmp.Compare(b, a))
		})
	})

	groups := make([]Group, 0, len(keys))
	for _, key := range keys {
		slices.SortFunc(byKey[key], compareServers)
		groups = append(groups, Group{Key: key, Name: groupName(key), Servers: byKey[key]})
	}
	return groups
}

func (exactPrimary) Group(servers []Server) []Group {
	prepare := func(server Server) Server {
		server.Services = slices.Clone(server.Services)
		slices.SortStableFunc(server.Services, func(a, b Service) int {
			return cmp.Compare(serviceTypeRank(a.ServiceType), serviceTypeRank(b.ServiceType))
		})
		return server
	}
	byKey := make(map[int][]Server)
	for _, input := range servers {
		server := prepare(input)
		key, special := groupingKey(server)
		if special {
			byKey[key] = append(byKey[key], server)
			continue
		}
		mask := serviceMask(server)
		byKey[mask] = append(byKey[mask], server)
	}

	keys := make([]int, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b int) int {
		return compareSpecialKeys(a, b, func(a, b int) int {
			primaryA, primaryB := primaryMask(a), primaryMask(b)
			return cmp.Or(
				cmp.Compare(primaryIndex(primaryA), primaryIndex(primaryB)),
				cmp.Compare(bits.OnesCount(uint(a)), bits.OnesCount(uint(b))),
				cmp.Compare(b, a),
			)
		})
	})

	groups := make([]Group, 0, len(keys))
	for _, key := range keys {
		slices.SortFunc(byKey[key], compareServers)
		groups = append(groups, Group{Key: key, Name: groupName(key), Servers: byKey[key]})
	}
	return groups
}

func (exactConvention) Group(servers []Server) []Group {
	prepare := func(server Server) Server {
		server.Services = slices.Clone(server.Services)
		slices.SortStableFunc(server.Services, func(a, b Service) int {
			return cmp.Compare(serviceTypeRank(a.ServiceType), serviceTypeRank(b.ServiceType))
		})
		return server
	}
	byKey := make(map[int][]Server)
	for _, input := range servers {
		server := prepare(input)
		key, special := groupingKey(server)
		if special {
			byKey[key] = append(byKey[key], server)
			continue
		}
		mask := serviceMask(server)
		byKey[mask] = append(byKey[mask], server)
	}

	keys := make([]int, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b int) int {
		return compareSpecialKeys(a, b, func(a, b int) int {
			rankA, rankB := slices.Index(conventionOrder, a), slices.Index(conventionOrder, b)
			if rankA >= 0 && rankB >= 0 {
				return cmp.Compare(rankA, rankB)
			}
			if rankA < 0 && rankB < 0 {
				return cmp.Compare(b, a)
			}
			if rankA < 0 {
				return 1
			}
			return -1
		})
	})

	groups := make([]Group, 0, len(keys))
	for _, key := range keys {
		slices.SortFunc(byKey[key], compareServers)
		groups = append(groups, Group{Key: key, Name: groupName(key), Servers: byKey[key]})
	}
	return groups
}

func (primaryService) Group(servers []Server) []Group {
	prepare := func(server Server) Server {
		server.Services = slices.Clone(server.Services)
		slices.SortStableFunc(server.Services, func(a, b Service) int {
			return cmp.Compare(serviceTypeRank(a.ServiceType), serviceTypeRank(b.ServiceType))
		})
		return server
	}
	byKey := make(map[int][]Server)
	for _, input := range servers {
		server := prepare(input)
		key, special := groupingKey(server)
		if special {
			byKey[key] = append(byKey[key], server)
			continue
		}
		mask := primaryMask(serviceMask(server))
		byKey[mask] = append(byKey[mask], server)
	}

	keys := make([]int, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b int) int {
		return compareSpecialKeys(a, b, func(a, b int) int { return cmp.Compare(b, a) })
	})

	groups := make([]Group, 0, len(keys))
	for _, key := range keys {
		slices.SortFunc(byKey[key], compareServers)
		groups = append(groups, Group{Key: key, Name: groupName(key), Servers: byKey[key]})
	}
	return groups
}

func groupingKey(server Server) (int, bool) {
	if len(server.Services) == 0 {
		return noServiceKey, true
	}
	for _, service := range server.Services {
		if serviceBits[service.ServiceType] == 0 {
			return unknownKey, true
		}
	}
	return 0, false
}

func serviceMask(server Server) int {
	mask := 0
	for _, service := range server.Services {
		mask |= serviceBits[service.ServiceType]
	}
	return mask
}

func primaryMask(mask int) int {
	for _, serviceType := range []ServiceType{AS, TS, SS, DS} {
		if bit := serviceBits[serviceType]; mask&bit != 0 {
			return bit
		}
	}
	return 0
}

func primaryIndex(mask int) int {
	for index, serviceType := range []ServiceType{AS, TS, SS, DS} {
		if mask&serviceBits[serviceType] != 0 {
			return index
		}
	}
	return 4
}

func serviceTypeRank(serviceType ServiceType) int {
	switch serviceType {
	case AS:
		return 0
	case TS:
		return 1
	case SS:
		return 2
	case DS:
		return 3
	default:
		return 4
	}
}

func roleRank(role Role) int {
	switch role {
	case Leader:
		return 0
	case Follower:
		return 1
	case LSS:
		return 2
	case RoleSS:
		return 3
	case RoleDS:
		return 4
	default:
		return 5
	}
}

func compareServers(a, b Server) int {
	roleOrder := func(a, b Role) int {
		return cmp.Or(cmp.Compare(roleRank(a), roleRank(b)), cmp.Compare(string(a), string(b)))
	}
	serviceOrder := func(a, b []Service) int {
		for index := 0; index < min(len(a), len(b)); index++ {
			if order := cmp.Or(
				cmp.Compare(serviceTypeRank(a[index].ServiceType), serviceTypeRank(b[index].ServiceType)),
				roleOrder(a[index].Role, b[index].Role),
			); order != 0 {
				return order
			}
		}
		return 0
	}
	stateRank := func(state State) int {
		switch state {
		case StateAbnormal:
			return 0
		case StateNormal:
			return 1
		default:
			return 2
		}
	}
	return cmp.Or(
		cmp.Compare(stateRank(a.State), stateRank(b.State)),
		cmp.Compare(string(a.State), string(b.State)),
		serviceOrder(a.Services, b.Services),
		cmp.Compare(len(a.Services), len(b.Services)),
		cmp.Compare(a.IP, b.IP),
		cmp.Compare(a.Name, b.Name),
	)
}

func compareSpecialKeys(a, b int, compareRegular func(int, int) int) int {
	if a == unknownKey && b != unknownKey && b != noServiceKey {
		return 1
	}
	if b == unknownKey && a != unknownKey && a != noServiceKey {
		return -1
	}
	if a == noServiceKey && b != noServiceKey {
		return 1
	}
	if b == noServiceKey && a != noServiceKey {
		return -1
	}
	return compareRegular(a, b)
}

func groupName(key int) string {
	switch key {
	case unknownKey:
		return "未知服务"
	case noServiceKey:
		return "无服务"
	case 0:
		return "键 0"
	default:
		var names []string
		for _, serviceType := range []ServiceType{AS, TS, SS, DS} {
			if key&serviceBits[serviceType] != 0 {
				names = append(names, string(serviceType))
			}
		}
		if len(names) == 0 {
			return fmt.Sprintf("键 %d", key)
		}
		return strings.Join(names, "+")
	}
}
