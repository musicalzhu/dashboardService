package baseline

import (
	"fmt"
	"slices"
)

// groupBy 是各分组规则共享的执行骨架。它先规范化服务器的 Services，
// 再将服务器分桶、排序各桶中的服务器、排序分组键，最后组装 Group 结果。
func groupBy(servers []Server, r RuleImpl) []Group {
	// 步骤1 遍历所有服务器，计算分组键并分桶。
	byKey := make(map[int][]Server)
	for _, server := range servers {
		// 复制并整理 Services，确保后续排序使用稳定顺序且不修改调用方输入。
		server = prepareServer(server)

		// 步骤1.1 根据注入的分组算法，计算当前服务器的分组键。
		key := r.KeyOf(server)

		// 步骤1.2 将当前服务器加入对应的分组桶。
		byKey[key] = append(byKey[key], server)
	}

	// 步骤2 在每个桶内排序服务器，并收集实际出现的分组键。
	keys := make([]int, 0, len(byKey))
	for key, group := range byKey {
		// 步骤2.1 对每个分组桶内的服务器进行排序。
		slices.SortFunc(group, compareServers)

		// 步骤2.2 收集所有实际出现的分组键。
		keys = append(keys, key)
	}

	// 步骤3
	slices.SortFunc(keys, r.Cmp)

	// 步骤 4 按已经确定的键顺序构造结果。
	groups := make([]Group, len(keys))
	for i, key := range keys {
		groups[i] = Group{
			Key:     key,
			Name:    groupName(key),
			Servers: byKey[key],
		}
	}
	return groups
}

// GroupServersByCluster 按集群 ID 聚合服务器，并在每个集群内使用指定键函数和比较器分组。
func GroupServersByCluster(servers []Server, r RuleImpl) []ClusterGroup {
	clusterByID := make(map[int]Cluster)
	serversByCluster := make(map[int][]Server)
	// 步骤1 遍历所有服务器，按集群 ID 分桶。
	for _, server := range servers {
		clusterID := server.ClusterID
		if _, exists := clusterByID[clusterID]; !exists {
			clusterByID[clusterID] = server.Cluster
		}
		serversByCluster[clusterID] = append(serversByCluster[clusterID], server)
	}

	clusters := make([]ClusterGroup, 0, len(clusterByID))
	// 步骤2 对每个集群内的服务器进行分组。
	for clusterID, cluster := range clusterByID {
		clusters = append(clusters, ClusterGroup{
			Cluster: cluster,
			Groups:  groupBy(serversByCluster[clusterID], r),
		})
	}
	// 步骤3 对集群分组进行排序。
	slices.SortFunc(clusters, compareClusters)
	return clusters
}

// RenderDashboard 按朴素 switch 选择规则，并返回前端所需的集群分组结构。
// 本方案的对比重点是 switch 分发；分组算法、比较器和数据结构保持与
// datadriven 一致。无效规则属于调用方编程错误，会 panic。
func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
	switch rule {
	case RuleCombination:
		return GroupServersByCluster(servers, RuleImpl{KeyOf: MaskOf, Cmp: CmpByMask})
	case RuleServiceCount:
		return GroupServersByCluster(servers, RuleImpl{KeyOf: MaskOf, Cmp: CmpByServiceCount})
	case RulePrimary:
		return GroupServersByCluster(servers, RuleImpl{KeyOf: MaskOf, Cmp: CmpByPrimary})
	case RuleConvention:
		return GroupServersByCluster(servers, RuleImpl{KeyOf: MaskOf, Cmp: CmpByConvention})
	case RulePrimaryService:
		return GroupServersByCluster(servers, RuleImpl{KeyOf: PrimaryServiceMask, Cmp: CmpByMask})
	default:
		panic(fmt.Sprintf("RenderDashboard: 无效的 GroupRule %d", int(rule)))
	}
}
