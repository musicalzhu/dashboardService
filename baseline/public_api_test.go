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
		{Name: "as-ts", Services: []baseline.Service{{ServiceType: baseline.AS}, {ServiceType: baseline.TS}}},
		{Name: "ts", Services: []baseline.Service{{ServiceType: baseline.TS}}},
		{Name: "as", Services: []baseline.Service{{ServiceType: baseline.AS}}},
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
			var nested []baseline.Group
			if err := json.Unmarshal(nestedJSON, &nested); err != nil {
				t.Fatalf("decode nested JSON: %v", err)
			}
			if len(nested) != len(want) {
				t.Fatalf("nested group count = %d, want %d", len(nested), len(want))
			}
			for i := range want {
				if nested[i].Name != want[i].Name || len(nested[i].Servers) != len(want[i].Servers) {
					t.Errorf("nested group[%d] = %#v, want %q with %d servers",
						i, nested[i], want[i].Name, len(want[i].Servers))
				}
			}

			flatJSON, err := baseline.MarshalDashboard(servers, rule, baseline.JSONFlat)
			if err != nil {
				t.Fatalf("MarshalDashboard(flat) error: %v", err)
			}
			var flat []baseline.FlatGroup
			if err := json.Unmarshal(flatJSON, &flat); err != nil {
				t.Fatalf("decode flat JSON: %v", err)
			}
			if len(flat) != len(want) {
				t.Fatalf("flat group count = %d, want %d", len(flat), len(want))
			}
			for i := range want {
				if flat[i].Name != want[i].Name || len(flat[i].Servers) != len(want[i].Servers) {
					t.Errorf("flat group[%d] = %#v, want name %q with %d servers",
						i, flat[i], want[i].Name, len(want[i].Servers))
				}
			}
		})
	}
}

func TestMarshalClusterDashboardUsesSelectedRuleAndFlatRows(t *testing.T) {
	servers := []baseline.Server{
		{
			Name:     "as-ts",
			Cluster:  baseline.Cluster{ClusterID: 2, ClusterRole: "primary", ClusterName: "cluster"},
			Services: []baseline.Service{{ServiceType: baseline.AS}, {ServiceType: baseline.TS}},
		},
		{
			Name:     "ts",
			Cluster:  baseline.Cluster{ClusterID: 2, ClusterRole: "primary", ClusterName: "ignored-name"},
			Services: []baseline.Service{{ServiceType: baseline.TS}},
		},
	}
	want := baseline.RenderDashboard(servers, baseline.RuleExactCombination)

	nestedJSON, err := baseline.MarshalClusterDashboard(servers, baseline.RuleExactCombination, baseline.JSONNested)
	if err != nil {
		t.Fatalf("MarshalClusterDashboard(nested) error: %v", err)
	}
	var nested []baseline.ClusterDashboard
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

	flatJSON, err := baseline.MarshalClusterDashboard(servers, baseline.RuleExactCombination, baseline.JSONFlat)
	if err != nil {
		t.Fatalf("MarshalClusterDashboard(flat) error: %v", err)
	}
	var flat []baseline.ClusterFlatGroupEntry
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

func TestMarshalClusterDashboardSupportsAllRules(t *testing.T) {
	servers := []baseline.Server{
		{Name: "as-ts", Cluster: baseline.Cluster{ClusterID: 1}, Services: []baseline.Service{{ServiceType: baseline.AS}, {ServiceType: baseline.TS}}},
		{Name: "as-ss", Cluster: baseline.Cluster{ClusterID: 1}, Services: []baseline.Service{{ServiceType: baseline.AS}, {ServiceType: baseline.SS}}},
		{Name: "ts", Cluster: baseline.Cluster{ClusterID: 1}, Services: []baseline.Service{{ServiceType: baseline.TS}}},
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
			want := baseline.RenderDashboard(servers, rule)

			data, err := baseline.MarshalClusterDashboard(servers, rule, baseline.JSONNested)
			if err != nil {
				t.Fatalf("MarshalClusterDashboard(nested) error: %v", err)
			}
			var nested []baseline.ClusterDashboard
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

			data, err = baseline.MarshalClusterDashboard(servers, rule, baseline.JSONFlat)
			if err != nil {
				t.Fatalf("MarshalClusterDashboard(flat) error: %v", err)
			}
			var flat []baseline.ClusterFlatGroupEntry
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
	if _, err := baseline.MarshalClusterDashboard(nil, baseline.RuleExactCombination, baseline.JSONFormat(99)); err == nil {
		t.Fatal("MarshalClusterDashboard() with invalid format returned nil error")
	}
}
