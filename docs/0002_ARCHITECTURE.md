# 🏗 系统架构

> 描述 KoalaExam 的整体技术架构、分层设计、数据流、关键设计决策与权衡。

---

## 1. 设计目标与非目标

### 设计目标

- **高并发**：单实例支持 3,000+ 学员同时在线考试
- **数据强一致**：考试记录、成绩防篡改
- **可观测**：全链路 TraceID + Prometheus 指标
- **可演进**：从单体到微服务的平滑过渡

### 非目标

- **不支持实时音视频监考**（属于 v2.0+）
- **不支持跨数据中心的全球部署**（v1.x 限定单 Region）
- **不追求极致单机性能**（优先水平扩展）

---

## 2. 整体架构图

```
┌──────────────────────────────────────────────────────────────────────┐
│                       客户端（浏览器）                                 │
│  Vue 3 + Vite + TypeScript + Pinia + Element Plus                    │
│  ├── 学员端（考试大厅 / 答题 / 错题本 / 收藏）                       │
│  ├── 教师端（出题 / 组卷 / 监考 / 阅卷 / 统计）                      │
│  └── 管理端（用户管理 / 系统配置 / 审计日志）                        │
└──────────────────────────────────────────────────────────────────────┘
                                  │ HTTP/HTTPS + JSON
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│                       Nginx 反向代理层                                │
│         (TLS 1.3 终止 + 静态资源 + 限流 + WAF 联动)                  │
└──────────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌──────────────────────────────────────────────────────────────────────┐
│             Hertz HTTP Server（基于 Netpoll 网络库）                   │
│   中间件链：RequestID → Recovery → CORS → JWT → 限流 → Audit         │
├──────────────────────────────────────────────────────────────────────┤
│                       应用层（internal/application）                   │
│  user │ question │ paper │ exam │ grading │ favorite │ statistics    │
├──────────────────────────────────────────────────────────────────────┤
│                       领域层（internal/domain）                       │
│         entity（充血模型）+ consts + errcode + 业务规则               │
├──────────────────────────────────────────────────────────────────────┤
│                     基础设施层（infrastructure）                       │
│         repository(GORM) | cache(go-redis) | oss | mq | tracer       │
└──────────────────────────────────────────────────────────────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        ▼                         ▼                         ▼
   ┌─────────┐              ┌──────────┐              ┌──────────┐
   │ MySQL   │              │  Redis   │              │   OSS    │
   │  8.0    │              │   7.x    │              │ (可选)   │
   │ 主从架构│              │ Cluster  │              │          │
   └─────────┘              └──────────┘              └──────────┘
                                  │
                                  ▼
                  ┌──────────────────────────────────┐
                  │   阅卷 Worker (Redis Stream)      │
                  │   (异步消费，可独立伸缩)           │
                  └──────────────────────────────────┘
```

---

## 3. 后端分层（Clean Architecture + DDD）

### 3.1 目录结构

```
koala-exam-backend/
├── cmd/
│   ├── hertz/                    # 主服务入口 (HTTP API)
│   ├── migrate/                  # 数据库迁移工具
│   └── worker/                   # 阅卷异步 Worker
├── internal/                     # 内部包（不可被外部引用）
│   ├── interfaces/              # 接口层
│   │   ├── handler/             # HTTP 处理器（Hertz 路由绑定）
│   │   ├── middleware/          # 中间件（JWT/限流/CORS/Audit）
│   │   └── router/              # 路由注册
│   ├── application/             # 应用层（业务编排 / UseCase）
│   │   ├── user/                # 用户应用服务
│   │   ├── question/            # 题库
│   │   ├── paper/               # 试卷
│   │   ├── exam/                # 考试
│   │   ├── grading/             # 阅卷
│   │   ├── favorite/            # 深度收藏/错题本（核心）
│   │   └── statistics/          # 统计
│   ├── domain/                  # 领域层（核心业务规则）
│   │   ├── entity/              # 实体（充血模型）
│   │   ├── valueobject/         # 值对象
│   │   ├── consts/              # 常量（角色/题型/状态）
│   │   └── errcode/             # 业务错误码
│   └── infrastructure/          # 基础设施层
│       ├── database/            # MySQL + GORM 初始化
│       ├── cache/               # Redis 客户端
│       ├── mq/                  # Redis Stream 封装
│       └── repository/          # 仓储实现（DAO）
├── pkg/                         # 公共库（可被外部项目引用）
│   ├── jwt/                     # JWT 工具
│   ├── encrypt/                 # bcrypt + SHA-256 + AES
│   ├── response/                # 统一返回
│   ├── config/                  # Viper 配置
│   ├── logger/                  # Zap 结构化日志
│   ├── utils/                   # 工具函数
│   └── tracer/                  # OpenTelemetry
├── configs/                     # 配置文件
├── migrations/                  # 数据库迁移 SQL
├── docs/                        # swag 生成的接口文档
├── scripts/                     # 脚本（迁移/压测/部署）
├── test/                        # 测试数据
├── .air.toml                    # air 热重载配置
├── Dockerfile                   # Docker 镜像构建
└── Makefile                     # 常用命令
```

