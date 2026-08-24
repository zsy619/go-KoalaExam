# 🐨 KoalaExam Backend

基于 **Hertz** + **GORM** + **MySQL** + **Redis** 的在线考试系统后端（Clean Architecture + DDD）。

## 📂 目录

```
cmd/
├── hertz/      # 主服务入口（端口 8080）
└── migrate/    # 数据库迁移 CLI（up/down/reset/seed）
internal/
├── application/    # 应用服务（业务编排）
├── domain/         # 领域模型（实体/常量/错误码）
├── infrastructure/ # 基础设施（DB/Redis/仓储）
└── interfaces/     # 接口层（Handler/中间件/路由）
pkg/                # 公共库（JWT/响应/加密/日志）
configs/            # YAML 配置
migrations/         # SQL 迁移
```

## 🚀 启动

```bash
# 1. 启动基础设施（项目根目录）
cd ..
docker compose up -d mysql redis

# 2. 启动后端（开发模式）
cd koala-exam-backend
go mod tidy
make run       # 或 make air（热重载）

# 3. 访问 API
curl http://localhost:8080/health
```

## 📡 主要路由

所有路由前缀 `/api/v1`。

### 公开
- `POST /auth/login`  登录
- `POST /auth/refresh`  刷新 token

### 鉴权
- `GET /user/profile`  个人信息
- `GET /exams/available`  可参加的考试
- `POST /exams/:id/start`  开始考试
- `POST /exams/answer`  保存答题
- `POST /exams/submit`  交卷

### 管理员（超管）
- `POST /admin/users`  创建用户
- `GET /admin/users`  用户列表

### 教师/超管
- `POST /questions`  创建题目
- `POST /papers`  创建试卷（手动/随机/GA）
- `POST /exams`  创建考试

### 深度收藏（核心）
- `POST /favorites/toggle`  单个收藏
- `POST /favorites/batch`  批量收藏（错题自动入库）
- `GET /favorites/check`  查询是否收藏
- `GET /wrong-book`  错题本（带掌握度筛选）
- `POST /wrong-log/:id/reviewed`  标记已复习

## 🔧 开发工具

```bash
make air          # 热重载
make migrate-up   # 迁移
make migrate-down # 回滚
make test         # 测试
make lint         # 代码检查
```

## 🐳 Docker

```bash
docker build -t koala-exam-backend .
docker run --rm -p 8080:8080 koala-exam-backend
```
