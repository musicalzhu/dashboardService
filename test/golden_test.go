// Package test 提供跨方案一致性 golden 测试。
//
// 目的：
//
//	验证 6 个 POC 方案的输出与 baseline 完全一致。
//	任何方案在重构后输出偏移，本测试立即失败。
//
// 测试数据来源：
//
//	baseline.AllServiceCombinations() —— 覆盖全部 15 种服务组合。
//
// 归一化策略：
//
//	把 []baseline.ClusterGroup 转成 []groupSnapshot（只保留有序的
//	集群 ID、分组名、服务器名列表），消除不同方案内部 Group.Key
//	等无关差异，只比对用户可见的输出。
package test

import (
	"reflect"
	"testing"

	"dashboardService/baseline"
	"dashboardService/builder"
	"dashboardService/datadriven"
	"dashboardService/decorator"
	"dashboardService/options"
	"dashboardService/strategy"
)

// groupSnapshot 是跨方案比较的统一快照结构。
//
// 为什么需要快照：
//   - 不同方案内部的 Group.Key 值可能不同（如 strategy 单独实现骨架，
//     可能产生相同的 Key 但顺序不同——虽然我们的目标是完全一致，
//     但快照能让我们精确指出“哪一层的哪个字段”不一致）。
//   - 快照只保留用户可见的输出（集群 ID、分组名、服务器名列表），
//     避免内部实现差异污染比较结果。
type groupSnapshot struct {
	ClusterID   int
	GroupName   string
	ServerNames []string
}

// snapshot 把 baseline.ClusterGroup 切片转成快照。
//
// 只保留用户可见的输出，忽略 Group.Key 等内部字段。
// 服务器名列表保持原顺序——分组内的服务器排序本身也是方案对比的一部分。
func snapshot(clusters []baseline.ClusterGroup) []groupSnapshot {
	result := make([]groupSnapshot, 0)
	for _, c := range clusters {
		for _, g := range c.Groups {
			names := make([]string, len(g.Servers))
			for i, s := range g.Servers {
				names[i] = s.Name
			}
			result = append(result, groupSnapshot{
				ClusterID:   c.ClusterID,
				GroupName:   g.Name,
				ServerNames: names,
			})
		}
	}
	return result
}

// TestGolden_AllSchemesMatchBaseline 验证 6 个方案的输出与 baseline 完全一致。
//
// 覆盖范围：
//   - 5 种规则（ExactCombination / ExactServiceCount / ExactPrimary / ExactConvention / PrimaryService）
//   - 15 种服务组合（AllServiceCombinations）
//   - 6 个方案（baseline / datadriven / options / decorator / builder / strategy）
func TestGolden_AllSchemesMatchBaseline(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	rules := []baseline.GroupRule{
		baseline.RuleCombination,
		baseline.RuleServiceCount,
		baseline.RulePrimary,
		baseline.RuleConvention,
		baseline.RulePrimaryService,
	}

	for _, rule := range rules {
		rule := rule // 避免 Go 1.21 之前的闭包捕获问题
		want := snapshot(baseline.RenderDashboard(servers, rule))

		// ---- 方案二：datadriven ----
		t.Run("datadriven/"+rule.String(), func(t *testing.T) {
			got := snapshot(datadriven.RenderDashboard(servers, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("datadriven 输出与 baseline 不一致\n got:  %v\n want: %v", got, want)
			}
		})

		// ---- 方案三：options ----
		t.Run("options/"+rule.String(), func(t *testing.T) {
			got := snapshot(options.RenderByRule(servers, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("options 输出与 baseline 不一致\n got:  %v\n want: %v", got, want)
			}
		})

		// ---- 方案四：decorator ----
		t.Run("decorator/"+rule.String(), func(t *testing.T) {
			got := snapshot(decorator.RenderDashboard(servers, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("decorator 输出与 baseline 不一致\n got:  %v\n want: %v", got, want)
			}
		})

		// ---- 方案五：builder ----
		t.Run("builder/"+rule.String(), func(t *testing.T) {
			got := snapshot(builder.RenderDashboard(servers, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("builder 输出与 baseline 不一致\n got:  %v\n want: %v", got, want)
			}
		})

		// ---- 方案六：strategy ----
		t.Run("strategy/"+rule.String(), func(t *testing.T) {
			got := snapshot(strategy.RenderDashboard(servers, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("strategy 输出与 baseline 不一致\n got:  %v\n want: %v", got, want)
			}
		})
	}
}

// TestGolden_EmptyInput 验证所有方案在空输入下的行为一致。
//
// 空输入是边界场景：
//   - 应返回非 nil 空切片（JSON 序列化为 [] 而非 null）
//   - 所有方案必须一致
func TestGolden_EmptyInput(t *testing.T) {
	rules := []baseline.GroupRule{
		baseline.RuleCombination,
		baseline.RulePrimaryService,
	}

	for _, rule := range rules {
		rule := rule
		t.Run("empty/"+rule.String(), func(t *testing.T) {
			want := snapshot(baseline.RenderDashboard(nil, rule))
			if len(want) != 0 {
				t.Fatalf("baseline 空输入应返回空结果，实际 %d 项", len(want))
			}

			got := snapshot(datadriven.RenderDashboard(nil, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("datadriven 空输入结果不一致")
			}

			got = snapshot(options.RenderByRule(nil, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("options 空输入结果不一致")
			}

			got = snapshot(decorator.RenderDashboard(nil, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("decorator 空输入结果不一致")
			}

			got = snapshot(builder.RenderDashboard(nil, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("builder 空输入结果不一致")
			}

			got = snapshot(strategy.RenderDashboard(nil, rule))
			if !reflect.DeepEqual(got, want) {
				t.Errorf("strategy 空输入结果不一致")
			}
		})
	}
}

// TestGolden_NoServiceGrouping 验证“无服务”服务器在所有方案下都归入同一个分组。
//
// 这是 C5（无服务服务器独立分组）变更的回归保护。
func TestGolden_NoServiceGrouping(t *testing.T) {
	// 构造混合数据：有服务 + 无服务 + 未知服务
	servers := []baseline.Server{
		{
			Name:     "with-as",
			Cluster:  baseline.Cluster{ClusterID: 1},
			Services: []baseline.Service{{ServiceType: baseline.ASType}},
		},
		{
			Name:     "no-service",
			Cluster:  baseline.Cluster{ClusterID: 1},
			Services: nil, // 无服务
		},
		{
			Name:     "empty-services",
			Cluster:  baseline.Cluster{ClusterID: 1},
			Services: []baseline.Service{}, // 空切片
		},
	}

	rules := []baseline.GroupRule{baseline.RuleCombination}

	for _, rule := range rules {
		rule := rule
		want := snapshot(baseline.RenderDashboard(servers, rule))

		t.Run("no-service/"+rule.String(), func(t *testing.T) {
			for name, got := range map[string][]groupSnapshot{
				"datadriven": snapshot(datadriven.RenderDashboard(servers, rule)),
				"options":    snapshot(options.RenderByRule(servers, rule)),
				"decorator":  snapshot(decorator.RenderDashboard(servers, rule)),
				"builder":    snapshot(builder.RenderDashboard(servers, rule)),
				"strategy":   snapshot(strategy.RenderDashboard(servers, rule)),
			} {
				if !reflect.DeepEqual(got, want) {
					t.Errorf("%s 输出与 baseline 不一致\n got:  %v\n want: %v", name, got, want)
				}
			}
		})
	}
}
