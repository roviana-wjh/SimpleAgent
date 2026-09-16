# Agent Runtime - 从零实现的最小可用 Agent

> 一个使用 Go 语言从零实现的 Agent Runtime，**不依赖任何 Agent Framework**（LangGraph、OpenHands 等），支持完整的 Agent Loop、Tool Calling、Session 管理和 Trace 记录。

## 🚀 快速开始

### 1. 设置环境

```bash
# 克隆项目
git clone <your-repo>
cd agent-runtime

# 安装依赖
go mod download
```

### 2. 配置 API Key

```bash
# 设置 DeepSeek API Key（或其他兼容 OpenAI 格式的 LLM）
export DEEPSEEK_API_KEY="your-api-key-here"
```

### 3. 运行示例

```bash
# 运行基础示例
go run cmd/deepseek_demo/main.go

# 运行增强版示例（推荐）
go run cmd/deepseek_demo_v2/main.go
```

### 4. 运行测试

```bash
# 运行所有测试（使用 Fake LLM，无需 API Key）
cd pkg/runtime
go test -v

# 运行特定测试
go test -v -run TestDirectAnswer
```

## 📐 系统设计

### 核心架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Agent Runtime                          │
│                                                             │
│  User Input                                                 │
│      ↓                                                      │
│  Load Session & Context                                     │
│      ↓                                                      │
│  ┌─────────────────────────────────────┐                   │
│  │     Agent Loop (Max N Steps)        │                   │
│  │                                     │                   │
│  │  1. Build LLM Request               │                   │
│  │     - Conversation History          │                   │
│  │     - Tool Schemas                  │                   │
│  │                                     │                   │
│  │  2. Call LLM (DeepSeek/OpenAI)      │                   │
│  │     - Record Trace                  │                   │
│  │                                     │                   │
│  │  3. Parse Response                  │                   │
│  │     ├─ Direct Answer → Return       │                   │
│  │     └─ Tool Calls → Execute         │                   │
│  │                                     │                   │
│  │  4. Execute Tools (if needed)       │                   │
│  │     - Calculator                    │                   │
│  │     - Search                        │                   │
│  │     - Weather                       │                   │
│  │     - Record Trace                  │                   │
│  │                                     │                   │
│  │  5. Add Tool Results to Context     │                   │
│  │                                     │                   │
│  │  6. Loop Continue or Max Steps      │                   │
│  └─────────────────────────────────────┘                   │
│      ↓                                                      │
│  Final Answer                                               │
└─────────────────────────────────────────────────────────────┘
```

### 核心组件

#### 1. Runtime (`pkg/runtime/`)

**Agent Loop 核心逻辑**：
- 非递归实现，使用 `for` 循环
- 最大步数控制（防止无限循环）
- 完整的错误处理和 Trace 记录

**关键接口**：
```go
type Runtime struct {
    config         *Config
    llmClient      LLMClient      // LLM 调用接口
    toolRegistry   ToolRegistry   // 工具注册中心
    sessionManager SessionManager // Session 管理器
    tracer         Tracer         // Trace 记录器
}

