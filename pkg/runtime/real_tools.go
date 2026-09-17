package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// ============ 真实工具实现 ============

// RealCalculatorTool 真实的计算器工具
type RealCalculatorTool struct{}

func NewRealCalculatorTool() *RealCalculatorTool {
	return &RealCalculatorTool{}
}

func (t *RealCalculatorTool) Name() string {
	return "calculator"
}

func (t *RealCalculatorTool) Description() string {
	return "执行数学计算。支持基本运算：加(+)、减(-)、乘(*)、除(/)。例如：'2+3'、'10*5'、'100/4'"
}

func (t *RealCalculatorTool) ParamSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"description": "数学表达式，如 '2+3' 或 '10*5'",
			},
		},
		"required": []string{"expression"},
	}
}

func (t *RealCalculatorTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	// 获取表达式
	expr, ok := params["expression"].(string)
	if !ok {
		return "", fmt.Errorf("expression parameter must be a string")
	}

	// 简化表达式（去除空格）
	expr = strings.ReplaceAll(expr, " ", "")

	// 解析并计算
	result, err := evaluateExpression(expr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.2f", result), nil
}

// evaluateExpression 简单的表达式求值（支持 +、-、*、/）
func evaluateExpression(expr string) (float64, error) {
	// 支持的运算符优先级：* / > + -

	// 第一步：处理乘除
	expr, err := processMulDiv(expr)
	if err != nil {
		return 0, err
	}

	// 第二步：处理加减
	result, err := processAddSub(expr)
	if err != nil {
		return 0, err
	}

	return result, nil
}

// processMulDiv 处理乘除运算
func processMulDiv(expr string) (string, error) {
	for {
		// 查找 * 或 /
		mulIdx := strings.Index(expr, "*")
		divIdx := strings.Index(expr, "/")

		if mulIdx == -1 && divIdx == -1 {
			break // 没有乘除运算
		}

		// 找到第一个乘除运算符
		opIdx := mulIdx
		op := "*"
		if divIdx != -1 && (mulIdx == -1 || divIdx < mulIdx) {
			opIdx = divIdx
			op = "/"
		}

		// 提取左右操作数
		left, right, err := extractOperands(expr, opIdx)
		if err != nil {
			return "", err
		}

		leftNum, err := strconv.ParseFloat(left, 64)
		if err != nil {
			return "", fmt.Errorf("invalid number: %s", left)
		}

		rightNum, err := strconv.ParseFloat(right, 64)
		if err != nil {
			return "", fmt.Errorf("invalid number: %s", right)
		}

		// 计算结果
		var result float64
		if op == "*" {
			result = leftNum * rightNum
		} else {
			if rightNum == 0 {
				return "", fmt.Errorf("division by zero")
			}
			result = leftNum / rightNum
		}

		// 替换表达式
		leftStart := opIdx - len(left)
		rightEnd := opIdx + 1 + len(right)
		expr = expr[:leftStart] + fmt.Sprintf("%f", result) + expr[rightEnd:]
	}

	return expr, nil
}

// processAddSub 处理加减运算
func processAddSub(expr string) (float64, error) {
	result := 0.0
	currentNum := ""
	op := "+"

	for i := 0; i < len(expr); i++ {
		c := expr[i]

		if c >= '0' && c <= '9' || c == '.' {
			currentNum += string(c)
		} else if c == '+' || c == '-' {
			if currentNum != "" {
				num, err := strconv.ParseFloat(currentNum, 64)
				if err != nil {
					return 0, fmt.Errorf("invalid number: %s", currentNum)
				}

				if op == "+" {
					result += num
				} else {
					result -= num
				}

				currentNum = ""
			}
			op = string(c)
		}
	}

	// 处理最后一个数字
	if currentNum != "" {
		num, err := strconv.ParseFloat(currentNum, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", currentNum)
		}

		if op == "+" {
			result += num
		} else {
			result -= num
		}
	}

	return result, nil
}

// extractOperands 提取运算符左右的操作数
func extractOperands(expr string, opIdx int) (string, string, error) {
	// 提取左操作数（向左查找数字）
	left := ""
	for i := opIdx - 1; i >= 0; i-- {
		c := expr[i]
		if c >= '0' && c <= '9' || c == '.' {
			left = string(c) + left
		} else {
			break
		}
	}

	if left == "" {
		return "", "", fmt.Errorf("missing left operand")
	}

	// 提取右操作数（向右查找数字）
	right := ""
	for i := opIdx + 1; i < len(expr); i++ {
		c := expr[i]
		if c >= '0' && c <= '9' || c == '.' {
			right += string(c)
		} else {
			break
		}
	}

	if right == "" {
		return "", "", fmt.Errorf("missing right operand")
	}

	return left, right, nil
}

// ============ 真实的搜索工具（使用 DuckDuckGo）============

type RealSearchTool struct{}

func NewRealSearchTool() *RealSearchTool {
	return &RealSearchTool{}
}

func (t *RealSearchTool) Name() string {
	return "search"
}

func (t *RealSearchTool) Description() string {
	return "在互联网上搜索信息。用于查找最新消息、事实、网站等。例如：'Go语言教程'、'2024年世界杯'"
}

func (t *RealSearchTool) ParamSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "搜索关键词",
			},
		},
		"required": []string{"query"},
	}
}

func (t *RealSearchTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	query, ok := params["query"].(string)
	if !ok {
		return "", fmt.Errorf("query parameter must be a string")
	}

	// 调用 DuckDuckGo Instant Answer API
	results, err := searchDuckDuckGo(ctx, query)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	return results, nil
}

