# KoalaExam 代码规范 (Google Go Style + DDD)

本文档定义 KoalaExam 项目的代码风格与架构原则,所有贡献者必须遵循。

## 分层架构 (DDD)

KoalaExam 采用 **领域驱动设计 (DDD)** 的分层架构,从内到外:

- **interfaces/** - HTTP handlers, middleware, routers (最外层)
- **application/** - 用例编排 (Use Cases)
- **domain/** - entities, repositories, consts, errcode (核心业务规则)
- **infrastructure/** - GORM 实现, Redis 缓存, JWT

### 各层职责

- **interfaces** - 处理 HTTP 请求、参数校验、响应格式化
- **application** - 编排领域对象完成用例,不含业务规则
- **domain** - 实体、值对象、领域仓储接口、常量、错误码
- **infrastructure** - 具体技术实现(GORM、Redis、JWT)

> **依赖方向**: 所有依赖都指向内层(domain)。基础设施通过实现 domain 的接口被注入。

## Google Go 风格

### 1. 包命名

- 包名小写、简短、有意义(避免 utils、helpers)
- 包名应该是名词(user、exam),不是动词
- 包名是源目录的最后一个元素

### 2. 命名约定

#### 变量

```go
// 推荐: 短而清晰
for i := 0; i < 10; i++ { ... }
for k, v := range m { ... }
buf := make([]byte, 1024)

// 避免: 冗长或类型前缀
userIndex := 0
i := 0  // 含义不明
```

#### 函数

```go
// 推荐: 动词+名词
func (a *ExamApp) StartExam(ctx context.Context, examID, userID int64) (*StartExamResp, error)
func (a *ExamApp) SaveAnswer(ctx context.Context, req *SaveAnswerReq) error

// 避免: 名词或缩写
func (a *ExamApp) DoExam(...)
func (a *ExamApp) ExamStart(...)
```

#### 接口

```go
// 推荐: 行为命名(动词)
type Reader interface { Read(p []byte) (n int, err error) }
type TokenService interface { Generate(...) }

// 避免: I 前缀或类名
type IReader interface { ... }
type UserService interface { ... }
```

#### 接收器

```go
// 推荐: 1-2 个字母的接收器,与类型相关
func (a *ExamApp) Create(...)     // a = app
func (u *User) FullName() string  // u = user
func (s *Server) Start()          // s = server

// 避免: this、self 或类型名
func (this *ExamApp) Create(...)
func (ExamApp) Create(...)
```

### 3. 注释

```go
// Package exam 考试应用服务。
//
// 遵循 Google Go 风格:
//   - 命名简洁(StartExam、SaveAnswer 等动词+名词)
//   - 显式错误返回(error 作为最后一个返回值)
//   - context 作为第一个参数传递
package exam

// StartExam 开始考试,支持断线续考。
//
// 已存在的记录若处于进行中状态会被复用,否则返回 ExamSubmitted 错误。
func (a *ExamApp) StartExam(ctx context.Context, examID, userID int64) (*StartExamResp, error)
```

### 4. 函数签名

#### context 作为第一个参数

```go
// 推荐
func (a *ExamApp) Login(ctx context.Context, username, password string) error

// 避免
func (a *ExamApp) Login(username, password string, ctx context.Context) error
```

#### error 作为最后一个返回值

```go
// 推荐
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error)

// 避免
func (r *UserRepository) GetByID(ctx context.Context, id int64) (error, *User)
```

### 5. 错误处理

```go
// 推荐: 显式包装错误
if err != nil {
    return fmt.Errorf("start exam %d: %w", examID, err)
}

// 推荐: 自定义业务错误
if user.Status == consts.UserStatusDisabled {
    return errcode.New(errcode.CodeUserDisabled, "用户已禁用")
}

// 避免: 吞掉错误
_ = doSomething()

// 避免: panic 用于业务错误
if !found {
    panic("user not found")
}
```

### 6. 构造函数

```go
// 推荐: 返回指针 + error
func NewExamApp(repo ExamRepository, ...) (*ExamApp, error)

// 推荐: 依赖注入,不做额外初始化
func NewExamApp(repo ExamRepository) *ExamApp {
    return &ExamApp{repo: repo}
}
```

### 7. 结构体

```go
// 推荐: 相关字段放一起,按重要性排序
type CreateExamReq struct {
    Title       string    // 必填
    Description string    // 可选
    PaperID     int64     // 必填
    StartTime   time.Time
    EndTime     time.Time
    Duration    int
    MaxAttempts int
    ShuffleQ    bool
}
```

## 目录结构

```
backend/
├── cmd/                    # 程序入口
│   ├── hertz/              # HTTP 服务
│   └── migrate/            # 数据库迁移
├── configs/                # YAML 配置
├── docs/                   # 文档
├── internal/
│   ├── application/        # 应用服务(用例编排)
│   │   ├── dto/            # 请求/响应 DTO
│   │   ├── user/           # 用户用例
│   │   ├── exam/           # 考试用例
│   │   └── ...
│   ├── domain/             # 领域核心
│   │   ├── consts/         # 业务常量
│   │   ├── entity/         # 实体(含 TableName)
│   │   ├── errcode/        # 错误码
│   │   └── repository/     # 仓储接口
│   ├── infrastructure/     # 技术实现
│   │   ├── cache/          # Redis 缓存
│   │   ├── database/       # GORM 初始化
│   │   └── repository/     # GORM 仓储实现
│   └── interfaces/         # HTTP 接口
│       ├── handler/        # HTTP handlers
│       ├── middleware/     # 中间件
│       └── router/         # 路由
├── pkg/                    # 通用工具(无业务依赖)
│   ├── config/             # 配置加载
│   ├── encrypt/            # 加密
│   ├── jwt/                # JWT
│   ├── logger/             # 日志
│   ├── response/           # 统一响应
│   └── utils/              # 工具函数
└── migrations/             # SQL 迁移文件
```

## 测试

### 单元测试

```go
package user

import "testing"

func TestBcryptPassword(t *testing.T) {
    hash, err := encrypt.BcryptPassword("koala123")
    if err != nil {
        t.Fatalf("BcryptPassword() error = %v", err)
    }
    if !encrypt.BcryptCheck(hash, "koala123") {
        t.Error("BcryptCheck() = false, want true")
    }
}
```

## 关键规则

1. **数据库表前缀** `ke_` - 所有表必须使用(entity.TableName + SQL DDL)
2. **时间字段** `CreatedAt/UpdatedAt` - GORM 自动维护
3. **软删除** - 使用 `DeletedAt gorm.DeletedAt`
4. **JSON 字段** - 不能为空字符串,使用 `nonEmptyJSON` 辅助函数
5. **权限控制** - 在 router 层通过 middleware 强制
6. **错误码** - 通过 `errcode.New(code, msg)` 返回业务错误
7. **Redis 键** - 集中在 `infrastructure/cache/redis.go` 定义

## 参考资料

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)


## 领域层核心构件

### 值对象 (Value Object)

值对象是不可变的、通过值相等比较的对象，用于封装业务规则：

```go
// 推荐：值对象
type Password string

func NewPassword(raw string) (Password, error) {
    // 业务规则校验
    if len(raw) < 8 {
        return "", errors.New("password too short")
    }
    return Password(raw), nil
}
```

KoalaExam 的值对象：
- `ExamPaper` 试卷快照
- `UserAnswers` 用户答案
- `ExamScore` 考试分值
- `AuditSummary` 防作弊汇总
- `TimeWindow` 时间窗口
- `Password` 密码强度

### 领域事件 (Domain Event)

领域事件代表已发生的事实，用于解耦副作用：

```go
// 定义事件
type ExamSubmittedEvent struct {
    UserID   int64
    RecordID int64
    Score    float64
}

func (e *ExamSubmittedEvent) EventName() string { return "exam.submitted" }

// 发布
_ = bus.Publish(ctx, &ExamSubmittedEvent{UserID: 1, RecordID: 2})

// 订阅
bus.Subscribe("exam.submitted", HandlerFunc(func(ctx context.Context, e Event) error {
    // 处理副作用（通知、统计等）
    return nil
}))
```

### 领域服务 (Domain Service)

处理不属于某个实体的业务逻辑：

```go
type AnswerComparator struct{}
func (c *AnswerComparator) Compare(correct, user interface{}) bool {
    // 跨实体业务逻辑
}
```

## 安全防护

### 1. 登录限流

- 5 分钟内最多 5 次失败
- 超过后锁定 5 分钟
- 通过 Redis `INCR` + `EXPIRE` 实现滑动窗口

### 2. Token 黑名单

- 登出时加入 Redis 黑名单
- TTL = token 剩余有效期
- 鉴权中间件每次检查黑名单

### 3. 密码强度

- 至少 8 位
- 包含字母+数字（至少 2 种字符类型）
- 通过 `valueobject.Password` 强制

### 4. 防作弊

- 切屏次数 ≥ 3 → 标记作弊
- 退出全屏 ≥ 2 次 → 标记作弊
- 开发者工具打开 ≥ 1 次 → 标记作弊
- 所有事件写入 `audit_log` 字段
