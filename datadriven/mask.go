package datadriven

import "strings"

// allServiceTypes 服务类型的优先级顺序，是全代码唯一权威来源。
// 索引越小优先级越高（AS=0 最高，DS=3 最低）。
var allServiceTypes = []ServiceType{AS, TS, SS, DS}

// ============================================================================
// 服务组合掩码
// ============================================================================
//
// 位布局（高优先级占据高位，是 mask 排序成立的前提）：
//
//	asBit = 1000
//	tsBit = 0100
//	ssBit = 0010
//	dsBit = 0001
//
// 用显式二进制字面量而非 iota，目的：
//  1. 一眼可见 asBit 是最高位，避免 1<<iota 从低位开始的陷阱
//  2. 直观表达"位 = 服务"的意图，无需推演 iota 序列
const (
	asBit = 0b1000 // AS 占据最高位
	tsBit = 0b0100 // TS
	ssBit = 0b0010 // SS
	dsBit = 0b0001 // DS 占据最低位
)

// serviceBit 服务类型到 mask 位的映射表。
var serviceBit = map[ServiceType]int{
	AS: asBit, TS: tsBit, SS: ssBit, DS: dsBit,
}

// comboOrderList 业务约定的精确组合展示顺序（RuleExactConvention 使用）。
//
// 业务规则（无公式可推导，纯业务约定）：
//   - AS 系列：不含 TS 的子集在前，含 TS 的在后
//   - TS 系列：TS → TS-SS → TS-DS → TS-SS-DS
//   - 整体：AS 组 → TS 组 → SS 组 → DS 组
//
// 修改后必须更新 TestConventionCoverage，防止 15 种组合遗漏或重复。
var comboOrderList = []int{
	// AS 系列（不含 TS）
	asBit,                 // AS
	asBit | ssBit,         // AS+SS
	asBit | dsBit,         // AS+DS
	asBit | ssBit | dsBit, // AS+SS+DS

	// AS 系列（含 TS）
	asBit | tsBit,                 // AS+TS
	asBit | tsBit | ssBit,         // AS+TS+SS
	asBit | tsBit | dsBit,         // AS+TS+DS
	asBit | tsBit | ssBit | dsBit, // AS+TS+SS+DS

	// TS 系列
	tsBit,                 // TS
	tsBit | ssBit,         // TS+SS
	tsBit | dsBit,         // TS+DS
	tsBit | ssBit | dsBit, // TS+SS+DS

	// SS 系列
	ssBit,         // SS
	ssBit | dsBit, // SS+DS

	// DS
	dsBit, // DS
}

// serviceTypeRank 返回服务类型的优先级索引（AS=0, TS=1, SS=2, DS=3）。
//
// 未知类型返回 len(allServiceTypes)，排在所有已知类型之后。
// 仅用于组内 Services 数组排序，不影响任何分组逻辑。
func serviceTypeRank(t ServiceType) int {
	for i, st := range allServiceTypes {
		if st == t {
			return i
		}
	}
	return len(allServiceTypes)
}

// primaryIndex 返回 mask 中最高优先级服务的索引；空 mask 返回 4。
//
// 例：
//
//	AS+SS (1010) → 0
//	TS+SS (0110) → 1
//	0000         → 4
func primaryIndex(mask int) int {
	for i, t := range allServiceTypes {
		if mask&serviceBit[t] != 0 {
			return i
		}
	}
	return len(allServiceTypes)
}

// primaryMask 返回 mask 中最高优先级服务的单个 bit；空 mask 返回 0。
//
// 例：
//
//	AS+TS+SS (1110) → 1000（AS 的 bit）
//	TS+SS    (0110) → 0100（TS 的 bit）
//	0000            → 0
func primaryMask(mask int) int {
	for _, t := range allServiceTypes {
		if b := serviceBit[t]; mask&b != 0 {
			return b
		}
	}
	return 0
}

// comboName 按优先级顺序拼接 mask 中的服务名。
//
// 例：
//
//	1100 → "AS+TS"
//	1000 → "AS"
//	0000 → ""
func comboName(mask int) string {
	var names []string
	for _, t := range allServiceTypes {
		if mask&serviceBit[t] != 0 {
			names = append(names, string(t))
		}
	}
	return strings.Join(names, "+")
}

// maskOf 计算服务器的服务组合 mask（OR 聚合，天然去重）。
//
// 同一服务类型多次出现也只贡献一个 bit。
// 返回值是精确组合 mask，用于前四种规则的分组键。
func maskOf(s Server) int {
	m := 0
	for _, svc := range s.Services {
		m |= serviceBit[svc.ServiceType]
	}
	return m
}

// primaryServiceMask 计算服务器"主服务"的单个 bit；无服务返回 0。
//
// 实现：先 maskOf 得到完整组合，再 primaryMask 取最高优先级位。
// 两步语义清晰："完整组合 → 主服务位"。
// 返回值用于第五种规则 RulePrimaryService 的分组键。
func primaryServiceMask(s Server) int {
	return primaryMask(maskOf(s))
}
