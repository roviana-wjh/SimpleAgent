package context_test

import (
	"testing"
	"time"

	"agent-runtime/pkg/context"
	"agent-runtime/pkg/models"

	"github.com/google/uuid"
)

func TestContextBuilder(t *testing.T) {
	config := &context.Config{
		MaxMessages:       10,
		KeepRecentCount:   5,
		SystemPrompt:      "Test system prompt",
		EnableCompression: true,
	}

	builder := context.NewBuilder(config)

	t.Run("Build with empty session", func(t *testing.T) {
		session := &models.Session{
			ID:        uuid.New().String(),
			UserID:    "test-user",
			Messages:  []*models.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		messages := builder.Build(session)

		// 应该只有 system message
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
		if messages[0].Role != "system" {
			t.Errorf("expected system role, got %s", messages[0].Role)
		}
	})

	t.Run("Build with messages", func(t *testing.T) {
		session := createTestSession(3)
		messages := builder.Build(session)

		// system + 3 messages
		if len(messages) != 4 {
			t.Errorf("expected 4 messages, got %d", len(messages))
		}
	})

	t.Run("Build with summary", func(t *testing.T) {
		session := createTestSession(3)
		session.Summary = "Previous conversation about weather"

		messages := builder.Build(session)

		// system + summary + 3 messages
		if len(messages) != 5 {
			t.Errorf("expected 5 messages, got %d", len(messages))
		}

		// 检查 summary message
		if messages[1].Role != "user" {
			t.Error("summary should be a user message")
		}
		if messages[1].Content == "" {
			t.Error("summary content should not be empty")
		}
	})
}

func TestCompression(t *testing.T) {
	config := &context.Config{
		MaxMessages:       10,
		KeepRecentCount:   5,
		SystemPrompt:      "Test system",
		EnableCompression: true,
	}

	builder := context.NewBuilder(config)

	t.Run("ShouldCompress", func(t *testing.T) {
		session := createTestSession(5)
		if builder.ShouldCompress(session) {
			t.Error("should not compress with 5 messages when threshold is 10")
		}

		session = createTestSession(15)
		if !builder.ShouldCompress(session) {
			t.Error("should compress with 15 messages when threshold is 10")
		}
	})

	t.Run("Compress session", func(t *testing.T) {
		session := createTestSession(15)
		originalCount := len(session.Messages)

		err := builder.Compress(session)
		if err != nil {
			t.Fatalf("compression failed: %v", err)
		}

		// 检查消息数量
		if len(session.Messages) != 5 {
			t.Errorf("expected 5 recent messages, got %d", len(session.Messages))
		}

		// 检查 summary 生成
		if session.Summary == "" {
			t.Error("summary should not be empty after compression")
		}

		// 检查压缩的消息数量
		compressedCount := originalCount - len(session.Messages)
		if compressedCount != 10 {
			t.Errorf("expected 10 compressed messages, got %d", compressedCount)
		}
	})

	t.Run("BuildWithCompression", func(t *testing.T) {
		session := createTestSession(15)

		messages, compressed := builder.BuildWithCompression(session)

		if !compressed {
			t.Error("session should have been compressed")
		}

		// system + summary + 5 recent messages
		if len(messages) != 7 {
			t.Errorf("expected 7 messages after compression, got %d", len(messages))
		}

		// 验证顺序
		if messages[0].Role != "system" {
			t.Error("first message should be system")
		}
		if messages[1].Role != "user" || messages[1].Content[:9] != "[Previous" {
			t.Error("second message should be summary")
		}
	})

	t.Run("Multiple compressions", func(t *testing.T) {
		session := createTestSession(15)

		// 第一次压缩
		builder.Compress(session)
		firstSummary := session.Summary

		// 添加更多消息
		for i := 0; i < 10; i++ {
			session.AddMessage(&models.Message{
				ID:        uuid.New().String(),
				SessionID: session.ID,
				Role:      "user",
				Content:   "New message after compression",
				CreatedAt: time.Now(),
			})
		}

		// 第二次压缩
		builder.Compress(session)
		secondSummary := session.Summary

		// 第二次 summary 应该包含第一次的内容
		if len(secondSummary) <= len(firstSummary) {
			t.Error("second summary should be longer than first")
		}
	})
}

func TestCompressionWithToolCalls(t *testing.T) {
	config := context.DefaultConfig()
	config.MaxMessages = 10
	config.KeepRecentCount = 3

	builder := context.NewBuilder(config)

	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    "test-user",
		Messages:  []*models.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 添加包含 Tool Call 的对话
	session.AddMessage(&models.Message{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Role:      "user",
		Content:   "What's the weather in Beijing?",
		CreatedAt: time.Now(),
	})

	session.AddMessage(&models.Message{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Role:      "assistant",
		ToolCalls: []models.ToolCall{
			{
				ID:   "call_1",
				Type: "function",
				Function: models.FunctionCall{
					Name:      "weather",
					Arguments: `{"location": "Beijing"}`,
				},
			},
		},
		CreatedAt: time.Now(),
	})

	session.AddMessage(&models.Message{
		ID:         uuid.New().String(),
		SessionID:  session.ID,
		Role:       "tool",
		Name:       "weather",
		ToolCallID: "call_1",
		Content:    "Temperature: 22°C, Sunny",
		CreatedAt:  time.Now(),
	})

	session.AddMessage(&models.Message{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Role:      "assistant",
		Content:   "The weather in Beijing is sunny with 22°C.",
		CreatedAt: time.Now(),
	})

	// 添加更多消息触发压缩
	for i := 0; i < 10; i++ {
		session.AddMessage(&models.Message{
			ID:        uuid.New().String(),
			SessionID: session.ID,
			Role:      "user",
			Content:   "Additional message",
			CreatedAt: time.Now(),
		})
	}

	err := builder.Compress(session)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}

	// 检查 summary 包含 tool call 信息
	if session.Summary == "" {
		t.Error("summary should not be empty")
	}

	t.Logf("Summary:\n%s", session.Summary)
}

func TestConfigValidation(t *testing.T) {
	t.Run("Default config", func(t *testing.T) {
		config := context.DefaultConfig()
		if config.MaxMessages <= 0 {
			t.Error("MaxMessages should be positive")
		}
		if config.KeepRecentCount <= 0 {
			t.Error("KeepRecentCount should be positive")
		}
		if config.SystemPrompt == "" {
			t.Error("SystemPrompt should not be empty")
		}
	})

	t.Run("Invalid config validation", func(t *testing.T) {
		config := &context.Config{
			MaxMessages:     5,
			KeepRecentCount: 10, // 错误：保留数量大于最大数量
		}

		config.Validate()

		if config.MaxMessages <= config.KeepRecentCount {
			t.Error("MaxMessages should be greater than KeepRecentCount after validation")
		}
	})
}

// Helper function
func createTestSession(messageCount int) *models.Session {
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    "test-user",
		Messages:  []*models.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	for i := 0; i < messageCount; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}

		msg := &models.Message{
			ID:        uuid.New().String(),
			SessionID: session.ID,
			Role:      role,
			Content:   "Test message " + uuid.New().String()[:8],
			CreatedAt: time.Now(),
		}
		session.AddMessage(msg)
	}

	return session
}
