package builder

import "dashboardService/baseline"

type State = baseline.State
type ServiceType = baseline.ServiceType
type Role = baseline.Role
type Service = baseline.Service
type Cluster = baseline.Cluster
type Server = baseline.Server
type ClusterGroup = baseline.ClusterGroup
type ServiceGroup = baseline.ServiceGroup
type ClusterFlatGroup = baseline.ClusterFlatGroup

const (
	StateNormal   = baseline.StateNormal
	StateAbnormal = baseline.StateAbnormal
	ASType        = baseline.ASType
	TSType        = baseline.TSType
	SSType        = baseline.SSType
	DSType        = baseline.DSType
	NoServiceType = baseline.NoServiceType
	Leader        = baseline.Leader
	Follower      = baseline.Follower
	LSS           = baseline.LSS
	SS            = baseline.RoleSS
	DS            = baseline.RoleDS
)
