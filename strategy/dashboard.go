package strategy

import (
	"sort"
)

// ========== 排序配置 ==========

// servicePriority 定义服务类型的分组排序优先级，值越小优先级越高
var servicePriority = map[ServiceType]int{
	ASType:        0,
	TSType:        1,
	SSType:        2,
	DSType:        3,
	NoServiceType: 4,
}

// statePriority 定义服务器或服务状态的排序优先级，值越小优先级越高
var statePriority = map[State]int{
	StateAbnormal: 0,
	StateNormal:   1,
}

// rolePriority 定义服务角色的排序优先级，值越小优先级越高
var rolePriority = map[Role]int{
	Leader:   0,
	Follower: 1,
	LSS:      2,
	SS:       3,
	DS:       4,
}

// clusterRolePriority 定义集群角色的排序优先级，值越小优先级越高
var clusterRolePriority = map[string]int{
	"primary":   0,
	"secondary": 1,
	"arbiter":   2,
}

// ========== 接口定义 ==========

// DashboardService 定义了生成 Dashboard 数据的接口（分组结构）
type DashboardService interface {
	// GetDashboard 返回分组和排序后的集群服务结构
	GetDashboard(servers []Server) []ClusterGroup
}

// DashboardFlatService 定义了生成平铺结构 Dashboard 数据的接口
type DashboardFlatService interface {
	// GetFlatDashboard 返回平铺结构的集群服务结构
	GetFlatDashboard(servers []Server) []ClusterFlatGroup
}

// Sorter 定义了排序逻辑的接口
type Sorter interface {
	// SortServiceGroups 对 ServiceGroup 切片按服务类型优先级排序
	SortServiceGroups(serviceGroups []ServiceGroup)
	// SortServers 对 Server 切片按状态、服务类型和角色优先级排序
	SortServers(servers []Server)
	// SortClusterGroups 对 ClusterGroup 切片按集群角色优先级排序
	SortClusterGroups(clusterGroups []ClusterGroup)
	// SortClusterFlatGroups 对 ClusterFlatGroup 切片按集群角色优先级排序
	SortClusterFlatGroups(groups []ClusterFlatGroup)
}

// ClusterGrouper 策略模式接口，支持不同聚合/输出结构体
type ClusterGrouper interface {
	Group(servers []Server) any
}

// ========== 默认实现 ==========

// DefaultSorter 是 Sorter 的默认实现，提供默认的排序规则
type DefaultSorter struct{}

// SortServiceGroups 对 ServiceGroup 切片按服务类型优先级排序
func (s *DefaultSorter) SortServiceGroups(serviceGroups []ServiceGroup) {
	sort.Slice(serviceGroups, func(i, j int) bool {
		return servicePriority[serviceGroups[i].ServiceType] < servicePriority[serviceGroups[j].ServiceType]
	})
}

// SortServers 对 Server 切片按状态优先级、服务类型优先级、角色优先级和 IP 排序
func (s *DefaultSorter) SortServers(servers []Server) {
	sort.Slice(servers, func(i, j int) bool {
		// 1. 按状态优先级排序
		if servers[i].State != servers[j].State {
			return statePriority[servers[i].State] < statePriority[servers[j].State]
		}
		// 2. 按服务器包含的服务类型和服务角色优先级排序
		if len(servers[i].Services) > 0 && len(servers[j].Services) > 0 {
			for idx := 0; idx < len(servers[i].Services) && idx < len(servers[j].Services); idx++ {
				serviceI := servers[i].Services[idx]
				serviceJ := servers[j].Services[idx]
				if servicePriority[serviceI.ServiceType] != servicePriority[serviceJ.ServiceType] {
					return servicePriority[serviceI.ServiceType] < servicePriority[serviceJ.ServiceType]
				}
				if rolePriority[serviceI.Role] != rolePriority[serviceJ.Role] {
					return rolePriority[serviceI.Role] < rolePriority[serviceJ.Role]
				}
			}
		}
		// 3. 按服务器 IP 排序
		return servers[i].IP < servers[j].IP
	})
}

