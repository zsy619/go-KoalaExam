# 🚨 故障排查手册（RUNBOOK）

> 面向运维与值守人员的实战手册。按故障现象给出 **症状 → 快速定位 → 修复 → 预防** 的四步流程。

---

## 🚦 故障分级

| 等级 | 含义 | 响应时间 | 处置人 |
|------|------|----------|--------|
| **P0** | 全站不可用、核心考试中断 | 5 分钟 | 全员 + On-call |
| **P1** | 关键功能故障（无法登录/交卷） | 15 分钟 | On-call + 后端 |
| **P2** | 部分功能异常（非阻塞） | 1 小时 | 后端 |
| **P3** | 体验问题、性能降级 | 4 小时 | 后端 |
| **P4** | 优化建议、信息咨询 | 1 工作日 | - |

---

## 📋 故障应急响应 SOP

### P0 应急流程

```
1. 收到告警 → 拉群 → 5min 内确认 P0
2. 同步 Status Page（正在恢复）
3. 启动应急：执行对应故障章节的修复脚本
4. 服务恢复 → 验证核心链路（login / start / submit）
5. 撰写事后报告（30 分钟内出初步 RCA）
```

### 应急联系人

| 角色 | 联系方式 | 备份联系 |
|------|---------|---------|
| On-call 主 | 值班表轮值 | +86-xxx |
| DBA | 张三 | 李四 |
| 安全 | 王五（CSO 值班） | - |
| 业务方 | 教学总监 | 客户成功 |

---

## 1️⃣ 服务无法访问（HTTP 5xx / 健康检查失败）

### 症状
- GET /health 返回 5xx 或超时
- 监控告警：koala_http_requests_5xx_total 飙升
- 用户反馈：网站 502/504

### 快速定位

```bash
# 1. 检查后端 Pod 是否存活
kubectl get pods -n koala -l app=koala-exam-backend

# 2. 查看最近日志
kubectl logs -n koala -l app=koala-exam-backend --tail=200 --since=10m | grep -i error

# 3. 检查依赖
curl -I http://mysql:3306
redis-cli -h redis ping

# 4. 查看资源
kubectl top pods -n koala
```

### 常见原因 & 修复

| 原因 | 修复 |
|------|------|
| **MySQL 连接耗尽** | 调 max_open_conns，查慢查询 |
| **Redis OOM** | INFO memory 看 used_memory，清理 keys |
| **配置错误** | kubectl rollout undo deployment/koala-exam-backend |
| **镜像拉取失败** | 检查 imagePullSecrets、镜像仓库网络 |
| **Netpoll 异常** | 临时设置 APP_ENABLE_NETPOLL=false 回滚 Go 原生网络库 |

---

## 2️⃣ 登录失败（401 / 403 风暴）

### 症状
- POST /auth/login 返回 401 比例 > 5%
- koala_jwt_token_invalid 告警

### 排查

```bash
# 1. JWT_SECRET 是否被改
kubectl get configmap koala-config -o yaml | grep jwt_secret

# 2. 时钟漂移
kubectl exec -it pod -- date

# 3. 数据库用户表是否被锁
mysql -e "SELECT * FROM information_schema.innodb_trx"
```

### 修复

| 场景 | 操作 |
|------|------|
| JWT_SECRET 误改 | 回滚 ConfigMap + 前端强制刷新 |
| 数据库行锁 | 找出长事务 innodb_lock_waits 强杀 |
| bcrypt 升级不兼容 | 临时回退到原算法版本，兼容期 7 天 |

---

## 3️⃣ 考试无法开始（/exams/:id/start 失败）

### 症状
- 接口返回 code: 40003（考试时间窗口错误）
- 或返回 500（GA 算法超时）

### 排查

```sql
-- 检查考试时间窗口
SELECT id, title, start_time, end_time, duration, status FROM ems_exam WHERE id = ?;

-- 检查用户是否已开始（唯一索引）
SELECT * FROM ems_exam_record WHERE exam_id = ? AND user_id = ?;

-- 检查试卷是否发布
SELECT id, status FROM ems_paper WHERE id = ?;
```

### 修复

| 场景 | 操作 |
|------|------|
| 时区错误 | 服务器统一 TZ=Asia/Shanghai，前端 dayjs timezone |
| GA 算法超时 | 临时切换 strategy=2（随机），记入工单 |
| 试卷快照缺失 | 重跑迁移脚本 make migrate-up |

---

## 4️⃣ 答题进度丢失

### 症状
- 用户刷新页面后，已答题目丢失
- Redis 中 koala:exam:progress:{record_id} 不存在

### 排查

