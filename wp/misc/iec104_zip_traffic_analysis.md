# IEC104 - Zip伪加密与流量分析

## 来源
CTF比赛: 第七届全国工控系统信息安全攻防竞赛
年份: 2021

## 类型
Misc - Zip伪加密 / IEC104 流量分析 / 3DES解密

## 题目分析
题目提供一个加密的ZIP压缩包，要求查看压缩包内隐藏的信息。ZIP文件存在伪加密，内部包含一个 `ice104.pcap` 流量包，流量包中隐藏了加密的 flag 文件，需要结合密钥进行解密。flag格式为 `flag{xxx}`。

## 解题思路
1. 信息收集
   - 在ZIP文件末尾发现一串字符 `KEY:XHR666`
   - 打开压缩包提示有密码，但实际为伪加密

2. 漏洞分析
   - ZIP伪加密：通过修改ZIP文件的加密标志位，使解压软件误以为文件被加密，实际数据并未加密
   - 流量包中隐藏了 `FLAG.txt` 文件

3. 漏洞利用
   - 使用 `zipCenOp.jar` 工具处理ZIP伪加密，成功解压出 `ice104.pcap` 流量包
   - 使用 `binwalk` 分析流量包，提取出 `FLAG.txt`
   - FLAG.txt 内容为：`U2FsdGVkX1/A8ta0AedfEebXcmyZnwA2U4SXuZ9oW2NQ1AuJbF9dmg==`

4. Flag获取
   - `U2Fsd` 开头为常见对称加密算法特征（AES、DES、3DES等）
   - 以 `XHR666` 作为密钥，依次尝试AES、DES、3DES解密
   - 确认加密算法为 **3DES**，解密成功获得 flag

## 使用工具
- zipCenOp.jar - ZIP伪加密处理
- binwalk - 文件分析/提取
- CyberChef / Python - 3DES 解密

## Payload
```
加密数据: U2FsdGVkX1/A8ta0AedfEebXcmyZnwA2U4SXuZ9oW2NQ1AuJbF9dmg==
密钥: XHR666
算法: 3DES
```

## Flag获取方式
从ZIP文件末尾提取密钥 `XHR666`，使用 `zipCenOp.jar` 去除伪加密后解压获取 `ice104.pcap`，用 `binwalk` 从流量包中提取 `FLAG.txt`，将加密内容以 `XHR666` 为密钥进行3DES解密即可得到 flag。

## 技术总结
关键词：ZIP伪加密, 流量分析, IEC104, 3DES解密, binwalk, 工控安全

核心要点：
1. ZIP伪加密仅修改加密标志位，数据本身未加密，可通过工具修复
2. `U2Fsd` 开头的Base64编码数据通常为 OpenSSL 的 `Salted__` 格式，常见于AES/DES/3DES加密
3. binwalk 可用于从二进制文件（包括pcap流量包）中提取嵌入文件
4. 工控协议（如IEC104）的流量分析是工业信息安全竞赛的常见考点