### 3.2 分层职责

| 层 | 职责 | 不应包含 |
|----|------|----------|
| **interfaces** | 接收 HTTP 请求，参数绑定，调用 application | 业务逻辑、SQL |
| **application** | 业务编排（UseCase）、事务控制、跨域协调 | HTTP 细节、SQL 拼接 |
| **domain** | 实体行为、值对象、业务规则、领域事件 | 任何外部依赖（DB/Cache） |
| **infrastructure** | 仓储实现、外部服务（DB/Redis/OSS）适配 | 业务规则 |

### 3.3 依赖倒置

```go
// application 层依赖 interface，infrastructure 实现 interface
type ExamRepository interface {
    GetByID(ctx context.Context, id uint64) (*entity.Exam, error)
    List(ctx context.Context, filter ExamFilter) ([]*entity.Exam, error)
    Create(ctx context.Context, exam *entity.Exam) error
    // ...
}

// infrastructure/repository/exam.go 实现
type examRepo struct { db *gorm.DB }
func (r *examRepo) GetByID(...) { ... }
```

这样 application 层可独立单测，repository 可替换（MySQL → PostgreSQL 几乎零成本）。

---

## 4. 前端架构

### 4.1 目录结构

```
koala-exam-frontend/
├── src/
│   ├── api/                    # Axios 封装 + 模块化 API
│   │   ├── modules/
│   │   │   ├── auth.ts
│   │   │   ├── exam.ts
│   │   │   ├── question.ts
│   │   │   └── favorite.ts
│   │   └── request.ts          # Axios 实例 + 拦截器
│   ├── components/
│   │   ├── common/             # 通用（按钮/弹窗/分页）
│   │   └── business/           # 业务（FavoriteStar / MasteryTag / QuestionRenderer）
│   ├── composables/            # 组合式函数（Vue3 逻辑复用）
│   │   ├── useExamTimer.ts     # 考试倒计时（防作弊核心）
│   │   ├── useAntiCheat.ts     # 切屏检测 + 复制粘贴拦截
│   │   └── useFavorite.ts      # 收藏乐观更新
│   ├── layouts/                # 布局（MainLayout / ExamLayout）
│   ├── router/                 # 路由 + 守卫（鉴权/角色控制）
│   ├── store/                  # Pinia 状态管理
│   │   ├── modules/
│   │   │   ├── user.ts         # 用户/Token
│   │   │   ├── favorite.ts     # 收藏全局态（避免重复请求）
│   │   │   └── exam.ts         # 考试进行中态
│   ├── types/                  # TS 类型
│   ├── views/                  # 页面（按模块划分）
│   ├── utils/                  # 工具函数
│   ├── App.vue
│   └── main.ts                 # 入口
├── public/                     # 静态资源
├── index.html
├── vite.config.ts
├── tsconfig.json
└── package.json
```

### 4.2 关键技术决策

#### 4.2.1 组合式 API + <script setup>

```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useExamStore } from '@/store/modules/exam'
import { useExamTimer } from '@/composables/useExamTimer'
import { useAntiCheat } from '@/composables/useAntiCheat'

const examStore = useExamStore()
const record = computed(() => examStore.currentRecord)
const { remaining, isWarning } = useExamTimer(record.value?.duration)

useAntiCheat({
  enabled: record.value?.antiCheat,
  onViolation: (type) => examStore.logAudit(type),
})
</script>

<template>
  <ExamQuestion v-for="q in record.questions" :key="q.id" :question="q" />
  <ExamTimer :remaining="remaining" :warning="isWarning" />
</template>
```

#### 4.2.2 Pinia 全局收藏态

