package ctfknowledge

import (
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"go.uber.org/zap"
)

// Manager WP 知识库管理器
type Manager struct {
	mu       sync.RWMutex
	index    *WPIndex
	wpDir    string
	logger   *zap.Logger
}

// NewManager 创建 WP 知识库管理器
func NewManager(wpDir string, logger *zap.Logger) *Manager {
	return &Manager{
		index:  &WPIndex{
			entries:    make([]*WPEntry, 0),
			byID:       make(map[string]*WPEntry),
			byPath:     make(map[string]*WPEntry),
			byCat:      make(map[string][]*WPEntry),
			keywordIdx: make(map[string][]*wpRef),
		},
		wpDir:  wpDir,
		logger: logger,
	}
}

// BuildIndex 构建 WP 知识库索引
func (m *Manager) BuildIndex() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := ScanWPDirectory(m.wpDir, m.logger)
	if err != nil {
		return err
	}

	newIndex := &WPIndex{
		entries:    make([]*WPEntry, 0, len(entries)),
		byID:       make(map[string]*WPEntry),
		byPath:     make(map[string]*WPEntry),
		byCat:      make(map[string][]*WPEntry),
		keywordIdx: make(map[string][]*wpRef),
	}

	for _, entry := range entries {
		newIndex.entries = append(newIndex.entries, entry)
		newIndex.byPath[entry.FilePath] = entry

		// Build category index
		if entry.Category != "" {
			newIndex.byCat[entry.Category] = append(newIndex.byCat[entry.Category], entry)
		}
		if entry.SubCategory != "" {
			subKey := entry.Category + "/" + entry.SubCategory
			newIndex.byCat[subKey] = append(newIndex.byCat[subKey], entry)
		}

		// Build keyword inverted index
		m.indexEntry(newIndex, entry)
	}

	newIndex.byCat["all"] = newIndex.entries
	m.index = newIndex
	m.logger.Info("WP 知识库索引构建完成",
		zap.Int("entries", len(newIndex.entries)),
		zap.Int("keywords", len(newIndex.keywordIdx)),
	)
	return nil
}

// indexEntry 为一个 WP 构建关键词索引
func (m *Manager) indexEntry(idx *WPIndex, entry *WPEntry) {
	// 从标题提取关键词
	for _, tk := range tokenize(entry.Title) {
		if len(tk) < 2 {
			continue
		}
		idx.keywordIdx[tk] = append(idx.keywordIdx[tk], &wpRef{entry: entry, field: "title", score: 0.9})
	}

	// 从关键词列表提取
	for _, kw := range entry.Keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		idx.keywordIdx[kw] = append(idx.keywordIdx[kw], &wpRef{entry: entry, field: "keyword", score: 1.0})
		// Also index sub-tokens of multi-char keywords
		for _, tk := range tokenize(kw) {
			if len(tk) >= 2 {
				idx.keywordIdx[tk] = append(idx.keywordIdx[tk], &wpRef{entry: entry, field: "keyword_sub", score: 0.7})
			}
		}
	}

	// 从工具名提取
	for _, tool := range entry.Tools {
		for _, tk := range tokenize(tool) {
			if len(tk) >= 2 {
				idx.keywordIdx[tk] = append(idx.keywordIdx[tk], &wpRef{entry: entry, field: "tool", score: 0.85})
			}
		}
	}

	// 从漏洞类型提取
	for _, tk := range tokenize(entry.VulnType) {
		if len(tk) >= 2 {
			idx.keywordIdx[tk] = append(idx.keywordIdx[tk], &wpRef{entry: entry, field: "vulntype", score: 0.8})
		}
	}

	// 从题目分析提取重要片段
	for _, tk := range tokenize(entry.Analysis) {
		if len(tk) < 2 {
			continue
		}
		idx.keywordIdx[tk] = append(idx.keywordIdx[tk], &wpRef{entry: entry, field: "analysis", score: 0.5})
	}

	// 从解题思路提取
	for _, tk := range tokenize(entry.Solution) {
		if len(tk) < 2 {
			continue
		}
		idx.keywordIdx[tk] = append(idx.keywordIdx[tk], &wpRef{entry: entry, field: "solution", score: 0.5})
	}
}

