package baseline

import (
	"cmp"
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
	clusters := RenderDashboard(servers, rule)
	switch format {
	case JSONNested:
		return json.MarshalIndent(clusters, "", "  ")
	case JSONFlat:
		return json.MarshalIndent(flattenClusterGroups(clusters), "", "  ")
	default:
		return nil, fmt.Errorf("MarshalDashboard: 无效的 JSONFormat %d", format)
	}
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
	return cmp.Or(
		cmp.Compare(priorityOf(clusterRolePriority, a.ClusterRole), priorityOf(clusterRolePriority, b.ClusterRole)),
		cmp.Compare(a.ClusterID, b.ClusterID),
	)
}
