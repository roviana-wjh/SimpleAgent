package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Session 定义会话
type Session struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Messages  []*Message
}

// SessionContext 定义会话上下文
type SessionContext struct {
	session *Session
	mu      sync.RWMutex
}

// NewSessionContext 创建会话上下文
func NewSessionContext(session *Session) *SessionContext {
	return &SessionContext{
		session: session,
	}
}

// GetMessages 获取所有消息
func (sc *SessionContext) GetMessages() []*Message {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	messages := make([]*Message, len(sc.session.Messages))
	copy(messages, sc.session.Messages)
	return messages
}

// AddMessage 添加消息
func (sc *SessionContext) AddMessage(msg *Message) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if msg.Role == "" {
		return fmt.Errorf("message role cannot be empty")
	}

	sc.session.Messages = append(sc.session.Messages, msg)
	sc.session.UpdatedAt = time.Now()
	return nil
}

// Clear 清空消息
func (sc *SessionContext) Clear() error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.session.Messages = make([]*Message, 0)
	sc.session.UpdatedAt = time.Now()
	return nil
}

// ============ Session Manager ============

// SessionManager 定义会话管理器接口
type SessionManager interface {
	Create(ctx context.Context, userID string) (*Session, error)
	Get(ctx context.Context, sessionID string) (*Session, error)
	GetContext(ctx context.Context, sessionID string) (*SessionContext, error)
	Save(ctx context.Context, session *Session) error
	Delete(ctx context.Context, sessionID string) error
}

// MemorySessionManager 基于内存的会话管理器
type MemorySessionManager struct {
	sessions map[string]*Session
	contexts map[string]*SessionContext
	mu       sync.RWMutex
}

// NewMemorySessionManager 创建内存会话管理器
func NewMemorySessionManager() *MemorySessionManager {
	return &MemorySessionManager{
		sessions: make(map[string]*Session),
		contexts: make(map[string]*SessionContext),
	}
}

// Create 创建会话
func (m *MemorySessionManager) Create(ctx context.Context, userID string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &Session{
		ID:        generateID(),
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages:  make([]*Message, 0),
	}

	m.sessions[session.ID] = session
	m.contexts[session.ID] = NewSessionContext(session)

	return session, nil
}

// Get 获取会话
func (m *MemorySessionManager) Get(ctx context.Context, sessionID string) (*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// GetContext 获取会话上下文
func (m *MemorySessionManager) GetContext(ctx context.Context, sessionID string) (*SessionContext, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionCtx, exists := m.contexts[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return sessionCtx, nil
}

// Save 保存会话
func (m *MemorySessionManager) Save(ctx context.Context, session *Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session.UpdatedAt = time.Now()
	m.sessions[session.ID] = session
	return nil
}

// Delete 删除会话
func (m *MemorySessionManager) Delete(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, sessionID)
	delete(m.contexts, sessionID)
	return nil
}

// 辅助函数
func generateID() string {
	// 简单实现，实际应该使用 UUID
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
