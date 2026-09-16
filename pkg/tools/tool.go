package tools

import (
	"context"
	"fmt"

	"agent-runtime/pkg/models"
)

// Tool 定义工具接口
type Tool interface {
	// Name 返回工具名称
	Name() string

	// Description 返回工具描述（供 LLM 理解）
	Description() string

	// Parameters 返回参数 Schema（JSON Schema 格式）
	Parameters() map[string]interface{}

	// Execute 执行工具
	Execute(ctx context.Context, args map[string]interface{}) (*models.ToolResult, error)
}

// BaseTool 提供 Tool 的基础实现
type BaseTool struct {
	name        string
	description string
	parameters  map[string]interface{}
}

// Name 实现 Tool 接口
func (t *BaseTool) Name() string {
	return t.name
}

// Description 实现 Tool 接口
func (t *BaseTool) Description() string {
	return t.description
}

// Parameters 实现 Tool 接口
func (t *BaseTool) Parameters() map[string]interface{} {
	return t.parameters
}

// ToDefinition 将 Tool 转换为 LLM API 所需的 ToolDefinition
func ToDefinition(tool Tool) *models.ToolDefinition {
	return &models.ToolDefinition{
		Type: "function",
		Function: models.FunctionDef{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters:  tool.Parameters(),
		},
	}
}

// ToolError 表示工具执行错误
type ToolError struct {
	ToolName string
	Reason   string
	Details  string
}

func (e *ToolError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("tool %s failed: %s - %s", e.ToolName, e.Reason, e.Details)
	}
	return fmt.Sprintf("tool %s failed: %s", e.ToolName, e.Reason)
}

// NewToolError 创建工具错误
func NewToolError(toolName, reason, details string) *ToolError {
	return &ToolError{
		ToolName: toolName,
		Reason:   reason,
		Details:  details,
	}
}
