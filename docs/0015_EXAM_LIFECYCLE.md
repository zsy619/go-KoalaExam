# 🎯 考试全生命周期（Exam Lifecycle）

> 深入剖析 KoalaExam 考试模块的完整生命周期：状态机、断线续考、重考机制、答题保存策略、阅卷流水线与错题本联动。
> 本文档基于 internal/domain/entity/exam.go、internal/application/exam/exam_app.go、internal/application/grading/grading_app.go 等真实代码实现。

---

## 1. 核心实体

### 1.1 考试表 ke_exam

```go
type Exam struct {
    ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    Title         string    `gorm:"size:128;not null" json:"title"`
    Description   string    `gorm:"type:text" json:"description"`
    PaperID       int64     `gorm:"index;not null" json:"paper_id"`
    StartTime     time.Time `gorm:"index" json:"start_time"`
    EndTime       time.Time `gorm:"index" json:"end_time"`
    Duration      int       `gorm:"default:60" json:"duration"`    // 考试时长（分钟）
    MaxAttempts   int       `gorm:"default:1" json:"max_attempts"` // 允许重考次数
    ShuffleQ      bool      `gorm:"default:true" json:"shuffle_q"`
    ShuffleOpt    bool      `gorm:"default:true" json:"shuffle_opt"`
    AntiCheat     bool      `gorm:"default:true" json:"anti_cheat"`
    Status        int8      `gorm:"default:1;index" json:"status"` // 0未发布 1进行中 2已结束
    CreatorID     int64     `gorm:"index" json:"creator_id"`
    TargetUsers   string    `gorm:"type:text" json:"target_users"`   // JSON 数组
    TargetClasses string    `gorm:"type:text" json:"target_classes"` // JSON 数组
}
```

### 1.2 考试记录表 ke_exam_record

```go
type ExamRecord struct {
    ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
    ExamID          int64      `gorm:"index:idx_exam_user,unique;not null" json:"exam_id"`
    UserID          int64      `gorm:"index:idx_exam_user,unique;not null" json:"user_id"`
    PaperSnapshot   string     `gorm:"type:json" json:"paper_snapshot"`  // 试卷快照
    Answers         string     `gorm:"type:json" json:"answers"`         // 答题内容JSON
    Status          int8       `gorm:"default:0;index" json:"status"`    // 0进行中 1已交卷 2已批改 3异常
    StartTime       time.Time  `json:"start_time"`
    SubmitTime      *time.Time `json:"submit_time"`
    Duration        int        `gorm:"default:0" json:"duration"`        // 实际用时（秒）
    TotalScore      float64    `gorm:"default:0" json:"total_score"`
    ObjectiveScore  float64    `gorm:"default:0" json:"objective_score"`
    SubjectiveScore float64    `gorm:"default:0" json:"subjective_score"`
    Passed          bool       `gorm:"default:false" json:"passed"`
    ScoreHash       string     `gorm:"size:128" json:"score_hash"`       // SHA-256 签名
    TabSwitchCnt    int        `gorm:"default:0" json:"tab_switch_cnt"`  // 切屏次数
    AuditLog        string     `gorm:"type:text" json:"audit_log"`       // 行为审计 JSON
}
```

**关键约束**：uniq_exam_user 索引 (exam_id, user_id) 唯一 + 软删除，强制一人一考一记录。

---

## 2. 考试状态机

### 2.1 考试状态（ke_exam.status）

```
  ┌─────────┐  创建        ┌─────────────┐  StartTime 到达    ┌──────────┐
  │  草稿    │ ──────────► │  未发布(0)   │ ────────────────► │ 进行中(1) │
  │ (Draft) │             │  (Pending)  │                  │ (Running) │
  └─────────┘             └─────────────┘                  └──────────┘
                                                                      │
                                                                      │ EndTime 到达 / 手动停止
                                                                      ▼
                                                              ┌──────────┐
                                                              │ 已结束(2) │
                                                              │ (Ended)  │
                                                              └──────────┘
```

