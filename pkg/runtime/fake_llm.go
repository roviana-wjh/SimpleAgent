package runtime

import (
	"context"
	"sync"
)

// LLMClient 定义 LLM 客户端接口
type LLMClient interface {
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

// FakeLLMClient 用于测试的 Fake LLM 客户端
type FakeLLMClient struct {
	responses []*ChatResponse
	callIndex int
	calls     []*ChatRequest
	mu        sync.Mutex
}

// NewFakeLLMClient 创建 Fake LLM Client
func NewFakeLLMClient(responses []*ChatResponse) *FakeLLMClient {
	return &FakeLLMClient{
		responses: responses,
		callIndex: 0,
		calls:     make([]*ChatRequest, 0),
	}
}

// Chat 实现 LLMClient 接口
func (c *FakeLLMClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 保存调用记录
	c.calls = append(c.calls, req)

	// 返回预定义的响应
	if c.callIndex >= len(c.responses) {
		return &ChatResponse{
			Content: "No more responses defined",
		}, nil
	}

	resp := c.responses[c.callIndex]
	c.callIndex++
	return resp, nil
}

// GetCalls 获取所有 LLM 调用
func (c *FakeLLMClient) GetCalls() []*ChatRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

// GetCallCount 获取调用次数
func (c *FakeLLMClient) GetCallCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.calls)
}

// Reset 重置状态
func (c *FakeLLMClient) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.callIndex = 0
	c.calls = make([]*ChatRequest, 0)
}

// 辅助函数用于创建常见的响应

// DirectAnswer 创建直接回答响应
func DirectAnswer(content string) *ChatResponse {
	return &ChatResponse{
		Content:   content,
		ToolCalls: nil,
	}
}

// ToolCallResponse 创建工具调用响应
func ToolCallResponse(toolName string, params map[string]interface{}) *ChatResponse {
	return &ChatResponse{
		Content: "",
		ToolCalls: []ToolCall{
			{
				ID:     "call_" + toolName,
				Name:   toolName,
				Params: params,
			},
		},
	}
}

// MultipleToolCallResponse 创建多个工具调用响应
func MultipleToolCallResponse(toolCalls []ToolCall) *ChatResponse {
	return &ChatResponse{
		Content:   "",
		ToolCalls: toolCalls,
	}
}

// ErrorResponse 创建错误响应
func ErrorResponse(errMsg string) *ChatResponse {
	return &ChatResponse{
		Content: "",
		Error:   errMsg,
	}
}
