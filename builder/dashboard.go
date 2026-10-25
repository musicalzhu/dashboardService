package builder

import (
	"sort"
)

// serviceTypeOrder 定义服务类型的业务优先级，越靠前优先级越高。
// 同一台服务器同时提供多种服务时，只会归入其中优先级最高的一种。
var serviceTypeOrder = []ServiceType{ASType, TSType, SSType, DSType}

// roleOrder 定义服务角色的排序优先级，数值越小越靠前。
var roleOrder = map[Role]int{
	Leader:   0,
	Follower: 1,
	LSS:      2,
	SS:       3,
	DS:       4,
}

// clusterRoleOrder 定义集群角色的排序优先级，数值越小越靠前。
var clusterRoleOrder = map[string]int{
	"primary":   0,
	"secondary": 1,
	"arbiter":   2,
}

// DashboardBuilder 根据服务器数据构建 Dashboard 视图。
// 它不保存请求状态；分组和排序规则由方法统一执行，嵌套与平铺视图共享同一份分组结果。
type DashboardBuilder struct{}

// NewDashboardBuilder 创建一个无状态的 DashboardBuilder。
func NewDashboardBuilder() *DashboardBuilder {
	return &DashboardBuilder{}
}

// GetDashboard 按集群和服务类型聚合服务器，并返回排序后的嵌套视图。
//
// 每台服务器依据其服务列表中优先级最高的服务类型，只进入一个服务分组。
// 没有服务的服务器进入 NoServiceType 分组；有服务但服务类型无法识别的服务器不会进入服务分组。
// 集群按 ClusterID 聚合；同一 ClusterID 的集群元数据采用输入中首次出现的值。
func (DashboardBuilder) GetDashboard(servers []Server) []ClusterGroup {
	clusters := make(map[int]*clusterAccumulator)
	for _, server := range servers {
		cluster, exists := clusters[server.ClusterID]
		if !exists {
			// 首次遇到该 ClusterID 时创建累加器，并保留该服务器携带的集群信息。
			cluster = &clusterAccumulator{
				cluster: server.Cluster,
				groups:  make(map[ServiceType][]Server),
			}
			clusters[server.ClusterID] = cluster
		}

		// 每台服务器只归入其最高优先级的服务类型，符合当前 Dashboard 的分组语义。
		serviceType := highestPriorityServiceType(server.Services)
		if serviceType == "" {
			serviceType = NoServiceType
		}
		// 复制并规范化服务列表，避免后续排序改写调用方传入的切片。
		cluster.groups[serviceType] = append(cluster.groups[serviceType], copyAndSortServer(server))
	}

	// map 的遍历顺序不稳定，因此收集结果后必须显式排序，保证输出顺序可预测。
	result := make([]ClusterGroup, 0, len(clusters))
	for _, cluster := range clusters {
		group := ClusterGroup{
			Cluster:       cluster.cluster,
			ServiceGroups: make([]ServiceGroup, 0, len(cluster.groups)),
		}
		for serviceType, groupedServers := range cluster.groups {
			// 各分组的服务器使用同一套状态、服务和网络地址排序规则。
			sortServers(groupedServers)
			group.ServiceGroups = append(group.ServiceGroups, ServiceGroup{
				ServiceType: serviceType,
				Servers:     groupedServers,
			})
		}
		sort.Slice(group.ServiceGroups, func(i, j int) bool {
			return serviceTypeLess(group.ServiceGroups[i].ServiceType, group.ServiceGroups[j].ServiceType)
		})
		result = append(result, group)
	}

	// 按集群角色、集群 ID 及名称确定可预测的集群顺序。
	sort.Slice(result, func(i, j int) bool {
		return clusterLess(result[i].Cluster, result[j].Cluster)
	})
	return result
}

// GetFlatDashboard 返回平铺格式的 Dashboard 视图。
// 它先调用 GetDashboard 复用聚合和排序规则，再把服务分组投影到 AS、TS、SS、DS 字段，
// 避免平铺视图重新实现一套可能与嵌套视图产生差异的聚合逻辑。
func (b DashboardBuilder) GetFlatDashboard(servers []Server) []ClusterFlatGroup {
	grouped := b.GetDashboard(servers)
	result := make([]ClusterFlatGroup, 0, len(grouped))
	for _, cluster := range grouped {
		flat := ClusterFlatGroup{Cluster: cluster.Cluster}
		for _, serviceGroup := range cluster.ServiceGroups {
			switch serviceGroup.ServiceType {
			case ASType:
				flat.AS = serviceGroup.Servers
			case TSType:
				flat.TS = serviceGroup.Servers
			case SSType:
				flat.SS = serviceGroup.Servers
			case DSType:
				flat.DS = serviceGroup.Servers
			case NoServiceType:
				flat.NoService = serviceGroup.Servers
			}
		}
		result = append(result, flat)
	}
	return result
}

