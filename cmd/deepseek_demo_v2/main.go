package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"agent-runtime/pkg/runtime"
)

func main() {
	// 从环境变量获取 API Key
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Println("Error: DEEPSEEK_API_KEY environment variable not set")
		log.Println("Please set it before running:")
		log.Println("  export DEEPSEEK_API_KEY=\"your-api-key\"")
		os.Exit(1)
	}

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║          Agent Runtime - DeepSeek Demo                ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 创建 DeepSeek LLM 客户端
	llmClient := runtime.NewDeepSeekClient(apiKey)
	fmt.Println("✓ DeepSeek LLM Client initialized")

	// 创建工具注册表
	toolRegistry := runtime.NewSimpleToolRegistry()

	// 注册真实工具（替代 Fake 工具）
	toolRegistry.Register(runtime.RealCalculator())
	toolRegistry.Register(runtime.RealSearch())
	toolRegistry.Register(runtime.RealWeather())

	fmt.Printf("✓ Registered 3 real tools: calculator, search, weather\n")

	// 创建 Session Manager
	sessionMgr := runtime.NewMemorySessionManager()
	fmt.Println("✓ Session Manager initialized")

	// 创建 Tracer
	tracer := runtime.NewMemoryTracer()
	fmt.Println("✓ Tracer initialized")

	// 配置 Runtime
	config := &runtime.Config{
		MaxSteps:          10,
		MaxTokens:         2000,
		Temperature:       0.7,
		CompressThreshold: 20,
	}

	// 创建 Runtime
	rt := runtime.NewRuntime(config, llmClient, toolRegistry, sessionMgr, tracer)
	fmt.Println("✓ Agent Runtime initialized")
	fmt.Println()

	// 创建一个新 Session
	session, err := sessionMgr.Create(context.Background(), "user1")
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	fmt.Printf("Session ID: %s\n", session.ID)
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	// 示例 1: 直接对话
	runExample(rt, session.ID, "示例 1: 直接对话", "你好，请用一句话介绍你自己")

	// 示例 2: 调用工具 - 计算
	runExample(rt, session.ID, "示例 2: 工具调用（计算器）", "请帮我计算 123 + 456")

	// 示例 3: 追问（基于历史上下文）
	runExample(rt, session.ID, "示例 3: 追问（上下文理解）", "那结果再乘以 2 是多少？")

	// 示例 4: 工具调用 - 天气
	runExample(rt, session.ID, "示例 4: 工具调用（天气）", "北京今天天气怎么样？")

	// 打印统计信息
	printStatistics(tracer, sessionMgr, session.ID)
}

// runExample 运行单个示例
func runExample(rt *runtime.Runtime, sessionID, title, userInput string) {
	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Printf("│ %s\n", title)
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	fmt.Printf("👤 User: %s\n", userInput)

	resp, err := rt.Execute(context.Background(), sessionID, userInput)
	if err != nil {
		fmt.Printf("❌ Error: %v\n\n", err)
		return
	}

	fmt.Printf("🤖 Assistant: %s\n", resp.FinalAnswer)
	fmt.Printf("📊 Stats: %d loops, finished=%t\n", resp.LoopCount, resp.Finished)

	// 打印 Tool 调用信息
	toolCalls := 0
	for _, trace := range resp.Traces {
		if trace.Type == "tool_call" {
			toolCalls++
		}
	}
	if toolCalls > 0 {
		fmt.Printf("🔧 Tools: %d tool call(s)\n", toolCalls)
	}

	fmt.Println()
}

// printStatistics 打印统计信息
func printStatistics(tracer runtime.Tracer, sessionMgr runtime.SessionManager, sessionID string) {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("                     统计信息")
	fmt.Println("═══════════════════════════════════════════════════════")

	// Trace 统计
	traces := tracer.GetTraces()
	traceStats := make(map[string]int)
	for _, trace := range traces {
		traceStats[trace.Type]++
	}

	fmt.Println("\n📈 Trace Statistics:")
	fmt.Printf("  Total traces: %d\n", len(traces))
	for traceType, count := range traceStats {
		fmt.Printf("  - %-15s: %d\n", traceType, count)
	}

	// 对话历史统计
	sessionCtx, _ := sessionMgr.GetContext(context.Background(), sessionID)
	messages := sessionCtx.GetMessages()

	fmt.Println("\n💬 Conversation Statistics:")
	fmt.Printf("  Total messages: %d\n", len(messages))

	roleStats := make(map[string]int)
	for _, msg := range messages {
		roleStats[msg.Role]++
	}

	for role, count := range roleStats {
		fmt.Printf("  - %-15s: %d\n", role, count)
	}

	fmt.Println()
}
