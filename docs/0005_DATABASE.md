# 🗄 数据库设计

> 描述 KoalaExam 的数据库 schema 设计、ER 关系、索引策略、性能优化与运维规范。

---

## 1. 设计原则

1. **统一前缀**：所有业务表前缀 `ems_` (Exam Management System)
2. **逻辑删除**：使用 `deleted_at DATETIME DEFAULT NULL` 实现软删除
3. **时间戳**：每张表都包含 `created_at` / `updated_at`
4. **字符集**：统一 `utf8mb4_unicode_ci`（支持 emoji + 中英混排）
5. **存储引擎**：InnoDB（事务、行级锁）
6. **主键**：自增 BIGINT（预估 50 年内不溢出）
7. **外键**：应用层维护一致性（MySQL 外键性能损耗大）

---

## 2. 表清单（核心 12 张）

| 表名 | 中文 | 行数预估 | 关键索引 |
|------|------|----------|----------|
| ems_user | 用户 | 100w | uniq_username, idx_role, idx_status |
| ems_department | 组织/院系 | 1w | idx_parent |
| ems_class | 班级 | 10w | idx_dept, idx_teacher |
| ems_question_category | 题库分类 | 1k | idx_parent |
| ems_question | 题目 | 50w | idx_category, idx_type, idx_difficulty |
| ems_paper | 试卷 | 5w | idx_creator, idx_status |
| ems_paper_question | 试卷-题目关联 | 500w | uniq_paper_q |
| ems_exam | 考试 | 10w | idx_paper, idx_status, idx_start, idx_end |
| ems_exam_record | 考试记录 | 1亿 | uniq_exam_user, idx_status |
| ems_favorite_folder | 收藏夹 | 100w | idx_user |
| ems_favorite | 收藏主表 | 1亿 | uniq_user_target |
| ems_wrong_log | 错题日志 | 500w | idx_user_q, idx_last, idx_mastery |

---

## 3. ER 关系图

```
                                ┌─────────────────┐
                                │ ems_department  │
                                │  (组织/院系)     │
                                └─────────────────┘
                                         │ 1:N
                                         ▼
┌──────────┐  1:N  ┌──────────┐  1:N  ┌──────────┐
│ ems_user ├──────►│ ems_class│◄─────┤ ems_user │
│  (用户)  │       │  (班级)  │       │ (班主任)  │
└──────────┘       └──────────┘       └──────────┘
     │ 1:N                              │
     ▼                                  │
┌──────────┐  1:N  ┌──────────┐  N:M  ┌──────────┐
│ ems_     │       │ ems_     │◄─────►│ ems_     │
│exam      │       │ paper    │       │ question │
│(考试)    │       │(试卷)    │       │(题目)    │
└──────────┘       └──────────┘       └──────────┘
     │ 1:N                              │
     ▼                                  │
┌──────────┐                       ┌──────────┐
│ ems_     │  N:1                  │ ems_     │
│exam_     ├──────────────►        │favorite_ │
│record    │                       │folder    │
│(考试记录)│                       │(收藏夹)  │
└──────────┘                       └──────────┘
     │                                  │ 1:N
     ▼                                  ▼
┌──────────┐                       ┌──────────┐
│ ems_     │                       │ ems_     │
│wrong_log │                       │favorite  │
│(错题)    │                       │(收藏)    │
└──────────┘                       └──────────┘
```

---

## 4. 核心表 DDL

### 4.1 用户表 ems_user

