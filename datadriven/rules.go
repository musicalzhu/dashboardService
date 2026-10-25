package datadriven

import "fmt"

// ============================================================================
// 分组规则：五种并列
// ============================================================================

// GroupRule 分组规则，共五种。
//
// 前四种按"精确组合"分组（最多 15 组），只改变组间排序；
// 第五种按"主服务"分组（最多 4 组），改变了分组键。
type GroupRule int

const (
	// RuleExactCombination 精确组合，mask 数值降序。
	//
	// 顺序：AS+TS+SS+DS > AS+TS+SS > ... > AS > ... > DS
	// 适用：容量评估（优先关注服务最复杂的服务器）。
	RuleExactCombination GroupRule = iota

	// RuleExactServiceCount 精确组合，服务数量升序 → mask 降序。
	//
	// 顺序：AS > TS > SS > DS > AS+TS > ... > AS+TS+SS+DS
	// 适用：快速定位单职责节点（易维护的服务器）。
	RuleExactServiceCount

	// RuleExactPrimary 精确组合，主服务索引 → 数量 → mask 降序。
	//
	// 顺序：AS > AS+TS > AS+SS > ... > TS > ... > DS
	// 适用：按优先级分运维区（先处理 AS 相关告警）。
	RuleExactPrimary

	// RuleExactConvention 精确组合，由 comboOrderList 完全决定。
	//
	// 顺序：AS > AS+SS > AS+DS > AS+SS+DS > AS+TS > ... > DS
	// 适用：与业务方约定的展示顺序。
	RuleExactConvention

	// RulePrimaryService 按主服务分组，投影到主服务位 + mask 降序。
	//
	// 分组：AS、TS、SS、DS（4 组，含 AS 的全部合并到 AS 组）
	// 顺序：AS > TS > SS > DS
	//
	// 原理：primaryServiceMask 把服务器投影成主服务的单个 bit
	// （1000/0100/0010/0001），由于 AS 占最高位，对投影值做
	// b-a 降序恰好得到此顺序。
	RulePrimaryService
)

// String 返回分组规则的可读名称，用于日志和 API 参数。
func (r GroupRule) String() string {
	return [...]string{
		"exact-combination",
		"exact-service-count",
		"exact-primary",
		"exact-convention",
		"primary-service",
	}[r]
}

// ParseGroupRule 将 API 参数字符串解析为 GroupRule。
//
// 空字符串视为默认规则 RuleExactCombination，便于前端省略参数。
// 未知字符串返回错误，调用方应当处理。
func ParseGroupRule(s string) (GroupRule, error) {
	switch s {
	case "", "exact-combination":
		return RuleExactCombination, nil
	case "exact-service-count":
		return RuleExactServiceCount, nil
	case "exact-primary":
		return RuleExactPrimary, nil
	case "exact-convention":
		return RuleExactConvention, nil
	case "primary-service":
		return RulePrimaryService, nil
	}
	return 0, fmt.Errorf("未知分组规则: %q", s)
}

// ruleImpl 描述一种分组规则的实现：分组键计算 + 组间排序。
//
// 两个字段恰好对应"五种规则"的两个正交维度：
//   - keyOf: 从 Server 计算分组键
//   - cmp:   分组键的排序函数
type ruleImpl struct {
	keyOf func(Server) int
	cmp   func(a, b int) int
}

// ruleImpls 五种分组规则的实现表。
//
// 这张表就是五种规则的全部声明。一眼可看出：
//   - 前四行 keyOf 相同（maskOf）→ 精确组合分组，只是排序不同
//   - 第五行 keyOf 不同（primaryServiceMask）→ 主服务分组
//   - 第五行 cmp 与第一行相同（cmpByMaskDesc）→ 复用组合优先排序
//
// 新增规则时：
//  1. 增加一个 GroupRule 常量
//  2. 本表增加一行 entry
//  3. RenderDashboard 函数不用改
//
// 完整性由 TestAllRulesHaveImpl 保护：
//
//	新增常量却忘记加 entry 时，该测试会立即失败。
var ruleImpls = map[GroupRule]ruleImpl{
	RuleExactCombination:  {maskOf, cmpByMaskDesc},
	RuleExactServiceCount: {maskOf, cmpByServiceCount},
	RuleExactPrimary:      {maskOf, cmpByPrimary},
	RuleExactConvention:   {maskOf, cmpByConvention},
	RulePrimaryService:    {primaryServiceMask, cmpByMaskDesc},
}
