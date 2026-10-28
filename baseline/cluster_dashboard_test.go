package baseline

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRenderDashboardGroupsByClusterAndPrimaryService(t *testing.T) {
	servers := []Server{
		{
			Name:     "secondary-ts",
			Cluster:  Cluster{ClusterID: 20, ClusterRole: SecondaryClusterRole, ClusterName: "secondary"},
			Services: []Service{{ServiceType: TSType, State: ServiceNormal}},
		},
		{
			Name:     "primary-ts",
			Cluster:  Cluster{ClusterID: 10, ClusterRole: PrimaryClusterRole, ClusterName: "primary"},
			Services: []Service{{ServiceType: TSType, State: ServiceNormal}},
		},
		{
			Name:     "primary-as",
			Cluster:  Cluster{ClusterID: 10, ClusterRole: PrimaryClusterRole, ClusterName: "conflicting-name"},
			Services: []Service{{ServiceType: TSType, State: ServiceNormal}, {ServiceType: ASType, State: ServiceNormal}},
		},
		{
			Name:     "primary-five",
			Cluster:  Cluster{ClusterID: 5, ClusterRole: PrimaryClusterRole, ClusterName: "primary-five"},
			Services: []Service{{ServiceType: ASType, State: ServiceNormal}},
		},
	}

	got := RenderDashboard(servers, RulePrimaryService)
	if len(got) != 3 {
		t.Fatalf("cluster count = %d, want 3", len(got))
	}
	if got[0].ClusterID != 5 || got[1].ClusterID != 10 || got[2].ClusterID != 20 {
		t.Fatalf("cluster order = [%d, %d, %d], want [5, 10, 20]",
			got[0].ClusterID, got[1].ClusterID, got[2].ClusterID)
	}
	if len(got[1].Groups) != 2 {
		t.Fatalf("primary cluster group count = %d, want 2", len(got[1].Groups))
	}
	if got[1].ClusterName != "primary" {
		t.Errorf("cluster metadata = %q, want first-seen value %q", got[1].ClusterName, "primary")
	}
	if got[1].Groups[0].Name != "AS" || got[1].Groups[1].Name != "TS" {
		t.Fatalf("primary cluster group order = [%s, %s], want [AS, TS]",
			got[1].Groups[0].Name, got[1].Groups[1].Name)
	}
	if got[1].Groups[0].Servers[0].Name != "primary-as" {
		t.Fatalf("AS group server = %q, want primary-as", got[1].Groups[0].Servers[0].Name)
	}
}

func TestRenderDashboardRetainsEmptyAndUnknownServers(t *testing.T) {
	servers := []Server{
		{
			Name:     "mixed",
			Cluster:  Cluster{ClusterID: 7, ClusterRole: PrimaryClusterRole},
			Services: []Service{{ServiceType: ASType, State: ServiceNormal}, {ServiceType: ServiceType("XX"), State: ServiceNormal}},
		},
		{Name: "nil-services", Cluster: Cluster{ClusterID: 7, ClusterRole: PrimaryClusterRole}},
		{Name: "empty-services", Cluster: Cluster{ClusterID: 7, ClusterRole: PrimaryClusterRole}, Services: []Service{}},
	}

	nested := RenderDashboard(servers, RulePrimaryService)
	if len(nested) != 1 || len(nested[0].Groups) != 2 {
		t.Fatalf("nested result = %#v, want one cluster with unknown and no-service groups", nested)
	}
	if nested[0].Groups[0].Name != unknownServiceGroupName ||
		len(nested[0].Groups[0].Servers) != 1 ||
		nested[0].Groups[0].Servers[0].Name != "mixed" {
		t.Errorf("unknown service group = %#v", nested[0].Groups[0])
	}
	if nested[0].Groups[1].Name != noServiceGroupName ||
		len(nested[0].Groups[1].Servers) != 2 {
		t.Errorf("no-service group = %#v, want both empty service servers", nested[0].Groups[1])
	}

	encoded, err := MarshalDashboard(servers, RulePrimaryService, JSONFlat)
	if err != nil {
		t.Fatalf("MarshalDashboard() error: %v", err)
	}
	var flat []struct {
		ClusterID int      `json:"cluster_id"`
		Group     string   `json:"group"`
		Servers   []Server `json:"servers"`
	}
	if err := json.Unmarshal(encoded, &flat); err != nil {
		t.Fatalf("decode flat JSON: %v", err)
	}
	if len(flat) != 2 || flat[0].Group != unknownServiceGroupName || flat[1].Group != noServiceGroupName {
		t.Fatalf("flat rows = %#v, want unknown then no-service rows", flat)
	}
	total := 0
	for _, row := range flat {
		total += len(row.Servers)
	}
	if total != len(servers) {
		t.Errorf("flat view retained %d servers, want %d", total, len(servers))
	}
}

func TestRenderDashboardAppliesAllRules(t *testing.T) {
	servers := []Server{
		{Name: "as-ts", Cluster: Cluster{ClusterID: 1}, Services: []Service{{ServiceType: ASType}, {ServiceType: TSType}}},
		{Name: "as", Cluster: Cluster{ClusterID: 1}, Services: []Service{{ServiceType: ASType}}},
		{Name: "ts", Cluster: Cluster{ClusterID: 1}, Services: []Service{{ServiceType: TSType}}},
	}
	rules := []GroupRule{
		RuleCombination,
		RuleServiceCount,
		RulePrimary,
		RuleConvention,
		RulePrimaryService,
	}
	for _, rule := range rules {
		t.Run(rule.String(), func(t *testing.T) {
			got := RenderDashboard(servers, rule)
			want := RenderDashboard(servers, rule)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("cluster result = %#v, want %#v", got, want)
			}
		})
	}
}

func TestRenderDashboardDoesNotMutateServiceOrder(t *testing.T) {
	servers := []Server{{
		Name:    "node",
		Cluster: Cluster{ClusterID: 1, ClusterRole: PrimaryClusterRole},
		Services: []Service{
			{ServiceType: TSType, State: ServiceNormal},
			{ServiceType: ASType, State: ServiceNormal},
		},
	}}
	original := append([]Service(nil), servers[0].Services...)

	result := RenderDashboard(servers, RulePrimaryService)
	if !reflect.DeepEqual(servers[0].Services, original) {
		t.Fatalf("input Services mutated: got %#v, want %#v", servers[0].Services, original)
	}
	if got := result[0].Groups[0].Servers[0].Services[0].ServiceType; got != ASType {
		t.Errorf("output service order begins with %s, want AS", got)
	}
}

func TestRenderDashboardSortsClusterRolesAndUnknownLast(t *testing.T) {
	servers := []Server{
		{Name: "unknown", Cluster: Cluster{ClusterID: 1, ClusterRole: ClusterRole("other")}, Services: []Service{{ServiceType: ASType}}},
		{Name: "secondary", Cluster: Cluster{ClusterID: 3, ClusterRole: SecondaryClusterRole}, Services: []Service{{ServiceType: ASType}}},
		{Name: "primary", Cluster: Cluster{ClusterID: 8, ClusterRole: PrimaryClusterRole}, Services: []Service{{ServiceType: ASType}}},
	}

	groups := RenderDashboard(servers, RulePrimaryService)
	got := make([]ClusterRole, len(groups))
	for i, group := range groups {
		got[i] = group.ClusterRole
	}
	want := []ClusterRole{PrimaryClusterRole, SecondaryClusterRole, ClusterRole("other")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cluster role order = %v, want %v", got, want)
	}
}
