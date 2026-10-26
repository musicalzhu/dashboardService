package options_test

import (
	"fmt"
	"reflect"
	"testing"

	"dashboardService/baseline"
	"dashboardService/options"
)

// TestDefaultOptionsMatchBaseline 验证不传选项时的默认行为与 switch 基线一致。
func TestDefaultOptionsMatchBaseline(t *testing.T) {
	grouping, err := options.NewGrouping()
	if err != nil {
		t.Fatalf("NewGrouping() error: %v", err)
	}
	servers := comparisonServers()
	got := grouping.Apply(servers)
	want := baseline.RenderDashboard(servers, options.RuleExactCombination)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("default options result differs from baseline\ngot:  %#v\nwant: %#v", got, want)
	}
}

// TestOptionsSelectPrimaryServiceGrouping 验证选项可分别指定分组键和组间排序函数。
func TestOptionsSelectPrimaryServiceGrouping(t *testing.T) {
	grouping, err := options.NewGrouping(
		options.WithKeyFn(baseline.PrimaryServiceMask),
		options.WithCmpFn(baseline.CompareByMaskDesc),
	)
	if err != nil {
		t.Fatalf("NewGrouping() error: %v", err)
	}
	servers := comparisonServers()
	got := grouping.Apply(servers)
	want := baseline.RenderDashboard(servers, options.RulePrimaryService)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("functional options result differs from primary-service rule\ngot:  %#v\nwant: %#v", got, want)
	}
}

// TestNewGroupingRejectsInvalidOptions 验证无效配置会明确返回错误。
func TestNewGroupingRejectsInvalidOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []options.Option
	}{
		{name: "nil option", opts: []options.Option{nil}},
		{name: "nil key function", opts: []options.Option{options.WithKeyFn(nil)}},
		{name: "nil comparator", opts: []options.Option{options.WithCmpFn(nil)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if grouping, err := options.NewGrouping(tt.opts...); err == nil || grouping != nil {
				t.Fatalf("NewGrouping() = (%v, %v), want (nil, error)", grouping, err)
			}
		})
	}
}

// TestFunctionalOptionsAreUsableOutsidePackage 验证外部调用方能够使用公开 API。
func TestFunctionalOptionsAreUsableOutsidePackage(t *testing.T) {
	grouping, err := options.NewGrouping(
		options.WithKeyFn(baseline.PrimaryServiceMask),
		options.WithCmpFn(baseline.CompareByMaskDesc),
	)
	if err != nil {
		t.Fatalf("NewGrouping() error: %v", err)
	}
	groups := grouping.Apply(comparisonServers())
	if len(groups) != 4 {
		t.Fatalf("Apply() returned %d groups, want 4", len(groups))
	}
}

func comparisonServers() []baseline.Server {
	return []baseline.Server{
		{Name: "as", Services: []baseline.Service{{ServiceType: baseline.AS}}},
		{Name: "ts", Services: []baseline.Service{{ServiceType: baseline.TS}}},
		{Name: "ss", Services: []baseline.Service{{ServiceType: baseline.SS}}},
		{Name: "ds", Services: []baseline.Service{{ServiceType: baseline.DS}}},
	}
}

// ExampleNewGrouping 展示通过函数式选项按主服务进行分组。
func ExampleNewGrouping() {
	grouping, err := options.NewGrouping(
		options.WithKeyFn(baseline.PrimaryServiceMask),
		options.WithCmpFn(baseline.CompareByMaskDesc),
	)
	if err != nil {
		fmt.Println("配置错误:", err)
		return
	}

	servers := []options.Server{
		{Name: "node-as", Services: []options.Service{{ServiceType: options.AS, State: options.SvcNormal}}},
		{Name: "node-ts", Services: []options.Service{{ServiceType: options.TS, State: options.SvcNormal}}},
	}
	for _, group := range grouping.Apply(servers) {
		fmt.Println(group.Name)
	}

	// Output:
	// AS
	// TS
}
