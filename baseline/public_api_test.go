package baseline_test

import (
	"encoding/json"
	"testing"

	"dashboardService/baseline"
)

func TestMarshalGroupsIsPublic(t *testing.T) {
	data, err := baseline.MarshalGroups([]baseline.Group{{
		Name:    "AS",
		Servers: []baseline.Server{},
	}})
	if err != nil {
		t.Fatalf("MarshalGroups() error: %v", err)
	}
	if string(data) != "[\n  {\n    \"name\": \"AS\",\n    \"servers\": []\n  }\n]" {
		t.Fatalf("MarshalGroups() = %s", data)
	}
}

func TestMarshalDashboardFormatsPreserveRuleOrder(t *testing.T) {
	servers := []baseline.Server{
		{Name: "as-ts", Services: []baseline.Service{{ServiceType: baseline.ASType}, {ServiceType: baseline.TSType}}},
		{Name: "ts", Services: []baseline.Service{{ServiceType: baseline.TSType}}},
		{Name: "as", Services: []baseline.Service{{ServiceType: baseline.ASType}}},
	}
	rules := []baseline.GroupRule{
		baseline.RuleExactCombination,
		baseline.RuleExactServiceCount,
		baseline.RuleExactPrimary,
		baseline.RuleExactConvention,
		baseline.RulePrimaryService,
	}

	for _, rule := range rules {
		t.Run(rule.String(), func(t *testing.T) {
			want := baseline.RenderDashboard(servers, rule)

			nestedJSON, err := baseline.MarshalDashboard(servers, rule, baseline.JSONNested)
			if err != nil {
				t.Fatalf("MarshalDashboard(nested) error: %v", err)
			}
			var nested []baseline.ClusterGroup
			if err := json.Unmarshal(nestedJSON, &nested); err != nil {
				t.Fatalf("decode nested JSON: %v", err)
			}
			if len(nested) != len(want) || len(nested[0].Groups) != len(want[0].Groups) {
				t.Fatalf("nested cluster count/groups = %#v, want %#v", nested, want)
			}
			for i := range want[0].Groups {
				if nested[0].Groups[i].Name != want[0].Groups[i].Name ||
					len(nested[0].Groups[i].Servers) != len(want[0].Groups[i].Servers) {
					t.Errorf("nested group[%d] = %#v, want %q with %d servers",
						i, nested[0].Groups[i], want[0].Groups[i].Name, len(want[0].Groups[i].Servers))
				}
			}

			flatJSON, err := baseline.MarshalDashboard(servers, rule, baseline.JSONFlat)
			if err != nil {
				t.Fatalf("MarshalDashboard(flat) error: %v", err)
			}
			var flat []struct {
				Name    string            `json:"group"`
				Servers []baseline.Server `json:"servers"`
			}
			if err := json.Unmarshal(flatJSON, &flat); err != nil {
				t.Fatalf("decode flat JSON: %v", err)
			}
			if len(flat) != len(want[0].Groups) {
				t.Fatalf("flat group count = %d, want %d", len(flat), len(want[0].Groups))
			}
			for i := range want[0].Groups {
				if flat[i].Name != want[0].Groups[i].Name ||
					len(flat[i].Servers) != len(want[0].Groups[i].Servers) {
					t.Errorf("flat group[%d] = %#v, want name %q with %d servers",
						i, flat[i], want[0].Groups[i].Name, len(want[0].Groups[i].Servers))
				}
			}
		})
	}
}

