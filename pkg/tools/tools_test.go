package tools_test

import (
	"context"
	"testing"

	"agent-runtime/pkg/tools"
)

func TestRegistry(t *testing.T) {
	registry := tools.NewRegistry()

	// 测试注册工具
	calc := tools.NewCalculatorTool()
	if err := registry.Register(calc); err != nil {
		t.Fatalf("failed to register calculator: %v", err)
	}

	// 测试重复注册
	if err := registry.Register(calc); err == nil {
		t.Fatal("expected error when registering duplicate tool")
	}

	// 测试获取工具
	tool, err := registry.Get("calculator")
	if err != nil {
		t.Fatalf("failed to get calculator: %v", err)
	}
	if tool.Name() != "calculator" {
		t.Errorf("expected calculator, got %s", tool.Name())
	}

	// 测试获取不存在的工具
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error when getting nonexistent tool")
	}

	// 测试列出所有工具
	tools := registry.List()
	if len(tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(tools))
	}

	// 测试获取所有定义
	definitions := registry.GetAllDefinitions()
	if len(definitions) != 1 {
		t.Errorf("expected 1 definition, got %d", len(definitions))
	}
	if definitions[0].Function.Name != "calculator" {
		t.Errorf("expected calculator, got %s", definitions[0].Function.Name)
	}
}

func TestCalculatorTool(t *testing.T) {
	calc := tools.NewCalculatorTool()
	ctx := context.Background()

	tests := []struct {
		name      string
		args      map[string]interface{}
		wantOK    bool
		wantValue string
	}{
		{
			name:      "add",
			args:      map[string]interface{}{"operation": "add", "a": 10.0, "b": 5.0},
			wantOK:    true,
			wantValue: "15.00",
		},
		{
			name:      "subtract",
			args:      map[string]interface{}{"operation": "subtract", "a": 10.0, "b": 5.0},
			wantOK:    true,
			wantValue: "5.00",
		},
		{
			name:      "multiply",
			args:      map[string]interface{}{"operation": "multiply", "a": 10.0, "b": 5.0},
			wantOK:    true,
			wantValue: "50.00",
		},
		{
			name:      "divide",
			args:      map[string]interface{}{"operation": "divide", "a": 10.0, "b": 5.0},
			wantOK:    true,
			wantValue: "2.00",
		},
		{
			name:   "divide by zero",
			args:   map[string]interface{}{"operation": "divide", "a": 10.0, "b": 0.0},
			wantOK: false,
		},
		{
			name:   "missing parameter",
			args:   map[string]interface{}{"operation": "add", "a": 10.0},
			wantOK: false,
		},
		{
			name:   "invalid operation",
			args:   map[string]interface{}{"operation": "modulo", "a": 10.0, "b": 5.0},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.Execute(ctx, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Success != tt.wantOK {
				t.Errorf("expected success=%v, got %v (error: %s)", tt.wantOK, result.Success, result.Error)
			}

			if tt.wantOK && result.Output != tt.wantValue {
				t.Errorf("expected output=%s, got %s", tt.wantValue, result.Output)
			}
		})
	}
}

func TestSearchTool(t *testing.T) {
	search := tools.NewSearchTool()
	ctx := context.Background()

	tests := []struct {
		name   string
		args   map[string]interface{}
		wantOK bool
	}{
		{
			name:   "basic search",
			args:   map[string]interface{}{"query": "golang tutorials"},
			wantOK: true,
		},
		{
			name:   "search with max_results",
			args:   map[string]interface{}{"query": "weather", "max_results": 3},
			wantOK: true,
		},
		{
			name:   "empty query",
			args:   map[string]interface{}{"query": ""},
			wantOK: false,
		},
		{
			name:   "missing query",
			args:   map[string]interface{}{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := search.Execute(ctx, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Success != tt.wantOK {
				t.Errorf("expected success=%v, got %v (error: %s)", tt.wantOK, result.Success, result.Error)
			}

			if tt.wantOK && result.Output == "" {
				t.Error("expected non-empty output")
			}
		})
	}
}

func TestWeatherTool(t *testing.T) {
	weather := tools.NewWeatherTool()
	ctx := context.Background()

	tests := []struct {
		name   string
		args   map[string]interface{}
		wantOK bool
	}{
		{
			name:   "celsius (default)",
			args:   map[string]interface{}{"location": "Beijing"},
			wantOK: true,
		},
		{
			name:   "fahrenheit",
			args:   map[string]interface{}{"location": "New York", "unit": "fahrenheit"},
			wantOK: true,
		},
		{
			name:   "unknown location",
			args:   map[string]interface{}{"location": "UnknownCity"},
			wantOK: true,
		},
		{
			name:   "empty location",
			args:   map[string]interface{}{"location": ""},
			wantOK: false,
		},
		{
			name:   "invalid unit",
			args:   map[string]interface{}{"location": "Tokyo", "unit": "kelvin"},
			wantOK: false,
		},
		{
			name:   "missing location",
			args:   map[string]interface{}{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := weather.Execute(ctx, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Success != tt.wantOK {
				t.Errorf("expected success=%v, got %v (error: %s)", tt.wantOK, result.Success, result.Error)
			}

			if tt.wantOK && result.Output == "" {
				t.Error("expected non-empty output")
			}
		})
	}
}

func TestToolDefinitions(t *testing.T) {
	// 测试所有工具的定义格式
	testTools := []tools.Tool{
		tools.NewCalculatorTool(),
		tools.NewSearchTool(),
		tools.NewWeatherTool(),
	}

	for _, tool := range testTools {
		t.Run(tool.Name(), func(t *testing.T) {
			// 检查基本信息
			if tool.Name() == "" {
				t.Error("tool name is empty")
			}
			if tool.Description() == "" {
				t.Error("tool description is empty")
			}

			// 检查参数 Schema
			params := tool.Parameters()
			if params == nil {
				t.Fatal("parameters is nil")
			}

			// 检查 Schema 结构
			if params["type"] != "object" {
				t.Errorf("expected type=object, got %v", params["type"])
			}

			properties, ok := params["properties"].(map[string]interface{})
			if !ok {
				t.Fatal("properties is not a map")
			}
			if len(properties) == 0 {
				t.Error("properties is empty")
			}

			required, ok := params["required"].([]string)
			if !ok {
				t.Fatal("required is not a string slice")
			}
			if len(required) == 0 {
				t.Error("required parameters is empty")
			}

			// 测试转换为 ToolDefinition
			def := tools.ToDefinition(tool)
			if def.Type != "function" {
				t.Errorf("expected type=function, got %s", def.Type)
			}
			if def.Function.Name != tool.Name() {
				t.Errorf("expected name=%s, got %s", tool.Name(), def.Function.Name)
			}
		})
	}
}
