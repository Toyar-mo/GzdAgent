# CTF Knowledge Base — CTF 解题知识库

## 概述

CTF Knowledge Base 是基于 gzdAI 平台的增量功能模块，让 AI Agent 在 CTF 自动化解题过程中能够检索历史优秀选手的 Write-up（WP）解题报告，提取其中的攻击方法论、攻击链、工具使用方式，融合到当前解题任务中，大幅提高解题效率与成功率。

---

## 核心架构

```
用户输入
  ↓
LLM 分析题目类型（web/pwn/reverse/crypto/misc）
  ↓
调用 search_ctf_wp(category, keywords, query)  ← MCP Tool
  ↓
ctfknowledge.Manager.Search()
  ├─ 分类过滤（目录匹配）
  ├─ 关键词倒排索引匹配（标题/类型/工具/关键字）
  └─ 全文子串匹配（题目分析/解题思路）
  ↓
返回相关 WP 经验（攻击思路 + Payload + 工具建议）
  ↓
LLM 融合 WP 参考 + 当前环境 → 生成攻击方案
  ↓
调用常规工具执行
```

### 技术特点

| 特性 | 方案 |
|---|---|
| **轻量无外部依赖** | 纯 Go 实现，无需第三方向量库或商业 API |
| **中文分词** | bi-gram + 英文 tokenize 混合分词 |
| **检索策略** | 倒排索引 + TF 评分 + 多字段匹配 |
| **索引方式** | 启动时全量扫描，运行时内存索引 |
| **集成方式** | MCP Tool 注册，对 Agent 透明 |

---

## 模块结构

```
gzdAI/
├── wp/                                    # Write-up 知识库目录
│   ├── README.md                          # 目录说明
│   ├── web/                               # Web 安全
│   │   ├── sql_injection/                 # SQL 注入
│   │   │   ├── basic_union_sqli.md        # 联合查询注入 WP
│   │   │   └── blind_sqli_time.md         # 时间盲注 WP
│   │   ├── xss/                           # XSS
│   │   ├── ssrf/                          # SSRF
│   │   ├── file_upload/                   # 文件上传
│   │   ├── rce/                           # 远程命令执行
│   │   └── other/                         # 其他 Web
│   ├── pwn/                               # 二进制漏洞利用
│   ├── reverse/                           # 逆向工程
│   ├── crypto/                            # 密码学
│   └── misc/                              # 杂项
│
├── internal/
│   └── ctfknowledge/                      # CTF 知识库核心模块
│       ├── types.go                       # 数据结构定义
│       ├── parser.go                      # Markdown WP 解析器
│       ├── index.go                       # 关键词倒排索引 + 搜索
│       └── tool.go                        # MCP 工具注册
│
├── cmd/ctfcli/                            # CLI 管理工具
│   └── main.go                            # 命令行入口
│
└── internal/app/
    └── ctf_tools.go                       # App 层工具装配
```

---

## WP 文件格式

每个 WP 文件使用标准 Markdown，包含以下章节：

```markdown
# 题目名称

## 来源
CTF比赛: BugKu
年份: 2023

## 类型
SQL Injection - Union Based

## 题目分析
描述漏洞环境与题目背景...

## 解题思路
1. 判断注入点
2. 确认注入类型
3. 确定字段数...

## 使用工具
- sqlmap
- Burp Suite
- Python

## Payload
关键 Payload 代码...

## Flag获取方式
Flag 位置与获取方法...

## 技术总结
关键词：union注入, 联合查询, information_schema
```

---

## 配置方式

在 `config.yaml` 中添加：

```yaml
ctf_knowledge:
  enabled: true      # 启用 CTF 知识库
  wp_dir: ./wp       # WP 目录路径
  top_k: 5           # 搜索结果返回数
```

---

## MCP 工具

### search_ctf_wp

在 CTF 解题知识库中搜索历史 Write-up，获取类似题目的攻击思路。

**输入参数：**

```json
{
  "category": "web",
  "keywords": ["sql", "bypass", "upload"],
  "query": "一个登录页面存在SQL注入，需要绕过WAF"
}
```

**返回结果：**

```json
{
  "matched_wp": [
    {
      "title": "基础联合查询注入",
      "similarity": 0.95,
      "summary": "一个简单的 SQL 注入漏洞...",
      "attack_method": "union select 1,2,3--+",
      "tools": ["sqlmap", "Burp Suite"]
    }
  ]
}
```

### ctf_wp_stats

查看 CTF 知识库状态信息（总条目数、分类分布）。

---

## CLI 管理工具

```bash
# 创建 WP 目录结构
go run ./cmd/ctfcli/ init

# 扫描并建立索引
go run ./cmd/ctfcli/ index

# 搜索知识库
go run ./cmd/ctfcli/ search "SQL注入绕过WAF"

# 查看状态
go run ./cmd/ctfcli/ status

# JSON 格式输出
JSON=1 go run ./cmd/ctfcli/ search "SQL注入"
```

---

## Agent Prompt 增强

系统 Prompt 中新增 CTF 解题流程指导：

```
当处理 CTF 题目时，使用以下增强流程：
1. 分析题目类型（web/pwn/reverse/crypto/misc）和具体考点
2. 使用 search_ctf_wp 工具搜索 CTF 知识库
3. 参考优秀选手 WP 中的方法论和攻击链，但不要直接复制答案
4. 结合当前靶机环境重新验证 Payload 和攻击方法
5. 优先选择成功率最高的方法

安全要求：
- WP 仅作参考，不替代真实漏洞分析
- 所有 Payload 必须根据当前环境重新验证
- 知识库辅助决策，核心分析能力仍需自主发挥
```

---

## 安全设计

- **WP 仅作参考**：Agent 不会无条件相信 WP 内容
- **需重新验证**：所有 Payload 必须根据当前靶机重新测试
- **环境适配**：攻击方法需确认环境兼容性
- **核心能力自主**：知识库不替代 Agent 的漏洞分析能力

---

## 应用场景

| 场景 | 说明 |
|---|---|
| **CTF 自动解题** | Agent 自动分析题目，检索 WP，生成攻击方案 |
| **知识积累** | 团队共享历史 WP，沉淀解题经验 |
| **新人培训** | 新手可通过知识库学习常见攻击模式 |
| **竞赛复盘** | 赛后整理 WP，构建可复用的知识体系 |

---

## 测试案例

**用户输入：**

> "帮我分析一道 SQL 注入 CTF 题"

**Agent 执行流程：**

1. 分析题目类型 → `web/sql_injection`
2. 调用 `search_ctf_wp(category="web", keywords=["SQL注入"])`
3. 知识库返回 2 条匹配结果：
   - **基础联合查询注入**（相关度 100%）
   - **时间盲注**（相关度 100%）
4. Agent 融合 WP 经验：
   - 参考联合查询的 `order by` 判字段数方法
   - 参考时间盲注的 `sleep()` 延迟判断技术
   - 结合当前环境选择最优攻击路径
5. 输出：
   - 历史攻击思路总结
   - 当前靶机分析
   - 下一步攻击建议
