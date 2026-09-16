package runtime

import (
	"context"
	"fmt"
	"sync"
)

// Tool 定义工具接口
type Tool interface {
	Name() string
	Description() string
	ParamSchema() map[string]interface{}
	Execute(ctx context.Context, params map[string]interface{}) (string, error)
}

// ToolRegistry 定义工具注册表接口
type ToolRegistry interface {
	Register(tool Tool) error
	Get(name string) Tool
	GetAll() []Tool
}

// SimpleToolRegistry 简单的工具注册表实现
type SimpleToolRegistry struct {
	tools map[string]Tool
	mu    sync.RWMutex
}

// NewSimpleToolRegistry 创建工具注册表
func NewSimpleToolRegistry() *SimpleToolRegistry {
	return &SimpleToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *SimpleToolRegistry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name()]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name())
	}

	r.tools[tool.Name()] = tool
	return nil
}

// Get 获取工具
func (r *SimpleToolRegistry) Get(name string) Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tools[name]
}

// GetAll 获取所有工具
func (r *SimpleToolRegistry) GetAll() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ============ Fake Tool 实现 ============

// FakeTool 用于测试的 Fake 工具
type FakeTool struct {
	name        string
	description string
	output      string
	shouldFail  bool
	errorMsg    string
	callCount   int
	mu          sync.Mutex
}

// NewFakeTool 创建成功的 Fake 工具
func NewFakeTool(name string, output string) *FakeTool {
	return &FakeTool{
		name:        name,
		description: "Fake tool for testing: " + name,
		output:      output,
		callCount:   0,
		shouldFail:  false,
	}
}

// NewFailingTool 创建会失败的 Fake 工具
func NewFailingTool(name string, errorMsg string) *FakeTool {
	return &FakeTool{
		name:        name,
		description: "Fake tool for testing: " + name,
		shouldFail:  true,
		errorMsg:    errorMsg,
		callCount:   0,
	}
}

// Name 返回工具名称
func (t *FakeTool) Name() string {
	return t.name
}

// Description 返回工具描述
func (t *FakeTool) Description() string {
	return t.description
}

// ParamSchema 返回参数 Schema
func (t *FakeTool) ParamSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type": "string",
			},
		},
	}
}

// Execute 执行工具
func (t *FakeTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.callCount++

	if t.shouldFail {
		return "", fmt.Errorf(t.errorMsg)
	}

	return t.output, nil
}

// GetCallCount 获取调用次数
func (t *FakeTool) GetCallCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.callCount
}

// Reset 重置调用次数
func (t *FakeTool) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.callCount = 0
}

// ============ 具体的 Fake Tool 实现 ============

// CalculatorTool Fake 计算器工具
func CalculatorTool() *FakeTool {
	return NewFakeTool("calculator", "4")
}

// SearchTool Fake 搜索工具
func SearchTool() *FakeTool {
	return NewFakeTool("search", "Found results: ...")
}

// WeatherTool Fake 天气工具
func WeatherTool() *FakeTool {
	return NewFakeTool("weather", "Sunny, 25°C")
}
