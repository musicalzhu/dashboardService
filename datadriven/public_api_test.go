package datadriven_test

import (
	"testing"

	"dashboardService/datadriven"
)

// TestMarshalGroupsIsPublic 验证包外应用可以调用 JSON 集成边界。
func TestMarshalGroupsIsPublic(t *testing.T) {
	data, err := datadriven.MarshalGroups([]datadriven.Group{{
		Name:    "AS",
		Servers: []datadriven.Server{},
	}})
	if err != nil {
		t.Fatalf("MarshalGroups() error: %v", err)
	}
	if string(data) != "[\n  {\n    \"name\": \"AS\",\n    \"servers\": []\n  }\n]" {
		t.Fatalf("MarshalGroups() = %s", data)
	}
}
