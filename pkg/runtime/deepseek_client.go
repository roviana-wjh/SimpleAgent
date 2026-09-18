package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DeepSeekClient DeepSeek LLM 客户端
type DeepSeekClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
}

// NewDeepSeekClient 创建 DeepSeek 客户端
func NewDeepSeekClient(apiKey string) *DeepSeekClient {
	return &DeepSeekClient{
		apiKey:  apiKey,
		baseURL: "https://api.deepseek.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		model: "deepseek-chat", // 默认模型
	}
}

// SetModel 设置模型
func (c *DeepSeekClient) SetModel(model string) {
	c.model = model
}

// DeepSeek API 请求结构
type deepseekRequest struct {
	Model       string                   `json:"model"`
	Messages    []deepseekMessage        `json:"messages"`
	Tools       []deepseekTool           `json:"tools,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
}

type deepseekMessage struct {
	Role       string                 `json:"role"`
	Content    string                 `json:"content,omitempty"`
	Name       string                 `json:"name,omitempty"`
	ToolCalls  []deepseekToolCall     `json:"tool_calls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
}

type deepseekToolCall struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"`
	Function deepseekFunctionCall `json:"function"`
}

type deepseekFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type deepseekTool struct {
	Type     string               `json:"type"`
	Function deepseekFunctionDef  `json:"function"`
}

type deepseekFunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// DeepSeek API 响应结构
type deepseekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int             `json:"index"`
		Message      deepseekMessage `json:"message"`
		FinishReason string          `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Chat 实现 LLMClient 接口
func (c *DeepSeekClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 转换消息格式
	messages := make([]deepseekMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		dsMsg := deepseekMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}

		// 处理 tool_calls
		if len(msg.ToolCalls) > 0 {
			dsMsg.ToolCalls = make([]deepseekToolCall, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				argsJSON, _ := json.Marshal(tc.Params)
				dsMsg.ToolCalls = append(dsMsg.ToolCalls, deepseekToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: deepseekFunctionCall{
						Name:      tc.Name,
						Arguments: string(argsJSON),
					},
				})
			}
		}

		// 处理 tool role 的 tool_call_id
		if msg.Role == "tool" && len(msg.ToolCalls) > 0 {
			dsMsg.ToolCallID = msg.ToolCalls[0].ID
		}

		messages = append(messages, dsMsg)
	}

	// 转换工具格式
	tools := make([]deepseekTool, 0, len(req.Tools))
	for _, tool := range req.Tools {
		tools = append(tools, deepseekTool{
			Type: "function",
			Function: deepseekFunctionDef{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.ParamSchema,
			},
		})
	}

	// 构建请求
	dsReq := deepseekRequest{
		Model:       c.model,
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	reqBody, err := json.Marshal(dsReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", httpResp.StatusCode, string(respBody))
	}

	// 解析响应
	var dsResp deepseekResponse
	if err := json.Unmarshal(respBody, &dsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if dsResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", dsResp.Error.Message)
	}

	if len(dsResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := dsResp.Choices[0]

	// 转换响应格式
	resp := &ChatResponse{
		Content: choice.Message.Content,
	}

	// 处理 tool calls
	if len(choice.Message.ToolCalls) > 0 {
		resp.ToolCalls = make([]ToolCall, 0, len(choice.Message.ToolCalls))
		for _, tc := range choice.Message.ToolCalls {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &params); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool call arguments: %w", err)
			}

			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				ID:     tc.ID,
				Name:   tc.Function.Name,
				Params: params,
			})
		}
	}

	return resp, nil
}
