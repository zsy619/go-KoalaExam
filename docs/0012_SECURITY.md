# 🔒 安全合规

> 描述 KoalaExam 系统在身份认证、数据加密、接口防护、审计合规等方面的安全设计与实施规范。

---

## 1. 安全架构总览

```
                ┌────────────────────────────────────────────┐
                │            WAF / 边缘防护                   │
                │  (Cloudflare / 阿里云 WAF)                  │
                │  • SQL 注入 / XSS / CC 攻击拦截             │
                │  • 地理位置 / IP 黑名单                        │
                └────────────────────────────────────────────┘
                                  │
                                  ▼
                ┌────────────────────────────────────────────┐
                │        Nginx (TLS 1.3 / HTTP/2)            │
                │  • HTTPS 终止 + HSTS                        │
                │  • 限流 (limit_req) + 防爬虫                 │
                │  • 安全 Headers (CSP, X-Frame-Options)     │
                └────────────────────────────────────────────┘
                                  │
                                  ▼
                ┌────────────────────────────────────────────┐
                │        Hertz 应用层防护                       │
                │  • JWT 鉴权 + Refresh 机制                  │
                │  • 接口签名 (HMAC-SHA256)                    │
                │  • 限流 (Redis 滑动窗口)                      │
                │  • 输入校验 (Hertz bind validator)            │
                │  • 全链路 TraceID + 审计日志                  │
                └────────────────────────────────────────────┘
                                  │
                                  ▼
                ┌────────────────────────────────────────────┐
                │        数据层防护                             │
                │  • GORM 参数化（防 SQL 注入）                 │
                │  • 敏感字段加密存储 (AES-256-GCM)             │
                │  • 数据脱敏 (手机号/身份证/邮箱)               │
                │  • 数据库最小权限账号                         │
                │  • 备份加密 + 异地容灾                       │
                └────────────────────────────────────────────┘
```

---

## 2. 身份认证与会话

### 2.1 JWT 双 Token

```go
type Claims struct {
    UserID   uint64 `json: "uid"
    Role     int    `json: "role"
    DeviceID string `json: "did"
    jwt.RegisteredClaims
}

// 签发：Access (2h) + Refresh (7d)
// 验签：HS256 + JWT_SECRET
// 续签：Refresh Token 单次使用，旋转颁发
```

**安全要点**：
- JWT_SECRET ≥ 32 字符强随机，定期轮换（每季度）
- Access Token 短期（2h），Refresh Token 长期（7d）且一次性
- Token 必须带 DeviceID（防 token 劫持到其他设备）
- 注销 / 改密码立即失效：通过黑名单 Redis（TTL = Access 剩余时间）

### 2.2 密码存储

```go
import "golang.org/x/crypto/bcrypt"

func Hash(pwd string) (string, error) {
    return bcrypt.GenerateFromPassword([]byte(pwd), 12) // cost = 12
}

func Verify(pwd, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)) == nil
}
```

**安全要点**：
- 密码 cost = 12（约 250ms 校验，对抗暴力破解）
- 不存储明文 / MD5 / SHA1 / SHA256（已全部拒绝）
- 强密码策略：8 位 + 大小写 + 数字（前端 + 后端双重校验）
- 密码重置：一次性 Token（10min 过期）+ 邮件验证

### 2.3 多端登录控制

```
用户A 登录设备 1 → 颁发 TokenA
用户A 登录设备 2 → 颁发 TokenB
策略 1 (默认)：TokenA 仍然有效（多端登录）
策略 2 (严格)：踢出设备 1，TokenA 加入黑名单
策略由 ems_user.allow_multi_device 控制
```

---

## 3. 数据加密

### 3.1 传输层 TLS

- 全站 HTTPS（TLS 1.3 优先，TLS 1.2 兼容）
- 证书：Let's Encrypt 自动续签 或 企业 OV 证书
- HSTS Header：Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
- OCSP Stapling 启用

### 3.2 存储加密

| 数据 | 加密方式 | 用途 |
|------|----------|------|
| 用户密码 | bcrypt(cost=12) | 密码哈希 |
| JWT 密钥 | AES-256-GCM | 配置文件加密 |
| 数据库备份 | AES-256-CBC | 异地备份文件 |
| 敏感字段（手机/身份证） | AES-256-GCM | 数据库字段加密 |
| 文件上传 | OSS 服务端加密 | 用户头像/附件 |

### 3.3 密钥管理

```
生产环境密钥分级：
  L0: 数据库主密码 → Vault / AWS KMS
  L1: JWT_SECRET   → 环境变量 + KMS 加密注入
  L2: API 签名密钥  → 配置中心 (Apollo / Nacos)
  L3: 业务密钥      → 数据库字段加密表

