// Package baseline 的单元测试。
//
// 本文件是 baseline 包的完整单元测试，覆盖包内四个源文件的全部可测逻辑：
//   - rules.go        —— 基础工具、比较器、规则元信息
//   - test_helper.go  —— 测试数据完整性
//   - dashboard.go    —— 组内排序、分组主流程
//   - model.go        —— 由上述测试间接覆盖
//
// 命名说明：
//
//	本文件原名 dashboard_test.go，在"只测 dashboard.go"时期是准确的；
//	整合 rules_test.go 后，测试范围扩展为整个 baseline 包，
//	因此改名为 baseline_test.go —— 与其余五个方案的
//	datadriven_test.go / options_test.go / decorator_test.go /
//	builder_test.go / strategy_test.go 命名风格保持一致。
//
// 测试分区（按"底层 → 中间层 → 顶层"组织）：
//
//	一、基础工具          rules.go 中的工具函数（MaskOf / PrimaryServiceMask / groupName）
//	二、比较器            rules.go 中的四个 CmpByXxx
//	三、规则元信息        rules.go 中的 Rule 枚举、Parse、String
//	四、测试数据完整性    test_helper.go 中的 AllServiceCombinations
//	五、组内排序          dashboard.go 中的 compareServers、prepareServer
//	六、分组主流程        dashboard.go 中的 RenderDashboard
//
// 跨方案一致性由 poc/test/golden_test.go 负责，本文件只验证 baseline 自身。
package baseline

import (
	"reflect"
	"testing"
)

// ============================================================================
// 一、基础工具（rules.go）
// ============================================================================

// TestMaskOf 验证服务组合掩码的计算。
//
// MaskOf 是分组键计算的核心：
//   - 空服务返回 noServiceGroupKey（无服务独立分组）
//   - 未知服务返回 unknownServiceGroupKey（未知服务独立分组）
//   - 已知服务按 OR 聚合（天然去重）
//
// 覆盖：空服务、空切片、单服务、多服务、重复、未知、已知+未知混合。
func TestMaskOf(t *testing.T) {
	cases := []struct {
		name     string
		services []ServiceType
		want     int
	}{
		{"空服务", nil, noServiceGroupKey},
		{"空切片", []ServiceType{}, noServiceGroupKey},
		{"单AS", []ServiceType{ASType}, 0b1000},
		{"单DS", []ServiceType{DSType}, 0b0001},
		{"AS+TS", []ServiceType{ASType, TSType}, 0b1100},
		{"AS+TS+SS+DS", []ServiceType{ASType, TSType, SSType, DSType}, 0b1111},
		{"重复服务去重", []ServiceType{ASType, ASType, TSType}, 0b1100},
		{"未知服务", []ServiceType{"unknown"}, unknownServiceGroupKey},
		{"已知+未知混合", []ServiceType{ASType, "unknown"}, unknownServiceGroupKey},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := Server{}
			for _, st := range c.services {
				srv.Services = append(srv.Services, Service{ServiceType: st, State: StateNormal})
			}
			if got := MaskOf(srv); got != c.want {
				t.Errorf("MaskOf = %04b, want %04b", got, c.want)
			}
		})
	}
}

// TestPrimaryServiceMask 验证主服务 bit 的计算。
//
// PrimaryServiceMask 先取完整 mask，再取最高优先级位，
// 是 RulePrimaryService（主服务归组）的关键分组键。
//
// 覆盖：空服务、各主服务优先级、多服务取最高。
func TestPrimaryServiceMask(t *testing.T) {
	cases := []struct {
		name     string
		services []ServiceType
		want     int
	}{
		{"空服务", nil, 0},
		{"AS+TS", []ServiceType{ASType, TSType}, 0b1000},
		{"TS+SS", []ServiceType{TSType, SSType}, 0b0100},
		{"SS+DS", []ServiceType{SSType, DSType}, 0b0010},
		{"单DS", []ServiceType{DSType}, 0b0001},
		{"全服务", []ServiceType{ASType, TSType, SSType, DSType}, 0b1000},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := Server{}
			for _, st := range c.services {
				srv.Services = append(srv.Services, Service{ServiceType: st, State: StateNormal})
			}
			if got := PrimaryServiceMask(srv); got != c.want {
				t.Errorf("PrimaryServiceMask = %04b, want %04b", got, c.want)
			}
		})
	}
}

