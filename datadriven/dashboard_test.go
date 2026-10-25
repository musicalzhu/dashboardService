package datadriven

import (
	"encoding/json"
	"math/bits"
	"slices"
	"testing"
)

// ============================================================================
// 不变量测试：bit 定义与优先级必须一致
// ============================================================================

// TestBitDefinitions 验证每个 bit 的数值正确。
//
// 硬编码断言，防止未来重构改错（如误写 asBit = 0b0001）。
func TestBitDefinitions(t *testing.T) {
	cases := map[string]struct{ got, want int }{
		"asBit": {asBit, 0b1000},
		"tsBit": {tsBit, 0b0100},
		"ssBit": {ssBit, 0b0010},
		"dsBit": {dsBit, 0b0001},
	}
	for name, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %04b, want %04b", name, c.got, c.want)
		}
	}
}

// TestBitOrderMatchesPriority 验证 bit 数值顺序与 allServiceTypes 优先级一致。
//
// 这是 groupBy(primaryServiceMask, cmpByMaskDesc) 产生 AS>TS>SS>DS 的前提。
// 若此不变量被破坏，RulePrimaryService 的顺序会静默出错。
func TestBitOrderMatchesPriority(t *testing.T) {
	for i := 1; i < len(allServiceTypes); i++ {
		hi := serviceBit[allServiceTypes[i-1]]
		lo := serviceBit[allServiceTypes[i]]
		if hi <= lo {
			t.Fatalf("bit 定义与优先级不一致：%s(%04b) 应 > %s(%04b)",
				allServiceTypes[i-1], hi, allServiceTypes[i], lo)
		}
	}
}

// ============================================================================
// 规则表完整性测试（map 版本必加）
// ============================================================================

// TestAllRulesHaveImpl 验证所有 GroupRule 常量都在 ruleImpls 表中有实现。
//
// 这是 map 版本的核心保障：新增 GroupRule 常量却忘记加 map entry 时，
// 只有本测试能捕获（编译器不会检查）。
func TestAllRulesHaveImpl(t *testing.T) {
	allRules := []GroupRule{
		RuleExactCombination,
		RuleExactServiceCount,
		RuleExactPrimary,
		RuleExactConvention,
		RulePrimaryService,
	}
	for _, r := range allRules {
		if _, ok := ruleImpls[r]; !ok {
			t.Errorf("规则 %v (%d) 缺少 ruleImpls entry", r, int(r))
		}
	}
	if len(ruleImpls) != len(allRules) {
		t.Errorf("ruleImpls 有 %d 项，GroupRule 有 %d 种（有冗余或遗漏）",
			len(ruleImpls), len(allRules))
	}
}

// ============================================================================
// 组合命名测试
// ============================================================================

// TestComboName 验证 mask 到名称的转换。
func TestComboName(t *testing.T) {
	cases := map[int]string{
		0b1000: "AS",
		0b1100: "AS+TS",
		0b1010: "AS+SS",
		0b1001: "AS+DS",
		0b1111: "AS+TS+SS+DS",
		0b0100: "TS",
		0b0001: "DS",
		0:      "",
	}
	for mask, want := range cases {
		if got := comboName(mask); got != want {
			t.Errorf("comboName(%04b) = %q, want %q", mask, got, want)
		}
	}
}

// ============================================================================
// 五种规则的组数与顺序测试
// ============================================================================

// TestGroupCounts 验证各规则的分组数量。
func TestGroupCounts(t *testing.T) {
	servers := buildAllCombinations()
	cases := map[GroupRule]int{
		RuleExactCombination:  15,
		RuleExactServiceCount: 15,
		RuleExactPrimary:      15,
		RuleExactConvention:   15,
		RulePrimaryService:    4, // 合并为 AS/TS/SS/DS 四组
	}
	for rule, want := range cases {
		t.Run(rule.String(), func(t *testing.T) {
			got := RenderDashboard(servers, rule)
			if len(got) != want {
				t.Errorf("组数 = %d, want %d", len(got), want)
			}
		})
	}
}

