package datadriven

import (
	"fmt"
)

// ExampleGroupRule_String 展示分组规则如何转换为稳定的字符串名称。
func ExampleGroupRule_String() {
	fmt.Println(RuleExactPrimary)

	// Output:
	// rule-primary
}

// ExampleParseGroupRule 展示外部规则名称的校验和默认规则解析。
func ExampleParseGroupRule() {
	rule, err := ParseGroupRule("rule-primary-service")
	if err != nil {
		fmt.Println("invalid rule:", err)
		return
	}
	fmt.Println(rule)

	defaultRule, err := ParseGroupRule("")
	if err != nil {
		fmt.Println("invalid rule:", err)
		return
	}
	fmt.Println(defaultRule)

	// Output:
	// rule-primary-service
	// rule-combination
}

// ExampleRenderDashboard 展示未知服务和没有运行服务的服务器都会保留在分组结果中。
func ExampleRenderDashboard() {
	servers := []Server{
		{
			Name: "node-as",
			Services: []Service{
				{ServiceType: AS, State: SvcNormal},
			},
		},
		{
			Name: "node-unknown",
			Services: []Service{
				{ServiceType: ServiceType("XX"), State: SvcNormal},
			},
		},
		{Name: "node-empty"},
	}

	clusters := RenderDashboard(servers, RuleExactCombination)
	for _, cluster := range clusters {
		for _, group := range cluster.Groups {
			fmt.Printf("%s: ", group.Name)
			for i, server := range group.Servers {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(server.Name)
			}
			fmt.Println()
		}
	}

	// Output:
	// AS: node-as
	// 未知服务: node-unknown
	// 无服务: node-empty
}

// ExampleMarshalGroups 展示应用层可直接调用的 JSON 编码边界。
func ExampleMarshalGroups() {
	data, err := MarshalGroups([]Group{{
		Name:    "AS",
		Servers: []Server{},
	}})
	if err != nil {
		fmt.Println("marshal error:", err)
		return
	}
	fmt.Println(string(data))

	// Output:
	// [
	//   {
	//     "name": "AS",
	//     "servers": []
	//   }
	// ]
}
