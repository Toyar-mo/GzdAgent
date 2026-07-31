# Pooop - PHP反序列化POP链

## 来源
CTF比赛: 2026年广东省大学生计算机设计大赛"网络与数据安全挑战赛"决赛
年份: 2026

## 类型
Web - PHP反序列化 / POP链构造

## 题目分析
题目提供PHP源码，包含三个类：`FileReader`、`DataProcessor`、`OutputHandler`。需要构造POP链通过反序列化读取`flag.php`文件。flag在`flag.php`中。

## 解题思路
1. 信息收集
   - 打开题目看到PHP源码，复制到本地调试
   - 源码包含三个类，目标文件为`flag.php`

2. 漏洞分析
   - `FileReader.read()`：检查文件是否存在，存在则读取并输出
   - `DataProcessor.process()`：调用成员变量`reader`的`read()`方法，`reader`可控
   - `DataProcessor.__toString()`：调用`process()`，对象被当作字符串时触发
   - `OutputHandler.render()`：输出成员变量`processor`，`processor`可控
   - `OutputHandler.__destruct()`：调用`render()`，程序结束时自动触发

3. 漏洞利用
   - 构造POP链：`OutputHandler.__destruct()` → `render()` → 输出`DataProcessor`对象 → 触发`__toString()` → `process()` → 调用`FileReader.read('flag.php')`
   - 生成序列化payload，提交到题目

4. Flag获取
   - 反序列化后，Flag需要Ctrl+U查看源代码获取（php代码被HTML渲染）

## 使用工具
- PHP本地环境
- 浏览器开发者工具

## Payload
```php
<?php
$FR = new FileReader('flag.php');
$DP = new DataProcessor($FR, 123);
$OH = new OutputHandler($DP, 'html');
$a = serialize($OH);
echo $a;
unserialize($a);
?>
```

## Flag获取方式
构造上述POP链序列化payload，提交到题目触发反序列化，在响应中Ctrl+U查看源代码获取flag。

## 技术总结
关键词：PHP反序列化, POP链, __destruct, __toString, 魔术方法

核心要点：
1. POP链构造需要从触发点（`__destruct`/`__wakeup`）反向追踪到目标点（文件读取）
2. `__destruct`在对象销毁时自动触发，是常见的反序列化入口
3. `__toString`在对象被当作字符串使用时触发
4. `method_exists`检查可以用同类型对象替代，构造调用链
5. 文件读取后通过echo输出，注意HTML渲染可能导致flag被隐藏