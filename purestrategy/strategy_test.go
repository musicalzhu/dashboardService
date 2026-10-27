package purestrategy_test

import (
	"reflect"
	"strings"
	"testing"

	"dashboardService/baseline"
	"dashboardService/purestrategy"
)

func TestGroupersMatchBaseline(t *testing.T) {
	servers := allServiceCombinations()
	servers = append(servers,
		baseline.Server{
			Name: "unknown",
			Services: []baseline.Service{
				{ServiceType: baseline.ServiceType("XX"), State: baseline.ServiceNormal},
			},
		},
		baseline.Server{
			Name: "mixed-unknown",
			Services: []baseline.Service{
				{ServiceType: baseline.ASType, State: baseline.ServiceNormal},
				{ServiceType: baseline.ServiceType("XX"), State: baseline.ServiceNormal},
			},
		},
		baseline.Server{Name: "no-service"},
		baseline.Server{
			Name:  "as-normal-follower",
			IP:    "10.0.0.1",
			State: baseline.ServerNormal,
			Services: []baseline.Service{
				{ServiceType: baseline.ASType, Role: baseline.Follower},
			},
		},
		baseline.Server{
			Name:  "as-abnormal-leader",
			IP:    "10.0.0.2",
			State: baseline.ServerAbnormal,
			Services: []baseline.Service{
				{ServiceType: baseline.ASType, Role: baseline.Leader},
			},
		},
	)

	rules := []baseline.GroupRule{
		baseline.RuleExactCombination,
		baseline.RuleExactServiceCount,
		baseline.RuleExactPrimary,
		baseline.RuleExactConvention,
		baseline.RulePrimaryService,
	}

	for _, rule := range rules {
		t.Run(rule.String(), func(t *testing.T) {
			got := purestrategy.RenderDashboard(servers, rule)
			want := baseline.RenderDashboard(servers, rule)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("pure strategy differs from baseline\ngot:  %#v\nwant: %#v", got, want)
			}
		})
	}
}

func allServiceCombinations() []baseline.Server {
	serviceTypes := []baseline.ServiceType{baseline.ASType, baseline.TSType, baseline.SSType, baseline.DSType}
	servers := make([]baseline.Server, 0, (1<<len(serviceTypes))-1)
	for mask := 1; mask < 1<<len(serviceTypes); mask++ {
		var services []baseline.Service
		var names []string
		for bit, serviceType := range serviceTypes {
			if mask&(1<<(len(serviceTypes)-1-bit)) == 0 {
				continue
			}
			services = append(services, baseline.Service{
				ServiceType: serviceType,
				State:       baseline.StateNormal,
			})
			names = append(names, string(serviceType))
		}
		servers = append(servers, baseline.Server{
			Name: strings.Join(names, "+"),
			Cluster: baseline.Cluster{
				ClusterID: mask%2 + 1,
				ClusterRole: []baseline.ClusterRole{
					baseline.PrimaryClusterRole,
					baseline.SecondaryClusterRole,
				}[mask%2],
			},
			IP:       "1.1.1.1",
			State:    baseline.StateNormal,
			Services: services,
		})
	}
	return servers
}

func TestStrategiesDoNotMutateServiceInput(t *testing.T) {
	for rule, grouper := range purestrategy.Groupers {
		t.Run(rule.String(), func(t *testing.T) {
			servers := []baseline.Server{{
				Name: "node",
				Services: []baseline.Service{
					{ServiceType: baseline.TSType, Role: baseline.Follower},
					{ServiceType: baseline.ASType, Role: baseline.Leader},
				},
			}}
			original := append([]baseline.Service(nil), servers[0].Services...)
			grouper.Group(servers)
			if !reflect.DeepEqual(servers[0].Services, original) {
				t.Fatalf("strategy mutated input services: got %#v, want %#v", servers[0].Services, original)
			}
		})
	}
}

func TestEmptyInputReturnsEmptyGroups(t *testing.T) {
	for rule := range purestrategy.Groupers {
		groups := purestrategy.RenderDashboard(nil, rule)
		if groups == nil || len(groups) != 0 {
			t.Errorf("%s returned %#v for nil input, want non-nil empty slice", rule, groups)
		}
	}
}

func TestUnknownRulePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("invalid rule should panic")
		}
	}()
	purestrategy.RenderDashboard(nil, baseline.GroupRule(999))
}