### 2.2 考试记录状态（ke_exam_record.status）

```
  ┌──────────┐  POST /exams/submit    ┌──────────┐  AutoGrade 完成    ┌──────────┐
  │ 进行中(0) │ ────────────────────► │ 已交卷(1) │ ─────────────────► │ 已批改(2) │
  │ (Ongoing) │                       │(Submitted)│                    │ (Graded) │
  └──────────┘                       └──────────┘                    └──────────┘
       ▲                                                                  │
       │ 断线续考（record 已存在且 Status==Ongoing）                          │
       └──────────────────────────────────────────────────────────────────┘

  异常(3): 阅卷失败 / 超时未交卷 / 系统崩溃恢复
```

### 2.3 状态流转规则

| 当前状态 | 触发动作 | 目标状态 | 守卫条件 |
|----------|----------|----------|----------|
| 草稿 | CreateExam | 未发布 | 试卷已发布 |
| 未发布 | 时间到达 start_time | 进行中 | cron 任务 / 首次访问时检查 |
| 进行中 | 时间到达 end_time | 已结束 | cron 任务 |
| 进行中 | 手动 ArchiveExam | 已结束 | 超管权限 |
| (record) Ongoing | 提交 | Submitted | 唯一索引 + 状态校验 |
| (record) Submitted | AutoGrade | Graded | 阅卷成功 |
| (record) Ongoing | 断线续考 | Ongoing | record 已存在 + Status==Ongoing |
| (record) Submitted | 重新进入 | 报错 | CodeExamSubmitted |

**当前实现说明**：
- CreateExam 直接写入 Status = Running（跳过 Pending 状态）—— 简化逻辑，建议 v1.1 增加 Pending 审核流
- 状态推进依赖 cron 任务，详见 §9「定时任务」

---

## 3. 考试创建流程

### 3.1 教师创建考试

```
POST /api/v1/exams
  ↓
ExamHandler.CreateExam
  ↓
ExamApp.CreateExam(req, creatorID)
  ├─ 1. 解析时间字符串（RFC3339）
  ├─ 2. JSON 序列化 TargetUsers / TargetClasses
  ├─ 3. 构造 Exam 实体
  ├─ 4. 默认 Status = Running（当前实现跳过 Pending）
  └─ 5. examRepo.Create(ex) → 返回 ex.ID
```

### 3.2 DTO 定义（exam_dto.go）

```go
type CreateExamReq struct {
    Title         string  `json:"title" binding:"required"`
    Description   string  `json:"description"`
    PaperID       int64   `json:"paper_id" binding:"required"`
    StartTime     string  `json:"start_time" binding:"required"` // RFC3339
    EndTime       string  `json:"end_time" binding:"required"`
    Duration      int     `json:"duration" binding:"required"`
    MaxAttempts   int     `json:"max_attempts"`;                // 默认 1
    ShuffleQ      bool    `json:"shuffle_q"`
    ShuffleOpt    bool    `json:"shuffle_opt"`
    AntiCheat     bool    `json:"anti_cheat"`
    TargetUsers   []int64 `json:"target_users"`
    TargetClasses []int64 `json:"target_classes"`
}
```

### 3.3 目标分配（多维度）

```sql
-- 表中存的是 JSON 字符串
TargetUsers   = [1,2,3,5,8]                -- 指定用户 ID
TargetClasses = [10,11]                   -- 班级 ID（自动展开为班级成员）

-- 学员查询可用考试
SELECT * FROM ke_exam
WHERE status = 1
  AND (
    JSON_CONTAINS(target_users, :uid)         -- 用户在目标列表
    OR JSON_CONTAINS(target_classes, :class_id) -- 用户班级在目标列表
    OR target_users = [] AND target_classes = []  -- 公开考试
  )
  AND start_time <= NOW() AND end_time >= NOW();
```

