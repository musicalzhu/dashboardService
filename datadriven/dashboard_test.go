package datadriven

import (
	"reflect"
	"testing"

	"dashboardService/baseline"
)

// TestAllRulesHaveImpl 验证每条共享规则都恰好有一条数据驱动映射；共同行为由 baseline 测试。
func TestAllRulesHaveImpl(t *testing.T) {
	rules := []GroupRule{
		RuleExactCombination,
		RuleExactServiceCount,
		RuleExactPrimary,
		RuleExactConvention,
		RulePrimaryService,
	}
	for _, rule := range rules {
		if _, ok := ruleImpls[rule]; !ok {
			t.Errorf("rule %v has no implementation mapping", rule)
		}
	}
	if len(ruleImpls) != len(rules) {
		t.Errorf("rule implementation count = %d, want %d", len(ruleImpls), len(rules))
	}
}

// TestDataDrivenDispatchMatchesBaseline 验证规则表入口与 switch 基线委托到相同的共享实现。
func TestDataDrivenDispatchMatchesBaseline(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	servers = append(servers,
		Server{
			Name:     "unknown",
			Services: []Service{{ServiceType: ServiceType("XX"), State: SvcNormal}},
		},
		Server{Name: "empty"},
	)

	for _, rule := range []GroupRule{
		RuleExactCombination,
		RuleExactServiceCount,
		RuleExactPrimary,
		RuleExactConvention,
		RulePrimaryService,
	} {
		t.Run(rule.String(), func(t *testing.T) {
			got := RenderDashboard(servers, rule)
			want := baseline.RenderDashboard(servers, rule)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("data-driven result differs from shared baseline\ngot:  %#v\nwant: %#v", got, want)
			}
		})
	}
}
