package strategy

import "dashboardService/model"

type State = model.State
type ServiceType = model.ServiceType
type Role = model.Role
type Service = model.Service
type Cluster = model.Cluster
type Server = model.Server
type ClusterGroup = model.ClusterGroup
type ServiceGroup = model.ServiceGroup
type ClusterFlatGroup = model.ClusterFlatGroup

const (
	StateNormal   = model.StateNormal
	StateAbnormal = model.StateAbnormal
	ASType        = model.ASType
	TSType        = model.TSType
	SSType        = model.SSType
	DSType        = model.DSType
	NoServiceType = model.NoServiceType
	Leader        = model.Leader
	Follower      = model.Follower
	LSS           = model.LSS
	SS            = model.SS
	DS            = model.DS
)
