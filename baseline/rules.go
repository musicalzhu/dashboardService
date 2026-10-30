package baseline

import (
	"cmp"
	"fmt"
	"math/bits"
	"slices"
	"strings"
)

var (
	// servicePriority 定义服务类型优先级；值越小优先级越高。
	servicePriority = map[ServiceType]int{
		ASType: 0,
		TSType: 1,
		SSType: 2,
		DSType: 3,
	}

	// statePriority 定义服务器状态顺序；值越小优先级越高，其他状态值排在已知状态之后。
	statePriority = map[State]int{
		StateAbnormal: 0,
		StateNormal:   1,
	}

	// rolePriority 定义服务角色顺序；值越小优先级越高，其他角色值排在已知角色之后。
	rolePriority = map[ServiceRole]int{
		Leader:   0,
		Follower: 1,
		LSS:      2,
		SSRole:   3,
		DSRole:   4,
	}

	// clusterRolePriority 定义集群角色顺序；值越小优先级越高，其他角色值排在已知角色之后。
	clusterRolePriority = map[ClusterRole]int{
		PrimaryClusterRole:   0,
		SecondaryClusterRole: 1,
	}
)

// allServiceTypes 定义服务类型的稳定遍历和展示顺序。
var allServiceTypes = []ServiceType{ASType, TSType, SSType, DSType}

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

// priorityOf 返回配置优先级；未登记的值排在所有已登记值之后。
// 举例：servicePriority 中 ASType=0、TSType=1、SSType=2、DSType=3，未登记类型优先级为 4。
// 举例：statePriority 中 StateAbnormal=0、StateNormal=1，未登记状态优先级为 2。
func priorityOf[T comparable](priorities map[T]int, value T) int {
	if priority, ok := priorities[value]; ok {
		return priority
	}
	return len(priorities)
}

// primaryIndex 返回 mask 中最高优先级服务的索引；空 mask 返回服务类型数。
// 举例：mask 0b1101 对应 AS+TS+DS，返回 0（AS 的索引）；mask 0b0000 返回 4（服务类型数）。
func primaryIndex(mask int) int {
	for i, serviceType := range allServiceTypes {
		if mask&serviceBit[serviceType] != 0 {
			return i
		}
	}
	return len(allServiceTypes)
}

// primaryMask 返回 mask 中最高优先级服务的单个 bit；空 mask 返回 0。
// 举例：mask 0b1101 对应 AS+TS+DS，返回 0b1000（AS 的 bit）；mask 0b0000 返回 0。
func primaryMask(mask int) int {
	for _, serviceType := range allServiceTypes {
		if bit := serviceBit[serviceType]; mask&bit != 0 {
			return bit
		}
	}
	return 0
}

// comboName 以 AS→TS→SS→DS 的顺序生成组合名称。
// 举例：mask 0b1101 对应 "AS+TS+DS"；mask 0b0000 对应 ""。
func comboName(mask int) string {
	var names []string
	for _, serviceType := range allServiceTypes {
		if mask&serviceBit[serviceType] != 0 {
			names = append(names, string(serviceType))
		}
	}
	return strings.Join(names, "+")
}

// groupName 返回给定 key 对应的组名称。
// 举例：key 为 unknownServiceGroupKey 返回 "未知服务组"；key 为 noServiceGroupKey 返回 "无服务组"；key 为 0 返回 "键 0"；key 为 0b1101 返回 "AS+TS+DS"。
// 返回的组名称用于在 UI 中显示，便于用户识别不同服务组合。
func groupName(key int) string {
	switch key {
	case unknownServiceGroupKey:
		return unknownServiceGroupName
	case noServiceGroupKey:
		return noServiceGroupName
	}
	if key <= maxServiceMask {
		return comboName(key)
	}
	return fmt.Sprintf("键 %d", key)
}

