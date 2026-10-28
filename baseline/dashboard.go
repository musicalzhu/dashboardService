package baseline

import (
	"cmp"
	"fmt"
	"slices"
)

// KeyFunc 根据服务器计算分组键。
type KeyFunc func(Server) int

// CompareFunc 按 slices.SortFunc 约定比较两个分组键。
type CompareFunc func(a, b int) int

// groupBy 是各分组规则共享的执行骨架。它先规范化服务器的 Services，
// 再将服务器分桶、排序各桶中的服务器、排序分组键，最后组装 Group 结果。
//
// 空 Services 和含未知服务类型的服务器分别进入专用分组，不调用 keyOf：
// 无服务组固定排最后，未知服务组排在普通分组之后、无服务组之前。
// 普通分组键必须非负；负键由 baseline 保留给这两个专用分组。
func groupBy(servers []Server, keyOf KeyFunc, compareKeys CompareFunc) []Group {
	byKey := make(map[int][]Server)
	for _, server := range servers {
		// 复制并整理 Services，确保后续排序使用稳定顺序且不修改调用方输入。
		server = prepareServer(server)

		// 无服务与未知服务拥有自己的组键，必须在计算业务分组键前识别。
		if len(server.Services) == 0 {
			byKey[noServiceGroupKey] = append(byKey[noServiceGroupKey], server)
			continue
		}

		if hasUnknownService(server) {
			byKey[unknownServiceGroupKey] = append(byKey[unknownServiceGroupKey], server)
			continue
		}

		// 只有含已知服务的服务器才参与当前规则的分组键计算。
		key := keyOf(server)
		if key < 0 {
			panic("baseline.groupBy: key function must return a non-negative key")
		}
		byKey[key] = append(byKey[key], server)
	}

	// 在每个桶内排序服务器，并收集实际出现的分组键。
	keys := make([]int, 0, len(byKey))
	for key, group := range byKey {
		slices.SortFunc(group, compareServers)
		keys = append(keys, key)
	}

	// 特殊组按配置的优先级排序；普通组之间的顺序完全委托给业务 comparator。
	// 特殊组优先级不同则不调用业务 comparator，避免保留键混入普通键排序。
	slices.SortFunc(keys, func(a, b int) int {
		if a == b {
			return 0
		}
		if order := cmp.Compare(specialGroupPriority[a], specialGroupPriority[b]); order != 0 {
			return order
		}
		return compareKeys(a, b)
	})

	// 按已经确定的键顺序构造结果；make 保证空输入也返回非 nil 的空切片。
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
// keyFn 必须返回非负键；负键由 baseline 保留给未知服务和无服务分组。
// 返回的集群及组顺序符合 Dashboard 展示顺序。
func GroupBy(servers []Server, keyFn KeyFunc, cmpFn CompareFunc) []ClusterGroup {
	if keyFn == nil {
		panic("baseline.GroupBy: nil key function")
	}
	if cmpFn == nil {
		panic("baseline.GroupBy: nil comparator")
	}
	return groupServersByCluster(servers, keyFn, cmpFn)
}

// groupServersByCluster 按集群 ID 聚合服务器，并在每个集群内调用 group 函数分组。
func groupServersByCluster(servers []Server, keyOf KeyFunc, cmp CompareFunc) []ClusterGroup {
	clusterByID := make(map[int]Cluster)
	serversByCluster := make(map[int][]Server)
	for _, server := range servers {
		clusterID := server.ClusterID
		if _, exists := clusterByID[clusterID]; !exists {
			clusterByID[clusterID] = server.Cluster
		}
		serversByCluster[clusterID] = append(serversByCluster[clusterID], server)
	}

	clusters := make([]ClusterGroup, 0, len(clusterByID))
	for clusterID, cluster := range clusterByID {
		clusters = append(clusters, ClusterGroup{
			Cluster: cluster,
			Groups:  groupBy(serversByCluster[clusterID], keyOf, cmp),
		})
	}
	slices.SortFunc(clusters, compareClusterGroups)
	return clusters
}

func compareClusterGroups(a, b ClusterGroup) int {
	return cmp.Or(
		cmp.Compare(priorityOf(clusterRolePriority, a.ClusterRole), priorityOf(clusterRolePriority, b.ClusterRole)),
		cmp.Compare(a.ClusterID, b.ClusterID),
	)
}

// RenderDashboard 按朴素 switch 选择规则，并返回前端所需的集群分组结构。
// 本方案的对比重点是 switch 分发；分组算法、比较器和数据结构保持与
// datadriven 一致。无效规则属于调用方编程错误，会 panic。
func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
	switch rule {
	case RuleCombination:
		return groupServersByCluster(servers, MaskOf, CompareByMaskDesc)
	case RuleServiceCount:
		return groupServersByCluster(servers, MaskOf, CmpByServiceCount)
	case RulePrimary:
		return groupServersByCluster(servers, MaskOf, CmpByPrimary)
	case RuleConvention:
		return groupServersByCluster(servers, MaskOf, CmpByConvention)
	case RulePrimaryService:
		return groupServersByCluster(servers, PrimaryServiceMask, CompareByMaskDesc)
	default:
		panic(fmt.Sprintf("RenderDashboard: 无效的 GroupRule %d", int(rule)))
	}
}
