package datadriven

import "dashboardService/model"

type State = model.State
type ServiceType = model.ServiceType
type Role = model.Role
type Service = model.Service
type Cluster = model.Cluster
type Server = model.Server

type ServiceStatus = model.State
type ServerStatus = model.State
type PaxosRole = model.Role

const (
	StateNormal   = model.StateNormal
	StateAbnormal = model.StateAbnormal
	AS            = model.ASType
	TS            = model.TSType
	SS            = model.SSType
	DS            = model.DSType
	SvcNormal     = model.StateNormal
	SvcAbnormal   = model.StateAbnormal
	SrvNormal     = model.StateNormal
	SrvAbnormal   = model.StateAbnormal
	Leader        = model.Leader
	Follower      = model.Follower
)