> v1.1 计划：拆为独立表 ke_exam_target (exam_id, target_type, target_id)，便于索引与查询。

---

## 4. 开始考试（核心流程）

### 4.1 完整时序图

```
学员           Frontend           Hertz         ExamApp            Repo          Redis
  │                │                │              │                │              │
  │ 点击 [开始考试] │                │              │                │              │
  ├───────────────►│ POST /exams/:id/start       │              │              │
  │                ├───────────────►│              │              │              │
  │                │                │ JWT middleware 解析 uid          │              │
  │                │                ├──────────────►│              │              │
  │                │                │              │ GetByID(examID)│              │
  │                │                │              ├──────────────►│              │
  │                │                │              │ 校验时间窗口 [start, end]    │              │
  │                │                │              │              │              │
  │                │                │              │ GetByExamAndUser(examID, uid)│              │
  │                │                │              ├──────────────►│              │
  │                │                │              │ rec==nil?     │              │
  │                │                │              │  是：组装试卷 + Create snapshot│              │
  │                │                │              │  否：续考（Status 必须 Ongoing）│              │
  │                │                │              │              │              │
  │                │                │              │ rand.Shuffle(questions)       │              │
  │                │                │              │ 对每题 Shuffle(options)       │              │
  │                │                │              │ 学员端隐藏 answer / analysis │              │
  │                │ 200 OK         │◄─────────────┤ StartExamResp               │              │
  │                │ record_id + questions[]    │              │              │
  │◄───────────────┤                │              │              │              │
  │                │ LocalStorage.setItem(exam_progress) │              │              │
```

### 4.2 关键代码（StartExam）

```go
func (a *ExamApp) StartExam(ctx context.Context, examID, userID int64) (*dto.StartExamResp, error) {
    // 1. 加载考试
    ex, err := a.examRepo.GetByID(ctx, examID)
    if err != nil { return nil, errcode.New(errcode.CodeExamNotExist, "ExamNotExist") }

    // 2. 校验时间窗口
    now := time.Now()
    if now.Before(ex.StartTime) || now.After(ex.EndTime) {
        return nil, errcode.New(errcode.CodeExamNotRunning, "ExamNotRunning")
    }

    // 3. 断线续考检测
    rec, _ := a.recordRepo.GetByExamAndUser(ctx, examID, userID)
    if rec != nil && rec.Status != consts.RecordStatusOngoing {
        return nil, errcode.New(errcode.CodeExamSubmitted, "ExamSubmitted")
    }

    // 4. 首次参加：组装试卷 → 快照
    if rec == nil {
        qs, _ := a.paperApp.Assemble(ctx, ex.PaperID)
        snap, _ := json.Marshal(qs)
        rec = &entity.ExamRecord{
            ExamID:        examID,
            UserID:        userID,
            PaperSnapshot: string(snap),
            Status:        consts.RecordStatusOngoing,
            StartTime:     now,
        }
        a.recordRepo.Create(ctx, rec)
    }

    // 5. 题目乱序 + 选项乱序（学员端隐藏答案）
    var qs []entity.Question
    json.Unmarshal([]byte(rec.PaperSnapshot), &qs)
    if ex.ShuffleQ { rand.Shuffle(len(qs), swap) }
    resps := []dto.QuestionResp{}
    for i := range qs {
        r := question.ToResp(&qs[i], false);  // false = 隐藏答案
        if ex.ShuffleOpt && r.Options != nil {
            rand.Shuffle(len(r.Options), swap)
        }
        resps = append(resps, *r)
    }

    return &dto.StartExamResp{
        ExamID: ex.ID, RecordID: rec.ID, Title: ex.Title,
        Duration: ex.Duration, Questions: resps,
    }, nil
}
```

### 4.3 关键设计点

#### 4.3.1 试卷快照（防篡改）

