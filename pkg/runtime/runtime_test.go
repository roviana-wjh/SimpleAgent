package runtime

import (
	"context"
	"strings"
	"testing"
)

// ============ 测试场景 1: DirectAnswer ============
func TestDirectAnswer(t *testing.T) {
	// 设置 Fake LLM：直接回答
	llm := NewFakeLLMClient([]*ChatResponse{
		DirectAnswer("Hello! How can I help you?"),
	})

	// 设置工具
	registry := NewSimpleToolRegistry()

	// 设置 Session Manager
	sessionMgr := NewMemorySessionManager()
	session, _ := sessionMgr.Create(context.Background(), "user1")

	// 设置 Tracer
	tracer := NewMemoryTracer()

	// 创建 Runtime
	config := &Config{
		MaxSteps:    10,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	// 执行
	resp, err := rt.Execute(context.Background(), session.ID, "Hello")

	// 验证
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Finished {
		t.Errorf("Expected Finished=true, got false")
	}

	if resp.LoopCount != 1 {
		t.Errorf("Expected LoopCount=1, got %d", resp.LoopCount)
	}

	if resp.FinalAnswer != "Hello! How can I help you?" {
		t.Errorf("Expected FinalAnswer='Hello! How can I help you?', got '%s'", resp.FinalAnswer)
	}

	if llm.GetCallCount() != 1 {
		t.Errorf("Expected 1 LLM call, got %d", llm.GetCallCount())
	}

	t.Logf("✓ DirectAnswer test passed")
}

// ============ 测试场景 2: SingleToolCall ============
func TestSingleToolCall(t *testing.T) {
	// 设置 Fake LLM：第一次调用工具，第二次返回答案
	llm := NewFakeLLMClient([]*ChatResponse{
		ToolCallResponse("calculator", map[string]interface{}{"expression": "2+2"}),
		DirectAnswer("The answer is 4"),
	})

	// 设置工具
	registry := NewSimpleToolRegistry()
	calcTool := CalculatorTool()
	registry.Register(calcTool)

	// 设置 Session Manager
	sessionMgr := NewMemorySessionManager()
	session, _ := sessionMgr.Create(context.Background(), "user1")

	// 设置 Tracer
	tracer := NewMemoryTracer()

	// 创建 Runtime
	config := &Config{
		MaxSteps:    10,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	// 执行
	resp, err := rt.Execute(context.Background(), session.ID, "Calculate 2+2")

	// 验证
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Finished {
		t.Errorf("Expected Finished=true, got false")
	}

	if resp.LoopCount != 2 {
		t.Errorf("Expected LoopCount=2, got %d", resp.LoopCount)
	}

	if resp.FinalAnswer != "The answer is 4" {
		t.Errorf("Expected FinalAnswer='The answer is 4', got '%s'", resp.FinalAnswer)
	}

	if llm.GetCallCount() != 2 {
		t.Errorf("Expected 2 LLM calls, got %d", llm.GetCallCount())
	}

	if calcTool.GetCallCount() != 1 {
		t.Errorf("Expected 1 calculator call, got %d", calcTool.GetCallCount())
	}

	t.Logf("✓ SingleToolCall test passed")
}

// ============ 测试场景 3: MultipleToolCalls ============
func TestMultipleToolCalls(t *testing.T) {
	// 设置 Fake LLM：一次调用多个工具
	llm := NewFakeLLMClient([]*ChatResponse{
		MultipleToolCallResponse([]ToolCall{
			{ID: "call_1", Name: "calculator", Params: map[string]interface{}{"expression": "2+2"}},
			{ID: "call_2", Name: "weather", Params: map[string]interface{}{"city": "Beijing"}},
		}),
		DirectAnswer("The calculation result is 4 and the weather is sunny"),
	})

	// 设置工具
	registry := NewSimpleToolRegistry()
	calcTool := CalculatorTool()
	weatherTool := WeatherTool()
	registry.Register(calcTool)
	registry.Register(weatherTool)

	// 设置 Session Manager
	sessionMgr := NewMemorySessionManager()
	session, _ := sessionMgr.Create(context.Background(), "user1")

	// 设置 Tracer
	tracer := NewMemoryTracer()

	// 创建 Runtime
	config := &Config{
		MaxSteps:    10,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	// 执行
	resp, err := rt.Execute(context.Background(), session.ID, "Calculate 2+2 and check weather")

	// 验证
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Finished {
		t.Errorf("Expected Finished=true, got false")
	}

	if resp.LoopCount != 2 {
		t.Errorf("Expected LoopCount=2, got %d", resp.LoopCount)
	}

	if calcTool.GetCallCount() != 1 {
		t.Errorf("Expected 1 calculator call, got %d", calcTool.GetCallCount())
	}

	if weatherTool.GetCallCount() != 1 {
		t.Errorf("Expected 1 weather call, got %d", weatherTool.GetCallCount())
	}

	t.Logf("✓ MultipleToolCalls test passed")
}

// ============ 测试场景 4: ToolFailure ============
func TestToolFailure(t *testing.T) {
	// 设置 Fake LLM
	llm := NewFakeLLMClient([]*ChatResponse{
		ToolCallResponse("broken_tool", map[string]interface{}{"query": "test"}),
		DirectAnswer("Sorry, the tool failed with error: connection error"),
	})

	// 设置工具（包含会失败的工具）
	registry := NewSimpleToolRegistry()
	brokenTool := NewFailingTool("broken_tool", "connection error")
	registry.Register(brokenTool)

	// 设置 Session Manager
	sessionMgr := NewMemorySessionManager()
	session, _ := sessionMgr.Create(context.Background(), "user1")

	// 设置 Tracer
	tracer := NewMemoryTracer()

	// 创建 Runtime
	config := &Config{
		MaxSteps:    10,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	// 执行
	resp, err := rt.Execute(context.Background(), session.ID, "Use broken tool")

	// 验证：不应该 panic，应该继续执行
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !resp.Finished {
		t.Errorf("Expected Finished=true, got false")
	}

	// 检查错误消息是否被传递到 LLM
	sessionCtx, _ := sessionMgr.GetContext(context.Background(), session.ID)
	messages := sessionCtx.GetMessages()

	foundError := false
	for _, msg := range messages {
		if msg.Role == "tool" && strings.Contains(msg.Content, "connection error") {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected tool error message in context")
	}

	t.Logf("✓ ToolFailure test passed")
}

// ============ 测试场景 5: UnknownTool ============
func TestUnknownTool(t *testing.T) {
	// 设置 Fake LLM：请求不存在的工具
	llm := NewFakeLLMClient([]*ChatResponse{
		ToolCallResponse("unknown_tool", map[string]interface{}{"query": "test"}),
	})

	// 设置工具（不包含 unknown_tool）
	registry := NewSimpleToolRegistry()

	// 设置 Session Manager
	sessionMgr := NewMemorySessionManager()
	session, _ := sessionMgr.Create(context.Background(), "user1")

	// 设置 Tracer
	tracer := NewMemoryTracer()

	// 创建 Runtime
	config := &Config{
		MaxSteps:    10,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	// 执行
	resp, err := rt.Execute(context.Background(), session.ID, "Use unknown tool")

	// 验证
	if err == nil {
		t.Errorf("Expected error for unknown tool, got nil")
	}

	if resp.Finished {
		t.Errorf("Expected Finished=false, got true")
	}

	if !strings.Contains(resp.Error, "not found") {
		t.Errorf("Expected error to contain 'not found', got: %s", resp.Error)
	}

	t.Logf("✓ UnknownTool test passed")
}

// ============ 测试场景 6: MaxSteps ============
func TestMaxSteps(t *testing.T) {
	// 设置 Fake LLM：无限循环调用工具
	responses := make([]*ChatResponse, 0)
	for i := 0; i < 20; i++ {
		responses = append(responses, ToolCallResponse("calculator", map[string]interface{}{"expression": "1+1"}))
	}
	llm := NewFakeLLMClient(responses)

	// 设置工具
	registry := NewSimpleToolRegistry()
	calcTool := CalculatorTool()
	registry.Register(calcTool)

	// 设置 Session Manager
	sessionMgr := NewMemorySessionManager()
	session, _ := sessionMgr.Create(context.Background(), "user1")

	// 设置 Tracer
	tracer := NewMemoryTracer()

	// 创建 Runtime（MaxSteps=5）
	config := &Config{
		MaxSteps:    5,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	// 执行
	resp, err := rt.Execute(context.Background(), session.ID, "Keep calculating")

	// 验证
	if err != ErrMaxStepsExceeded {
		t.Errorf("Expected ErrMaxStepsExceeded, got: %v", err)
	}

	if resp.Finished {
		t.Errorf("Expected Finished=false, got true")
	}

	if resp.LoopCount != 5 {
		t.Errorf("Expected LoopCount=5, got %d", resp.LoopCount)
	}

	if resp.Error != ErrMaxStepsExceeded.Error() {
		t.Errorf("Expected Error='%s', got '%s'", ErrMaxStepsExceeded.Error(), resp.Error)
	}

	t.Logf("✓ MaxSteps test passed")
}

// ============ 测试场景 7: SessionIsolation ============
func TestSessionIsolation(t *testing.T) {
	// 设置 Fake LLM
	llm := NewFakeLLMClient([]*ChatResponse{
		DirectAnswer("Response for session 1"),
		DirectAnswer("Response for session 2"),
	})

	// 设置工具
	registry := NewSimpleToolRegistry()

	// 设置 Session Manager
	sessionMgr := NewMemorySessionManager()
	session1, _ := sessionMgr.Create(context.Background(), "user1")
	session2, _ := sessionMgr.Create(context.Background(), "user1")

	// 设置 Tracer
	tracer := NewMemoryTracer()

	// 创建 Runtime
	config := &Config{
		MaxSteps:    10,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	// 执行 Session 1
	resp1, err1 := rt.Execute(context.Background(), session1.ID, "Hello from session 1")
	if err1 != nil {
		t.Fatalf("Session 1 error: %v", err1)
	}

	// 执行 Session 2
	resp2, err2 := rt.Execute(context.Background(), session2.ID, "Hello from session 2")
	if err2 != nil {
		t.Fatalf("Session 2 error: %v", err2)
	}

	// 验证：两个 Session 的响应不同
	if resp1.SessionID == resp2.SessionID {
		t.Errorf("Expected different session IDs")
	}

	// 获取两个 Session 的 Context
	ctx1, _ := sessionMgr.GetContext(context.Background(), session1.ID)
	ctx2, _ := sessionMgr.GetContext(context.Background(), session2.ID)

	messages1 := ctx1.GetMessages()
	messages2 := ctx2.GetMessages()

	// 验证：Session 1 的消息数量
	if len(messages1) != 2 { // user + assistant
		t.Errorf("Expected 2 messages in session 1, got %d", len(messages1))
	}

	// 验证：Session 2 的消息数量
	if len(messages2) != 2 {
		t.Errorf("Expected 2 messages in session 2, got %d", len(messages2))
	}

	// 验证：Session 1 的消息不包含 Session 2 的内容
	for _, msg := range messages1 {
		if strings.Contains(msg.Content, "session 2") {
			t.Errorf("Session 1 should not contain session 2 content")
		}
	}

	// 验证：Session 2 的消息不包含 Session 1 的内容
	for _, msg := range messages2 {
		if strings.Contains(msg.Content, "session 1") {
			t.Errorf("Session 2 should not contain session 1 content")
		}
	}

	t.Logf("✓ SessionIsolation test passed")
}

// ============ 测试场景 8: FollowUpConversation ============
func TestFollowUpConversation(t *testing.T) {
	// 设置 Fake LLM
	llm := NewFakeLLMClient([]*ChatResponse{
		DirectAnswer("My name is Claude"),
		DirectAnswer("As I mentioned, my name is Claude"),
	})

	// 设置工具
	registry := NewSimpleToolRegistry()

	// 设置 Session Manager
	sessionMgr := NewMemorySessionManager()
	session, _ := sessionMgr.Create(context.Background(), "user1")

	// 设置 Tracer
	tracer := NewMemoryTracer()

	// 创建 Runtime
	config := &Config{
		MaxSteps:    10,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	// 第一次提问
	resp1, err1 := rt.Execute(context.Background(), session.ID, "What's your name?")
	if err1 != nil {
		t.Fatalf("First question error: %v", err1)
	}

	// 第二次提问（追问）
	resp2, err2 := rt.Execute(context.Background(), session.ID, "Can you repeat that?")
	if err2 != nil {
		t.Fatalf("Second question error: %v", err2)
	}

	// 验证：第二次调用能读取第一次的历史
	sessionCtx, _ := sessionMgr.GetContext(context.Background(), session.ID)
	messages := sessionCtx.GetMessages()

	// 应该有 4 条消息：user1, assistant1, user2, assistant2
	if len(messages) != 4 {
		t.Errorf("Expected 4 messages, got %d", len(messages))
	}

	// 验证消息顺序
	expectedRoles := []string{"user", "assistant", "user", "assistant"}
	for i, msg := range messages {
		if msg.Role != expectedRoles[i] {
			t.Errorf("Message %d: expected role '%s', got '%s'", i, expectedRoles[i], msg.Role)
		}
	}

	// 验证第一条用户消息
	if !strings.Contains(messages[0].Content, "name") {
		t.Errorf("First user message should contain 'name'")
	}

	// 验证第二条用户消息
	if !strings.Contains(messages[2].Content, "repeat") {
		t.Errorf("Second user message should contain 'repeat'")
	}

	t.Logf("✓ FollowUpConversation test passed")
}

// ============ 测试场景 9: FollowUpWithTool ============
func TestFollowUpWithTool(t *testing.T) {
	// 设置 Fake LLM
	llm := NewFakeLLMClient([]*ChatResponse{
		// 第一次：调用计算器
		ToolCallResponse("calculator", map[string]interface{}{"expression": "2+2"}),
		DirectAnswer("The result is 4"),
		// 第二次：基于之前的结果再调用
		ToolCallResponse("calculator", map[string]interface{}{"expression": "4*2"}),
		DirectAnswer("The result is 8, which is double of the previous result"),
	})

	// 设置工具
	registry := NewSimpleToolRegistry()
	calcTool := CalculatorTool()
	registry.Register(calcTool)

	// 设置 Session Manager
	sessionMgr := NewMemorySessionManager()
	session, _ := sessionMgr.Create(context.Background(), "user1")

	// 设置 Tracer
	tracer := NewMemoryTracer()

	// 创建 Runtime
	config := &Config{
		MaxSteps:    10,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	// 第一次提问：计算 2+2
	resp1, err1 := rt.Execute(context.Background(), session.ID, "Calculate 2+2")
	if err1 != nil {
		t.Fatalf("First question error: %v", err1)
	}

	if !resp1.Finished {
		t.Errorf("First question should be finished")
	}

	// 第二次提问：基于之前的结果
	resp2, err2 := rt.Execute(context.Background(), session.ID, "Now double that result")
	if err2 != nil {
		t.Fatalf("Second question error: %v", err2)
	}

	if !resp2.Finished {
		t.Errorf("Second question should be finished")
	}

	// 验证：Context 包含两次交互的所有消息
	sessionCtx, _ := sessionMgr.GetContext(context.Background(), session.ID)
	messages := sessionCtx.GetMessages()

	// 应该有：
	// user1, assistant1(tool_call), tool1, assistant1(response)
	// user2, assistant2(tool_call), tool2, assistant2(response)
	// = 8 条消息
	if len(messages) < 6 {
		t.Errorf("Expected at least 6 messages, got %d", len(messages))
	}

	// 验证第一次工具调用的结果在 Context 中
	foundFirstToolResult := false
	for _, msg := range messages {
		if msg.Role == "tool" && msg.Name == "calculator" {
			foundFirstToolResult = true
			break
		}
	}

	if !foundFirstToolResult {
		t.Errorf("First tool result should be in context")
	}

	t.Logf("✓ FollowUpWithTool test passed")
}

// ============ 测试场景 10: ContextCompression ============
func TestContextCompression(t *testing.T) {
	t.Skip("Context compression not yet implemented - placeholder test")

	// TODO: 实现 Context Compression 后取消 Skip
	// 这个测试需要：
	// 1. 创建大量消息（超过阈值）
	// 2. 触发压缩
	// 3. 验证消息数量减少
	// 4. 验证关键信息保留

	t.Logf("✓ ContextCompression test (placeholder)")
}

// ============ 辅助测试函数 ============

func TestTraceRecording(t *testing.T) {
	// 验证 Trace 是否正确记录
	llm := NewFakeLLMClient([]*ChatResponse{
		ToolCallResponse("calculator", map[string]interface{}{"expression": "2+2"}),
		DirectAnswer("The answer is 4"),
	})

	registry := NewSimpleToolRegistry()
	calcTool := CalculatorTool()
	registry.Register(calcTool)

	sessionMgr := NewMemorySessionManager()
	session, _ := sessionMgr.Create(context.Background(), "user1")

	tracer := NewMemoryTracer()

	config := &Config{
		MaxSteps:    10,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	rt := NewRuntime(config, llm, registry, sessionMgr, tracer)

	resp, _ := rt.Execute(context.Background(), session.ID, "Calculate 2+2")

	// 验证 Traces
	traces := resp.Traces
	if len(traces) == 0 {
		t.Errorf("Expected traces to be recorded")
	}

	// 验证 Trace 类型
	traceTypes := make(map[string]int)
	for _, trace := range traces {
		traceTypes[trace.Type]++
	}

	if traceTypes["llm_call"] < 2 {
		t.Errorf("Expected at least 2 llm_call traces, got %d", traceTypes["llm_call"])
	}

	if traceTypes["tool_call"] < 1 {
		t.Errorf("Expected at least 1 tool_call trace, got %d", traceTypes["tool_call"])
	}

	if traceTypes["loop_start"] < 2 {
		t.Errorf("Expected at least 2 loop_start traces, got %d", traceTypes["loop_start"])
	}

	t.Logf("✓ TraceRecording test passed")
}
