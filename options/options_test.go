package options_test

import (
	"slices"
	"testing"

	"dashboardService/baseline"
	"dashboardService/options"
)

// TestPresetMatchBaseline 验证五个预定义选项组与 baseline 一致。
func TestPresetMatchBaseline(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	cases := map[options.GroupRule][]options.Option{
		options.RuleExactCombination:  options.ExactCombination,
		options.RuleExactServiceCount: options.ExactServiceCount,
		options.RuleExactPrimary:      options.ExactPrimary,
		options.RuleExactConvention:   options.ExactConvention,
		options.RulePrimaryService:    options.PrimaryService,
	}
	for rule, opts := range cases {
		t.Run(rule.String(), func(t *testing.T) {
			got := names(options.RenderDashboard(servers, opts...))
			want := names(baseline.RenderDashboard(servers, rule))
			if !slices.Equal(got, want) {
				t.Errorf("不一致\n got:  %v\n want:  %v", got, want)
			}
		})
	}
}

// TestFreeComposition 验证自由组合能力：主服务归组 + 数量升序。
//
// 这个组合在 datadriven 的 ruleImpls 表中不存在，但 options 可以构造。
// 这是 options 方案相对 datadriven 的独特价值。
func TestFreeComposition(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	clusters := options.RenderDashboard(servers,
		options.WithKeyOf(baseline.PrimaryServiceMask),
		options.WithCmp(baseline.CmpByServiceCount),
	)
	totalGroups := 0
	for _, c := range clusters {
		totalGroups += len(c.Groups)
	}
	if totalGroups != 4 {
		t.Errorf("主服务归组后分组数 = %d, want 4", totalGroups)
	}
}

// TestNilOptionPanics 验证 nil 立即 panic（fail-fast）。
func TestNilOptionPanics(t *testing.T) {
	t.Run("WithKeyOf", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("WithKeyOf(nil) 应当 panic")
			}
		}()
		options.WithKeyOf(nil)
	})
	t.Run("WithCmp", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("WithCmp(nil) 应当 panic")
			}
		}()
		options.WithCmp(nil)
	})
}

// TestDefaultNoOptions 验证无选项时默认等价于 RuleExactCombination。
func TestDefaultNoOptions(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	got := names(options.RenderDashboard(servers))
	want := names(baseline.RenderDashboard(servers, baseline.RuleCombination))
	if !slices.Equal(got, want) {
		t.Error("无选项默认行为不一致")
	}
}

// TestOptionOverride 验证后传入的 Option 覆盖先传入的相同字段。
func TestOptionOverride(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	got := names(options.RenderDashboard(servers,
		options.WithKeyOf(baseline.PrimaryServiceMask),
		options.WithKeyOf(baseline.MaskOf), // 覆盖前一个
		options.WithCmp(baseline.CmpByMask),
	))
	want := names(baseline.RenderDashboard(servers, baseline.RuleCombination))
	if !slices.Equal(got, want) {
		t.Error("后传入的 Option 应覆盖先传入的相同字段")
	}
}

// TestRenderByRule 验证 RenderByRule 与 RenderDashboard + OptionsOfRule 等价。
func TestRenderByRule(t *testing.T) {
	servers := baseline.AllServiceCombinations()
	for _, r := range []options.GroupRule{
		options.RuleExactCombination,
		options.RulePrimaryService,
	} {
		got1 := names(options.RenderByRule(servers, r))
		opts, err := options.OptionsOfRule(r)
		if err != nil {
			t.Fatal(err)
		}
		got2 := names(options.RenderDashboard(servers, opts...))
		if !slices.Equal(got1, got2) {
			t.Errorf("RenderByRule 与 RenderDashboard 不一致")
		}
	}
}

// TestInvalidRulePanics 验证无效 GroupRule 会 panic。
func TestInvalidRulePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("无效 GroupRule 应当 panic")
		}
	}()
	options.RenderByRule(nil, options.GroupRule(999))
}

// TestOptionsOfRuleInvalid 验证无效规则的错误返回。
func TestOptionsOfRuleInvalid(t *testing.T) {
	if _, err := options.OptionsOfRule(options.GroupRule(999)); err == nil {
		t.Error("无效规则应返回错误")
	}
}

// TestParseGroupRule 验证解析委托。
func TestParseGroupRule(t *testing.T) {
	got, err := options.ParseGroupRule("rule-primary-service")
	if err != nil {
		t.Fatal(err)
	}
	if got != options.RulePrimaryService {
		t.Errorf("ParseGroupRule = %v, want %v", got, options.RulePrimaryService)
	}
	if _, err := options.ParseGroupRule("invalid"); err == nil {
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
