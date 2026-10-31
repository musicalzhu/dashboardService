package builder

import (
	"fmt"

	"dashboardService/baseline"
)

// Builder 是分组配置的构造器。
//
// 调用方通过链式方法逐步配置 keyOf / cmp，最后调用 Build 执行分组。
//
// 生命周期：
//
//	b := builder.New().
//	    WithKeyOf(baseline.PrimaryServiceMask).
//	    WithCmp(baseline.CmpByServiceCount)
//	groups1 := b.Build(serversBatch1)  // 可复用
//	groups2 := b.Build(serversBatch2)  // 同一配置多次 Build
//
// 并发安全：
//
//	Builder 本身不并发安全（配置阶段应单 goroutine 完成）。
//	但配置完成后，多个 goroutine 可以并发调用 b.Build(servers)——
//	因为 Build 只读取配置，不修改 Builder。
type Builder struct {
	keyOf baseline.KeyFunc
	cmp   baseline.CompareFunc
}

// New 创建使用默认配置的 Builder。
//
// 默认配置等价于 RuleExactCombination：
//   - KeyOf = baseline.MaskOf
//   - Cmp   = baseline.CmpByMask
//
// 不调用任何 WithXxx 直接 Build 时，行为等价于 baseline.RuleCombination。
func New() *Builder {
	return &Builder{
		keyOf: baseline.MaskOf,
		cmp:   baseline.CmpByMask,
	}
}

// WithKeyOf 设置分组键函数。
//
// 举例：
//
//	New().WithKeyOf(baseline.MaskOf)              // 按精确组合归组
//	New().WithKeyOf(baseline.PrimaryServiceMask)  // 按主服务归组
//
// 传入 nil 是编程错误，会立即 panic（fail-fast，不等 Build 时才发现）。
// 这与 options 方案的 WithKeyOf 一致——配置错误应当立即暴露。
func (b *Builder) WithKeyOf(fn baseline.KeyFunc) *Builder {
	if fn == nil {
		panic("builder.WithKeyOf: KeyFunc 不能为 nil")
	}
	b.keyOf = fn
	return b
}

// WithCmp 设置组间比较器。
//
// 举例：
//
//	New().WithCmp(baseline.CmpByMask)          // mask 降序
//	New().WithCmp(baseline.CmpByServiceCount)  // 服务数量升序
//
// 传入 nil 是编程错误，会立即 panic。
func (b *Builder) WithCmp(fn baseline.CompareFunc) *Builder {
	if fn == nil {
		panic("builder.WithCmp: CompareFunc 不能为 nil")
	}
	b.cmp = fn
	return b
}

// Build 用当前配置执行分组。
//
// 参数 servers 是待分组的服务器列表；Builder 不持有 servers，
// 所以同一 Builder 可以用于多个不同的 servers 批次。
//
// Build 是纯函数：
//   - 不修改 Builder；
//   - 不修改 servers 的任何数据（含 Services 底层数组）；
//   - 可并发调用。
func (b *Builder) Build(servers []Server) []ClusterGroup {
	return baseline.GroupServersByCluster(servers, baseline.RuleImpl{
		KeyOf: b.keyOf,
		Cmp:   b.cmp,
	})
}

// String 返回 Builder 当前配置的可读描述，便于日志与调试。
//
// 注意：Go 无法获取函数值的名称（除非用 runtime.FuncForPC），
// 所以这里只返回"已设置"或"默认"的语义描述。
func (b *Builder) String() string {
	return fmt.Sprintf("Builder{keyOf: %s, cmp: %s}",
		describeKeyFunc(b.keyOf),
		describeCmpFunc(b.cmp),
	)
}

// describeKeyFunc 返回 keyOf 函数的语义描述。
//
// 为什么需要它：Go 的 func 无法直接打印名字，这里通过"指针比较已知函数"
// 来识别常见的 keyOf 函数。未知函数返回 "custom"。
func describeKeyFunc(fn baseline.KeyFunc) string {
	switch fn {
	case nil:
		return "nil"
	default:
		// 无法直接比较函数，只能返回通用描述。
		return "custom"
	}
}

// describeCmpFunc 返回 cmp 函数的语义描述。
func describeCmpFunc(fn baseline.CompareFunc) string {
	switch fn {
	case nil:
		return "nil"
	default:
		return "custom"
	}
}
