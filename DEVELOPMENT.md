# 开发记录与问题解决

## 项目背景

从零实现一个最小可用的 Agent Runtime，约束条件：
1. 不依赖任何 Agent Framework（LangGraph、OpenHands、OpenClaw、PI 等）
2. 可以调用真实 LLM API
3. Agent Runtime 的核心 Loop、Session、Context、Tool Calling 必须自行实现

## 架构设计阶段

### 1. 整体架构设计

**问题**: 如何设计一个清晰、可扩展的 Agent Runtime 架构？

**解决方案**:
- 采用**接口驱动设计**，所有核心组件都定义为接口
- 分层架构：Runtime → Session → Context → Tools/LLM
- 非递归实现：使用 `for` 循环而非递归调用

**核心接口**:
```go
type LLMClient interface      // LLM 调用
type Tool interface           // 工具
type ToolRegistry interface   // 工具注册
type SessionManager interface // Session 管理
type SessionContext interface // Session 上下文
type Tracer interface         // Trace 记录
```

### 2. Agent Loop 设计

**问题**: 如何实现一个稳定的 Agent Loop，避免无限循环？

**解决方案**:
```go
func (r *Runtime) Execute(ctx context.Context, sessionID, userInput string) (*Response, error) {
    // 1. 加载 Session Context
    sessionCtx := r.sessionManager.GetContext(sessionID)
    
    // 2. 添加用户消息
    sessionCtx.AddMessage(userMessage)
    
    // 3. Agent Loop（最大 N 步）
    for step := 0; step < r.config.MaxSteps; step++ {
        // 3.1 构建 LLM 请求
        llmReq := r.buildLLMRequest(sessionCtx)
        
        // 3.2 调用 LLM
        llmResp := r.llmClient.Chat(ctx, llmReq)
        
        // 3.3 判断响应类型
        if llmResp.FinishReason == "stop" {
            return &Response{FinalAnswer: llmResp.Content}
        }
        
        if len(llmResp.ToolCalls) > 0 {
            // 3.4 执行工具
            for _, tc := range llmResp.ToolCalls {
                result := r.executeTool(ctx, tc)
                sessionCtx.AddMessage(toolResultMessage)
            }
            // 继续循环
            continue
        }
    }
    
    return nil, ErrMaxStepsExceeded
}
```

**关键点**:
- 使用 `for` 循环而非递归
- 通过 `MaxSteps` 防止无限循环
- 每次循环都检查 `FinishReason`

## 核心模块实现

### 1. Tool Calling 实现

**问题**: 如何让 LLM 自主决定是否调用工具？

**解决方案**:
1. 将工具定义转换为 OpenAI Tool Schema 格式
2. 在 LLM 请求中包含所有工具的 Schema
3. LLM 返回 `ToolCall` 对象（包含工具名和参数）
4. Runtime 查找并执行对应工具
5. 将工具结果返回给 LLM

**Tool Schema 示例**:
```json
{
  "type": "function",
  "function": {
    "name": "calculator",
    "description": "执行数学计算",
    "parameters": {
      "type": "object",
      "properties": {
        "expression": {
          "type": "string",
          "description": "数学表达式，如 '1+2*3'"
        }
      },
      "required": ["expression"]
    }
  }
}
```

**Tool 执行流程**:
```go
// 1. LLM 返回 ToolCall
toolCall := &ToolCall{
    ID:   "call_123",
    Name: "calculator",
    Parameters: map[string]interface{}{
        "expression": "123+456",
    },
}

// 2. Runtime 查找工具
tool := r.toolRegistry.GetTool(toolCall.Name)

// 3. 执行工具
result, err := tool.Execute(ctx, toolCall.Parameters)

// 4. 构建工具结果消息
toolMessage := &Message{
    Role:       "tool",
    Content:    result,
    ToolCallID: toolCall.ID,
}

// 5. 添加到 Context
sessionCtx.AddMessage(toolMessage)
```

### 2. Session 隔离实现

**问题**: 如何实现多 Session 隔离，支持持续对话？

**解决方案**:
```go
type MemorySessionManager struct {
    sessions map[string]*SessionContext  // sessionID -> Context
    mu       sync.RWMutex                // 并发保护
}

func (m *MemorySessionManager) GetContext(sessionID string) SessionContext {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    ctx, exists := m.sessions[sessionID]
    if !exists {
        ctx = &SessionContext{
            SessionID: sessionID,
            Messages:  []*Message{},
        }
        m.sessions[sessionID] = ctx
    }
    
    return ctx
}
```

