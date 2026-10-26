package options

import (
	"errors"
	"fmt"

	"dashboardService/baseline"
)

// Grouping 保存共享分组流程所需的函数配置。
// 配置由 NewGrouping 创建后固定，调用方不能通过公开 API 修改。
type Grouping struct {
	keyFn baseline.KeyFunc
	cmpFn baseline.CompareFunc
}

// Option 配置 Grouping 实例。
type Option func(*Grouping)

// WithKeyFn 设置分组键计算函数，覆盖默认的完整服务组合 mask。
func WithKeyFn(fn baseline.KeyFunc) Option {
	return func(grouping *Grouping) {
		grouping.keyFn = fn
	}
}

// WithCmpFn 设置组间排序比较函数，覆盖默认的 mask 降序。
func WithCmpFn(fn baseline.CompareFunc) Option {
	return func(grouping *Grouping) {
		grouping.cmpFn = fn
	}
}

// NewGrouping 创建一个配置固定的分组器。
// 默认按完整服务组合分组，并按 mask 降序排列，与 baseline 的默认规则一致。
// nil 选项、nil 键函数或 nil 比较器都会返回错误。
func NewGrouping(opts ...Option) (*Grouping, error) {
	grouping := &Grouping{
		keyFn: baseline.MaskOf,
		cmpFn: baseline.CompareByMaskDesc,
	}
	for i, option := range opts {
		if option == nil {
			return nil, fmt.Errorf("options.NewGrouping: option %d is nil", i)
		}
		option(grouping)
	}
	if grouping.keyFn == nil {
		return nil, errors.New("options.NewGrouping: key function is nil")
	}
	if grouping.cmpFn == nil {
		return nil, errors.New("options.NewGrouping: comparator is nil")
	}
	return grouping, nil
}

// Apply 使用已配置的函数执行共享分组、排序及未知服务处理流程。
func (grouping *Grouping) Apply(servers []Server) []Group {
	if grouping == nil {
		panic("options.Grouping.Apply: nil Grouping")
	}
	return baseline.GroupBy(servers, grouping.keyFn, grouping.cmpFn)
}

// ParseGroupRule 调用 baseline 解析外部规则名称。
func ParseGroupRule(value string) (GroupRule, error) {
	return baseline.ParseGroupRule(value)
}
