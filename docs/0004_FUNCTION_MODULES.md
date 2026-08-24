# 📦 功能模块详解

> 按 8 大业务域详细描述每个模块的功能要点、技术实现与扩展点。

---

## 1. 用户权限管理

### 1.1 角色体系

```ts
enum UserRole {
  Admin      = 1  // 超管（全部权限）
  Teacher    = 2  // 教师（题库/阅卷/统计）
  Student    = 3  // 学员（考试/错题本/收藏）
}

// 权限粒度通过 RBAC 标签实现（如 question:create / exam:start）
```

### 1.2 功能清单

- [x] 用户注册 / 登录（手机号 + 密码 / 验证码）
- [x] 密码强度校验（前端 + 后端双重）
- [x] 用户资料修改（昵称 / 头像 / 邮箱 / 手机）
- [x] 密码重置（邮件 Token + 短信验证码）
- [x] 角色分配（仅超管可操作）
- [x] 班级 / 院系组织树
- [x] 批量导入用户（Excel）
- [x] 多端登录控制（设备 ID + 黑名单）
- [ ] SSO 单点登录（v1.2）
- [ ] 人脸识别（v1.2）

### 1.3 权限校验中间件

```go
// middleware/auth.go
func AuthRequired() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        token := c.GetHeader("Authorization")
        claims, err := jwt.Parse(token)
        if err != nil {
            return response.Error(ctx, errcode.ErrUnauthorized)
        }
        ctx = context.WithValue(ctx, "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "role", claims.Role)
        c.Next(ctx)
    }
}

func RequireRole(role int) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        userRole := ctx.Value("role").(int)
        if userRole > role {  // 数字越小权限越大
            return response.Error(ctx, errcode.ErrPermissionDenied)
        }
        c.Next(ctx)
    }
}

// 使用：
router.POST("/api/v1/admin/users", AuthRequired(), RequireRole(UserRoleAdmin), handler.CreateUser)
```

---

## 2. 题库管理

### 2.1 题型支持

| 题型 | type 常量 | 选项存储 | 评分方式 |
|------|----------|----------|----------|
| 单选 | 1 | JSON [{ key, content }] | 等值匹配 |
| 多选 | 2 | JSON | 集合匹配（部分得分） |
| 判断 | 3 | JSON [{ key, content }] | 等值 |
| 填空 | 4 | JSON [{ blank_id, content }] | 字符串匹配 / 正则 |
| 不定项 | 5 | JSON | 多选 + 评分规则 |
| 编程 | 6 | 代码 + 测试用例 | 评测沙箱 |
| 简答 | 7 | 富文本 | 人工评分 |

### 2.2 题库分类（树形）

```sql
-- 支持无限层级（parent_id 自引用）
CREATE TABLE ems_question_category (
  id BIGINT PRIMARY KEY,
  parent_id BIGINT DEFAULT 0,
  name VARCHAR(64) NOT NULL,
  sort INT DEFAULT 0,
  -- ...
);

-- 查询整棵树（WITH RECURSIVE）
WITH RECURSIVE tree AS (
  SELECT id, parent_id, name FROM ems_question_category WHERE id = ?
  UNION ALL
  SELECT c.id, c.parent_id, c.name FROM ems_question_category c
  JOIN tree t ON c.parent_id = t.id
)
SELECT * FROM tree;
```

### 2.3 功能清单

- [x] 题目 CRUD
- [x] 题干富文本（Markdown + 图片上传）
- [x] 选项 JSON
- [x] 标签管理
- [x] 难度分级（1-5）
- [x] 分值配置
- [x] Excel 批量导入（标题、选项、答案、解析、标签）
- [x] 题目统计（被引用次数 / 正确率 / 错误率）
- [x] 题目预览
- [ ] 题目协作（多人编辑 + 版本）
- [ ] 智能题目推荐（v2.0）

### 2.4 Excel 导入流程

```
1. 前端上传 .xlsx
2. 后端解析（excelize/v2）
3. 校验（必填 / 格式 / 数据合法性）
4. 返回预览（成功行 / 失败行 + 错误原因）
5. 用户确认 → 入库
6. 异步返回导入结果（导入 N 条 / 失败 M 条）

模板示例：
  | 题型 | 题干 | 选项A | 选项B | 选项C | 选项D | 答案 | 解析 | 难度 | 分值 | 标签 |
  | 单选 | 1+1=? | 1 | 2 | 3 | 4 | B | 基础加法 | 1 | 2 | 数学/基础 |
```

---

## 3. 组卷管理

### 3.1 三种策略