学员开始考试时，将完整试卷内容（题目、选项、分值）JSON 序列化存入 paper_snapshot。

- 收益：阅卷时不依赖当前题库，题目被修改/删除不影响已开始的考试
- 风险：存储成本（每条记录几 KB），JSON 字段不可建索引
- 优化：阅卷完成后可清理快照（保留 30 天）

#### 4.3.2 题目乱序 + 选项乱序

```go
// 全局 rand（不推荐！v1.1 应使用 crypto/rand + per-exam seed）
rand.Shuffle(len(qs), func(i, j int) { qs[i], qs[j] = qs[j], qs[i] })

// ⚠️ 问题：所有学员共享全局随机状态，多个并发 startExam 时可能产生相关性
// v1.1 修复：使用 crypto/rand 独立种子，或基于 user_id + exam_id 哈希
```

#### 4.3.3 学员端隐藏答案

```go
func ToResp(q *entity.Question, showAnswer bool) *QuestionResp {
    resp := &QuestionResp{ID: q.ID, Type: q.Type, Title: q.Title, Options: q.Options}
    if showAnswer { resp.Answer = q.Answer; resp.Analysis = q.Analysis }
    return resp
}
```

**关键**：学员永远拿不到正确答案，即使翻前端 JS 也不行。

---

## 5. 答题进度保存策略

### 5.1 三级存储架构

```
前端 LocalStorage  ←──→  Redis Hash（4h TTL）  ←──→  MySQL exam_record.answers
  实时缓存（毫秒）       中间层（秒级）                最终持久化（交卷时合并）
```

### 5.2 保存时机

| 触发 | 写入目标 | 频率 | 失败容忍 |
|------|----------|------|----------|
| 用户切换题目 | LocalStorage | 每次 | 不上报 |
| 节流器触发（10s） | Redis HSet | 每 10s 一次 | 降级到 LocalStorage |
| 提交答卷 | MySQL UPDATE | 一次性 | 阻塞至成功 |
| 网络中断 | LocalStorage | 立即 | 重连上报 |

### 5.3 Redis Hash 结构

```
Key:    koala:exam:progress:BLANK
Type:   Hash
TTL:    4 小时（考试时长的 2-4 倍）

Field:  BLANKquestion_id
Value:  BLANK
          "qid": 12345,
          "ans": "B",
          "elapsed": 42,    // 已用秒（防作弊）
          "ts": 1700000000  // 时间戳
        BLANK
```

### 5.4 SaveAnswer 代码

```go
func (a *ExamApp) SaveAnswer(ctx context.Context, req *dto.SaveAnswerReq) error {
    key := cache.KeyExamProgress.Build(req.RecordID)
    data, _ := json.Marshal(map[string]interface{}{
        "qid":     req.QuestionID,
        "ans":     req.Answer,
        "elapsed": req.Elapsed,
        "ts":      time.Now().Unix(),
    })
    if err := a.rdb.HSet(ctx, key, fmt.Sprintf("%d", req.QuestionID), data).Err(); err != nil {
        return err;
    }
    a.rdb.Expire(ctx, key, 4*time.Hour);
    return nil
}

// ⚠️ 注意事项：
// 1. Redis 写入是 fire-and-forget 风格，调用方需自己处理失败重试
// 2. Elapsed 字段记录答题时长，可用于检测异常答题模式（秒答全对 = 疑似作弊）
// 3. 4h TTL 是经验值，>2x duration 即可
```

### 5.5 前端实现（节流 + LocalStorage）

