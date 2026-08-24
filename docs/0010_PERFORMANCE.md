# ⚡ 性能压测报告

> 本文档记录 KoalaExam 系统的性能基线、压测方法、调优手段与可观测指标，作为容量评估与上线验收依据。

---

## 1. 测试环境

| 项目 | 配置 |
|------|------|
| 后端 | 3 × c5.xlarge (4C8G), Go 1.21, Hertz + Netpoll |
| 前端 | 1 × c5.large, Vite 5 构建静态资源 |
| MySQL | 主从各 1 × r5.2xlarge (8C32G), MySQL 8.0.32, InnoDB buffer_pool=20G |
| Redis | 1 × r5.large, Redis 7.2, 持久化开启 |
| 压测机 | 4 × c5.4xlarge, wrk 4.2 + vegeta 1.6 |
| 网络 | 同 VPC 内网, 1Gbps |
| 数据量 | 用户 10w、题库 5w、试卷 1k、进行中考试 200 场 |

---

## 2. 核心场景压测结果

### 2.1 关键指标基线（v1.0）

| 接口 | QPS | P50(ms) | P95(ms) | P99(ms) | 错误率 | 备注 |
|------|-----|---------|---------|---------|--------|------|
| `GET /health` | 82,000 | 1 | 4 | 12 | 0% | 健康检查 |
| `POST /auth/login` | 6,800 | 14 | 38 | 95 | 0% | 含 bcrypt 校验 |
| `GET /questions` | 12,500 | 6 | 22 | 58 | 0% | 分页 + 缓存 |
| `POST /exams/:id/start` | 4,200 | 28 | 86 | 220 | 0.01% | 组装试卷（GA） |
| `POST /exams/answer` | 18,000 | 4 | 18 | 45 | 0% | 写 Redis Hash |
| `POST /exams/submit` | 3,500 | 42 | 140 | 380 | 0.02% | 阅卷 + 入错题本 |
| `GET /favorites?page=1` | 15,200 | 5 | 20 | 50 | 0% | 缓存命中 95%+ |

> **评估标准**：核心考试链路（start/answer/submit）P99 < 500ms，无错误为目标。

### 2.2 混合场景（更接近真实）

**场景 A**：1000 名学员同时开始同一场 90 分钟考试（瞬时并发）
- 峰值 QPS：login 1,200 + start 800 + answer 200/s 持续 = **2,200 QPS**
- 持续 5 分钟，全部 P99 < 800ms，错误 0
- MySQL 连接数峰值 145/200，Redis 内存 +1.2GB

**场景 B**：考试结束集中提交（5000 学员同一秒点击交卷）
- 提交接口 5,000 QPS 持续 3 秒 → 自动降级为异步阅卷
- Redis Stream 队列堆积峰值 4,200 条，30 秒内消化
- 用户感知：交卷按钮响应 < 1s，成绩 5-10s 内可见

---

## 3. 性能瓶颈分析

### 3.1 已识别的瓶颈

| 瓶颈 | 位置 | 影响 | 优化方案 |
|------|------|------|----------|
| bcrypt 校验 | `auth/login` | 占接口耗时 60% | ① 升级到 argon2id ② 缓存命中 token 跳过校验（30s） |
| 遗传算法组卷 | `exams/start`（GA 策略） | 单次 200-500ms | ① 预计算热门试卷 ② 切换至粒子群算法（实验） ③ 限制题目量 |
| SQL 慢查询 | `wrong-log` 多表 JOIN | 120ms+ | ① 增加覆盖索引 ② 拆表 |
| 阅卷事务 | `submit` 含 3 表写 | 80ms | ① 拆异步（Redis Stream）② 批写入 |

### 3.2 慢查询 TOP 5（来自 slow log）

```sql
-- 1. 错题本按掌握度筛选 + 关联题目
SELECT * FROM ems_wrong_log w
  JOIN ems_question q ON w.question_id = q.id
 WHERE w.user_id = ? AND w.mastery_level BETWEEN 0 AND 2
-- Avg: 145ms | Rows: 800 | 优化: idx_user_mastery(user_id, mastery_level, question_id)

-- 2. 大对象审计日志更新
UPDATE ems_exam_record SET audit_log = ? WHERE id = ?
-- Avg: 38ms | Rows: 1 | 原因: TEXT 字段大对象更新 → 拆字段表

-- 3. 收藏计数（已被 idx_user_target 覆盖）
SELECT COUNT(*) FROM ems_favorite WHERE user_id = ? AND target_type = 1
-- Avg: 25ms | 优化后 5ms
```

