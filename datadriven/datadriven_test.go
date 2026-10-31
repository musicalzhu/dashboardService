package datadriven

import (
	"slices"
	"testing"

	"dashboardService/baseline"
)

// TestRuleTableCompleteness 验证 ruleImpls 覆盖所有 GroupRule。
//
// 这是 datadriven 方案的关键测试：新增 GroupRule 常量时如果忘记
// 在 ruleImpls 中加 entry，只有本测试能捕获。
func TestRuleTableCompleteness(t *testing.T) {
	allRules := []GroupRule{
		RuleExactCombination,
		RuleExactServiceCount,
		RuleExactPrimary,
		RuleExactConvention,
		RulePrimaryService,
	}
	for _, r := range allRules {
		if _, ok := ruleImpls[r]; !ok {
			t.Errorf("规则 %v 缺少 ruleImpls entry", r)
		}
	}
	if len(ruleImpls) != len(allRules) {
		t.Errorf("ruleImpls 有 %d 项, GroupRule 有 %d 种", len(ruleImpls), len(allRules))
	}
}

// TestRenderDashboardMatchesBaseline 验证每种规则与 baseline 一致。
func TestRenderDashboardMatchesBaseline(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	for _, r := range []GroupRule{
		RuleExactCombination,
		RuleExactServiceCount,
		RuleExactPrimary,
		RuleExactConvention,
		RulePrimaryService,
	} {
		t.Run(r.String(), func(t *testing.T) {
			got := names(RenderDashboard(servers, r))
			want := names(baseline.RenderDashboard(servers, r))
			if !slices.Equal(got, want) {
				t.Errorf("不一致\n got:  %v\n want: %v", got, want)
			}
		})
	}
}

// TestInvalidRulePanics 验证无效规则会 panic（fail-fast）。
func TestInvalidRulePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("无效 GroupRule 应当 panic")
		}
	}()
	RenderDashboard(nil, GroupRule(999))
}

// TestParseGroupRule 验证 ParseGroupRule 委托给 baseline。
func TestParseGroupRule(t *testing.T) {
	got, err := ParseGroupRule("rule-primary-service")
	if err != nil {
		t.Fatal(err)
	}
	if got != RulePrimaryService {
		t.Errorf("ParseGroupRule = %v, want %v", got, RulePrimaryService)
	}
	if _, err := ParseGroupRule("invalid"); err == nil {
		t.Error("无效输入应返回错误")
	}
}

func names(clusters []ClusterGroup) []string {
	var result []string
	for _, c := range clusters {
		for _, g := range c.Groups {
			result = append(result, g.Name)
		}
	}
	return result
}
