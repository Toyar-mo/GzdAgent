# double_rsa - RSA双重加密

## 来源
CTF比赛: 2026年广东省大学生计算机设计大赛"网络与数据安全挑战赛"决赛
年份: 2026

## 类型
Crypto - RSA双重加密 / 已知phi分解n

## 题目分析
题目使用两轮RSA加密，flag被分为两部分分别加密。第一部分：利用参数h与n1的关系分解n1；第二部分：已知phi值，通过`sum_pq = n + 1 - phi`计算p+q，再解二次方程分解n2，加密指数e2为q2的前一个素数。

## 解题思路
1. 信息收集
   - 获取两轮加密的参数：h、c1、n1、phi、c2、n2
   - 第一轮利用h与n1的关系
   - 第二轮已知phi，加密指数e2需根据q2计算

2. 漏洞分析
   - 第一轮：`D = (h-1)^2 + 4*n1`为完全平方数，可解出q1和p1
   - 第二轮：已知phi可通过`p+q = n+1-phi`和`pq = n`解二次方程分解n2

3. 漏洞利用
   - 第一轮：计算`sqrtD = isqrt(D)`，`q1 = (1-h+sqrtD)//2`，`p1 = n1//q1`
   - 第二轮：计算`sum_pq = n2+1-phi`，解二次方程得p2、q2
   - 计算`e2 = prev_prime(q2)`
   - 分别解密后用`m1 + m2`拼接获取flag

4. Flag获取
   - 解密两部分密文，拼接后得到完整flag

## 使用工具
- Python
- gmpy2库
- Crypto.Util.number

## Payload
```python
import gmpy2
from Crypto.Util.number import long_to_bytes

# 第一轮：利用h分解n1
D = (h - 1)**2 + 4 * n1
sqrtD = gmpy2.isqrt(D)
q1 = (1 - h + sqrtD) // 2
p1 = n1 // q1
phi1 = (p1 - 1) * (q1 - 1)
d1 = gmpy2.invert(65537, phi1)
m1 = pow(c1, d1, n1)

# 第二轮：已知phi分解n2
sum_pq = n2 + 1 - phi
discriminant = sum_pq**2 - 4 * n2
sqrt_disc = gmpy2.isqrt(discriminant)
p2 = (sum_pq - sqrt_disc) // 2
q2 = (sum_pq + sqrt_disc) // 2

# e2为q2的前一个素数
def prev_prime(n):
    while True:
        n -= 1
        if gmpy2.is_prime(n):
            return n

e2 = prev_prime(q2)
d2 = gmpy2.invert(e2, phi)
m2 = pow(c2, d2, n2)

flag = long_to_bytes(m1) + long_to_bytes(m2)
print(flag)
```

## Flag获取方式
利用参数h和phi分别分解两轮RSA的模数n，解密后拼接得到flag。

## 技术总结
关键词：RSA, 双重加密, 已知phi分解n, 二次方程, 完全平方数

核心要点：
1. 已知`(p-1)(q-1)`和`pq`可解`p+q = n+1-phi`，进而解二次方程分解n
2. 参数h关联n的完全平方关系可用于分解
3. `prev_prime`函数用于计算前一个素数
4. 多轮加密中flag分段加密是常见题型