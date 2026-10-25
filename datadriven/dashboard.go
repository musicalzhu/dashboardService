package datadriven

import (
	"fmt"
	"slices"
	"strings"
)

// ============================================================================
// 领域模型
// ============================================================================

// Group 表示一个分组，是 API 输出的顶层单元。
//
// Key 的语义随规则变化：
//   - 前四种规则（RuleExact*）：Key = 精确组合 mask（如 1010 = AS+SS）
//   - 第五种规则（RulePrimaryService）：Key = 主服务 bit（如 1000 = AS 组）
type Group struct {
	Key     int      `json:"-"`       // 分组键，内部使用，不输出
	Name    string   `json:"name"`    // 组合名称，如 "AS+TS"
	Servers []Server `json:"servers"` // 分组内的服务器列表（已排序）
}

const noServiceGroupName = "无服务"

// ============================================================================
// 分组骨架（执行层）
// ============================================================================

// groupBy 是分组流程的唯一骨架，所有规则共享。
//
// 流程四个阶段：
//  1. 分桶：遍历服务器，clone Services，按分组键聚合
//  2. 组内排序：每组内服务器按 compareServers 排列
//  3. 组间排序：按 cmp 排列分组键
//  4. 组装：把有序分组键转为 []Group
//
// 参数：
//   - servers: 待分组的服务器列表
//   - keyOf:   从 Server 计算分组键；返回 0 表示该服务器不参与分组
//   - cmp:     分组键的排序函数
//
// 纯函数：不修改输入 servers 的任何数据（含 Services 底层数组）。
// Services 为空的服务器会进入 Key=0 的“无服务”分组；有 Services 但没有可识别服务类型的服务器仍会跳过。
func groupBy(servers []Server, keyOf func(Server) int, cmp func(a, b int) int) []Group {
	// 阶段1：按分组键分桶
	byKey := make(map[int][]Server)
	for _, srv := range servers {
		srv = prepareServer(srv) // clone + 服务排序，避免副作用
		if len(srv.Services) == 0 {
			byKey[0] = append(byKey[0], srv)
			continue
		}

		key := keyOf(srv)
		if key == 0 {
			continue // 有服务列表但无可识别服务类型，不参与服务分组
		}
		byKey[key] = append(byKey[key], srv)
	}

	// 阶段2：组内排序 + 收集分组键
	keys := make([]int, 0, len(byKey))
	for k, group := range byKey {
		slices.SortFunc(group, compareServers)
		keys = append(keys, k)
	}

	// 阶段3：组间排序
	slices.SortFunc(keys, func(a, b int) int {
		if a == 0 && b != 0 {
			return 1
		}
		if b == 0 && a != 0 {
			return -1
		}
		return cmp(a, b)
	})

	// 阶段4：组装
	groups := make([]Group, len(keys))
	for i, k := range keys {
		name := comboName(k)
		if k == 0 {
			name = noServiceGroupName
		}
		groups[i] = Group{
			Key:     k,
			Name:    name,
			Servers: byKey[k],
		}
	}
	return groups
}

// ============================================================================
// 统一入口（规则分发层）
// ============================================================================

// RenderDashboard 按指定规则分组服务器，是唯一对外入口。
//
// 规则分发由 ruleImpls 表完成：
//  1. 查表得到 (keyOf, cmp)
//  2. 交给 groupBy 执行
//
// 无效规则是编程错误，会 panic（fail-fast，不是运行时错误）。
//
// 参数：
//   - servers: 待分组的服务器列表
//   - rule:    分组规则
//
// 返回：按规则排序的分组列表，每组内服务器已排序。
func RenderDashboard(servers []Server, rule GroupRule) []Group {
	impl, ok := ruleImpls[rule]
	if !ok {
		panic(fmt.Sprintf("RenderDashboard: 无效的 GroupRule %d", int(rule)))
	}
	return groupBy(servers, impl.keyOf, impl.cmp)
}

// ============================================================================
// 演示
// ============================================================================

// main 演示五种规则的输出差异。
func main() {
	servers := buildAllCombinations()

	for _, rule := range []GroupRule{
		RuleExactCombination,
		RuleExactServiceCount,
		RuleExactPrimary,
		RuleExactConvention,
		RulePrimaryService,
	} {
		groups := RenderDashboard(servers, rule)
		names := make([]string, len(groups))
		for i, g := range groups {
			names[i] = g.Name
		}
		fmt.Printf("[%-20s] %2d 组: %s\n", rule, len(groups), strings.Join(names, " > "))
	}
}

// buildAllCombinations 构造覆盖全部 15 种服务组合的演示数据。
//
// 仅用于演示和测试，不参与生产逻辑。
func buildAllCombinations() []Server {
	var servers []Server
	for _, mask := range comboOrderList {
		var svcs []Service
		for _, t := range allServiceTypes {
			if mask&serviceBit[t] != 0 {
				svcs = append(svcs, Service{ServiceType: t, State: SvcNormal})
			}
		}
		servers = append(servers, Server{
			Name:     comboName(mask),
			IP:       "1.1.1.1",
			State:    SrvNormal,
			Services: svcs,
		})
	}
	return servers
}
