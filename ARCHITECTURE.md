# Agent Runtime - 架构文档

> 从零实现的最小可用 Agent Runtime，不依赖任何 Agent Framework

## 📁 项目结构

```
agent-runtime/
├── cmd/                      # 入口程序（2 个示例）
│   ├── deepseek_demo/        # 基础示例（简洁输出）
│   └── deepseek_demo_v2/     # 增强示例（美化输出）
│
├── pkg/                      # 核心实现（28 个文件）
│   ├── models/               # 数据模型（4 个文件）
│   ├── runtime/              # 核心运行时（10 个文件）⭐
│   ├── tools/                # 工具系统（7 个文件）
│   ├── session/              # Session 管理（4 个文件）
│   └── context/              # Context 管理（3 个文件）
│
├── go.mod
├── go.sum
└── README.md
```

## 🎯 核心模块详解

### 1. pkg/models/ - 数据模型层

定义所有核心数据结构，符合 OpenAI Tool Calling 标准。

| 文件 | 功能 | 行数 |
|------|------|------|
| `message.go` | 消息模型（Message, ToolCall） | 87 |
| `session.go` | Session 模型 | 18 |
| `tool.go` | Tool Schema 模型 | 22 |
| `trace.go` | Trace 记录模型 | 40 |

**关键设计**：
- 完整支持 OpenAI Tool Calling 格式
- 支持 4 种消息角色：system, user, assistant, tool
- 完整的 Trace 类型：loop_start, llm_call, tool_call, loop_end

### 2. pkg/runtime/ - 核心运行时 ⭐

Agent Runtime 的核心实现，包含完整的 Agent Loop。

| 文件 | 功能 | 行数 | 重要性 |
|------|------|------|--------|
| `types.go` | 核心接口定义 | 160 | ⭐⭐⭐⭐⭐ |
| `runtime.go` | Agent Loop 核心逻辑 | 240+ | ⭐⭐⭐⭐⭐ |
| `deepseek_client.go` | LLM 客户端 | 180 | ⭐⭐⭐ |
| `session.go` | Session 管理器 | 120 | ⭐⭐⭐ |
| `tracer.go` | Trace 记录器 | 50 | ⭐⭐ |
| `fake_llm.go` | 测试用 Fake LLM | 100 | ⭐ |
| `fake_tools.go` | 测试用 Fake Tools | 80 | ⭐ |
| `real_tools.go` | 真实工具实现 | 320 | ⭐⭐⭐⭐ |
| `runtime_test.go` | 集成测试（10 个场景） | 400+ | ⭐⭐⭐ |
| `real_tools_test.go` | 真实工具测试 | 60 | ⭐⭐ |

**核心 Agent Loop 流程**（`runtime.go`）：
```go
func (r *Runtime) Execute(ctx context.Context, sessionID, userInput string) (*Response, error) {
    for step < MaxSteps {
        1. 加载 Session Context
        2. 构建 LLM 请求（含 Tool Schemas）
        3. 调用 LLM
        4. 解析响应（Direct Answer or Tool Calls）
        5. 执行 Tools（如果需要）
        6. 更新 Context
        7. 记录 Trace
        8. 检查是否需要压缩 Context
    }
    return FinalAnswer
}
```

**关键设计**：
- ✅ 非递归实现（for 循环）
- ✅ 自动 Context 压缩
- ✅ 完整的错误处理
- ✅ 详细的 Trace 记录

### 3. pkg/tools/ - 工具系统

可插拔的工具系统，支持动态注册。

| 文件 | 功能 | 行数 |
|------|------|------|
| `tool.go` | Tool 接口定义 | 40 |
| `registry.go` | 工具注册中心 | 100 |
| `validator.go` | 参数验证器 | 40 |
| `calculator.go` | 计算器工具（真实实现） | 150 |
| `search.go` | 搜索工具（DuckDuckGo） | 120 |
| `weather.go` | 天气工具（OpenWeatherMap） | 140 |
| `tools_test.go` | 工具单元测试 | 150 |

**Tool 接口**：
```go
type Tool interface {
    Name() string
    Description() string
    ParamSchema() map[string]interface{}
    Execute(ctx context.Context, params map[string]interface{}) (string, error)
}
```

**已实现的工具**：
1. **Calculator** - 真实数学计算（支持 +、-、*、/）
2. **Search** - DuckDuckGo 搜索（免费，无需 API Key）
3. **Weather** - OpenWeatherMap 天气查询（免费层 1000 calls/day）

### 4. pkg/session/ - Session 管理

支持多 Session 隔离，线程安全。

| 文件 | 功能 | 行数 |
|------|------|------|
| `manager.go` | SessionManager 接口 | 30 |
| `memory_store.go` | 内存存储实现 | 80 |
| `context.go` | SessionContext 实现 | 120 |
| `session_test.go` | Session 测试 | 100 |

**Session 隔离**：
- 同一用户可以创建多个 Session
- 不同 Session 的 Context 完全隔离
- 支持 Session 级别的 Context 管理

### 5. pkg/context/ - Context 管理

智能的 Context 压缩和构建。

