# dataTypeV - SQLite数据分类

## 来源
CTF比赛: 2026年广东省大学生计算机设计大赛"网络与数据安全挑战赛"决赛
年份: 2026

## 类型
Misc - 数据分析 / SQLite / 数据分类

## 题目分析
题目提供一个SQLite数据库文件`data.db`，需要遍历所有表中的所有列，将数据按照类型（姓名、手机号、身份证号、邮箱）进行分类，并将分类结果保存为CSV文件。

## 解题思路
1. 信息收集
   - 获取SQLite数据库文件
   - 分析数据库结构和内容

2. 漏洞分析
   - 需要自动识别数据类型：身份证号（18位）、手机号（11位）、邮箱（含@和.）、姓名（2-4个中文字符）

3. 漏洞利用
   - 编写Python脚本连接SQLite数据库
   - 遍历所有表和所有列
   - 对每个值进行正则匹配分类

4. Flag获取
   - 分类结果输出到CSV文件中，从中获取flag

## 使用工具
- Python
- sqlite3库
- csv库
- re（正则表达式）

## Payload
```python
import sqlite3
import csv
import re

def classify_value(value):
    s = str(value).strip()
    if not s:
        return None
    # 身份证号 (18位，前17位数字，最后1位数字或X)
    if len(s) == 18 and s[:17].isdigit() and (s[17].isdigit() or s[17].upper() == 'X'):
        return 'idcard'
    # 手机号 (11位纯数字)
    if len(s) == 11 and s.isdigit():
        return 'phone'
    # 邮箱
    if '@' in s and '.' in s and s.index('@') > 0:
        domain_part = s[s.index('@')+1:]
        if '.' in domain_part and domain_part.index('.') > 0:
            return 'email'
    # 姓名 (2-4个中文字符)
    if 2 <= len(s) <= 4 and re.search(r'[\u4e00-\u9fa5]', s):
        return 'name'
    return None

conn = sqlite3.connect('data.db')
cursor = conn.cursor()
cursor.execute("SELECT name FROM sqlite_master WHERE type='table';")
tables = [t[0] for t in cursor.fetchall()]

results = []
for table in tables:
    cursor.execute(f"PRAGMA table_info({table});")
    columns = [col[1] for col in cursor.fetchall()]
    for col in columns:
        cursor.execute(f"SELECT {col} FROM {table} WHERE {col} IS NOT NULL;")
        for row in cursor.fetchall():
            typ = classify_value(row[0])
            if typ:
                results.append((typ, str(row[0])))

with open('result.csv', 'w', newline='', encoding='utf-8') as f:
    writer = csv.writer(f)
    writer.writerow(['type', 'value'])
    writer.writerows(results)
```

## Flag获取方式
运行脚本遍历SQLite数据库全表全列，自动分类数据并输出CSV，从分类结果中获取flag。

## 技术总结
关键词：SQLite, 数据分类, 正则表达式, CSV, Python

核心要点：
1. SQLite使用`sqlite_master`表查询所有用户表名
2. `PRAGMA table_info`获取表结构信息
3. 正则表达式用于匹配中文姓名：`[\u4e00-\u9fa5]`
4. 身份证号校验：18位，前17位数字，最后1位数字或X
5. 数据分类核心是编写准确的匹配规则