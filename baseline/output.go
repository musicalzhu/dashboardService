package baseline

// ClusterFlatGroup 是平铺后的集群结构：集群元信息 + 分组顺序 + 分组数据。
//
// 两种输出形式：
//
//	Flatten:             每个分组字段包含全部服务器（完整版）
//	FlattenWithPreview:  每个分组字段只含前 previewSize 台，
//	                     剩余放入 "rest{分组名}" 字段（预览版）
//
// 完整版 JSON：
//
//	{
//	  "cluster_id": 1,
//	  "order": ["AS", "TS"],
//	  "AS": [全部服务器],
//	  "TS": [全部服务器]
//	}
//
// 预览版 JSON（previewSize=3）：
//
//	{
//	  "cluster_id": 1,
//	  "order": ["AS", "TS"],
//	  "AS":     [前 3 台],
//	  "restAS": [第 4 台起],
//	  "TS":     [前 3 台]
//	}
//
// 为什么用 map 而非 struct：分组名是动态的，struct 不支持动态字段。
//
// 为什么需要 order 字段：
//
//	Go 的 encoding/json 序列化 map 时按 key 字典序输出，
//	JSON 规范（RFC 8259）规定 object 是无序集合。
//	order 显式记录分组顺序，前端按 order 遍历即可正确渲染。
type ClusterFlatGroup map[string]any

// Flatten 把 []ClusterGroup 平铺为 []ClusterFlatGroup。
//
// 每个分组字段包含完整服务器列表，不截断。
// 与 FlattenWithPreview 的区别：
//   - Flatten:           完整数据，适合导出 / 审计
//   - FlattenWithPreview: 截断预览，适合前端渲染
//
// 纯格式转换：
//   - 不过滤任何分组（无服务 / 未知服务 / 普通分组一视同仁）
//   - 不修改服务器列表顺序
func Flatten(clusters []ClusterGroup) []ClusterFlatGroup {
	result := make([]ClusterFlatGroup, len(clusters))
	for i, c := range clusters {
		flat := make(ClusterFlatGroup, 5+len(c.Groups))
		flat["cluster_id"] = c.ClusterID
		flat["cluster_role"] = c.ClusterRole
		flat["cluster_name"] = c.ClusterName
		flat["state"] = c.State

		order := make([]string, len(c.Groups))
		for j, g := range c.Groups {
			order[j] = g.Name
			flat[g.Name] = g.Servers
		}
		flat["order"] = order

		result[i] = flat
	}
	return result
}

// FlattenWithPreview 把 []ClusterGroup 平铺为 []ClusterFlatGroup，
// 每个分组只保留前 previewSize 台服务器，剩余放入 "rest{分组名}" 字段。
//
// 转换规则：
//   - 分组服务器数 <= previewSize：完整放入 "分组名" 字段，不输出 rest 字段
//   - 分组服务器数 >  previewSize：前 previewSize 台放入 "分组名" 字段，
//                                  剩余放入 "rest{分组名}" 字段
//
// 示例（previewSize=3，AS 组 5 台）：
//
//	"AS":     [前 3 台]       // 3 台
//	"restAS": [后 2 台]       // 2 台
//
// 前端消费：
//
//	for (const groupName of cluster.order) {
//	    renderGroup(groupName, cluster[groupName]);          // 前 3 台
//	    const rest = cluster["rest" + groupName];            // 剩余（可能不存在）
//	    if (rest && rest.length > 0) {
//	        renderMoreButton(groupName, rest);               // "..." 按钮
//	    }
//	}
//
// 为什么不输出空 rest 字段：
//   - 只有存在剩余服务器时才输出 "rest{分组名}"，
//     前端通过"字段是否存在"即可判断是否需要显示"更多"
//   - 避免 JSON 膨胀（每个分组一个空数组 × 几十个分组）
//
// previewSize <= 0 是编程错误，会 panic（fail-fast）。
func FlattenWithPreview(clusters []ClusterGroup, previewSize int) []ClusterFlatGroup {
	if previewSize <= 0 {
		panic("baseline.FlattenWithPreview: previewSize 必须大于 0")
	}

	result := make([]ClusterFlatGroup, len(clusters))
	for i, c := range clusters {
		flat := make(ClusterFlatGroup, 5+2*len(c.Groups))
		flat["cluster_id"] = c.ClusterID
		flat["cluster_role"] = c.ClusterRole
		flat["cluster_name"] = c.ClusterName
		flat["state"] = c.State

		order := make([]string, len(c.Groups))
		for j, g := range c.Groups {
			order[j] = g.Name
			if len(g.Servers) <= previewSize {
				flat[g.Name] = g.Servers
				continue
			}
			flat[g.Name] = g.Servers[:previewSize]
			flat["rest"+g.Name] = g.Servers[previewSize:]
		}
		flat["order"] = order

		result[i] = flat
	}
	return result
}
