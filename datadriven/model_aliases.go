// Package datadriven 实现 POC 方案二：数据驱动（Data-Driven）。
//
// 核心思想：
//   - 规则不是写在 switch 分支中，而是以"表"的形式集中声明：
//     ruleImpls map[GroupRule]RuleImpl。
//   - 入口 RenderDashboard 从 ruleImpls 表中查一个 RuleImpl，
//     再委托 baseline.GroupServersByCluster 执行。
//   - 与 baseline 的唯一差异是"规则存储方式"：switch 分支 vs map 表项。
//   - 规则的可读性、可枚举性、可测试性都由"表"承载，而不是"分支"。
package datadriven

import "dashboardService/baseline"

// State 是 baseline.State 的别名，表示服务器或服务的运行状态。
type State = baseline.State

// ServiceType 是 baseline.ServiceType 的别名，表示 AS、TS、SS 或 DS 服务类型。
type ServiceType = baseline.ServiceType

// Role 是 baseline.ServiceRole 的别名，表示微服务角色。
type Role = baseline.ServiceRole

// Service 是 baseline.Service 的别名，复用共享服务实例结构。
type Service = baseline.Service

// Cluster 是 baseline.Cluster 的别名，复用共享集群元数据结构。
type Cluster = baseline.Cluster

// ClusterGroup 是 baseline.ClusterGroup 的别名，复用前端分组结构。
type ClusterGroup = baseline.ClusterGroup

// Server 是 baseline.Server 的别名，复用共享服务器及其服务列表结构。
type Server = baseline.Server

// ServiceStatus 是兼容 datadriven 领域用语的状态别名，底层仍为 baseline.State。
type ServiceStatus = baseline.State

// ServerStatus 是兼容 datadriven 领域用语的状态别名，底层仍为 baseline.State。
type ServerStatus = baseline.State

// PaxosRole 是兼容 datadriven 领域用语的角色别名，底层仍为 baseline.ServiceRole。
type PaxosRole = baseline.ServiceRole

const (
	// StateNormal 是 baseline.StateNormal 的别名。
	StateNormal = baseline.StateNormal
	// StateAbnormal 是 baseline.StateAbnormal 的别名。
	StateAbnormal = baseline.StateAbnormal
	// AS 是 baseline.ASType 的短名称。
	AS = baseline.ASType
	// TS 是 baseline.TSType 的短名称。
	TS = baseline.TSType
	// SS 是 baseline.SSType 的短名称。
	SS = baseline.SSType
	// DS 是 baseline.DSType 的短名称。
	DS = baseline.DSType
	// SvcNormal 是服务状态正常值的别名。
	SvcNormal = baseline.ServiceNormal
	// SvcAbnormal 是服务异常状态值的兼容别名。
	SvcAbnormal = baseline.ServiceAbnormal
	// SrvNormal 是服务器状态正常值的别名。
	SrvNormal = baseline.ServerNormal
	// SrvAbnormal 是服务器异常状态值的别名。
	SrvAbnormal = baseline.ServerAbnormal
	// Leader 是 baseline.Leader 的别名。
	Leader = baseline.Leader
	// Follower 是 baseline.Follower 的别名。
	Follower = baseline.Follower
)

// GroupRule 复用 baseline 的规则类型，使各 POC 的规则标识保持一致。
type GroupRule = baseline.GroupRule

// RuleExactCombination 是 datadriven 方案的规则标识，复用 baseline 的规则值。
const (
	RuleExactCombination  = baseline.RuleCombination
	RuleExactServiceCount = baseline.RuleServiceCount
	RuleExactPrimary      = baseline.RulePrimary
	RuleExactConvention   = baseline.RuleConvention
	RulePrimaryService    = baseline.RulePrimaryService
)
