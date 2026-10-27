package baseline

import (
	"fmt"
	"strings"
)

// allServiceTypes 是服务优先级及展示顺序的唯一来源。
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

// serviceTypeRank 返回服务类型的展示优先级；未知类型排在已知类型之后。
func serviceTypeRank(serviceType ServiceType) int {
	for i, known := range allServiceTypes {
		if known == serviceType {
			return i
		}
	}
	return len(allServiceTypes)
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
