# KoalaExam 深度功能增强文档

本文档汇总 v2 版本的深度优化功能与架构改进。

## 一、领域层增强

### 1.1 领域事件 (Domain Event)

**位置**: `internal/domain/event/`

事件总线实现进程内 pub-sub，解耦应用层副作用：

| 事件 | 触发时机 | 订阅方 |
| --- | --- | --- |
| `ExamStartedEvent` | 考试开始 | 通知、统计 |
| `ExamSubmittedEvent` | 考试提交 | 自动阅卷 |
| `ExamCheatedEvent` | 防作弊阈值 | 监考告警 |
| `FavoriteToggledEvent` | 收藏切换 | 推荐统计 |
| `WrongAnswerRecordedEvent` | 错题记录 | 学习曲线 |
| `WrongBookReviewedEvent` | 错题复习 | 掌握度更新 |
| `UserLoggedOutEvent` | 用户登出 | Token 黑名单 |
| `UserLoginFailedEvent` | 登录失败 | 风控统计 |

### 1.2 值对象 (Value Object)

**位置**: `internal/domain/valueobject/`

| 值对象 | 职责 |
| --- | --- |
| `ExamPaper` | 试卷快照（不可变） |
| `UserAnswers` | 用户答案（map） |
| `ExamScore` | 考试分值（含计算与及格判定） |
| `AuditSummary` | 防作弊汇总（切屏/全屏/Copy/Paste/Devtools） |
| `TimeWindow` | 时间窗口（开始/结束/包含） |
| `Password` | 密码强度校验 |

### 1.3 领域服务 (Domain Service)

**位置**: `internal/domain/service/`

- `QuestionSelector` 题目选择器（固定/随机/遗传）
- `AnswerComparator` 答案比对器（支持多选/大小写无关）
- `GradingStrategy` 自动阅卷策略
- `ScoreSigner` 成绩 SHA-256 签名（防篡改）
- `PaperAssembler` 试卷组装器
- `ScoreCalculator` 分数计算器

## 二、应用层增强

### 2.1 收藏/错题本 (`favorite_app.go`)

#### 深度功能

- **批量收藏事务**：批量添加时使用 GORM Transaction，失败回滚
- **错题本智能筛选**：按掌握度、已复习状态筛选
- **自动收藏到"错题本"**：错题自动入库到系统文件夹
- **事件驱动副作用**：所有操作发布领域事件
- **掌握度分布统计**：`GetWrongStats` 接口

#### 关键 API

```go
// 批量收藏
result, err := favApp.BatchAdd(ctx, &BatchAddReq{
    UserID:      uid,
    QuestionIDs: qids,
    FolderID:    folderID,
})
// result: { AddedIDs: [...], SkippedIDs: [...] }

// 错题本
list, total, err := favApp.GetWrongBook(ctx, uid, WrongBookQuery{
    MasteryLevel: 3,
    IsReviewed:   &false,
    Page:         1,
    PageSize:     20,
})

// 标记复习
err := favApp.MarkReviewed(ctx, logID, masteryLevel)
```

### 2.2 用户应用 (`user_app.go`)

#### 安全增强

- **登录限流**：5 分钟内 5 次失败锁定（Redis `INCR`）
- **Token 黑名单**：登出加入 Redis 黑名单
- **密码强度**：使用 `valueobject.Password` 强制
- **随机密码生成**：12 位含大小写+数字
- **Token 适配器**：`JwtTokenAdapter` 隔离 JWT 实现

### 2.3 考试应用 (`exam_app.go`)

#### 防作弊增强

- **事件驱动审计**：每次防作弊事件发布到 EventBus
- **实时统计**：`<examRecord>.AuditSummary` JSON 字段
- **自动判作弊**：切屏≥3/退出全屏≥2/Devtools≥1 即标记

## 三、前端深度优化

### 3.1 useFavorite 组合式 API

**位置**: `frontend/src/composables/useFavorite.ts`

特性：
- 切换/批量收藏
- 错题本智能筛选（按掌握度/复习状态）
- 收藏统计（总数、按类型、按文件夹、近一周）
- 防抖自动保存

### 3.2 exam Pinia Store

**位置**: `frontend/src/store/modules/exam.ts`

特性：
- 答题进度实时持久化（localStorage）
- 倒计时管理（自动超时提交）
- 防作弊事件统计（切屏/全屏/Copy/Paste/Devtools）
- 断线续考支持

### 3.3 user Pinia Store

**位置**: `frontend/src/store/modules/user.ts`

特性：
- Token 持久化 + 自动 refresh
- 前端登录失败锁定（5 次/5 分钟）
- Token 过期检测（剩余 5 分钟提醒）

### 3.4 Favorites.vue 深度重写

- 收藏夹侧边栏（全部/错题本/自定义）
- 多维度筛选（类型/搜索）
- 批量操作（多选+批量移除+批量移动）
- 创建收藏夹对话框（颜色+图标）

### 3.5 WrongBook.vue 深度重写

- 智能概览（累计/未掌握/已掌握/平均掌握度）
- 掌握度筛选（全部/未掌握/部分掌握/已掌握）
- 已复习筛选
- 智能搜索（防抖）
- 错题详情对话框（含题目/选项/你的答案/正确答案/解析）
- 错题练习模式（开发中）

## 四、安全防护体系

| 防护层 | 实现 | 文件 |
| --- | --- | --- |
| 登录限流 | Redis `INCR` + `EXPIRE` | `infrastructure/cache/ratelimit.go` |
| Token 黑名单 | Redis SET | `infrastructure/cache/ratelimit.go` |
| 密码强度 | valueobject 校验 | `domain/valueobject/valueobject.go` |
| 成绩签名 | SHA-256 + 随机盐 | `domain/service/service.go` |
| 防作弊 | 事件驱动 + 阈值 | `application/exam/exam_app.go` |
| RBAC | 中间件 + 路由分组 | `interfaces/router/router.go` |

## 五、性能与可用性

### 5.1 缓存策略

- **考试进度**：Redis Hash + 4 小时 TTL
- **登录限流**：Redis INCR + 5 分钟 TTL
- **Token 黑名单**：Redis SET + 跟随 Token 过期

### 5.2 数据库优化

- **复合唯一索引**：`(exam_id, user_id)` 防止重复记录
- **软删除**：`deleted_at` 字段
- **审计汇总**：JSON 字段，避免多表关联

### 5.3 事务一致性

- 批量收藏使用 GORM `Transaction`
- 考试提交事务（更新记录 + 发布事件）

## 六、测试覆盖

| 包 | 测试文件 | 用例数 |
| --- | --- | --- |
| domain/entity | entity_test.go | 9 |
| domain/event | event_test.go | 4 |
| domain/service | service_test.go | 7 |
| domain/valueobject | valueobject_test.go | 9 |
| pkg/encrypt | encrypt_test.go | 3 |
| pkg/jwt | jwt_test.go | 5 |
| pkg/response | response_test.go | 4 |
| pkg/utils | utils_test.go | 6 |

**总计**: 47 个单元测试，全部通过。

## 七、API 端到端验证

15 个 API 端点全部通过：

- 11 个管理端（用户/统计/题库/试卷/考试/收藏/错题本）
- 4 个学员端（可参加考试/错题本/收藏/学习统计）
