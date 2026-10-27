package baseline

// State 表示服务器或服务的运行状态。
// 状态以字符串形式参与 JSON 序列化；常用值由 StateNormal 和 StateAbnormal 提供。
type State string

// ServiceType 标识服务的类型。
// 目前由 ASType、TSType、SSType 和 DSType 定义支持的服务类型。
type ServiceType string

// ServiceRole 表示服务角色。
// 常用角色值由 Leader、Follower、LSS、SSRole 和 DSRole 定义。
type ServiceRole string

// ClusterRole 表示集群角色。
type ClusterRole string

const (
	// StateNormal 表示资源处于正常运行状态。
	StateNormal State = "normal"
	// StateAbnormal 表示资源处于异常状态。
	StateAbnormal State = "abnormal"
)

const (
	// ASType 是 AS 服务类型。
	ASType ServiceType = "AS"
	// TSType 是 TS 服务类型。
	TSType ServiceType = "TS"
	// SSType 是 SS 服务类型。
	SSType ServiceType = "SS"
	// DSType 是 DS 服务类型。
	DSType ServiceType = "DS"

	// NoServiceType 表示 Dashboard 中没有服务的分类，不是实际服务类型。
	NoServiceType ServiceType = "no-service"
	// UnknownServiceType 表示 Dashboard 中存在未识别服务的分类。
	UnknownServiceType ServiceType = "unknown-service"
)

const (
	noServiceGroupName      = "无服务"
	unknownServiceGroupName = "未知服务"
	unknownServiceGroupKey  = -1
	noServiceGroupKey       = -2
)

const (
	// Leader 表示 Leader 角色。
	Leader ServiceRole = "leader"
	// Follower 表示 Follower 角色。
	Follower ServiceRole = "follower"
	// LSS 表示 LSS 角色。
	LSS ServiceRole = "lss"
	// SSRole 表示 SS 角色。
	SSRole ServiceRole = "ss"
	// DSRole 表示 DS 角色。
	DSRole ServiceRole = "ds"
)

const (
	// PrimaryClusterRole 表示主集群。
	PrimaryClusterRole ClusterRole = "primary"
	// SecondaryClusterRole 表示从集群。
	SecondaryClusterRole ClusterRole = "secondary"
	// ArbiterClusterRole 表示仲裁集群。
	ArbiterClusterRole ClusterRole = "arbiter"
)

const (
	// ServiceNormal 和 ServerNormal 是服务与服务器正常状态的短名称。
	ServiceNormal = StateNormal
	ServerNormal  = StateNormal

	// ServiceAbnormal 和 ServerAbnormal 是服务与服务器异常状态的短名称。
	ServiceAbnormal = StateAbnormal
	ServerAbnormal  = StateAbnormal
)

// Cluster 描述集群的身份信息和角色信息。
type Cluster struct {
	// ClusterID 是集群标识，用于识别并聚合同一集群的数据。
	ClusterID int `json:"cluster_id"`
	// ClusterRole 是集群角色，例如 primary、secondary。
	ClusterRole ClusterRole `json:"cluster_role"`
	// ClusterName 是面向展示的集群名称。
	ClusterName string `json:"cluster_name"`
}

// Server 描述一台服务器及其所属集群和运行的服务。
type Server struct {
	// Name 是服务器名称。
	Name string `json:"name"`
	// Cluster 是该服务器所属的集群信息。
	Cluster
	// IP 是服务器 IP 地址。
	IP string `json:"ip"`
	// State 是服务器整体运行状态。
	State State `json:"state"`
	// Services 是该服务器上运行的服务实例列表。
	Services []Service `json:"services"`
}

// Service 描述一台服务器上运行的一个服务实例。
type Service struct {
	// ServiceType 是服务类型，JSON 字段名为 type。
	ServiceType ServiceType `json:"type"`
	// ServiceRole 是该服务的角色，不适用时可为空。
	Role ServiceRole `json:"role"`
	// State 是该服务的运行状态。
	State State `json:"state"`
}

// Group 表示一个分组。Key 仅供程序内部使用，不写入 JSON。
type Group struct {
	// Key 是分组内部键，不写入 JSON。
	Key int `json:"-"`
	// Name 是分组显示名称。
	Name string `json:"name"`
	// Servers 是归入该分组的按照排序规则排序的服务器。
	Servers []Server `json:"servers"`
}

// ClusterGroup 表示集群及其按选定规则排序的分组。
type ClusterGroup struct {
	Cluster
	// Groups 是该集群下按分组规则和组间顺序排列的分组。
	Groups []Group `json:"groups"`
}
