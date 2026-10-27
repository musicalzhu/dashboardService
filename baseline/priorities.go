package baseline

var (
	// servicePriority 按服务展示顺序排列；未登记的服务类型排在已知类型之后。
	servicePriority = buildServicePriority()

	// statePriority 定义服务器状态顺序；其他状态值排在已知状态之后。
	statePriority = map[State]int{
		StateAbnormal: 0,
		StateNormal:   1,
	}

	// rolePriority 定义服务角色顺序；其他角色值排在已知角色之后。
	rolePriority = map[ServiceRole]int{
		Leader:   0,
		Follower: 1,
		LSS:      2,
		SSRole:   3,
		DSRole:   4,
	}

	// clusterRolePriority 定义集群角色顺序；其他角色值排在已知角色之后。
	clusterRolePriority = map[ClusterRole]int{
		PrimaryClusterRole:   0,
		SecondaryClusterRole: 1,
	}

	// specialGroupPriority 只登记特殊组；普通组键查表默认优先级为 0。
	specialGroupPriority = map[int]int{
		unknownServiceGroupKey: 1,
		noServiceGroupKey:      2,
	}
)

func buildServicePriority() map[ServiceType]int {
	priority := make(map[ServiceType]int, len(allServiceTypes))
	for i, serviceType := range allServiceTypes {
		priority[serviceType] = i
	}
	return priority
}

// priorityOf 返回配置优先级；未登记的值排在所有已登记值之后。
func priorityOf[T comparable](priorities map[T]int, value T) int {
	if priority, ok := priorities[value]; ok {
		return priority
	}
	return len(priorities)
}
