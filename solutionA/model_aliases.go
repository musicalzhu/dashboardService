package strategy

import "dashboardService/baseline"

type State = baseline.State
type ServiceType = baseline.ServiceType
type Role = baseline.ServiceRole
type ClusterRole = baseline.ClusterRole
type Service = baseline.Service
type Cluster = baseline.Cluster
type Server = baseline.Server

// ServiceGroup 是 solutionA 特定主服务视图的输出结构。
type ServiceGroup struct {
	ServiceType ServiceType
	Servers     []Server
}

// ClusterGroup 是 solutionA 的嵌套输出结构。
type ClusterGroup struct {
	Cluster
	ServiceGroups []ServiceGroup
}

// ClusterFlatGroup 是 solutionA 针对主服务分类的固定字段输出结构。
type ClusterFlatGroup struct {
	Cluster
	AS              []Server `json:"AS,omitempty"`
	TS              []Server `json:"TS,omitempty"`
	SS              []Server `json:"SS,omitempty"`
	DS              []Server `json:"DS,omitempty"`
	NoService       []Server `json:"no_service,omitempty"`
	UnknownServices []Server `json:"unknown_services,omitempty"`
}

const (
	StateNormal          = baseline.StateNormal
	StateAbnormal        = baseline.StateAbnormal
	ASType               = baseline.ASType
	TSType               = baseline.TSType
	SSType               = baseline.SSType
	DSType               = baseline.DSType
	NoServiceType        = baseline.NoServiceType
	Leader               = baseline.Leader
	Follower             = baseline.Follower
	LSS                  = baseline.LSS
	SS                   = baseline.SSRole
	DS                   = baseline.DSRole
	SSRole               = baseline.SSRole
	DSRole               = baseline.DSRole
	PrimaryClusterRole   = baseline.PrimaryClusterRole
	SecondaryClusterRole = baseline.SecondaryClusterRole
	ArbiterClusterRole   = baseline.ArbiterClusterRole
)