轮换周期：
  - JWT_SECRET: 90 天
  - 数据库密码: 180 天
  - API 签名密钥: 30 天
```

---

## 4. 接口安全

### 4.1 鉴权中间件链

```go
// 顺序：RequestID → Recovery → CORS → JWT → RateLimit → Audit → Handler
```

### 4.2 接口签名（防重放）

对关键接口（如 /exams/submit、/admin/users）启用 HMAC-SHA256 签名：

```http
POST /api/v1/exams/submit HTTP/1.1
X-Timestamp: 1700000000
X-Nonce: random_uuid
X-Signature: hmac_sha256(...)
```

### 4.3 限流策略

| 接口 | 限制 | 算法 |
|------|------|------|
| /auth/login | 5 req/min/IP | 滑动窗口 |
| /auth/refresh | 60 req/hour/user | 令牌桶 |
| 全局 | 200 QPS/IP | 滑动窗口 |
| /exams/submit | 10 req/min/user | 滑动窗口 |

### 4.4 输入校验

```go
type CreateUserReq struct {
    Username string `json: "username" validate: "required,min=3,max=32,alphanum"
    Password string `json: "password" validate: "required,min=8,max=64"
    Email    string `json: "email"    validate: "required,email"
    Role     int    `json: "role"     validate: "required,oneof=1 2 3"
}

// 防止 XSS：所有用户输入在入库前过 HTML 过滤
```

---

## 5. 防作弊与考试安全

### 5.1 前端防护

| 措施 | 实现 | 绕过难度 |
|------|------|----------|
| 全屏监控 | requestFullscreen + 失焦告警 | 中 |
| 切屏检测 | visibilitychange 事件 | 中 |
| 复制粘贴拦截 | copy/paste/contextmenu preventDefault | 高 |
| 题目乱序 | 后端每次随机 | - |
| 选项乱序 | 后端每次随机 | - |
| 答题自动保存 | 每 10s 同步 Redis + LocalStorage | - |

### 5.2 后端防护

阅卷必须在服务端进行（不允许客户端对比答案）：

```go
func (s *GradingApp) Grade(record *ExamRecord) {
    for _, ans := range record.Answers {
        correct := s.repo.GetCorrectAnswer(ans.QuestionID) // 服务端取
        if compareAnswer(ans.UserAnswer, correct) {
            // 客观题自动给分
        }
    }
}

// 成绩签名
func (s *ExamApp) Sign(record *ExamRecord) string {
    payload := fmt.Sprintf("%d|%.2f|%.2f|%.2f|%s",
        record.ID, record.ObjectiveScore,
        record.SubjectiveScore, record.TotalScore,
        s.config.ScoreSalt)
    return sha256Hex(payload)
}
```

### 5.3 监考审计

```
ems_exam_record.audit_log: [
  { type: "tab_switch", payload: { from: "visible", to: "hidden" }, ts: 1700000010 },
  { type: "copy_attempt", payload: { question_id: 123 }, ts: 1700000020 },
  { type: "fullscreen_exit", payload: {}, ts: 1700000030 },
]

