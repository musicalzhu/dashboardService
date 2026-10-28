package baseline

import "testing"

func TestGroupRuleString(t *testing.T) {
	tests := []struct {
		rule GroupRule
		want string
	}{
		{RuleCombination, "rule-combination"},
		{RuleServiceCount, "rule-service-count"},
		{RulePrimary, "rule-primary"},
		{RuleConvention, "rule-convention"},
		{RulePrimaryService, "rule-primary-service"},
		{GroupRule(999), "GroupRule(999)"},
	}

	for _, tt := range tests {
		if got := tt.rule.String(); got != tt.want {
			t.Errorf("GroupRule(%d).String() = %q, want %q", tt.rule, got, tt.want)
		}
	}
}

func TestParseGroupRule(t *testing.T) {
	tests := map[string]GroupRule{
		"":                     RuleCombination,
		"rule-combination":     RuleCombination,
		"rule-service-count":   RuleServiceCount,
		"rule-primary":         RulePrimary,
		"rule-convention":      RuleConvention,
		"rule-primary-service": RulePrimaryService,
	}
	for input, want := range tests {
		got, err := ParseGroupRule(input)
		if err != nil {
			t.Fatalf("ParseGroupRule(%q) error: %v", input, err)
		}
		if got != want {
			t.Errorf("ParseGroupRule(%q) = %v, want %v", input, got, want)
		}
	}
	if _, err := ParseGroupRule("invalid"); err == nil {
		t.Error("未知规则应返回错误")
	}
}
