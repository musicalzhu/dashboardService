package options

import "dashboardService/baseline"

// Group 是 baseline 共享的分组结果类型。
type Group = baseline.Group

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
	AS            = baseline.AS
	TS            = baseline.TS
	SS            = baseline.SS
	DS            = baseline.DS
	SvcNormal     = baseline.SvcNormal
	SvcAbnormal   = baseline.SvcAbnormal
	SrvNormal     = baseline.SrvNormal
	SrvAbnormal   = baseline.SrvAbnormal
)
