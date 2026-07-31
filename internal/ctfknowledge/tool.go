package ctfknowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Toyar-mo/GzdAgent/internal/mcp"
	"github.com/Toyar-mo/GzdAgent/internal/mcp/builtin"

	"go.uber.org/zap"
)

// RegisterWPKnowledgeTool 注册 CTF WP 知识检索工具到 MCP 服务器
func RegisterWPKnowledgeTool(mcpServer *mcp.Server, manager *Manager, logger *zap.Logger) {
	registerSearchCTFWP(mcpServer, manager, logger)
	registerWPStats(mcpServer, manager, logger)
	logger.Info("CTF WP 知识库工具注册完成")
}

func registerSearchCTFWP(mcpServer *mcp.Server, manager *Manager, logger *zap.Logger) {
	tool := mcp.Tool{
		Name:             builtin.ToolSearchCTFWP,
		Description:      "在 CTF 解题知识库中搜索历史 Write-up（优秀选手的解题报告），获取类似题目的攻击思路、Payload 和工具使用方式。输入当前 CTF 题目的类别和关键词，返回相关的历史 WP 供参考。注意：WP 仅作辅助决策，不要直接复制答案，需要结合当前环境重新验证。",
		ShortDescription: "搜索 CTF 历史解题思路（Write-up 知识库）",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"category": map[string]interface{}{
					"type":        "string",
					"description": "题目类别，可选值：web, pwn, reverse, crypto, misc，或者细分类别如 web/sql_injection",
				},
				"keywords": map[string]interface{}{
					"type":        "array",
					"description": "关键词列表，如 [\"sql\", \"bypass\", \"upload\"]，将与题目类型匹配",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "当前题目的自然语言描述，如「一个登录页面存在SQL注入，需要绕过WAF」",
				},
			},
			"required": []string{},
		},
	}

	handler := func(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
		category, _ := args["category"].(string)
		query, _ := args["query"].(string)

		var keywords []string
		if kw, ok := args["keywords"].([]interface{}); ok {
			for _, k := range kw {
				if s, ok := k.(string); ok && s != "" {
					keywords = append(keywords, s)
				}
			}
		}

		req := &SearchRequest{
			Category: category,
			Keywords: keywords,
			Query:    query,
			TopK:     5,
		}

		results := manager.Search(req)
		if len(results) == 0 {
			return &mcp.ToolResult{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: fmt.Sprintf("未找到与当前题目匹配的历史 WP。建议：\n1. 尝试更换关键词\n2. 使用更通用的类别\n3. 在 wp/ 目录中添加更多 Write-up"),
					},
				},
			}, nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("在 CTF 知识库中找到 %d 条相关历史经验：\n\n", len(results)))

		for i, r := range results {
			entry := r.Entry
			sb.WriteString(fmt.Sprintf("=== WP %d（相关度: %.0f%%）===\n", i+1, r.Similarity*100))
			sb.WriteString(fmt.Sprintf("题目: %s\n", entry.Title))
			if entry.Source != "" {
				sb.WriteString(fmt.Sprintf("来源: %s\n", entry.Source))
			}
			if entry.VulnType != "" {
				sb.WriteString(fmt.Sprintf("类型: %s\n", entry.VulnType))
			}
			if entry.Category != "" {
				sb.WriteString(fmt.Sprintf("分类: %s", entry.Category))
				if entry.SubCategory != "" {
					sb.WriteString("/" + entry.SubCategory)
				}
				sb.WriteString("\n")
			}

			// 题目分析摘要（前 200 字）
			if entry.Analysis != "" {
				analysis := truncateText(entry.Analysis, 200)
				sb.WriteString(fmt.Sprintf("题目分析: %s\n", analysis))
			}

			// 解题思路摘要
			if entry.Solution != "" {
				sb.WriteString(fmt.Sprintf("解题思路: %s\n", truncateText(entry.Solution, 300)))
			}

			// 使用工具
			if len(entry.Tools) > 0 {
				sb.WriteString(fmt.Sprintf("使用工具: %s\n", strings.Join(entry.Tools, ", ")))
			}

			// 关键词
			if len(entry.Keywords) > 0 {
				sb.WriteString(fmt.Sprintf("关键词: %s\n", strings.Join(entry.Keywords, ", ")))
			}

			// Payload（如果有且长度合理）
			if entry.Payload != "" && len(entry.Payload) < 500 {
				sb.WriteString(fmt.Sprintf("参考 Payload:\n%s\n", entry.Payload))
			} else if entry.Payload != "" {
				sb.WriteString(fmt.Sprintf("参考 Payload:（内容较长，请参考原始 WP 文件 %s）\n", entry.FilePath))
			}

			sb.WriteString("\n")
		}

		// 安全提示
		sb.WriteString("---\n")
		sb.WriteString("⚠️ 安全提醒：WP 仅作参考，请务必：\n")
		sb.WriteString("1. 根据当前靶机环境重新验证 Payload\n")
		sb.WriteString("2. 不要直接复制 WP 中的 Flag\n")
		sb.WriteString("3. 每种攻击方法需要先确认环境兼容性\n")

		return &mcp.ToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: sb.String(),
				},
			},
		}, nil
	}

	mcpServer.RegisterTool(tool, handler)
	logger.Debug("CTF WP 搜索工具已注册", zap.String("toolName", tool.Name))
}

