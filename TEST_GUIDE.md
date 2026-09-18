# Agent Runtime 测试指南

## 测试概览

本项目包含完整的集成测试套件，**不依赖真实 LLM API**，使用 Fake LLM 实现所有测试场景。

## 快速开始

```bash
# 进入测试目录
cd pkg/runtime

# 运行所有测试
go test -v

# 运行特定测试
go test -v -run TestDirectAnswer

# 查看测试覆盖率
go test -v -cover
```

## 测试场景

### ✅ 已实现的测试（10 个）

| # | 测试名称 | 测试场景 | 验证内容 |
|---|---------|---------|---------|
| 1 | `TestDirectAnswer` | 直接回答 | LLM 不调用工具，直接返回答案 |
| 2 | `TestSingleToolCall` | 单个工具调用 | LLM 调用一个工具，并使用结果回答 |
| 3 | `TestMultipleToolCalls` | 多个工具并行调用 | LLM 同时调用多个工具 |
| 4 | `TestToolFailure` | 工具执行失败 | 工具返回错误时的处理 |
| 5 | `TestUnknownTool` | 未知工具错误 | LLM 尝试调用不存在的工具 |
| 6 | `TestMaxSteps` | 最大步数限制 | 达到最大循环次数时自动终止 |
| 7 | `TestSessionIsolation` | Session 隔离 | 不同 Session 的消息完全隔离 |
| 8 | `TestFollowUpConversation` | 追问对话 | 在同一 Session 中追问 |
| 9 | `TestFollowUpWithTool` | 追问并调用工具 | 追问时调用工具并使用历史上下文 |
| 10 | `TestTraceRecording` | Trace 记录 | 验证完整的 Trace 记录 |

### ⏳ 待实现的测试

| # | 测试名称 | 测试场景 |
|---|---------|---------|
| 11 | `TestContextCompression` | Context 压缩（当前为占位测试）|

## 测试架构

### Fake LLM 实现

测试使用 `FakeLLMClient` 模拟 LLM 行为，预定义响应：

```go
type FakeLLMClient struct {
    // 预定义的响应序列
    responses []*ChatResponse
    callCount int
}
```

**优势**：
- ✅ 测试不依赖网络
- ✅ 完全可预测的结果
- ✅ 快速执行（< 1 秒）
- ✅ 可模拟各种边界情况

### Fake Tools 实现

测试使用简单的 Fake Tools：

```go
// FakeCalculatorTool - 始终返回 "42"
// FakeSearchTool - 返回 mock 搜索结果
// FakeWeatherTool - 返回 mock 天气信息
// FakeFailingTool - 始终返回错误
```

## 测试示例

### 示例 1: 直接回答

```go
func TestDirectAnswer(t *testing.T) {
    // 1. 创建 Fake LLM，预定义直接回答
    fakeLLM := &FakeLLMClient{
        responses: []*ChatResponse{
            {
                Content:      "Hello! I'm doing great.",
                FinishReason: "stop",
            },
        },
    }
    
    // 2. 创建 Runtime
    rt := createTestRuntime(fakeLLM)
    
    // 3. 执行测试
    resp, err := rt.Execute(context.Background(), "session1", "Hello")
    
    // 4. 验证
    assert.NoError(t, err)
    assert.Equal(t, "Hello! I'm doing great.", resp.FinalAnswer)
}
```

### 示例 2: 工具调用

```go
func TestSingleToolCall(t *testing.T) {
    // 1. 第一次调用 LLM - 返回工具调用
    // 2. 执行工具 - 返回结果
    // 3. 第二次调用 LLM - 使用工具结果回答
    fakeLLM := &FakeLLMClient{
        responses: []*ChatResponse{
            {
                ToolCalls: []*ToolCall{
                    {ID: "call_1", Name: "calculator", ...},
                },
                FinishReason: "tool_calls",
            },
            {
                Content:      "The result is 42",
                FinishReason: "stop",
            },
        },
    }
    
    rt := createTestRuntime(fakeLLM)
    resp, err := rt.Execute(context.Background(), "session1", "Calculate 21*2")
    
    assert.Equal(t, 2, resp.LoopCount)
    assert.Contains(t, resp.FinalAnswer, "42")
}
```

## 运行真实 LLM 测试

如果需要使用真实 LLM API（如 DeepSeek）进行测试：

```bash
# 设置 API Key
export DEEPSEEK_API_KEY="your-api-key"

# 运行 demo
go run cmd/deepseek_demo/main.go
```

## 测试覆盖率

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# 查看覆盖率统计
go test -cover
```

## 调试技巧

### 1. 查看详细日志

```go
// 在测试中启用调试输出
t.Logf("Current loop count: %d", resp.LoopCount)
t.Logf("Messages: %+v", messages)
```

### 2. 检查 Trace

```go
// 验证 Trace 记录
for _, trace := range resp.Traces {
    t.Logf("Trace: type=%s, step=%d", trace.Type, trace.StepIndex)
}
```

### 3. 验证 Session 状态

```go
// 获取 Session 历史
history, err := rt.GetSessionHistory(sessionID)
for i, msg := range history {
    t.Logf("Message %d: role=%s, content=%s", i, msg.Role, msg.Content)
}
```

## 常见问题

### Q1: 为什么不使用真实 LLM API 测试？

**A**: 使用 Fake LLM 的好处：
- 测试快速且稳定
- 不需要 API Key
- 可以模拟边界情况
- CI/CD 环境友好

### Q2: 如何添加新的测试场景？

**A**: 
1. 在 `runtime_test.go` 中添加新的测试函数
2. 创建 `FakeLLMClient` 并预定义响应
3. 执行 `Runtime.Execute()`
4. 使用 `assert` 验证结果

### Q3: Session ID 重复怎么办？

**A**: 已修复！使用 `time.Now().UnixNano() + rand.Int63()` 确保唯一性。

## 测试最佳实践

1. **每个测试独立**：不依赖其他测试的状态
2. **清晰的命名**：`Test<Scenario>` 格式
3. **完整的验证**：检查返回值、Loop 次数、Trace 等
4. **有意义的断言消息**：失败时容易定位问题
5. **使用 t.Helper()**：辅助函数标记为 helper

## 下一步

- [ ] 实现 `TestContextCompression`
- [ ] 添加性能测试（Benchmark）
- [ ] 添加并发测试
- [ ] 添加压力测试