```ts
// composables/useExamProgress.ts
export function useExamProgress(recordId: number, duration: number) {
  const answers = ref(new Map())
  const startTime = Date.now()

  // 1. 切题即写 LocalStorage
  function saveLocal(qid: number, ans: Answer) {
    answers.value.set(qid, ans)
    localStorage.setItem(`exam:BLANKrecordId`, JSON.stringify([...answers.value]))
  }

  // 2. 节流同步到 Redis（10s 一次）
  const throttledSync = throttle(async () => {
    for (const [qid, ans] of answers.value) {
      try {
        await api.saveAnswer({
          record_id: recordId,
          question_id: qid,
          answer: ans,
          elapsed: Math.floor((Date.now() - startTime) / 1000),
        })
      } catch (e) {
        // 失败保留在 LocalStorage，等下次重试
        console.warn('sync failed, will retry', e)
      }
    }
  }, 10_000)

  // 3. 切题时触发
  function onQuestionChange(qid: number, ans: Answer) {
    saveLocal(qid, ans)
    throttledSync()
  }

  // 4. 离开页面时强制同步
  onBeforeUnmount(() => throttledSync.flush())

  return { answers, onQuestionChange }
}
```

### 5.6 断线续考

**场景**：学员答题中刷新页面 / 网络中断 / 切到其他应用。

**恢复流程**：

1. 前端从 LocalStorage 读取 exam:BLANKrecord_id 缓存
2. 调用 POST /exams/:id/start 再次进入（record 已存在 + Status==Ongoing → 续考）
3. 后端反序列化 PaperSnapshot，重新乱序（乱序结果可能与之前不同）
4. 前端用 LocalStorage 答案覆盖默认空答
5. 用户继续答题，10s 后 Redis 会用 LocalStorage 数据兜底

**已知问题**：续考时题目顺序可能变化（依赖 rand.Shuffle）。建议 v1.1 改用确定性随机（基于 user_id+exam_id 的 hash）。

---

## 6. 交卷与阅卷流水线

### 6.1 提交流程（同步）

```go
func (a *ExamApp) SubmitExam(ctx context.Context, recordID int64) (*entity.ExamRecord, error) {
    rec, err := a.recordRepo.GetByID(ctx, recordID)
    if rec.Status != consts.RecordStatusOngoing {
        return nil, errcode.New(errcode.CodeExamSubmitted, "ExamSubmitted")
    }

    // 合并 Redis 答题数据 → MySQL.answers
    key := cache.KeyExamProgress.Build(recordID)
    hash, _ := a.rdb.HGetAll(ctx, key).Result()
    answers := map[string]interface{}{}
    for k, v := range hash {
        var m map[string]interface{}
        _ = json.Unmarshal([]byte(v), &m)
        answers[k] = m["ans"];  // 只取 ans 字段
    }
    b, _ := json.Marshal(answers)
    rec.Answers = string(b)

    now := time.Now()
    rec.SubmitTime = &now
    rec.Status = consts.RecordStatusSubmitted;
    rec.Duration = int(now.Sub(rec.StartTime).Seconds());
    return rec, a.recordRepo.Update(ctx, rec);
}
```

**关键点**：
- 只合并 ans 字段（不持久化 elapsed / ts，这些仅用于防作弊分析）
- Duration 记录实际用时（秒），可用于异常答题检测
- 阅卷触发在 submit 之后（v1.0 同步，v1.1 异步）

### 6.2 阅卷流水线（计划异步化）

```
v1.0 当前：
  POST /exams/submit
    → ExamApp.SubmitExam  （合并 Redis → MySQL）
    → GradingApp.AutoGrade （同步阅卷）
    → FavoriteApp.RecordWrongAnswers （自动入错题本）
    → 200 OK + 成绩
    → 接口耗时 300-500ms（高频提交时易打爆 DB）

v1.1 计划：
  POST /exams/submit
    → ExamApp.SubmitExam   （合并 + 入 Redis Stream）
    → 立即返回 { record_id, status: submitted }
    → 阅卷 Worker 消费 Stream
    → AutoGrade → RecordWrongAnswers → 通知前端
    → 接口耗时 < 100ms，可应对 5000 QPS
```

详见 [0017_GRADING.md](0017_GRADING.md)

### 6.3 阅卷后的成绩签名