| 文件 | 功能 | 行数 |
|------|------|------|
| `config.go` | 压缩配置 | 30 |
| `builder.go` | Context Builder | 100 |
| `context_test.go` | Context 测试 | 80 |

**压缩策略**：
- **Sliding Window**：保留最近 N 条消息（默认 10 条）
- **System Message**：始终保留
- **当前 Loop**：当前循环的消息不压缩

## 🔄 数据流向

```
User Input
    ↓
Runtime.Execute()
    ↓
SessionManager.GetContext()
    ↓
ContextBuilder.BuildForLLM()  ← 压缩历史消息
    ↓
LLMClient.Chat()  ← 调用 DeepSeek/OpenAI
    ↓
┌───────────────────┐
│  Response Type?   │
└─────────┬─────────┘
          │
    ┌─────┴─────┐
    ▼           ▼
Direct      Tool Calls
Answer          ↓
    │     ToolRegistry.Execute()
    │           ↓
    │     Update Context
    │           ↓
    └─────┬─────┘
          ▼
    Tracer.Record()
          ↓
    SessionManager.Update()
          ↓
    Return Response
```

## 🧪 测试覆盖

### 集成测试（runtime_test.go）

10 个完整的测试场景：

1. **DirectAnswer** - 直接回答，不调用工具
2. **SingleToolCall** - 单个工具调用
3. **MultipleToolCalls** - 多个工具并行调用
4. **ToolFailure** - 工具执行失败处理
5. **UnknownTool** - 未知工具错误处理
6. **MaxSteps** - 最大步数限制
7. **SessionIsolation** - Session 完全隔离
8. **FollowUpConversation** - 追问对话（上下文理解）
9. **FollowUpWithTool** - 追问并调用工具
10. **ContextCompression** - Context 压缩（TODO）

### 单元测试

- `real_tools_test.go` - 真实工具测试
- `tools_test.go` - 工具系统测试
- `session_test.go` - Session 管理测试
- `context_test.go` - Context 压缩测试

**运行测试**：
```bash
cd pkg/runtime
go test -v
```

## 🎯 核心特性

### 1. 完全自主实现

- ✅ 没有依赖 LangGraph、OpenHands 等 Agent Framework
- ✅ Agent Loop、Session、Context、Tool Calling 全部自行实现
- ✅ 接口驱动设计，高度解耦

### 2. 非递归实现

使用 `for` 循环而非递归：
- 避免栈溢出
- 支持任意长的 Agent Loop
- 更容易理解和调试

### 3. 测试友好

- **Fake LLM**: 测试不依赖网络，使用预定义响应
- **完整测试**: 10+ 个测试场景，覆盖所有核心功能
- **快速执行**: 所有测试在 1 秒内完成

### 4. 易于扩展

- **添加新工具**: 实现 `Tool` 接口，注册到 Registry
- **替换 LLM**: 实现 `LLMClient` 接口
- **更换存储**: 实现 `SessionManager` 接口

## 📊 关键指标

| 指标 | 数值 |
|------|------|
| 总文件数 | 30 |
| 实现文件 | 25 |
| 测试文件 | 5 |
| 总代码行数 | ~3,500 |
| 测试代码行数 | ~800 |
| 核心模块 | 5 |
| 外部依赖 | 2（go-openai, uuid） |

## 🌟 五颗星核心文件（必读）

1. **pkg/runtime/runtime.go** - Agent Loop 核心逻辑
2. **pkg/runtime/types.go** - 核心接口定义
3. **pkg/tools/tool.go** - Tool 接口
4. **pkg/context/builder.go** - Context 压缩逻辑
5. **pkg/runtime/real_tools.go** - 真实工具实现

## 🚀 快速开始

### 1. 设置 API Key

```bash
export DEEPSEEK_API_KEY="your-api-key"
export OPENWEATHER_API_KEY="your-weather-key"  # 可选
```

### 2. 运行示例

```bash
# 基础示例
go run cmd/deepseek_demo/main.go

# 增强版示例（推荐）
go run cmd/deepseek_demo_v2/main.go
```

### 3. 运行测试

```bash
cd pkg/runtime
go test -v
```

## 🔌 依赖管理

### 必须自己实现的核心模块

| 模块 | 原因 |
|------|------|
| **Runtime Loop** | 核心业务逻辑 |
| **Session Manager** | Session 隔离和管理 |
| **Context Manager** | Context 压缩策略 |
| **Tool Registry** | 工具注册和调度 |
| **Trace System** | 追踪逻辑 |

### 可以依赖第三方的模块

| 模块 | 推荐依赖 |
|------|----------|
| **LLM SDK** | `github.com/sashabaranov/go-openai` |
| **UUID** | `github.com/google/uuid` |
| **HTTP Client** | 标准库 `net/http` |

## 💡 设计哲学

1. **接口驱动** - 所有核心组件都是接口，高度解耦
2. **简单优先** - 先实现 MVP，再逐步优化
3. **测试友好** - Fake 实现支持无网络测试
4. **可观测性** - 完整的 Trace 系统
5. **生产可用** - 真实工具实现，支持实际场景

---

**最后更新**: 2026-09-17