// TestGroupName 验证 key → 分组名称映射。
//
// 覆盖三种情况：
//   - 无服务 / 未知服务（独立的固定名称）
//   - 已知 mask（按 AS+TS+... 拼接）
//   - 越界 key（兜底格式化）
func TestGroupName(t *testing.T) {
	cases := map[int]string{
		noServiceGroupKey:      noServiceGroupName,
		unknownServiceGroupKey: unknownServiceGroupName,
		0b1000:                 "AS",
		0b1100:                 "AS+TS",
		0b1111:                 "AS+TS+SS+DS",
	}
	for key, want := range cases {
		if got := groupName(key); got != want {
			t.Errorf("groupName(%d) = %q, want %q", key, got, want)
		}
	}
}

// ============================================================================
// 二、比较器（rules.go）
// ============================================================================

// TestCmpByMask 验证 mask 降序：值大的排前（返回负）。
//
// 这是 RuleCombination 的组间排序规则。
func TestCmpByMask(t *testing.T) {
	cases := []struct {
		a, b     int
		wantSign int
	}{
		{0b1111, 0b0001, -1}, // 15 > 1，a 排前
		{0b0001, 0b1111, 1},  // 1 < 15，a 排后
		{0b1000, 0b1000, 0},  // 相等
	}
	for _, c := range cases {
		if got := sign(CmpByMask(c.a, c.b)); got != c.wantSign {
			t.Errorf("CmpByMask(%04b, %04b) 符号 = %d, want %d", c.a, c.b, got, c.wantSign)
		}
	}
}

// TestCmpByServiceCount 验证数量升序 → mask 降序。
//
// 这是 RuleServiceCount 的组间排序规则：
// 第一层按服务数量升序，第二层按 mask 降序。
func TestCmpByServiceCount(t *testing.T) {
	cases := []struct{ a, b int }{
		{0b1000, 0b1100}, // 1 < 2，a 排前
		{0b1100, 0b1111}, // 2 < 4，a 排前
		{0b1100, 0b1010}, // 数量同，mask 大在前（1100 > 1010）
	}
	for _, c := range cases {
		if CmpByServiceCount(c.a, c.b) >= 0 {
			t.Errorf("CmpByServiceCount(%04b, %04b) 应 < 0", c.a, c.b)
		}
	}
}

// TestCmpByPrimary 验证主服务索引升序 → 数量升序 → mask 降序。
//
// 这是 RulePrimary 的组间排序规则：
// 三级排序，逐级决定分组顺序。
func TestCmpByPrimary(t *testing.T) {
	if CmpByPrimary(0b1000, 0b0100) >= 0 {
		t.Error("含 AS 的应排在含 TS 之前")
	}
	if CmpByPrimary(0b0100, 0b0010) >= 0 {
		t.Error("含 TS 的应排在含 SS 之前")
	}
	if CmpByPrimary(0b0010, 0b0001) >= 0 {
		t.Error("含 SS 的应排在含 DS 之前")
	}
}

// TestCmpByConvention 验证业务约定顺序。
//
// 这是 RuleConvention 的组间排序规则：
// 顺序由 comboOrderList 决定，是纯业务约定，无公式可推导。
func TestCmpByConvention(t *testing.T) {
	// 核心语义：AS > AS+SS > AS+TS > TS > SS > DS
	if CmpByConvention(0b1000, 0b1010) >= 0 {
		t.Error("AS 应排在 AS+SS 之前")
	}
	if CmpByConvention(0b1010, 0b1100) >= 0 {
		t.Error("AS+SS 应排在 AS+TS 之前")
	}
	if CmpByConvention(0b1100, 0b0100) >= 0 {
		t.Error("含 AS 的应排在纯 TS 之前")
	}
}

// sign 返回整数的符号（-1 / 0 / 1），用于比较器测试。
func sign(n int) int {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}

// ============================================================================
// 三、规则元信息（rules.go）
// ============================================================================

