package runtime

import (
	"context"
	"strings"
	"testing"
)

// ============ 真实工具测试 ============

func TestRealCalculatorTool(t *testing.T) {
	calc := NewRealCalculatorTool()

	tests := []struct {
		name       string
		expression string
		want       string
		wantErr    bool
	}{
		{
			name:       "简单加法",
			expression: "2+3",
			want:       "5.00",
			wantErr:    false,
		},
		{
			name:       "简单减法",
			expression: "10-3",
			want:       "7.00",
			wantErr:    false,
		},
		{
			name:       "简单乘法",
			expression: "4*5",
			want:       "20.00",
			wantErr:    false,
		},
		{
			name:       "简单除法",
			expression: "20/4",
			want:       "5.00",
			wantErr:    false,
		},
		{
			name:       "运算符优先级",
			expression: "2+3*4",
			want:       "14.00",
			wantErr:    false,
		},
		{
			name:       "复杂表达式",
			expression: "10*5/2",
			want:       "25.00",
			wantErr:    false,
		},
		{
			name:       "小数运算",
			expression: "3.5*2",
			want:       "7.00",
			wantErr:    false,
		},
		{
			name:       "除零错误",
			expression: "10/0",
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"expression": tt.expression,
			}

			result, err := calc.Execute(context.Background(), params)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.want {
				t.Errorf("Expected %s, got %s", tt.want, result)
			}
		})
	}
}

func TestRealSearchTool(t *testing.T) {
	search := NewRealSearchTool()

	// 测试基本功能
	params := map[string]interface{}{
		"query": "Go programming language",
	}

	result, err := search.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证返回包含查询关键词
	if !strings.Contains(result, "Go") || !strings.Contains(result, "搜索结果") {
		t.Errorf("Result should contain query keyword and search indicator")
	}

	t.Logf("Search result: %s", result)
}

func TestRealWeatherTool(t *testing.T) {
	weather := NewRealWeatherTool()

	// 测试基本功能
	params := map[string]interface{}{
		"city": "London",
	}

	result, err := weather.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证返回包含城市名或提示信息
	if !strings.Contains(result, "London") && !strings.Contains(result, "API Key") {
		t.Errorf("Result should contain city name or API key hint")
	}

	t.Logf("Weather result: %s", result)
}

func TestRealToolsIntegration(t *testing.T) {
	// 测试真实工具与 Registry 集成
	registry := NewSimpleToolRegistry()

	// 注册真实工具
	registry.Register(RealCalculator())
	registry.Register(RealSearch())
	registry.Register(RealWeather())

	// 验证注册成功
	tools := registry.GetAll()
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(tools))
	}

	// 测试通过 Registry 调用工具
	calcTool := registry.Get("calculator")
	if calcTool == nil {
		t.Fatal("Calculator tool not found")
	}

	result, err := calcTool.Execute(context.Background(), map[string]interface{}{
		"expression": "5+3",
	})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "8.00" {
		t.Errorf("Expected 8.00, got %s", result)
	}
}

func TestCalculatorErrorHandling(t *testing.T) {
	calc := NewRealCalculatorTool()

	tests := []struct {
		name   string
		params map[string]interface{}
	}{
		{
			name:   "缺少参数",
			params: map[string]interface{}{},
		},
		{
			name: "参数类型错误",
			params: map[string]interface{}{
				"expression": 123, // 应该是 string
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := calc.Execute(context.Background(), tt.params)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}
		})
	}
}
