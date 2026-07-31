package ctfknowledge

import (
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// section extracts a section from markdown content by heading name.
// Returns the section content (lines between heading and next heading).
func extractSection(content string, heading string) string {
	idx := strings.Index(content, heading)
	if idx < 0 {
		return ""
	}
	start := idx + len(heading)
	for start < len(content) && content[start] == '\n' {
		start++
	}

	// Count heading level
	level := 0
	for _, c := range heading {
		if c == '#' {
			level++
		} else {
			break
		}
	}

	// Find next heading at the same or higher level (fewer or equal #)
	end := len(content)
	lines := strings.Split(content[start:], "\n")
	currentPos := start
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			currentPos += len(line) + 1
			continue
		}
		// Check if line is a heading
		headLevel := 0
		for _, c := range trimmed {
			if c == '#' {
				headLevel++
			} else {
				break
			}
		}
		if headLevel > 0 && headLevel <= level && strings.TrimSpace(trimmed[headLevel:]) != "" {
			end = currentPos
			break
		}
		currentPos += len(line) + 1
	}

	text := strings.TrimSpace(content[start:end])
	return text
}

// parseKeywords 从"关键词："行提取关键词列表
func parseKeywords(text string) []string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 匹配 "关键词：" 或 "keywords:"
		if strings.HasPrefix(line, "关键词") || strings.HasPrefix(line, "Keywords") || strings.HasPrefix(line, "keywords") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				raw := strings.TrimSpace(parts[1])
				items := strings.Split(raw, ",")
				var result []string
				for _, item := range items {
					item = strings.TrimSpace(item)
					if item != "" {
						result = append(result, item)
					}
				}
				return result
			}
		}
	}
	return nil
}

// extractToolList 从"使用工具"章节提取工具列表（每行一个工具）
func extractToolList(toolsText string) []string {
	if toolsText == "" {
		return nil
	}
	lines := strings.Split(toolsText, "\n")
	var tools []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			tool := strings.TrimSpace(line[2:])
			if tool != "" {
				tools = append(tools, tool)
			}
		} else if strings.HasPrefix(line, "* ") {
			tool := strings.TrimSpace(line[2:])
			if tool != "" {
				tools = append(tools, tool)
			}
		}
	}
	return tools
}

// ParseWP 解析一篇 WP Markdown 内容
func ParseWP(filePath string, content string) *WPEntry {
	entry := &WPEntry{
		FilePath: filePath,
		Content:  content,
		Tools:    nil,
		Keywords: nil,
	}

	// 提取标题（# 开头第一行）
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			entry.Title = strings.TrimSpace(line[2:])
			break
		}
	}

	// 提取各章节
	entry.Source = extractSection(content, "## 来源")
	entry.VulnType = extractSection(content, "## 类型")
	entry.Analysis = extractSection(content, "## 题目分析")
	entry.Solution = extractSection(content, "## 解题思路")
	toolsText := extractSection(content, "## 使用工具")
	entry.Payload = extractSection(content, "## Payload")
	entry.FlagMethod = extractSection(content, "## Flag获取方式")
	summaryText := extractSection(content, "## 技术总结")

	entry.Summary = summaryText
	entry.Keywords = parseKeywords(summaryText)

	if toolsText != "" {
		entry.Tools = extractToolList(toolsText)
	}

	// 从来源提取比赛名称和年份
	if entry.Source != "" {
		for _, line := range strings.Split(entry.Source, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "CTF比赛") || strings.HasPrefix(line, "CTF比赛:") || strings.HasPrefix(line, "CTF比赛：") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					entry.Source = strings.TrimSpace(parts[1])
				} else {
					parts = strings.SplitN(line, "：", 2)
					if len(parts) == 2 {
						entry.Source = strings.TrimSpace(parts[1])
					}
				}
			}
			if strings.HasPrefix(line, "年份") || strings.HasPrefix(line, "年份:") || strings.HasPrefix(line, "年份：") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					entry.Year = strings.TrimSpace(parts[1])
				} else {
					parts = strings.SplitN(line, "：", 2)
					if len(parts) == 2 {
						entry.Year = strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}

	// 从目录路径推断分类
	relParts := strings.Split(filepath.ToSlash(filePath), "/")
	for i, part := range relParts {
		if part == "wp" && i+1 < len(relParts) {
			if i+2 < len(relParts) {
				entry.Category = relParts[i+1]
				if strings.HasSuffix(entry.Category, ".md") {
					entry.Category = ""
				} else if i+2 < len(relParts)-1 || (i+2 == len(relParts)-1 && !strings.HasSuffix(relParts[i+2], ".md")) {
					entry.SubCategory = relParts[i+2]
				}
			}
			break
		}
	}

	// 如果 title 为空，用文件名
	if entry.Title == "" {
		entry.Title = strings.TrimSuffix(filepath.Base(filePath), ".md")
	}

	return entry
}

// ScanWPDirectory 扫描 WP 目录，返回所有解析后的 WPEntry
func ScanWPDirectory(wpDir string, logger *zap.Logger) ([]*WPEntry, error) {
	info, err := os.Stat(wpDir)
	if err != nil || !info.IsDir() {
		logger.Warn("WP 目录不存在或不可读，跳过", zap.String("path", wpDir), zap.Error(err))
		return nil, nil
	}

	var entries []*WPEntry
	err = filepath.WalkDir(wpDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), "readme.md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			logger.Warn("读取 WP 文件失败", zap.String("path", path), zap.Error(err))
			return nil
		}

		entry := ParseWP(path, string(content))
		entries = append(entries, entry)
		logger.Debug("扫描到 WP", zap.String("title", entry.Title), zap.String("category", entry.Category))
		return nil
	})

	if err != nil {
		logger.Warn("扫描 WP 目录失败", zap.String("path", wpDir), zap.Error(err))
		return nil, nil
	}

	logger.Info("WP 知识库扫描完成", zap.Int("count", len(entries)))
	return entries, nil
}
