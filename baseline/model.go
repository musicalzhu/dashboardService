package baseline

// State 表示服务器或微服务的运行状态。
// 状态以字符串形式参与 JSON 序列化；常用值由 StateNormal 和 StateAbnormal 提供。
type State string

// ServiceType 标识微服务的类型。
// 目前由 ASType、TSType、SSType 和 DSType 定义支持的服务类型。
type ServiceType string

// Role 表示微服务在其服务组或一致性协议中的角色。
// 常用角色值由 Leader、Follower、LSS、RoleSS 和 RoleDS 定义。
type Role string

const (
	// StateNormal 表示资源处于正常运行状态。
	StateNormal State = "normal"
	// StateAbnormal 表示资源处于异常状态。
	StateAbnormal State = "abnormal"
)

const (
	// ASType 是 AS 微服务类型。
	ASType ServiceType = "AS"
	// TSType 是 TS 微服务类型。
	TSType ServiceType = "TS"
	// SSType 是 SS 微服务类型。
	SSType ServiceType = "SS"
	// DSType 是 DS 微服务类型。
	DSType ServiceType = "DS"

	// AS、TS、SS 和 DS 是服务类型的短名称。
	AS = ASType
	TS = TSType
	SS = SSType
	DS = DSType

	// NoServiceType 表示 Dashboard 中没有服务的分类，不是实际服务类型。
	NoServiceType ServiceType = "no-service"
	// UnknownServiceType 表示 Dashboard 中存在未识别服务的分类。
	UnknownServiceType ServiceType = "unknown-service"
)

const (
	// Leader 表示 Leader 角色。
	Leader Role = "leader"
	// Follower 表示 Follower 角色。
	Follower Role = "follower"
	// LSS 表示 LSS 角色。
	LSS Role = "lss"
)

const (
	// SvcNormal 和 SrvNormal 是服务与服务器正常状态的兼容短名称。
	SvcNormal = StateNormal
	SrvNormal = StateNormal

	// SvcAbnormal 和 SrvAbnormal 是服务与服务器异常状态的兼容短名称。
	SvcAbnormal = StateAbnormal
	SrvAbnormal = StateAbnormal

	// RoleSS 和 RoleDS 是 SS、DS 服务角色的短名称。
	RoleSS Role = "ss"
	RoleDS Role = "ds"
)

// Service 描述一台服务器上运行的一个微服务实例。
type Service struct {
	// ServiceType 是服务类型，JSON 字段名为 type。
	ServiceType ServiceType `json:"type"`
	// PaxosGroup 是该服务所属的 Paxos 组；不适用时可为空。
	PaxosGroup string `json:"paxos_group"`
	// Role 是该服务在相关服务组或协议中的角色。
	Role Role `json:"role"`
	// State 是该微服务的运行状态。
	State State `json:"state"`
}

// Cluster 描述集群的身份信息和角色信息。
type Cluster struct {
	// ClusterID 是集群标识，用于识别并聚合同一集群的数据。
	ClusterID int `json:"cluster_id"`
	// ClusterRole 是集群角色，例如 primary、secondary 或 arbiter。
	ClusterRole string `json:"cluster_role"`
	// ClusterName 是面向展示的集群名称。
	ClusterName string `json:"cluster_name"`
}

// Server 描述一台服务器及其所属集群和运行的微服务。
type Server struct {
	// Name 是服务器名称。
	Name string `json:"name"`
	Cluster
	// IP 是服务器 IP 地址。
	IP string `json:"ip"`
	// State 是服务器整体运行状态。
	State State `json:"state"`
	// Services 是该服务器上运行的微服务实例列表。
	Services []Service `json:"services"`
}

// Group 表示一个服务分组。Key 仅供程序内部使用，不写入 JSON。
type Group struct {
	// Key 是分组内部键，不写入 JSON。
	Key int `json:"-"`
	// Name 是分组显示名称。
	Name string `json:"name"`
	// Servers 是归入该分组的服务器。
	Servers []Server `json:"servers"`
}

// ClusterGroup 是兼容的集群嵌套视图，按服务类型组织服务器。
type ClusterGroup struct {
	Cluster
	// ServiceGroups 是该集群下的服务类型分组。
	ServiceGroups []ServiceGroup `json:"service_groups"`
}

// ServiceGroup 表示一个服务类型分组及属于该分组的服务器。
type ServiceGroup struct {
	// ServiceType 是该分组对应的服务类型。
	ServiceType ServiceType `json:"service_type"`
	// Servers 是属于该服务分组的服务器。
	Servers []Server `json:"servers"`
}

// FlatGroup 是通用平铺 JSON 中按排序顺序输出的分组条目。
type FlatGroup struct {
	// Name 是分组显示名称。
	Name string `json:"name"`
	// Servers 是归入该分组的服务器。
	Servers []Server `json:"servers"`
}

// ClusterDashboard 表示一个集群及其有序的分组条目。
type ClusterDashboard struct {
	Cluster
	// Groups 是该集群内按选定规则排序的分组。
	Groups []FlatGroup `json:"groups"`
}

// ClusterFlatGroupEntry 是集群平铺 JSON 中的单个集群/分组行。
type ClusterFlatGroupEntry struct {
	Cluster
	// GroupName 是该行对应的分组名称。
	GroupName string `json:"group"`
	// Servers 是该分组中的服务器。
	Servers []Server `json:"servers"`
}

// ClusterFlatGroup 是按固定服务类型字段展开的兼容输出结构。
type ClusterFlatGroup struct {
	Cluster
	// AS、TS、SS 和 DS 分别是对应主服务类型的服务器列表。
	AS []Server `json:"AS,omitempty"`
	TS []Server `json:"TS,omitempty"`
	SS []Server `json:"SS,omitempty"`
	DS []Server `json:"DS,omitempty"`
	// NoService 是没有运行服务的服务器列表。
	NoService []Server `json:"no_service,omitempty"`
	// UnknownServices 是包含未识别服务的服务器列表。
	UnknownServices []Server `json:"unknown_services,omitempty"`
}

// ServiceStatus、ServerStatus 和 PaxosRole 是兼容历史领域用语的别名。
type ServiceStatus = State
type ServerStatus = State
type PaxosRole = Role