| 策略 | 适用场景 | 算法 | 耗时 |
|------|----------|------|------|
| 手动组卷 | 教师精选 | 人工挑选题目 | 即时 |
| 随机组卷 | 日常小测 | 按规则随机抽取 | < 100ms |
| 遗传算法 | 大型考试 | GA 优化难度/知识点分布 | 200-500ms |

### 3.2 遗传算法实现（简化）

```go
// pkg/paper/ga.go
type Individual struct {
    Genes     []uint64  // 题目 ID 序列
    Fitness   float64   // 适应度（难度匹配度 + 知识点覆盖率）
}

func GeneticAssemble(req *GAReq) ([]uint64, error) {
    // 1. 初始化种群（随机抽取题目）
    population := make([]*Individual, 100)
    for i := range population {
        population[i] = randomIndividual(req)
    }

    // 2. 迭代（100 代）
    for gen := 0; gen < 100; gen++ {
        // 评估适应度
        for _, ind := range population {
            ind.Fitness = evaluate(ind, req)
        }
        // 选择 / 交叉 / 变异
        population = evolve(population, 0.1)  // 变异率 10%
    }

    // 3. 返回最优个体
    return best(population).Genes, nil
}

func evaluate(ind *Individual, req *GAReq) float64 {
    score := 0.0
    score -= math.Abs(ind.totalDifficulty - req.TargetDifficulty) * 10
    score += ind.knowledgePointCoverage(req.KPs) * 50
    return score
}
```

### 3.3 试卷快照