func (r *Runtime) Execute(ctx context.Context, sessionID, userInput string) (*Response, error)
```

#### 2. Session 管理 (`pkg/runtime/session.go`)

**Session 隔离**：
- 每个 Session 有独立的 Context
- 支持同一用户创建多个 Session
- 不同 Session 的消息完全隔离

**Context 内容**：
```go
type SessionContext struct {
    Messages    []*Message  // 对话历史（user、assistant、tool）
    LoopCount   int         // 当前循环次数
    Traces      []*Trace    // Trace 记录
}
```

**Context 压缩**：
- 当消息数量超过 `CompressThreshold` 时自动压缩
- 保留所有 `system` 消息（重要指令）
- 保留最近的 N 条消息（N = CompressThreshold / 2）
- 丢弃中间的历史消息，减少 Token 消耗

#### 3. Tool System (`pkg/tools/`)

**Tool 接口**：
```go
type Tool interface {
    Name() string                         // 工具名称
    Description() string                  // 工具描述（供 LLM 理解）
    ParamSchema() map[string]interface{}  // JSON Schema
    Execute(ctx context.Context, params map[string]interface{}) (string, error)
}
```

**已实现的工具**：
- **Calculator**: 数学表达式计算（真实实现）
- **Search**: 搜索功能（Mock）
- **Weather**: 天气查询（Mock）

#### 4. LLM 集成 (`pkg/runtime/deepseek_client.go`)

**支持 OpenAI 兼容的 API**：
- DeepSeek
- OpenAI
- 其他兼容 OpenAI Tool Calling 格式的 LLM

**LLM 接口**：
```go
type LLMClient interface {
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

type ChatResponse struct {
    Content      string      // 直接回复
    ToolCalls    []*ToolCall // 工具调用请求
    FinishReason string      // "stop" | "tool_calls" | "length"
}
```

#### 5. Trace 系统 (`pkg/runtime/tracer.go`)

**记录内容**：
- `loop_start`: Agent Loop 开始
- `llm_call`: LLM 调用（请求和响应）
- `tool_call`: Tool 执行（参数和结果）
- `loop_end`: Agent Loop 结束
- `error`: 错误信息

**用途**：
- 调试和问题排查
- 性能分析
- 审计和合规

## 📁 项目结构

```
agent-runtime/
├── cmd/
│   ├── deepseek_demo/          # 基础示例程序
│   └── deepseek_demo_v2/       # 增强版示例程序
│
├── pkg/
│   ├── models/                 # 数据模型
│   │   ├── message.go          # 消息模型
│   │   ├── session.go          # Session 模型
│   │   ├── tool.go             # Tool 模型
│   │   └── trace.go            # Trace 模型
│   │
│   ├── runtime/                # 核心实现
│   │   ├── types.go            # 核心类型定义
│   │   ├── runtime.go          # Runtime 主逻辑
│   │   ├── session.go          # Session 管理
│   │   ├── tracer.go           # Trace 记录
│   │   ├── deepseek_client.go  # DeepSeek/OpenAI 客户端
│   │   ├── fake_llm.go         # 测试用 Fake LLM
│   │   ├── fake_tools.go       # 测试用 Fake Tools
│   │   └── runtime_test.go     # 集成测试
│   │
│   ├── tools/                  # 工具实现
│   │   ├── tool.go             # Tool 接口
│   │   ├── registry.go         # 工具注册中心
│   │   ├── calculator.go       # 计算器
│   │   ├── search.go           # 搜索（Mock）
│   │   ├── weather.go          # 天气（Mock）
│   │   └── tools_test.go       # 工具测试
│   │
│   ├── session/                # Session 管理（可选，更复杂的场景）
│   └── context/                # Context 管理（可选，更复杂的场景）
│
├── go.mod
├── go.sum
└── README.md
```

## 🧪 测试场景

项目包含 10 个完整的集成测试场景：

| # | 测试场景 | 验证内容 |
|---|---------|---------|
| 1 | DirectAnswer | 直接回答，不调用工具 |
| 2 | SingleToolCall | 单个工具调用 |
| 3 | MultipleToolCalls | 多个工具并行调用 |
| 4 | ToolFailure | 工具执行失败处理 |
| 5 | UnknownTool | 未知工具错误处理 |
| 6 | MaxSteps | 最大步数限制 |
| 7 | SessionIsolation | Session 完全隔离 |
| 8 | FollowUpConversation | 追问对话（上下文理解）|
| 9 | FollowUpWithTool | 追问并调用工具 |
| 10 | ContextCompression | Context 压缩（TODO）|

```bash
# 运行所有测试
cd pkg/runtime
go test -v

# 查看测试覆盖率
go test -v -cover
```

## 🎯 核心特性

### 1. 完全自主实现

- ✅ 没有依赖 LangGraph、OpenHands、OpenClaw、PI 等 Agent Framework
- ✅ Agent Loop、Session、Context、Tool Calling 全部自行实现
- ✅ 接口驱动设计，高度解耦

### 2. 非递归实现

使用 `for` 循环而非递归：
- 避免栈溢出
- 支持任意长的 Agent Loop
- 更容易理解和调试

### 3. 测试友好

- **Fake LLM**: 测试不依赖网络，使用预定义响应
- **完整测试**: 25 个测试场景，覆盖所有核心功能
- **快速执行**: 所有测试在 1 秒内完成

### 4. 易于扩展

- **添加新工具**: 实现 `Tool` 接口，注册到 Registry
- **替换 LLM**: 实现 `LLMClient` 接口
- **更换存储**: 实现 `SessionManager` 接口

## 💡 使用示例

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "agent-runtime/pkg/runtime"
)

func main() {
    // 1. 创建 LLM 客户端
    llmClient := runtime.NewDeepSeekClient("your-api-key")
    
    // 2. 创建工具注册表并注册工具
    toolRegistry := runtime.NewSimpleToolRegistry()
    toolRegistry.Register(runtime.CalculatorTool())
    toolRegistry.Register(runtime.SearchTool())
    
    // 3. 创建 Session 管理器
    sessionMgr := runtime.NewMemorySessionManager()
    
    // 4. 创建 Tracer
    tracer := runtime.NewMemoryTracer()
    
    // 5. 配置并创建 Runtime
    config := &runtime.Config{
        MaxSteps:    10,
        MaxTokens:   2000,
        Temperature: 0.7,
    }
    
    rt := runtime.NewRuntime(config, llmClient, toolRegistry, sessionMgr, tracer)
    
    // 6. 创建 Session
    session, _ := sessionMgr.Create(context.Background(), "user1")
    
    // 7. 执行 Agent
    resp, err := rt.Execute(context.Background(), session.ID, "请帮我计算 123 + 456")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Answer: %s\n", resp.FinalAnswer)
    fmt.Printf("Loops: %d\n", resp.LoopCount)
}
```

### 追问对话

```go
// 第一次对话
resp1, _ := rt.Execute(ctx, sessionID, "请帮我计算 123 + 456")
// Answer: 579

// 追问（基于历史上下文）
resp2, _ := rt.Execute(ctx, sessionID, "那结果再乘以 2 是多少？")
// Answer: 1158
```

## 🔧 配置

### Runtime 配置

```go
type Config struct {
    MaxSteps          int     // 最大循环次数（默认 10）
    MaxTokens         int     // LLM 最大 Token（默认 2000）
    Temperature       float64 // LLM 温度（默认 0.7）
    CompressThreshold int     // Context 压缩阈值（默认 20）
}
```

### 环境变量

```bash
# DeepSeek API Key
export DEEPSEEK_API_KEY="sk-your-key"

# 或者使用其他兼容 OpenAI 的 LLM
export OPENAI_API_KEY="sk-your-key"
export OPENAI_BASE_URL="https://api.openai.com/v1"
```

## 🐛 故障排查

### 问题 1: API 调用失败

**症状**: `context deadline exceeded` 或 `connection refused`

**解决方案**:
1. 检查 API Key 是否正确
2. 检查网络连接
3. 确认 API 端点是否正确

### 问题 2: Tool 调用失败

**症状**: `tool not found` 或 `tool execution failed`

**解决方案**:
1. 确认工具已注册到 Registry
2. 检查工具参数是否符合 Schema
3. 查看 Trace 记录定位问题

### 问题 3: 超过最大步数

**症状**: `max steps exceeded`

**解决方案**:
1. 增加 `Config.MaxSteps`
2. 检查 LLM 是否陷入循环
3. 优化工具描述，帮助 LLM 更快决策

## 📝 开发记录

详见 [DEVELOPMENT.md](./DEVELOPMENT.md)

## 🚧 TODO

- [ ] 实现 Context Compression（总结历史对话）
- [ ] 添加更多实用工具
- [ ] 集成真实搜索 API
- [ ] 添加 Web API 接口
- [ ] 持久化存储（数据库）
- [ ] Streaming 响应支持