// searchDuckDuckGo 调用 DuckDuckGo Instant Answer API（免费，无需 API Key）
func searchDuckDuckGo(ctx context.Context, query string) (string, error) {
	// DuckDuckGo Instant Answer API
	baseURL := "https://api.duckduckgo.com/"
	params := url.Values{}
	params.Add("q", query)
	params.Add("format", "json")
	params.Add("no_html", "1")
	params.Add("skip_disambig", "1")

	apiURL := baseURL + "?" + params.Encode()

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析 JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// 提取结果
	abstractText := ""
	if val, ok := result["AbstractText"].(string); ok && val != "" {
		abstractText = val
	}

	abstractURL := ""
	if val, ok := result["AbstractURL"].(string); ok && val != "" {
		abstractURL = val
	}

	// 提取 Related Topics
	relatedTopics := []string{}
	if topics, ok := result["RelatedTopics"].([]interface{}); ok {
		for i, topic := range topics {
			if i >= 3 {
				break // 只取前 3 个
			}
			if topicMap, ok := topic.(map[string]interface{}); ok {
				if text, ok := topicMap["Text"].(string); ok && text != "" {
					relatedTopics = append(relatedTopics, text)
				}
			}
		}
	}

	// 构建结果
	resultStr := fmt.Sprintf("搜索结果：'%s'\n\n", query)

	if abstractText != "" {
		resultStr += fmt.Sprintf("摘要：%s\n", abstractText)
		if abstractURL != "" {
			resultStr += fmt.Sprintf("来源：%s\n", abstractURL)
		}
		resultStr += "\n"
	}

	if len(relatedTopics) > 0 {
		resultStr += "相关话题：\n"
		for i, topic := range relatedTopics {
			resultStr += fmt.Sprintf("%d. %s\n", i+1, topic)
		}
	}

	if resultStr == fmt.Sprintf("搜索结果：'%s'\n\n", query) {
		// 没有找到结果
		resultStr += "未找到详细信息，建议访问：https://duckduckgo.com/?q=" + url.QueryEscape(query)
	}

	return resultStr, nil
}

// ============ 真实的天气工具（使用 OpenWeatherMap）============

type RealWeatherTool struct{}

func NewRealWeatherTool() *RealWeatherTool {
	return &RealWeatherTool{}
}

func (t *RealWeatherTool) Name() string {
	return "weather"
}

func (t *RealWeatherTool) Description() string {
	return "查询天气信息。可以查询任何城市的当前天气。例如：'北京'、'上海'、'New York'"
}

func (t *RealWeatherTool) ParamSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"city": map[string]interface{}{
				"type":        "string",
				"description": "城市名称",
			},
		},
		"required": []string{"city"},
	}
}

func (t *RealWeatherTool) Execute(ctx context.Context, params map[string]interface{}) (string, error) {
	city, ok := params["city"].(string)
	if !ok {
		return "", fmt.Errorf("city parameter must be a string")
	}

	// 从环境变量获取 API Key
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		// 如果没有 API Key，返回提示信息
		return fmt.Sprintf("天气查询需要 OpenWeatherMap API Key。\n\n请设置环境变量：\nexport OPENWEATHER_API_KEY=\"your-api-key\"\n\n免费注册：https://openweathermap.org/api\n\n城市：%s", city), nil
	}

	// 调用 OpenWeatherMap API
	weather, err := getWeather(ctx, city, apiKey)
	if err != nil {
		return "", fmt.Errorf("weather query failed: %w", err)
	}

	return weather, nil
}

// getWeather 调用 OpenWeatherMap API
func getWeather(ctx context.Context, city, apiKey string) (string, error) {
	// OpenWeatherMap Current Weather API
	baseURL := "https://api.openweathermap.org/data/2.5/weather"
	params := url.Values{}
	params.Add("q", city)
	params.Add("appid", apiKey)
	params.Add("units", "metric") // 摄氏度
	params.Add("lang", "zh_cn")    // 中文描述

	apiURL := baseURL + "?" + params.Encode()

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析 JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// 提取天气信息
	cityName := ""
	if val, ok := result["name"].(string); ok {
		cityName = val
	}

	temp := 0.0
	if main, ok := result["main"].(map[string]interface{}); ok {
		if val, ok := main["temp"].(float64); ok {
			temp = val
		}
	}

	feelsLike := 0.0
	if main, ok := result["main"].(map[string]interface{}); ok {
		if val, ok := main["feels_like"].(float64); ok {
			feelsLike = val
		}
	}

	humidity := 0.0
	if main, ok := result["main"].(map[string]interface{}); ok {
		if val, ok := main["humidity"].(float64); ok {
			humidity = val
		}
	}

	description := ""
	if weather, ok := result["weather"].([]interface{}); ok && len(weather) > 0 {
		if w, ok := weather[0].(map[string]interface{}); ok {
			if val, ok := w["description"].(string); ok {
				description = val
			}
		}
	}

	windSpeed := 0.0
	if wind, ok := result["wind"].(map[string]interface{}); ok {
		if val, ok := wind["speed"].(float64); ok {
			windSpeed = val
		}
	}

	// 构建结果
	resultStr := fmt.Sprintf("%s 当前天气：\n\n", cityName)
	resultStr += fmt.Sprintf("天气状况：%s\n", description)
	resultStr += fmt.Sprintf("温度：%.1f°C（体感 %.1f°C）\n", temp, feelsLike)
	resultStr += fmt.Sprintf("湿度：%.0f%%\n", humidity)
	resultStr += fmt.Sprintf("风速：%.1f m/s\n", windSpeed)

	return resultStr, nil
}
