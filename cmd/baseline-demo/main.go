package main

import (
	"encoding/json"
	"fmt"

	"dashboardService/baseline"
)

// main 演示四部分：
//   - 第一部分：[]ClusterGroup 嵌套结构（五种规则的分组与排序）
//   - 第二部分：[]ClusterFlatGroup 完整版（全部服务器，JSON 输出）
//   - 第三部分：[]ClusterFlatGroup 预览版（前 3 台 + rest 字段，JSON 输出）
//   - 第四部分：组内角色排序（相同服务组合、不同角色时的排序规则）
func main() {
	servers := baseline.AllServiceCombinations()
	rules := []baseline.GroupRule{
		baseline.RuleCombination,
		baseline.RuleServiceCount,
		baseline.RulePrimary,
		baseline.RuleConvention,
		baseline.RulePrimaryService,
	}

	// ============================================================
	// 一、[]ClusterGroup 嵌套结构
	// ============================================================
	fmt.Println("================ 一、[]ClusterGroup 嵌套结构 ================")
	for _, rule := range rules {
		fmt.Printf("\n=== %s ===\n", rule)
		baseline.PrintClusters(baseline.RenderDashboard(servers, rule))
	}

	// ============================================================
	// 二、[]ClusterFlatGroup 完整版（全部服务器）
	// ============================================================
	// 用 RulePrimaryService 演示：4 组（AS/TS/SS/DS），数据量适中
	fmt.Println("\n================ 二、[]ClusterFlatGroup 完整版 ================")
	fmt.Println("（用 RulePrimaryService 演示，完整服务器列表）")
	primaryClusters := baseline.RenderDashboard(servers, baseline.RulePrimaryService)
	printJSON(baseline.Flatten(primaryClusters))

	// ============================================================
	// 三、[]ClusterFlatGroup 预览版（前 3 台 + rest 字段）
	// ============================================================
	// 用 RulePrimaryService + previewSize=3：
	//   AS 组 8 台 → AS 前 3 台 + restAS 后 5 台
	//   TS 组 4 台 → TS 前 3 台 + restTS 后 1 台
	//   SS 组 2 台 → SS 2 台（不触发 rest）
	//   DS 组 1 台 → DS 1 台（不触发 rest）
	//
	// 注意：具体是哪 3 台由 compareServers 决定（状态 → 服务列表 → IP → Name），
	// 不受分组规则影响。运行本程序即可看到实际顺序。
	fmt.Println("\n================ 三、[]ClusterFlatGroup 预览版（前 3 台） ================")
	printJSON(baseline.FlattenWithPreview(primaryClusters, 3))

	// ============================================================
	// 四、组内角色排序演示
	// ============================================================
	// 五台服务器服务组合均为 AS，角色不同：Leader > Follower > LSS > SSRole > DSRole。
	// IP 与角色反序（Leader IP 最大），证明排序依据是角色而非 IP。
	fmt.Println("\n================ 四、组内角色排序演示 ================")
	fmt.Println("五台服务器服务组合均为 AS，角色不同：Leader > Follower > LSS > SSRole > DSRole")
	fmt.Println("IP 与角色反序（Leader IP 最大），证明排序依据是角色而非 IP")
	fmt.Println()

	roleServers := baseline.RoleVariations()
	roleClusters := baseline.RenderDashboard(roleServers, baseline.RuleCombination)
	baseline.PrintClusters(roleClusters)
}

// printJSON 以缩进 JSON 打印，便于阅读。
//
// 生产环境的 API handler 用 json.NewEncoder(w).Encode(flat) 即可，
// 不需要缩进——缩进只是为了让演示输出可读。
func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("JSON 序列化失败: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