教师监考大屏实时展示异常事件，可强制收卷。
```

---

## 6. 数据脱敏与隐私

### 6.1 字段脱敏

| 场景 | 字段 | 脱敏方式 |
|------|------|----------|
| 列表展示 | 手机号 | 138****5678 |
| 列表展示 | 身份证 | 110101********1234 |
| 列表展示 | 邮箱 | z***@example.com |
| 日志输出 | 密码 / Token | 完全过滤（禁止打印） |
| API 响应 | 密码字段 | 直接删除 |

### 6.2 隐私合规

- GDPR / 个人信息保护法：用户可导出 / 删除个人数据
- 账号注销：30 天软删除 → 永久物理删除
- 操作日志保留 180 天（监管要求）
- 考试记录保留 3 年（教学评估需要）

### 6.3 数据导出审计

所有数据导出（Excel / CSV / API）记录到 ems_data_export_log：

```sql
CREATE TABLE ems_data_export_log (
  id BIGINT PRIMARY KEY,
  operator_id BIGINT,
  export_type VARCHAR(32),
  filter JSON,
  row_count INT,
  file_url VARCHAR(255),
  created_at DATETIME
);
```

---

## 7. 审计与合规

### 7.1 操作审计

所有写操作写入 ems_audit_log：

```sql
CREATE TABLE ems_audit_log (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  action VARCHAR(64) NOT NULL,
  resource_type VARCHAR(32),
  resource_id VARCHAR(64),
  before JSON,
  after JSON,
  ip VARCHAR(45),
  ua VARCHAR(255),
  trace_id VARCHAR(32),
  created_at DATETIME
);
```

### 7.2 审计查询接口

仅超管可访问：

```
GET /admin/audit-logs?user_id=&action=&start=&end=&page=
GET /admin/data-exports?operator_id=&export_type=
```

### 7.3 安全事件响应

| 事件 | 响应时间 | 流程 |
|------|----------|------|
| SQL 注入尝试 | 实时拦截 | WAF 自动 + 告警 |
| 大规模数据导出 | 实时告警 | 自动限流 + 二次确认 |
| 异常登录（异地） | 实时告警 | 短信通知 + 强制验证 |
| 数据泄露 | 1h 内 | CSO 牵头 + 法务 + 公告 |

---

## 8. 安全 Checklist

### 8.1 上线前必须

- [ ] JWT_SECRET ≥ 32 字符，从 Vault/KMS 注入（非明文配置）
- [ ] 数据库账号最小权限（应用账号无 DROP/ALTER）
- [ ] Redis 密码 + 仅监听内网
- [ ] TLS 1.3 + HSTS + OCSP Stapling
- [ ] WAF 规则开启（SQL/XSS/CC）
- [ ] 限流配置审核（防爆破）
- [ ] 审计日志表就绪 + 告警
- [ ] 数据备份 + 异地容灾演练（季度）
- [ ] 依赖漏洞扫描（govulncheck）
- [ ] 日志中无密码 / 无 Token / 无身份证（脱敏验证）

### 8.2 每月检查

- [ ] 慢查询 / 异常登录分析
- [ ] 失败 JWT 签发率
- [ ] 限流命中次数（异常高 → 攻击迹象）
- [ ] 备份恢复演练（抽样验证）

### 8.3 每季度检查

- [ ] JWT_SECRET / 数据库密码轮换
- [ ] 渗透测试（外部第三方）
- [ ] 权限矩阵审计
- [ ] 员工安全培训

---

## 9. 事故响应预案

### 9.1 数据泄露应急

```
1. 确认泄露范围（受影响用户数 / 字段 / 时间）
2. 立即通知：CSO → 法务 → CEO → 业务方
3. 启动技术封堵（封账号 / 改密钥 / 关接口）
4. 保留证据（日志 / 数据库快照）
5. 监管报告（72h 内 / 个保法要求）
6. 用户通知（短信 + 站内信 + 公告）
7. 复盘 + 改进（30 天内出改进方案）
```

### 9.2 攻击应急

```
1. WAF 自动拦截（99% 已处理）
2. 拉黑源 IP（Cloudflare API）
3. 启动备用 IP（Anycast）
4. 限流阈值动态调整
5. 24h 监控加固
```

---

> 维护者：CSO + 后端架构组 / 季度评审