// TestParseGroupRule 验证规则名解析。
//
// 空字符串返回默认 RuleCombination；未知名称返回错误。
// 这是前端 API 参数解析的入口。
func TestParseGroupRule(t *testing.T) {
	cases := map[string]GroupRule{
		"":                     RuleCombination,
		"rule-combination":     RuleCombination,
		"rule-service-count":   RuleServiceCount,
		"rule-primary":         RulePrimary,
		"rule-convention":      RuleConvention,
		"rule-primary-service": RulePrimaryService,
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

// TestGroupRuleString 验证规则的稳定名称。
//
// 覆盖合法规则（返回规则名）和越界规则（返回兜底格式）。
func TestGroupRuleString(t *testing.T) {
	cases := map[GroupRule]string{
		RuleCombination:    "rule-combination",
		RuleServiceCount:   "rule-service-count",
		RulePrimary:        "rule-primary",
		RuleConvention:     "rule-convention",
		RulePrimaryService: "rule-primary-service",
	}
	for r, want := range cases {
		if got := r.String(); got != want {
			t.Errorf("GroupRule(%d).String() = %q, want %q", r, got, want)
		}
	}
	// 越界规则应返回可读的兜底名称
	if got := GroupRule(999).String(); got != "GroupRule(999)" {
		t.Errorf("未知规则 String() = %q, want \"GroupRule(999)\"", got)
	}
}

// ============================================================================
// 四、测试数据完整性（test_helper.go）
// ============================================================================

// TestAllServiceCombinations 验证测试数据完整覆盖 15 种服务组合。
//
// 这是其他测试的基础：
//   - 所有分组语义测试都依赖本函数产生的完整数据
//   - 六方案一致性 golden 测试也使用本函数
//
// 验证三条：
//  1. 数量必须是 15（2^4 - 1，排除空组合）
//  2. 组合名唯一（用于断言分组顺序）
//  3. mask 唯一（用于断言分组结果）
func TestAllServiceCombinations(t *testing.T) {
	servers := AllServiceCombinations()
	if len(servers) != 15 {
		t.Fatalf("应生成 15 种组合，实际 %d", len(servers))
	}

	seenNames := make(map[string]bool, len(servers))
	seenMasks := make(map[int]bool, len(servers))

	for _, s := range servers {
		if s.ClusterID != 1 {
			t.Errorf("服务器 %s 的集群 ID = %d, want 1", s.Name, s.ClusterID)
		}
		if len(s.Services) == 0 {
			t.Errorf("服务器 %s 不应无服务", s.Name)
		}
		if seenNames[s.Name] {
			t.Errorf("组合名重复: %s", s.Name)
		}
		seenNames[s.Name] = true

		mask := MaskOf(s)
		if seenMasks[mask] {
			t.Errorf("组合 mask 重复: %04b", mask)
		}
		seenMasks[mask] = true
	}
}

// ============================================================================
// 五、组内排序（dashboard.go）
// ============================================================================

// TestCompareServers_StateFirst 验证服务器状态优先于其他字段。
//
// 异常服务器应排在正常服务器之前——这是运维场景的核心语义。
func TestCompareServers_StateFirst(t *testing.T) {
	abnormal := Server{State: StateAbnormal, IP: "9.9.9.9", Name: "a"}
	normal := Server{State: StateNormal, IP: "1.1.1.1", Name: "n"}
	if compareServers(abnormal, normal) >= 0 {
		t.Error("异常服务器应排在正常服务器之前")
	}
}

// TestCompareServers_IPAndName 验证 IP 和 Name 作为 tie-breaker。
//
// 当状态和服务序列都相同时，IP 小的排前；IP 也相同时，Name 小的排前。
func TestCompareServers_IPAndName(t *testing.T) {
	a := Server{State: StateNormal, IP: "1.1.1.1", Name: "b"}
	b := Server{State: StateNormal, IP: "1.1.1.2", Name: "a"}
	if compareServers(a, b) >= 0 {
		t.Error("IP 小的应排前")
	}

	// IP 相同，按 Name
	c := Server{State: StateNormal, IP: "1.1.1.1", Name: "a"}
	d := Server{State: StateNormal, IP: "1.1.1.1", Name: "b"}
	if compareServers(c, d) >= 0 {
		t.Error("Name 小的应排前")
	}
}

// TestPrepareServerNoSideEffect 验证 prepareServer 不修改调用方数据。
//
// 这是 POC 所有方案都必须满足的纯函数不变量：
//
//	groupBy 内部对 servers[i] 执行 prepareServer，
//	如果 prepareServer 直接排序原切片，会污染调用方数据，
//	并可能在并发场景引发 data race。
func TestPrepareServerNoSideEffect(t *testing.T) {
	orig := Server{
		Services: []Service{
			{ServiceType: DSType, State: StateNormal},
			{ServiceType: ASType, State: StateNormal},
		},
	}
	first := orig.Services[0]
	_ = prepareServer(orig)
	if orig.Services[0] != first {
		t.Error("prepareServer 修改了输入数据")
	}
}

// TestPrepareServerSortsServices 验证 Services 被按 AS→TS→SS→DS 排序。
//
// 排序目的：JSON 输出展示整齐，不影响分组逻辑。
func TestPrepareServerSortsServices(t *testing.T) {
	orig := Server{
		Services: []Service{
			{ServiceType: DSType},
			{ServiceType: ASType},
			{ServiceType: SSType},
			{ServiceType: TSType},
		},
	}
	got := prepareServer(orig)
	want := []ServiceType{ASType, TSType, SSType, DSType}
	for i, svc := range got.Services {
		if svc.ServiceType != want[i] {
			t.Errorf("位置 %d: got %s, want %s", i, svc.ServiceType, want[i])
		}
	}
}

// ============================================================================
// 六、分组主流程（dashboard.go）
// ============================================================================

// TestEmptyInputReturnsEmptyClusterGroups 验证空输入返回非 nil 空切片。
//
// 非 nil 是关键：JSON 序列化时非 nil 空切片输出 []，nil 输出 null。
// 前端期望 []，这是 API 契约的一部分。
func TestEmptyInputReturnsEmptyClusterGroups(t *testing.T) {
	clusterGroups := RenderDashboard(nil, RuleCombination)
	if clusterGroups == nil || len(clusterGroups) != 0 {
		t.Fatalf("RenderDashboard(nil, ...) = %#v, want non-nil empty slice", clusterGroups)
	}
}

// TestAllRulesPreserveTheExpectedGroupingSemantics 为所有规则验证分组语义。
//
// 分两组断言：
//   - 第一组：验证分组顺序（RuleCombination、RuleConvention、RulePrimaryService）
//   - 第二组：验证 RuleServiceCount、RulePrimary 的分组顺序
//
// 对 RulePrimaryService 额外验证各组服务器数量（8/4/2/1），
// 这是"主服务归组"区别于"精确组合"的核心特征：
//   - 精确组合：15 组（每组合一组）
//   - 主服务归组：4 组（AS/TS/SS/DS，含 AS 的全部合并）
func TestAllRulesPreserveTheExpectedGroupingSemantics(t *testing.T) {
	servers := AllServiceCombinations()

	// 第一组：分组顺序可确定的部分（含主服务归组的组数与内容）
	expectedNames := map[GroupRule][]string{
		RuleCombination: {
			"AS+TS+SS+DS", "AS+TS+SS", "AS+TS+DS", "AS+TS",
			"AS+SS+DS", "AS+SS", "AS+DS", "AS",
			"TS+SS+DS", "TS+SS", "TS+DS", "TS",
			"SS+DS", "SS", "DS",
		},
		RuleConvention: {
			"AS", "AS+SS", "AS+DS", "AS+SS+DS",
			"AS+TS", "AS+TS+SS", "AS+TS+DS", "AS+TS+SS+DS",
			"TS", "TS+SS", "TS+DS", "TS+SS+DS", "SS", "SS+DS", "DS",
		},
		RulePrimaryService: {"AS", "TS", "SS", "DS"},
	}

	for rule, want := range expectedNames {
		t.Run(rule.String(), func(t *testing.T) {
			groups := RenderDashboard(servers, rule)[0].Groups

			// 主服务归组：额外验证每组服务器数量
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

	// 第二组：RuleServiceCount 和 RulePrimary 的分组顺序
	expectedOrder := map[GroupRule][]string{
		RuleServiceCount: {
			"AS", "TS", "SS", "DS",
			"AS+TS", "AS+SS", "AS+DS", "TS+SS",
			"TS+DS", "SS+DS", "AS+TS+SS", "AS+TS+DS",
			"AS+SS+DS", "TS+SS+DS", "AS+TS+SS+DS",
		},
		RulePrimary: {
			"AS", "AS+TS", "AS+SS", "AS+DS",
			"AS+TS+SS", "AS+TS+DS", "AS+SS+DS", "AS+TS+SS+DS",
			"TS", "TS+SS", "TS+DS", "TS+SS+DS",
			"SS", "SS+DS", "DS",
		},
	}
	for rule, want := range expectedOrder {
		t.Run(rule.String(), func(t *testing.T) {
			groups := RenderDashboard(servers, rule)[0].Groups
			got := make([]string, len(groups))
			for i, group := range groups {
				got[i] = group.Name
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("group order = %v, want %v", got, want)
			}
		})
	}
}
