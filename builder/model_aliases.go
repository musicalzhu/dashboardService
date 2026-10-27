package builder

import "dashboardService/baseline"

type State = baseline.State
type ServiceType = baseline.ServiceType
type Role = baseline.ServiceRole
type ClusterRole = baseline.ClusterRole
type Service = baseline.Service
type Cluster = baseline.Cluster
type Server = baseline.Server

// ServiceGroup 是 Builder POC 的主服务视图分组。
type ServiceGroup struct {
	ServiceType ServiceType
	Servers     []Server
}

// ClusterGroup 是 Builder POC 的嵌套输出形状。
type ClusterGroup struct {
	Cluster
	ServiceGroups []ServiceGroup
}

// ClusterFlatGroup 是 Builder POC 针对主服务规则的固定字段输出形状。
type ClusterFlatGroup struct {
	Cluster
	AS        []Server
	TS        []Server
	SS        []Server
	DS        []Server
	NoService []Server
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
	SSRole               = baseline.SSRole
	DSRole               = baseline.DSRole
	SS                   = baseline.SSRole
	DS                   = baseline.DSRole
	PrimaryClusterRole   = baseline.PrimaryClusterRole
	SecondaryClusterRole = baseline.SecondaryClusterRole
	ArbiterClusterRole   = baseline.ArbiterClusterRole
)
