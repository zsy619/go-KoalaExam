# ⚖️ 阅卷与判分专题（Grading）

> 深入剖析 KoalaExam 的阅卷流水线：自动阅卷算法、主观题评分、错题本联动、成绩签名。
> 基于 internal/application/grading/grading_app.go、internal/application/favorite/favorite_app.go 真实代码实现。

---

## 1. 阅卷流水线概览

```
┌──────────────────────────────────────────────────────────────┐
│  v1.0 同步阅卷（当前）                                          │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  POST /exams/submit                                          │
│       ↓                                                      │
│  ExamApp.SubmitExam                                          │
│    ├─ 合并 Redis progress → MySQL.answers                    │
│    ├─ 状态 → Submitted                                       │
│    └─ return record                                          │
│       ↓                                                      │
│  GradingApp.AutoGrade (同事务)                              │
│    ├─ 反序列化 PaperSnapshot + Answers                       │
│    ├─ 题型分支：单选/多选/判断/填空/编程/简答                │
│    ├─ 逐题打分 → objective_score                             │
│    ├─ 计算总分 → total_score                                  │
│    └─ SignScore → score_hash                                 │
│       ↓                                                      │
│  FavoriteApp.RecordWrongAnswers                              │
│    ├─ 错题写入 ke_wrong_log                                  │
│    ├─ 自动加入收藏 (FavoriteSourceAuto)                      │
│    └─ 计算 mastery_level                                     │
│       ↓                                                      │
│  return 200 OK + 完整成绩                                    │
│                                                              │
│  总耗时：300-500ms（高频提交时易打爆 DB）                    │
└──────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│  v1.1 异步阅卷（计划）                                          │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  POST /exams/submit                                          │
│       ↓                                                      │
│  ExamApp.SubmitExam                                          │
│    ├─ 合并 Redis → MySQL                                     │
│    ├─ 入队 Redis Stream: koala:stream:grading               │
│    └─ return { status: submitted }                           │
│                                                              │
│  ... 用户立即看到「已交卷，阅卷中」                            │
│                                                              │
│  Grading Worker (独立部署)                                   │
│    ├─ 消费 Stream                                            │
│    ├─ AutoGrade + RecordWrongAnswers                         │
│    ├─ WebSocket 通知前端                                     │
│    └─ return 成绩                                            │
│                                                              │
│  总耗时：<100ms（提交），5-10s（阅卷）                       │
└──────────────────────────────────────────────────────────────┘
```

---

## 2. 客观题自动阅卷

### 2.1 支持的题型

| 题型 | type 常量 | 评分算法 | 边界处理 |
|------|----------|----------|----------|
| 单选 | 1 | 等值匹配 | 完全相同才得分 |
| 多选 | 2 | 集合匹配 | 对对错错按比例 |
| 判断 | 3 | 等值匹配 | true/false 比对 |
| 填空 | 4 | 字符串模糊 | normalize 后比对 |
| 不定项 | 5 | 自定义规则 | 视配置 |
| 编程 | 6 | 沙箱评测 | 测试用例通过率 |
| 简答 | 7 | 人工评分 | 教师打分 |

### 2.2 核心评分代码

