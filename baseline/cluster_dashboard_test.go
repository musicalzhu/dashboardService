package baseline

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestRenderClusterDashboardGroupsByClusterAndPrimaryService(t *testing.T) {
	servers := []Server{
		{
			Name: "secondary-ts",
			Cluster: Cluster{
				ClusterID:   20,
				ClusterRole: "secondary",
				ClusterName: "secondary",
			},
			Services: []Service{{ServiceType: TS, State: SvcNormal}},
		},
		{
			Name: "primary-ts",
			Cluster: Cluster{
				ClusterID:   10,
				ClusterRole: "primary",
				ClusterName: "primary",
			},
			Services: []Service{{ServiceType: TS, State: SvcNormal}},
		},
		{
			Name: "primary-as",
			Cluster: Cluster{
				ClusterID:   10,
				ClusterRole: "primary",
				ClusterName: "conflicting-name",
			},
			Services: []Service{
				{ServiceType: TS, State: SvcNormal},
				{ServiceType: AS, State: SvcNormal},
			},
		},
		{
			Name:     "primary-five",
			Cluster:  Cluster{ClusterID: 5, ClusterRole: "primary", ClusterName: "primary-five"},
			Services: []Service{{ServiceType: AS, State: SvcNormal}},
		},
	}

	got := RenderClusterDashboard(servers)
	if len(got) != 3 {
		t.Fatalf("cluster count = %d, want 3", len(got))
	}
	if got[0].ClusterID != 5 || got[1].ClusterID != 10 || got[2].ClusterID != 20 {
		t.Fatalf("cluster order = [%d, %d, %d], want [5, 10, 20]",
			got[0].ClusterID, got[1].ClusterID, got[2].ClusterID)
	}
	if len(got[1].ServiceGroups) != 2 {
		t.Fatalf("primary cluster service group count = %d, want 2", len(got[1].ServiceGroups))
	}
	if got[1].ClusterName != "primary" {
		t.Errorf("cluster metadata = %q, want first-seen value %q", got[1].ClusterName, "primary")
	}
	if got[1].ServiceGroups[0].ServiceType != AS || got[1].ServiceGroups[1].ServiceType != TS {
		t.Fatalf("primary cluster service order = [%s, %s], want [AS, TS]",
			got[1].ServiceGroups[0].ServiceType, got[1].ServiceGroups[1].ServiceType)
	}
	if got[1].ServiceGroups[0].Servers[0].Name != "primary-as" {
		t.Fatalf("AS group server = %q, want primary-as", got[1].ServiceGroups[0].Servers[0].Name)
	}
}

func TestRenderClusterViewsRetainEmptyAndUnknownServers(t *testing.T) {
	servers := []Server{
		{
			Name:     "mixed",
			Cluster:  Cluster{ClusterID: 7, ClusterRole: "primary"},
			Services: []Service{{ServiceType: AS, State: SvcNormal}, {ServiceType: ServiceType("XX"), State: SvcNormal}},
		},
		{Name: "nil-services", Cluster: Cluster{ClusterID: 7, ClusterRole: "primary"}},
		{Name: "empty-services", Cluster: Cluster{ClusterID: 7, ClusterRole: "primary"}, Services: []Service{}},
	}

	nested := RenderClusterDashboard(servers)
	if len(nested) != 1 || len(nested[0].ServiceGroups) != 2 {
		t.Fatalf("nested result = %#v, want one cluster with unknown and no-service groups", nested)
	}
	if nested[0].ServiceGroups[0].ServiceType != UnknownServiceType ||
		len(nested[0].ServiceGroups[0].Servers) != 1 ||
		nested[0].ServiceGroups[0].Servers[0].Name != "mixed" {
		t.Errorf("unknown service group = %#v", nested[0].ServiceGroups[0])
	}
	if nested[0].ServiceGroups[1].ServiceType != NoServiceType ||
		len(nested[0].ServiceGroups[1].Servers) != 2 {
		t.Errorf("no-service group = %#v, want both empty service servers", nested[0].ServiceGroups[1])
	}

	flat := RenderClusterFlatDashboard(servers)
	if len(flat) != 1 || len(flat[0].UnknownServices) != 1 || len(flat[0].NoService) != 2 {
		t.Fatalf("flat result = %#v, want unknown=1 and no-service=2", flat)
	}
	total := len(flat[0].UnknownServices) + len(flat[0].NoService)
	for _, group := range [][]Server{flat[0].AS, flat[0].TS, flat[0].SS, flat[0].DS} {
		total += len(group)
	}
	if total != len(servers) {
		t.Errorf("flat view retained %d servers, want %d", total, len(servers))
	}

	encoded, err := json.Marshal(flat[0])
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	if !bytes.Contains(encoded, []byte(`"unknown_services":[`)) || !bytes.Contains(encoded, []byte(`"no_service":[`)) {
		t.Errorf("flat JSON omitted special server groups: %s", encoded)
	}
}

func TestRenderClusterDashboardDoesNotMutateServiceOrder(t *testing.T) {
	servers := []Server{{
		Name:    "node",
		Cluster: Cluster{ClusterID: 1, ClusterRole: "primary"},
		Services: []Service{
			{ServiceType: TS, State: SvcNormal},
			{ServiceType: AS, State: SvcNormal},
		},
	}}
	original := append([]Service(nil), servers[0].Services...)

	result := RenderClusterDashboard(servers)
	if !reflect.DeepEqual(servers[0].Services, original) {
		t.Fatalf("input Services mutated: got %#v, want %#v", servers[0].Services, original)
	}
	if got := result[0].ServiceGroups[0].Servers[0].Services[0].ServiceType; got != AS {
		t.Errorf("output service order begins with %s, want AS", got)
	}
}

func TestRenderClusterDashboardSortsClusterRolesAndUnknownLast(t *testing.T) {
	servers := []Server{
		{Name: "unknown", Cluster: Cluster{ClusterID: 1, ClusterRole: "other"}, Services: []Service{{ServiceType: AS}}},
		{Name: "arbiter", Cluster: Cluster{ClusterID: 4, ClusterRole: "arbiter"}, Services: []Service{{ServiceType: AS}}},
		{Name: "secondary", Cluster: Cluster{ClusterID: 3, ClusterRole: "secondary"}, Services: []Service{{ServiceType: AS}}},
		{Name: "primary", Cluster: Cluster{ClusterID: 8, ClusterRole: "primary"}, Services: []Service{{ServiceType: AS}}},
	}

	groups := RenderClusterDashboard(servers)
	got := make([]string, len(groups))
	for i, group := range groups {
		got[i] = group.ClusterRole
	}
	want := []string{"primary", "secondary", "arbiter", "other"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cluster role order = %v, want %v", got, want)
	}
}
