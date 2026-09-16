package session

import (
	"fmt"

	"agent-runtime/pkg/models"
)

// Context 表示会话上下文（实现 runtime.SessionContext 接口）
type Context struct {
	session   *models.Session
	loopCount int
	maxLoops  int
}

// NewContext 创建新的会话上下文
func NewContext(session *models.Session, maxLoops int) *Context {
	if maxLoops <= 0 {
		maxLoops = 10
	}
	return &Context{
		session:   session,
		loopCount: 0,
		maxLoops:  maxLoops,
	}
}

// GetMessages 获取所有消息
func (c *Context) GetMessages() []*models.Message {
	return c.session.GetMessages()
}

// AddMessage 添加消息
func (c *Context) AddMessage(msg *models.Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	c.session.AddMessage(msg)
	return nil
}

// GetLoopCount 获取当前循环计数
func (c *Context) GetLoopCount() int {
	return c.loopCount
}

// IncrementLoop 增加循环计数并返回新值
func (c *Context) IncrementLoop() int {
	c.loopCount++
	return c.loopCount
}

// GetSession 获取底层 Session 对象
func (c *Context) GetSession() *models.Session {
	return c.session
}

// Reset 重置上下文（清空循环计数）
func (c *Context) Reset() error {
	c.loopCount = 0
	return nil
}
