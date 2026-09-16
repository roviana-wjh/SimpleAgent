package tools

import (
	"context"
	"fmt"
	"math"

	"agent-runtime/pkg/models"
)

// CalculatorTool 计算器工具
type CalculatorTool struct {
	BaseTool
	validator *Validator
}

// NewCalculatorTool 创建计算器工具
func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{
		BaseTool: BaseTool{
			name:        "calculator",
			description: "Perform basic arithmetic operations: add, subtract, multiply, divide. Use this tool when you need to calculate numbers.",
			parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "The arithmetic operation to perform",
						"enum":        []string{"add", "subtract", "multiply", "divide"},
					},
					"a": map[string]interface{}{
						"type":        "number",
						"description": "The first number",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "The second number",
					},
				},
				"required": []string{"operation", "a", "b"},
			},
		},
		validator: NewValidator(),
	}
}

// Execute 执行计算
func (t *CalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (*models.ToolResult, error) {
	// 验证必需参数
	if err := t.validator.ValidateRequired(args, []string{"operation", "a", "b"}); err != nil {
		return &models.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取参数
	operation, err := GetString(args, "operation")
	if err != nil {
		return &models.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	a, err := GetNumber(args, "a")
	if err != nil {
		return &models.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	b, err := GetNumber(args, "b")
	if err != nil {
		return &models.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 执行计算
	var result float64
	switch operation {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return &models.ToolResult{
				Success: false,
				Error:   "division by zero",
			}, nil
		}
		result = a / b
	default:
		return &models.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported operation: %s", operation),
		}, nil
	}

	// 检查结果是否有效
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return &models.ToolResult{
			Success: false,
			Error:   "calculation resulted in invalid number",
		}, nil
	}

	return &models.ToolResult{
		Success: true,
		Output:  fmt.Sprintf("%.2f", result),
	}, nil
}
