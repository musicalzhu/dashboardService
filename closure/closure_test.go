package closure_test

import (
	"fmt"
	"reflect"
	"testing"

	"dashboardService/baseline"
	"dashboardService/closure"
)

// TestRulesMatchBaseline 验证五个可调用闭包分别对应共享基线中的预定义规则。
func TestRulesMatchBaseline(t *testing.T) {
	servers := []closure.Server{
		{Name: "as", Services: []closure.Service{{ServiceType: closure.AS, State: closure.SvcNormal}}},
		{Name: "ts-ss", Services: []closure.Service{
			{ServiceType: closure.TS, State: closure.SvcNormal},
			{ServiceType: closure.SS, State: closure.SvcNormal},
		}},
		{Name: "ds", Services: []closure.Service{{ServiceType: closure.DS, State: closure.SvcNormal}}},
		{Name: "empty"},
	}
	rules := []struct {
		name    string
		grouper closure.Grouper
		rule    baseline.GroupRule
	}{
		{"exact combination", closure.RuleExactCombination, baseline.RuleExactCombination},
		{"exact service count", closure.RuleExactServiceCount, baseline.RuleExactServiceCount},
		{"exact primary", closure.RuleExactPrimary, baseline.RuleExactPrimary},
		{"exact convention", closure.RuleExactConvention, baseline.RuleExactConvention},
		{"primary service", closure.RulePrimaryService, baseline.RulePrimaryService},
	}

	for _, test := range rules {
		t.Run(test.name, func(t *testing.T) {
			got := test.grouper(servers)
			want := baseline.RenderDashboard(servers, test.rule)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("closure result differs from baseline\ngot:  %#v\nwant: %#v", got, want)
			}
		})
	}
}

// ExampleRulePrimaryService 展示规则本身就是函数，可直接传入服务器列表调用。
func ExampleRulePrimaryService() {
	servers := []closure.Server{
		{Name: "node-as", Services: []closure.Service{{ServiceType: closure.AS, State: closure.SvcNormal}}},
		{Name: "node-ts", Services: []closure.Service{{ServiceType: closure.TS, State: closure.SvcNormal}}},
	}

	for _, group := range closure.RulePrimaryService(servers) {
		fmt.Println(group.Name)
	}

	// Output:
	// AS
	// TS
}
