package baseline

import "slices"

type clusterServers struct {
	cluster Cluster
	servers []Server
}

// RenderClusterDashboard 按 ClusterID 分组，并在每个集群内按主服务类型组织服务器。
// 服务类型顺序为 AS、TS、SS、DS、未知服务、无服务；集群内服务器使用 baseline 的共享排序。
// 同一 ClusterID 的集群元数据取输入中首次出现的值。
func RenderClusterDashboard(servers []Server) []ClusterGroup {
	byCluster := make(map[int]*clusterServers)
	for _, server := range servers {
		clusterID := server.ClusterID
		cluster := byCluster[clusterID]
		if cluster == nil {
			cluster = &clusterServers{cluster: server.Cluster}
			byCluster[clusterID] = cluster
		}
		cluster.servers = append(cluster.servers, server)
	}

	groups := make([]ClusterGroup, 0, len(byCluster))
	for _, cluster := range byCluster {
		serviceGroups := GroupBy(cluster.servers, PrimaryServiceMask, CompareByMaskDesc)
		group := ClusterGroup{
			Cluster:       cluster.cluster,
			ServiceGroups: make([]ServiceGroup, 0, len(serviceGroups)),
		}
		for _, serviceGroup := range serviceGroups {
			serviceType := ServiceType(serviceGroup.Name)
			switch serviceGroup.Key {
			case unknownServiceGroupKey:
				serviceType = UnknownServiceType
			case 0:
				serviceType = NoServiceType
			}
			group.ServiceGroups = append(group.ServiceGroups, ServiceGroup{
				ServiceType: serviceType,
				Servers:     serviceGroup.Servers,
			})
		}
		groups = append(groups, group)
	}

	slices.SortFunc(groups, compareClusterGroups)
	return groups
}

// RenderClusterFlatDashboard 返回与 RenderClusterDashboard 相同的分类结果，
// 但按服务类型放入固定字段，供需要平铺 JSON 结构的调用方使用。
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