// SortClusterGroups 对 ClusterGroup 切片按集群角色优先级排序
func (s *DefaultSorter) SortClusterGroups(clusterGroups []ClusterGroup) {
	sort.Slice(clusterGroups, func(i, j int) bool {
		pi := clusterRolePriority[clusterGroups[i].ClusterRole]
		pj := clusterRolePriority[clusterGroups[j].ClusterRole]
		if pi != pj {
			return pi < pj
		}
		return clusterGroups[i].ClusterID < clusterGroups[j].ClusterID
	})
}

// SortClusterFlatGroups 对 ClusterFlatGroup 切片按集群角色优先级排序
func (s *DefaultSorter) SortClusterFlatGroups(ClusterFlatGroup []ClusterFlatGroup) {
	sort.Slice(ClusterFlatGroup, func(i, j int) bool {
		pi := clusterRolePriority[ClusterFlatGroup[i].ClusterRole]
		pj := clusterRolePriority[ClusterFlatGroup[j].ClusterRole]
		if pi != pj {
			return pi < pj
		}
		return ClusterFlatGroup[i].ClusterID < ClusterFlatGroup[j].ClusterID
	})
}

// DefaultDashboardService 是 DashboardService 的默认实现（分组结构）
type DefaultDashboardService struct {
	sorter Sorter
}

// NewDefaultDashboardService 构造函数，注入排序器
func NewDefaultDashboardService(sorter Sorter) *DefaultDashboardService {
	return &DefaultDashboardService{sorter: sorter}
}

// GetDashboard 实现 DashboardService 接口，返回分组和排序后的数据
func (d *DefaultDashboardService) GetDashboard(servers []Server) []ClusterGroup {
	// 使用 map[int]*ClusterGroup 动态构建 ClusterGroup
	clusterMap := make(map[int]*ClusterGroup)

	for _, server := range servers {
		// 获取或创建对应的 ClusterGroup
		if _, exists := clusterMap[server.ClusterID]; !exists {
			clusterMap[server.ClusterID] = &ClusterGroup{
				Cluster:       server.Cluster,
				ServiceGroups: []ServiceGroup{},
			}
		}
		clusterGroup := clusterMap[server.ClusterID]

		// 获取服务器中优先级最高的服务类型
		highestPriorityServiceType := getHighestPriorityServiceType(server.Services)
		if highestPriorityServiceType == "" {
			highestPriorityServiceType = NoServiceType
		}
		if highestPriorityServiceType != "" {
			// 获取或创建对应的 ServiceGroup
			var serviceGroup *ServiceGroup
			for i := range clusterGroup.ServiceGroups {
				if clusterGroup.ServiceGroups[i].ServiceType == highestPriorityServiceType {
					serviceGroup = &clusterGroup.ServiceGroups[i]
					break
				}
			}
			if serviceGroup == nil {
				// 如果不存在对应的 ServiceGroup，则创建一个新的
				newServiceGroup := ServiceGroup{
					ServiceType: highestPriorityServiceType,
					Servers:     []Server{},
				}
				clusterGroup.ServiceGroups = append(clusterGroup.ServiceGroups, newServiceGroup)
				// 获取这个新创建的 ServiceGroup 的指针
				serviceGroup = &clusterGroup.ServiceGroups[len(clusterGroup.ServiceGroups)-1]
			}
			// 将服务器添加到对应的 ServiceGroup
			serviceGroup.Servers = append(serviceGroup.Servers, server)
		}
	}

	// 将 map 转换为切片并排序
	clusterGroups := make([]ClusterGroup, 0, len(clusterMap))
	for _, clusterGroup := range clusterMap {
		// 对每个 ClusterGroup 的 ServiceGroups 排序
		d.sorter.SortServiceGroups(clusterGroup.ServiceGroups)
		// 对每个 ServiceGroup 的 Servers 排序
		for i := range clusterGroup.ServiceGroups {
			d.sorter.SortServers(clusterGroup.ServiceGroups[i].Servers)
		}
		clusterGroups = append(clusterGroups, *clusterGroup)
	}

	// 对 ClusterGroups 排序
	d.sorter.SortClusterGroups(clusterGroups)
	return clusterGroups
}

