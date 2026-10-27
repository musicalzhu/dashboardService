# POC 共享核心与 switch 基线

`baseline` 同时是朴素 switch 规则分发方案，也是多个分发 POC 的共享核心。领域模型、
分组算法、掩码计算、比较器、JSON 编码和共同行为测试由本 package
维护。`datadriven`、`options`、`closure`、`builder` 和 `solutionA` 通过别名复用 baseline 的领域模型；
各规则分发 POC 通过公开 API 复用共享算法，只保留各自的规则选择逻辑；
`purestrategy` 是刻意复制分组骨架的对照方案，用于评估纯策略模式的维护成本。

## 文件结构

| 文件 | 职责 |
| --- | --- |
| [`dashboard.go`](./dashboard.go) | `Group`、共享 `GroupBy` 骨架及 switch 基线入口 `RenderDashboard` |
| [`model.go`](./model.go) | baseline 自有领域模型及 dashboard 输出结构 |
| [`rules.go`](./rules.go) | 规则常量、名称解析和 `String` |
| [`mask.go`](./mask.go) | 服务优先级、mask、组合名称、主服务投影 |
| [`compare.go`](./compare.go) | Services 预处理、组内服务器排序、组间比较器 |
| [`json.go`](./json.go) | 分组和集群 UI 结构转换、JSON 编码边界 |
| [`dashboard_test.go`](./dashboard_test.go) | 共享规则行为、边界和排序的唯一权威测试 |
| [`cluster_dashboard_test.go`](./cluster_dashboard_test.go) | 集群 JSON 视图、集群排序、未知/无服务保留及输入不可变性测试 |
| [`example_test.go`](./example_test.go) | 可执行的 switch 分组和 JSON 编码示例 |

## 五种分发方案

| 方案 | 包 | 规则选择方式 | 代码位置 |
| --- | --- | --- | --- |
| 朴素 switch | `baseline` | `RenderDashboard` 的 `switch` 直接选择 key/comparator | [`dashboard.go`](./dashboard.go) |
| 数据驱动 | `datadriven` | `ruleImpls` 静态映射表选择 key/comparator | [`../datadriven/rules.go`](../datadriven/rules.go) |
| 函数式选项 | `options` | `NewGrouping(WithKeyFn(...), WithCmpFn(...))` 配置函数对 | [`../options/options.go`](../options/options.go) |
| 闭包组合 | `closure` | 直接调用规则函数变量，如 `RulePrimaryService(servers)` | [`../closure/closure.go`](../closure/closure.go) |
| 纯策略模式 | `purestrategy` | `Grouper` 接口 + `map[GroupRule]Grouper`，每个策略独立实现分组流程 | [`../purestrategy/implementations.go`](../purestrategy/implementations.go) |

`baseline` 对外提供 `GroupBy`、`MaskOf`、`PrimaryServiceMask` 和各组间比较函数，目的是让
候选方案只改变“如何选择规则”，而不是重复实现服务分组和排序算法。`GroupBy` 也集中处理：

- `Services` 为空：进入专用的“无服务”组，固定最后。
- 任意服务类型未知：整台服务器进入专用的“未知服务”组；已知与未知混合时也按未知处理，
  固定在已知组之后、无服务组之前。
- 组内服务器排序：服务器状态（异常优先）、按 AS→TS→SS→DS 排列的服务类型及服务角色、服务列表长度、IP、Name。
- 服务器内 Services 排序：先克隆，再按 AS→TS→SS→DS 整理，不修改调用方输入。

服务器状态优先级为 abnormal→normal→其他值；服务列表按类型顺序比较，同类型再按
Leader→Follower→LSS→SS→DS→其他角色比较。服务状态不参与组内服务器排序；服务类型/角色的共同前缀
比较完成后，以服务列表长度作确定性次级顺序，再按 IP、Name 排序。长度比较使变长服务列表也构成
传递的全序；solutionA 直接在共同前缀后比较 IP，在服务列表长度不同时可能产生非传递比较结果。

## 集群 Dashboard 输出

`RenderDashboard` 按五种规则返回已排序的 `[]Group`。输出结构属于 JSON/UI 边界，而不是另一种
分组规则：`MarshalDashboard` 可将任一规则结果编码为分组数组，或按排序保留组名的平铺条目。
集群输出只是这一转换的一个聚合特例：在集群内应用所选规则，再输出嵌套分组或逐组平铺的行。
平铺结构由 JSON 边界按需生成，不在领域模型中定义某一种规则专属的固定字段结构。