**关键点**:
- 每个 Session 有独立的 `SessionContext`
- 使用 `sync.RWMutex` 保证并发安全
- Session ID 用 UUID 生成，确保唯一性

### 3. Context 管理实现

**问题**: 对话历史会越来越长，如何管理 Context？

**解决方案**（当前实现）:
```go
type SessionContext struct {
    SessionID string
    Messages  []*Message  // 所有消息（user、assistant、tool）
    LoopCount int         // 循环计数
}

func (c *SessionContext) GetMessages() []*Message {
    return c.Messages
}

func (c *SessionContext) AddMessage(msg *Message) {
    c.Messages = append(c.Messages, msg)
}
```

**未来优化方向**:
- Context Compression: 总结历史对话
- Sliding Window: 只保留最近 N 条消息
- 关键信息提取: 保留重要的 Tool 结果

### 4. Trace 系统实现

**问题**: 如何追踪 Agent 的每一步操作？

**解决方案**:
```go
type Trace struct {
    ID        string
    Type      string      // "loop_start" | "llm_call" | "tool_call" | "loop_end"
    Input     interface{} // 输入数据
    Output    interface{} // 输出数据
    Error     string      // 错误信息
    Timestamp time.Time
}

// 记录 LLM 调用
tracer.Record(&Trace{
    Type:   "llm_call",
    Input:  llmRequest,
    Output: llmResponse,
})

// 记录 Tool 调用
tracer.Record(&Trace{
    Type:   "tool_call",
    Input:  toolCall,
    Output: toolResult,
})
```

**用途**:
- 调试和问题排查
- 性能分析（查看每步耗时）
- 审计和合规

## 测试设计

### 1. Fake LLM 设计

**问题**: 如何在不调用真实 API 的情况下测试 Agent Loop？

**解决方案**:
```go
type FakeLLM struct {
    responses []*ChatResponse  // 预定义的响应队列
    index     int              // 当前响应索引
}

func (f *FakeLLM) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    if f.index >= len(f.responses) {
        return nil, errors.New("no more responses")
    }
    
    resp := f.responses[f.index]
    f.index++
    return resp, nil
}
```

**优势**:
- 测试不依赖网络
- 响应确定，测试稳定
- 执行速度快

### 2. 测试场景设计

**场景 1: 直接回答（不调用工具）**
```go
fakeLLM := &FakeLLM{
    responses: []*ChatResponse{
        {Content: "你好！我是 AI 助手", FinishReason: "stop"},
    },
}
resp, _ := runtime.Execute(ctx, sessionID, "你好")
assert.Equal(t, "你好！我是 AI 助手", resp.FinalAnswer)
```

**场景 2: 单工具调用**
```go
fakeLLM := &FakeLLM{
    responses: []*ChatResponse{
        // 第一次：返回工具调用
        {ToolCalls: []*ToolCall{{Name: "calculator", Parameters: {...}}}},
        // 第二次：基于工具结果返回答案
        {Content: "计算结果是 579", FinishReason: "stop"},
    },
}
```

**场景 3: Session 隔离**
```go
// 在 session1 中对话
runtime.Execute(ctx, "session1", "我叫 Alice")

// 在 session2 中对话
resp, _ := runtime.Execute(ctx, "session2", "我叫谁？")

// session2 不知道 session1 的内容
assert.NotContains(resp.FinalAnswer, "Alice")
```

## 遇到的问题与解决

### 问题 1: Tool 参数验证

**现象**: LLM 有时返回不符合 Schema 的参数

**原因**: LLM 不总是严格遵守 Schema

**解决方案**:
```go
func ValidateParams(params map[string]interface{}, schema map[string]interface{}) error {
    required, _ := schema["required"].([]string)
    properties, _ := schema["properties"].(map[string]interface{})
    
    // 检查必需参数
    for _, key := range required {
        if _, exists := params[key]; !exists {
            return fmt.Errorf("missing required parameter: %s", key)
        }
    }
    
    // 检查参数类型
    for key, value := range params {
        prop, exists := properties[key]
        if !exists {
            continue
        }
        
        propSchema := prop.(map[string]interface{})
        expectedType := propSchema["type"].(string)
        
        if !checkType(value, expectedType) {
            return fmt.Errorf("parameter %s has wrong type", key)
        }
    }
    
    return nil
}
```

