package models

import "time"

// Session 表示一个会话
type Session struct {
	ID        string                 `json:"id"`         // Session UUID
	UserID    string                 `json:"user_id"`    // 用户 ID
	Messages  []*Message             `json:"messages"`   // 对话消息列表
	Summary   string                 `json:"summary"`    // 会话摘要（用于长对话压缩）
	CreatedAt time.Time              `json:"created_at"` // 创建时间
	UpdatedAt time.Time              `json:"updated_at"` // 更新时间
	Metadata  map[string]interface{} `json:"metadata"`   // 扩展元数据
}

// MessageCount 返回消息数量
func (s *Session) MessageCount() int {
	return len(s.Messages)
}

// AddMessage 添加消息到会话
func (s *Session) AddMessage(msg *Message) {
	s.Messages = append(s.Messages, msg)
	s.UpdatedAt = time.Now()
}

// GetMessages 获取所有消息
func (s *Session) GetMessages() []*Message {
	return s.Messages
}

// GetRecentMessages 获取最近 N 条消息
func (s *Session) GetRecentMessages(n int) []*Message {
	if n <= 0 || n >= len(s.Messages) {
		return s.Messages
	}
	return s.Messages[len(s.Messages)-n:]
}
