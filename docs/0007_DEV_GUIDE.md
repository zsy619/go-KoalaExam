# 🛠 开发指南

> KoalaExam 项目开发规范，含环境准备、开发流程、代码规范、测试与发布。

---

## 一、环境准备

### 1.1 必备工具

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | 1.21+ | 后端编译 |
| Node.js | 20+ / pnpm | 前端构建 |
| Docker / Docker Compose | 24+ | 本地基础设施 |
| Git | 2.30+ | 版本管理 |
| Make | - | 命令简化 |
| IDE | GoLand / VS Code | 推荐 |
| Air | latest | Go 热重载 |
| golangci-lint | latest | 代码检查 |
| swag | latest | API 文档生成 |

### 1.2 一键安装脚本

```bash
# 后端工具链
cd koala-exam-backend
make deps   # 安装 air / wire / golangci-lint / swag

# 前端依赖
cd ../koala-exam-frontend
pnpm install
```

### 1.3 推荐 IDE 插件

**VS Code**：
- Vue Language Features (Volar)
- ESLint
- Prettier
- Go
- Go Test Explorer
- GitLens
- Docker

**GoLand**：
- 内置功能已足够，无需额外插件

---

## 二、启动开发环境

### 2.1 启动基础设施

```bash
# 项目根目录
cd go-KoalaExam

# 启动 MySQL + Redis（端口 3306 / 6379）
docker compose up -d

# 查看日志
docker compose logs -f mysql

# 停止
docker compose down
```

### 2.2 启动后端（开发模式）

```bash
cd koala-exam-backend

# 拉取依赖
go mod tidy

# 初始化数据库表结构（首次）
make migrate-up

# 启动（热重载）
make air
# 等价于：air -c .air.toml

# 普通启动
make run

# 编译二进制
make build       # 输出 bin/koala-exam
```

服务地址：http://localhost:8080

Swagger 文档：http://localhost:8080/swagger/index.html

### 2.3 启动前端

```bash
cd koala-exam-frontend

# 安装依赖（首次）
pnpm install

# 启动 dev server
pnpm dev
# → http://localhost:5173

# 类型检查
pnpm type-check

# 代码风格检查
pnpm lint

# 构建生产版本
pnpm build     # → dist/
```

### 2.4 访问入口

| 服务 | 地址 | 默认账号 |
|------|------|----------|
| 前端 | http://localhost:5173 | - |
| 后端 | http://localhost:8080 | - |
| Swagger | http://localhost:8080/swagger/index.html | - |
| Grafana | http://localhost:3000 (生产) | admin/admin |
| MySQL | localhost:3306 | koala / koala123 |
| Redis | localhost:6379 | koala123 |

---

## 三、目录结构

### 3.1 后端

