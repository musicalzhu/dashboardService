package closure

import "dashboardService/baseline"

// Group、Server、Service 和相关领域类型复用 baseline 的共享模型。
type Group = baseline.Group
type ClusterGroup = baseline.ClusterGroup
type Role = baseline.ServiceRole
type State = baseline.State
type ServiceType = baseline.ServiceType
type Service = baseline.Service
type Cluster = baseline.Cluster
type Server = baseline.Server

const (
	StateNormal   = baseline.StateNormal
	StateAbnormal = baseline.StateAbnormal
	AS            = baseline.ASType
	TS            = baseline.TSType
	SS            = baseline.SSType
	DS            = baseline.DSType
	SvcNormal     = baseline.ServiceNormal
	SvcAbnormal   = baseline.ServiceAbnormal
	SrvNormal     = baseline.ServerNormal
	SrvAbnormal   = baseline.ServerAbnormal
	Leader        = baseline.Leader
	Follower      = baseline.Follower
)
