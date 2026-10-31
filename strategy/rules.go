package strategy

import (
	"slices"

	"dashboardService/baseline"
)

// 本文件是纯策略模式的核心，也是 POC 要展示的"重复代价"所在。
//
// 每个策略的 Group 方法都完整重复了相同的四步模板：
//  1. 集群分桶
//  2. 组内分桶（按 keyOf）
//  3. 组内排序 + 组间排序（仅此处不同）
//  4. 组装
//
// 五个策略中，唯一不同的是"组间排序"调用哪个比较函数
// （以及 primaryService 的 keyOf 不同）。
// 其余约 30 行模板完全重复——这就是纯策略模式在"骨架高度相似"
// 场景下的真实代价。POC 评审时应指出这一点。

// ============================================================================
// 策略 1：exactCombination（精确组合 + mask 降序）
// ============================================================================

type exactCombination struct{}

func (exactCombination) Name() string { return "rule-combination" }

func (exactCombination) Group(servers []Server) []ClusterGroup {
	// 1. 按集群分桶
	clusterByID := make(map[int]Cluster)
	serversByCluster := make(map[int][]Server)
	for _, srv := range servers {
		cid := srv.ClusterID
		clusterByID[cid] = srv.Cluster
		serversByCluster[cid] = append(serversByCluster[cid], srv)
	}

	// 2. 每个集群内分组
	clusters := make([]ClusterGroup, 0, len(clusterByID))
	for cid, cluster := range clusterByID {
		// 2.1 按 keyOf 分桶（本策略 keyOf = baseline.MaskOf）
		byKey := make(map[int][]Server)
		for _, srv := range serversByCluster[cid] {
			srv = prepareServer(srv)
			key := baseline.MaskOf(srv)
			byKey[key] = append(byKey[key], srv)
		}
		// 2.2 组内排序 + 收集 key
		keys := make([]int, 0, len(byKey))
		for k, group := range byKey {
			slices.SortFunc(group, compareServers)
			keys = append(keys, k)
		}
		// 2.3 组间排序（策略差异点：baseline.CmpByMask）
		slices.SortFunc(keys, baseline.CmpByMask)
		// 2.4 组装
		groups := make([]Group, len(keys))
		for i, k := range keys {
			groups[i] = Group{Key: k, Name: groupName(k), Servers: byKey[k]}
		}
		clusters = append(clusters, ClusterGroup{Cluster: cluster, Groups: groups})
	}

	// 3. 集群排序
	slices.SortFunc(clusters, compareClusters)
	return clusters
}

// ============================================================================
// 策略 2：exactServiceCount（精确组合 + 服务数量升序）
// ============================================================================

type exactServiceCount struct{}

func (exactServiceCount) Name() string { return "rule-service-count" }

func (exactServiceCount) Group(servers []Server) []ClusterGroup {
	// 1. 按集群分桶
	clusterByID := make(map[int]Cluster)
	serversByCluster := make(map[int][]Server)
	for _, srv := range servers {
		cid := srv.ClusterID
		clusterByID[cid] = srv.Cluster
		serversByCluster[cid] = append(serversByCluster[cid], srv)
	}

	// 2. 每个集群内分组
	clusters := make([]ClusterGroup, 0, len(clusterByID))
	for cid, cluster := range clusterByID {
		byKey := make(map[int][]Server)
		for _, srv := range serversByCluster[cid] {
			srv = prepareServer(srv)
			key := baseline.MaskOf(srv)
			byKey[key] = append(byKey[key], srv)
		}
		keys := make([]int, 0, len(byKey))
		for k, group := range byKey {
			slices.SortFunc(group, compareServers)
			keys = append(keys, k)
		}
		// 2.3 组间排序（策略差异点：baseline.CmpByServiceCount）
		slices.SortFunc(keys, baseline.CmpByServiceCount)
		groups := make([]Group, len(keys))
		for i, k := range keys {
			groups[i] = Group{Key: k, Name: groupName(k), Servers: byKey[k]}
		}
		clusters = append(clusters, ClusterGroup{Cluster: cluster, Groups: groups})
	}

	slices.SortFunc(clusters, compareClusters)
	return clusters
}

// ============================================================================
// 策略 3：exactPrimary（精确组合 + 主服务优先）
// ============================================================================

type exactPrimary struct{}

func (exactPrimary) Name() string { return "rule-primary" }

