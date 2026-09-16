package session

import (
	"context"
	"fmt"
	"sync"

	"agent-runtime/pkg/models"
)

// MemoryStore 内存存储实现
type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*models.Session // sessionID -> Session
	userSessions map[string][]string    // userID -> []sessionID
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions:     make(map[string]*models.Session),
		userSessions: make(map[string][]string),
	}
}

// Save 保存会话
func (s *MemoryStore) Save(ctx context.Context, session *models.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}
	if session.ID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否是新会话
	_, exists := s.sessions[session.ID]

	// 保存会话
	s.sessions[session.ID] = session

	// 如果是新会话，更新 userSessions 索引
	if !exists {
		s.userSessions[session.UserID] = append(s.userSessions[session.UserID], session.ID)
	}

	return nil
}

// Get 获取会话
func (s *MemoryStore) Get(ctx context.Context, sessionID string) (*models.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

// Delete 删除会话
func (s *MemoryStore) Delete(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// 从 sessions 中删除
	delete(s.sessions, sessionID)

	// 从 userSessions 索引中删除
	userID := session.UserID
	if sessionIDs, ok := s.userSessions[userID]; ok {
		newSessionIDs := make([]string, 0, len(sessionIDs)-1)
		for _, sid := range sessionIDs {
			if sid != sessionID {
				newSessionIDs = append(newSessionIDs, sid)
			}
		}
		if len(newSessionIDs) > 0 {
			s.userSessions[userID] = newSessionIDs
		} else {
			delete(s.userSessions, userID)
		}
	}

	return nil
}

// ListByUser 列出用户的所有会话
func (s *MemoryStore) ListByUser(ctx context.Context, userID string) ([]*models.Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	sessionIDs, exists := s.userSessions[userID]
	if !exists {
		return []*models.Session{}, nil
	}

	sessions := make([]*models.Session, 0, len(sessionIDs))
	for _, sessionID := range sessionIDs {
		if session, ok := s.sessions[sessionID]; ok {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// Count 返回总会话数（用于测试和监控）
func (s *MemoryStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// Clear 清空所有会话（用于测试）
func (s *MemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions = make(map[string]*models.Session)
	s.userSessions = make(map[string][]string)
}