```bash
# 1. Redis 健康
redis-cli -h redis cluster info | grep cluster_known_nodes

# 2. 是否有大 key / 热 key
redis-cli --hotkeys -h redis

# 3. 看应用日志是否抛 Redis 异常
kubectl logs ... | grep -i redis
```

### 修复

```bash
# 临时：恢复 MySQL 答题数据 → 写回 Redis
mysql -e "SELECT * FROM ems_exam_answer WHERE record_id = ?"

# 长期：
# - 检查 Netpoll 连接复用
# - 升级 Redis 客户端连接池
# - 前端双写 LocalStorage
```

---

## 5️⃣ 阅卷卡顿 / 错题未入库

### 症状
- /exams/submit 后长时间不出成绩
- ems_wrong_log 无新数据
- 告警 koala_grading_queue_size > 5000

### 排查

```bash
# 1. 检查 Redis Stream 队列
redis-cli XLEN koala:stream:grading
redis-cli XINFO GROUPS koala:stream:grading

# 2. Worker 是否在跑
kubectl get pods -l app=koala-grading-worker

# 3. 死信队列
redis-cli XLEN koala:stream:grading:dlq
```

### 修复

```bash
# 1. 扩容 Worker（临时）
kubectl scale deployment koala-grading-worker --replicas=5

# 2. 重放死信（人工确认后）
redis-cli XRANGE koala:stream:grading:dlq - +

# 3. 若 Worker 全部宕机：重启 + 跳过失败任务
kubectl rollout restart deployment/koala-grading-worker
```

---

## 6️⃣ 数据库连接池耗尽

### 症状
Error 1040: Too many connections / connection refused

### 应急

```sql
-- 1. 查看连接数
SHOW VARIABLES LIKE 'max_connections';
SHOW STATUS LIKE 'Threads_connected';

-- 2. 找出慢查询
SELECT * FROM information_schema.processlist
 WHERE COMMAND != 'Sleep' AND TIME > 5 ORDER BY TIME DESC;

-- 3. 杀掉长事务（请先确认影响）
KILL <id>;
```

```yaml
# 调整 GORM 连接池（永久修复）
mysql:
  max_open_conns: 300
  max_idle_conns: 80
  conn_max_lifetime: 1800
```

---

## 7️⃣ 缓存雪崩 / 击穿

### 症状
- 接口 P99 突然飙升
- Redis 大量 GET 返回 nil

### 应急

```bash
# 1. 开启本地缓存兜底（Caffeine）
# 设置 APP_LOCAL_CACHE_ENABLED=true

# 2. 临时延长热点 key TTL
redis-cli EXPIRE koala:question:hot:list 86400

# 3. 加分布式锁防止击穿
redis-cli SET koala:lock:question:list:1 1 NX EX 5
```

---

## 8️⃣ 安全事件

### SQL 注入 / XSS 攻击

1. WAF 已自动拦截（99% 场景无需人工）
2. 异常高频 IP：拉入 Nginx deny 名单
3. 审计日志查 koala_security_events_total

### 数据泄露疑似

1. 立即通知 CSO / 法务
2. 启动数据泄露应急预案（保留证据）
3. 详见 0012_SECURITY.md

---

## 🔧 常用排查命令

```bash
# 后端实时日志
kubectl logs -f -n koala -l app=koala-exam-backend --tail=100

# 进入 Pod 调试
kubectl exec -it -n koala <pod-name> -- /bin/sh

# 慢查询分析
pt-query-digest /var/log/mysql/slow.log

# 火焰图（性能问题）
go tool pprof http://backend:6060/debug/pprof/profile?seconds=30

# 链路追踪
jaeger-ui → service: koala-exam-backend → 找 P99 高的 trace
```

---

## 📞 升级路径

```
自助排查 (本文档)
    ↓ 不解决
企业微信群 @On-call
    ↓ 不解决
拉群：后端 + DBA + SRE
    ↓ 不解决
升级至架构师 / CTO
```

---

## 📚 事后总结模板（RCA）

```markdown
## 故障概要
- 时间：YYYY-MM-DD HH:MM ~ HH:MM（持续 X 分钟）
- 影响：XX 用户受影响，XX 笔订单（考试）失败
- 等级：P?

## 时间线
- HH:MM 告警触发
- HH:MM On-call 响应
- HH:MM 定位原因
- HH:MM 临时缓解
- HH:MM 完全恢复

## 根因
（详细描述 + 截图 + 日志链接）

## 改进措施
- [ ] TODO-1 (P0, 负责人)
- [ ] TODO-2 (P1, 负责人)
```

---

> 最后更新：每次 P0/P1 故障后由 On-call 同步更新本手册。