```go
// application/grading/grading_app.go
func (a *GradingApp) AutoGrade(ctx context.Context, record *entity.ExamRecord) error {
    // 1. 解析快照与答案
    var paper []entity.Question
    if err := json.Unmarshal([]byte(record.PaperSnapshot), &paper); err != nil {
        return err
    }

    answers := map[string]interface{}{}
    json.Unmarshal([]byte(record.Answers), &answers)

    totalObjective := 0.0
    wrongAnswers := []entity.WrongAnswer{}

    // 2. 逐题判分
    for _, q := range paper {
        userAns, answered := answers[strconv.FormatInt(q.ID, 10)]
        score := 0.0
        correct := false

        if !answered || userAns == nil {
            // 未作答
            score = 0
        } else {
            switch q.Type {
            case QuestionTypeSingle, QuestionTypeJudge:
                correct = compareAnswer(q.Answer, userAns)
                if correct { score = q.Score }

            case QuestionTypeMultiple:
                correct, partial = compareSet(q.Answer, userAns)
                score = q.Score * partial  // partial = (correct - wrong) / total

            case QuestionTypeBlank:
                correct, partial = compareBlank(q.Answer, userAns)
                score = q.Score * partial

            case QuestionTypeUncertain:
                // 自定义规则：选部分得 50%
                correct = compareAnswer(q.Answer, userAns)
                if correct { score = q.Score }
                else if containsAny(q.Answer, userAns) { score = q.Score * 0.5 }

            case QuestionTypeProgramming:
                score = 0  // 编程题由评测沙箱打分（异步）
                pending = append(pending, q.ID)

            case QuestionTypeShort:
                // 主观题不入自动阅卷
                pending = append(pending, q.ID)
            }
        }

        totalObjective += score
        if !correct && answered {
            wrongAnswers = append(wrongAnswers, entity.WrongAnswer{
                QuestionID: q.ID,
                UserAnswer: userAns,
                CorrectAnswer: q.Answer,
            })
        }
    }

    // 3. 更新成绩
    record.ObjectiveScore = totalObjective
    record.TotalScore = totalObjective + record.SubjectiveScore
    record.Passed = record.TotalScore >= record.PassScore

    // 4. 签名防篡改
    record.ScoreHash = a.SignScore(record)

    // 5. 写入（事务）
    return a.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Save(record).Error; err != nil {
            return err
        }
        // 6. 错题自动收录
        if len(wrongAnswers) > 0 {
            return a.favoriteApp.RecordWrongAnswers(ctx, record.UserID, wrongAnswers)
        }
        return nil
    })
}
```

### 2.3 答案比对算法

#### 2.3.1 单选/判断（等值）

```go
func compareAnswer(correct, user interface{}) bool {
    if correct == nil || user == nil { return false }
    cs := fmt.Sprintf("%v", correct)
    us := fmt.Sprintf("%v", user)
    if cs == us { return true }
    // 简化比对：去空格、转小写
    return normalize(cs) == normalize(us)
}

func normalize(s string) string {
    out := ""
    for _, r := range s {
        if r != ' ' && r != '\t' && r != '\n' {
            out += string(r)
        }
    }
    return out
}
```

#### 2.3.2 多选（集合匹配 + 部分得分）

```go
func compareSet(correct, user interface{}) (bool, float64) {
    cs := toSet(correct)  // 解析为 BLANK[a, b, c]
    us := toSet(user)

    if cs.Equals(us) {
        return true, 1.0
    }

    // 部分得分：(correct - wrong) / total
    correctCnt := 0
    wrongCnt := 0
    for x := range us {
        if cs.Contains(x) {
            correctCnt++
        } else {
            wrongCnt++
        }
    }
    partial := float64(correctCnt - wrongCnt) / float64(cs.Size())
    if partial < 0 { partial = 0 }

    return false, partial
}
```

#### 2.3.3 填空（字符串模糊匹配）

```go
func compareBlank(correct, user interface{}) (bool, float64) {
    // correct: "BLANKapple|orange|banana"  (| 分隔多个空)
    cs := strings.Split(fmt.Sprintf("%v", correct), "|")
    us := strings.Split(fmt.Sprintf("%v", user), "|")

    correctCnt := 0
    for i, c := range cs {
        if i >= len(us) { break }
        if normalize(strings.ToLower(c)) == normalize(strings.ToLower(us[i])) {
            correctCnt++
        }
    }

    return correctCnt == len(cs), float64(correctCnt) / float64(len(cs))
}
```

**注意：填空题不支持模糊匹配（typo 容忍）**。v1.1 应支持 Levenshtein 距离阈值。

---

## 3. 主观题人工评分

### 3.1 评分流程

```
1. 教师打开「阅卷中心」 → 选择某场考试
2. 选择某学员的某道主观题 → 进入阅卷界面
3. 看到题目 + 学员作答 + AI 辅助参考（v1.2）
4. 输入分数 + 评语
5. POST /api/v1/grading/subjective
6. 后端：update score + recompute total + sign score + 通知学员
```

### 3.2 接口与 DTO

```go
// application/dto/exam_dto.go
type SubjectiveGradeReq struct {
    RecordID   int64   `json:"record_id" binding:"required"`
    QuestionID int64   `json:"question_id" binding:"required"`
    Score      float64 `json:"score" binding:"required,min=0"`
    Comment    string  `json:"comment" binding:"max=1000"`
}

type SubjectiveGradeResp struct {
    RecordID        int64   `json:"record_id"`
    NewScore        float64 `json:"new_score"`
    NewTotalScore   float64 `json:"new_total_score"`
    ScoreHash       string  `json:"score_hash"`
    AlreadyGraded    bool    `json:"already_graded"`
}
```