func (exactPrimary) Group(servers []Server) []ClusterGroup {
	clusterByID := make(map[int]Cluster)
	serversByCluster := make(map[int][]Server)
	for _, srv := range servers {
		cid := srv.ClusterID
		clusterByID[cid] = srv.Cluster
		serversByCluster[cid] = append(serversByCluster[cid], srv)
	}

	clusters := make([]ClusterGroup, 0, len(clusterByID))
	for cid, cluster := range clusterByID {
		byKey := make(map[int][]Server)
		for _, srv := range serversByCluster[cid] {
			srv = prepareServer(srv)
			key := baseline.MaskOf(srv)
			byKey[key] = append(byKey[key], srv)
		}
		keys := make([]int, 0, len(byKey))
		for k, group := range byKey {
			slices.SortFunc(group, compareServers)
			keys = append(keys, k)
		}
		// 2.3 组间排序（策略差异点：baseline.CmpByPrimary）
		slices.SortFunc(keys, baseline.CmpByPrimary)
		groups := make([]Group, len(keys))
		for i, k := range keys {
			groups[i] = Group{Key: k, Name: groupName(k), Servers: byKey[k]}
		}
		clusters = append(clusters, ClusterGroup{Cluster: cluster, Groups: groups})
	}

	slices.SortFunc(clusters, compareClusters)
	return clusters
}

// ============================================================================
// 策略 4：exactConvention（精确组合 + 业务约定顺序）
// ============================================================================

type exactConvention struct{}

func (exactConvention) Name() string { return "rule-convention" }

func (exactConvention) Group(servers []Server) []ClusterGroup {
	clusterByID := make(map[int]Cluster)
	serversByCluster := make(map[int][]Server)
	for _, srv := range servers {
		cid := srv.ClusterID
		clusterByID[cid] = srv.Cluster
		serversByCluster[cid] = append(serversByCluster[cid], srv)
	}

	clusters := make([]ClusterGroup, 0, len(clusterByID))
	for cid, cluster := range clusterByID {
		byKey := make(map[int][]Server)
		for _, srv := range serversByCluster[cid] {
			srv = prepareServer(srv)
			key := baseline.MaskOf(srv)
			byKey[key] = append(byKey[key], srv)
		}
		keys := make([]int, 0, len(byKey))
		for k, group := range byKey {
			slices.SortFunc(group, compareServers)
			keys = append(keys, k)
		}
		// 2.3 组间排序（策略差异点：baseline.CmpByConvention）
		slices.SortFunc(keys, baseline.CmpByConvention)
		groups := make([]Group, len(keys))
		for i, k := range keys {
			groups[i] = Group{Key: k, Name: groupName(k), Servers: byKey[k]}
		}
		clusters = append(clusters, ClusterGroup{Cluster: cluster, Groups: groups})
	}

	slices.SortFunc(clusters, compareClusters)
	return clusters
}

// ============================================================================
// 策略 5：primaryService（主服务归组 + mask 降序）
// ============================================================================

type primaryService struct{}

func (primaryService) Name() string { return "rule-primary-service" }

func (primaryService) Group(servers []Server) []ClusterGroup {
	clusterByID := make(map[int]Cluster)
	serversByCluster := make(map[int][]Server)
	for _, srv := range servers {
		cid := srv.ClusterID
		clusterByID[cid] = srv.Cluster
		serversByCluster[cid] = append(serversByCluster[cid], srv)
	}

	clusters := make([]ClusterGroup, 0, len(clusterByID))
	for cid, cluster := range clusterByID {
		byKey := make(map[int][]Server)
		for _, srv := range serversByCluster[cid] {
			srv = prepareServer(srv)
			// 2.1 分桶（策略差异点：keyOf = baseline.PrimaryServiceMask）
			key := baseline.PrimaryServiceMask(srv)
			byKey[key] = append(byKey[key], srv)
		}
		keys := make([]int, 0, len(byKey))
		for k, group := range byKey {
			slices.SortFunc(group, compareServers)
			keys = append(keys, k)
		}
		// 2.3 组间排序（与策略 1 相同：baseline.CmpByMask）
		slices.SortFunc(keys, baseline.CmpByMask)
		groups := make([]Group, len(keys))
		for i, k := range keys {
			groups[i] = Group{Key: k, Name: groupName(k), Servers: byKey[k]}
		}
		clusters = append(clusters, ClusterGroup{Cluster: cluster, Groups: groups})
	}

	slices.SortFunc(clusters, compareClusters)
	return clusters
}

// ============================================================================
// 包级入口：从字符串解析策略
// ============================================================================

// Parse 从 API 参数字符串解析策略。
func Parse(value string) (Strategy, error) {
	rule, err := baseline.ParseGroupRule(value)
	if err != nil {
		return nil, err
	}
	return ByRule(rule), nil
}
