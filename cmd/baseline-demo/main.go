package main

import (
	"encoding/json"
	"fmt"

	"dashboardService/baseline"
)

// main 演示三部分：
//   - 第一部分：[]ClusterGroup 嵌套结构（原有演示）
//   - 第二部分：[]ClusterFlatGroup 完整版（全部服务器）
//   - 第三部分：[]ClusterFlatGroup 预览版（前 3 台 + rest 字段）
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
	clusters := baseline.RenderDashboard(servers, baseline.RulePrimaryService)
	printJSON(baseline.Flatten(clusters))

	// ============================================================
	// 三、[]ClusterFlatGroup 预览版（前 3 台 + rest 字段）
	// ============================================================
	// 用 RulePrimaryService + previewSize=3：
	fmt.Println("\n================ 三、[]ClusterFlatGroup 预览版（前 3 台） ================")
	printJSON(baseline.FlattenWithPreview(clusters, 3))
}

// printJSON 以缩进 JSON 打印，便于阅读。
func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("JSON 序列化失败: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