// maskOf 将服务器服务类型合并成完整组合 mask；重复类型自然去重。
// 举例：服务器含 AS+TS+DS，返回 0b1101；
func MaskOf(server Server) int {
	mask := 0

	// 无服务与未知服务拥有自己的组键，必须在计算服务分组键前识别。
	if len(server.Services) == 0 {
		return noServiceGroupKey
	}

	if hasUnknownService(server) {
		return unknownServiceGroupKey
	}

	// 服务类型合并成完整组合 mask
	for _, service := range server.Services {
		mask |= serviceBit[service.ServiceType]
	}
	return mask
}

// PrimaryServiceMask 返回服务器最高优先级的已知服务对应的 bit；没有已知服务时返回 0。
// 举例：服务器含 AS+TS+DS，返回 0b1000（AS 的 bit）；服务器无服务，返回 0b0000。
func PrimaryServiceMask(server Server) int {
	mask := MaskOf(server)
	return primaryMask(mask)
}

// hasUnknownService 检查服务器是否包含未识别的服务类型。
// 已知和未知类型混合时，整台服务器归入未知服务组，不按已知部分分组。
// 举例：服务器含 AS+TS+XX未知类型，返回 true；服务器仅含 AS+TS，返回 false。
func hasUnknownService(server Server) bool {
	for _, service := range server.Services {
		if serviceBit[service.ServiceType] == 0 {
			return true
		}
	}
	return false
}

// prepareServer 克隆 Services 后按服务优先级排序，避免修改调用方切片。
// 举例：服务器含 TS+AS，返回 AS+TS；服务器含 DS+SS+TS，返回 TS+SS+DS。
func prepareServer(server Server) Server {
	server.Services = slices.Clone(server.Services)
	slices.SortStableFunc(server.Services, func(a, b Service) int {
		return cmp.Compare(priorityOf(servicePriority, a.ServiceType), priorityOf(servicePriority, b.ServiceType))
	})
	return server
}

func compareClusters(a, b ClusterGroup) int {
	return cmp.Or(
		cmp.Compare(priorityOf(clusterRolePriority, a.ClusterRole), priorityOf(clusterRolePriority, b.ClusterRole)),
		cmp.Compare(a.ClusterID, b.ClusterID),
	)
}

// compareServers 服务器之间排序，按状态、服务类型/角色序列、IP、Name 排序。
// 举例：状态 Abnormal 排在 Normal 前；服务 AS+TS 排在 AS+DS 前；IP 192.168.1.1 排在 192.168.1.2 前。
func compareServers(a, b Server) int {
	return cmp.Or(
		cmp.Compare(priorityOf(statePriority, a.State), priorityOf(statePriority, b.State)),
		compareServerServices(a.Services, b.Services),
		cmp.Compare(a.IP, b.IP),
		cmp.Compare(a.Name, b.Name),
	)
}

// compareServerServices 按序比较服务器的服务类型和角色。
// 举例：服务 AS+TS 排在 AS+DS 前；服务 TS+SS 排在 TS+DS 前；服务 AS+TS 与 AS+TS 相等。
func compareServerServices(a, b []Service) int {
	// 逐个比较服务类型和角色，直到发现不同或比较完毕。
	for i := 0; i < min(len(a), len(b)); i++ {
		serviceA, serviceB := a[i], b[i]
		if c := cmp.Or(
			cmp.Compare(priorityOf(servicePriority, serviceA.ServiceType), priorityOf(servicePriority, serviceB.ServiceType)),
			cmp.Compare(priorityOf(rolePriority, serviceA.Role), priorityOf(rolePriority, serviceB.Role)),
		); c != 0 {
			return c
		}
	}
	return 0
}

// 服务类型分组算法
// CmpByMask 按 mask 降序比较分组键。
// 举例：mask 0b1101（AS+TS+DS）排在 mask 0b1001（AS+DS）之后；mask 0b0011（SS+DS）排在 mask 0b0001（DS）之后。
func CmpByMask(a, b int) int {
	return cmp.Compare(b, a)
}

