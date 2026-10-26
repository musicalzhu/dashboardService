package purestrategy

import "dashboardService/baseline"

// 分组实现复用共享领域模型和常量；只有策略与算法本身在本 package 独立实现。
type State = baseline.State
type ServiceType = baseline.ServiceType
type Role = baseline.Role
type Service = baseline.Service

const (
	StateNormal   = baseline.StateNormal
	StateAbnormal = baseline.StateAbnormal
	AS            = baseline.AS
	TS            = baseline.TS
	SS            = baseline.SS
	DS            = baseline.DS
	Leader        = baseline.Leader
	Follower      = baseline.Follower
	LSS           = baseline.LSS
	RoleSS        = baseline.RoleSS
	RoleDS        = baseline.RoleDS
)