```go
func (a *ExamApp) SignScore(rec *entity.ExamRecord) {
    rec.ScoreHash = encrypt.SHA256Hex(
        fmt.Sprintf("%d|%.2f|%.2f|%.2f", rec.ID, rec.TotalScore,
            rec.ObjectiveScore, rec.SubjectiveScore),
        "koala-exam-salt",  // ⚠️ 应从 KMS 注入
    )
}
```

**安全权衡**：
- 防单字段篡改（修改 total_score 会导致 hash 不匹配）
- ⚠️ salt 硬编码在源码中（应改为 KMS 注入）
- ⚠️ 不防整体 record 重放攻击（应加时间戳）
- 教师阅卷后重新签名（前端无法绕过）

v1.1 改进：

```go
payload := fmt.Sprintf("%d|%d|%d|%.2f|%.2f|%.2f|%s|%d",
    recordID, examID, userID,
    totalScore, objectiveScore, subjectiveScore,
    jwtSecret,         // KMS 注入
    time.Now().Unix()) // 时间戳
hash := sha256.Sum256([]byte(payload))
```

---

## 7. 行为审计与防作弊

### 7.1 审计事件类型

| 事件类型 | 触发 | 字段 |
|----------|------|------|
| tab_switch | visibilitychange → hidden | from, to |
| copy_attempt | copy/paste/contextmenu | question_id |
| fullscreen_exit | fullscreenchange | - |
| answer_change | 切换答案 | question_id, new_answer |
| time_warning | 剩余 5 分钟 | remaining_seconds |
| force_submit | 超时 | reason |

### 7.2 AuditEvent 实现

```go
func (a *ExamApp) AuditEvent(ctx context.Context, req *dto.AuditReq) error {
    rec, _ := a.recordRepo.GetByID(ctx, req.RecordID)

    // 累加切屏次数
    if ev, ok := req.Events["type"]; ok && ev == "tab_switch" {
        rec.TabSwitchCnt++;
    }

    // 追加审计日志
    var audit []map[string]interface{};
    if rec.AuditLog != "" { _ = json.Unmarshal([]byte(rec.AuditLog), &audit) }
    audit = append(audit, req.Events)
    b, _ := json.Marshal(audit)
    rec.AuditLog = string(b)
    return a.recordRepo.Update(ctx, rec)
}

// ⚠️ 问题：
// 1. 每次审计都 UPDATE整条 record（包含 audit_log JSON），写放大
// 2. audit_log 无界增长，单个考试可达 MB 级

// v1.1 改进：拆分为独立表 ke_exam_audit_log，批量 INSERT
```

### 7.3 切屏处置策略

```
TabSwitchCnt == 1: 黄色警告（前端 toast）
TabSwitchCnt == 2: 橙色警告 + 写入审计
TabSwitchCnt >= 3: 红色警告 + 教师可见 + 自动收卷（可配置）

配置项（ke_exam 表新增字段，v1.1 计划）：
  max_tab_switch_count INT DEFAULT 3
  auto_submit_on_violation BOOLEAN DEFAULT false
```

详见 [0016_ANTI_CHEAT.md](0016_ANTI_CHEAT.md)

---

## 8. 重考机制

### 8.1 当前实现（v1.0）

**关键字段**：MaxAttempts INT DEFAULT 1

**当前逻辑**：
- uniq_exam_user 索引唯一 → 同一 exam+user 只能有一条 record
- 重考 = 重新创建 record（覆盖原 record? 还是新建？）
- **当前代码未实现重考**（只有 MaxAttempts 字段）

### 8.2 重考流程（v1.1 计划）