### 3.3 实现

```go
func (a *GradingApp) SubjectiveGrade(ctx context.Context, req *dto.SubjectiveGradeReq) (*dto.SubjectiveGradeResp, error) {
    // 1. 权限校验：教师 / 超管
    role := ctx.Value("role").(int)
    if role > consts.RoleTeacher { return nil, errcode.ErrPermissionDenied }

    rec, _ := a.recordRepo.GetByID(ctx, req.RecordID)

    // 2. 更新某题分数
    var answers map[string]float64
    json.Unmarshal([]byte(rec.SubjectiveAnswers), &answers)
    if answers == nil { answers = map[string]float64{} }
    answers[strconv.FormatInt(req.QuestionID, 10)] = req.Score

    // 3. 重新计算总分
    subjectiveTotal := 0.0
    for _, s := range answers { subjectiveTotal += s }
    rec.SubjectiveScore = subjectiveTotal
    rec.TotalScore = rec.ObjectiveScore + subjectiveTotal
    rec.Passed = rec.TotalScore >= rec.PassScore

    // 4. 重新签名
    rec.ScoreHash = a.SignScore(rec)

    // 5. 保存
    a.recordRepo.Update(ctx, rec)

    // 6. 通知学员（WebSocket / 短信）
    a.notify.GradeComplete(rec.UserID, rec)

    return &dto.SubjectiveGradeResp{
        RecordID: rec.ID,
        NewScore: req.Score,
        NewTotalScore: rec.TotalScore,
        ScoreHash: rec.ScoreHash,
    }, nil
}
```

### 3.4 AI 辅助评分（v1.2）

```ts
// 调用 GPT-4 给出参考分数
const aiSuggestion = await api.aiGrade({
  question: q,
  student_answer: ans,
  reference_answer: q.Answer,
  rubric: q.Rubric,
})

// 教师界面显示：
// AI 参考分：8/10
// 评分理由：回答基本正确，缺少关键步骤 X
// 教师确认：BLANK 8.5 BLANK 9.0 BLANK 9.5 BLANK 10 BLANK 自定义
```

---

## 4. 错题本自动收录

### 4.1 流程

```
阅卷完成 → GradingApp.AutoGrade
  ↓
检测每题正确性
  ↓
错题列表 → FavoriteApp.RecordWrongAnswers
  ├─ 写入 ke_wrong_log（错题记录）
  ├─ 写入 ke_favorite（自动收藏，source_type=2）
  ├─ 关联到「我的错题本」收藏夹（自动创建）
  └─ 计算 mastery_level
```

### 4.2 错题记录实体（ke_wrong_log）

```sql
CREATE TABLE ke_wrong_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    question_id BIGINT NOT NULL,
    exam_record_id BIGINT,
    user_answer TEXT,
    correct_answer TEXT,
    wrong_count INT DEFAULT 1,         -- 累计错次
    review_count INT DEFAULT 0,        -- 已复习次数
    correct_count INT DEFAULT 0,       -- 复习正确次数
    mastery_level TINYINT DEFAULT 1,   -- 1薄弱 2一般 3良好 4掌握
    last_wrong_at DATETIME,
    last_review_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uniq_user_q (user_id, question_id),
    KEY idx_user_mastery (user_id, mastery_level),
    KEY idx_last_wrong (last_wrong_at)
) ENGINE=InnoDB;
```

### 4.3 收藏关联

```sql
-- 自动收藏
INSERT INTO ke_favorite (user_id, target_type, target_id, source_type, wrong_log_id, folder_id)
SELECT
    :user_id,
    1,                    -- target_type=题目
    wl.question_id,
    2,                    -- source_type=错题自动
    wl.id,                -- wrong_log_id
    (SELECT id FROM ke_favorite_folder WHERE user_id = :user_id AND is_system = 1 LIMIT 1)
FROM ke_wrong_log wl
WHERE wl.user_id = :user_id AND wl.id IN (...错题 IDs...);
```

### 4.4 掌握度动态算法

