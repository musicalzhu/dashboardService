package purestrategy_test

import (
	"reflect"
	"testing"

	"dashboardService/baseline"
	"dashboardService/purestrategy"
)

func TestGroupersMatchBaseline(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	servers = append(servers,
		baseline.Server{
			Name: "unknown",
			Services: []baseline.Service{
				{ServiceType: baseline.ServiceType("XX"), State: baseline.SvcNormal},
			},
		},
		baseline.Server{
			Name: "mixed-unknown",
			Services: []baseline.Service{
				{ServiceType: baseline.AS, State: baseline.SvcNormal},
				{ServiceType: baseline.ServiceType("XX"), State: baseline.SvcNormal},
			},
		},
		baseline.Server{Name: "no-service"},
		baseline.Server{
			Name:  "as-normal-follower",
			IP:    "10.0.0.1",
			State: baseline.SrvNormal,
			Services: []baseline.Service{
				{ServiceType: baseline.AS, Role: baseline.Follower},
			},
		},
		baseline.Server{
			Name:  "as-abnormal-leader",
			IP:    "10.0.0.2",
			State: baseline.SrvAbnormal,
			Services: []baseline.Service{
				{ServiceType: baseline.AS, Role: baseline.Leader},
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

func TestStrategiesDoNotMutateServiceInput(t *testing.T) {
	for rule, grouper := range purestrategy.Groupers {
		t.Run(rule.String(), func(t *testing.T) {
			servers := []baseline.Server{{
				Name: "node",
				Services: []baseline.Service{
					{ServiceType: baseline.TS, Role: baseline.Follower},
					{ServiceType: baseline.AS, Role: baseline.Leader},
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
