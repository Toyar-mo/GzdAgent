# CTF Write-up 知识库

本目录存储 CTF 比赛历史优秀选手的 Write-up（WP）解题报告，供 AI Agent 在自动化解题时参考。

## 目录结构

```
wp/
├── web/           # Web 安全
│   ├── sql_injection/  # SQL 注入
│   ├── xss/            # 跨站脚本
│   ├── ssrf/           # 服务端请求伪造
│   ├── file_upload/    # 文件上传
│   ├── rce/            # 远程命令执行
│   └── other/          # 其他 Web
├── pwn/           # 二进制漏洞利用
├── reverse/       # 逆向工程
├── crypto/        # 密码学
└── misc/          # 杂项
```

## WP 格式要求

每个 WP 文件使用 Markdown 格式，必须包含以下内容：

```markdown
# 题目名称

## 来源
CTF比赛:
年份:

## 类型
漏洞/题目类型

## 题目分析
描述漏洞环境与题目背景

## 解题思路
1. 信息收集
2. 漏洞分析
3. 漏洞利用
4. Flag获取

## 使用工具
工具列表

## Payload
关键 Payload

## Flag获取方式
Flag 位置与获取方法

## 技术总结
关键词：xxx, xxx
```

## 使用说明

Agent 在解题时会自动检索本目录下的 WP 作为参考。WP 仅作辅助决策，不替代真实漏洞分析。
