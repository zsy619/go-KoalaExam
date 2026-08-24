# 🐨 KoalaExam（考拉智测）

> 基于 **Golang (Hertz)** + **Vue 3** 的企业级在线考试系统

![arch](https://img.shields.io/badge/arch-clean--architecture-blue) ![backend](https://img.shields.io/badge/backend-hertz-00ADD8) ![frontend](https://img.shields.io/badge/frontend-vue3-42b883) ![db](https://img.shields.io/badge/db-mysql-4479A1) ![table](https://img.shields.io/badge/table-ke__-orange) ![license](https://img.shields.io/badge/license-MIT-green)

## 🎯 项目简介

**KoalaExam** 是一套面向 K12 / 高校 / 企业培训场景的**在线考试系统**，采用前后端分离架构 + Clean Architecture 分层 + GORM 实现。

- **后端**：字节跳动开源的 **Hertz** 框架（基于自研高性能网络库 **Netpoll**）+ **GORM** + MySQL（`KoalaExam` 库，`ke_` 前缀）+ Redis
- **前端**：**Vue 3 + TypeScript + Vite + Pinia + Element Plus**，现代化、类型安全

## 📂 仓库结构

```
koala-exam/
├── koala-exam-backend/     # Go + Hertz 后端
│   ├── cmd/
│   │   ├── hertz/          # 主服务入口（端口 8080）
│   │   └── migrate/        # 数据库迁移 CLI（up/down/reset/seed/fresh）
│   ├── internal/
│   │   ├── application/   # 应用服务（业务编排）
│   │   ├── domain/        # 领域模型（实体/常量/错误码）
│   │   ├── infrastructure/# 基础设施（MySQL/Redis/仓储）
│   │   └── interfaces/    # 接口层（Handler/中间件/路由）
│   ├── pkg/               # 公共库（JWT/响应/加密/日志）
│   ├── configs/           # YAML 配置
│   └── migrations/init/   # SQL 初始化脚本
├── koala-exam-frontend/   # Vue3 + Vite 前端
│   └── src/（api/components/composables/layouts/router/store/views）
├── docker-compose.yml     # MySQL + Redis 一键启动
├── scripts/               # start.sh / stop.sh / db.sh
└── docs/                  # 知识库
```

## 🚀 快速开始

### 一键启动（推荐）

```bash
# 1. 启动 MySQL + Redis + 自动建库建表 + 测试数据 + 后端 + 前端
./scripts/start.sh

# 访问
# 前端:  http://localhost:5173
# 后端:  http://localhost:8080
# MySQL: localhost:3306 (root/123456, db=KoalaExam)

# 2. 停止所有服务
./scripts/stop.sh
```

### 手动启动

```bash
# 1. 启动基础设施
docker compose up -d

# 2. 后端：建库建表 + 写入测试数据
cd koala-exam-backend
go mod tidy
make fresh       # 或 make seed（仅写数据）

# 3. 后端：启动服务
make air         # 热重载（推荐）
# 或：make run

# 4. 前端：安装并启动
cd ../koala-exam-frontend
pnpm install
pnpm dev
```

### 数据库管理

```bash
./scripts/db.sh fresh    # 重置数据库（含测试数据）
./scripts/db.sh reset    # 重置表结构（不含数据）
./scripts/db.sh seed     # 仅写入测试数据
./scripts/db.sh init     # 直接执行 SQL（不需要 Go）
./scripts/db.sh shell    # 进入 MySQL Shell
```

## 🔑 默认账号（密码均为 `koala123`）

| 账号 | 角色 | 说明 |
|------|------|------|
| admin | 超管 (role=1) | 全功能管理 |
| teacher | 教师 (role=2) | 题库/试卷/考试管理 |
| teacher2 | 教师 (role=2) | 备用教师账号 |
| student | 学员 (role=3) | 软工1班，含错题/收藏 |
| student2 | 学员 (role=3) | 软工1班 |
| student3 | 学员 (role=3) | 软工2班 |
| student4 | 学员 (role=3) | AI1班 |
| student5 | 学员 (role=3) | 软工1班 |

## 🗄 数据库配置

```yaml
database: KoalaExam      # 数据库名
username: root          # 用户
password: "123456"      # 密码
table_prefix: ke_       # 表前缀
```

### 表清单（12 张，全部 `ke_` 前缀）

| 表名 | 说明 | 关键索引 |
|------|------|----------|
| ke_department | 组织/院系 | idx_parent |
| ke_class | 班级 | idx_dept, idx_teacher |
| ke_user | 用户 | uniq_username, idx_role, idx_status |
| ke_question_category | 题库分类 | idx_parent |
| ke_question | 题目 | idx_category, idx_type, idx_difficulty |
| ke_paper | 试卷 | idx_creator, idx_status |
| ke_paper_question | 试卷-题目关联 | uniq_paper_q |
| ke_exam | 考试 | idx_paper, idx_status, idx_start, idx_end |
| ke_exam_record | 考试记录 | uniq_exam_user, idx_status |
| ke_favorite_folder | 收藏夹 | idx_user |
| ke_favorite | 收藏主表 | uniq_user_target |
| ke_wrong_log | 错题日志 | idx_user_q, idx_mastery |

## 🧪 测试数据概览

`make fresh` 或 `./scripts/db.sh fresh` 将创建：

- **5 个组织**（计算机学院 / 软件工程系 / 人工智能系 / 外语学院 / 经管学院）
- **3 个班级**（软工1班/软工2班/AI1班）
- **8 个用户**（1 超管 + 2 教师 + 5 学员）
- **6 个题库分类**
- **20 道题目**（涵盖 6 种题型：单选/多选/判断/填空/不定项/编程）
- **5 张试卷**（固定 + 随机 + 入门）
- **13 条试卷-题目关联**
- **5 场考试**（含长期开放的可重复考试）
- **12 个收藏夹**（每个学员 2 个系统夹 + 部分自定义夹）
- **5 条收藏**（手动 + 错题自动）
- **7 条错题日志**（含掌握度评级 1-5 星）
- **6 条考试记录**（已批改 + 分数签名）

## 🎯 核心特性

### 功能完整（与原始需求一一对应）
- ✅ 用户/权限（RBAC）：超管/教师/学生三级
- ✅ 题库管理：单选/多选/判断/填空/不定项/编程 6 种题型
- ✅ 智能组卷：手动 + 随机 + 遗传算法
- ✅ 限时考试：倒计时、超时自动交卷、断线续考
- ✅ 防作弊：全屏监控、切屏检测、题目乱序、选项随机、禁止复制粘贴
- ✅ 阅卷：客观题自动 + 主观题人工
- ✅ 成绩统计：ECharts 可视化、SHA-256 防篡改
- ✅ **深度收藏/错题本**：智能归类、掌握度动态评级、薄弱点分析

### 技术亮点
- 🚀 Hertz + Netpoll 高性能
- 🏛 Clean Architecture + DDD 分层
- 🛡 JWT + 中间件链（Recovery / CORS / RateLimit / Audit）
- 💾 Redis 缓存：答题进度、断线续考、限流
- 🔒 数据安全：bcrypt 密码 + SHA-256 成绩签名
- 📈 可观测：Zap 结构化日志

## 📜 License

MIT


## ✅ 最新功能

- 完整后端：50+ Go 文件，支持 Hertz + Netpoll
- **数据库 MySQL `KoalaExam`（表前缀 `ke_`）**：12 张表，含完整 DDL + 测试数据
- 完整前端：Vue 3 + TS + Vite + Pinia + Element Plus
- **统计模块**（Dashboard / ExamOverview / UserLearningStats）
- **个人中心**（资料编辑 + 修改密码）
- 5 个单元测试通过（加密、JWT、响应、工具、实体）
- 一键启动 / 停止 / 数据库管理 Shell 脚本
- Docker Compose 全栈部署（MySQL+Redis+后端+前端）
- 完整文档：架构/技术栈/功能/数据库/API/风险/开发指南/部署指南
