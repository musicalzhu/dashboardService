package closure

import "dashboardService/baseline"

// Group、Server、Service 和相关领域类型复用 baseline 的共享模型。
type Group = baseline.Group
type Role = baseline.Role
type State = baseline.State
type ServiceType = baseline.ServiceType
type Service = baseline.Service
type Cluster = baseline.Cluster
type Server = baseline.Server

const (
	StateNormal   = baseline.StateNormal
	StateAbnormal = baseline.StateAbnormal
	AS            = baseline.AS
	TS            = baseline.TS
	SS            = baseline.SS
	DS            = baseline.DS
	SvcNormal     = baseline.SvcNormal
	SvcAbnormal   = baseline.SvcAbnormal
	SrvNormal     = baseline.SrvNormal
	SrvAbnormal   = baseline.SrvAbnormal
	Leader        = baseline.Leader
	Follower      = baseline.Follower
)