```
学员请求重考：
  1. 检查当前已交卷 record 数 vs MaxAttempts
  2. 检查时间窗口（通常重考有独立时间窗）
  3. 删除旧 record（软删除） → 允许 uniq 索引创建新 record
  4. 创建新 record（attempt_no = 旧 attempt_no + 1）
  5. 重新组装试卷（题目随机性可重置）

数据库改动：
  ALTER TABLE ke_exam_record
    ADD COLUMN attempt_no INT DEFAULT 1,
    DROP INDEX uniq_exam_user,
    ADD UNIQUE INDEX uniq_exam_user_attempt (exam_id, user_id, attempt_no);
```

### 8.3 取最高分 vs 取最近分

```sql
-- 取最高分
SELECT MAX(total_score) FROM ke_exam_record WHERE exam_id = ? AND user_id = ?;

-- 取最近分
SELECT * FROM ke_exam_record WHERE exam_id = ? AND user_id = ? ORDER BY attempt_no DESC LIMIT 1;

-- 配置项：ke_exam.score_strategy TINYINT (1=最高 2=最近 3=平均)
```

---

## 9. 定时任务与状态推进

### 9.1 需要的 cron 任务

| 任务 | 频率 | 实现 |
|------|------|------|
| 考试状态推进（Pending→Running） | 1 分钟 | crontab + UPDATE ke_exam SET status=1 WHERE status=0 AND start_time<=NOW() |
| 考试自动结束（Running→Ended） | 1 分钟 | UPDATE ke_exam SET status=2 WHERE status=1 AND end_time<=NOW() |
| 超时未交卷处理 | 5 分钟 | 找出超时 record → 强制 submit → 触发阅卷 |
| Redis 进度清理 | 1 小时 | 扫描 4h+ 未删除的 progress key |
| 阅卷队列监控 | 30 秒 | 检查 Redis Stream 长度 → 告警 |

### 9.2 K8s CronJob 示例

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: koala-exam-state-advancer
  namespace: koala
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: advancer
              image: koala-exam-backend:v1.0.0
              command: ["/app/koala-exam", "job", "advance-state"]
              env: [MYSQL_HOST, REDIS_HOST, JWT_SECRET]
          restartPolicy: OnFailure
```

---

## 10. 已知技术债与改进

### v1.0 → v1.1 待办

| 项 | 优先级 | 工作量 |
|----|--------|--------|
| 全局 rand 替换为 crypto/rand + seed | P1 | 1d |
| 阅卷异步化（Redis Stream Worker） | P0 | 5d |
| 重考机制（MaxAttempts 实现） | P1 | 3d |
| 审计Log 拆表 + 批量 INSERT | P1 | 2d |
| 状态机 Pending 阶段 | P2 | 2d |
| 自动超时收卷 cron | P0 | 2d |
| SignScore 用 KMS 注入 | P0 | 1d |
| TargetUsers/TargetClasses 拆表 | P2 | 3d |
| paper_snapshot 阅卷后清理（30 天后归档） | P2 | 2d |
| 题目顺序确定性随机（user_id+exam_id hash） | P2 | 1d |

### 长期（v2.0+）

- 微服务化：考试服务独立部署
- WebSocket 推送替代轮询答题同步
- PWA + 离线答题（IndexedDB）
- AI 辅助监考（视频流分析）

---

## 11. 相关文档

- [0002_ARCHITECTURE.md](0002_ARCHITECTURE.md) - 整体架构与考试数据流
- [0004_FUNCTION_MODULES.md](0004_FUNCTION_MODULES.md) - 防作弊功能列表
- [0005_DATABASE.md](0005_DATABASE.md) - 表结构（注意：实际表名前缀 ke_，文档用 ems_，已规划统一）
- [0006_API.md](0006_API.md) - 考试接口清单
- [0009_TECH_RISK.md](0009_TECH_RISK.md) - 提交风暴、断线续考风险
- [0016_ANTI_CHEAT.md](0016_ANTI_CHEAT.md) - 防作弊专题
- [0017_GRADING.md](0017_GRADING.md) - 阅卷流水线专题

---

> 维护者：考试组 + 架构组 / 月度评审
