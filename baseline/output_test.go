// baseline/output_test.go

package baseline

import (
	"fmt"
	"reflect"
	"testing"
)

// TestFlattenOrderMatchesGroups 验证 order 与 Groups 顺序完全一致。
func TestFlattenOrderMatchesGroups(t *testing.T) {
	servers := AllServiceCombinations()

	for _, rule := range []GroupRule{
		RuleCombination,
		RuleServiceCount,
		RulePrimary,
		RuleConvention,
		RulePrimaryService,
	} {
		t.Run(rule.String(), func(t *testing.T) {
			clusters := RenderDashboard(servers, rule)
			flat := Flatten(clusters)

			for i, c := range clusters {
				wantOrder := make([]string, len(c.Groups))
				for j, g := range c.Groups {
					wantOrder[j] = g.Name
				}
				gotOrder, ok := flat[i]["order"].([]string)
				if !ok {
					t.Fatalf("cluster %d 缺少 order 字段或类型错误", i)
				}
				if !reflect.DeepEqual(gotOrder, wantOrder) {
					t.Errorf("cluster %d order 顺序错误\n got:  %v\n want: %v",
						i, gotOrder, wantOrder)
				}

				// 验证每个分组名都能在 flat 里取到
				for _, name := range wantOrder {
					if _, ok := flat[i][name]; !ok {
						t.Errorf("cluster %d 缺少分组字段 %q", i, name)
					}
				}
			}
		})
	}
}

// TestFlattenPreservesAllGroups 验证空分组也被保留。
func TestFlattenPreservesAllGroups(t *testing.T) {
	servers := AllServiceCombinations()
	clusters := RenderDashboard(servers, RuleCombination)
	flat := Flatten(clusters)

	for i, c := range clusters {
		for _, g := range c.Groups {
			if _, ok := flat[i][g.Name]; !ok {
				t.Errorf("cluster %d 缺少分组 %q", i, g.Name)
			}
		}
	}
}

// TestFlattenClusterMetadata 验证集群元信息完整保留。
func TestFlattenClusterMetadata(t *testing.T) {
	servers := AllServiceCombinations()
	clusters := RenderDashboard(servers, RuleCombination)
	flat := Flatten(clusters)

	for i, c := range clusters {
		if flat[i]["cluster_id"] != c.ClusterID {
			t.Errorf("cluster %d: cluster_id = %v, want %v",
				i, flat[i]["cluster_id"], c.ClusterID)
		}
		if flat[i]["cluster_role"] != c.ClusterRole {
			t.Errorf("cluster %d: cluster_role = %v, want %v",
				i, flat[i]["cluster_role"], c.ClusterRole)
		}
		if flat[i]["cluster_name"] != c.ClusterName {
			t.Errorf("cluster %d: cluster_name = %v, want %v",
				i, flat[i]["cluster_name"], c.ClusterName)
		}
		if flat[i]["state"] != c.State {
			t.Errorf("cluster %d: state = %v, want %v",
				i, flat[i]["state"], c.State)
		}
	}
}

// TestFlattenWithPreview_TruncatesAndKeepsRest 验证分组超过 previewSize 时，
// 前 N 台放入分组字段，剩余放入 rest 字段。
func TestFlattenWithPreview_TruncatesAndKeepsRest(t *testing.T) {
	// 构造 5 台 AS 服务器（RulePrimaryService 下全部归入 AS 组）
	servers := make([]Server, 5)
	for i := 0; i < 5; i++ {
		servers[i] = Server{
			Name:     fmt.Sprintf("as-%d", i),
			Cluster:  Cluster{ClusterID: 1, ClusterName: "c1"},
			IP:       fmt.Sprintf("1.1.1.%d", i),
			State:    StateNormal,
			Services: []Service{{ServiceType: ASType, State: StateNormal}},
		}
	}

	clusters := RenderDashboard(servers, RulePrimaryService)
	flat := FlattenWithPreview(clusters, 3)

	// AS 组应有 3 台
	asServers, ok := flat[0]["AS"].([]Server)
	if !ok {
		t.Fatal("缺少 AS 字段或类型错误")
	}
	if len(asServers) != 3 {
		t.Errorf("AS 字段 = %d 台, want 3", len(asServers))
	}

	// restAS 应有 2 台
	restAS, ok := flat[0]["restAS"].([]Server)
	if !ok {
		t.Fatal("缺少 restAS 字段或类型错误")
	}
	if len(restAS) != 2 {
		t.Errorf("restAS 字段 = %d 台, want 2", len(restAS))
	}

	// 验证前后顺序：AS 是前 3 台，restAS 是后 2 台（顺序不重叠）
	for i := 0; i < 3; i++ {
		if asServers[i].Name != servers[i].Name {
			t.Errorf("AS[%d] = %s, want %s", i, asServers[i].Name, servers[i].Name)
		}
	}
	for i := 0; i < 2; i++ {
		if restAS[i].Name != servers[3+i].Name {
			t.Errorf("restAS[%d] = %s, want %s", i, restAS[i].Name, servers[3+i].Name)
		}
	}
}

// TestFlattenWithPreview_NoRestWhenUnderPreviewSize 验证分组不足 previewSize 时
// 不输出 rest 字段。
func TestFlattenWithPreview_NoRestWhenUnderPreviewSize(t *testing.T) {
	servers := []Server{
		{Name: "as-1", Cluster: Cluster{ClusterID: 1}, Services: []Service{{ServiceType: ASType}}},
		{Name: "as-2", Cluster: Cluster{ClusterID: 1}, Services: []Service{{ServiceType: ASType}}},
	}
	clusters := RenderDashboard(servers, RulePrimaryService)
	flat := FlattenWithPreview(clusters, 3)

	asServers, ok := flat[0]["AS"].([]Server)
	if !ok {
		t.Fatal("缺少 AS 字段")
	}
	if len(asServers) != 2 {
		t.Errorf("AS 字段 = %d 台, want 2", len(asServers))
	}

	if _, ok := flat[0]["restAS"]; ok {
		t.Error("分组不足 previewSize 时不应输出 restAS 字段")
	}
}

// TestFlattenWithPreview_OrderPreserved 验证 order 字段与 Groups 顺序一致。
func TestFlattenWithPreview_OrderPreserved(t *testing.T) {
	servers := AllServiceCombinations()
	for _, rule := range []GroupRule{
		RuleCombination,
		RuleServiceCount,
		RulePrimary,
		RuleConvention,
		RulePrimaryService,
	} {
		t.Run(rule.String(), func(t *testing.T) {
			clusters := RenderDashboard(servers, rule)
			flat := FlattenWithPreview(clusters, 3)

			for i, c := range clusters {
				wantOrder := make([]string, len(c.Groups))
				for j, g := range c.Groups {
					wantOrder[j] = g.Name
				}
				gotOrder, ok := flat[i]["order"].([]string)
				if !ok {
					t.Fatalf("cluster %d 缺少 order 字段", i)
				}
				if !reflect.DeepEqual(gotOrder, wantOrder) {
					t.Errorf("cluster %d order 错误\n got:  %v\n want: %v",
						i, gotOrder, wantOrder)
				}
			}
		})
	}
}

// TestFlattenWithPreview_InvalidPreviewSizePanics 验证 previewSize <= 0 时 panic。
func TestFlattenWithPreview_InvalidPreviewSizePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("previewSize <= 0 应当 panic")
		}
	}()
	FlattenWithPreview(nil, 0)
}
