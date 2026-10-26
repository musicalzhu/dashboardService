package baseline

import "dashboardService/model"

// State、ServiceType、Service 和 Server 直接复用共享领域模型，
// 使 baseline 与 datadriven 使用完全相同的输入结构。
type State = model.State
type ServiceType = model.ServiceType
type Role = model.Role
type Service = model.Service
type Cluster = model.Cluster
type Server = model.Server
type ClusterGroup = model.ClusterGroup
type ServiceGroup = model.ServiceGroup
type ClusterFlatGroup = model.ClusterFlatGroup
type ServiceStatus = model.State
type ServerStatus = model.State
type PaxosRole = model.Role

const (
	StateNormal   = model.StateNormal
	StateAbnormal = model.StateAbnormal

	AS = model.ASType
	TS = model.TSType
	SS = model.SSType
	DS = model.DSType

	NoServiceType      = model.NoServiceType
	UnknownServiceType = model.UnknownServiceType

	SvcNormal   = model.StateNormal
	SvcAbnormal = model.StateAbnormal
	SrvNormal   = model.StateNormal
	SrvAbnormal = model.StateAbnormal

	Leader   = model.Leader
	Follower = model.Follower
	LSS      = model.LSS
	RoleSS   = model.SS
	RoleDS   = model.DS
)