详见 [0002_ARCHITECTURE.md §3.1](0002_ARCHITECTURE.md#31-目录结构)

### 3.2 前端

详见 [0002_ARCHITECTURE.md §4.1](0002_ARCHITECTURE.md#41-目录结构)

---

## 四、开发流程

### 4.1 添加新接口（标准 6 步）

```
1. domain/entity/     添加实体（如 MyEntity）
2. infrastructure/repository/  添加仓储（CRUD）
3. application/dto/   添加请求/响应 DTO
4. application/<module>/  添加应用服务（UseCase）
5. interfaces/handler/  添加 Handler（HTTP 绑定）
6. interfaces/router/router.go  注册路由
7. （可选）添加 swag 注释 → swag init 自动生成文档
```

完整示例：

```go
// 1. entity
package entity
type MyEntity struct {
    ID    uint64 `gorm:"primaryKey"`
    Name  string `gorm:"size:64"`
    CreatedAt time.Time
}

// 2. repository
package repository
type MyEntityRepository interface {
    GetByID(ctx context.Context, id uint64) (*entity.MyEntity, error)
    List(ctx context.Context, page, pageSize int) ([]*entity.MyEntity, int64, error)
    Create(ctx context.Context, e *entity.MyEntity) error
}

// 3. dto
package dto
type CreateMyEntityReq struct {
    Name string `json:"name" validate:"required,min=2,max=32"`
}

// 4. service
package service
func (s *MyEntityApp) Create(ctx context.Context, req *dto.CreateMyEntityReq) (*entity.MyEntity, error) {
    e := &entity.MyEntity{Name: req.Name}
    if err := s.repo.Create(ctx, e); err != nil {
        return nil, err
    }
    return e, nil
}

// 5. handler
// @Summary 创建我的实体
// @Router /api/v1/my-entities [POST]
func (h *MyEntityHandler) Create(ctx context.Context, c *app.RequestContext) {
    var req dto.CreateMyEntityReq
    if err := c.BindAndValidate(&req); err != nil {
        response.Error(ctx, errcode.ErrInvalidParam.WithErr(err))
        return
    }
    e, err := h.app.Create(ctx, &req)
    if err != nil {
        response.Error(ctx, errcode.FromError(err))
        return
    }
    response.OK(ctx, e)
}

// 6. router
router.POST("/api/v1/my-entities", auth.AuthRequired(), handler.Create)
```

### 4.2 添加新页面（前端 5 步）

```
1. src/api/modules/  添加 API
2. src/types/entity.ts  添加 TS 类型
3. src/views/<module>/  添加页面组件
4. src/router/index.ts  注册路由
5. src/layouts/MainLayout.vue  加入菜单
```

### 4.3 数据库变更

```bash
# 1. 在 entity 修改结构
vim koala-exam-backend/internal/domain/entity/user.go

# 2. 生成迁移 SQL（推荐 goose）
goose create add_user_avatar sql

# 3. 编写 up + down
vim migrations/00010_add_user_avatar.sql

# 4. 应用迁移
make migrate-up

# 5. 回滚（如需）
make migrate-down

# 开发环境快速重置（清空数据）
make migrate-reset
```

---

## 五、代码规范

### 5.1 Go 代码规范

遵循 [Effective Go](https://go.dev/doc/effective_go) + [Uber Go Style Guide](https://github.com/uber-go/guide)

强制检查（CI 阻断）：
- `golangci-lint run` （启用 govet / errcheck / ineffassign / staticcheck）
- `gofmt -s`

**命名**：
- 包名小写单词：`exam`、`favorite`
- 公开 PascalCase：`ExamService`
- 私有 camelCase：`examRepo`
- 常量驼峰 + 分组：`const ( ErrNotFound = ...  ErrInvalid = ... )`

**错误处理**：
- 不要 `errors.New("xxx")` 后直接吞掉
- 用 `fmt.Errorf("xxx: %w", err)` 包装
- 用 `errcode.ErrXxx.WithErr(err)` 关联业务码

**注释**：
- 公开函数 / 类型必须有 godoc 注释（以函数名开头）
- 复杂逻辑加 `//` 行内注释说明 WHY

### 5.2 TypeScript / Vue 规范

强制（CI 阻断）：
- `eslint --max-warnings=0`
- `vue-tsc --noEmit` 类型检查

**命名**：
- 组件 PascalCase：`ExamQuestion.vue`
- 组合式 camelCase：`useExamTimer.ts`
- 类型 PascalCase：`ExamRecord`
- 变量 camelCase：`currentRecord`
- 常量 UPPER_SNAKE：`MAX_QUESTIONS`

**组合式 API 优先**：
- 使用 `<script setup lang="ts">` 语法
- 业务逻辑抽到 `composables/`
- Props 用 `defineProps<{}>()` 类型化

---

## 六、测试

### 6.1 后端测试

```bash
# 单元测试
make test
# 等价：go test ./... -v -count=1

# 覆盖率
make test-coverage
# 输出 coverage.html（浏览器打开）

# 性能基准
go test -bench=. ./...

# 模糊测试（Go 1.18+）
go test -fuzz=FuzzTest -fuzztime=30s ./pkg/encrypt/
```

### 6.2 测试组织

```
xxx.go           → xxx_test.go        （同包）
xxx.go           → xxx_internal_test.go （仅内部）

表驱动测试：
  tests := []struct{ name string; ... }{ ... }
  for _, tt := range tests { t.Run(tt.name, func(t *testing.T) { ... }) }
```

### 6.3 Mock & 数据库

```
- 单元测试：mock repository（用 gomock / sqlmock）
- 集成测试：testcontainers 起 MySQL/Redis，或用 sqlite
- E2E 测试：air 起服务 + 真实请求
```

### 6.4 前端测试

```bash
# 单元（Vitest）
pnpm test:unit

# E2E（Playwright）
pnpm test:e2e

# Lint + 类型
pnpm lint && pnpm type-check
```

---

## 七、提交规范（Conventional Commits）

### 7.1 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 7.2 Type

| Type | 用途 |
|------|------|
| feat | 新功能 |
| fix | 修复 Bug |
| docs | 文档 |
| style | 格式（不影响代码） |
| refactor | 重构 |
| perf | 性能优化 |
| test | 测试 |
| chore | 构建 / 工具 |
| ci | CI/CD |

### 7.3 示例

```
feat(exam): 支持遗传算法出卷

- 新增 pkg/paper/ga.go 实现遗传算法组卷
- 试卷表 strategy=3 时启用 GA
- 组卷平均耗时从 50ms 降至 200ms（复杂试卷）

Closes #123

---

fix(favorite): 修复错题自动收藏幂等性问题

- 之前：同一错题可能重复入库
- 修复：使用 INSERT ... ON DUPLICATE KEY UPDATE

Reviewed-by: alice
```

### 7.4 分支策略

```
main        ← 发布分支（受保护）
develop     ← 集成分支
feature/*   ← 功能开发
fix/*       ← 紧急修复
release/*   ← 发布准备

流程：feature → develop → release → main
```

---

## 八、PR 规范

### 8.1 Checklist

PR 提交前自检：

- [ ] 单元测试通过（覆盖率 > 70%）
- [ ] Lint + 类型检查无 error
- [ ] 已添加/更新文档（接口、注释）
- [ ] 已关联 Issue
- [ ] 数据库迁移有 up + down
- [ ] 配置变更已在 docs 中说明
- [ ] 截图 / 录屏（前端 UI 变更）

### 8.2 Code Review 要点

- 业务逻辑正确性
- 性能影响（DB 查询、缓存使用）
- 安全（输入校验、权限、SQL 注入）
- 可观测（日志、指标）
- 测试覆盖

---

## 九、常见任务速查

### 9.1 添加新角色权限

```go
// 1. domain/consts/role.go
const RoleReviewer = 4  // 阅卷员

// 2. middleware/RequireRole 识别新角色
// 3. 数据库迁移：ALTER TABLE ems_user MODIFY role TINYINT;
```

### 9.2 添加新题型

```go
// 1. domain/consts/question_type.go
const QuestionTypeEssay = 7

// 2. application/grading/auto_grade.go 增加 switch case
// 3. 前端 components/QuestionRenderer.vue 增加类型分支
```

### 9.3 修改 JWT 密钥

```bash
# 1. 生成新密钥（32+ 字符）
openssl rand -hex 32

# 2. 更新配置（生产用 KMS 注入）
# config.yaml: jwt.secret: "新密钥"

# 3. 重启服务 + 强制用户重新登录
```

### 9.4 数据导出

```go
// pkg/excel/exporter.go
func ExportExamRecords(records []*entity.ExamRecord, w io.Writer) error {
    f := excelize.NewFile()
    // 写表头
    // 写数据
    return f.Write(w)
}
```

---

## 十、调试技巧

### 10.1 本地调试后端

```bash
# 1. 启用 pprof
go build -tags pprof -o bin/koala-exam ./cmd/hertz

# 2. 访问火焰图
# go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# 3. 用 Delve 调试
dlv debug ./cmd/hertz
```

### 10.2 查看 SQL 日志

```yaml
# config.yaml
mysql:
  log_level: info  # debug / info / warn / error / silent
```

### 10.3 调试 Redis

```bash
# 查看答题进度
redis-cli -a koala123 HGETALL koala:exam:progress:12345

# 清空某用户缓存
redis-cli -a koala123 --scan --pattern "koala:*:user:100" | xargs redis-cli DEL
```

### 10.4 前端调试

```
Chrome DevTools → Vue Devtools 扩展 → 切换组件查看状态
Network 面板 → 筛选 XHR → 看接口耗时 / Payload
Console → import { useUserStore } from '@/store'  // 调试
```

---

> 相关文档：[0002 架构](0002_ARCHITECTURE.md) · [0006 API](0006_API.md) · [0008 部署](0008_DEPLOY.md)