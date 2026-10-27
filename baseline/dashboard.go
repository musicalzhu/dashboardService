package baseline

import (
	"fmt"
	"slices"
)

// KeyFunc 根据服务器计算分组键。
type KeyFunc func(Server) int

// CompareFunc 按 slices.SortFunc 约定比较两个分组键。
type CompareFunc func(a, b int) int

// groupBy 是五种规则共享的执行骨架：准备输入、按 key 分桶、组内排序、
// 按 cmp 排组，并组装返回结果。含未知类型的服务器进入独立未知服务组，
// 该组排在已知组合后、无服务组前；空 Services 的服务器进入独立无服务组并固定排最后。
func groupBy(servers []Server, keyOf func(Server) int, cmp func(a, b int) int) []Group {
	byKey := make(map[int][]Server)
	for _, server := range servers {
		server = prepareServer(server)
		if len(server.Services) == 0 {
			byKey[noServiceGroupKey] = append(byKey[noServiceGroupKey], server)
			continue
		}

		if hasUnknownService(server) {
			byKey[unknownServiceGroupKey] = append(byKey[unknownServiceGroupKey], server)
			continue
		}

		key := keyOf(server)
		if key < 0 {
			panic("baseline.groupBy: key function must return a non-negative key")
		}
		byKey[key] = append(byKey[key], server)
	}

	keys := make([]int, 0, len(byKey))
	for key, group := range byKey {
		slices.SortFunc(group, compareServers)
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(a, b int) int {
		if a == unknownServiceGroupKey && b != unknownServiceGroupKey && b != noServiceGroupKey {
			return 1
		}
		if b == unknownServiceGroupKey && a != unknownServiceGroupKey && a != noServiceGroupKey {
			return -1
		}
		if a == noServiceGroupKey && b != noServiceGroupKey {
			return 1
		}
		if b == noServiceGroupKey && a != noServiceGroupKey {
			return -1
		}
		return cmp(a, b)
	})

	groups := make([]Group, len(keys))
	for i, key := range keys {
		name := groupName(key)
		switch key {
		case unknownServiceGroupKey:
			name = unknownServiceGroupName
		case noServiceGroupKey:
			name = noServiceGroupName
		}
		groups[i] = Group{Key: key, Name: name, Servers: byKey[key]}
	}
	return groups
}

// GroupBy 按 ClusterID 聚合服务器，并在每个集群内使用指定键函数和比较器分组。
// keyFn 必须返回非负键；0 是有效的普通分组键。负键由 baseline 保留给未知服务和无服务分组。
// 返回的集群及组顺序符合 Dashboard 展示顺序。
func GroupBy(servers []Server, keyOf KeyFunc, cmp CompareFunc) []ClusterGroup {
	if keyOf == nil {
		panic("baseline.GroupBy: nil key function")
	}
	if cmp == nil {
		panic("baseline.GroupBy: nil comparator")
	}
	return groupServersByCluster(servers, func(clusterServers []Server) []Group {
		return groupBy(clusterServers, keyOf, cmp)
	})
}

// RenderDashboard 按朴素 switch 选择规则，并返回前端所需的集群分组结构。
//
// 本方案的对比重点是 switch 分发；分组算法、比较器和数据结构保持与
// datadriven 一致。无效规则属于调用方编程错误，会 panic。
func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
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
