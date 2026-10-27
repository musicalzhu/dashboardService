package options

import "dashboardService/baseline"

// Group 是 baseline 共享的分组结果类型。
type Group = baseline.Group

// ClusterGroup 是 baseline 共享的前端集群分组类型。
type ClusterGroup = baseline.ClusterGroup

// GroupRule 是 baseline 共享的规则标识类型。
type GroupRule = baseline.GroupRule

// State、ServiceType、Service 和 Server 复用 baseline 对应的领域类型。
type State = baseline.State
type ServiceType = baseline.ServiceType
type Service = baseline.Service
type Server = baseline.Server

const (
	RuleExactCombination  = baseline.RuleExactCombination
	RuleExactServiceCount = baseline.RuleExactServiceCount
	RuleExactPrimary      = baseline.RuleExactPrimary
	RuleExactConvention   = baseline.RuleExactConvention
	RulePrimaryService    = baseline.RulePrimaryService

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
)
