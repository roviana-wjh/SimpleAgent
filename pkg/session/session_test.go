package session_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"agent-runtime/pkg/models"
	"agent-runtime/pkg/session"
)

func TestMemoryStore(t *testing.T) {
	store := session.NewMemoryStore()
	ctx := context.Background()

	// 测试保存和获取
	t.Run("Save and Get", func(t *testing.T) {
		sess := &models.Session{
			ID:        "session-1",
			UserID:    "user-1",
			Messages:  []*models.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Metadata:  make(map[string]interface{}),
		}

		err := store.Save(ctx, sess)
		if err != nil {
			t.Fatalf("failed to save session: %v", err)
		}

		retrieved, err := store.Get(ctx, "session-1")
		if err != nil {
			t.Fatalf("failed to get session: %v", err)
		}

		if retrieved.ID != sess.ID {
			t.Errorf("expected ID %s, got %s", sess.ID, retrieved.ID)
		}
		if retrieved.UserID != sess.UserID {
			t.Errorf("expected UserID %s, got %s", sess.UserID, retrieved.UserID)
		}
	})

	// 测试按用户列出
	t.Run("ListByUser", func(t *testing.T) {
		store.Clear()

		// 为同一用户创建多个会话
		for i := 1; i <= 3; i++ {
			sess := &models.Session{
				ID:        fmt.Sprintf("session-%d", i),
				UserID:    "user-1",
				Messages:  []*models.Message{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Metadata:  make(map[string]interface{}),
			}
			store.Save(ctx, sess)
		}

		sessions, err := store.ListByUser(ctx, "user-1")
		if err != nil {
			t.Fatalf("failed to list sessions: %v", err)
		}

		if len(sessions) != 3 {
			t.Errorf("expected 3 sessions, got %d", len(sessions))
		}
	})

	// 测试删除
	t.Run("Delete", func(t *testing.T) {
		store.Clear()

		sess := &models.Session{
			ID:        "session-1",
			UserID:    "user-1",
			Messages:  []*models.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Metadata:  make(map[string]interface{}),
		}
		store.Save(ctx, sess)

		err := store.Delete(ctx, "session-1")
		if err != nil {
			t.Fatalf("failed to delete session: %v", err)
		}

		_, err = store.Get(ctx, "session-1")
		if err == nil {
			t.Error("expected error when getting deleted session")
		}
	})

	// 测试会话隔离
	t.Run("Session Isolation", func(t *testing.T) {
		store.Clear()

		// 创建两个用户的会话
		sess1 := &models.Session{
			ID:        "session-1",
			UserID:    "user-1",
			Messages:  []*models.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Metadata:  make(map[string]interface{}),
		}
		sess2 := &models.Session{
			ID:        "session-2",
			UserID:    "user-2",
			Messages:  []*models.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Metadata:  make(map[string]interface{}),
		}

		store.Save(ctx, sess1)
		store.Save(ctx, sess2)

		// 验证用户 1 只能看到自己的会话
		user1Sessions, _ := store.ListByUser(ctx, "user-1")
		if len(user1Sessions) != 1 || user1Sessions[0].ID != "session-1" {
			t.Error("session isolation failed for user-1")
		}

		// 验证用户 2 只能看到自己的会话
		user2Sessions, _ := store.ListByUser(ctx, "user-2")
		if len(user2Sessions) != 1 || user2Sessions[0].ID != "session-2" {
			t.Error("session isolation failed for user-2")
		}
	})
}

func TestSessionManager(t *testing.T) {
	store := session.NewMemoryStore()
	manager := session.NewManager(store, 10)
	ctx := context.Background()

	// 测试创建会话
	t.Run("Create Session", func(t *testing.T) {
		sess, err := manager.Create(ctx, "user-1")
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		if sess.ID == "" {
			t.Error("session ID is empty")
		}
		if sess.UserID != "user-1" {
			t.Errorf("expected UserID user-1, got %s", sess.UserID)
		}
		if sess.Messages == nil {
			t.Error("messages should be initialized")
		}
	})

	// 测试获取会话上下文
	t.Run("GetOrCreateContext", func(t *testing.T) {
		store.Clear()

		sess, _ := manager.Create(ctx, "user-1")

		sessionCtx, err := manager.GetOrCreateContext(ctx, sess.ID)
		if err != nil {
			t.Fatalf("failed to get context: %v", err)
		}

		if sessionCtx.GetLoopCount() != 0 {
			t.Errorf("expected loop count 0, got %d", sessionCtx.GetLoopCount())
		}
	})

	// 测试保存上下文
	t.Run("SaveContext", func(t *testing.T) {
		store.Clear()

		sess, _ := manager.Create(ctx, "user-1")
		sessionCtx, _ := manager.GetOrCreateContext(ctx, sess.ID)

		// 添加消息
		msg := &models.Message{
			ID:        "msg-1",
			SessionID: sess.ID,
			Role:      "user",
			Content:   "Hello",
			CreatedAt: time.Now(),
		}
		sessionCtx.AddMessage(msg)

		// 保存上下文
		err := manager.SaveContext(ctx, sess.ID, sessionCtx)
		if err != nil {
			t.Fatalf("failed to save context: %v", err)
		}

		// 重新获取并验证
		newCtx, _ := manager.GetOrCreateContext(ctx, sess.ID)
		messages := newCtx.GetMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
		if messages[0].Content != "Hello" {
			t.Errorf("expected content 'Hello', got %s", messages[0].Content)
		}
	})

	// 测试多会话隔离
	t.Run("Multiple Session Isolation", func(t *testing.T) {
		store.Clear()

		// 创建两个会话
		sess1, _ := manager.Create(ctx, "user-1")
		sess2, _ := manager.Create(ctx, "user-1")

		ctx1, _ := manager.GetOrCreateContext(ctx, sess1.ID)
		ctx2, _ := manager.GetOrCreateContext(ctx, sess2.ID)

		// 向会话 1 添加消息
		msg1 := &models.Message{
			ID:        "msg-1",
			SessionID: sess1.ID,
			Role:      "user",
			Content:   "Message in session 1",
			CreatedAt: time.Now(),
		}
		ctx1.AddMessage(msg1)
		manager.SaveContext(ctx, sess1.ID, ctx1)

		// 向会话 2 添加消息
		msg2 := &models.Message{
			ID:        "msg-2",
			SessionID: sess2.ID,
			Role:      "user",
			Content:   "Message in session 2",
			CreatedAt: time.Now(),
		}
		ctx2.AddMessage(msg2)
		manager.SaveContext(ctx, sess2.ID, ctx2)

		// 验证隔离
		reloadCtx1, _ := manager.GetOrCreateContext(ctx, sess1.ID)
		reloadCtx2, _ := manager.GetOrCreateContext(ctx, sess2.ID)

		if len(reloadCtx1.GetMessages()) != 1 {
			t.Error("session 1 should have 1 message")
		}
		if reloadCtx1.GetMessages()[0].Content != "Message in session 1" {
			t.Error("session 1 message mismatch")
		}

		if len(reloadCtx2.GetMessages()) != 1 {
			t.Error("session 2 should have 1 message")
		}
		if reloadCtx2.GetMessages()[0].Content != "Message in session 2" {
			t.Error("session 2 message mismatch")
		}
	})
}

func TestSessionContext(t *testing.T) {
	sess := &models.Session{
		ID:        "session-1",
		UserID:    "user-1",
		Messages:  []*models.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	ctx := session.NewContext(sess, 10)

	// 测试循环计数
	t.Run("Loop Count", func(t *testing.T) {
		if ctx.GetLoopCount() != 0 {
			t.Error("initial loop count should be 0")
		}

		count := ctx.IncrementLoop()
		if count != 1 {
			t.Errorf("expected count 1, got %d", count)
		}

		if ctx.GetLoopCount() != 1 {
			t.Error("loop count should be 1")
		}
	})

	// 测试消息管理
	t.Run("Message Management", func(t *testing.T) {
		msg := &models.Message{
			ID:        "msg-1",
			SessionID: "session-1",
			Role:      "user",
			Content:   "Test",
			CreatedAt: time.Now(),
		}

		err := ctx.AddMessage(msg)
		if err != nil {
			t.Fatalf("failed to add message: %v", err)
		}

		messages := ctx.GetMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
	})

	// 测试重置
	t.Run("Reset", func(t *testing.T) {
		ctx.IncrementLoop()
		ctx.IncrementLoop()

		err := ctx.Reset()
		if err != nil {
			t.Fatalf("failed to reset: %v", err)
		}

		if ctx.GetLoopCount() != 0 {
			t.Error("loop count should be 0 after reset")
		}
	})
}