// Search 搜索 WP 知识库
func (m *Manager) Search(req *SearchRequest) []*SearchResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.index == nil || len(m.index.entries) == 0 {
		return nil
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 5
	}

	// Collect candidates with scores
	candidates := make(map[*WPEntry]*searchScore)

	// 1. Filter by category first
	var pool []*WPEntry
	if req.Category != "" {
		catKey := req.Category
		if pool = m.index.byCat[catKey]; pool == nil {
			// Try category/subcategory
			for key, entries := range m.index.byCat {
				if strings.HasPrefix(key, catKey+"/") {
					pool = append(pool, entries...)
				}
			}
		}
	}
	if pool == nil {
		pool = m.index.entries
	}

	// 2. Score by keywords
	queryTokens := make(map[string]bool)
	if req.Query != "" {
		for _, tk := range tokenize(req.Query) {
			if len(tk) >= 2 {
				queryTokens[tk] = true
			}
		}
	}
	for _, kw := range req.Keywords {
		kw = strings.TrimSpace(kw)
		if kw != "" {
			queryTokens[kw] = true
		}
	}

	if len(queryTokens) > 0 {
		// Score using inverted index
		for token := range queryTokens {
			refs := m.index.keywordIdx[token]
			for _, ref := range refs {
				score, ok := candidates[ref.entry]
				if !ok {
					score = &searchScore{}
					candidates[ref.entry] = score
				}
				score.score += ref.score
			}
		}

		// Also do full-text content matching for query
		if req.Query != "" {
			queryLower := strings.ToLower(req.Query)
			for _, entry := range pool {
				if _, exists := candidates[entry]; exists {
					continue
				}
				contentLower := strings.ToLower(entry.Content)
				if strings.Contains(contentLower, queryLower) {
					candidates[entry] = &searchScore{score: 0.3}
				}
			}
		}
	} else {
		// No query tokens, return pool entries with base score
		for _, entry := range pool {
			candidates[entry] = &searchScore{score: 0.1}
		}
	}

	// 3. Sort and build results
	sorted := make([]*entryScore, 0, len(candidates))
	for entry, score := range candidates {
		// Normalize score to 0-1 range
		normalized := score.score
		if normalized > 1.0 {
			normalized = 1.0
		}
		sorted = append(sorted, &entryScore{entry: entry, score: normalized})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	if len(sorted) > topK {
		sorted = sorted[:topK]
	}

	results := make([]*SearchResult, 0, len(sorted))
	for _, es := range sorted {
		results = append(results, &SearchResult{
			Entry:      es.entry,
			Similarity: es.score,
			MatchField: bestMatchField(es.entry, queryTokens),
		})
	}

	return results
}

// GetStats 获取知识库统计信息
func (m *Manager) GetStats() *WPStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &WPStats{
		TotalEntries:  len(m.index.entries),
		Categories:    make(map[string]int),
		SubCategories: make(map[string]int),
	}

	for _, entry := range m.index.entries {
		if entry.Category != "" {
			stats.Categories[entry.Category]++
		}
		if entry.SubCategory != "" {
			subKey := entry.Category + "/" + entry.SubCategory
			stats.SubCategories[subKey]++
		}
	}

	return stats
}

// GetEntries 获取所有 WP 条目
func (m *Manager) GetEntries(category string) []*WPEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if category == "" || category == "all" {
		result := make([]*WPEntry, len(m.index.entries))
		copy(result, m.index.entries)
		return result
	}

	if entries, ok := m.index.byCat[category]; ok {
		result := make([]*WPEntry, len(entries))
		copy(result, entries)
		return result
	}
	return nil
}

// GetEntryByPath 根据文件路径获取 WP
func (m *Manager) GetEntryByPath(path string) *WPEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.index.byPath[path]
}

// tokenize 将文本分词为 token 列表
// 对中文使用 bi-gram + uni-gram 混合，对英文按空格拆分
func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	seen := make(map[string]bool)

	// Process rune by rune
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		r := runes[i]

		if r > 0x4e00 && r < 0x9fff {
			// Chinese character: extract bi-grams and uni-grams
			if i+1 < len(runes) && runes[i+1] > 0x4e00 && runes[i+1] < 0x9fff {
				bigram := string(runes[i]) + string(runes[i+1])
				if !seen[bigram] {
					tokens = append(tokens, bigram)
					seen[bigram] = true
				}
			}
			ch := string(r)
			if !seen[ch] {
				tokens = append(tokens, ch)
				seen[ch] = true
			}
			i++
		} else if r > 0x3000 && r < 0x4e00 {
			// CJK punctuation
			i++
		} else if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			// English word / number
			start := i
			for i < len(runes) {
				c := runes[i]
				if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
					i++
				} else {
					break
				}
			}
			word := string(runes[start:i])
			if utf8.RuneCountInString(word) >= 2 && !seen[word] {
				tokens = append(tokens, word)
				seen[word] = true
			}
		} else {
			i++
		}
	}

	return tokens
}

type searchScore struct {
	score float64
}

type entryScore struct {
	entry *WPEntry
	score float64
}

// bestMatchField 找出最佳匹配字段
func bestMatchField(entry *WPEntry, queryTokens map[string]bool) string {
	if len(queryTokens) == 0 {
		return "all"
	}
	for token := range queryTokens {
		for _, kw := range entry.Keywords {
			if strings.Contains(strings.ToLower(kw), token) || strings.Contains(token, strings.ToLower(kw)) {
				return "keyword"
			}
		}
		if strings.Contains(strings.ToLower(entry.Title), token) {
			return "title"
		}
		if strings.Contains(strings.ToLower(entry.VulnType), token) {
			return "vulntype"
		}
	}
	return "content"
}
