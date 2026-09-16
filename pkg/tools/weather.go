package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"agent-runtime/pkg/models"
)

// WeatherTool 天气查询工具（Mock 实现）
type WeatherTool struct {
	BaseTool
	validator *Validator
}

// NewWeatherTool 创建天气工具
func NewWeatherTool() *WeatherTool {
	return &WeatherTool{
		BaseTool: BaseTool{
			name:        "weather",
			description: "Get current weather information for a specific location. Use this tool when users ask about weather conditions.",
			parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"location": map[string]interface{}{
						"type":        "string",
						"description": "The city or location name (e.g., 'Beijing', 'New York', 'Tokyo')",
					},
					"unit": map[string]interface{}{
						"type":        "string",
						"description": "Temperature unit",
						"enum":        []string{"celsius", "fahrenheit"},
						"default":     "celsius",
					},
				},
				"required": []string{"location"},
			},
		},
		validator: NewValidator(),
	}
}

// Execute 执行天气查询（Mock 实现）
func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (*models.ToolResult, error) {
	// 验证必需参数
	if err := t.validator.ValidateRequired(args, []string{"location"}); err != nil {
		return &models.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取参数
	location, err := GetString(args, "location")
	if err != nil {
		return &models.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	if strings.TrimSpace(location) == "" {
		return &models.ToolResult{
			Success: false,
			Error:   "location cannot be empty",
		}, nil
	}

	// 获取温度单位（可选参数）
	unit := GetStringWithDefault(args, "unit", "celsius")
	if unit != "celsius" && unit != "fahrenheit" {
		return &models.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("invalid unit: %s (must be 'celsius' or 'fahrenheit')", unit),
		}, nil
	}

	// Mock 天气数据
	weather := t.mockWeather(location, unit)

	return &models.ToolResult{
		Success: true,
		Output:  weather,
	}, nil
}

// mockWeather 生成 Mock 天气数据
func (t *WeatherTool) mockWeather(location, unit string) string {
	// 根据城市生成不同的 Mock 数据
	weatherData := map[string]map[string]interface{}{
		"beijing": {
			"condition":  "Partly Cloudy",
			"celsius":    22,
			"fahrenheit": 72,
			"humidity":   65,
			"wind":       "15 km/h NE",
		},
		"shanghai": {
			"condition":  "Sunny",
			"celsius":    28,
			"fahrenheit": 82,
			"humidity":   70,
			"wind":       "10 km/h SE",
		},
		"new york": {
			"condition":  "Rainy",
			"celsius":    18,
			"fahrenheit": 64,
			"humidity":   80,
			"wind":       "20 km/h W",
		},
		"london": {
			"condition":  "Cloudy",
			"celsius":    15,
			"fahrenheit": 59,
			"humidity":   75,
			"wind":       "12 km/h SW",
		},
		"tokyo": {
			"condition":  "Clear",
			"celsius":    25,
			"fahrenheit": 77,
			"humidity":   60,
			"wind":       "8 km/h E",
		},
	}

	// 标准化城市名称
	locationKey := strings.ToLower(strings.TrimSpace(location))

	// 获取天气数据，如果城市不存在则使用默认值
	data, exists := weatherData[locationKey]
	if !exists {
		data = map[string]interface{}{
			"condition":  "Clear",
			"celsius":    20,
			"fahrenheit": 68,
			"humidity":   55,
			"wind":       "10 km/h N",
		}
	}

	// 根据单位选择温度
	var temperature string
	if unit == "fahrenheit" {
		temperature = fmt.Sprintf("%d°F", data["fahrenheit"].(int))
	} else {
		temperature = fmt.Sprintf("%d°C", data["celsius"].(int))
	}

	// 格式化输出
	output := fmt.Sprintf(`Weather for %s:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🌡️  Temperature: %s
☁️  Condition: %s
💧 Humidity: %d%%
🌬️  Wind: %s
📅 Updated: %s
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
(Mock weather data)`,
		location,
		temperature,
		data["condition"].(string),
		data["humidity"].(int),
		data["wind"].(string),
		time.Now().Format("2006-01-02 15:04:05"),
	)

	return output
}