// 服务类型分组算法
// CmpByServiceCount 按服务数量升序，再按 mask 降序排列。
// 举例：mask 0b1101（AS+TS+DS）排在 mask 0b1001（AS+DS）之后；mask 0b0011（SS+DS）排在 mask 0b0001（DS）之后。
func CmpByServiceCount(a, b int) int {
	return cmp.Or(
		cmp.Compare(bits.OnesCount(uint(a)), bits.OnesCount(uint(b))),
		cmp.Compare(b, a),
	)
}

// 服务类型分组算法
// CmpByPrimary 按主服务索引、服务数量、mask 依次排序。
// 举例：mask 0b1101（AS+TS+DS）排在 mask 0b1001（AS+DS）之后；mask 0b0011（SS+DS）排在 mask 0b0001（DS）之后。
func CmpByPrimary(a, b int) int {
	return cmp.Or(
		cmp.Compare(primaryIndex(a), primaryIndex(b)),
		cmp.Compare(bits.OnesCount(uint(a)), bits.OnesCount(uint(b))),
		cmp.Compare(b, a),
	)
}

// 服务类型分组算法
// CmpByConvention 按业务顺序表排序；未登记组合稳定地排在登记组合之后。
// 举例：mask 0b1101（AS+TS+DS）排在 mask 0b1001（AS+DS）之后；mask 0b0011（SS+DS）排在 mask 0b0001（DS）之后。
func CmpByConvention(a, b int) int {
	rankA := slices.Index(comboOrderList, a)
	rankB := slices.Index(comboOrderList, b)
	switch {
	case rankA >= 0 && rankB >= 0:
		return cmp.Compare(rankA, rankB)
	case rankA < 0 && rankB < 0:
		return cmp.Compare(b, a)
	case rankA < 0:
		return 1
	default:
		return -1
	}
}

const (
	// RuleCombination 按完整组合 mask 降序排列。
	// 举例：服务组合 AS+TS+DS 排在 AS+DS 之后，AS+DS 排在 SS+DS 之后。
	// RuleCombination 是默认规则，适合大多数场景。
	RuleCombination GroupRule = iota

	// RuleServiceCount 按服务数升序、再按 mask 降序排列。
	// 举例：服务组合 AS+DS 排在 AS+TS+DS 之前，AS+TS+DS 排在 SS+DS 之前。
	// RuleServiceCount 适合关注服务数的场景。
	RuleServiceCount

	// RulePrimary 按主服务优先级、服务数、mask 排列。
	// 举例：服务组合 AS+DS 排在 AS+TS+DS 之前，AS+TS+DS 排在 SS+DS 之前。
	// RulePrimary 适合关注主服务的场景。
	RulePrimary

	// RuleConvention 按 comboOrderList 中的业务顺序排列。
	// 举例：服务组合 AS+DS 排在 AS+TS+DS 之前，AS+TS+DS 排在 SS+DS 之前。
	// RuleConvention 适合关注业务顺序的场景。
	RuleConvention

	// RulePrimaryService 按最高优先级服务归组。
	// 举例：服务组合 AS+TS+DS 与 AS+DS 归为 AS 组。
	// RulePrimaryService 适合关注主服务的场景。
	RulePrimaryService
)

// String 返回规则的稳定名称。
func (r GroupRule) String() string {
	names := [...]string{
		"rule-combination",
		"rule-service-count",
		"rule-primary",
		"rule-convention",
		"rule-primary-service",
	}
	if r < 0 || int(r) >= len(names) {
		return fmt.Sprintf("GroupRule(%d)", r)
	}
	return names[r]
}

// ParseGroupRule 将配置或请求中的名称解析为规则。
// 空字符串使用 RuleCombination；未知名称返回错误，不静默回退。
func ParseGroupRule(value string) (GroupRule, error) {
	switch value {
	case "", "rule-combination":
		return RuleCombination, nil
	case "rule-service-count":
		return RuleServiceCount, nil
	case "rule-primary":
		return RulePrimary, nil
	case "rule-convention":
		return RuleConvention, nil
	case "rule-primary-service":
		return RulePrimaryService, nil
	default:
		return 0, fmt.Errorf("未知分组规则: %q", value)
	}
}
