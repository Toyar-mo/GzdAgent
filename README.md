<p align="center">
  <img src="images/gzd.png" alt="GzdAgent" height="48">
</p>

<h1 align="center">GzdAgent</h1>

<p align="center">
  <strong>面向 CTF 安全竞赛的智能辅助分析 Agent 平台</strong>
</p>

<p align="center">
  <a href="#快速开始">快速开始</a> ·
  <a href="#功能特性">功能特性</a> ·
  <a href="#架构概览">架构概览</a> ·
  <a href="#部署">部署</a> ·
  <a href="#配置">配置</a>
</p>

---

## 简介

GzdAgent 是一个面向 CTF（Capture The Flag）安全竞赛场景的智能辅助分析平台。用户通过自然语言与 AI Agent 交互，Agent 自动调用工具、检索知识库、参考历史 Write-up 解题报告，为参赛者提供解题思路和攻击路径建议。

**核心能力**：AI 对话 + MCP 工具调用 + 知识库检索 + CTF WP 经验匹配。

## 功能特性

### Agent 对话

- 多轮对话，SSE 流式输出，Markdown 渲染
- 单 Agent ReAct 模式 + 多 Agent 编排（Deep / Plan-Execute / Supervisor）
- 支持 OpenAI 兼容接口和 Claude 原生 API
- 推理模式（thinking/effort）可配置

### CTF 解题辅助

- 题目类型自动识别（Web / PWN / Reverse / Crypto / Misc）
- 调用 `search_ctf_wp` 工具检索历史 WP 知识库
- 基于关键词匹配 + 类别过滤，返回相似题目的解题思路和 Payload 参考
- WP 知识库按类别目录组织，Markdown 格式，支持 CLI 管理

### 工具系统

- 55+ 内置 MCP 工具：漏洞管理、资产扫描、WebShell 管理、C2 框架、知识库检索、项目事实图等
- 支持外部 MCP 服务器接入，扩展工具能力
- 工具执行监控：超时控制、取消、结果截断、并发限制

### 知识库

- RAG 向量检索：文档上传、嵌入、多查询检索、重排序
- WP 倒排索引：CJK + 英文分词，关键词精确匹配
- 检索历史审计

### 项目管理

- 项目创建、事实黑板（Fact Graph）
- 攻击链可视化（节点 + 边 DAG）
- 漏洞记录与资产关联

### 企业功能

- RBAC 权限控制（用户 → 角色 → 权限，三级资源范围）
- 操作审计日志
- 可视化工作流编排
- 批量任务调度
- 8 平台聊天机器人集成（微信、企业微信、钉钉、飞书、Telegram、Slack、Discord、QQ）

## 快速开始

### 环境要求

- Go 1.25+
- Linux 64-bit（推荐）或 macOS
- 可访问的 LLM API（OpenAI 兼容或 Claude）

### 编译

```bash
git clone https://github.com/Toyar-mo/GzdAgent.git
cd GzdAgent
go build -o gzd-agent ./cmd/server/
```

### 配置

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml`，至少配置模型 API：

```yaml
ai:
  default_channel: "deepseek"
  channels:
    deepseek:
      provider: "openai"
      base_url: "https://api.deepseek.com/v1"
      api_key: "${DEEPSEEK_API_KEY}"
      model: "deepseek-chat"
```

### 启动

```bash
./gzd-agent --config config.yaml
```

浏览器访问 `http://localhost:8080`，默认管理员账号 `admin`，初始密码在启动日志中打印。

## 架构概览

```
┌─────────────────────────────────────────┐
│              客户端层                     │
│  Web SPA (Vanilla JS)  │  机器人 (8平台)  │
└─────────────────────────────────────────┘
                    │ HTTP/SSE
┌─────────────────────────────────────────┐
│            HTTP 服务层 (Gin)              │
│   认证 + RBAC + 限流 + 审计               │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│             应用服务层                     │
│  ┌─────────────────────────────────┐    │
│  │ Agent 引擎 (Eino ReAct)         │    │
│  │  单Agent │ Deep │ Plan │ Sup    │    │
│  ├─────────────────────────────────┤    │
│  │ MCP 工具系统 (55+ 内置工具)      │    │
│  ├─────────────────────────────────┤    │
│  │ 知识库 (RAG + WP 倒排索引)       │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│              数据层 (SQLite)              │
│   30+ 表: 对话/消息/工具执行/项目/审计...   │
└─────────────────────────────────────────┘
```

## 技术栈

| 层 | 技术 |
|----|------|
| 后端语言 | Go 1.25 |
| HTTP 框架 | Gin |
| Agent 框架 | CloudWeGo Eino |
| 工具协议 | MCP (Model Context Protocol) |
| 数据库 | SQLite (WAL 模式) |
| 前端 | Vanilla JS SPA |
| 日志 | zap |
| 配置 | YAML |

## 项目结构

```
GzdAgent/
├── cmd/server/          # Web 服务入口
├── cmd/ctfcli/          # CTF WP CLI 工具
├── internal/
│   ├── agent/           # Agent 引擎
│   ├── app/             # 应用启动
│   ├── config/          # 配置管理
│   ├── ctfknowledge/    # CTF WP 知识库
│   ├── database/        # 数据库层
│   ├── handler/         # HTTP API
│   ├── knowledge/       # RAG 知识库
│   ├── mcp/             # MCP 协议
│   ├── multiagent/      # 多 Agent 编排
│   ├── openai/          # LLM 客户端
│   ├── project/         # 项目事实图
│   ├── robot/           # 机器人集成
│   ├── security/        # 安全工具/鉴权
│   ├── workflow/        # 工作流引擎
│   └── ...
├── web/                 # 前端 SPA
├── wp/                  # CTF WP 知识库
├── skills/              # Agent Skills
├── agents/              # 多 Agent 定义
├── roles/               # 角色配置
└── config.example.yaml  # 配置模板
```

## 配置

主要配置项：

| 配置节 | 说明 |
|--------|------|
| `server` | HTTP/HTTPS 服务端口、TLS |
| `ai` | 多频道 AI 模型配置 |
| `agent` | Agent 迭代次数、超时、并发 |
| `mcp` | MCP 服务器开关 |
| `external_mcp` | 外部 MCP 服务器 |
| `knowledge` | RAG 知识库配置 |
| `ctf_knowledge` | CTF WP 知识库开关和路径 |
| `database` | SQLite 路径 |
| `multi_agent` | 多 Agent 编排模式 |
| `robots` | 机器人平台配置 |
| `security` | 工具目录、描述模式 |

API Key 支持环境变量：`api_key: "${DEEPSEEK_API_KEY}"`

## 添加 CTF WP

在 `wp/` 目录下按类别创建 Markdown 文件：

```
wp/
├── web/sql_injection/basic_union_sqli.md
├── pwn/rop/ret2libc.md
└── reverse/crackme.md
```

WP 格式参考 `wp/README.md`。重启服务后自动索引。

CLI 管理：

```bash
go run ./cmd/ctfcli/ index
go run ./cmd/ctfcli/ search "sql注入"
go run ./cmd/ctfcli/ status
```

## 安全说明

- 仅可对自有系统或已获得明确授权的目标使用
- 在共享或生产环境启用高风险工具前，请审阅安全配置
- API Key 通过环境变量注入，不写入配置文件
- 默认生成随机管理员密码，启动时打印

## 贡献者

<a href="https://github.com/Toyar-mo"><img src="https://github.com/Toyar-mo.png" width="48" height="48" alt="Toyar-mo"></a>

贡献者列表持续更新中，欢迎提交 Pull Request 加入！

## License

Apache License 2.0