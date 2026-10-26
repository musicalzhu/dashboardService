package datadriven

import "dashboardService/baseline"

// MarshalGroups 将分组结果序列化为带两空格缩进的 JSON 字节流，供应用层直接调用。
//
// 本函数只负责编码，不执行分组、排序或错误恢复；编码错误原样返回给调用方。
// 空的非 nil groups 编码为 []，无服务组的名称通过 name 字段输出，Key 不输出。
//
// 示例：一个无服务组会编码为：
//
//	[
//	  {
//	    "name": "无服务",
//	    "servers": []
//	  }
//	]
func MarshalGroups(groups []Group) ([]byte, error) {
	return baseline.MarshalGroups(groups)
}
