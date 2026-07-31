package ctfknowledge

import "time"

// WPEntry 表示一篇 CTF Write-up
type WPEntry struct {
	Title       string   `json:"title"`
	FilePath    string   `json:"filePath"`
	Category    string   `json:"category"`    // web/pwn/reverse/crypto/misc
	SubCategory string   `json:"subCategory"` // sql_injection/xss/...
	Source      string   `json:"source"`      // CTF比赛名称
	Year        string   `json:"year"`        // 年份
	VulnType    string   `json:"vulnType"`    // 漏洞类型
	Analysis    string   `json:"analysis"`    // 题目分析
	Solution    string   `json:"solution"`    // 解题思路
	Tools       []string `json:"tools"`       // 使用工具
	Payload     string   `json:"payload"`     // 关键 Payload
	FlagMethod  string   `json:"flagMethod"`  // Flag获取方式
	Keywords    []string `json:"keywords"`    // 技术总结关键词
	Summary     string   `json:"summary"`     // 技术总结全文
	Content     string   `json:"content"`     // 完整原始内容
	IndexedAt   time.Time
}

// SearchRequest WP 知识库搜索请求
type SearchRequest struct {
	Category  string   `json:"category,omitempty"`  // web/pwn/reverse/crypto/misc
	Keywords  []string `json:"keywords,omitempty"`  // 关键词列表
	Query     string   `json:"query,omitempty"`     // 自然语言查询
	TopK      int      `json:"topK,omitempty"`      // 返回结果数，默认 5
}

// SearchResult WP 知识库搜索结果
type SearchResult struct {
	Entry      *WPEntry `json:"entry"`
	Similarity float64  `json:"similarity"`  // 相关度 0-1
	MatchField string   `json:"matchField"`  // 匹配字段：category/keyword/title/content
}

// WPIndex WP 知识库索引
type WPIndex struct {
	entries   []*WPEntry
	byID      map[string]*WPEntry
	byPath    map[string]*WPEntry
	byCat     map[string][]*WPEntry
	keywordIdx map[string][]*wpRef // 关键词 → WP 引用
}

type wpRef struct {
	entry *WPEntry
	field string // 匹配的字段名
	score float64
}

// WPStats 知识库统计
type WPStats struct {
	TotalEntries  int               `json:"totalEntries"`
	Categories    map[string]int    `json:"categories"`
	SubCategories map[string]int    `json:"subCategories"`
	IndexedAt     string            `json:"indexedAt"`
}
