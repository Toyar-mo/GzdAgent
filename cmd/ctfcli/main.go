package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Toyar-mo/GzdAgent/internal/ctfknowledge"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	wpDir := flag.String("wp", "wp", "WP 知识库目录路径")
	flag.CommandLine.Parse(os.Args[2:])

	cmd := os.Args[1]
	absWpDir, _ := filepath.Abs(*wpDir)

	switch cmd {
	case "init":
		cmdInit(absWpDir, logger)
	case "index":
		cmdIndex(absWpDir, logger)
	case "search":
		query := strings.Join(os.Args[2:], " ")
		if query == "" || query == "--wp" {
			// Try to parse remaining args
			remaining := flag.Args()
			query = strings.Join(remaining, " ")
		}
		if query == "" {
			fmt.Println("请提供搜索关键词")
			return
		}
		cmdSearch(absWpDir, query, logger)
	case "status":
		cmdStatus(absWpDir, logger)
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Print(`CTF WP 知识库管理工具

用法:
  ctfcli init                    初始化 WP 知识库目录结构
  ctfcli index                   扫描并索引 WP 目录
  ctfcli search <关键词>          搜索 WP 知识库
  ctfcli status                  查看知识库状态

选项:
  --wp <路径>    WP 知识库目录路径（默认: ./wp）
`)
}

func cmdInit(wpDir string, logger *zap.Logger) {
	dirs := []string{
		wpDir,
		filepath.Join(wpDir, "web", "sql_injection"),
		filepath.Join(wpDir, "web", "xss"),
		filepath.Join(wpDir, "web", "ssrf"),
		filepath.Join(wpDir, "web", "file_upload"),
		filepath.Join(wpDir, "web", "rce"),
		filepath.Join(wpDir, "web", "other"),
		filepath.Join(wpDir, "pwn"),
		filepath.Join(wpDir, "reverse"),
		filepath.Join(wpDir, "crypto"),
		filepath.Join(wpDir, "misc"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录失败 %s: %v\n", dir, err)
			return
		}
	}

	readmePath := filepath.Join(wpDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		readme := `# CTF Write-up 知识库

本目录存储 CTF 比赛历史优秀选手的 Write-up（WP）解题报告。

## 目录结构

- web/sql_injection/ - SQL 注入
- web/xss/ - 跨站脚本
- web/ssrf/ - 服务端请求伪造
- web/file_upload/ - 文件上传
- web/rce/ - 远程命令执行
- web/other/ - 其他 Web
- pwn/ - 二进制漏洞利用
- reverse/ - 逆向工程
- crypto/ - 密码学
- misc/ - 杂项

## WP 格式

参考 web/sql_injection/ 中的示例文件。
`
		os.WriteFile(readmePath, []byte(readme), 0644)
	}

	fmt.Printf("✅ WP 知识库目录已初始化: %s\n", wpDir)
	fmt.Printf("总目录数: %d\n", len(dirs))
	fmt.Println("\n请将 CTF Write-up Markdown 文件放入对应分类目录，然后运行: ctfcli index")
}

func cmdIndex(wpDir string, logger *zap.Logger) {
	manager := ctfknowledge.NewManager(wpDir, logger)
	if err := manager.BuildIndex(); err != nil {
		fmt.Printf("❌ 索引失败: %v\n", err)
		return
	}

	stats := manager.GetStats()
	fmt.Printf("✅ WP 知识库索引完成\n")
	fmt.Printf("总条目数: %d\n", stats.TotalEntries)
	for cat, count := range stats.Categories {
		fmt.Printf("  %s: %d 篇\n", cat, count)
	}
}

func cmdSearch(wpDir string, query string, logger *zap.Logger) {
	manager := ctfknowledge.NewManager(wpDir, logger)
	if err := manager.BuildIndex(); err != nil {
		fmt.Printf("❌ 索引失败: %v\n", err)
		return
	}

	req := &ctfknowledge.SearchRequest{
		Query: query,
		TopK:  10,
	}
	results := manager.Search(req)

	if len(results) == 0 {
		fmt.Println("未找到匹配结果")
		return
	}

	jsonOut := os.Getenv("JSON")
	if jsonOut == "1" || jsonOut == "true" {
		output := ctfknowledge.FormatSearchResultJSON(results)
		fmt.Println(output)
		return
	}

	fmt.Printf("找到 %d 条匹配结果:\n\n", len(results))
	for i, r := range results {
		entry := r.Entry
		fmt.Printf("=== %d. %s（相关度: %.0f%%）===\n", i+1, entry.Title, r.Similarity*100)
		if entry.Category != "" {
			fmt.Printf("   分类: %s", entry.Category)
			if entry.SubCategory != "" {
				fmt.Printf("/%s", entry.SubCategory)
			}
			fmt.Println()
		}
		if entry.VulnType != "" {
			fmt.Printf("   类型: %s\n", entry.VulnType)
		}
		if len(entry.Keywords) > 0 {
			fmt.Printf("   关键词: %s\n", strings.Join(entry.Keywords, ", "))
		}
		if len(entry.Tools) > 0 {
			fmt.Printf("   工具: %s\n", strings.Join(entry.Tools, ", "))
		}
		fmt.Println()
	}
}

func cmdStatus(wpDir string, logger *zap.Logger) {
	manager := ctfknowledge.NewManager(wpDir, logger)
	if err := manager.BuildIndex(); err != nil {
		fmt.Printf("❌ 读取索引失败: %v\n", err)
		return
	}

	stats := manager.GetStats()
	fmt.Printf("CTF WP 知识库状态\n")
	fmt.Printf("================\n")
	fmt.Printf("WP 目录: %s\n", wpDir)
	fmt.Printf("总条目数: %d\n\n", stats.TotalEntries)

	if len(stats.Categories) > 0 {
		fmt.Println("分类分布:")
		for name, count := range stats.Categories {
			fmt.Printf("  %s: %d 篇\n", name, count)
		}
	}

	// Output JSON if requested
	if os.Getenv("JSON") == "1" {
		data, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(data))
	}
}
