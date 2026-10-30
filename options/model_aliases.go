package options

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

// GroupRule 复用 baseline 的规则类型，使三个 POC 方案的规则标识保持一致。
type GroupRule = baseline.GroupRule

// GroupRule 常量别名：与 datadriven / baseline 一致，便于三个方案对比。
const (
	// RuleExactCombination 精确组合 + mask 降序。
	RuleExactCombination = baseline.RuleCombination
	// RuleExactServiceCount 精确组合 + 服务数量升序。
	RuleExactServiceCount = baseline.RuleServiceCount
	// RuleExactPrimary 精确组合 + 主服务优先。
	RuleExactPrimary = baseline.RulePrimary
	// RuleExactConvention 精确组合 + 业务约定顺序。
	RuleExactConvention = baseline.RuleConvention
	// RulePrimaryService 主服务归组 + mask 降序。
	RulePrimaryService = baseline.RulePrimaryService
)
