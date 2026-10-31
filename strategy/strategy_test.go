package strategy

import (
	"slices"
	"testing"

	"dashboardService/baseline"
)

// TestStrategiesTableCompleteness 验证 strategies 表覆盖所有 GroupRule。
//
// 新增 GroupRule 常量时如果忘记在 strategies 中加 entry，只有本测试能捕获。
func TestStrategiesTableCompleteness(t *testing.T) {
	allRules := []GroupRule{
		RuleExactCombination,
		RuleExactServiceCount,
		RuleExactPrimary,
		RuleExactConvention,
		RulePrimaryService,
	}
	for _, r := range allRules {
		if _, ok := strategies[r]; !ok {
			t.Errorf("规则 %v 缺少 strategies entry", r)
		}
	}
	if len(strategies) != len(allRules) {
		t.Errorf("strategies 有 %d 项, GroupRule 有 %d 种", len(strategies), len(allRules))
	}
}

// TestStrategiesMatchBaseline 验证每个策略与 baseline 一致。
func TestStrategiesMatchBaseline(t *testing.T) {
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
				t.Errorf("不一致\n got:  %v\n want:  %v", got, want)
			}
		})
	}
}

// TestStrategySelfContained 验证每个策略的 Group 方法可独立调用。
//
// 这是纯策略模式的核心特征：策略之间不共享 Group 实现，
// 每个策略可以单独使用，不需要其它策略存在。
func TestStrategySelfContained(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	for _, r := range []GroupRule{
		RuleExactCombination,
		RulePrimaryService,
	} {
		s := ByRule(r)
		if s.Name() == "" {
			t.Errorf("策略 %v Name() 为空", r)
		}
		result := s.Group(servers)
		if len(result) == 0 {
			t.Errorf("策略 %v Group() 返回空", r)
		}
	}
}

// TestPrimaryServiceGrouping 验证主服务归组的分组数。
func TestPrimaryServiceGrouping(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	clusters := RenderDashboard(servers, RulePrimaryService)
	total := 0
	for _, c := range clusters {
		total += len(c.Groups)
	}
	if total != 4 {
		t.Errorf("分组数 = %d, want 4", total)
	}
}

// TestExactCombinationGrouping 验证精确组合的分组数。
func TestExactCombinationGrouping(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	clusters := RenderDashboard(servers, RuleExactCombination)
	total := 0
	for _, c := range clusters {
		total += len(c.Groups)
	}
	if total != 15 {
		t.Errorf("分组数 = %d, want 15", total)
	}
}

// TestByRuleInvalidPanics 验证无效规则 panic。
func TestByRuleInvalidPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("无效 GroupRule 应当 panic")
		}
	}()
	ByRule(GroupRule(999))
}

// TestParse 验证解析委托。
func TestParse(t *testing.T) {
	valid := []string{
		"", "rule-combination", "rule-service-count",
		"rule-primary", "rule-convention", "rule-primary-service",
	}
	for _, in := range valid {
		if _, err := Parse(in); err != nil {
			t.Errorf("Parse(%q) error: %v", in, err)
		}
	}
	if _, err := Parse("invalid"); err == nil {
		t.Error("无效输入应返回错误")
	}
}

func names(clusters []baseline.ClusterGroup) []string {
	var result []string
	for _, c := range clusters {
		for _, g := range c.Groups {
			result = append(result, g.Name)
		}
	}
	return result
}