详见 [0002_ARCHITECTURE.md - 6.1](0002_ARCHITECTURE.md#61-试卷快照防篡改--兼容题目修改)

---

## 4. 考试管理

### 4.1 状态流转

```
草稿 (Draft) → 待发布 (Pending) → 进行中 (Ongoing) → 已结束 (Ended) → 已归档 (Archived)
       │           │              │               │              │
       ↓           ↓              ↓               ↓              ↓
   教师编辑    待超管审核     学员可参加       不可参加        数据分析

状态机：
  - Draft → Pending: 教师提交
  - Pending → Ongoing: 时间到达 start_time（自动）
  - Ongoing → Ended: 时间到达 end_time 或手动停止
  - Ended → Archived: 30 天后自动归档
```

### 4.2 限时考试

- 倒计时：前端 `setInterval` + 服务端 `start_time + duration`
- 超时自动交卷：`POST /exams/submit` 由前端 / 后端 cron 触发
- 断线续考：基于 Redis 答题进度 + LocalStorage

### 4.3 试卷分配

```
方式 1：按用户 - 选择特定学员
方式 2：按班级 - 全班同时参加
方式 3：按组织 - 按院系 / 年级批量
方式 4：开放考试 - 满足条件即可参加（链接 + 验证码）
```

### 4.4 功能清单

- [x] 试卷分配（4 种方式）
- [x] 状态流转
- [x] 限时考试 + 自动交卷
- [x] 断线续考
- [x] 考试通知（站内信 / 邮件）
- [ ] 考试预约（v1.2）
- [ ] 考场分配（v1.2）

---

## 5. 防作弊

### 5.1 措施一览

| 措施 | 实现 | 绕过难度 | 后端配合 |
|------|------|----------|----------|
| 全屏监控 | requestFullscreen() + 失焦告警 | 中 | 审计日志 |
| 切屏检测 | visibilitychange 事件 | 中 | 计数 + 超限自动收卷 |
| 复制粘贴拦截 | copy/paste/contextmenu preventDefault | 高 | 审计 + 警告 |
| 题目乱序 | 后端 Random.Shuffle | - | - |
| 选项乱序 | 后端每次随机 | - | - |
| 答题自动保存 | 每 10s 同步 Redis + LocalStorage | - | 断线续考 |
| 行为审计日志 | 写入 ems_exam_record.audit_log | - | 监考大屏 |
| IP 绑定 | 异常 IP 告警 | 低 | CSO 审核 |
- [ ] 人脸识别（v1.2）
- [ ] 屏幕录制（v1.2）
- [ ] AI 异常行为检测（v2.0）

### 5.2 防作弊核心实现（前端）

```ts
// composables/useAntiCheat.ts
export function useAntiCheat(options: {
  enabled: boolean
  onViolation: (type: string) => void
}) {
  const tabSwitchCount = ref(0)

  function onVisibilityChange() {
    if (document.hidden && options.enabled) {
      tabSwitchCount.value++
      options.onViolation('tab_switch')
      if (tabSwitchCount.value >= 3) {
        options.onViolation('force_submit')
      }
    }
  }

  function preventCopy(e: ClipboardEvent) {
    if (options.enabled) {
      e.preventDefault()
      options.onViolation('copy_attempt')
    }
  }

  onMounted(() => {
    document.addEventListener('visibilitychange', onVisibilityChange)
    document.addEventListener('copy', preventCopy)
    document.addEventListener('paste', preventCopy)
    document.addEventListener('contextmenu', preventCopy)
  })

  onUnmounted(() => {
    document.removeEventListener('visibilitychange', onVisibilityChange)
    // ...
  })

  return { tabSwitchCount }
}
```

### 5.3 监考大屏

```
教师端实时面板：
  - 当前考试列表（按场次分组）
  - 每场：在线人数 / 已交卷 / 异常事件
  - 单考生：实时屏幕 + 操作流（鼠标轨迹 + 切屏次数）
  - 强制收卷 / 延长考试时间 / 标记异常

实现：
  - 后端 SSE（Server-Sent Events）推送
  - 前端 EventSource 订阅
```

---

## 6. 阅卷管理

### 6.1 自动阅卷（客观题）

```go
// application/grading/auto_grade.go
func (s *GradingApp) AutoGrade(record *entity.ExamRecord) error {
    for _, ans := range record.Answers {
        q := s.repo.GetQuestion(ans.QuestionID)
        switch q.Type {
        case QuestionTypeSingle, QuestionTypeJudge:
            // 单选/判断：等值匹配
            if ans.UserAnswer == q.CorrectAnswer {
                ans.Score = q.Score
            }
        case QuestionTypeMultiple, QuestionTypeUncertain:
            // 多选：集合匹配（按对错比例给分）
            correct, wrong := compareSets(ans.UserAnswer, q.CorrectAnswer)
            ans.Score = q.Score * (correct - wrong) / len(q.CorrectAnswer)
        case QuestionTypeBlank:
            // 填空：每个空独立匹配（支持模糊）
            blanks := parseBlanks(ans.UserAnswer)
            correctCount := 0
            for i, b := range blanks {
                if matchBlank(b, q.Blanks[i]) {  // 支持模糊匹配
                    correctCount++
                }
            }
            ans.Score = q.Score * correctCount / len(q.Blanks)
        }
    }
    return s.repo.UpdateAnswers(record.Answers)
}
```

### 6.2 主观题人工评分

```
教师阅卷界面：
  - 单题模式：一次只评一道题（效率高）
  - 双评模式：两位教师独立评分 → 系统对比 → 差异过大重评
  - AI 辅助评分（v1.2）：GPT-4 给出参考分数 + 教师确认

评分维度：
  - 准确性（占比 70%）
  - 完整性（占比 20%）
  - 表达（占比 10%）
```

### 6.3 代码题自动评测

```
v1.2 计划：对接 Judge0 / 自建沙箱

评测流程：
  1. 拉取用户代码 + 测试用例
  2. 沙箱编译（沙箱内 Docker）
  3. 运行测试用例（限资源 / 限时间）
  4. 比对输出 → 计算得分（通过的测试用例比例）
  5. 返回结果（耗时 / 内存 / 通过率）

沙箱要求：
  - 资源隔离（cgroup）
  - 网络隔离
  - 超时熔断（最长 30s）
  - 安全审计（禁止执行危险命令）
```

---

## 7. 成绩统计分析

### 7.1 多维统计

```sql
-- 平均分 / 及格率
SELECT AVG(total_score) avg_score,
       AVG(CASE WHEN total_score >= pass_score THEN 1 ELSE 0 END) pass_rate
FROM ems_exam_record
WHERE exam_id = ? AND status IN (3, 4);  -- 已交卷/已批改

-- 分数段分布
SELECT CASE
         WHEN total_score >= 90 THEN 'A'
         WHEN total_score >= 80 THEN 'B'
         WHEN total_score >= 70 THEN 'C'
         WHEN total_score >= 60 THEN 'D'
         ELSE 'F'
       END grade_bucket,
       COUNT(*) cnt
FROM ems_exam_record
WHERE exam_id = ?
GROUP BY grade_bucket;

-- 知识点薄弱分析
SELECT q.knowledge_point,
       COUNT(*) wrong_count,
       AVG(qr.user_score / qr.question_score) avg_correct_rate
FROM ems_exam_answer qr
JOIN ems_question q ON qr.question_id = q.id
WHERE qr.exam_id = ? AND qr.user_score < qr.question_score
GROUP BY q.knowledge_point
ORDER BY wrong_count DESC;
```

### 7.2 可视化

| 图表 | 类型 | 用途 |
|------|------|------|
| 成绩分布 | 柱状图 | 分数段人数 |
| 趋势分析 | 折线图 | 历次成绩对比 |
| 知识点雷达 | 雷达图 | 学生能力画像 |
| 班级对比 | 堆叠柱状图 | 班级均分对比 |
| 答题时长 | 散点图 | 异常考生识别 |

### 7.3 导出

```
Excel 导出（成绩单）：
  - 单场：包含所有考生成绩 + 排名
  - 个人：含分题得分 + 错题列表 + 解析

PDF 导出（成绩证明）：
  - 含 Logo + 二维码（扫码核验）
  - SHA-256 签名

v1.2 计划：
  - 定制化报表（BI）
  - 数据看板（实时刷新）
```

---

## 8. 深度收藏 / 错题本（核心卖点 🐨）

### 8.1 三种来源

```
主动收藏（source_type=1）：
  学员点击 ⭐ → POST /favorites/toggle → 加入收藏

错题自动（source_type=2）：
  阅卷完成 → GradingApp.AutoGrade → 检测错题 → 自动加入收藏（无需手动）

智能推荐（source_type=3，v2.0）：
  基于错题率推相似题目 → 用户确认收藏
```

### 8.2 收藏夹管理

```sql
CREATE TABLE ems_favorite_folder (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,         -- 高频错题 / 必背大题 / 算法专项
  color VARCHAR(16),
  icon VARCHAR(32),
  sort INT DEFAULT 0,
  is_system TINYINT DEFAULT 0,       -- 系统默认（我的收藏/我的错题本）
  created_at DATETIME
);
```

- 系统自动生成：「我的收藏」「我的错题本」
- 用户自定义：自由命名 + 排序 + 颜色 + 图标

### 8.3 错题本

每次错题记录：

```sql
CREATE TABLE ems_wrong_log (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  question_id BIGINT NOT NULL,
  exam_record_id BIGINT,
  user_answer TEXT,
  correct_answer TEXT,
  wrong_count INT DEFAULT 1,         -- 累计错次
  review_count INT DEFAULT 0,        -- 已复习次数
  correct_count INT DEFAULT 0,       -- 复习正确次数
  mastery_level TINYINT DEFAULT 1,   -- 1=薄弱 2=一般 3=良好 4=掌握
  last_wrong_at DATETIME,
  last_review_at DATETIME,
  UNIQUE KEY uniq_user_q (user_id, question_id)
);
```

### 8.4 掌握度动态计算

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
        w.MasteryLevel = 1  // 薄弱
    case correctRate < 0.7:
        w.MasteryLevel = 2  // 一般
    case correctRate < 0.9:
        w.MasteryLevel = 3  // 良好
    default:
        w.MasteryLevel = 4  // 掌握
    }
}
```

UI 展示：
- 1 星红 + 「薄弱重点」
- 2 星黄 + 「需加强」
- 3 星蓝 + 「基本掌握」
- 4 星绿 + 「已掌握」

### 8.5 复习模式

```
1. 进入「错题本复习」
2. 系统按掌握度优先级推荐（薄弱优先）
3. 展示题目 → 用户作答
4. 判分 + 更新 review_count / correct_count / mastery_level
5. 推荐下一题（艾宾浩斯遗忘曲线间隔）
6. 复习完成 → 统计（本次复习 X 题 / 掌握度提升 Y%）
```

### 8.6 扩展点

- [ ] 智能推荐相似题（v2.0）
- [ ] 错题导出 PDF（v1.2）
- [ ] 错题本分享（v1.2）
- [ ] 错题讲解（GPT-4，v2.0）

---

## 9. 通知与消息

### 9.1 渠道

- 站内信（必选）
- 邮件（可选）
- 短信（可选，付费）
- 钉钉 / 企微 / 飞书 Webhook（v1.2）

### 9.2 触发场景

```
- 考试开始前 10 分钟：提醒
- 考试结束：成绩发布
- 错题自动入库：本周错题汇总（每周日）
- 系统公告：管理员发布
- 审核结果：教师操作完成
```

---

## 10. 管理员后台

### 10.1 功能

- [x] 用户管理（CRUD + 角色分配 + 批量导入）
- [x] 部门 / 班级管理
- [x] 系统配置（限流 / 防作弊 / 邮件）
- [x] 审计日志查询
- [x] 数据导出审计
- [ ] 仪表盘（v1.2）
- [ ] 数据字典管理（v1.2）

---

> **相关文档**：[0002 架构](0002_ARCHITECTURE.md) · [0005 数据库](0005_DATABASE.md) · [0006 API](0006_API.md) · [0012 安全](0012_SECURITY.md)