package datadriven

import (
	"math/bits"
	"slices"
	"strings"
)

// prepareServer 克隆 Services 并按优先级排序。
//
// 必须 clone 的原因：
//
//	range 得到的 srv.Services 是 slice header 拷贝，底层数组与输入共享。
//	后续排序会污染上游数据，并可能在并发场景引发 data race。
//	clone 是纳秒级成本，彻底切断共享。
//
// Services 排序的目的：
//
//	仅用于 JSON 输出展示整齐（AS → TS → SS → DS）。
func prepareServer(s Server) Server {
	s.Services = slices.Clone(s.Services)
	slices.SortFunc(s.Services, func(a, b Service) int {
		return serviceTypeRank(a.ServiceType) - serviceTypeRank(b.ServiceType)
	})
	return s
}

// findService 按类型查找服务，未找到返回 nil。
//
// 返回的是切片元素的指针，调用方只读使用，不得修改。
func findService(s *Server, t ServiceType) *Service {
	for i := range s.Services {
		if s.Services[i].ServiceType == t {
			return &s.Services[i]
		}
	}
	return nil
}

// ============================================================================
// 组内服务器排序
// ============================================================================

// compareServiceStatus 比较两个同类型服务的状态。
//
// 三种情况：
//   - 双方都无此服务 → 相等
//   - 仅一方有 → 有的一方排前
//   - 双方都有 → normal 优先
func compareServiceStatus(a, b *Service) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return 1
	case b == nil:
		return -1
	}
	if a.State == b.State {
		return 0
	}
	if a.State == SvcNormal {
		return -1
	}
	return 1
}

// compareServers 组内服务器排序规则（四级）：
//  1. 服务器状态（normal 优先）
//  2. 按 AS→TS→SS→DS 逐一比较服务状态（normal 优先）
//  3. IP 字典序
//  4. Name 字典序（最终 tie-breaker，保证结果稳定）
//
// 目的：让运维最需关注的服务器（异常 + 服务异常 + IP 靠前）排最前。
func compareServers(a, b Server) int {
	if a.State != b.State {
		if a.State == SrvNormal {
			return -1
		}
		return 1
	}
	for _, t := range allServiceTypes {
		if c := compareServiceStatus(findService(&a, t), findService(&b, t)); c != 0 {
			return c
		}
	}
	if c := strings.Compare(a.IP, b.IP); c != 0 {
		return c
	}
	return strings.Compare(a.Name, b.Name)
}

// ============================================================================
// 组间比较函数
// ============================================================================

// cmpByMaskDesc mask 数值降序。
//
// 语义：mask 越大（服务越复杂、优先级越高），排越前。
// 依赖不变量：asBit > tsBit > ssBit > dsBit（bit 定义即优先级）。
func cmpByMaskDesc(a, b int) int {
	return b - a
}

// cmpByServiceCount 服务数量升序 → mask 降序。
//
// 第一层：bit 1 的个数少的排前（单服务在前）
// 第二层：同数量时 mask 大的排前（高优先级组合在前）
func cmpByServiceCount(a, b int) int {
	if c := bits.OnesCount(uint(a)) - bits.OnesCount(uint(b)); c != 0 {
		return c
	}
	return b - a
}

// cmpByPrimary 主服务索引升序 → 数量升序 → mask 降序。
func cmpByPrimary(a, b int) int {
	if c := primaryIndex(a) - primaryIndex(b); c != 0 {
		return c
	}
	if c := bits.OnesCount(uint(a)) - bits.OnesCount(uint(b)); c != 0 {
		return c
	}
	return b - a
}

// cmpByConvention 业务约定顺序：由 comboOrderList 完全决定。
//
// 兜底规则（未来新增服务类型时的保护）：
//   - 两个都在列表中 → 按列表索引升序
//   - 只有 a 不在列表 → a 排后
//   - 只有 b 不在列表 → b 排后
//   - 都不在列表 → 按 mask 降序（保证确定性）
func cmpByConvention(a, b int) int {
	ra := slices.Index(comboOrderList, a)
	rb := slices.Index(comboOrderList, b)
	switch {
	case ra >= 0 && rb >= 0:
		return ra - rb
	case ra < 0 && rb < 0:
		return b - a
	case ra < 0:
		return 1
	default:
		return -1
	}
}
