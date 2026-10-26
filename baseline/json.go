package baseline

import "encoding/json"

// MarshalGroups 将分组结果编码为缩进 JSON，供应用层直接调用。
// 编码错误会原样返回给调用方。
func MarshalGroups(groups []Group) ([]byte, error) {
	return json.MarshalIndent(groups, "", "  ")
}