```ts
// store/modules/favorite.ts
export const useFavoriteStore = defineStore('favorite', () => {
  const favMap = ref<Map<string, boolean>>(new Map())  // key: `${type}_${id}`

  async function check(targetType: number, targetId: number) {
    const key = `${targetType}_${targetId}`
    if (favMap.value.has(key)) return favMap.value.get(key)!
    const { favorited } = await api.check({ targetType, targetId })
    favMap.value.set(key, favorited)
    return favorited
  }

  async function toggle(targetType: number, targetId: number) {
    const key = `${targetType}_${targetId}`
    // 乐观更新
    const newState = !favMap.value.get(key)
    favMap.value.set(key, newState)
    try {
      await api.toggle({ targetType, targetId })
    } catch (e) {
      // 回滚
      favMap.value.set(key, !newState)
      throw e
    }
  }

  return { favMap, check, toggle }
})
```

#### 4.2.3 路由守卫

```ts
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }
  if (to.meta.role && !userStore.hasRole(to.meta.role)) {
    next({ name: 'Forbidden' })
    return
  }
  next()
})
```

---

## 5. 数据流（典型场景）

### 5.1 学员开始考试

```
1. 学员点击 [开始考试]
   Frontend → POST /api/v1/exams/:id/start
   → Router → ExamHandler.StartExam
   → ExamAppService.StartExam (application)
      ├─ 校验时间窗口（domain rule）
      ├─ 校验用户未开始（唯一索引兜底）
      ├─ 组装试卷（固定/随机/GA）
      │  └─ PaperService.Assemble (strategy)
      ├─ 题目乱序 + 选项乱序
      ├─ 写入 ems_exam_record (paper_snapshot)
      └─ 返回 { exam_id, record_id, questions[], duration, ... }

2. 前端缓存到 LocalStorage（断线续考用）
```

### 5.2 学员作答（节流 + 容错）

```
每 10 秒批量上报：
Frontend → POST /api/v1/exams/answer { record_id, answers: [...] }
→ ExamHandler.SaveAnswer
→ ExamAppService.SaveAnswer
   ├─ 写入 Redis Hash: koala:exam:progress:{record_id}
   ├─ 异常容错：Redis 失败 → 降级到 MySQL（不影响答题）
   └─ 成功 → 200 OK（前端 LocalStorage 清理已同步项）

（关键设计：答题进度不直接写 MySQL，避免高频写打爆 DB）
```

### 5.3 交卷与阅卷

```
1. 学员点击 [交卷] / 超时自动触发
   Frontend → POST /api/v1/exams/submit { record_id }
   → ExamAppService.SubmitExam
      ├─ 合并 Redis 答题数据 → MySQL (ems_exam_answer)
      ├─ 标记 ems_exam_record.status = submitted
      ├─ 推送阅卷任务到 Redis Stream: koala:stream:grading
      └─ 返回 { record_id, status: submitted }  （90ms 内完成）

2. 阅卷 Worker 异步消费（独立部署）
   Worker → 消费 Stream → GradingAppService.AutoGrade
      ├─ 客观题自动评分
      ├─ 主观题入队待人工评分
      ├─ → FavoriteApp.RecordWrongAnswers
      │   ├─ 写入 ems_wrong_log
      │   ├─ 自动加入收藏（错题本）
      │   └─ 计算掌握度（mastery_level）
      ├─ 计算成绩 + SHA-256 签名
      └─ 通知前端（WebSocket / SSE）

3. 教师在线批改主观题 → 更新成绩
   Frontend → POST /api/v1/grading/subjective
   → GradingAppService.SubjectiveGrade
   → 重新计算总分 + 签名 + 入错题本
```

---

## 6. 关键设计

### 6.1 试卷快照（防篡改 + 兼容题目修改）

```go
// domain/entity/exam_record.go
type ExamRecord struct {
    ID            uint64
    ExamID        uint64
    UserID        uint64
    PaperSnapshot datatypes.JSON `gorm:"type:json"`  // 试卷完整快照
    Status        int
    // ...
}

// 开始考试时：
func (s *ExamApp) StartExam(...) {
    snapshot := paper.ToJSON()  // 题目 + 选项 + 分值 + 顺序
    record := &ExamRecord{
        PaperSnapshot: snapshot,
        // ...
    }
    s.repo.Create(record)
    // 阅卷时只读快照，不依赖当前题目表
}
```

**收益**：即使题目被修改 / 删除，已开始的考试不受影响；阅卷可独立完成。

### 6.2 成绩签名（防篡改）

