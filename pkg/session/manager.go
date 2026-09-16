package session

import (
	"context"
	"fmt"
	"time"

	"agent-runtime/pkg/models"

	"github.com/google/uuid"
)

// Manager 会话管理器
type Manager struct {
	store    Store
	maxLoops int
}

// NewManager 创建新的会话管理器
func NewManager(store Store, maxLoops int) *Manager {
	if maxLoops <= 0 {
		maxLoops = 10
	}
	return &Manager{
		store:    store,
		maxLoops: maxLoops,
	}
}

// Create 创建新会话
func (m *Manager) Create(ctx context.Context, userID string) (*models.Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	now := time.Now()
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Messages:  make([]*models.Message, 0),
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]interface{}),
	}

	if err := m.store.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return session, nil
}

// Get 获取会话
func (m *Manager) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID cannot be empty")
	}

	session, err := m.store.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// Update 更新会话
func (m *Manager) Update(ctx context.Context, session *models.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	session.UpdatedAt = time.Now()
	return m.store.Save(ctx, session)
}

// Delete 删除会话
func (m *Manager) Delete(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}

// ListByUser 列出用户的所有会话
func (m *Manager) ListByUser(ctx context.Context, userID string) ([]*models.Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	return m.store.ListByUser(ctx, userID)
}

// GetOrCreateContext 获取或创建会话上下文（实现 runtime.SessionManager 接口）
func (m *Manager) GetOrCreateContext(ctx context.Context, sessionID string) (*Context, error) {
	// 尝试获取现有会话
	session, err := m.store.Get(ctx, sessionID)
	if err != nil {
		// 如果会话不存在，返回错误（不自动创建）
		// 用户需要显式调用 Create 创建会话
		return nil, fmt.Errorf("session %s not found: %w", sessionID, err)
	}

	// 创建会话上下文
	return NewContext(session, m.maxLoops), nil
}

// SaveContext 保存会话上下文（实现 runtime.SessionManager 接口）
func (m *Manager) SaveContext(ctx context.Context, sessionID string, sessionCtx *Context) error {
	if sessionCtx == nil {
		return fmt.Errorf("session context cannot be nil")
	}

	// 获取底层 Session 对象并保存
	session := sessionCtx.GetSession()
	return m.Update(ctx, session)
}

// Store 定义会话存储接口
type Store interface {
	// Save 保存会话
	Save(ctx context.Context, session *models.Session) error

	// Get 获取会话
	Get(ctx context.Context, sessionID string) (*models.Session, error)

	// Delete 删除会话
	Delete(ctx context.Context, sessionID string) error

	// ListByUser 列出用户的所有会话
	ListByUser(ctx context.Context, userID string) ([]*models.Session, error)
}
