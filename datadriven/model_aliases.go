package datadriven

import "dashboardService/baseline"

// State 是 model.State 的别名，表示服务器或服务的运行状态。
type State = baseline.State

// ServiceType 是 model.ServiceType 的别名，表示 AS、TS、SS 或 DS 服务类型。
type ServiceType = baseline.ServiceType

// Role 是 model.Role 的别名，表示微服务角色。
type Role = baseline.Role

// Service 是 model.Service 的别名，复用共享服务实例结构。
type Service = baseline.Service

// Cluster 是 model.Cluster 的别名，复用共享集群元数据结构。
type Cluster = baseline.Cluster

// Server 是 model.Server 的别名，复用共享服务器及其服务列表结构。
type Server = baseline.Server

// ServiceStatus 是兼容 datadriven 领域用语的状态别名，底层仍为 model.State。
type ServiceStatus = baseline.State

// ServerStatus 是兼容 datadriven 领域用语的状态别名，底层仍为 model.State。
type ServerStatus = baseline.State

// PaxosRole 是兼容 datadriven 领域用语的角色别名，底层仍为 model.Role。
type PaxosRole = baseline.Role

const (
	// StateNormal 是 model.StateNormal 的别名。
	StateNormal = baseline.StateNormal
	// StateAbnormal 是 model.StateAbnormal 的别名。
	StateAbnormal = baseline.StateAbnormal
	// AS 是 model.ASType 的短名称。
	AS = baseline.AS
	// TS 是 model.TSType 的短名称。
	TS = baseline.TS
	// SS 是 model.SSType 的短名称。
	SS = baseline.SS
	// DS 是 model.DSType 的短名称。
	DS = baseline.DS
	// SvcNormal 是服务状态正常值的别名。
	SvcNormal = baseline.SvcNormal
	// SvcAbnormal 是服务异常状态值的兼容别名。
	SvcAbnormal = baseline.SvcAbnormal
	// SrvNormal 是服务器状态正常值的别名。
	SrvNormal = baseline.SrvNormal
	// SrvAbnormal 是服务器异常状态值的别名。
	SrvAbnormal = baseline.SrvAbnormal
	// Leader 是 model.Leader 的别名。
	Leader = baseline.Leader
	// Follower 是 model.Follower 的别名。
	Follower = baseline.Follower
)
