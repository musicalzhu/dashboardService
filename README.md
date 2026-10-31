# Dashboard 分组方案 POC：六种实现方案对比分析

> 一份关于可维护性、SOLID 原则与设计模式落地的工程实践记录。
> 适用于任何面向对象或函数式语言，Go 仅作为本次 POC 的实现载体。

---

## 📖 目录

- [引言：从代码可维护性到设计模式](#引言从代码可维护性到设计模式)
- [一、背景与目标](#一背景与目标)
- [二、六种 POC 方案总览](#二六种-poc-方案总览)
- [三、六种 POC 方案详解](#三六种-poc-方案详解)
- [四、三种未实现的方案及理由](#四三种未实现的方案及理由)
- [五、需求变更分析总结](#五需求变更分析总结)
- [六、决策建议与演进路径](#六决策建议与演进路径)
- [七、项目结构](#七项目结构)
- [八、三种验证途径](#八三种验证途径)
- [九、Golden 测试说明](#九golden-测试说明)
- [十、POC 评审的核心问题](#十poc-评审的核心问题)
- [十一、总结](#十一总结)
- [附录 A：术语表](#附录-a术语表)
- [附录 B：变更类型速查](#附录-b变更类型速查)
- [附录 C：验证途径速查](#附录-c验证途径速查)

---

## 引言：从代码可维护性到设计模式

### 代码的“禅意”：可维护性是一种长期主义
写代码容易，改代码难。真正考验工程的，不是第一次实现有多快，而是三个月后、三年后，别人（包括未来的自己）能否在不破坏现有行为的前提下，安全地增加一个功能、修改一处逻辑、删除一个不再需要的规则。

> 代码的“优雅”不是花哨的技巧，而是让变更的成本与变更的范围成正比：
> - 改一个规则，就只改一个地方；
> - 加一个规则，不需要理解全部规则；
> - 删一个规则，不会留下隐蔽的引用；
> - 读一个函数，不需要在多个文件之间反复跳转。

这一原则与语言无关：无论是 Go、Java、Python 还是 TypeScript，可维护性的本质都是控制变更的传播范围。本文讨论的设计原则和模式适用于任何面向对象或函数式语言，Go 只是 POC 的实现载体。

### SOLID：把“可维护性”翻译成可检查的准则
“可维护性”听起来抽象，SOLID 把它拆成了五条可操作的准则：

| 原则 | 原始涵义 | 在分组场景中的含义 |
| :--- | :--- | :--- |
| **S** 单一职责 | 一个类应该只有一个引起它变化的原因 | 规则分发、分组执行、输出序列化，各自独立 |
| **O** 开闭原则 | 软件实体应对扩展开放，对修改关闭 | 新增规则不修改既有分发逻辑 |
| **L** 里氏替换 | 子类型必须能够替换其基类型，而不破坏程序正确性 | 任何规则实现都能替换另一个而不破坏调用方 |
| **I** 接口隔离 | 客户端不应被迫依赖它不使用的方法 | 不强迫调用方依赖它不需要的方法 |
| **D** 依赖倒置 | 高层模块不应依赖低层模块，两者都应依赖抽象 | 高层策略不依赖低层实现细节 |

SOLID 的底层逻辑只有一条：让“不该变的地方”在变更时不需要变，让“该变的地方”只变一处。这正是后文所有方案对比的评判标尺。

### 设计模式：原则在具体场景下的具象化
设计模式是 SOLID 在特定问题下的常见解法。策略模式对应“开闭 + 里氏替换”，Builder 对应“单一职责 + 分步构造”，装饰器对应“开闭 + 组合优于继承”。

但模式的价值完全取决于场景匹配度。一个模式是否适用，不在于它是否符合教科书定义，而在于它是否让当前场景的变更成本更低、代码更可读、测试更容易。如果套用模式反而增加了理解负担、引入了不必要的间接层，那么再“标准”的模式也是过度设计。

> **判断是否过度设计的唯一标准是变更成本，而不是模式名称。** 一个 20 行的 switch 如果让新增规则只需加一个 case，它就是比 200 行的策略模式更好的设计——前提是当前业务确实只有少量固定规则。


### 从原则到实践：Dashboard POC 的由来

UMP Dashboard 的分组需求实现相对独立——这个真实痛点，才是本次 POC 的直接动因。

**生产环境代码现状**

UMP 集群拓扑分组展示已实现 **2 种分组排序视图**：

- **Dashboard 视图（先上线）**：按主服务分组（按最高优先级服务合并成 AS/TS/SS/DS 分组），按页面要求的 flat 格式输出。
- **Dashboard Detail 视图（后续实现）**：按自定义服务组合分组（按 AS/TS/SS/DS 的完整组合和自定义顺序分组），**相对独立地实现了一份逻辑**。

**共享部分**：只有「获取集群拓扑原始信息」。除此之外，两套实现的分组规则、排序逻辑、输出转换基本上是各写各的。

**痛点不止一层：**

**第一层——两套实现都要改。** 每次业务需求变更（如增减服务类型、调整排序规则、调整输出格式、新增分组视图），两套实现都要分别修改。变更集越大，改动成本越高，且容易漏改。历史原因造成的两份相对独立的实现，会持续放大变更成本。**这是“重复”的代价。**

**第二层——每套实现内部，应对变更的变更点也比 `baseline` 多。** 这不是“改两个文件”这么简单——**现有实现的每一套内部，都没有按单一职责原则拆分**。

| 变更操作 | `baseline` 方案 | 现有生产实现 |
|---------|---------------|-------------|
| **调整排序规则** | 改 `compareServers` 1 处 | 每套实现内部，排序逻辑可能与分组逻辑、输出逻辑交织在同一函数中——改排序可能牵动分组、改分组可能影响输出 |
| **新增分组视图** | 加一个 `switch case`（2 行） | 每套实现新增一套完整逻辑（分组 + 排序 + 输出） |
| **调整输出格式** | 改 `flat.go` 中的 `Flatten` 1 处 | 每套实现都有独立的输出转换代码，改动点成倍增加 |
| **增减服务类型** | 改 `allServiceTypes` 1 处 | 两套实现的分组、排序、输出中可能都有硬编码的服务类型判断 |
| **改通用骨架**（如加日志） | 改 `groupBy` 1 处 | 两套实现各自的骨架都要改，且骨架内部可能有多处重复逻辑 |

**这是“违反单一职责 + 开闭原则”的代价**——每套实现都是一个“分组 + 排序 + 输出”的混合体，**任何变更都会在多个关注点之间传播**。

**两层痛点的根源是同一个：**

- **没有统一的规则分发骨架**——规则散落在各套实现的各个角落，新增规则、修改规则都需要“找到所有相关代码”
- **没有独立的输出层**——输出转换与分组逻辑耦合，改输出可能破坏分组
- **没有单一的数据定义源**——服务分组类型、排序优先级等领域模型和规则数据可能在各套实现中各自定义，改一处漏一处

**这不是“某个方案写得不好”，而是“从一开始就没有按 SOLID 原则设计”——历史原因造成的“分散实现”，会持续放大变更成本。**

**POC 的目的**：用可运行的代码对比六种技术方案，找出“当前业务最匹配、未来演进成本最低”的方案——**让下一次需求变更只改一处，且那一处是真正该改的地方**。

UMP Dashboard 的分组需求恰好是一个微缩的设计试验场：

- 为验证技术方案对集群拓扑展示扩展的适应性，POC 中额外引入了三种精确组合的排序变体（服务数量升序、主服务优先、业务约定顺序），以模拟“规则数量增长”的场景。分组骨架高度相似，只有分组键和排序不同。

业务可能的需求变更包括（详见第五章）：
- 增减服务类型（如 4 种 → 3 种，或新增第 5 种）
- 每台服务器内不同服务的排序规则调整
- 同一个分组内不同服务器之间的排序规则调整
- 集群中“无服务的服务器”和“包含未知服务的服务器”显示或分组逻辑调整
- 更多页面出现不同分组逻辑和排序展现要求（如按集群服务器Zone分组、按状态分组）
- API 解析逻辑调整
    - 输出平铺（按照分组的分组名称 + 服务器列表展开平铺）
    - 截断展示（每分组只显示前 N 台服务器 + 加“更多”入口）


本次 POC 是技术学习与方案讨论，目的是在多种可运行实现中对比，用代码事实而非直觉评估各自取舍。

> **POC 的目的不是选出“最模式”的方案，而是选出“最匹配当前业务、且未来演进成本最低”的方案。** 模式只是候选工具，变更成本才是裁判。所有结论都来自可运行的代码和可复现的验证，而非直觉或偏好。

---

## 一、背景与目标

### 业务场景
UMP平台的 Dashboard 需要按“服务组合”将服务器分组展示：
- 每台服务器运行 AS、TS、SS、DS 四种服务中的一种或多种。
- 系统需要从不同视角将服务器分组，供前端展示。
- **业务核心视图**：主服务分组、自定义服务组合分组（共 2 种）。**当前生产环境是两份相对独立实现（详见引言“生产环境代码现状”）。**
- **POC 扩展验证**：为模拟分组排序规则变化，额外引入 3 种精确组合的排序变体（共 5 种分组排序规则）。



### POC 目标
对比 6 种技术实现方案，用可运行的代码事实进行技术方案讨论与学习。

**评估维度：**

| 维度 | 说明 |
| :--- | :--- |
| **变更规则的成本** | 变更一个规则要改几行代码 |
| **修改骨架的成本** | 改 `prepareServer` 等通用逻辑要改几处 |
| **可读性** | 新人理解所需时间 |
| **测试难度** | 规则是否可独立测试 |
| **扩展性** | 未来需求变化的适应能力 |

---

## 二、六种 POC 方案总览

| # | 方案 | 包名 | 核心机制 | 规则载体 |
| :---: | :--- | :--- | :--- | :--- |
| 1 | 基线 | `baseline` | switch 分发 | switch case |
| 2 | 数据驱动 | `datadriven` | map 查表 | map entry |
| 3 | 函数式选项 | `options` | Option 列表构造 | `[]Option` |
| 4 | 装饰器 | `decorator` | `makeGrouper` 冻结函数 | Grouper 函数值 |
| 5 | Builder 模式 | `builder` | 链式配置 + Build | Builder 对象 |
| 6 | 纯策略模式 | `strategy` | 独立类型实现接口 | 接口 + 类型 |

**所有方案共享：**
- `baseline` 的领域模型（`Server`、`Service`、`Cluster`、`Group`、`ClusterGroup`）。
- `baseline.GroupServersByCluster` 分组流程（`strategy` 除外，它不共享骨架）。
- 输出结果完全一致——本 POC 通过三种互补的验证途径保证一致性（详见第八章）。

---

## 三、六种 POC 方案详解

### 方案 1：baseline（基线）
**核心思想**：零抽象——一个 switch 直接选 RuleImpl，交给共享骨架执行。

```go
func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
    switch rule {
    case RuleCombination:
        return GroupServersByCluster(servers, RuleImpl{KeyOf: MaskOf, Cmp: CmpByMask})
    case RuleServiceCount:
        return GroupServersByCluster(servers, RuleImpl{KeyOf: MaskOf, Cmp: CmpByServiceCount})
    // ... 其余 case
    }
    panic("invalid rule")
}
```

| 维度 | 评价 |
| :--- | :--- |
| **新增规则** | 加一个 case（2 行） |
| **修改骨架** | 改 `groupBy` 一处 |
| **可读性** | ★★★★★（读完 switch 就懂全部） |
| **可枚举** | ✅（switch 可见所有规则） |
| **POC 定位** | **基线**——其他方案与它对比“多出的抽象换来了什么” |

**经典适用场景**（不限本 POC）：命令分发（CLI 参数 → 具体函数）、有限状态机的状态转移、协议消息类型分发、操作系统的系统调用号分发——任何“固定数量的枚举 → 明确行为”的场景。

### 方案 2：datadriven（数据驱动）
**核心思想**：规则收敛到一张表——`ruleImpls` map 集中声明所有规则。

```go
var ruleImpls = map[GroupRule]RuleImpl{
    RuleExactCombination:  {KeyOf: MaskOf, Cmp: CmpByMask},
    RuleExactServiceCount: {KeyOf: MaskOf, Cmp: CmpByServiceCount},
    // ... 其余行
}

func RenderDashboard(servers []Server, rule GroupRule) []ClusterGroup {
    r, ok := ruleImpls[rule]
    if !ok { panic(...) }
    return baseline.GroupServersByCluster(servers, r)
}
```

| 维度 | 评价 |
| :--- | :--- |
| **新增规则** | 加一行 map entry |
| **修改骨架** | 改 `groupBy` 一处（共享） |
| **可读性** | ★★★★（表本身就是规则清单） |
| **可枚举** | ✅（遍历 map） |
| **规则可解析** | ✅（map 键天然是字符串来源） |
| **POC 定位** | **候选 A**——最简洁的“有抽象”方案 |

**经典适用场景**：词法分析器关键字表、编码/解码映射表、HTTP 路由表（`http.ServeMux`）、国际化消息表、数据库驱动注册表（`database/sql.Register`）、状态机中的状态转移表——任何“配置驱动”的分发场景。

### 方案 3：options（函数式选项）
**核心思想**：规则由 Option 函数组合构造，而非从预定义表中查找。

```go
func RenderDashboard(servers []Server, opts ...Option) []ClusterGroup {
    return baseline.GroupServersByCluster(servers, apply(opts))
}

// 预定义选项组合
var PrimaryService = []Option{
    WithKeyOf(baseline.PrimaryServiceMask),
    WithCmp(baseline.CmpByMask),
}

// 自由组合（表中没有的规则）
options.RenderDashboard(servers,
    options.WithKeyOf(baseline.PrimaryServiceMask),
    options.WithCmp(baseline.CmpByServiceCount),
)
```

| 维度 | 评价 |
| :--- | :--- |
| **新增规则** | 加一个 `[]Option` 变量 + 1 行 map |
| **修改骨架** | 改 `groupBy` 一处（共享） |
| **可读性** | ★★★（Option 数量多时需追踪） |
| **自由组合** | ✅ 独有（可构造表中没有的组合） |
| **暴露实现细节** | ✅ 调用方要写 `WithKeyOf(baseline.MaskOf)` |
| **POC 定位** | **候选 B**——验证“自由组合”能力 |

**经典适用场景**：gRPC 客户端配置（`grpc.Dial(addr, WithInsecure(), WithBlock())`）、数据库连接配置、日志库配置（`logrus.New().WithField(...)`）、HTTP 服务器配置、指标/追踪客户端配置——任何“可选参数多、默认值合理”的构造函数。

### 方案 4：decorator（装饰器）
**核心思想**：规则是一等函数——`makeGrouper` 冻结 keyOf/cmp 成 Grouper，可被装饰器包装。

```go
// 规则 = 冻结 keyOf/cmp 的闭包
var PrimaryService = makeGrouper(baseline.PrimaryServiceMask, baseline.CmpByMask)

// 直接调用
groups := decorator.PrimaryService(servers)

// 只给这个规则加日志（装饰器独特能力）
primaryWithLog := decorator.WithLogging(logFn)(decorator.PrimaryService)
```

| 维度 | 评价 |
| :--- | :--- |
| **新增规则** | 加一个变量 + 1 行 map（如需 Parse） |
| **修改骨架** | 改 `groupBy` 一处（共享） |
| **规则级独立增强** | ✅ 独有（只给某个规则加缓存/日志/容错） |
| **可枚举** | ⚠️ 需维护 `allGroupers` map |
| **调试难度** | ⚠️ 闭包内断点难以对应规则名 |
| **POC 定位** | **候选 C**——验证“规则级增强”能力 |

**经典适用场景**：HTTP 中间件链（Gin 的 `router.Use(...)`）、IO 流处理（`bufio.NewReader` → `gzip.NewReader` → `io.LimitReader`）、RPC 拦截器（gRPC 的 `UnaryInterceptor`）、AOP 切面（日志、事务、鉴权）、Java 的 `java.io` 流体系、Python 的装饰器语法——任何“在不修改原对象的前提下增加行为”的场景。

### 方案 5：builder（Builder 模式）
**核心思想**：链式配置——`New().WithKeyOf(...).WithCmp(...).Build(servers)`。

```go
// 预定义构造器
func NewPrimaryService() *Builder {
    return New().
        WithKeyOf(baseline.PrimaryServiceMask).
        WithCmp(baseline.CmpByMask)
}

// 调用
builder.NewPrimaryService().Build(servers)

// 自由组合
builder.New().
    WithKeyOf(baseline.PrimaryServiceMask).
    WithCmp(baseline.CmpByServiceCount).
    Build(servers)

// 配置复用（Builder 独有）
b := builder.NewPrimaryService()
r1 := b.Build(servers1)
r2 := b.Build(servers2)
```

| 维度 | 评价 |
| :--- | :--- |
| **新增规则** | 加一个构造器 + 1 行 switch |
| **分步配置** | ✅ 独有（配置分散在不同函数中） |
| **配置复用** | ✅ 独有（一次配置，多次 Build） |
| **状态管理** | ⚠️ Builder 对象有状态，需追踪 |
| **POC 定位** | **候选 D**——验证“配置复用”能力 |

**经典适用场景**：HTTP 请求构建、SQL 查询构建（`sqlbuilder.Select(...).From(...).Where(...)`）、复杂对象构造（`strings.Builder`、`bytes.Buffer`、Protobuf 消息链）、测试数据构建、配置文件构造（`viper.New()`）——任何“可选字段多、分步构造、需要复用配置”的场景。

### 方案 6：strategy（纯策略模式）
**核心思想**：每个规则是独立类型，实现 Group 接口，完整重写分组骨架。

```go
type Strategy interface {
    Name() string
    Group(servers []Server) []ClusterGroup
}

type exactCombination struct{}
func (exactCombination) Group(servers []Server) []ClusterGroup {
    // 30 行完整骨架（与其他策略 90% 重复，仅组间排序不同）
}
```

| 维度 | 评价 |
| :--- | :--- |
| **新增规则** | 写完整类型 + 30 行 Group 方法 |
| **修改骨架** | 改 5 处（每个策略独立） |
| **本包代码量** | ~400 行（datadriven 的 16 倍） |
| **POC 定位** | **反例 1**——展示“骨架重复”的代价 |

**经典适用场景**：支付方式（支付宝/微信/银行卡，鉴权、签名、回调完全不同）、压缩算法（Gzip/Bzip2/Zstd）、排序算法（快排/归并/堆排）、路径规划算法（Dijkstra/A*/BFS）、图像编码格式（PNG/JPEG/WebP）。核心判断标准：每个策略的“执行流程”本身就不同，而非仅仅“参数不同”。
**不适合场景**：规则只是配置不同（如当前分组场景）——这正是本 POC 中 strategy 被作为反例的原因。

---

## 四、三种未实现的方案及理由

### 方案 7：注册表模式（Registry）
**核心思想**：每个规则在 `init()` 中自己注册，而非集中在一张表里。

```go
var registry = make(map[Rule]RuleImpl)

func init() {
    registry[RuleExactCombination] = RuleImpl{MaskOf, CmpByMask}
    // ...
}
```
- **与 datadriven 差异**：表是 `init()` 填充，不是包级变量
- **适用场景**：规则分布在多个文件/模块（如插件系统）
- **不实现理由**：当前 5 个规则在一个文件，集中表比分散注册更直观
- **经典场景**：数据库驱动注册（`database/sql.Register`）、图像格式注册（`image.RegisterFormat`）、编码格式注册、插件系统、DI 容器（`uber/dig`、`google/wire`）。

### 方案 8：泛型（Generics）
**核心思想**：用泛型参数化分组键类型，支持任意实体分组。

```go
func GroupBy[S any, K comparable](items []S, keyOf func(S) K, cmp func(a, b K) int) map[K][]S
```
- **优势**：可分组任意实体（`[]Server`、`[]Alert`、`[]User`）
- **不实现理由**：当前只有 Server 一种实体，泛型是 YAGNI
- **未来适用**：多种实体需要统一分组时
- **经典场景**：容器数据结构（`slices.Sort`、`maps.Keys`）、通用算法库、ORM 的泛型仓储（`Repository[T]`）、API 响应包装（`Response[T]`）、函数式工具（`Map[T, U]`、`Filter[T]`）。
- **当前场景**：分组对象只有 Server，引入泛型只增加复杂度，无实际收益。

### 方案 9：代码生成（go generate）
**核心思想**：用 `go generate` 从配置生成 switch 或 map，让“数据表”在编译期展开。

```go
//go:generate go run gen.go
```
- **优势**：规则清单来自外部配置（YAML/DB）时，避免手写代码
- **不实现理由**：当前 5 个规则手工维护完全可控，代码生成增加构建复杂度
- **未来适用**：规则数量 > 50 或规则来自动态配置
- **经典场景**：Protobuf/Thrift 的 stub 生成、`stringer`（枚举 → `String()`）、`mockgen`（interface → mock）、`wire`（依赖注入）、状态机代码生成、模板渲染。

---

## 五、需求变更分析总结

### 变更全集

| 类别 | 编号 | 变更描述 |
| :--- | :---: | :--- |
| **业务数据** | C1 | 增减服务类型（4→3 或 4→5） |
| | C2 | 每台服务器内不同服务的排序规则调整 |
| | C3 | 同一个分组内不同服务器之间的排序规则调整 |
| | C4 | 业务约定顺序变化（`comboOrderList` 调整） |
| | C5 | 集群中“无服务的服务器”和“包含未知服务的服务器”需要独立分组显示 |
| **规则集** | C6 | 新增第 N 种规则 |
| | C7 | 删除某规则 |
| | C8 | 规则数量显著增长（如 >10） |
| | C9 | 更多页面出现不同分组逻辑和排序展现要求（如按集群服务器Zone分组、按状态分组） |
| **骨架** | C10 | 修改通用骨架（加日志/缓存/去重） |
| | C11 | 分组键类型从 int 改为 string（如按 cluster 分组） |
| **规则级增强** | C12 | 只给某个规则加缓存（规则级独立增强） |
| | C13 | 规则需要独立状态（灰度/AB 测试） |
| | C14 | 动态加载规则（从配置中心） |
| **能力扩展** | C15 | 用户自选 key+cmp 组合 |
| | C16 | API 解析（前端按规则名调用） |
| | C17 | 规则枚举（前端下拉框列出所有规则） |
| | C18 | 输出平铺（按照分组的分组名称 + 服务器列表展开平铺） |
| | C19 | 截断展示（每分组只显示前 N 台服务器 + 加“更多”入口） |



### 变更影响矩阵

| 编号 | 变更 | baseline | datadriven | options | decorator | builder | strategy |
| :---: | :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| C1 | 增减服务类型 | 相同 | 相同 | 相同 | 相同 | 相同 | 相同 |
| C2 | 服务器内服务排序规则 | 相同 | 相同 | 相同 | 相同 | 相同 | 相同 |
| C3 | 同分组内服务器排序规则 | 相同 | 相同 | 相同 | 相同 | 相同 | 相同 |
| C4 | 业务约定顺序变化 | 相同 | 相同 | 相同 | 相同 | 相同 | 相同 |
| C5 | 无服务/未知服务独立分组 | 相同 | 相同 | 相同 | 相同 | 相同 | 相同 |
| C6 | 新增规则 | 2 行 | 2 行 | 2 行 | 2 行 | 2 行 | 30 行 |
| C7 | 删除规则 | ⚠️ 无编译保护 | ⚠️ 测试保护 | ⚠️ 测试保护 | ⚠️ 测试保护 | ⚠️ 测试保护 | ✅ 编译保护 |
| C8 | 规则 >10 | ⚠️ switch 过长 | ✅ 好 | ✅ 好 | ⚠️ 变量膨胀 | ✅ 好 | ❌ 类型爆炸 |
| C9 | 更多页面不同分组逻辑 | 需新 switch | ✅ 新表 | ✅ 新选项 | ✅ 新 Grouper | ✅ 新构造器 | ❌ 新类型族 |
| C10 | 修改骨架 | 1 处 | 1 处 | 1 处 | 1 处 | 1 处 | 5 处 |
| C11 | 键类型变化 | 大改 | 大改 | 大改 | 大改 | 大改 | 大改 |
| C12 | 规则级缓存 | ❌ 污染 | ❌ 污染 | ❌ | ✅ 最干净 | ❌ | ✅ 干净 |
| C13 | 规则级状态 | ❌ | ❌ | ❌ | ✅ | ⚠️ | ✅ |
| C14 | 动态加载规则 | ❌ 无法 | ✅ | ✅ | ✅ | ✅ | ✅ |
| C15 | 自由组合 key+cmp | ❌ | ❌ | ✅ 独有 | ✅ | ✅ | ❌ |
| C16 | API 解析 | 加 switch case | ✅ 表键天然 | ✅ 需额外 map | ✅ 需额外 map | 加 switch case | 加 switch case |
| C17 | 规则枚举 | 遍历 switch（不可）| ✅ 遍历表 | ⚠️ 需维护 | ⚠️ 需维护 | ⚠️ 需维护 | ⚠️ 需维护 |
| C18 | 输出平铺 | 输出层与规则层解耦 | 输出层与规则层解耦 | 输出层与规则层解耦 | 输出层与规则层解耦 | 输出层与规则层解耦 | 输出层与规则层解耦 |
| C19 | 截断展示 | 输出层与规则层解耦 | 输出层与规则层解耦 | 输出层与规则层解耦 | 输出层与规则层解耦 | 输出层与规则层解耦 | 输出层与规则层解耦 |

### 关键结论
- **结论 1：业务数据变更（C1-C5）不区分方案。** 增减服务类型、排序规则变化修改点完全相同——都改 `allServiceTypes`、`serviceBit`、`comboOrderList`、`compareServers`。这些是数据定义，不是分发结构。业务数据变更不能作为选型依据。
- **结论 2：规则集变更（C6-C9）区分方案能力。** `strategy` 新增规则需要 30 行（完整重写骨架），其余方案只需 2 行（配置声明）。`strategy` 在规则集扩展上成本最高，这是排除它的核心依据。
- **结论 3：骨架变更（C10-C11）区分方案共享程度。** `strategy` 修改骨架需要改 5 处，其余方案改 1 处（共享 `groupBy`）。C11 键类型变化所有方案都需要大改，不区分方案——这是泛型方案（未实现）的唯一优势场景。
- **结论 4：规则级增强（C12-C14）暴露“函数”与“配置”的本质差异。** `decorator` 的 `WithCache(PrimaryService)` 一行装饰，其他规则不受影响；其他方案必须在 `RenderDashboard` 里加 if 判断，污染通用分发。`decorator` 的独特价值是“规则级独立增强”。
- **结论 5：能力扩展（C15-C17）区分“组合能力”与“枚举能力”。** C15 只有 `options` / `decorator` / `builder` 支持；C16/C17 `datadriven` 最自然（map 键天然是字符串来源）。`datadriven` 在“API 层能力”上最简洁。
- **结论 6：输出层变更（C18-C19）不区分方案。** 所有方案都输出有序的 `[]Group`，平铺、截断都在输出层完成，与规则分发完全解耦。

### 综合评分

| 方案 | 业务数据 | 规则集 | 骨架 | 规则级增强 | 能力扩展 | 输出层 | 总分 |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| baseline | 5 | 3 | 5 | 1 | 2 | 5 | 21 |
| datadriven | 5 | 5 | 5 | 2 | 4 | 5 | **26 ⭐** |
| options | 5 | 5 | 5 | 2 | 5 | 5 | 27 |
| decorator | 5 | 5 | 5 | 5 | 4 | 5 | 29 |
| builder | 5 | 5 | 5 | 2 | 5 | 5 | 27 |
| strategy | 5 | 2 | 1 | 4 | 2 | 5 | 19 |

*注：评分基于“当前业务是否需要”加权。decorator 总分最高，但“规则级增强”（C12-C13）当前业务不需要——这就是为什么最终推荐 datadriven 而非 decorator。评分数值只是辅助参考，真正的决策依据是“当前业务需要什么”。*

---

## 六、决策建议与演进路径

### 决策树

```text
问题 1：是否需要"规则级独立增强"（只给某个规则加缓存/日志/容错）？
├── 是 → decorator
└── 否 → 问题 2

问题 2：是否需要"用户自由组合 key+cmp"？
├── 是 → options / decorator
└── 否 → 问题 3

问题 3：是否需要"分步配置"或"配置复用"？
├── 是 → builder
└── 否 → 问题 4

问题 4：规则数量是否可能 >10？
├── 是 → datadriven
└── 否 → baseline / datadriven（两者皆可，datadriven 更易扩展）
```

### 当前业务的答案

| 问题 | 答案 |
| :--- | :--- |
| 需要规则级增强？ | ❌ 无此需求 |
| 需要自由组合？ | ❌ 业务方“选一个规则” |
| 需要分步配置/复用？ | ❌ 一次性配置 |
| 规则数量 >10？ | ❌ 稳定在 5 种 |

**➡️ 推荐 `datadriven`**

**推荐理由：**
1. 比 `baseline` 多一行抽象，换来“规则清单即数据表”的可读性和可枚举性。
2. 比 `options`/`decorator`/`builder` 少一层间接，调用方只需 `RenderDashboard(servers, rule)`。
3. 未来可平滑演进——如需自由组合可升级到 `options`，如需规则级增强可升级到 `decorator`。
4. `strategy` 作为反例被排除——90% 代码重复，修改骨架改 5 处。

### 演进路径

```text
baseline  →  datadriven  →  options  →  decorator
  │            │               │             │
零抽象      数据表         自由组合      规则级增强
```
**当前：`datadriven`**
- 若未来需要自由组合：从 `datadriven` 升级到 `options`（加 Option 包装，规则表不变）
- 若未来需要规则级增强：从 `datadriven` 升级到 `decorator`（`ruleImpls` 表项变成 `makeGrouper` 变量）
- 若未来需要配置复用：从 `datadriven` 升级到 `builder`

`datadriven` 是“演进成本最低”的起点——所有后续方案都可以从它平滑迁移，无需重写共享骨架。

---

## 七、项目结构

```text
poc/
├── go.mod
├── README.md
│
├── baseline/                 # 方案一：基线（共享模型 + 骨架 + 规则）
│   ├── model.go
│   ├── rules.go
│   ├── dashboard.go
│   ├── flat.go
│   ├── flat_test.go
│   ├── test_helper.go
│   └── baseline_test.go      
│
├── datadriven/               # 方案二：数据驱动
│   ├── model_aliases.go
│   ├── rules.go
│   ├── dashboard.go
│   └── datadriven_test.go
│
├── options/                  # 方案三：函数式选项
│   ├── model_aliases.go
│   ├── options.go
│   ├── rules.go
│   ├── dashboard.go
│   └── options_test.go
│
├── decorator/                # 方案四：装饰器
│   ├── model_aliases.go
│   ├── grouper.go
│   ├── rules.go
│   ├── decorators.go
│   ├── dashboard.go
│   ├── decorator_test.go          # 外部测试：等价性与装饰器能力
│   └── decorators_internal_test.go # 内部测试：命名装饰器单元测试
│
├── builder/                  # 方案五：Builder
│   ├── model_aliases.go
│   ├── builder.go
│   ├── presets.go
│   ├── dashboard.go
│   └── builder_test.go
│
├── strategy/                 # 方案六：纯策略（反例）
│   ├── model_aliases.go
│   ├── helpers.go
│   ├── strategy.go
│   ├── rules.go
│   ├── dashboard.go
│   └── strategy_test.go
│
├── test/                     # 跨方案一致性测试
│   └── golden_test.go
│
└── cmd/                      # 各方案的演示入口
    ├── baseline-demo/
    │   └── main.go
    ├── datadriven-demo/
    │   └── main.go
    ├── options-demo/
    │   └── main.go
    ├── decorator-demo/
    │   └── main.go
    ├── builder-demo/
    │   └── main.go
    └── strategy-demo/
        └── main.go
```

**组织说明：**
- **每个方案独立成包**：不共享代码（`strategy` 除外，它不共享 `groupBy` 骨架），保证对比的纯粹性。
- **各方案的测试文件与包名一致**：`baseline_test.go`、`datadriven_test.go`、`options_test.go`、`decorator_test.go`、`builder_test.go`、`strategy_test.go`——读者在任何方案里都知道“测试文件在 `包名_test.go`”。
- **`cmd/*-demo/main.go`**：各方案的演示入口，用同一批输入数据（`baseline.AllServiceCombinations()`）分别运行，人工观察分组结果。
- **`test/golden_test.go`**：跨方案一致性测试，自动断言六方案输出与 baseline 相同。
- **未实现的方案**：注册表、泛型、代码生成——在本文档“第四章”有详细说明，POC 中不提供代码实现。

---

## 八、三种验证途径

POC 提供三种互补的验证途径：

| 途径 | 目的 | 特点 | 适合场景 |
| :--- | :--- | :--- | :--- |
| **途径一：运行各方案演示** | 人工观察各方案的分组输出 | 直观、可读 | 讲解与演示 |
| **途径二：运行各方案单元测试** | 验证本方案内部逻辑正确 | 精准、覆盖方案独有特性 | 日常开发 |
| **途径三：运行 golden 测试** | 自动断言六方案输出一致 | 可回归、可 CI | 重构保护、CI 集成 |

**三种途径互为补充：**
- **演示程序**用于“看”——让人理解方案差异
- **单元测试**用于“查”——保证本方案内部正确
- **golden 测试**用于“守”——保证重构不破坏跨方案一致性

**环境要求**：Go 1.21+（使用 `slices`、`cmp` 等标准库）

### 途径一：运行各方案演示
每个方案都有独立的演示入口，用同一批输入数据打印分组结果，便于人工比对：

```bash
# 方案一：基线
go run ./cmd/baseline-demo

# 方案二：数据驱动
go run ./cmd/datadriven-demo

# 方案三：函数式选项
go run ./cmd/options-demo

# 方案四：装饰器
go run ./cmd/decorator-demo

# 方案五：Builder
go run ./cmd/builder-demo

# 方案六：纯策略
go run ./cmd/strategy-demo
```
**观察要点**：五个规则在同一批服务器上的分组结果、分组名称、组内服务器顺序、组间顺序。六个方案的输出应当逐行一致。

### 途径二：运行各方案单元测试
每个方案包内部有自己的单元测试（`xxx_test.go`），验证本方案内部逻辑的正确性：

```bash
go test ./baseline/ -v
go test ./datadriven/ -v
go test ./options/ -v
go test ./decorator/ -v
go test ./builder/ -v
go test ./strategy/ -v
```

**各方案单元测试覆盖的核心内容：**

| 方案 | 测试重点 | 覆盖的独有特性 |
| :--- | :--- | :--- |
| **baseline** | 基础工具、比较器、规则元信息、组内排序、分组主流程 | —（基线，作为其他方案的参考模板） |
| **datadriven** | 规则表完整性（防漏 entry）、与 baseline 一致性 | 表键解析、无效规则 panic |
| **options** | 预定义选项组等价性、nil 选项 panic | 自由组合（WithKeyOf + WithCmp 构造表中没有的规则） |
| **decorator** | 预定义 Grouper 等价性、Parse/Names | 装饰器链（WithLogging / WithMetrics / WithRecover） |
| **builder** | 预定义构造器等价性、Builder 状态 | 配置复用（一次配置，多次 Build）、状态可变 |
| **strategy** | 策略表完整性、策略自包含性 | 每个 Group 方法独立可用 |

### 途径三：运行 golden 测试
`test/golden_test.go` 自动断言六个方案的输出与 baseline 完全一致：

```bash
# 运行全部 golden 测试
go test ./test/ -v

# 只跑跨方案一致性测试
go test -run TestGolden_AllSchemesMatchBaseline ./test/ -v
```

**运行全部测试：**
```bash
# 运行全部单元测试 + golden 测试
go test ./...

# 全量测试 + race 检测 + 覆盖率
go test ./... -race -cover
```

**推荐验证流程：**
```bash
# 日常开发：
# 1. 先跑单方案测试，定位改动是否影响本方案内部
go test ./baseline/ ./datadriven/ ./options/ -v
# 2. 再跑 golden 测试，确认改动没有破坏跨方案一致性
go test ./test/ -v

# 发布前 / PR 合并前：全量测试 + race 检测 + 覆盖率
go test ./... -race -cover
```

---

## 九、Golden 测试说明

### 为什么需要 golden 测试
POC 涉及六个方案的并行实现，所有方案的输出必须与 baseline 完全一致。这是 POC 的核心不变量，也是后续任何重构必须保持的契约。如果只依赖“运行演示程序人工比对”，存在易漏、无回归保护、CI 无法自动化三个问题。golden 测试的价值在于自动保护“六方案输出一致”这一不变量。

### 核心设计：归一化快照
golden 测试的关键不是“直接比 `[]ClusterGroup`”，而是先归一化——把各方案的输出转成统一快照，只保留用户可见的字段，消除内部实现差异：

```go
// groupSnapshot 是跨方案比较的统一快照结构。
type groupSnapshot struct {
    ClusterID   int
    GroupName   string
    ServerNames []string
}

// snapshot 把 baseline.ClusterGroup 切片转成快照。
func snapshot(clusters []baseline.ClusterGroup) []groupSnapshot {
    // ... 提取核心字段，忽略内部 ID 等数据属性
}
```

### 测试失败时的输出示例
```text
=== RUN   TestGolden_AllSchemesMatchBaseline/datadriven/rule-primary-service
    golden_test.go:xx: datadriven 输出与 baseline 不一致
         got:  [{ClusterID:1 GroupName:"AS" ServerNames:[AS AS+SS ...]}]
         want: [{ClusterID:1 GroupName:"AS" ServerNames:[AS AS+DS ...]}]
```
输出会精确指出：哪个方案、哪个规则、哪个集群、哪个分组的服务器列表不一致。

### CI 集成建议

**GitHub Actions (`.github/workflows/ci.yml`)：**
```yaml
name: POC Verification

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run all tests (unit + golden)
        run: go test ./... -race -cover

      - name: Run golden tests verbosely
        run: go test ./test/ -v
```

**重构保护约定**：未来任何 PR 修改 `baseline.groupBy` 或各方案 `RenderDashboard`，golden 测试必须通过。它是合并的必要条件，保证“六方案一致”（正确性由各方案的单元测试保证）。

---

## 十、POC 评审的核心问题

分享时，建议让参会者回答以下问题：

| 问题 | 目的 |
| :--- | :--- |
| “如果给 `prepareServer` 加一行日志，`strategy` 要改几处？” | 展示纯策略模式的重复代价（5 处） |
| “如果业务要求‘只给主服务分组加缓存’，哪个方案最自然？” | 展示 decorator 的独特价值 |
| “如果未来规则数量从 5 个涨到 20 个，哪个方案最合适？” | 展示 datadriven 的扩展优势 |
| “当前业务真的需要‘自由组合 key+cmp’吗？” | 验证 options 的假设是否成立 |
| “服务从 4 种改成 3 种，六个方案修改点相同——这说明了什么？” | 证明“业务数据变更”不能作为选型依据 |
| “如果要实现行转列的输出平铺，需要改哪些方案的核心逻辑？” | 验证“输出层与规则层是否解耦” |

---

## 十一、总结

POC 的价值不是“证明某个方案最好”，而是让团队在 6 个可运行的实现面前，用事实而非直觉做出决策。

**核心结论：**
- **baseline** 是零抽象基线——最简，但缺少可枚举性。
- **datadriven** 是“刚好匹配当前业务”的方案——比 baseline 多一行抽象，换来规则清单即数据。
- **options** 验证“自由组合”能力——当前不需要，但未来可能需要。
- **decorator** 验证“规则级增强”能力——当前不需要，但未来可能需要。
- **builder** 验证“配置复用”能力——当前不需要。
- **strategy** 是反例——90% 代码重复，修改骨架改 5 处，被排除。
- **注册表/泛型/代码生成** 不实现——当前场景无需求。

**最终选择 `datadriven` 的理由**：它既是“刚好匹配”的方案，也是“演进成本最低”的起点。这是 POC 最重要的产出：不是选一个“终极方案”，而是选一个“未来可演进的起点”。

---

## 附录 A：术语表

| 术语 | 说明 |
| :--- | :--- |
| **规则（Rule）** | 一种分组+排序策略，如“精确组合 + mask 降序” |
| **骨架（Skeleton）** | 分组流程的公共实现（分桶 → 组内排序 → 组间排序 → 组装） |
| **RuleImpl** | 规则配置：`KeyOf`（分组键函数）+ `Cmp`（组间比较器） |
| **Grouper** | 已冻结 KeyOf 和 Cmp 的可执行分组函数 |
| **golden 测试** | 用固定输入 + 固定期望输出验证各方案输出一致性的测试 |
| **YAGNI** | You Aren't Gonna Need It，不要为“假想的需求”提前设计 |

## 附录 B：变更类型速查

| 类别 | 变更示例 | 关键判断 |
| :--- | :--- | :--- |
| **业务数据** | 增减服务类型、排序规则变化 | 不区分方案（数据定义变更） |
| **规则集** | 新增/删除规则、规则数量增长 | 区分方案（strategy 成本最高） |
| **骨架** | 修改通用分组流程、键类型变化 | 区分方案（strategy 改 5 处） |
| **规则级增强** | 单规则缓存、AB 测试、动态加载 | 区分方案（decorator 独有） |
| **能力扩展** | 自由组合、API 解析、规则枚举 | 区分方案（options 独有自由组合，datadriven 独有解析） |
| **输出层** | 平铺、截断展示 | 不区分方案（输出层与规则层解耦） |

## 附录 C：验证途径速查

| 验证途径 | 命令 | 验证什么 | 何时运行 |
| :--- | :--- | :--- | :--- |
| **演示程序** | `go run ./cmd/xxx-demo` | 人工观察分组结果 | 讲解、演示 |
| **单元测试** | `go test ./xxx/ -v` | 本方案内部正确性 | 日常开发、修改某方案后 |
| **golden 测试** | `go test ./test/ -v` | 跨方案输出一致性 | 修改共享骨架后、PR 合并前、CI |
| **全量测试** | `go test ./... -race -cover` | 全部内容 | 发布前、CI |

---
*Happy Coding & Happy Designing! 🚀*