```go
// domain/entity/wrong_log.go
func (w *WrongLog) RecomputeMastery() {
    if w.ReviewCount == 0 {
        w.MasteryLevel = 1  // 未复习
        return
    }
    correctRate := float64(w.CorrectCount) / float64(w.ReviewCount)
    switch {
    case correctRate < 0.4:
        w.MasteryLevel = 1  // 薄弱（红色）
    case correctRate < 0.7:
        w.MasteryLevel = 2  // 一般（黄色）
    case correctRate < 0.9:
        w.MasteryLevel = 3  // 良好（蓝色）
    default:
        w.MasteryLevel = 4  // 掌握（绿色）
    }
}
```

**v1.1 改进**：引入艾宾浩斯遗忘曲线，推荐下次复习时间。

---

## 5. 成绩签名（防篡改）

### 5.1 当前实现

```go
func (a *GradingApp) SignScore(rec *entity.ExamRecord) string {
    payload := fmt.Sprintf("%d|%.2f|%.2f|%.2f|%s",
        rec.ID,
        rec.TotalScore,
        rec.ObjectiveScore,
        rec.SubjectiveScore,
        "koala-exam-salt",  // ⚠️ 硬编码！v1.1 改为 KMS 注入
    )
    h := sha256.Sum256([]byte(payload))
    return hex.EncodeToString(h[:])
}
```

### 5.2 验证流程

```go
func (a *GradingApp) VerifyScore(rec *entity.ExamRecord) bool {
    expected := a.SignScore(rec)
    return expected == rec.ScoreHash
}

// 申诉接口 / 审计日志调用
```

### 5.3 安全权衡

| 维度 | 当前 | v1.1 改进 |
|------|------|----------|
| Salt | 硬编码 | KMS 注入 + 定期轮换 |
| Payload | 仅分数 | 含 exam_id + user_id + 时间戳 |
| 防单字段篡改 | ✅ | ✅ |
| 防整体重放 | ❌ | ✅ 加时间戳 |
| 防密钥泄露 | ❌ | ✅ 定期轮换 + 审计 |

### 5.4 v1.1 改进版

```go
func SignScore(rec *entity.ExamRecord, jwtSecret string) string {
    payload := fmt.Sprintf("%d|%d|%d|%.2f|%.2f|%.2f|%s|%d",
        rec.ID, rec.ExamID, rec.UserID,
        rec.TotalScore, rec.ObjectiveScore, rec.SubjectiveScore,
        jwtSecret, time.Now().Unix())
    h := sha256.Sum256([]byte(payload))
    return hex.EncodeToString(h[:])
}
```

---

## 6. 编程题自动评测（v1.2 计划）

### 6.1 评测流程

```
1. 用户提交代码 + 测试用例
2. 后端转交评测 Worker
3. Worker 启动 Docker 容器（沙箱）
4. 编译用户代码 → 运行测试用例
5. 收集结果（通过 / 失败 / 超时 / 错误）
6. 计算得分 = 通过测试数 / 总测试数 * 题目分值
7. 回写 record
8. 通知用户
```

### 6.2 沙箱设计

```yaml
# 评测容器配置
image: ubuntu:22.04          # 或 alpine
cpu_limit: 1                 # 1 核
memory_limit: 256M           # 256 MB
disk_limit: 1G
network: none                # 完全离线
timeout: 30s                 # 最长 30s
user: nobody                # 非 root
read_only: true              # 只读文件系统
```

### 6.3 Judge0 对接

```ts
// 前端调用后端 API
const result = await api.judgeProgramming({
  source_code: userCode,
  language_id: 71,           // Python 3
  stdin: '',
  expected_output: '42',
  cpu_time_limit: 5,
  memory_limit: 128000,
})

// 后端转发到 Judge0
const resp = await fetch(process.env.JUDGE0_URL + '/submissions', {
  method: 'POST',
  body: JSON.stringify({ source_code, language_id: 71, ... }),
})

// 轮询结果 → 写回
```

### 6.4 评测得分公式

```
score = question.score * (passed_cases / total_cases)
```

**边界情况**：
- 部分用例通过 → 按比例
- 编译错误 → 0 分（错误信息展示给用户）
- 超时 → 0 分 + 标记「超时」
- 内存超限 → 0 分

---

## 7. 异步阅卷架构（v1.1 计划）

### 7.1 Redis Stream 设计

