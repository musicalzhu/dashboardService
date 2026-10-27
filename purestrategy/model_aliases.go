package purestrategy

import "dashboardService/baseline"

// 分组实现复用共享领域模型和常量；只有策略与算法本身在本 package 独立实现。
type State = baseline.State
type ServiceType = baseline.ServiceType
type Role = baseline.ServiceRole
type Service = baseline.Service
type Cluster = baseline.Cluster
type ClusterGroup = baseline.ClusterGroup
type ClusterRole = baseline.ClusterRole

const (
	StateNormal   = baseline.StateNormal
	StateAbnormal = baseline.StateAbnormal
	AS            = baseline.ASType
	TS            = baseline.TSType
	SS            = baseline.SSType
	DS            = baseline.DSType
	Leader        = baseline.Leader
	Follower      = baseline.Follower
	LSS           = baseline.LSS
	RoleSS        = baseline.SSRole
	RoleDS        = baseline.DSRole
)