// TestRulePrimaryServiceOrderAndContent 验证主服务分组的顺序和内容。
//
// 期望：
//   - 顺序：AS → TS → SS → DS
//   - AS 组 8 台（所有含 AS 的组合）
//   - TS 组 4 台（不含 AS 但含 TS 的组合）
//   - SS 组 2 台（不含 AS/TS 但含 SS 的组合）
//   - DS 组 1 台（只含 DS）
func TestRulePrimaryServiceOrderAndContent(t *testing.T) {
	servers := buildAllCombinations()
	groups := RenderDashboard(servers, RulePrimaryService)

	wantNames := []string{"AS", "TS", "SS", "DS"}
	wantCounts := []int{8, 4, 2, 1}

	if len(groups) != 4 {
		t.Fatalf("组数 = %d, want 4", len(groups))
	}
	for i, g := range groups {
		if g.Name != wantNames[i] {
			t.Errorf("第 %d 组名 = %s, want %s", i, g.Name, wantNames[i])
		}
		if len(g.Servers) != wantCounts[i] {
			t.Errorf("%s 组服务器数 = %d, want %d", g.Name, len(g.Servers), wantCounts[i])
		}
	}
}

// TestRulePrimaryServiceNoDuplicates 验证每台服务器只出现一次。
func TestRulePrimaryServiceNoDuplicates(t *testing.T) {
	servers := buildAllCombinations()
	groups := RenderDashboard(servers, RulePrimaryService)

	seen := make(map[string]string)
	for _, g := range groups {
		for _, s := range g.Servers {
			if prev, ok := seen[s.Name]; ok {
				t.Errorf("服务器 %s 重复出现在 %s 和 %s", s.Name, prev, g.Name)
			}
			seen[s.Name] = g.Name
		}
	}
	if len(seen) != len(servers) {
		t.Errorf("去重后服务器数 = %d, want %d（不应丢数据）", len(seen), len(servers))
	}
}

// TestRuleExactConventionOrder 锁定业务约定顺序。
func TestRuleExactConventionOrder(t *testing.T) {
	servers := buildAllCombinations()
	groups := RenderDashboard(servers, RuleExactConvention)

	got := make([]string, len(groups))
	for i, g := range groups {
		got[i] = g.Name
	}

	want := []string{
		"AS", "AS+SS", "AS+DS", "AS+SS+DS",
		"AS+TS", "AS+TS+SS", "AS+TS+DS", "AS+TS+SS+DS",
		"TS", "TS+SS", "TS+DS", "TS+SS+DS",
		"SS", "SS+DS",
		"DS",
	}
	if !slices.Equal(got, want) {
		t.Errorf("约定顺序不符\n got:  %v\n want: %v", got, want)
	}
}

// TestRuleExactCombinationOrder 验证组合优先（mask 降序）。
func TestRuleExactCombinationOrder(t *testing.T) {
	servers := buildAllCombinations()
	groups := RenderDashboard(servers, RuleExactCombination)

	for i := 1; i < len(groups); i++ {
		if groups[i-1].Key <= groups[i].Key {
			t.Errorf("第 %d 组 Key(%04b) 应 > 第 %d 组 Key(%04b)",
				i-1, groups[i-1].Key, i, groups[i].Key)
		}
	}
}

// TestRuleExactServiceCountOrder 验证单服务优先：数量升序。
func TestRuleExactServiceCountOrder(t *testing.T) {
	servers := buildAllCombinations()
	groups := RenderDashboard(servers, RuleExactServiceCount)

	prevCount := -1
	for i, g := range groups {
		count := bits.OnesCount(uint(g.Key))
		if count < prevCount {
			t.Errorf("第 %d 组数量(%d) 应 >= 前一组的数量(%d)", i, count, prevCount)
		}
		prevCount = count
	}
}

// ============================================================================
// 边界与防御性测试
// ============================================================================

// TestConventionCoverage 验证 comboOrderList 覆盖全部 15 种组合且无重复。
func TestConventionCoverage(t *testing.T) {
	if len(comboOrderList) != 15 {
		t.Errorf("comboOrderList 长度 = %d, want 15", len(comboOrderList))
	}
	seen := make(map[int]bool)
	for _, m := range comboOrderList {
		if seen[m] {
			t.Errorf("mask %04b (%s) 重复", m, comboName(m))
		}
		seen[m] = true
	}
	for m := 1; m < 16; m++ {
		if !seen[m] {
			t.Errorf("缺少 mask %04b (%s)", m, comboName(m))
		}
	}
}

