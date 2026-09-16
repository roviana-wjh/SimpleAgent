package context

import (
	"time"
)

// Config 定义 Context 管理配置
type Config struct {
	// MaxMessages 触发压缩的消息数量阈值
	MaxMessages int

	// KeepRecentCount 压缩后保留的最近消息数量
	KeepRecentCount int

	// SystemPrompt Agent 的系统提示
	SystemPrompt string

	// EnableCompression 是否启用自动压缩
	EnableCompression bool

	// MaxAge 会话最大年龄（可选，0 表示不限制）
	MaxAge time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxMessages:       20,
		KeepRecentCount:   10,
		SystemPrompt:      "You are a helpful AI assistant. Use the available tools when needed to help the user.",
		EnableCompression: true,
		MaxAge:            0,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.KeepRecentCount <= 0 {
		c.KeepRecentCount = 10
	}
	if c.MaxMessages <= c.KeepRecentCount {
		c.MaxMessages = c.KeepRecentCount + 10
	}
	if c.SystemPrompt == "" {
		c.SystemPrompt = "You are a helpful AI assistant."
	}
	return nil
}