```
Stream Key: koala:stream:grading
Consumer Group: koala-grading-workers
Consumer: koala-grading-worker-1, -2, -3...

Entry:
  record_id: 12345
  exam_id: 88
  user_id: 1001
  enqueued_at: 1700000000
```

### 7.2 Worker 实现

```go
// cmd/worker/main.go
func main() {
    worker := grading.NewWorker(redisClient, mysqlClient)
    for {
        // 1. XREADGROUP 消费
        streams, _ := worker.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
            Group: "koala-grading-workers",
            Consumer: "worker-1",
            Streams: []string{"koala:stream:grading", ">"},
            Count: 10,
            Block: 5 * time.Second,
        }).Result()

        for _, stream := range streams {
            for _, msg := range stream.Messages {
                // 2. 处理
                var task GradingTask
                json.Unmarshal([]byte(msg.Values["data"]), &task)
                worker.Process(task)

                // 3. ACK
                worker.rdb.XAck(ctx, "koala:stream:grading",
                    "koala-grading-workers", msg.ID)
            }
        }
    }
}
```

### 7.3 死信队列

```go
// 失败 3 次后入死信
if attempts >= 3 {
    worker.rdb.XAdd(ctx, &redis.XAddArgs{
        Stream: "koala:stream:grading:dlq",
        Values: msg.Values,
    })
}
```

### 7.4 WebSocket 通知

```go
// 阅卷完成后通知前端
func (n *Notifier) GradeComplete(userID int64, rec *ExamRecord) {
    payload := map[string]interface{}{
        "event":": record_graded",
        "data":": rec,
    }
    msg, _ := json.Marshal(payload)
    n.rdb.Publish(ctx, fmt.Sprintf("koala:notify:user:%d", userID), msg)
}
```

---

## 8. 阅卷配置项

```yaml
# config.yaml
grading:
  mode: sync            # sync | async（v1.1 切换）
  async:
    stream_name: koala:stream:grading
    consumer_group: koala-grading-workers
    max_retries: 3
    worker_count: 3     # 默认 3 个 Worker
  ai_assist:
    enabled: false      # v1.2
    provider: openai    # openai | azure | wenxin
  programming:
    enabled: false      # v1.2
    provider: judge0    # judge0 | self-host
    sandbox:
      cpu_limit: 1
      memory_limit_mb: 256
      timeout_sec: 30
```

---

## 9. 阅卷质量评估

### 9.1 指标

```
- 阅卷延迟 P50 / P95 / P99
- 自动阅卷准确率（与人工对比抽样 1%）
- 错题收录完整率（与人工标注对比）
- AI 辅助评分采纳率（教师是否修改 AI 建议）
- 死信队列堆积率
```

### 9.2 抽样校对

每月抽样 100 份记录：
- 客观题：对比人工判分（误差应 = 0）
- 主观题：对比两位教师分数（差异 < 5% 为合格）
- 错题收录：随机抽查 50 道错题是否入库

---

## 10. 已知技术债与改进

### v1.0 → v1.1

- [ ] 阅卷异步化（Redis Stream Worker）
- [ ] 主观题 AI 辅助评分
- [ ] 填空题模糊匹配（Levenshtein）
- [ ] 错题本推荐算法（艾宾浩斯曲线）
- [ ] 阅卷 Worker 独立部署 + 自动扩容
- [ ] SignScore 用 KMS 注入
- [ ] 阅卷配置化（rubric / 评分规则）

### v2.0+

- [ ] 编程题沙箱评测
- [ ] 阅卷中心可视化大屏
- [ ] 申诉流程（学员对成绩有异议）
- [ ] 复阅机制（教师可重新评判）

---

## 11. 相关文档

- [0015_EXAM_LIFECYCLE.md](0015_EXAM_LIFECYCLE.md) - 阅卷在考试生命周期中的位置
- [0016_ANTI_CHEAT.md](0016_ANTI_CHEAT.md) - 阅卷前的防作弊
- [0004_FUNCTION_MODULES.md](0004_FUNCTION_MODULES.md) - 第 6 节阅卷模块
- [0010_PERFORMANCE.md](0010_PERFORMANCE.md) - 阅卷性能调优
- [0011_RUNBOOK.md](0011_RUNBOOK.md) - 阅卷失败应急流程

---

> 维护者：阅卷组 + 算法组 / 月度评审
