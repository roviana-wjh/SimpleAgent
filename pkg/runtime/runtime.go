package runtime

import (
	"context"
	"fmt"
	"time"
)

// Runtime 定义 Agent Runtime
type Runtime struct {
	config       *Config
	llmClient    LLMClient
	toolRegistry ToolRegistry
	sessionMgr   SessionManager
	tracer       Tracer
}

// NewRuntime 创建 Runtime
func NewRuntime(
	config *Config,
	llmClient LLMClient,
	toolRegistry ToolRegistry,
	sessionMgr SessionManager,
	tracer Tracer,
) *Runtime {
	return &Runtime{
		config:       config,
		llmClient:    llmClient,
		toolRegistry: toolRegistry,
		sessionMgr:   sessionMgr,
		tracer:       tracer,
	}
}

// Execute 执行 Agent 交互
func (rt *Runtime) Execute(ctx context.Context, sessionID string, userInput string) (*Response, error) {
	response := &Response{
		SessionID:  sessionID,
		LoopCount:  0,
		Traces:     make([]*Trace, 0),
		Finished:   false,
	}

	// 获取会话
	session, err := rt.sessionMgr.Get(ctx, sessionID)
	if err != nil {
		response.Error = err.Error()
		return response, err
	}

	// 获取会话上下文
	sessionCtx, err := rt.sessionMgr.GetContext(ctx, sessionID)
	if err != nil {
		response.Error = err.Error()
		return response, err
	}

	// 添加用户消息
	userMsg := &Message{
		Role:    "user",
		Content: userInput,
	}
	if err := sessionCtx.AddMessage(userMsg); err != nil {
		response.Error = err.Error()
		return response, err
	}

	// 主循环
	for response.LoopCount < rt.config.MaxSteps {
		response.LoopCount++

		// 记录 Loop 开始
		rt.recordTrace(&Trace{
			Type:      "loop_start",
			StepIndex: response.LoopCount,
			Timestamp: time.Now(),
		})

		// 获取消息（应用 Context 压缩）
		allMessages := sessionCtx.GetMessages()
		messages := rt.compressMessages(allMessages)

		// 准备 LLM 请求
		tools := rt.toolRegistry.GetAll()
		toolDefs := make([]*ToolDef, 0, len(tools))
		for _, tool := range tools {
			toolDefs = append(toolDefs, &ToolDef{
				Name:        tool.Name(),
				Description: tool.Description(),
				ParamSchema: tool.ParamSchema(),
			})
		}

		// 调用 LLM
		llmReq := &ChatRequest{
			Messages: messages,
			Tools:    toolDefs,
		}

		llmStart := time.Now()
		llmResp, err := rt.llmClient.Chat(ctx, llmReq)
		llmDuration := time.Since(llmStart)

		// 记录 LLM 调用
		rt.recordTrace(&Trace{
			Type:      "llm_call",
			StepIndex: response.LoopCount,
			Input:     llmReq,
			Output:    llmResp,
			Duration:  llmDuration,
			Timestamp: time.Now(),
		})

		if err != nil {
			response.Error = err.Error()
			response.Finished = false
			response.Traces = rt.tracer.GetTraces()
			return response, err
		}

		// 处理 LLM 响应错误
		if llmResp.Error != "" {
			response.Error = llmResp.Error
			response.Finished = false
			response.Traces = rt.tracer.GetTraces()
			return response, fmt.Errorf(llmResp.Error)
		}

		// 如果没有 Tool Call，返回最终答案
		if len(llmResp.ToolCalls) == 0 {
			response.FinalAnswer = llmResp.Content
			response.Finished = true

			// 添加 Assistant 消息
			assistantMsg := &Message{
				Role:    "assistant",
				Content: llmResp.Content,
			}
			sessionCtx.AddMessage(assistantMsg)

			// 记录 Loop 结束
			rt.recordTrace(&Trace{
				Type:      "loop_end",
				StepIndex: response.LoopCount,
				Timestamp: time.Now(),
			})

			// 保存会话
			if err := rt.sessionMgr.Save(ctx, session); err != nil {
				response.Error = err.Error()
				return response, err
			}

			response.Traces = rt.tracer.GetTraces()
			return response, nil
		}

		// 添加 Assistant 消息（包含 Tool Call）
		assistantMsg := &Message{
			Role:      "assistant",
			Content:   "",
			ToolCalls: llmResp.ToolCalls,
		}
		sessionCtx.AddMessage(assistantMsg)

		// 执行 Tool Call
		for _, toolCall := range llmResp.ToolCalls {
			tool := rt.toolRegistry.Get(toolCall.Name)
			if tool == nil {
				errMsg := fmt.Sprintf("tool %s not found", toolCall.Name)
				response.Error = errMsg
				response.Finished = false
				response.Traces = rt.tracer.GetTraces()
				return response, fmt.Errorf(errMsg)
			}

			// 执行工具
			toolStart := time.Now()
			result, toolErr := tool.Execute(ctx, toolCall.Params)
			toolDuration := time.Since(toolStart)

			// 记录 Tool 调用
			toolErrMsg := ""
			if toolErr != nil {
				toolErrMsg = toolErr.Error()
			}
			rt.recordTrace(&Trace{
				Type:      "tool_call",
				StepIndex: response.LoopCount,
				Input: map[string]interface{}{
					"tool_name": toolCall.Name,
					"params":    toolCall.Params,
				},
				Output:    result,
				Error:     toolErrMsg,
				Duration:  toolDuration,
				Timestamp: time.Now(),
			})

			// 添加 Tool 结果消息
			toolContent := result
			if toolErr != nil {
				toolContent = fmt.Sprintf("Error: %s", toolErr.Error())
			}

			toolMsg := &Message{
				Role:    "tool",
				Content: toolContent,
				Name:    toolCall.Name,
				ToolCalls: []ToolCall{
					{
						ID: toolCall.ID, // 必须包含 tool_call_id
					},
				},
			}
			sessionCtx.AddMessage(toolMsg)
		}

		// 记录 Loop 结束
		rt.recordTrace(&Trace{
			Type:      "loop_end",
			StepIndex: response.LoopCount,
			Timestamp: time.Now(),
		})
	}

	// 超过最大步数
	response.Error = ErrMaxStepsExceeded.Error()
	response.Finished = false
	response.Traces = rt.tracer.GetTraces()

	return response, ErrMaxStepsExceeded
}

// recordTrace 记录追踪
func (rt *Runtime) recordTrace(trace *Trace) {
	if rt.tracer != nil {
		_ = rt.tracer.Record(trace)
	}
}

// compressMessages 压缩消息（保留 system 和最近的消息）
func (rt *Runtime) compressMessages(messages []*Message) []*Message {
	// 如果消息数量未超过阈值，直接返回
	if len(messages) <= rt.config.CompressThreshold {
		return messages
	}

	compressed := make([]*Message, 0)

	// 1. 保留所有 system 消息
	for _, msg := range messages {
		if msg.Role == "system" {
			compressed = append(compressed, msg)
		}
	}

	// 2. 计算要保留的最近消息数量（保留一半）
	keepRecent := rt.config.CompressThreshold / 2
	if keepRecent < 5 {
		keepRecent = 5 // 至少保留 5 条
	}

	// 3. 保留最近的消息
	startIdx := len(messages) - keepRecent
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(messages); i++ {
		if messages[i].Role != "system" { // 避免重复添加 system 消息
			compressed = append(compressed, messages[i])
		}
	}

	return compressed
}

