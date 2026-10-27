package baseline

import (
	"reflect"
	"strings"
	"testing"
)

func allServiceCombinations() []Server {
	serviceTypes := []ServiceType{ASType, TSType, SSType, DSType}
	servers := make([]Server, 0, (1<<len(serviceTypes))-1)
	for mask := 1; mask < 1<<len(serviceTypes); mask++ {
		var services []Service
		var names []string
		for bit, serviceType := range serviceTypes {
			if mask&(1<<(len(serviceTypes)-1-bit)) == 0 {
				continue
			}
			services = append(services, Service{ServiceType: serviceType, State: StateNormal})
			names = append(names, string(serviceType))
		}
		servers = append(servers, Server{
			Name:     strings.Join(names, "+"),
			IP:       "1.1.1.1",
			State:    StateNormal,
			Services: services,
		})
	}
	return servers
}

// TestEmptyInputReturnsEmptyGroups 验证空输入返回非 nil 空切片。
func TestEmptyInputReturnsEmptyGroups(t *testing.T) {
	groups := RenderDashboard(nil, RuleExactCombination)
	if groups == nil || len(groups) != 0 {
		t.Fatalf("RenderDashboard(nil, ...) = %#v, want non-nil empty slice", groups)
	}
}

// TestGroupByRetainsZeroKeyServers 验证零是有效普通键，且不会与无服务组混合或遗漏服务器。
func TestGroupByRetainsZeroKeyServers(t *testing.T) {
	servers := []Server{
		{Name: "zero-key-a", Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
		{Name: "zero-key-b", Services: []Service{{ServiceType: TSType, State: ServiceNormal}}},
		{Name: "no-service"},
	}

	groups := GroupBy(servers, func(Server) int { return 0 }, CompareByMaskDesc)
	if len(groups) != 2 {
		t.Fatalf("group count = %d, want zero-key and no-service groups", len(groups))
	}
	if groups[0].Key != 0 || groups[0].Name != "键 0" || len(groups[0].Servers) != 2 {
		t.Errorf("zero-key group = %#v, want two keyed servers", groups[0])
	}
	if groups[1].Key != noServiceGroupKey || groups[1].Name != noServiceGroupName ||
		len(groups[1].Servers) != 1 || groups[1].Servers[0].Name != "no-service" {
		t.Errorf("no-service group = %#v", groups[1])
	}
}

// TestGroupByRejectsReservedNegativeKeys 验证自定义函数不能返回内部特殊组键。
func TestGroupByRejectsReservedNegativeKeys(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("negative key should panic")
		}
	}()
	GroupBy(
		[]Server{{Name: "node", Services: []Service{{ServiceType: ASType}}}},
		func(Server) int { return -1 },
		CompareByMaskDesc,
	)
}