// DefaultDashboardFlatService 是 DashboardFlatService 的默认实现（平铺结构）
type DefaultDashboardFlatService struct {
	sorter Sorter
}

// NewDefaultDashboardFlatService 构造函数，注入排序器
func NewDefaultDashboardFlatService(sorter Sorter) *DefaultDashboardFlatService {
	return &DefaultDashboardFlatService{sorter: sorter}
}

// GetFlatDashboard 实现 DashboardFlatService 接口，返回平铺结构数据
func (d *DefaultDashboardFlatService) GetFlatDashboard(servers []Server) []ClusterFlatGroup {
	clusterMap := make(map[int]*ClusterFlatGroup)
	for _, server := range servers {
		if _, exists := clusterMap[server.ClusterID]; !exists {
			clusterMap[server.ClusterID] = &ClusterFlatGroup{
				Cluster: server.Cluster,
			}
		}
		cg := clusterMap[server.ClusterID]
		// 只分配到优先级最高的服务类型组
		highest := getHighestPriorityServiceType(server.Services)
		switch highest {
		case ASType:
			cg.AS = append(cg.AS, server)
		case TSType:
			cg.TS = append(cg.TS, server)
		case SSType:
			cg.SS = append(cg.SS, server)
		case DSType:
			cg.DS = append(cg.DS, server)
		default:
			if len(server.Services) == 0 {
				cg.NoService = append(cg.NoService, server)
			}
		}
	}
	clusterGroups := make([]ClusterFlatGroup, 0, len(clusterMap))
	for _, cg := range clusterMap {
		d.sorter.SortServers(cg.AS)
		d.sorter.SortServers(cg.TS)
		d.sorter.SortServers(cg.SS)
		d.sorter.SortServers(cg.DS)
		d.sorter.SortServers(cg.NoService)
		clusterGroups = append(clusterGroups, *cg)
	}
	d.sorter.SortClusterFlatGroups(clusterGroups)
	return clusterGroups
}

// ========== 策略模式实现 ==========

// ServiceGroupStrategy 分组结构策略
type ServiceGroupStrategy struct {
	sorter Sorter
}

// Group 实现 ClusterGrouper 接口，返回分组结构
func (s *ServiceGroupStrategy) Group(servers []Server) any {
	dashboard := NewDefaultDashboardService(s.sorter)
	return dashboard.GetDashboard(servers) // 返回[]ClusterGroup
}

// FlatTypeGroupStrategy 平铺结构策略
type FlatTypeGroupStrategy struct {
	sorter Sorter
}

// Group 实现 ClusterGrouper 接口，返回平铺结构
func (s *FlatTypeGroupStrategy) Group(servers []Server) any {
	dashboard := NewDefaultDashboardFlatService(s.sorter)
	return dashboard.GetFlatDashboard(servers) // 返回[]ClusterFlatGroup
}

// ========== 工具函数 ==========

// getHighestPriorityServiceType 获取服务列表中优先级最高的服务类型
func getHighestPriorityServiceType(services []Service) ServiceType {
	if len(services) == 0 {
		return ""
	}
	// 按服务类型优先级排序
	sort.Slice(services, func(i, j int) bool {
		return servicePriority[services[i].ServiceType] < servicePriority[services[j].ServiceType]
	})
	return services[0].ServiceType
}

// ========== 策略选择入口 ==========

// RenderDashboard 根据策略输出不同结构体
func RenderDashboard(grouper ClusterGrouper, servers []Server) any {
	return grouper.Group(servers)
}

// ========== 示例数据 ==========

// main 仅作占位，实际业务入口应由调用方实现
