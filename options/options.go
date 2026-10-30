package options

import "dashboardService/baseline"

// Option 修改一个 RuleImpl。返回 Option 而非直接修改，
// 是为了让调用方可以用 opts ...Option 的形式表达"配置的组合过程"。
//
// 为什么是 func(*baseline.RuleImpl) 而非自定义类型：
//   - 复用 baseline.RuleImpl，避免引入额外的 "Grouping" 类型；
//   - Option 的产物就是 baseline.RuleImpl，最终直接传给 GroupServersByCluster；
//   - 与 datadriven 的 ruleImpls 表项类型相同，便于对比。
type Option func(*baseline.RuleImpl)

// WithKeyOf 设置分组键函数。
//
// 举例：
//
//	WithKeyOf(baseline.MaskOf)              // 按精确组合归组
//	WithKeyOf(baseline.PrimaryServiceMask)  // 按主服务归组
//
// 传入 nil 是编程错误，会立即 panic（fail-fast，不等 Build 时才发现）。
func WithKeyOf(fn baseline.KeyFunc) Option {
	if fn == nil {
		panic("options.WithKeyOf: KeyFunc 不能为 nil")
	}
	return func(r *baseline.RuleImpl) { r.KeyOf = fn }
}

// WithCmp 设置组间比较器。
//
// 举例：
//
//	WithCmp(baseline.CmpByMask)          // mask 降序
//	WithCmp(baseline.CmpByServiceCount)  // 服务数量升序
//
// 传入 nil 是编程错误，会立即 panic。
func WithCmp(fn baseline.CompareFunc) Option {
	if fn == nil {
		panic("options.WithCmp: CompareFunc 不能为 nil")
	}
	return func(r *baseline.RuleImpl) { r.Cmp = fn }
}

// newRule 把一组 Option 应用到默认配置上，返回最终的 RuleImpl。
//
// 默认值：
//   - KeyOf = baseline.MaskOf     （按精确组合归组）
//   - Cmp   = baseline.CmpByMask  （mask 降序）
//
// 不传任何 Option 时，结果等价于 RuleExactCombination。
// 后传入的 Option 覆盖先传入的相同字段——这与"选项"的直觉一致。
func newRule(opts []Option) baseline.RuleImpl {
	r := baseline.RuleImpl{
		KeyOf: baseline.MaskOf,
		Cmp:   baseline.CmpByMask,
	}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}
