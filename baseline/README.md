# POC 共享核心与 switch 基线

`baseline` 同时是朴素 switch 规则分发方案，也是四种分发 POC 的共享核心。数据模型、
分组算法、掩码计算、比较器、JSON 编码、测试 fixture 和共同行为测试只在本 package
维护。`datadriven`、`options` 与 `closure` 通过公开 API 复用这些实现，只保留各自的规则选择逻辑。

## 文件结构

| 文件 | 职责 |
| --- | --- |
| [`dashboard.go`](./dashboard.go) | `Group`、共享 `GroupBy` 骨架及 switch 基线入口 `RenderDashboard` |
| [`cluster_dashboard.go`](./cluster_dashboard.go) | 按集群聚合的嵌套视图和平铺视图 API |
| [`rules.go`](./rules.go) | 规则常量、名称解析和 `String` |
| [`mask.go`](./mask.go) | 服务优先级、mask、组合名称、主服务投影 |
| [`compare.go`](./compare.go) | Services 预处理、组内服务器排序、组间比较器 |
| [`model_aliases.go`](./model_aliases.go) | 复用共享 `model` 领域类型 |
| [`json.go`](./json.go) | 对应用公开的 `MarshalGroups` 集成边界 |
| [`fixtures.go`](./fixtures.go) | 演示和各 POC 共用的 15 种服务组合 fixture |
| [`dashboard_test.go`](./dashboard_test.go) | 共享规则行为、边界和排序的唯一权威测试 |
| [`cluster_dashboard_test.go`](./cluster_dashboard_test.go) | 集群视图、集群排序、未知/无服务保留及输入不可变性测试 |
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

- `Services` 为空：进入专用的“无服务”组，固定最后。
- 任意服务类型未知：整台服务器进入专用的“未知服务”组；已知与未知混合时也按未知处理，
  固定在已知组之后、无服务组之前。
- 组内服务器排序：服务器状态、各服务存在性/状态、IP、Name。
- 服务器内 Services 排序：先克隆，再按 AS→TS→SS→DS 整理，不修改调用方输入。

## 集群 Dashboard 输出

当前 `RenderDashboard` 按完整服务组合或主服务规则返回全局 `[]Group`。solutionA 还提出了
两种输出需求：先按 `ClusterID` 聚合集群，再按主服务类型分类；以及把分类写入固定字段的平铺
JSON 结构。baseline 通过独立 API 支持这两种形状，不改变原有规则入口：

| API | 输出 | 用途 |
| --- | --- | --- |
| `RenderClusterDashboard(servers)` | `[]ClusterGroup`，每个集群含 `ServiceGroups` | 可扩展的嵌套服务分组 |
| `RenderClusterFlatDashboard(servers)` | `[]ClusterFlatGroup`，每个服务类型对应固定字段 | 适配固定字段的 Dashboard/JSON 消费方 |

这两个 API 均按 `ClusterID` 聚合，并采用 AS→TS→SS→DS 的主服务优先级。组内共享 baseline 的
服务器排序逻辑。无 Services 的服务器归入 `NoServiceType`；含未知服务（包括已知与未知混合）
的服务器归入 `UnknownServiceType`。平铺结构以 `unknown_services` 输出未知服务服务器，避免
solutionA 原平铺实现中未知类型服务器被遗漏的问题。集群按 `primary`→`secondary`→`arbiter`
排序，未识别的集群角色排最后，再按 `ClusterID` 升序；相同 ID 的元数据使用输入中首次出现的值。

```go
nested := baseline.RenderClusterDashboard(servers)
flat := baseline.RenderClusterFlatDashboard(servers)
```

此功能扩展的是结果聚合/展示形状，不是新增一种规则分发技术；原有 baseline、Data-driven、
Options 和 Closure POC 的规则 API 与语义保持不变。

### solutionA 差异取舍

| solutionA 特征 | baseline 当前状态 | 处理 |
| --- | --- | --- |
| 按 ClusterID 聚合集群 | 原规则入口是全局服务组 | 采纳，新增两个集群视图 API |
| 服务器按最高优先级服务归组 | 原有规则支持完整组合和主服务 bit | 采纳为集群视图分类规则 |
| 嵌套与平铺两种输出模型 | 原有 API 仅返回 `[]Group` | 采纳，复用 model 中已有输出类型 |
| 保留无服务服务器 | 原规则已保留 | 保持；集群视图映射到 `NoServiceType` |
| 未知服务在平铺输出 | solutionA 的 flat default 分支可能跳过非空未知服务 | 改进采纳，显式输出 `unknown_services` |
| 异常状态排在正常状态前 | 与 baseline 既有排序契约相反 | 不采纳，保留 baseline 的正常优先行为 |
| 按服务 Role 排序 | baseline 的组内顺序按状态、服务状态/存在性、IP、Name | 暂不采纳；这是不同的领域排序策略，需业务确认后再改变契约 |
| 自定义 `Sorter` 注入 | 现有 POC 比较共享默认规则分发 | 不采纳到基线；避免为单个输出方案扩大排序接口 |

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

`GroupBy` 的普通分组键允许为 0；负键保留给内部未知服务组和无服务组，回调返回负键会明确 panic，
不会与特殊组混淆。键为 0 的普通组显示名为“键 0”。无效的 `GroupRule.String()` 会返回
`GroupRule(n)`，不会因枚举越界而 panic。

## 运行与测试

```powershell
go run ./cmd/baseline-demo
go test ./baseline
go test ./solutionA
go test ./...
```

共同行为测试只在 `baseline/dashboard_test.go` 维护。候选方案 package 仅测试它们独有的规则
映射、选项配置或闭包关联，避免复制分组和比较器测试；修改共享算法时只需在 baseline 更新并验证行为。
