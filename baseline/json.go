package baseline

import (
	"encoding/json"
	"fmt"
	"slices"
)

// JSONFormat 选择 Dashboard JSON 的输出结构。
type JSONFormat int

const (
	// JSONNested 将分组作为包含服务器列表的嵌套对象输出。
	JSONNested JSONFormat = iota
	// JSONFlat 将分组转换为按规则排序的条目，不输出内部 Group.Key。
	JSONFlat
)

// MarshalGroups 将有序分组编码为缩进 JSON。
func MarshalGroups(groups []Group) ([]byte, error) {
	return json.MarshalIndent(groups, "", "  ")
}

// FlattenGroups 将分组转换为不含内部键的有序 UI 条目。
func FlattenGroups(groups []Group) []FlatGroup {
	flat := make([]FlatGroup, 0, len(groups))
	for _, group := range groups {
		flat = append(flat, FlatGroup{
			Name:    group.Name,
			Servers: group.Servers,
		})
	}
	return flat
}

// MarshalFlatGroups 将有序平铺分组编码为缩进 JSON。
func MarshalFlatGroups(groups []Group) ([]byte, error) {
	return json.MarshalIndent(FlattenGroups(groups), "", "  ")
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
	clusters := renderClusterDashboard(servers, rule)
	switch format {
	case JSONNested:
		return json.MarshalIndent(clusters, "", "  ")
	case JSONFlat:
		return json.MarshalIndent(flattenClusterDashboard(clusters), "", "  ")
	default:
		return nil, fmt.Errorf("MarshalClusterDashboard: 无效的 JSONFormat %d", format)
	}
}

// RenderClusterDashboard 是固定使用主服务规则的兼容视图，返回历史嵌套结构。
func RenderClusterDashboard(servers []Server) []ClusterGroup {
	rendered := renderClusterDashboard(servers, RulePrimaryService)
	groups := make([]ClusterGroup, 0, len(rendered))
	for _, cluster := range rendered {
		group := ClusterGroup{
			Cluster:       cluster.Cluster,
			ServiceGroups: make([]ServiceGroup, 0, len(cluster.Groups)),
		}
		for _, flatGroup := range cluster.Groups {
			serviceType := ServiceType(flatGroup.Name)
			switch flatGroup.Name {
			case unknownServiceGroupName:
				serviceType = UnknownServiceType
			case noServiceGroupName:
				serviceType = NoServiceType
			}
			group.ServiceGroups = append(group.ServiceGroups, ServiceGroup{
				ServiceType: serviceType,
				Servers:     flatGroup.Servers,
			})
		}
		groups = append(groups, group)
	}
	return groups
}

// RenderClusterFlatDashboard 是固定使用主服务规则的兼容视图，返回历史固定字段结构。
func RenderClusterFlatDashboard(servers []Server) []ClusterFlatGroup {
	clusterGroups := RenderClusterDashboard(servers)
	flatGroups := make([]ClusterFlatGroup, 0, len(clusterGroups))
	for _, cluster := range clusterGroups {
		flat := ClusterFlatGroup{Cluster: cluster.Cluster}
		for _, group := range cluster.ServiceGroups {
			switch group.ServiceType {
			case AS:
				flat.AS = group.Servers
			case TS:
				flat.TS = group.Servers
			case SS:
				flat.SS = group.Servers
			case DS:
				flat.DS = group.Servers
			case NoServiceType:
				flat.NoService = group.Servers
			default:
				flat.UnknownServices = append(flat.UnknownServices, group.Servers...)
			}
		}
		flatGroups = append(flatGroups, flat)
	}
	return flatGroups
}

func renderClusterDashboard(servers []Server, rule GroupRule) []ClusterDashboard {
	clusterByID := make(map[int]Cluster)
	serversByCluster := make(map[int][]Server)
	for _, server := range servers {
		clusterID := server.ClusterID
		if _, exists := clusterByID[clusterID]; !exists {
			clusterByID[clusterID] = server.Cluster
		}
		serversByCluster[clusterID] = append(serversByCluster[clusterID], server)
	}

	clusters := make([]ClusterDashboard, 0, len(clusterByID))
	for clusterID, cluster := range clusterByID {
		groups := FlattenGroups(RenderDashboard(serversByCluster[clusterID], rule))
		clusters = append(clusters, ClusterDashboard{Cluster: cluster, Groups: groups})
	}
	slices.SortFunc(clusters, compareClusterDashboards)
	return clusters
}

func flattenClusterDashboard(clusters []ClusterDashboard) []ClusterFlatGroupEntry {
	flat := make([]ClusterFlatGroupEntry, 0)
	for _, cluster := range clusters {
		for _, group := range cluster.Groups {
			flat = append(flat, ClusterFlatGroupEntry{
				Cluster:   cluster.Cluster,
				GroupName: group.Name,
				Servers:   group.Servers,
			})
		}
	}
	return flat
}

func compareClusterDashboards(a, b ClusterDashboard) int {
	if rankA, rankB := clusterRoleRank(a.ClusterRole), clusterRoleRank(b.ClusterRole); rankA != rankB {
		if rankA < rankB {
			return -1
		}
		return 1
	}
	if a.ClusterID < b.ClusterID {
		return -1
	}
	if a.ClusterID > b.ClusterID {
		return 1
	}
	return 0
}

func clusterRoleRank(role string) int {
	switch role {
	case "primary":
		return 0
	case "secondary":
		return 1
	case "arbiter":
		return 2
	default:
		return 3
	}
}
