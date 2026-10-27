// Package purestrategy 提供“每条规则一个完整策略实现”的对照 POC。
package purestrategy

import (
	"fmt"
	"slices"

	"dashboardService/baseline"
)

// Group 是共享的分组结果类型。
type Group = baseline.Group

// Server 是共享的服务器领域类型。
type Server = baseline.Server

// Grouper 是每种分组规则必须实现的策略接口。
type Grouper interface {
	Group(servers []Server) []ClusterGroup
}

// Groupers 按规则类型注册五种独立的分组策略。
//
// 此 map 是 POC 注册表，调用方应只读；不要在运行期间修改。
var Groupers = map[baseline.GroupRule]Grouper{
	baseline.RuleExactCombination:  exactCombination{},
	baseline.RuleExactServiceCount: exactServiceCount{},
	baseline.RuleExactPrimary:      exactPrimary{},
	baseline.RuleExactConvention:   exactConvention{},
	baseline.RulePrimaryService:    primaryService{},
}

// RenderDashboard 通过规则注册表选择策略并运行。
// 未注册的规则表示调用方配置错误，会 panic。
func RenderDashboard(servers []Server, rule baseline.GroupRule) []ClusterGroup {
	grouper, ok := Groupers[rule]
	if !ok {
		panic(fmt.Sprintf("purestrategy.RenderDashboard: 无效的 GroupRule %d", rule))
	}
	return grouper.Group(servers)
}

func groupByCluster(servers []Server, group func([]Server) []Group) []ClusterGroup {
	clusterByID := make(map[int]Cluster)
	serversByCluster := make(map[int][]Server)
	for _, server := range servers {
		if _, exists := clusterByID[server.ClusterID]; !exists {
			clusterByID[server.ClusterID] = server.Cluster
		}
		serversByCluster[server.ClusterID] = append(serversByCluster[server.ClusterID], server)
	}

	clusters := make([]ClusterGroup, 0, len(clusterByID))
	for clusterID, cluster := range clusterByID {
		clusters = append(clusters, ClusterGroup{
			Cluster: cluster,
			Groups:  group(serversByCluster[clusterID]),
		})
	}
	slices.SortFunc(clusters, compareClusters)
	return clusters
}

func compareClusters(a, b ClusterGroup) int {
	if rankA, rankB := clusterRoleRank(a.ClusterRole), clusterRoleRank(b.ClusterRole); rankA != rankB {
		return rankA - rankB
	}
	if a.ClusterID < b.ClusterID {
		return -1
	}
	if a.ClusterID > b.ClusterID {
		return 1
	}
	return 0
}

func clusterRoleRank(role ClusterRole) int {
	switch role {
	case baseline.PrimaryClusterRole:
		return 0
	case baseline.SecondaryClusterRole:
		return 1
	case baseline.ArbiterClusterRole:
		return 2
	default:
		return 3
	}
}
