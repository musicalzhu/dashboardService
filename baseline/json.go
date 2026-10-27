package baseline

import (
	"encoding/json"
	"fmt"
	"slices"
)

// JSONFormat 选择 Dashboard JSON 的输出结构。
type JSONFormat int

const (
	// JSONNested 将集群及其分组嵌套输出。
	JSONNested JSONFormat = iota
	// JSONFlat 将每个集群/分组展开为独立条目。
	JSONFlat
)

type flatGroupEntry struct {
	Name    string   `json:"group"`
	Servers []Server `json:"servers"`
}

type flatClusterRow struct {
	Cluster
	Name    string   `json:"group"`
	Servers []Server `json:"servers"`
}

// MarshalGroups 将有序分组编码为缩进 JSON。
func MarshalGroups(groups []Group) ([]byte, error) {
	return json.MarshalIndent(groups, "", "  ")
}

// MarshalFlatGroups 将有序分组展开为平铺条目并编码为缩进 JSON。
func MarshalFlatGroups(groups []Group) ([]byte, error) {
	flat := make([]flatGroupEntry, 0, len(groups))
	for _, group := range groups {
		flat = append(flat, flatGroupEntry{
			Name:    group.Name,
			Servers: group.Servers,
		})
	}
	return json.MarshalIndent(flat, "", "  ")
}

// MarshalDashboard 按指定分组规则生成结果，并编码为嵌套或平铺 UI 结构。
func MarshalDashboard(servers []Server, rule GroupRule, format JSONFormat) ([]byte, error) {
	groups := RenderDashboard(servers, rule)
	switch format {
	case JSONNested:
		return MarshalGroups(groups)
	case JSONFlat:
		return MarshalFlatGroups(groups)
	default:
		return nil, fmt.Errorf("MarshalDashboard: 无效的 JSONFormat %d", format)
	}
}

// MarshalClusterDashboard 按集群聚合服务器，在每个集群内应用所选规则，
// 并编码为嵌套分组或平铺行。
func MarshalClusterDashboard(servers []Server, rule GroupRule, format JSONFormat) ([]byte, error) {
	clusters := RenderClusterDashboardByRule(servers, rule)
	switch format {
	case JSONNested:
		return json.MarshalIndent(clusters, "", "  ")
	case JSONFlat:
		return json.MarshalIndent(flattenClusterGroups(clusters), "", "  ")
	default:
		return nil, fmt.Errorf("MarshalClusterDashboard: 无效的 JSONFormat %d", format)
	}
}

// RenderClusterDashboard 按主服务规则生成集群分组，保留原有调用入口。
func RenderClusterDashboard(servers []Server) []ClusterGroup {
	return RenderClusterDashboardByRule(servers, RulePrimaryService)
}

// RenderClusterDashboardByRule 按 ClusterID 聚合服务器，并在每个集群中应用指定规则。
// 集群元数据取相同 ClusterID 在输入中首次出现的值。
func RenderClusterDashboardByRule(servers []Server, rule GroupRule) []ClusterGroup {
	return groupServersByCluster(servers, func(clusterServers []Server) []Group {
		return RenderDashboard(clusterServers, rule)
	})
}

func groupServersByCluster(servers []Server, group func([]Server) []Group) []ClusterGroup {
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
			Groups:  group(serversByCluster[clusterID]),
		})
	}
	slices.SortFunc(clusters, compareClusterGroups)
	return clusters
}

func flattenClusterGroups(clusters []ClusterGroup) []flatClusterRow {
	flat := make([]flatClusterRow, 0)
	for _, cluster := range clusters {
		for _, group := range cluster.Groups {
			flat = append(flat, flatClusterRow{
				Cluster: cluster.Cluster,
				Name:    group.Name,
				Servers: group.Servers,
			})
		}
	}
	return flat
}

func compareClusterGroups(a, b ClusterGroup) int {
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
	case PrimaryClusterRole:
		return 0
	case SecondaryClusterRole:
		return 1
	case ArbiterClusterRole:
		return 2
	default:
		return 3
	}
}
