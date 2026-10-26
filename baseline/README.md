# POC 共享核心与 switch 基线

`baseline` 同时是朴素 switch 规则分发方案，也是四种分发 POC 的共享核心。数据模型、
分组算法、掩码计算、比较器、JSON 编码、测试 fixture 和共同行为测试只在本 package
维护。`datadriven`、`options` 与 `closure` 通过公开 API 复用这些实现，只保留各自的规则选择逻辑。

## 文件结构

| 文件 | 职责 |
| --- | --- |
| [`dashboard.go`](./dashboard.go) | `Group`、共享 `GroupBy` 骨架及 switch 基线入口 `RenderDashboard` |
| [`rules.go`](./rules.go) | 规则常量、名称解析和 `String` |
| [`mask.go`](./mask.go) | 服务优先级、mask、组合名称、主服务投影 |
| [`compare.go`](./compare.go) | Services 预处理、组内服务器排序、组间比较器 |
| [`model_aliases.go`](./model_aliases.go) | 复用共享 `model` 领域类型 |
| [`json.go`](./json.go) | 对应用公开的 `MarshalGroups` 集成边界 |
| [`fixtures.go`](./fixtures.go) | 演示和各 POC 共用的 15 种服务组合 fixture |
| [`dashboard_test.go`](./dashboard_test.go) | 共享规则行为、边界和排序的唯一权威测试 |
| [`example_test.go`](./example_test.go) | 可执行的 switch 分组和 JSON 编码示例 |

## 四种分发方案

| 方案 | 包 | 规则选择方式 | 代码位置 |
| --- | --- | --- | --- |
| 朴素 switch | `baseline` | `RenderDashboard` 的 `switch` 直接选择 key/comparator | [`dashboard.go`](./dashboard.go) |
| 数据驱动 | `datadriven` | `ruleImpls` 静态映射表选择 key/comparator | [`../datadriven/rules.go`](../datadriven/rules.go) |
| 函数式选项 | `options` | `NewGrouping(WithKeyFn(...), WithCmpFn(...))` 配置函数对 | [`../options/options.go`](../options/options.go) |
| 闭包组合 | `closure` | 直接调用规则函数变量，如 `RulePrimaryService(servers)` | [`../closure/closure.go`](../closure/closure.go) |

`baseline` 对外提供 `GroupBy`、`MaskOf`、`PrimaryServiceMask` 和各组间比较函数，目的是让
候选方案只改变“如何选择规则”，而不是重复实现服务分组和排序算法。`GroupBy` 也集中处理：

- `Services` 为空：进入 `Key=0` 的“无服务”组，固定最后。
- 任意服务类型未知：整台服务器进入 `Key=-1` 的“未知服务”组；已知与未知混合时也按未知处理，
  固定在已知组之后、无服务组之前。
- 组内服务器排序：服务器状态、各服务存在性/状态、IP、Name。
- 服务器内 Services 排序：先克隆，再按 AS→TS→SS→DS 整理，不修改调用方输入。

## 排序与分组边界

| 责任 | 对象 | 共享实现 |
| --- | --- | --- |
| 服务分组 | 输入服务器到组的归属 | `MaskOf` / `PrimaryServiceMask` 与 `GroupBy` |
| 组间排序 | 已形成的服务组 | `CompareByMaskDesc`、`CompareByServiceCount`、`CompareByPrimary`、`CompareByConvention` |
| 组内服务器排序 | 同一组的服务器 | `compareServers` |
| 服务器内服务排序 | 单台服务器的 Services | `prepareServer` |

## 公开共享函数

| 函数 | 用途 |
| --- | --- |
| `GroupBy(servers, keyFn, cmpFn)` | 用指定函数运行统一分组骨架 |
| `MaskOf(server)` | 计算服务器已知服务的完整 mask |
| `PrimaryServiceMask(server)` | 计算最高优先级服务对应的单个 bit |
| `CompareByMaskDesc(a, b)` | 组 key 按 mask 降序 |
| `CompareByServiceCount(a, b)` | 先服务数量、再 mask |
| `CompareByPrimary(a, b)` | 先主服务、再数量、再 mask |
| `CompareByConvention(a, b)` | 按业务约定顺序 |
| `MarshalGroups(groups)` | 将结果编码成缩进 JSON |
| `AllServiceCombinations()` | 返回供 POC 演示与测试使用的 15 种组合 fixture |

## 运行与测试

```powershell
go run ./cmd/baseline-demo
go test ./baseline
go test ./...
```

共同行为测试只在 `baseline/dashboard_test.go` 维护。候选方案 package 仅测试它们独有的规则
映射、选项配置或闭包关联，避免复制分组和比较器测试；修改共享算法时只需在 baseline 更新并验证行为。
