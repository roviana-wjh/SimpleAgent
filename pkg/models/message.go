package models

import "time"

// Message 表示对话中的一条消息
type Message struct {
	ID         string      `json:"id"`
	SessionID  string      `json:"session_id"`
	Role       string      `json:"role"` // "system" | "user" | "assistant" | "tool"
	Content    string      `json:"content"`
	Name       string      `json:"name,omitempty"`        // Tool 名称（role=tool 时）
	ToolCallID string      `json:"tool_call_id,omitempty"` // Tool Call ID（role=tool 时）
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`   // Tool 调用（role=assistant 时）
	CreatedAt  time.Time   `json:"created_at"`
}

// ToolCall 表示一次工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // "function"
	Function FunctionCall `json:"function"`
}

// FunctionCall 表示函数调用详情
type FunctionCall struct {
	Name      string `json:"name"`      // Tool 名称
	Arguments string `json:"arguments"` // JSON 格式的参数
}