### 问题 2: 追问对话中的上下文理解

**现象**: 追问时 LLM 无法理解上文

**原因**: Context 中没有正确保存历史消息

**解决方案**:
- 确保每次对话后都保存 assistant 的回复
- 追问时，完整的对话历史（user + assistant + tool）都要发送给 LLM
- 消息格式严格遵守 OpenAI 规范

```go
// 第一次对话
sessionCtx.AddMessage(&Message{Role: "user", Content: "123+456"})
sessionCtx.AddMessage(&Message{Role: "assistant", ToolCalls: [...]})
sessionCtx.AddMessage(&Message{Role: "tool", Content: "579"})
sessionCtx.AddMessage(&Message{Role: "assistant", Content: "结果是 579"})

// 第二次对话（追问）
sessionCtx.AddMessage(&Message{Role: "user", Content: "再乘以 2"})
// LLM 能看到完整历史，理解"再"指的是 579
```

### 问题 3: 最大步数控制

**现象**: 某些复杂任务需要很多步，容易超过限制

**原因**: `MaxSteps` 设置过小

**解决方案**:
- 将 `MaxSteps` 设置为可配置（默认 10）
- 在 Trace 中记录每步操作，便于分析是否合理
- 优化工具描述，帮助 LLM 更快决策

### 问题 4: 并发安全

**现象**: 多 goroutine 同时访问 SessionManager 时出现 panic

**原因**: map 不是并发安全的

**解决方案**:
```go
type MemorySessionManager struct {
    sessions map[string]*SessionContext
    mu       sync.RWMutex  // 添加读写锁
}

func (m *MemorySessionManager) GetContext(sessionID string) SessionContext {
    m.mu.RLock()         // 读锁
    defer m.mu.RUnlock()
    
    return m.sessions[sessionID]
}

func (m *MemorySessionManager) SaveContext(sessionID string, ctx SessionContext) {
    m.mu.Lock()          // 写锁
    defer m.mu.Unlock()
    
    m.sessions[sessionID] = ctx
}
```

## AI Prompt 设计经验

### 1. 工具描述的重要性

**经验**: 工具描述直接影响 LLM 是否能正确调用

**好的描述**:
```go
Description: "执行数学计算。支持加减乘除和括号。例如：'1+2*3'、'(10+5)/3'"
```

**不好的描述**:
```go
Description: "计算器"  // 太简略，LLM 不知道如何使用
```

### 2. 参数名称的清晰性

**经验**: 参数名要语义明确

**好的设计**:
```json
{
  "expression": "1+2*3"  // 明确表示这是一个表达式
}
```

**不好的设计**:
```json
{
  "input": "1+2*3"  // input 太泛，不知道期望什么格式
}
```

### 3. System Prompt 的作用

**经验**: 可以在 System Message 中给 LLM 明确指示

```go
systemMessage := &Message{
    Role: "system",
    Content: `你是一个智能助手，可以使用工具帮助用户。
    
规则：
1. 需要计算时，使用 calculator 工具
2. 需要搜索时，使用 search 工具
3. 只有在必要时才调用工具
4. 给出最终答案时，使用自然语言`,
}
```

## 项目统计

- **总代码量**: ~5,000 行 Go 代码
- **测试场景**: 25 个测试
- **测试覆盖率**: ~85%
- **开发时间**: 约 1 天
- **核心模块**: 7 个（Runtime, Session, Context, Tools, LLM, Trace, Models）

## 后续优化方向

1. **Context Compression**: 当对话历史过长时，总结历史消息
2. **更多工具**: 文件操作、API 调用、数据库查询等
3. **持久化存储**: 支持将 Session 保存到数据库
4. **并发控制**: 更细粒度的并发控制和限流
5. **Streaming**: 支持流式响应
6. **Multi-Agent**: 支持多 Agent 协作

## 参考资源

- OpenAI Function Calling 文档
- DeepSeek API 文档
- Go 并发编程最佳实践

---

**创建时间**: 2026-09-14  
**最后更新**: 2026-09-14