func TestMarshalDashboardUsesSelectedRuleAndFlatRows(t *testing.T) {
	servers := []baseline.Server{
		{
			Name:     "as-ts",
			Cluster:  baseline.Cluster{ClusterID: 2, ClusterRole: "primary", ClusterName: "cluster"},
			Services: []baseline.Service{{ServiceType: baseline.ASType}, {ServiceType: baseline.TSType}},
		},
		{
			Name:     "ts",
			Cluster:  baseline.Cluster{ClusterID: 2, ClusterRole: "primary", ClusterName: "ignored-name"},
			Services: []baseline.Service{{ServiceType: baseline.TSType}},
		},
	}
	want := baseline.RenderDashboard(servers, baseline.RuleExactCombination)[0].Groups

	nestedJSON, err := baseline.MarshalDashboard(servers, baseline.RuleExactCombination, baseline.JSONNested)
	if err != nil {
		t.Fatalf("MarshalDashboard(nested) error: %v", err)
	}
	var nested []baseline.ClusterGroup
	if err := json.Unmarshal(nestedJSON, &nested); err != nil {
		t.Fatalf("decode nested cluster JSON: %v", err)
	}
	if len(nested) != 1 || len(nested[0].Groups) != len(want) {
		t.Fatalf("nested cluster JSON = %#v, want one cluster with %d groups", nested, len(want))
	}
	for i := range want {
		if nested[0].Groups[i].Name != want[i].Name {
			t.Errorf("nested cluster group[%d] = %q, want %q",
				i, nested[0].Groups[i].Name, want[i].Name)
		}
	}
	if nested[0].ClusterName != "cluster" {
		t.Errorf("cluster metadata = %q, want first-seen value", nested[0].ClusterName)
	}

	flatJSON, err := baseline.MarshalDashboard(servers, baseline.RuleExactCombination, baseline.JSONFlat)
	if err != nil {
		t.Fatalf("MarshalDashboard(flat) error: %v", err)
	}
	var flat []struct {
		ClusterID int               `json:"cluster_id"`
		GroupName string            `json:"group"`
		Servers   []baseline.Server `json:"servers"`
	}
	if err := json.Unmarshal(flatJSON, &flat); err != nil {
		t.Fatalf("decode flat cluster JSON: %v", err)
	}
	if len(flat) != len(want) {
		t.Fatalf("flat cluster row count = %d, want %d", len(flat), len(want))
	}
	for i := range want {
		if flat[i].GroupName != want[i].Name || flat[i].ClusterID != 2 {
			t.Errorf("flat cluster row[%d] = %#v, want group %q in cluster 2", i, flat[i], want[i].Name)
		}
	}
}

func TestMarshalDashboardSupportsAllRules(t *testing.T) {
	servers := []baseline.Server{
		{Name: "as-ts", Cluster: baseline.Cluster{ClusterID: 1}, Services: []baseline.Service{{ServiceType: baseline.ASType}, {ServiceType: baseline.TSType}}},
		{Name: "as-ss", Cluster: baseline.Cluster{ClusterID: 1}, Services: []baseline.Service{{ServiceType: baseline.ASType}, {ServiceType: baseline.SSType}}},
		{Name: "ts", Cluster: baseline.Cluster{ClusterID: 1}, Services: []baseline.Service{{ServiceType: baseline.TSType}}},
		{Name: "unknown", Cluster: baseline.Cluster{ClusterID: 1}, Services: []baseline.Service{{ServiceType: baseline.ServiceType("XX")}}},
		{Name: "no-service", Cluster: baseline.Cluster{ClusterID: 1}},
	}
	rules := []baseline.GroupRule{
		baseline.RuleExactCombination,
		baseline.RuleExactServiceCount,
		baseline.RuleExactPrimary,
		baseline.RuleExactConvention,
		baseline.RulePrimaryService,
	}

	for _, rule := range rules {
		t.Run(rule.String(), func(t *testing.T) {
			want := baseline.RenderDashboard(servers, rule)[0].Groups

			data, err := baseline.MarshalDashboard(servers, rule, baseline.JSONNested)
			if err != nil {
				t.Fatalf("MarshalDashboard(nested) error: %v", err)
			}
			var nested []baseline.ClusterGroup
			if err := json.Unmarshal(data, &nested); err != nil {
				t.Fatalf("decode nested JSON: %v", err)
			}
			if len(nested) != 1 || len(nested[0].Groups) != len(want) {
				t.Fatalf("nested cluster groups = %#v, want %d groups", nested, len(want))
			}
			for i := range want {
				if nested[0].Groups[i].Name != want[i].Name ||
					len(nested[0].Groups[i].Servers) != len(want[i].Servers) {
					t.Errorf("nested group[%d] = %#v, want %q with %d servers",
						i, nested[0].Groups[i], want[i].Name, len(want[i].Servers))
				}
			}

			data, err = baseline.MarshalDashboard(servers, rule, baseline.JSONFlat)
			if err != nil {
				t.Fatalf("MarshalDashboard(flat) error: %v", err)
			}
			var flat []struct {
				GroupName string            `json:"group"`
				Servers   []baseline.Server `json:"servers"`
			}
			if err := json.Unmarshal(data, &flat); err != nil {
				t.Fatalf("decode flat JSON: %v", err)
			}
			if len(flat) != len(want) {
				t.Fatalf("flat cluster row count = %d, want %d", len(flat), len(want))
			}
			for i := range want {
				if flat[i].GroupName != want[i].Name || len(flat[i].Servers) != len(want[i].Servers) {
					t.Errorf("flat group[%d] = %#v, want %q with %d servers",
						i, flat[i], want[i].Name, len(want[i].Servers))
				}
			}
		})
	}
}

func TestMarshalDashboardRejectsInvalidFormat(t *testing.T) {
	if _, err := baseline.MarshalDashboard(nil, baseline.RuleExactCombination, baseline.JSONFormat(99)); err == nil {
		t.Fatal("MarshalDashboard() with invalid format returned nil error")
	}
}
