package baseline

import (
	"reflect"
	"testing"
)

// TestEmptyInputReturnsEmptyGroups 验证空输入返回非 nil 空切片。
func TestEmptyInputReturnsEmptyClusterGroups(t *testing.T) {
	clustGroups := RenderDashboard(nil, RuleCombination)
	if clustGroups == nil || len(clustGroups) != 0 {
		t.Fatalf("RenderDashboard(nil, ...) = %#v, want non-nil empty slice", clustGroups)
	}
}

// TestAllRulesPreserveTheExpectedGroupingSemantics 为所有分发方案验证共享规则行为。
func TestAllRulesPreserveTheExpectedGroupingSemantics(t *testing.T) {
	servers := AllServiceCombinations()
	expectedNames := map[GroupRule][]string{
		RuleCombination: {
			"AS+TS+SS+DS", "AS+TS+SS", "AS+TS+DS", "AS+TS",
			"AS+SS+DS", "AS+SS", "AS+DS", "AS",
			"TS+SS+DS", "TS+SS", "TS+DS", "TS",
			"SS+DS", "SS", "DS",
		},
		RuleConvention: {
			"AS", "AS+SS", "AS+DS", "AS+SS+DS",
			"AS+TS", "AS+TS+SS", "AS+TS+DS", "AS+TS+SS+DS",
			"TS", "TS+SS", "TS+DS", "TS+SS+DS", "SS", "SS+DS", "DS",
		},
		RulePrimaryService: {"AS", "TS", "SS", "DS"},
	}
	for rule, want := range expectedNames {
		t.Run(rule.String(), func(t *testing.T) {
			groups := RenderDashboard(servers, rule)[0].Groups
			if rule == RulePrimaryService {
				counts := []int{8, 4, 2, 1}
				if len(groups) != len(counts) {
					t.Fatalf("group count = %d, want %d", len(groups), len(counts))
				}
				for i, group := range groups {
					if len(group.Servers) != counts[i] {
						t.Errorf("%s group size = %d, want %d", group.Name, len(group.Servers), counts[i])
					}
				}
			}
			got := make([]string, len(groups))
			for i, group := range groups {
				got[i] = group.Name
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("group order = %v, want %v", got, want)
			}
		})
	}

	expectedOrder := map[GroupRule][]string{
		RuleServiceCount: {
			"AS", "TS", "SS", "DS",
			"AS+TS", "AS+SS", "AS+DS", "TS+SS",
			"TS+DS", "SS+DS", "AS+TS+SS", "AS+TS+DS",
			"AS+SS+DS", "TS+SS+DS", "AS+TS+SS+DS",
		},
		RulePrimary: {
			"AS", "AS+TS", "AS+SS", "AS+DS",
			"AS+TS+SS", "AS+TS+DS", "AS+SS+DS", "AS+TS+SS+DS",
			"TS", "TS+SS", "TS+DS", "TS+SS+DS",
			"SS", "SS+DS", "DS",
		},
	}
	for _, rule := range []GroupRule{RuleServiceCount, RulePrimary} {
		t.Run(rule.String(), func(t *testing.T) {
			groups := RenderDashboard(servers, rule)[0].Groups
			got := make([]string, len(groups))
			for i, group := range groups {
				got[i] = group.Name
			}
			if !reflect.DeepEqual(got, expectedOrder[rule]) {
				t.Errorf("group order = %v, want %v", got, expectedOrder[rule])
			}
		})
	}
}
