package strategy

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetDashboard(t *testing.T) {
	tests := []struct {
		name    string
		servers []Server
		want    []ClusterGroup
	}{
		{
			name: "单机",
			servers: []Server{
				{
					Name: "node-01",
					Cluster: Cluster{
						ClusterID:   1,
						ClusterRole: "primary",
						ClusterName: "集群A",
					},
					IP:    "192.168.1.101",
					State: StateNormal,
					Services: []Service{
						{
							ServiceType: TSType,
							PaxosGroup:  "group-1",
							Role:        Leader,
							State:       StateNormal,
						},
						{
							ServiceType: DSType,
							PaxosGroup:  "group-1",
							Role:        DS,
							State:       StateNormal,
						},
						{
							ServiceType: ASType,
							PaxosGroup:  "group-1",
							Role:        Leader,
							State:       StateNormal,
						},
						{
							ServiceType: SSType,
							PaxosGroup:  "group-1",
							Role:        LSS,
							State:       StateNormal,
						},
					},
				},
			},
			want: []ClusterGroup{
				{
					Cluster: Cluster{
						ClusterID:   1,
						ClusterRole: "primary",
						ClusterName: "集群A",
					},
					ServiceGroups: []ServiceGroup{
						{
							ServiceType: ASType,
							Servers: []Server{
								{
									Name: "node-01",
									Cluster: Cluster{
										ClusterID:   1,
										ClusterRole: "primary",
										ClusterName: "集群A",
									},
									IP:    "192.168.1.101",
									State: StateNormal,
									Services: []Service{
										{
											ServiceType: ASType,
											PaxosGroup:  "group-1",
											Role:        Leader,
											State:       StateNormal,
										},
										{
											ServiceType: TSType,
											PaxosGroup:  "group-1",
											Role:        Leader,
											State:       StateNormal,
										},
										{
											ServiceType: SSType,
											PaxosGroup:  "group-1",
											Role:        LSS,
											State:       StateNormal,
										},
										{
											ServiceType: DSType,
											PaxosGroup:  "group-1",
											Role:        DS,
											State:       StateNormal,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "三节点",
			servers: []Server{
				{
					Name: "node-01",
					Cluster: Cluster{
						ClusterID:   1,
						ClusterRole: "primary",
						ClusterName: "集群A",
					},
					IP:    "192.168.1.101",
					State: StateNormal,
					Services: []Service{
						{
							ServiceType: ASType,
							PaxosGroup:  "group-1",
							Role:        Leader,
							State:       StateNormal,
						},
						{
							ServiceType: TSType,
							PaxosGroup:  "group-1",
							Role:        Follower,
							State:       StateNormal,
						},
						{
							ServiceType: SSType,
							PaxosGroup:  "group-1",
							Role:        LSS,
							State:       StateNormal,
						},
						{
							ServiceType: DSType,
							PaxosGroup:  "group-1",
							Role:        DS,
							State:       StateNormal,
						},
					},
				},
				{
					Name: "node-02",
					Cluster: Cluster{
						ClusterID:   1,
						ClusterRole: "primary",
						ClusterName: "集群A",
					},
					IP:    "192.168.1.102",
					State: StateNormal,
					Services: []Service{
						{
							ServiceType: ASType,
							PaxosGroup:  "group-1",
							Role:        Follower,
							State:       StateNormal,
						},
						{
							ServiceType: TSType,
							PaxosGroup:  "group-1",
							Role:        Leader,
							State:       StateNormal,
						},
						{
							ServiceType: SSType,
							PaxosGroup:  "group-1",
							Role:        SS,
							State:       StateNormal,
						},
						{
							ServiceType: DSType,
							PaxosGroup:  "group-1",
							Role:        DS,
							State:       StateNormal,
						},
					},
				},

				{
					Name: "node-03",
					Cluster: Cluster{
						ClusterID:   1,
						ClusterRole: "primary",
						ClusterName: "集群A",
					},
					IP:    "192.168.1.103",
					State: StateNormal,
					Services: []Service{
						{
							ServiceType: ASType,
							PaxosGroup:  "group-1",
							Role:        Follower,
							State:       StateNormal,
						},
						{
							ServiceType: TSType,
							PaxosGroup:  "group-1",
							Role:        Follower,
							State:       StateNormal,
						},
						{
							ServiceType: SSType,
							PaxosGroup:  "group-1",
							Role:        SS,
							State:       StateNormal,
						},
						{
							ServiceType: DSType,
							PaxosGroup:  "group-1",
							Role:        DS,
							State:       StateNormal,
						},
					},
				},
			},
			want: []ClusterGroup{
				{
					Cluster: Cluster{
						ClusterID:   1,
						ClusterRole: "primary",
						ClusterName: "集群A",
					},
					ServiceGroups: []ServiceGroup{
						{
							ServiceType: ASType,
							Servers: []Server{
								{
									Name: "node-01",
									Cluster: Cluster{
										ClusterID:   1,
										ClusterRole: "primary",
										ClusterName: "集群A",
									},
									IP:    "192.168.1.101",
									State: StateNormal,
									Services: []Service{
										{
											ServiceType: ASType,
											PaxosGroup:  "group-1",
											Role:        Leader,
											State:       StateNormal,
										},
										{
											ServiceType: TSType,
											PaxosGroup:  "group-1",
											Role:        Follower,
											State:       StateNormal,
										},
										{
											ServiceType: SSType,
											PaxosGroup:  "group-1",
											Role:        LSS,
											State:       StateNormal,
										},
										{
											ServiceType: DSType,
											PaxosGroup:  "group-1",
											Role:        DS,
											State:       StateNormal,
										},
									},
								},
								{
									Name: "node-02",
									Cluster: Cluster{
										ClusterID:   1,
										ClusterRole: "primary",
										ClusterName: "集群A",
									},
									IP:    "192.168.1.102",
									State: StateNormal,
									Services: []Service{
										{
											ServiceType: ASType,
											PaxosGroup:  "group-1",
											Role:        Follower,
											State:       StateNormal,
										},
										{
											ServiceType: TSType,
											PaxosGroup:  "group-1",
											Role:        Leader,
											State:       StateNormal,
										},
										{
											ServiceType: SSType,
											PaxosGroup:  "group-1",
											Role:        SS,
											State:       StateNormal,
										},
										{
											ServiceType: DSType,
											PaxosGroup:  "group-1",
											Role:        DS,
											State:       StateNormal,
										},
									},
								},
								{
									Name: "node-03",
									Cluster: Cluster{
										ClusterID:   1,
										ClusterRole: "primary",
										ClusterName: "集群A",
									},
									IP:    "192.168.1.103",
									State: StateNormal,
									Services: []Service{
										{
											ServiceType: ASType,
											PaxosGroup:  "group-1",
											Role:        Follower,
											State:       StateNormal,
										},
										{
											ServiceType: TSType,
											PaxosGroup:  "group-1",
											Role:        Follower,
											State:       StateNormal,
										},
										{
											ServiceType: SSType,
											PaxosGroup:  "group-1",
											Role:        SS,
											State:       StateNormal,
										},
										{
											ServiceType: DSType,
											PaxosGroup:  "group-1",
											Role:        DS,
											State:       StateNormal,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sorter := &DefaultSorter{}
			got := RenderDashboard(&ServiceGroupStrategy{sorter}, tt.servers)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("GetDashboard() mismatch (-want +got):\n%s", diff)
			}

			// 输出JSON
			jsonData, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				fmt.Printf("JSON编码错误: %v\n", err)
			}
			fmt.Println("服务分组结果:")
			fmt.Println(string(jsonData))
		})
	}
}

func TestDashboardIncludesServersWithoutServices(t *testing.T) {
	servers := []Server{
		{
			Name:     "node-service",
			IP:       "192.168.1.101",
			Cluster:  Cluster{ClusterID: 7, ClusterRole: "primary"},
			Services: []Service{{ServiceType: ASType, State: StateNormal}},
		},
		{
			Name:    "node-empty",
			IP:      "192.168.1.102",
			Cluster: Cluster{ClusterID: 7, ClusterRole: "primary"},
		},
		{
			Name:     "node-no-services",
			IP:       "192.168.1.103",
			Cluster:  Cluster{ClusterID: 7, ClusterRole: "primary"},
			Services: []Service{},
		},
	}
	sorter := &DefaultSorter{}

	nested := NewDefaultDashboardService(sorter).GetDashboard(servers)
	if len(nested) != 1 {
		t.Fatalf("GetDashboard() clusters = %d, want 1", len(nested))
	}
	var noServiceGroup *ServiceGroup
	for i := range nested[0].ServiceGroups {
		if nested[0].ServiceGroups[i].ServiceType == NoServiceType {
			noServiceGroup = &nested[0].ServiceGroups[i]
			break
		}
	}
	if noServiceGroup == nil {
		t.Fatal("GetDashboard() missing no-service group")
	}
	if len(noServiceGroup.Servers) != 2 {
		t.Fatalf("no-service group server count = %d, want 2", len(noServiceGroup.Servers))
	}

	flat := NewDefaultDashboardFlatService(sorter).GetFlatDashboard(servers)
	if len(flat) != 1 {
		t.Fatalf("GetFlatDashboard() clusters = %d, want 1", len(flat))
	}
	if len(flat[0].NoService) != 2 {
		t.Fatalf("GetFlatDashboard() no-service server count = %d, want 2", len(flat[0].NoService))
	}
}
