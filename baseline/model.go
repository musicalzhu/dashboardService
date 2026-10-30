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
// 常用角色值由 PrimaryClusterRole 和 SecondaryClusterRole 定义。
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
	noServiceGroupKey       = 0
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
)

const (
	// ClusterNormal 表示集群处于正常运行状态。
	// ServerNormal 表示服务器处于正常运行状态。
	// ServiceNormal 表示服务处于正常运行状态。
	ClusterNormal = StateNormal
	ServerNormal  = StateNormal
	ServiceNormal = StateNormal

	// ClusterAbnormal 表示集群处于异常状态。
	// ServerAbnormal 表示服务器处于异常状态。
	// ServiceAbnormal 表示服务处于异常状态。
	clusterAbnormal = StateAbnormal
	ServerAbnormal  = StateAbnormal
	ServiceAbnormal = StateAbnormal
)

// Cluster 描述集群的身份信息和角色信息。
type Cluster struct {
	// ClusterID 是集群标识，用于识别并聚合同一集群的数据。
	ClusterID int `json:"cluster_id"`
	// ClusterRole 是集群角色，例如 primary、secondary。
	ClusterRole ClusterRole `json:"cluster_role"`
	// ClusterName 是面向展示的集群名称。
	ClusterName string `json:"cluster_name"`
	// State 是集群整体运行状态。
	State State `json:"state"`
}

// 领域模型定义了服务器、服务和集群的基本结构，以及它们之间的关系。
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

// KeyFunc 根据服务器计算分组键。
type KeyFunc func(Server) int

// CompareFunc 按 slices.SortFunc 约定比较两个分组键。
type CompareFunc func(a, b int) int

// ruleImpl 封装分组键函数和比较器，用于分组规则的分发。
type RuleImpl struct {
	KeyOf KeyFunc
	Cmp   CompareFunc
}

// GroupRule 表示一个服务器分组和排序规则。
// 前四种规则，服务器按服务组合分组，规定了分组之间的顺序；
// 最后一种规则，服务器只是按最高优先级的主服务分组，规定了主服务之间的顺序。
type GroupRule int
