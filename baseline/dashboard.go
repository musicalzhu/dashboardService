package baseline

import (
	"fmt"
	"slices"
)

// Group 表示一个服务分组。Key 仅供程序内部使用，不写入 JSON。
type Group struct {
	Key     int      `json:"-"`
	Name    string   `json:"name"`
	Servers []Server `json:"servers"`
}

const noServiceGroupName = "无服务"
const unknownServiceGroupName = "未知服务"
const unknownServiceGroupKey = -1

// KeyFunc 根据服务器计算分组键。
type KeyFunc func(Server) int

// CompareFunc 按 slices.SortFunc 约定比较两个分组键。
type CompareFunc func(a, b int) int

// groupBy 是五种规则共享的执行骨架：准备输入、按 key 分桶、组内排序、
// 按 cmp 排组，并组装返回结果。含未知类型的服务器进入 key=-1 的未知服务组，
// 该组排在已知组合后、无服务组前；空 Services 的服务器归入 key=0 并固定排最后。
func groupBy(servers []Server, keyOf func(Server) int, cmp func(a, b int) int) []Group {
	byKey := make(map[int][]Server)
	for _, server := range servers {
		server = prepareServer(server)
		if len(server.Services) == 0 {
			byKey[0] = append(byKey[0], server)
			continue
		}

		if hasUnknownService(server) {
			byKey[unknownServiceGroupKey] = append(byKey[unknownServiceGroupKey], server)
			continue
		}

		key := keyOf(server)
		if key == 0 {
			continue
		}
		byKey[key] = append(byKey[key], server)
	}

	keys := make([]int, 0, len(byKey))
	for key, group := range byKey {
		slices.SortFunc(group, compareServers)
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b int) int {
		if a == unknownServiceGroupKey && b != unknownServiceGroupKey && b != 0 {
			return 1
		}
		if b == unknownServiceGroupKey && a != unknownServiceGroupKey && a != 0 {
			return -1
		}
		if a == 0 && b != 0 {
			return 1
		}
		if b == 0 && a != 0 {
			return -1
		}
		return cmp(a, b)
	})

	groups := make([]Group, len(keys))
	for i, key := range keys {
		name := comboName(key)
		if key == unknownServiceGroupKey {
			name = unknownServiceGroupName
		} else if key == 0 {
			name = noServiceGroupName
		}
		groups[i] = Group{Key: key, Name: name, Servers: byKey[key]}
	}
	return groups
}

// GroupBy 使用指定的键函数和组间比较器执行共享分组流程。
// 该函数供不同 POC 复用算法，以便只比较规则选择方式。
func GroupBy(servers []Server, keyOf KeyFunc, cmp CompareFunc) []Group {
	if keyOf == nil {
		panic("baseline.GroupBy: nil key function")
	}
	if cmp == nil {
		panic("baseline.GroupBy: nil comparator")
	}
	return groupBy(servers, keyOf, cmp)
}

// RenderDashboard 按朴素 switch 选择分组键与组间排序函数，再调用共享骨架。
//
// 本方案的对比重点是 switch 分发；分组算法、比较器和数据结构保持与
// datadriven 一致。无效规则属于调用方编程错误，会 panic。
func RenderDashboard(servers []Server, rule GroupRule) []Group {
	switch rule {
	case RuleExactCombination:
		return GroupBy(servers, MaskOf, CompareByMaskDesc)
	case RuleExactServiceCount:
		return GroupBy(servers, MaskOf, CompareByServiceCount)
	case RuleExactPrimary:
		return GroupBy(servers, MaskOf, CompareByPrimary)
	case RuleExactConvention:
		return GroupBy(servers, MaskOf, CompareByConvention)
	case RulePrimaryService:
		return GroupBy(servers, PrimaryServiceMask, CompareByMaskDesc)
	default:
		panic(fmt.Sprintf("RenderDashboard: 无效的 GroupRule %d", int(rule)))
	}
}