| API | 输出 | 用途 |
| --- | --- | --- |
| `MarshalDashboard(servers, rule, JSONNested)` | `[]Group` | 通用分组 JSON |
| `MarshalDashboard(servers, rule, JSONFlat)` | 分组名及服务器组成的平铺条目数组 | 保留规则排序 |
| `MarshalClusterDashboard(servers, rule, JSONNested)` | `[]ClusterGroup` | 按集群聚合的嵌套分组 |
| `MarshalClusterDashboard(servers, rule, JSONFlat)` | 集群/分组/服务器组成的平铺条目数组 | 按集群和规则排序输出 |
| `RenderClusterDashboard(servers)` | `[]ClusterGroup` | 兼容入口，固定使用主服务规则 |
| `RenderClusterDashboardByRule(servers, rule)` | `[]ClusterGroup` | 按指定规则生成集群分组 |

集群结果按 `primary`→`secondary`→`arbiter` 排序，未识别的集群角色排最后，再按 `ClusterID`
升序；相同 ID 的元数据使用输入中首次出现的值。每个集群内部完整复用指定规则的分组及排序结果，
未知服务和无服务服务器也因此保留。平铺行包含 group 名称，不依赖固定的 AS/TS/SS/DS 字段。

```go
nested, err := baseline.MarshalClusterDashboard(servers, baseline.RuleExactCombination, baseline.JSONNested)
flat, err := baseline.MarshalClusterDashboard(servers, baseline.RulePrimaryService, baseline.JSONFlat)
```

此功能扩展的是结果聚合/展示形状，不是新增一种规则分发技术；原有 baseline、Data-driven、
Options 和 Closure POC 的规则 API 与语义保持不变。

### solutionA 差异取舍

| solutionA 特征 | baseline 当前状态 | 处理 |
| --- | --- | --- |
| 按 ClusterID 聚合集群 | 原规则入口是全局服务组 | 采纳，新增可应用任一规则的集群 JSON 边界，保留主服务兼容入口 |
| 服务器按分组规则归类 | 原有入口支持完整组合和主服务 bit | 集群 JSON API 接受五种规则；兼容视图默认主服务规则 |
| 嵌套与平铺两种输出模型 | 原有 API 仅返回 `[]Group` | 采纳，由 baseline JSON 边界动态转换 |
| 保留无服务服务器 | 原规则已保留 | 保持；兼容集群视图映射到 `NoServiceType` |
| 未知服务在平铺输出 | solutionA 的 flat default 分支可能跳过非空未知服务 | 改进采纳，通用平铺保留未知服务组 |
| 异常状态排在正常状态前 | 与 baseline 原有排序契约相反 | 采纳，按 solutionA 的优先级调整为异常优先 |
| 按服务 Role 排序 | baseline 原先按服务状态/存在性排序 | 采纳，按服务类型后比较角色优先级，并取消服务状态的排序影响 |
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
| `MarshalGroups(groups)` | 将分组结果编码成嵌套 JSON |
| `MarshalFlatGroups(groups)` | 将已排序分组转换为有序平铺条目并编码 |
| `MarshalDashboard(servers, rule, format)` | 按规则分组后编码成嵌套或平铺 JSON |
| `MarshalClusterDashboard(servers, rule, format)` | 按集群聚合并编码所选规则的嵌套或平铺 JSON |

`GroupBy` 的普通分组键允许为 0；负键保留给内部未知服务组和无服务组，回调返回负键会明确 panic，
不会与特殊组混淆。键为 0 的普通组显示名为“键 0”。无效的 `GroupRule.String()` 会返回
`GroupRule(n)`，不会因枚举越界而 panic。

演示用的 15 种完整组合 fixture 仅由 `cmd/baseline-demo/main.go` 构造，不属于 baseline 的领域 API；
各方案测试使用各自的局部输入样例。

## 运行与测试

```powershell
go run ./cmd/baseline-demo
go test ./baseline
go test ./solutionA
go test ./...
```

共享算法的权威行为测试维护在 `baseline/dashboard_test.go`。`purestrategy` 另以 parity test
验证独立实现与 baseline 的结果一致；其实现及测试重复是该反例 POC 的评估对象，不是共享核心。
其他候选方案 package 仅测试它们独有的规则映射、选项配置或闭包关联。
