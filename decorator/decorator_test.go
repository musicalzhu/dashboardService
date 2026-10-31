package decorator_test

import (
	"slices"
	"testing"
	"time"

	"dashboardService/baseline"
	"dashboardService/decorator"
)

// TestPresetGroupersMatchBaseline 验证五个预定义 Grouper 与 baseline 一致。
func TestPresetGroupersMatchBaseline(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	cases := map[decorator.GroupRule]decorator.Grouper{
		decorator.RuleExactCombination:  decorator.ExactCombination,
		decorator.RuleExactServiceCount: decorator.ExactServiceCount,
		decorator.RuleExactPrimary:      decorator.ExactPrimary,
		decorator.RuleExactConvention:   decorator.ExactConvention,
		decorator.RulePrimaryService:    decorator.PrimaryService,
	}
	for rule, g := range cases {
		t.Run(rule.String(), func(t *testing.T) {
			got := names(g(servers))
			want := names(baseline.RenderDashboard(servers, rule))
			if !slices.Equal(got, want) {
				t.Errorf("不一致\n got:  %v\n want:  %v", got, want)
			}
		})
	}
}

// TestDirectInvocation 验证"规则即函数"：直接调用 Grouper。
func TestDirectInvocation(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	got := names(decorator.PrimaryService(servers))
	want := names(baseline.RenderDashboard(servers, baseline.RulePrimaryService))
	if !slices.Equal(got, want) {
		t.Error("直接调用 Grouper 结果不一致")
	}
}

// TestParse 验证字符串解析。
func TestParse(t *testing.T) {
	valid := []string{
		"", "rule-combination", "rule-service-count",
		"rule-primary", "rule-convention", "rule-primary-service",
	}
	for _, in := range valid {
		if _, err := decorator.Parse(in); err != nil {
			t.Errorf("Parse(%q) error: %v", in, err)
		}
	}
	if _, err := decorator.Parse("invalid"); err == nil {
		t.Error("无效输入应返回错误")
	}
}

// TestNames 验证规则名称枚举。
func TestNames(t *testing.T) {
	got := decorator.Names()
	want := []string{
		"rule-combination", "rule-convention", "rule-primary",
		"rule-primary-service", "rule-service-count",
	}
	if !slices.Equal(got, want) {
		t.Errorf("Names = %v, want %v", got, want)
	}
}

// TestWithLogging 验证日志装饰器只影响被装饰的规则。
func TestWithLogging(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	var logCount int
	logger := func(_ string, _ ...any) { logCount++ }

	// 只装饰 PrimaryService
	decorated := decorator.WithLogging(logger)(decorator.PrimaryService)
	decorated(servers)
	if logCount != 2 { // start + done
		t.Errorf("装饰后应有 2 条日志，实际 %d", logCount)
	}

	// 未装饰的 ExactCombination 不产生日志
	before := logCount
	decorator.ExactCombination(servers)
	if logCount != before {
		t.Error("未装饰的规则不应产生日志")
	}
}

// TestWithMetrics 验证指标装饰器记录调用耗时。
func TestWithMetrics(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	var count int
	var lastDur time.Duration
	decorated := decorator.WithMetrics(func(d time.Duration) {
		count++
		lastDur = d
	})(decorator.ExactCombination)
	decorated(servers)
	if count != 1 {
		t.Errorf("指标次数 = %d, want 1", count)
	}
	if lastDur < 0 {
		t.Errorf("耗时为负: %v", lastDur)
	}
}

// TestWithRecover 验证容错装饰器：panic 时返回非 nil 空切片。
func TestWithRecover(t *testing.T) {
	panicky := func(_ []decorator.Server) []decorator.ClusterGroup {
		panic("boom")
	}
	var captured any
	g := decorator.WithRecover(func(v any) { captured = v })(panicky)
	result := g(nil)
	if result == nil {
		t.Error("容错后应返回非 nil 空切片")
	}
	if len(result) != 0 {
		t.Errorf("容错后应返回空切片，实际 %d", len(result))
	}
	if captured != "boom" {
		t.Errorf("panic 值 = %v, want boom", captured)
	}
}

// TestApplyDecorators 验证装饰器链按顺序应用。
func TestApplyDecorators(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	var logCount, metricCount int
	composed := decorator.ApplyDecorators(
		decorator.PrimaryService,
		decorator.WithMetrics(func(_ time.Duration) { metricCount++ }),
		decorator.WithLogging(func(_ string, _ ...any) { logCount++ }),
	)
	composed(servers)
	if logCount != 2 {
		t.Errorf("日志次数 = %d, want 2", logCount)
	}
	if metricCount != 1 {
		t.Errorf("指标次数 = %d, want 1", metricCount)
	}
}

// TestByRuleInvalidPanics 验证无效 GroupRule 会 panic。
func TestByRuleInvalidPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("无效 GroupRule 应当 panic")
		}
	}()
	decorator.ByRule(decorator.GroupRule(999))
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