```sql
CREATE TABLE ems_user (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(32) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(64),
  email VARCHAR(128),
  phone VARCHAR(20),
  avatar VARCHAR(255),
  role TINYINT NOT NULL DEFAULT 3,
  department_id BIGINT,
  class_id BIGINT,
  status TINYINT NOT NULL DEFAULT 1,
  allow_multi_device TINYINT DEFAULT 1,
  last_login_at DATETIME,
  last_login_ip VARCHAR(45),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  UNIQUE KEY uniq_username (username, deleted_at),
  KEY idx_role (role, status),
  KEY idx_dept (department_id, role),
  KEY idx_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.2 题目表 ems_question

```sql
CREATE TABLE ems_question (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  category_id BIGINT NOT NULL,
  type TINYINT NOT NULL,
  difficulty TINYINT NOT NULL DEFAULT 1,
  title TEXT NOT NULL,
  options JSON,
  correct_answer TEXT NOT NULL,
  analysis TEXT,
  tags JSON,
  knowledge_points JSON,
  score DECIMAL(6,2) NOT NULL DEFAULT 1.00,
  creator_id BIGINT NOT NULL,
  reference_count INT DEFAULT 0,
  correct_rate DECIMAL(5,4) DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  KEY idx_category (category_id, status, deleted_at),
  KEY idx_type (type, difficulty),
  KEY idx_creator (creator_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.3 试卷表 ems_paper + 关联表

```sql
CREATE TABLE ems_paper (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(128) NOT NULL,
  strategy TINYINT NOT NULL,
  total_score DECIMAL(8,2) NOT NULL,
  pass_score DECIMAL(8,2) NOT NULL,
  duration INT NOT NULL,
  config_rule JSON,
  question_count INT NOT NULL,
  creator_id BIGINT NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  KEY idx_creator (creator_id, status),
  KEY idx_strategy (strategy)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE ems_paper_question (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  paper_id BIGINT NOT NULL,
  question_id BIGINT NOT NULL,
  section VARCHAR(32),
  sort INT NOT NULL DEFAULT 0,
  score DECIMAL(6,2) NOT NULL,
  UNIQUE KEY uniq_paper_q (paper_id, question_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.4 考试表 ems_exam + 考试记录 ems_exam_record

```sql
CREATE TABLE ems_exam (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  title VARCHAR(128) NOT NULL,
  paper_id BIGINT NOT NULL,
  start_time DATETIME NOT NULL,
  end_time DATETIME NOT NULL,
  duration INT NOT NULL,
  shuffle_questions TINYINT DEFAULT 1,
  shuffle_options TINYINT DEFAULT 1,
  anti_cheat JSON,
  target_type TINYINT NOT NULL,
  target_ids JSON,
  max_tab_switch_count INT DEFAULT 3,
  enable_fullscreen TINYINT DEFAULT 1,
  enable_copy_paste_block TINYINT DEFAULT 1,
  status TINYINT NOT NULL DEFAULT 1,
  creator_id BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  KEY idx_paper (paper_id, status),
  KEY idx_time (start_time, end_time, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE ems_exam_record (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  exam_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  paper_snapshot JSON NOT NULL,
  objective_score DECIMAL(8,2) DEFAULT 0,
  subjective_score DECIMAL(8,2) DEFAULT 0,
  total_score DECIMAL(8,2) DEFAULT 0,
  pass_score DECIMAL(8,2),
  passed TINYINT DEFAULT 0,
  score_hash VARCHAR(128),
  audit_log JSON,
  start_time DATETIME NOT NULL,
  submit_time DATETIME,
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  UNIQUE KEY uniq_exam_user (exam_id, user_id, deleted_at),
  KEY idx_status (status, submit_time),
  KEY idx_user (user_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.5 收藏 + 错题（核心卖点）

```sql
CREATE TABLE ems_favorite_folder (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
  color VARCHAR(16),
  icon VARCHAR(32),
  sort INT DEFAULT 0,
  is_system TINYINT DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  KEY idx_user (user_id, deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE ems_favorite (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  target_type TINYINT NOT NULL,
  target_id BIGINT NOT NULL,
  source_type TINYINT NOT NULL,
  folder_id BIGINT,
  note VARCHAR(500),
  wrong_log_id BIGINT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  UNIQUE KEY uniq_user_target (user_id, target_type, target_id, deleted_at),
  KEY idx_folder (folder_id, target_type),
  KEY idx_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE ems_wrong_log (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  question_id BIGINT NOT NULL,
  exam_record_id BIGINT,
  user_answer TEXT,
  correct_answer TEXT,
  wrong_count INT DEFAULT 1,
  review_count INT DEFAULT 0,
  correct_count INT DEFAULT 0,
  mastery_level TINYINT DEFAULT 1,
  last_wrong_at DATETIME,
  last_review_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME DEFAULT NULL,
  UNIQUE KEY uniq_user_q (user_id, question_id),
  KEY idx_user_mastery (user_id, mastery_level),
  KEY idx_last_wrong (last_wrong_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## 5. 关键设计

### 5.1 多态收藏 (ems_favorite)

```sql
UNIQUE INDEX uniq_user_target (user_id, target_type, target_id, deleted_at)
```

- `target_type`: 1=题目, 2=试卷, 3=知识点
- 删除后通过 `deleted_at` 区分，可恢复
- 应用层统一接口 `POST /favorites/toggle`

### 5.2 错题日志 (ems_wrong_log)

记录每一次错误，累积错题次数与掌握度。

```sql
KEY idx_user_q (user_id, question_id)
KEY idx_user_mastery (user_id, mastery_level)
KEY idx_last_wrong (last_wrong_at)
```

### 5.3 考试记录 (ems_exam_record)

- `paper_snapshot` JSON 字段：试卷快照，避免题目被删影响
- `score_hash` VARCHAR(128)：SHA-256 签名防篡改
- `audit_log` TEXT：行为审计 JSON（切屏/复制粘贴）
- UNIQUE INDEX `uniq_exam_user`：同一考试同一用户只能一条记录

### 5.4 试卷快照设计权衡

| 方案 | 优点 | 缺点 | 选型 |
|------|------|------|------|
| 存外键（实时查题目） | 题目可更新 | 题目改了影响考试 | ❌ |
| 存完整 JSON 快照 | 考试独立、可恢复 | 占用空间 | ✅ |
| 双写（外键+快照） | 都可查 | 写入放大 | ⚠️ 大型考试考虑 |

---

## 6. 索引策略

### 6.1 通用原则

1. 高频查询字段必建索引：role/status/type/difficulty/exam_id/user_id
2. 组合索引优于单列：`(user_id, question_id)`、`(user_id, target_type, target_id)`
3. 唯一索引保证一致性：username、user+target、exam+user
4. 覆盖索引避免回表：`(user_id, mastery_level, ...)`
5. 前缀索引：长 VARCHAR 字段（如 title）前 32 字符索引

### 6.2 索引清单

```sql
-- ems_question
ALTER TABLE ems_question ADD INDEX idx_search (category_id, type, difficulty, deleted_at);

-- ems_exam_record
ALTER TABLE ems_exam_record ADD INDEX idx_user_time (user_id, start_time DESC);
ALTER TABLE ems_exam_record ADD INDEX idx_exam_status (exam_id, status, total_score);

-- ems_wrong_log
ALTER TABLE ems_wrong_log ADD INDEX idx_user_review (user_id, last_review_at);
```

### 6.3 反例：避免过度索引

```
❌ 每列都加索引 → 写入性能下降
❌ 索引顺序不当 → 无法命中
❌ 函数索引 → MySQL 8 才支持，且易失效
```

---

## 7. 性能优化

### 7.1 大表分区

ems_exam_record 按 created_at 月分区：

```sql
ALTER TABLE ems_exam_record PARTITION BY RANGE (TO_DAYS(created_at)) (
  PARTITION p202410 VALUES LESS THAN (TO_DAYS(2024-11-01)),
  PARTITION p202411 VALUES LESS THAN (TO_DAYS(2024-12-01)),
  PARTITION p202412 VALUES LESS THAN (TO_DAYS(2025-01-01)),
  PARTITION pmax    VALUES LESS THAN MAXVALUE
);
```

收益：
- 单次扫描只命中部分分区
- 历史数据可独立 DROP（无需 DELETE）

### 7.2 冷热数据分离

```sql
-- 6 个月前的考试记录归档到 history_db
INSERT INTO history_db.ems_exam_record
SELECT * FROM ems_exam_record
WHERE created_at < DATE_SUB(NOW(), INTERVAL 6 MONTH);
```

### 7.3 读写分离

```
主库（写）：ems-master → INSERT / UPDATE / DELETE
从库（读）：ems-slave-1, ems-slave-2 → SELECT

GORM 配置：
  dsn_master: user:pass@tcp(master:3306)/koala_exam?parseTime=true
  dsn_slave:  user:pass@tcp(slave:3306)/koala_exam?parseTime=true

应用层：
  db.Master().Create(...)  // 写
  db.Slave().Find(...)     // 读

延迟监控：从库延迟 > 1s 告警
```

### 7.4 分库分表（ems_wrong_log）

```
按 user_id % 8 分 8 张表：
  ems_wrong_log_0, ems_wrong_log_1, ..., ems_wrong_log_7

实现：ShardingSphere / 自研路由

路由规则：
  table = ems_wrong_log + "_" + (user_id % 8)

合并查询（业务层 UNION ALL）：
  SELECT * FROM ems_wrong_log_0 WHERE user_id = ?
  UNION ALL
  SELECT * FROM ems_wrong_log_1 WHERE user_id = ?
```

### 7.5 慢查询治理

```sql
-- 开启慢日志
SET GLOBAL slow_query_log = ON;
SET GLOBAL long_query_time = 0.2;  -- 200ms

-- 分析工具
pt-query-digest /var/log/mysql/slow.log > report.txt

-- EXPLAIN 看执行计划
EXPLAIN SELECT * FROM ems_wrong_log WHERE user_id = 100 AND mastery_level = 1;
-- type: ref / rows: 50 / Extra: Using index
```

---

## 8. 备份与恢复

### 8.1 备份策略

```
每日全量（凌晨 2 点）：mysqldump → 对象存储
每小时增量：binlog 备份
每 15 分钟：Redis AOF + RDB

保留：
  全量：30 天
  binlog：7 天
  Redis：3 天

加密：AES-256-CBC 加密后上传
异地：跨 Region 复制（容灾）
```

### 8.2 恢复演练

每月抽样恢复 1 次（季度全量演练），验证：
- 数据完整性（checksum 比对）
- 应用兼容（导入测试库跑回归）
- 恢复时间 < 30 分钟（10GB 数据）

### 8.3 PITR（Point In Time Recovery）

```bash
# 恢复到任意秒（v1.1 计划）
mysqlbinlog --stop-datetime="2024-12-15 14:30:00" \
  /var/lib/mysql/binlog.000123 | mysql -u root -p koala_exam
```

---

## 9. 安全合规

### 9.1 字段加密

```sql
-- 用户手机号 / 身份证号（敏感）
ALTER TABLE ems_user ADD COLUMN phone_encrypted VARBINARY(255);
-- 应用层：AES-256-GCM 加密后存储
```

### 9.2 数据脱敏

列表展示：`138****5678` / `110101********1234`
导出：可申请导出（留痕），二次审批

### 9.3 审计

所有 DDL / 敏感表 DML 写入 `ems_audit_log`（详见 [0012_SECURITY.md](0012_SECURITY.md)）

---

## 10. Schema 演进

### 10.1 迁移工具

推荐工具：
- [golang-migrate/migrate](https://github.com/golang-migrate/migrate)
- [pressly/goose](https://github.com/pressly/goose)

```sql
-- migrations/0001_init.up.sql
CREATE TABLE ...;

-- migrations/0001_init.down.sql
DROP TABLE ...;

-- 顺序：
0001_init.up.sql
0002_add_question_knowledge.up.sql
0003_add_exam_anti_cheat.up.sql
...

每次表结构变更必须有 up + down
```

### 10.2 演进规则

1. 向后兼容：新字段可加 NOT NULL DEFAULT
2. 不可逆变更：拆分表 / 删字段 → 数据迁移 + 双写
3. 大表 DDL：使用 `pt-online-schema-change`
4. 索引：先用 `ALTER TABLE ... ADD INDEX CONCURRENTLY`（PG）或 `pt-osc`（MySQL）

---

> 相关文档：[0002 架构](0002_ARCHITECTURE.md) · [0010 性能](0010_PERFORMANCE.md) · [0012 安全](0012_SECURITY.md)