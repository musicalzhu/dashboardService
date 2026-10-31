package builder_test

import (
	"slices"
	"testing"

	"dashboardService/baseline"
	"dashboardService/builder"
)

// TestPresetConstructorsMatchBaseline 验证五个预定义构造器与 baseline 一致。
func TestPresetConstructorsMatchBaseline(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	cases := map[builder.GroupRule]func() *builder.Builder{
		builder.RuleExactCombination:  builder.NewExactCombination,
		builder.RuleExactServiceCount: builder.NewExactServiceCount,
		builder.RuleExactPrimary:      builder.NewExactPrimary,
		builder.RuleExactConvention:   builder.NewExactConvention,
		builder.RulePrimaryService:    builder.NewPrimaryService,
	}
	for rule, ctor := range cases {
		t.Run(rule.String(), func(t *testing.T) {
			got := names(ctor().Build(servers))
			want := names(baseline.RenderDashboard(servers, rule))
			if !slices.Equal(got, want) {
				t.Errorf("不一致\n got:  %v\n want:  %v", got, want)
			}
		})
	}
}

// TestNewFromRule 验证按 GroupRule 构造与预定义构造器等价。
func TestNewFromRule(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	for _, r := range []builder.GroupRule{
		builder.RuleExactCombination,
		builder.RulePrimaryService,
	} {
		got := names(builder.NewFromRule(r).Build(servers))
		want := names(baseline.RenderDashboard(servers, r))
		if !slices.Equal(got, want) {
			t.Errorf("NewFromRule(%v) 不一致", r)
		}
	}
}

// TestDefaultBuilder 验证无配置时默认等价于 RuleExactCombination。
func TestDefaultBuilder(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	got := names(builder.New().Build(servers))
	want := names(baseline.RenderDashboard(servers, baseline.RuleCombination))
	if !slices.Equal(got, want) {
		t.Error("默认配置不一致")
	}
}

// TestOverride 验证后设置覆盖先设置。
func TestOverride(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	got := names(builder.New().
		WithKeyOf(baseline.PrimaryServiceMask).
		WithKeyOf(baseline.MaskOf). // 覆盖
		WithCmp(baseline.CmpByMask).
		Build(servers))
	want := names(baseline.RenderDashboard(servers, baseline.RuleCombination))
	if !slices.Equal(got, want) {
		t.Error("后设置应覆盖")
	}
}

// TestConfigReuse 验证 Builder 的独特能力：一次配置，多次 Build。
func TestConfigReuse(t *testing.T) {
	all := baseline.AllServiceCombinations()
	b1 := all[:5]
	b2 := all[5:]
	b := builder.NewPrimaryService()
	r1 := b.Build(b1)
	r2 := b.Build(b2)
	if countServers(r1)+countServers(r2) != len(all) {
		t.Errorf("复用结果不完整")
	}
}

// TestConfigImmutability 验证 Build 不修改 Builder。
func TestConfigImmutability(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	b := builder.NewPrimaryService()
	first := names(b.Build(servers))
	second := names(b.Build(servers))
	if !slices.Equal(first, second) {
		t.Error("Build 修改了 Builder")
	}
}

// TestNilPanics 验证 nil 配置立即 panic。
func TestNilPanics(t *testing.T) {
	t.Run("WithKeyOf", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("WithKeyOf(nil) 应当 panic")
			}
		}()
		builder.New().WithKeyOf(nil)
	})
	t.Run("WithCmp", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("WithCmp(nil) 应当 panic")
			}
		}()
		builder.New().WithCmp(nil)
	})
}

// TestInvalidRulePanics 验证无效规则 panic。
func TestInvalidRulePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("无效 GroupRule 应当 panic")
		}
	}()
	builder.NewFromRule(builder.GroupRule(999))
}

// TestParseGroupRule 验证解析委托。
func TestParseGroupRule(t *testing.T) {
	got, err := builder.ParseGroupRule("rule-primary-service")
	if err != nil {
		t.Fatal(err)
	}
	if got != builder.RulePrimaryService {
		t.Errorf("ParseGroupRule = %v", got)
	}
	if _, err := builder.ParseGroupRule("invalid"); err == nil {
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

func countServers(clusters []baseline.ClusterGroup) int {
	total := 0
	for _, c := range clusters {
		for _, g := range c.Groups {
			total += len(g.Servers)
		}
	}
	return total
}