func registerWPStats(mcpServer *mcp.Server, manager *Manager, logger *zap.Logger) {
	tool := mcp.Tool{
		Name:             builtin.ToolCTFWPStats,
		Description:      "查看 CTF 解题知识库（Write-up 库）的状态信息，包括总条目数、分类分布等。",
		ShortDescription: "查看 CTF 知识库状态",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
	}

	handler := func(ctx context.Context, args map[string]interface{}) (*mcp.ToolResult, error) {
		stats := manager.GetStats()
		if stats == nil {
			return &mcp.ToolResult{
				Content: []mcp.Content{
					{
						Type: "text",
						Text: "CTF 知识库中暂无内容。请先在 wp/ 目录中添加 Write-up 文件。",
					},
				},
			}, nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("CTF 知识库状态\n"))
		sb.WriteString(fmt.Sprintf("总条目数: %d\n\n", stats.TotalEntries))

		if len(stats.Categories) > 0 {
			sb.WriteString("分类分布:\n")
			// Sort categories
			catNames := make([]string, 0, len(stats.Categories))
			for name := range stats.Categories {
				catNames = append(catNames, name)
			}
			sort.Strings(catNames)
			for _, name := range catNames {
				sb.WriteString(fmt.Sprintf("  %s: %d 篇\n", name, stats.Categories[name]))
			}
		}

		return &mcp.ToolResult{
			Content: []mcp.Content{
				{
					Type: "text",
					Text: sb.String(),
				},
			},
		}, nil
	}

	mcpServer.RegisterTool(tool, handler)
	logger.Debug("CTF WP 状态工具已注册", zap.String("toolName", tool.Name))
}

func truncateText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

// FormatSearchResultJSON 格式化搜索结果（用于 CLI）
func FormatSearchResultJSON(results []*SearchResult) string {
	type resultItem struct {
		Title      string   `json:"title"`
		Category   string   `json:"category"`
		SubCat     string   `json:"subCategory"`
		Source     string   `json:"source"`
		VulnType   string   `json:"vulnType"`
		Similarity float64  `json:"similarity"`
		Keywords   []string `json:"keywords"`
		Tools      []string `json:"tools"`
		Summary    string   `json:"summary"`
	}

	items := make([]resultItem, 0, len(results))
	for _, r := range results {
		items = append(items, resultItem{
			Title:      r.Entry.Title,
			Category:   r.Entry.Category,
			SubCat:     r.Entry.SubCategory,
			Source:     r.Entry.Source,
			VulnType:   r.Entry.VulnType,
			Similarity: r.Similarity,
			Keywords:   r.Entry.Keywords,
			Tools:      r.Entry.Tools,
			Summary:    r.Entry.Summary,
		})
	}

	data, _ := json.MarshalIndent(items, "", "  ")
	return string(data)
}