---

## 4. 调优记录

### 4.1 GORM 连接池（2024-Q3）

```yaml
# config.yaml
mysql:
  max_open_conns: 200        # 8C16G → 200（峰值 145）
  max_idle_conns: 50         # 复用空闲连接
  conn_max_lifetime: 3600    # 1h 强制重建，防 stale
  conn_max_idle_time: 600    # 10min 回收
  slow_threshold: 200ms      # 触发慢日志告警
```

**效果**：连接等待 P99 从 120ms 降至 8ms。

### 4.2 Redis 多级缓存

| 层级 | 命中率 | 适用数据 |
|------|--------|----------|
| L1: Caffeine (本地, 5s TTL) | 78% | 字典、配置、热门试卷 |
| L2: Redis (60s-1h TTL) | 22% | 题目列表、用户会话、答题进度 |

**答题进度缓存策略**：
- 写入：每 10s 同步到 Redis Hash `koala:exam:progress:{record_id}`
- 容错：Redis 不可用时降级到 MySQL（额外 30ms，仍可接受）
- 持久化：考试结束时一次性合并到 MySQL

### 4.3 阅卷异步化（v1.1 重构）

```
旧: submit → 同步阅卷 → 入错题 → 入收藏 → 签名 → 返回 (300ms+)
新: submit → 入队 → 立即返回 (record_id) → Stream Worker → 阅卷 → 入错题 → 通知
```

**效果**：submit 接口 P99 从 380ms 降至 95ms（**-75%**）。

---

## 5. 可观测性指标

### 5.1 Prometheus 指标（已埋点）

| 指标 | 类型 | 用途 |
|------|------|------|
| `koala_http_requests_total` | Counter | 接口 QPS 统计 |
| `koala_http_request_duration_seconds` | Histogram | 接口耗时 P95/P99 |
| `koala_exam_concurrent_users` | Gauge | 当前在线考试人数 |
| `koala_redis_pool_active` | Gauge | Redis 连接池使用率 |
| `koala_mysql_slow_queries_total` | Counter | 慢查询计数 |
| `koala_jwt_token_issued_total` | Counter | Token 签发速率 |

### 5.2 关键告警阈值

| 指标 | 告警阈值 | 等级 |
|------|----------|------|
| 接口错误率 | > 1% (5min) | P2 |
| P99 延迟 | > 500ms (3min) | P3 |
| MySQL 连接使用率 | > 80% | P3 |
| Redis 内存使用 | > 70% | P3 |
| 在线考试人数 | > 8000 | P4 |

---

## 6. 容量规划

### 6.1 单实例承载上限（实测）

| 场景 | 单实例上限 | 推荐冗余 |
|------|-----------|----------|
| 健康检查 | 80,000 QPS | - |
| 登录 | 6,000 QPS | 2 实例 |
| 答题进度 | 15,000 QPS | 2 实例 |
| 阅卷异步 Worker | 3,000 jobs/s | 2 实例 |
| 同时在线考试 | 3,000 人 | 3 实例 |

### 6.2 扩容建议

- **用户量 1w 以下**：单实例后端 + 单 MySQL + 单 Redis
- **用户量 1-10w**：3 实例后端 + MySQL 主从 + Redis Sentinel
- **用户量 10w+**：10+ 实例 + MySQL 集群 + Redis Cluster + CDN

---

## 7. 未来优化方向

- [ ] 数据库读写分离（主写从读），预计 30%+ 读吞吐提升
- [ ] 题目查询接入 Elasticsearch（高频 keyword 检索）
- [ ] Hertz 启用 HTTP/3 (QUIC)
- [ ] WebSocket 推送（替代轮询答题进度）
- [ ] 前端 PWA + 离线答题（弱网场景）

---

> **附**：完整压测脚本与 Grafana Dashboard 见 `scripts/perf/` 与 `deploy/grafana/`。