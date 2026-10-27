package baseline

import (
	"cmp"
	"fmt"
	"math/bits"
	"slices"
	"strings"
)

var (
	// servicePriority 按服务展示顺序排列；未登记的服务类型排在已知类型之后。
	servicePriority = buildServicePriority()

	// statePriority 定义服务器状态顺序；其他状态值排在已知状态之后。
	statePriority = map[State]int{
		StateAbnormal: 0,
		StateNormal:   1,
	}

	// rolePriority 定义服务角色顺序；其他角色值排在已知角色之后。
	rolePriority = map[ServiceRole]int{
		Leader:   0,
		Follower: 1,
		LSS:      2,
		SSRole:   3,
		DSRole:   4,
	}

	// clusterRolePriority 定义集群角色顺序；其他角色值排在已知角色之后。
	clusterRolePriority = map[ClusterRole]int{
		PrimaryClusterRole:   0,
		SecondaryClusterRole: 1,
	}

	// specialGroupPriority 只登记特殊组；普通组键查表默认优先级为 0。
	specialGroupPriority = map[int]int{
		unknownServiceGroupKey: 1,
		noServiceGroupKey:      2,
	}
)

// allServiceTypes 是服务优先级及展示顺序的唯一来源。
var allServiceTypes = []ServiceType{ASType, TSType, SSType, DSType}

func buildServicePriority() map[ServiceType]int {
	priority := make(map[ServiceType]int, len(allServiceTypes))
	for i, serviceType := range allServiceTypes {
		priority[serviceType] = i
	}
	return priority
}

// priorityOf 返回配置优先级；未登记的值排在所有已登记值之后。
func priorityOf[T comparable](priorities map[T]int, value T) int {
	if priority, ok := priorities[value]; ok {
		return priority
	}
	return len(priorities)
}

// 高优先级服务使用高位，以支持 mask 数值降序。
const (
	asBit          = 0b1000
	tsBit          = 0b0100
	ssBit          = 0b0010
	dsBit          = 0b0001
	maxServiceMask = asBit | tsBit | ssBit | dsBit
)

var serviceBit = map[ServiceType]int{
	ASType: asBit,
	TSType: tsBit,
	SSType: ssBit,
	DSType: dsBit,
}

// comboOrderList 表示 RuleExactConvention 的 15 种组合顺序。
var comboOrderList = []int{
	asBit, asBit | ssBit, asBit | dsBit, asBit | ssBit | dsBit,
	asBit | tsBit, asBit | tsBit | ssBit, asBit | tsBit | dsBit,
	asBit | tsBit | ssBit | dsBit,
	tsBit, tsBit | ssBit, tsBit | dsBit, tsBit | ssBit | dsBit,
	ssBit, ssBit | dsBit,
	dsBit,
}

// primaryIndex 返回 mask 中最高优先级服务的索引；空 mask 返回服务类型数。
func primaryIndex(mask int) int {
	for i, serviceType := range allServiceTypes {
		if mask&serviceBit[serviceType] != 0 {
			return i
		}
	}
	return len(allServiceTypes)
}

// primaryMask 返回 mask 中最高优先级服务的单个 bit；空 mask 返回 0。
func primaryMask(mask int) int {
	for _, serviceType := range allServiceTypes {
		if bit := serviceBit[serviceType]; mask&bit != 0 {
			return bit
		}
	}
	return 0
}

// comboName 以 AS→TS→SS→DS 的顺序生成组合名称。
func comboName(mask int) string {
	var names []string
	for _, serviceType := range allServiceTypes {
		if mask&serviceBit[serviceType] != 0 {
			names = append(names, string(serviceType))
		}
	}
	return strings.Join(names, "+")
}

func groupName(key int) string {
	if key == 0 {
		return "键 0"
	}
	if key <= maxServiceMask {
		return comboName(key)
	}
	return fmt.Sprintf("键 %d", key)
}

// maskOf 将服务器服务类型合并成完整组合 mask；重复类型自然去重。
func maskOf(server Server) int {
	mask := 0
	for _, service := range server.Services {
		mask |= serviceBit[service.ServiceType]
	}
	return mask
}

// MaskOf 返回服务器中已识别服务类型对应的 bit mask。
func MaskOf(server Server) int {
	return maskOf(server)
}

// hasUnknownService 检查服务器是否包含未识别的服务类型。
// 已知和未知类型混合时，整台服务器归入未知服务组，不按已知部分分组。
func hasUnknownService(server Server) bool {
	for _, service := range server.Services {
		if serviceBit[service.ServiceType] == 0 {
			return true
		}
	}
	return false
}

// primaryServiceMask 将完整组合投影为最高优先级服务对应的单个 bit。
func primaryServiceMask(server Server) int {
	return primaryMask(maskOf(server))
}

// PrimaryServiceMask 返回服务器最高优先级的已知服务对应的 bit；没有已知服务时返回 0。
func PrimaryServiceMask(server Server) int {
	return primaryServiceMask(server)
}

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
