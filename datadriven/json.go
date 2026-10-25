package datadriven

import "encoding/json"

// marshalGroups 将分组结果序列化为 JSON 字节流。
//
// 与业务逻辑解耦：若前端需其他格式（Protobuf、CSV 等），只需替换本函数。
func marshalGroups(groups []Group) ([]byte, error) {
	return json.MarshalIndent(groups, "", "  ")
}
