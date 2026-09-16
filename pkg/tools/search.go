package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"agent-runtime/pkg/models"
)

// SearchTool 搜索工具（Mock 实现）
type SearchTool struct {
	BaseTool
	validator *Validator
}

// NewSearchTool 创建搜索工具
func NewSearchTool() *SearchTool {
	return &SearchTool{
		BaseTool: BaseTool{
			name:        "search",
			description: "Search for information on the internet. Use this tool when you need to find current information or facts.",
			parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
					"max_results": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 5)",
						"default":     5,
					},
				},
				"required": []string{"query"},
			},
		},
		validator: NewValidator(),
	}
}

// Execute 执行搜索（Mock 实现）
func (t *SearchTool) Execute(ctx context.Context, args map[string]interface{}) (*models.ToolResult, error) {
	// 验证必需参数
	if err := t.validator.ValidateRequired(args, []string{"query"}); err != nil {
		return &models.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// 获取参数
	query, err := GetString(args, "query")
	if err != nil {
		return &models.ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	if strings.TrimSpace(query) == "" {
		return &models.ToolResult{
			Success: false,
			Error:   "query cannot be empty",
		}, nil
	}

	// 获取最大结果数（可选参数）
	maxResults := 5
	if _, exists := args["max_results"]; exists {
		if mr, err := GetInt(args, "max_results"); err == nil {
			maxResults = mr
		}
	}

	if maxResults < 1 {
		maxResults = 1
	}
	if maxResults > 10 {
		maxResults = 10
	}

	// Mock 搜索结果
	results := t.mockSearch(query, maxResults)

	return &models.ToolResult{
		Success: true,
		Output:  results,
	}, nil
}

// mockSearch 生成 Mock 搜索结果
func (t *SearchTool) mockSearch(query string, maxResults int) string {
	// 简单的 Mock 数据
	mockData := map[string][]string{
		"weather": {
			"Weather.com - Current weather conditions show sunny skies with temperature of 22°C",
			"AccuWeather - 7-day forecast predicts clear weather throughout the week",
			"Local Weather Station - Live updates showing humidity at 65%",
		},
		"news": {
			"Breaking News - Latest developments in technology sector",
			"World News Today - International updates from around the globe",
			"Tech News Daily - New AI breakthroughs announced by research teams",
		},
		"default": {
			fmt.Sprintf("Search result 1 for '%s' - Relevant information found at example.com", query),
			fmt.Sprintf("Search result 2 for '%s' - Detailed article at source.org", query),
			fmt.Sprintf("Search result 3 for '%s' - Latest updates at news.net", query),
		},
	}

	// 根据查询选择合适的 Mock 数据
	var results []string
	queryLower := strings.ToLower(query)
	if strings.Contains(queryLower, "weather") || strings.Contains(queryLower, "天气") {
		results = mockData["weather"]
	} else if strings.Contains(queryLower, "news") || strings.Contains(queryLower, "新闻") {
		results = mockData["news"]
	} else {
		results = mockData["default"]
	}

	// 限制结果数量
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	// 格式化输出
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d results for query '%s':\n\n", len(results), query))
	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, result))
	}
	output.WriteString(fmt.Sprintf("\n(Mock search performed at %s)", time.Now().Format("2006-01-02 15:04:05")))

	return output.String()
}
