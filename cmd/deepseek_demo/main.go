package main

import (
	"context"
	"fmt"
	"log"

	"agent-runtime/pkg/runtime"
)

func main() {
	// DeepSeek API Key
	apiKey := "sk-7646c9243c5643d195212c5987695751"

	// 创建 DeepSeek LLM 客户端
	llmClient := runtime.NewDeepSeekClient(apiKey)

	// 创建工具注册表
	toolRegistry := runtime.NewSimpleToolRegistry()

	// 注册工具
	calcTool := runtime.CalculatorTool()
	searchTool := runtime.SearchTool()
	weatherTool := runtime.WeatherTool()

	toolRegistry.Register(calcTool)
	toolRegistry.Register(searchTool)
	toolRegistry.Register(weatherTool)

	// 创建 Session Manager
	sessionMgr := runtime.NewMemorySessionManager()

	// 创建 Tracer
	tracer := runtime.NewMemoryTracer()

	// 配置 Runtime
	config := &runtime.Config{
		MaxSteps:          10,
		MaxTokens:         2000,
		Temperature:       0.7,
		CompressThreshold: 20,
	}

	// 创建 Runtime
	rt := runtime.NewRuntime(config, llmClient, toolRegistry, sessionMgr, tracer)

	// 创建一个新 Session
	session, err := sessionMgr.Create(context.Background(), "user1")
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	fmt.Printf("Created session: %s\n\n", session.ID)

	// 示例 1: 直接对话
	fmt.Println("=== Example 1: Direct Conversation ===")
	resp1, err := rt.Execute(context.Background(), session.ID, "你好，请介绍一下你自己")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("User: 你好，请介绍一下你自己\n")
		fmt.Printf("Assistant: %s\n", resp1.FinalAnswer)
		fmt.Printf("Loop Count: %d\n", resp1.LoopCount)
		fmt.Printf("Finished: %t\n\n", resp1.Finished)
	}

	// 示例 2: 调用工具
	fmt.Println("=== Example 2: Tool Call ===")
	resp2, err := rt.Execute(context.Background(), session.ID, "请帮我计算 123 + 456")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("User: 请帮我计算 123 + 456\n")
		fmt.Printf("Assistant: %s\n", resp2.FinalAnswer)
		fmt.Printf("Loop Count: %d\n", resp2.LoopCount)
		fmt.Printf("Finished: %t\n\n", resp2.Finished)
	}

	// 示例 3: 追问（基于历史上下文）
	fmt.Println("=== Example 3: Follow-up Question ===")
	resp3, err := rt.Execute(context.Background(), session.ID, "那结果再乘以 2 是多少？")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("User: 那结果再乘以 2 是多少？\n")
		fmt.Printf("Assistant: %s\n", resp3.FinalAnswer)
		fmt.Printf("Loop Count: %d\n", resp3.LoopCount)
		fmt.Printf("Finished: %t\n\n", resp3.Finished)
	}

	// 打印 Trace 信息
	fmt.Println("=== Traces ===")
	traces := tracer.GetTraces()
	fmt.Printf("Total traces: %d\n", len(traces))

	// 统计 Trace 类型
	traceStats := make(map[string]int)
	for _, trace := range traces {
		traceStats[trace.Type]++
	}

	fmt.Println("\nTrace Statistics:")
	for traceType, count := range traceStats {
		fmt.Printf("  %s: %d\n", traceType, count)
	}

	// 打印对话历史
	fmt.Println("\n=== Conversation History ===")
	sessionCtx, _ := sessionMgr.GetContext(context.Background(), session.ID)
	messages := sessionCtx.GetMessages()
	fmt.Printf("Total messages: %d\n\n", len(messages))

	for i, msg := range messages {
		fmt.Printf("[%d] Role: %s\n", i+1, msg.Role)
		if msg.Content != "" {
			fmt.Printf("    Content: %s\n", truncate(msg.Content, 100))
		}
		if len(msg.ToolCalls) > 0 {
			fmt.Printf("    Tool Calls: %d\n", len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				fmt.Printf("      - %s (ID: %s)\n", tc.Name, tc.ID)
			}
		}
		if msg.Name != "" {
			fmt.Printf("    Tool Name: %s\n", msg.Name)
		}
		fmt.Println()
	}
}

// truncate 截断长文本
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
