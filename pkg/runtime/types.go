package runtime

import (
	"errors"
	"time"
)

// Message 定义消息
type Message struct {
	Role      string      // "system", "user", "assistant", "tool"
	Content   string      // 消息内容
	Name      string      // tool role 时使用（工具名称）
	ToolCalls []ToolCall  // assistant role 时可能有工具调用
}

// ToolCall 定义工具调用
type ToolCall struct {
	ID     string                 // 调用 ID
	Name   string                 // 工具名称
	Params map[string]interface{} // 参数
}

// ToolResult 定义工具结果
type ToolResult struct {
	Success bool
	Output  string
	Error   string
}

// ChatRequest LLM 请求
type ChatRequest struct {
	Messages []*Message
	Tools    []*ToolDef
}

// ChatResponse LLM 响应
type ChatResponse struct {
	Content   string
	ToolCalls []ToolCall
	Error     string
}

// ToolDef 工具定义
type ToolDef struct {
	Name        string
	Description string
	ParamSchema map[string]interface{}
}

// Trace 追踪记录
type Trace struct {
	Type      string        // "llm_call", "tool_call", "loop_start", "loop_end", "error"
	StepIndex int           // 第几步
	Input     interface{}   // 输入
	Output    interface{}   // 输出
	Error     string        // 错误信息
	Duration  time.Duration // 执行时间
	Timestamp time.Time     // 时间戳
}

// Response Runtime 响应
type Response struct {
	SessionID   string
	FinalAnswer string
	LoopCount   int
	Traces      []*Trace
	Finished    bool
	Error       string
}

// Config Runtime 配置
type Config struct {
	MaxSteps      int
	MaxTokens     int
	Temperature   float64
	CompressThreshold int // Context 压缩阈值（消息数量）
}

// 定义错误
var (
	ErrMaxStepsExceeded = errors.New("maximum steps exceeded")
	ErrSessionNotFound  = errors.New("session not found")
	ErrToolNotFound     = errors.New("tool not found")
	ErrInvalidInput     = errors.New("invalid input")
)
