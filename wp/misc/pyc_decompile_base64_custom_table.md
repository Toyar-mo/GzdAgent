# Protocol_decryption - pyc反编译与Base64换表

## 来源
CTF比赛: 2026年广东省大学生计算机设计大赛"网络与数据安全挑战赛"决赛
年份: 2026

## 类型
Misc - 流量分析 / pyc反编译 / Base64换表 / 命令执行

## 题目分析
题目涉及HTTP流量分析，从流量中提取pyc文件，反编译后发现Base64换表加密算法，通过分析命令执行流量获取flag。

## 解题思路
1. 信息收集
   - 分析HTTP流量，发现返回了多个文件
   - 其中包含一个pyc文件

2. 漏洞分析
   - 使用pycdc.exe反编译pyc文件
   - 发现是Base64换表加密算法（自定义Base64编码表）

3. 漏洞利用
   - 分析命令执行的流量包
   - 发现`cat flag`命令执行记录
   - 找到加密后的flag数据

4. Flag获取
   - 使用自定义Base64表解码加密数据，得到flag

## 使用工具
- Wireshark
- pycdc.exe（pyc反编译工具）
- Python（Base64换表解码）

## Payload
无（Base64换表解码）

## Flag获取方式
从HTTP流量中提取pyc文件，反编译获取自定义Base64编码表，对命令执行返回的加密数据进行解码得到flag。

## 技术总结
关键词：pyc反编译, Base64换表, 流量分析, 命令执行, pycdc

核心要点：
1. pyc文件可使用pycdc等工具反编译为Python源码
2. Base64换表是常见编码混淆手段，需要提取自定义编码表
3. 流量分析中关注命令执行请求和响应，可能包含flag
4. HTTP流量中文件传输可能包含关键脚本文件