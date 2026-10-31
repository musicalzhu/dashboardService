// Package options 实现 POC 方案三：函数式选项（Functional Options）。
//
// 核心思想：
//   - 分组规则由一系列 Option 函数构造，而不是从预定义表中查表：
//     RenderDashboard(servers, WithKeyOf(...), WithCmp(...))。
//   - Option 是 func(*baseline.RuleImpl)，每个 Option 修改 RuleImpl 的一个字段。
//   - 调用方可以自由组合 keyOf 和 cmp，构造出 ruleImpls 表中不存在的规则组合
//     （如"主服务归组 + 服务数量升序"）。
//   - 所有分组、排序、序列化行为仍由 baseline 包实现和测试，本包只负责"构造配置"。
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