// TestEmptyServers 验证空输入返回非 nil 空切片（JSON 输出 [] 而非 null）。
func TestEmptyServers(t *testing.T) {
	for _, rule := range []GroupRule{
		RuleExactCombination, RulePrimaryService,
	} {
		t.Run(rule.String(), func(t *testing.T) {
			groups := RenderDashboard(nil, rule)
			if groups == nil {
				t.Error("返回应为非 nil 空切片")
			}
			data, err := json.Marshal(groups)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "[]" {
				t.Errorf("JSON = %q, want \"[]\"", string(data))
			}
		})
	}
}

// TestEmptyServicesIncluded 验证无服务的服务器进入独立分组并保留在输出中。
func TestEmptyServicesIncluded(t *testing.T) {
	servers := []Server{
		{Name: "n1", Services: []Service{{ServiceType: AS}}},
		{Name: "n2", Services: nil},
		{Name: "n3", Services: []Service{}},
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
			var noServiceGroup *Group
			for i := range groups {
				g := &groups[i]
				total += len(g.Servers)
				if g.Key == 0 {
					noServiceGroup = g
					if i != len(groups)-1 {
						t.Errorf("无服务分组位置 = %d, want 最后一个分组", i)
					}
				}
			}
			if total != len(servers) {
				t.Fatalf("分组内服务器总数 = %d, want %d（不应遗漏无服务服务器）", total, len(servers))
			}
			if noServiceGroup == nil {
				t.Fatal("缺少 Key=0 的无服务分组")
			}
			if noServiceGroup.Name != noServiceGroupName {
				t.Errorf("无服务分组名称 = %q, want %q", noServiceGroup.Name, noServiceGroupName)
			}
			if len(noServiceGroup.Servers) != 2 {
				t.Fatalf("无服务分组服务器数 = %d, want 2", len(noServiceGroup.Servers))
			}
			gotNames := []string{noServiceGroup.Servers[0].Name, noServiceGroup.Servers[1].Name}
			if !slices.Contains(gotNames, "n2") || !slices.Contains(gotNames, "n3") {
				t.Errorf("无服务分组服务器 = %v, want n2 和 n3", gotNames)
			}
		})
	}
}

// TestGroupNoSideEffect 验证 Group 不修改输入数据。
func TestGroupNoSideEffect(t *testing.T) {
	servers := []Server{
		{
			Name:    "n1",
			Cluster: Cluster{ClusterID: 1, ClusterRole: "primary", ClusterName: "cluster-1"},
			State:   SrvNormal,
			Services: []Service{
				{ServiceType: DS, State: SvcAbnormal},
				{ServiceType: AS, State: SvcNormal},
			},
		}, // 服务故意乱序
	}
	orig0 := servers[0].Services[0] // 记录 DS

	var groups []Group
	for _, rule := range []GroupRule{
		RuleExactCombination, RulePrimaryService,
	} {
		groups = RenderDashboard(servers, rule)
	}

	if servers[0].Services[0] != orig0 {
		t.Errorf("Group 修改了输入数据！原 DS 被改成 %s", servers[0].Services[0].ServiceType)
	}
	if got := groups[0].Servers[0].Cluster; got != servers[0].Cluster {
		t.Errorf("分组结果未保留共享 model 的集群信息：got %+v, want %+v", got, servers[0].Cluster)
	}
}

// TestInvalidRulePanics 验证无效规则会 panic。
func TestInvalidRulePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("无效 GroupRule 应当 panic")
		}
	}()
	RenderDashboard(nil, GroupRule(999))
}

// ============================================================================
// 解析测试
// ============================================================================

// TestParseGroupRule 验证字符串到规则的解析。
func TestParseGroupRule(t *testing.T) {
	cases := map[string]GroupRule{
		"":                    RuleExactCombination,
		"exact-combination":   RuleExactCombination,
		"exact-service-count": RuleExactServiceCount,
		"exact-primary":       RuleExactPrimary,
		"exact-convention":    RuleExactConvention,
		"primary-service":     RulePrimaryService,
	}
	for in, want := range cases {
		got, err := ParseGroupRule(in)
		if err != nil {
			t.Fatalf("ParseGroupRule(%q) error: %v", in, err)
		}
		if got != want {
			t.Errorf("ParseGroupRule(%q) = %v, want %v", in, got, want)
		}
	}
	if _, err := ParseGroupRule("invalid"); err == nil {
		t.Error("无效输入应返回错误")
	}
}
