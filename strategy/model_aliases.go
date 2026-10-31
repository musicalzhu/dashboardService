// Package strategy 实现 POC 方案六：纯策略模式（Pure Strategy）。
//
// 核心思想：
//   - 每个分组规则是独立类型，实现 Strategy 接口。
//   - 每个 Group 方法完整实现"集群分桶 → 组内分桶 → 排序 → 组装"。
//   - 策略之间不共享骨架（这是"纯策略"的定义）。
package strategy

import "dashboardService/baseline"

// State 是 baseline.State 的别名。
type State = baseline.State

// ServiceType 是 baseline.ServiceType 的别名。
type ServiceType = baseline.ServiceType

// Role 是 baseline.ServiceRole 的别名。
type Role = baseline.ServiceRole

// Service 是 baseline.Service 的别名。
type Service = baseline.Service

// Cluster 是 baseline.Cluster 的别名。
type Cluster = baseline.Cluster

// ClusterGroup 是 baseline.ClusterGroup 的别名。
type ClusterGroup = baseline.ClusterGroup

// Server 是 baseline.Server 的别名。
type Server = baseline.Server

// Group 是 baseline.Group 的别名。
type Group = baseline.Group

// GroupRule 复用 baseline 的规则类型。
type GroupRule = baseline.GroupRule

// GroupRule 常量别名。
const (
	RuleExactCombination  = baseline.RuleCombination
	RuleExactServiceCount = baseline.RuleServiceCount
	RuleExactPrimary      = baseline.RulePrimary
	RuleExactConvention   = baseline.RuleConvention
	RulePrimaryService    = baseline.RulePrimaryService
)
