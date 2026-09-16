package context

import (
	"fmt"
	"strings"
	"time"

	"agent-runtime/pkg/models"

	"github.com/google/uuid"
)

// Builder 负责构建 LLM Context
type Builder struct {
	config *Config
}

// NewBuilder 创建新的 Context Builder
func NewBuilder(config *Config) *Builder {
	if config == nil {
		config = DefaultConfig()
	}
	config.Validate()

	return &Builder{
		config: config,
	}
}

// Build 构建用于 LLM 的消息列表
func (b *Builder) Build(session *models.Session) []*models.Message {
	messages := make([]*models.Message, 0)

	// 1. System Prompt (必须)
	systemMsg := &models.Message{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Role:      "system",
		Content:   b.config.SystemPrompt,
		CreatedAt: time.Now(),
	}
	messages = append(messages, systemMsg)

	// 2. Session Summary (如果存在)
	if session.Summary != "" {
		summaryMsg := &models.Message{
			ID:        uuid.New().String(),
			SessionID: session.ID,
			Role:      "user",
			Content:   fmt.Sprintf("[Previous Conversation Summary]\n%s\n[End of Summary]", session.Summary),
			CreatedAt: time.Now(),
		}
		messages = append(messages, summaryMsg)
	}

	// 3. Recent Conversation Messages
	recentMessages := session.GetMessages()
	messages = append(messages, recentMessages...)

	return messages
}

// BuildWithCompression 构建消息列表，并在需要时自动压缩
func (b *Builder) BuildWithCompression(session *models.Session) ([]*models.Message, bool) {
	compressed := false

	// 检查是否需要压缩
	if b.config.EnableCompression && b.ShouldCompress(session) {
		if err := b.Compress(session); err == nil {
			compressed = true
		}
	}

	messages := b.Build(session)
	return messages, compressed
}

// ShouldCompress 判断是否需要压缩
func (b *Builder) ShouldCompress(session *models.Session) bool {
	// 条件 1: 消息数量超过阈值
	if len(session.Messages) > b.config.MaxMessages {
		return true
	}

	// 条件 2: 会话年龄超过阈值（可选）
	if b.config.MaxAge > 0 {
		age := time.Since(session.CreatedAt)
		if age > b.config.MaxAge {
			return true
		}
	}

	return false
}

// Compress 压缩会话历史
func (b *Builder) Compress(session *models.Session) error {
	messages := session.Messages

	// 如果消息数量不超过保留数量，无需压缩
	if len(messages) <= b.config.KeepRecentCount {
		return nil
	}

	// 1. 分离早期消息和最近消息
	splitPoint := len(messages) - b.config.KeepRecentCount
	oldMessages := messages[:splitPoint]
	recentMessages := messages[splitPoint:]

	// 2. 生成 Summary
	summary := b.generateSummary(oldMessages, session.Summary)

	// 3. 更新 Session
	session.Summary = summary
	session.Messages = recentMessages
	session.UpdatedAt = time.Now()

	return nil
}

// generateSummary 生成对话摘要
func (b *Builder) generateSummary(messages []*models.Message, existingSummary string) string {
	var summaryParts []string

	// 如果已有摘要，先添加
	if existingSummary != "" {
		summaryParts = append(summaryParts, existingSummary)
		summaryParts = append(summaryParts, "\n--- New Messages ---")
	}

	// 遍历消息生成摘要
	for _, msg := range messages {
		switch msg.Role {
		case "user":
			content := truncate(msg.Content, 80)
			summaryParts = append(summaryParts,
				fmt.Sprintf("• User: %s", content))

		case "assistant":
			if len(msg.ToolCalls) > 0 {
				// 有 Tool Calls
				toolNames := make([]string, 0, len(msg.ToolCalls))
				for _, tc := range msg.ToolCalls {
					toolNames = append(toolNames, tc.Function.Name)
				}
				summaryParts = append(summaryParts,
					fmt.Sprintf("• Assistant called: %s", strings.Join(toolNames, ", ")))
			} else if msg.Content != "" {
				// 普通回复
				content := truncate(msg.Content, 80)
				summaryParts = append(summaryParts,
					fmt.Sprintf("• Assistant: %s", content))
			}

		case "tool":
			// Tool 结果（简化）
			if msg.Name != "" {
				content := truncate(msg.Content, 60)
				summaryParts = append(summaryParts,
					fmt.Sprintf("  ↳ %s returned: %s", msg.Name, content))
			}
		}
	}

	return strings.Join(summaryParts, "\n")
}

// truncate 截断文本
func truncate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// GetConfig 获取配置
func (b *Builder) GetConfig() *Config {
	return b.config
}