// clusterAccumulator 是构建过程中的内部累加结构。
// groups 以服务类型为键保存该类型对应的服务器；最终转换为有序的 ServiceGroup 切片。
type clusterAccumulator struct {
	cluster Cluster
	groups  map[ServiceType][]Server
}

// highestPriorityServiceType 返回服务列表中优先级最高的服务类型。
// 空服务类型会被忽略；未配置在 serviceTypeOrder 中的类型排在所有已知类型之后，
// 多个未知类型之间按类型名称字典序比较。
func highestPriorityServiceType(services []Service) ServiceType {
	var highest ServiceType
	for _, service := range services {
		if service.ServiceType == "" {
			continue
		}
		if highest == "" || serviceTypeLess(service.ServiceType, highest) {
			highest = service.ServiceType
		}
	}
	return highest
}

// copyAndSortServer 复制服务器的 Services 切片，并按服务类型、角色、Paxos 组和状态排序。
// Server 本身按值复制，但 Services 是切片，因此需要单独复制底层数组以避免输入副作用。
func copyAndSortServer(server Server) Server {
	server.Services = append([]Service(nil), server.Services...)
	sort.SliceStable(server.Services, func(i, j int) bool {
		left, right := server.Services[i], server.Services[j]
		if left.ServiceType != right.ServiceType {
			return serviceTypeLess(left.ServiceType, right.ServiceType)
		}
		if left.Role != right.Role {
			return roleLess(left.Role, right.Role)
		}
		if left.PaxosGroup != right.PaxosGroup {
			return left.PaxosGroup < right.PaxosGroup
		}
		return left.State < right.State
	})
	return server
}

// sortServers 按服务器状态、服务列表及 IP 和名称对服务器切片排序。
// 异常状态优先于正常状态；服务列表先按服务类型、再按角色比较；
// 如果前述字段相同，则按服务数量、IP 和名称作为后续比较条件。
func sortServers(servers []Server) {
	sort.SliceStable(servers, func(i, j int) bool {
		left, right := servers[i], servers[j]
		if left.State != right.State {
			if stateRank(left.State) != stateRank(right.State) {
				return stateRank(left.State) < stateRank(right.State)
			}
			return left.State < right.State
		}
		for index := 0; index < len(left.Services) && index < len(right.Services); index++ {
			leftService, rightService := left.Services[index], right.Services[index]
			if leftService.ServiceType != rightService.ServiceType {
				return serviceTypeLess(leftService.ServiceType, rightService.ServiceType)
			}
			if leftService.Role != rightService.Role {
				return roleLess(leftService.Role, rightService.Role)
			}
		}
		if len(left.Services) != len(right.Services) {
			return len(left.Services) < len(right.Services)
		}
		if left.IP != right.IP {
			return left.IP < right.IP
		}
		return left.Name < right.Name
	})
}

// serviceTypeLess 按预设服务优先级比较两个服务类型。
// 优先级相同时（包括未知类型），按类型名称字典序比较，保证顺序确定。
func serviceTypeLess(left, right ServiceType) bool {
	leftRank, rightRank := serviceTypeRank(left), serviceTypeRank(right)
	if leftRank != rightRank {
		return leftRank < rightRank
	}
	return left < right
}

// serviceTypeRank 返回服务类型在 serviceTypeOrder 中的序号。
// 未知类型返回切片长度，因此排在所有已知类型之后。
func serviceTypeRank(serviceType ServiceType) int {
	for rank, candidate := range serviceTypeOrder {
		if serviceType == candidate {
			return rank
		}
	}
	return len(serviceTypeOrder)
}

// roleLess 按预设角色优先级比较角色。
// 已知角色始终排在未知角色之前；同属已知或未知时按角色名称字典序比较。
func roleLess(left, right Role) bool {
	leftRank, leftKnown := roleOrder[left]
	rightRank, rightKnown := roleOrder[right]
	if leftKnown != rightKnown {
		return leftKnown
	}
	if leftRank != rightRank {
		return leftRank < rightRank
	}
	return left < right
}

// stateRank 返回服务器状态的排序等级：异常优先，其次正常，未知状态最后。
func stateRank(state State) int {
	switch state {
	case StateAbnormal:
		return 0
	case StateNormal:
		return 1
	default:
		return 2
	}
}

// clusterLess 按集群角色优先级、集群 ID 和集群名称依次比较集群。
// 未知集群角色排在已知角色之后；同级集群继续按 ID 和名称确定顺序。
func clusterLess(left, right Cluster) bool {
	leftRank, leftKnown := clusterRoleOrder[left.ClusterRole]
	rightRank, rightKnown := clusterRoleOrder[right.ClusterRole]
	if leftKnown != rightKnown {
		return leftKnown
	}
	if leftRank != rightRank {
		return leftRank < rightRank
	}
	if left.ClusterID != right.ClusterID {
		return left.ClusterID < right.ClusterID
	}
	return left.ClusterName < right.ClusterName
}
