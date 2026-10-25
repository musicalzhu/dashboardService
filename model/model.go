package model

// State 表示服务器或微服务的运行状态。
// 状态以字符串形式参与 JSON 序列化；常用值由 StateNormal 和 StateAbnormal 提供。
type State string

// ServiceType 标识微服务的类型。
// 目前由 ASType、TSType、SSType 和 DSType 定义支持的服务类型。
type ServiceType string

// Role 表示微服务在其服务组或一致性协议中的角色。
// 常用角色值由 Leader、Follower、LSS、SS 和 DS 定义。
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

	// NoServiceType 是 Dashboard 分组使用的“无服务”类别标识，不代表实际微服务类型。
	NoServiceType ServiceType = "no-service"
)

const (
	// Leader 表示 Leader 角色。
	Leader Role = "leader"

	// Follower 表示 Follower 角色。
	Follower Role = "follower"

	// LSS 表示 LSS 角色。
	LSS Role = "lss"

	// SS 表示 SS 角色。
	SS Role = "ss"

	// DS 表示 DS 角色。
	DS Role = "ds"
)

// Service 描述一台服务器上运行的一个微服务实例。
type Service struct {
	// ServiceType 是服务类型，JSON 字段名为 type。
	ServiceType ServiceType `json:"type"`

	// PaxosGroup 是该服务所属的 Paxos 组名称；不适用时可为空。
	PaxosGroup string `json:"paxos_group"`

	// Role 是该服务在相关服务组或协议中的角色；不适用时可为空。
	Role Role `json:"role"`

	// State 是该微服务的运行状态。
	State State `json:"state"`
}

// Cluster 描述集群的身份信息和角色信息。
type Cluster struct {
	// ClusterID 是集群标识，用于识别和聚合同一集群的数据。
	ClusterID int `json:"cluster_id"`

	// ClusterRole 是集群角色标识，例如 primary、secondary 或 arbiter。
	ClusterRole string `json:"cluster_role"`

	// ClusterName 是面向展示的集群名称。
	ClusterName string `json:"cluster_name"`
}

// Server 描述一台服务器及其所属集群和运行的微服务。
type Server struct {
	// Name 是服务器名称。
	Name string `json:"name"`

	// Cluster 匿名嵌入集群信息。序列化时 Cluster 的字段提升到 Server 对象，
	// 因此 JSON 中 cluster_id、cluster_role、cluster_name 与其他服务器字段处于同一层级。
	Cluster

	// IP 是服务器的 IP 地址。
	IP string `json:"ip"`

	// State 是服务器整体运行状态，与单个 Service 的 State 分别表示不同层级的状态。
	State State `json:"state"`

	// Services 是该服务器上运行的微服务实例列表。
	Services []Service `json:"services"`
}

// ClusterGroup 表示一个集群及其按服务类型组织的嵌套分组。
// 没有服务的服务器放在 ServiceType 为 NoServiceType 的独立分组中。
type ClusterGroup struct {
	// Cluster 匿名嵌入集群信息；JSON 中集群字段与 service_groups 位于同一层级。
	Cluster

	// ServiceGroups 是该集群下的服务类型分组列表。
	ServiceGroups []ServiceGroup `json:"service_groups"`
}

// ServiceGroup 表示一个服务类型分组及属于该分组的服务器。
type ServiceGroup struct {
	// ServiceType 是该分组对应的微服务类型，或 NoServiceType 这一无服务类别标识。
	ServiceType ServiceType `json:"service_type"`

	// Servers 是归入该服务类型分组的服务器列表。
	Servers []Server `json:"servers"`
}

// ClusterFlatGroup 表示一个集群及按固定服务类型字段展开的平铺视图。
// 与 ClusterGroup 的 ServiceGroups 切片不同，该结构为 AS、TS、SS、DS 分别提供字段，
// 并通过 NoService 单独返回没有运行服务的服务器。
type ClusterFlatGroup struct {
	// Cluster 匿名嵌入集群信息；JSON 中集群字段与各服务字段位于同一层级。
	Cluster

	// AS 是 AS 服务类型对应的服务器列表；为空时 JSON 中省略该字段。
	AS []Server `json:"AS,omitempty"`

	// TS 是 TS 服务类型对应的服务器列表；为空时 JSON 中省略该字段。
	TS []Server `json:"TS,omitempty"`

	// SS 是 SS 服务类型对应的服务器列表；为空时 JSON 中省略该字段。
	SS []Server `json:"SS,omitempty"`

	// DS 是 DS 服务类型对应的服务器列表；为空时 JSON 中省略该字段。
	DS []Server `json:"DS,omitempty"`

	// NoService 是没有运行任何服务的服务器列表；为空时 JSON 中省略该字段。
	NoService []Server `json:"no_service,omitempty"`
}