```go
// pkg/encrypt/score.go
func SignScore(record *ExamRecord, salt string) string {
    payload := fmt.Sprintf("%d|%d|%.2f|%.2f|%.2f|%s",
        record.ID, record.UserID,
        record.ObjectiveScore,
        record.SubjectiveScore,
        record.TotalScore,
        salt)
    h := sha256.Sum256([]byte(payload))
    return hex.EncodeToString(h[:])
}

func VerifyScore(record *ExamRecord, salt string) bool {
    return SignScore(record, salt) == record.ScoreHash
}
```

**调用**：每次阅卷完成后调用 `SignScore`；任何修改（用户/教师）都会重新签名。
**核验**：审计日志 + 申诉接口可用 `VerifyScore` 验证。

### 6.3 错题自动入库（深度收藏核心）

```go
// application/favorite/wrong_answers.go
func (s *FavoriteApp) RecordWrongAnswers(ctx context.Context, record *ExamRecord) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        for _, ans := range record.WrongAnswers {
            // 1. 写入错题日志
            wrongLog := &entity.WrongLog{
                UserID:     record.UserID,
                QuestionID: ans.QuestionID,
                UserAnswer: ans.UserAnswer,
                CorrectAns: ans.CorrectAnswer,
                Mastery:    1,  // 初始掌握度
            }
            if err := tx.Create(wrongLog).Error; err != nil {
                return err
            }

            // 2. 自动加入收藏（错题本）
            favorite := &entity.Favorite{
                UserID:      record.UserID,
                TargetType:  1,  // 题目
                TargetID:    ans.QuestionID,
                SourceType:  2,  // 错题自动
                WrongLogID:  &wrongLog.ID,
                FolderID:    s.getOrCreateFolder(record.UserID, "我的错题本"),
            }
            if err := tx.Create(favorite).Error; err != nil {
                return err
            }
        }
        return nil
    })
}
```

### 6.4 多态收藏

```sql
-- ems_favorite 表
CREATE TABLE ems_favorite (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  target_type TINYINT NOT NULL,        -- 1=题目 2=试卷 3=知识点
  target_id BIGINT NOT NULL,
  source_type TINYINT NOT NULL,        -- 1=主动 2=错题自动 3=智能推荐
  folder_id BIGINT,
  note VARCHAR(500),
  wrong_log_id BIGINT,                 -- 错题关联
  deleted_at DATETIME,
  UNIQUE KEY uniq_user_target (user_id, target_type, target_id, deleted_at)
);
```

支持收藏题目 / 试卷 / 知识点，统一接口（`POST /favorites/toggle`）。

### 6.5 掌握度动态计算

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

每次复习错题后调用 `RecomputeMastery`，错题本按掌握度分级。

---

## 7. 横切关注点

### 7.1 中间件执行顺序

```
HTTP Request
  ↓
RequestID（生成 TraceID）
  ↓
Reception（日志记录）
  ↓
Reception（panic recover）
  ↓
CORS（跨域处理）
  ↓
JWT（解析 token，注入 user 到 ctx）
  ↓
RateLimit（基于 IP/用户限流）
  ↓
Audit（操作审计日志）
  ↓
Handler（业务处理）
  ↓
HTTP Response（TraceID 写入响应头）
```

### 7.2 错误处理

```go
// pkg/errcode/errcode.go
var (
    ErrUnauthorized     = New(100001, "未登录")
    ErrPermissionDenied = New(100003, "无权限")
    ErrExamTimeInvalid  = New(400003, "考试时间窗口错误")
    ErrGradingFailed    = New(500001, "阅卷失败")
)

// 使用：
if exam.Status != entity.ExamStatusPublished {
    return errcode.ErrExamNotPublished.WithMsg("考试未发布")
}

// Handler：
if err != nil {
    ec := errcode.FromError(err)
    response.JSON(ctx, ec.Code, ec.Message, nil)
}
```

### 7.3 统一返回结构

```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "trace_id": "abc123",
  "ts": 1700000000
}
```

---

## 8. 演进路径

### v1.0（当前）
- 单体应用 + 主从数据库
- Redis 单实例 + Sentinel
- 单 Region 部署

### v1.1（Q1 2025）
- 阅卷 Worker 独立部署
- 数据库读写分离
- 引入 OpenTelemetry + Jaeger

### v2.0（Q3 2025）
- 微服务化（按 domain 拆分）
- Service Mesh（Istio）
- 多 Region 容灾

---

> **相关文档**：[0003 技术栈](0003_TECH_STACK.md) · [0005 数据库](0005_DATABASE.md) · [0010 性能](0010_PERFORMANCE.md) · [0012 安全](0012_SECURITY.md)