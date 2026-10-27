package baseline_test

import (
	"fmt"

	"dashboardService/baseline"
)

// ExampleRenderDashboard 展示 switch 基线如何按业务规则分组并保留空服务服务器。
func ExampleRenderDashboard() {
	servers := []baseline.Server{
		{Name: "node-as", Services: []baseline.Service{{ServiceType: baseline.ASType, State: baseline.ServiceNormal}}},
		{Name: "node-empty"},
	}

	clusters := baseline.RenderDashboard(servers, baseline.RuleExactCombination)
	for _, cluster := range clusters {
		for _, group := range cluster.Groups {
			fmt.Printf("%s: %d 台服务器\n", group.Name, len(group.Servers))
		}
	}

	// Output:
	// AS: 1 台服务器
	// 无服务: 1 台服务器
}

// ExampleMarshalGroups 展示应用层直接将分组结果编码为 JSON。
func ExampleMarshalGroups() {
	data, err := baseline.MarshalGroups([]baseline.Group{{
		Name:    "AS",
		Servers: []baseline.Server{},
	}})
	if err != nil {
		fmt.Println("编码失败:", err)
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