// TestAllRulesPreserveTheExpectedGroupingSemantics 为所有分发方案验证共享规则行为。
func TestAllRulesPreserveTheExpectedGroupingSemantics(t *testing.T) {
	servers := allServiceCombinations()
	expectedNames := map[GroupRule][]string{
		RuleExactCombination: {
			"AS+TS+SS+DS", "AS+TS+SS", "AS+TS+DS", "AS+TS",
			"AS+SS+DS", "AS+SS", "AS+DS", "AS",
			"TS+SS+DS", "TS+SS", "TS+DS", "TS",
			"SS+DS", "SS", "DS",
		},
		RuleExactConvention: {
			"AS", "AS+SS", "AS+DS", "AS+SS+DS",
			"AS+TS", "AS+TS+SS", "AS+TS+DS", "AS+TS+SS+DS",
			"TS", "TS+SS", "TS+DS", "TS+SS+DS", "SS", "SS+DS", "DS",
		},
		RulePrimaryService: {"AS", "TS", "SS", "DS"},
	}
	for rule, want := range expectedNames {
		t.Run(rule.String(), func(t *testing.T) {
			groups := RenderDashboard(servers, rule)
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

	for _, rule := range []GroupRule{RuleExactServiceCount, RuleExactPrimary} {
		t.Run(rule.String(), func(t *testing.T) {
			groups := RenderDashboard(servers, rule)
			if len(groups) != 15 {
				t.Fatalf("group count = %d, want 15", len(groups))
			}
			seenKeys := make(map[int]bool, len(groups))
			for _, group := range groups {
				if seenKeys[group.Key] {
					t.Errorf("duplicate group key %04b", group.Key)
				}
				seenKeys[group.Key] = true
			}
			if len(seenKeys) != 15 {
				t.Errorf("unique group key count = %d, want 15", len(seenKeys))
			}
		})
	}
}

// TestUnknownServicesIncluded 验证未知类型及混合类型服务器均进入独立分组。
func TestUnknownServicesIncluded(t *testing.T) {
	servers := []Server{
		{Name: "known", Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
		{Name: "unknown", Services: []Service{{ServiceType: ServiceType("XX"), State: ServiceNormal}}},
		{
			Name: "mixed",
			Services: []Service{
				{ServiceType: ASType, State: ServiceNormal},
				{ServiceType: ServiceType("XX"), State: ServiceNormal},
			},
		},
		{Name: "empty"},
	}
	for _, rule := range []GroupRule{
		RuleExactCombination,
		RuleExactServiceCount,
		RuleExactPrimary,
		RuleExactConvention,
		RulePrimaryService,
	} {
		t.Run(rule.String(), func(t *testing.T) {
			groups := RenderDashboard(servers, rule)
			total := 0
			var unknownGroup, emptyGroup *Group
			for i := range groups {
				group := &groups[i]
				total += len(group.Servers)
				switch group.Key {
				case unknownServiceGroupKey:
					unknownGroup = group
					if group.Name != unknownServiceGroupName {
						t.Errorf("unknown group name = %q, want %q", group.Name, unknownServiceGroupName)
					}
					if i != len(groups)-2 {
						t.Errorf("unknown group position = %d, want penultimate", i)
					}
				case noServiceGroupKey:
					emptyGroup = group
					if i != len(groups)-1 {
						t.Errorf("empty group position = %d, want last", i)
					}
				}
			}
			if total != len(servers) {
				t.Fatalf("grouped server count = %d, want %d", total, len(servers))
			}
			if unknownGroup == nil || len(unknownGroup.Servers) != 2 {
				t.Fatalf("unknown group = %#v, want two servers", unknownGroup)
			}
			names := []string{unknownGroup.Servers[0].Name, unknownGroup.Servers[1].Name}
			if !reflect.DeepEqual(names, []string{"mixed", "unknown"}) {
				t.Errorf("unknown group server order = %v, want [mixed unknown]", names)
			}
			if emptyGroup == nil || len(emptyGroup.Servers) != 1 || emptyGroup.Servers[0].Name != "empty" {
				t.Errorf("empty group = %#v, want empty server", emptyGroup)
			}
		})
	}
}

// TestServersAreSortedWithinGroup 直接验证组内服务器的各级排序行为。
func TestServersAreSortedWithinGroup(t *testing.T) {
	tests := []struct {
		name    string
		servers []Server
		want    []string
	}{
		{
			name: "abnormal server state first",
			servers: []Server{
				{Name: "abnormal", State: ServerAbnormal, Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
				{Name: "normal", State: ServerNormal, Services: []Service{{ServiceType: ASType, State: ServiceAbnormal}}},
			},
			want: []string{"abnormal", "normal"},
		},
		{
			name: "unknown server states have a consistent order",
			servers: []Server{
				{Name: "state-z", State: State("maintenance-z"), Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
				{Name: "state-abnormal", State: ServerAbnormal, Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
				{Name: "state-a", State: State("maintenance-a"), Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
			},
			want: []string{"state-abnormal", "state-a", "state-z"},
		},
		{
			name: "service role",
			servers: []Server{
				{Name: "follower", IP: "10.0.0.1", State: ServerNormal, Services: []Service{{ServiceType: ASType, Role: Follower}}},
				{Name: "leader", IP: "10.0.0.2", State: ServerNormal, Services: []Service{{ServiceType: ASType, Role: Leader}}},
			},
			want: []string{"leader", "follower"},
		},
		{
			name: "service roles follow declared precedence",
			servers: []Server{
				{Name: "ds-role", IP: "10.0.0.1", State: ServerNormal, Services: []Service{{ServiceType: ASType, Role: DSRole}}},
				{Name: "ss-role", IP: "10.0.0.2", State: ServerNormal, Services: []Service{{ServiceType: ASType, Role: SSRole}}},
				{Name: "lss-role", IP: "10.0.0.3", State: ServerNormal, Services: []Service{{ServiceType: ASType, Role: LSS}}},
				{Name: "follower-role", IP: "10.0.0.4", State: ServerNormal, Services: []Service{{ServiceType: ASType, Role: Follower}}},
				{Name: "leader-role", IP: "10.0.0.5", State: ServerNormal, Services: []Service{{ServiceType: ASType, Role: Leader}}},
			},
			want: []string{"leader-role", "follower-role", "lss-role", "ss-role", "ds-role"},
		},
		{
			name: "service type",
			servers: []Server{
				{Name: "ds", IP: "10.0.0.1", State: ServerNormal, Services: []Service{{ServiceType: DSType, Role: Leader}}},
				{Name: "ts", IP: "10.0.0.2", State: ServerNormal, Services: []Service{{ServiceType: TSType, Role: Leader}}},
			},
			want: []string{"ts", "ds"},
		},
		{
			name: "service list length precedes IP after common prefix",
			servers: []Server{
				{Name: "has-ts", IP: "10.0.0.1", State: ServerNormal, Services: []Service{{ServiceType: ASType, Role: Leader}, {ServiceType: TSType, Role: Follower}}},
				{Name: "as-only", IP: "10.0.0.2", State: ServerNormal, Services: []Service{{ServiceType: ASType, Role: Leader}}},
			},
			want: []string{"as-only", "has-ts"},
		},
		{
			name: "ip",
			servers: []Server{
				{Name: "later-ip", IP: "10.0.0.2", State: ServerNormal, Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
				{Name: "earlier-ip", IP: "10.0.0.1", State: ServerNormal, Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
			},
			want: []string{"earlier-ip", "later-ip"},
		},
		{
			name: "name",
			servers: []Server{
				{Name: "zulu", IP: "10.0.0.1", State: ServerNormal, Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
				{Name: "alpha", IP: "10.0.0.1", State: ServerNormal, Services: []Service{{ServiceType: ASType, State: ServiceNormal}}},
			},
			want: []string{"alpha", "zulu"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := GroupBy(tt.servers, func(Server) int { return 1 }, func(int, int) int { return 0 })
			if len(groups) != 1 {
				t.Fatalf("groups = %d, want one shared test group", len(groups))
			}
			got := make([]string, len(groups[0].Servers))
			for i, server := range groups[0].Servers {
				got[i] = server.Name
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("server order = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMarshalGroups 验证对外公开的 JSON 集成边界。
func TestMarshalGroups(t *testing.T) {
	got, err := MarshalGroups([]Group{{Key: noServiceGroupKey, Name: noServiceGroupName, Servers: []Server{}}})
	if err != nil {
		t.Fatalf("MarshalGroups() error: %v", err)
	}
	const want = "[\n  {\n    \"name\": \"无服务\",\n    \"servers\": []\n  }\n]"
	if string(got) != want {
		t.Errorf("MarshalGroups() =\n%s\nwant:\n%s", got, want)
	}
}

// TestInvalidRulePanics 验证 switch 的 default 分支拒绝未知规则。
func TestInvalidRulePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("无效规则应触发 panic")
		}
	}()
	RenderDashboard(nil, GroupRule(999))
}

func TestInvalidRuleStringDoesNotPanic(t *testing.T) {
	if got := GroupRule(999).String(); got != "GroupRule(999)" {
		t.Fatalf("GroupRule(999).String() = %q, want %q", got, "GroupRule(999)")
	}
}

// TestParseGroupRule 验证规则名称、默认值和未知名称错误。
func TestParseGroupRule(t *testing.T) {
	tests := map[string]GroupRule{
		"":                    RuleExactCombination,
		"exact-combination":   RuleExactCombination,
		"exact-service-count": RuleExactServiceCount,
		"exact-primary":       RuleExactPrimary,
		"exact-convention":    RuleExactConvention,
		"primary-service":     RulePrimaryService,
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
