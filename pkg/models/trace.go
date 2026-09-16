package models

import "time"

// TraceType 定义追踪类型
type TraceType string

const (
	TraceLLMCall   TraceType = "llm_call"
	TraceToolCall  TraceType = "tool_call"
	TraceLoopStart TraceType = "loop_start"
	TraceLoopEnd   TraceType = "loop_end"
	TraceError     TraceType = "error"
)

// Trace 表示一条追踪记录
type Trace struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	StepIndex int                    `json:"step_index"`
	Type      TraceType              `json:"type"`
	Input     interface{}            `json:"input,omitempty"`
	Output    interface{}            `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
