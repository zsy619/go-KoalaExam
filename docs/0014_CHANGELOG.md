# 📝 变更日志（CHANGELOG）

> 所有版本变更均按 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/) 规范记录。
> 版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)（SemVer）。

---

## [Unreleased] - 计划中

### Documentation（文档 - 本次更新）
- 📚 新增 [0015_EXAM_LIFECYCLE.md](0015_EXAM_LIFECYCLE.md)（666 行）—— 考试全生命周期专题
  - 考试状态机（草稿 / 未发布 / 进行中 / 已结束）
  - 开始考试全流程（含断线续考 / 题目乱序 / 学员端答案隐藏）
  - 三级存储进度保存（LocalStorage → Redis Hash 4h TTL → MySQL）
  - 交卷 / 阅卷 / 成绩签名 / 重考机制
  - cron 状态推进（K8s CronJob 示例）
  - 10 项已知技术债与改进清单
- 🛡 新增 [0016_ANTI_CHEAT.md](0016_ANTI_CHEAT.md)（582 行）—— 防作弊专题
  - 14 项防作弊措施全景图（前端拦截 / 后端审计 / AI 监考）
  - 切屏 / 全屏 / 复制粘贴 / DevTools 检测代码
  - 后端审计 + v1.1 拆表改进
  - 可疑模式判定（秒答全对 / 多 IP 切换 / 同一答案）
  - 教师监考大屏 SSE 推送方案
  - AI 智能监考（人脸识别 / 屏幕录制 / 鼠标轨迹）
  - 防作弊与隐私合规（个人信息保护法 / GDPR / 等保 2.0）
- ⚖️ 新增 [0017_GRADING.md](0017_GRADING.md)（691 行）—— 阅卷与判分专题
  - v1.0 同步阅卷 vs v1.1 异步阅卷（Redis Stream）对比
  - 客观题 7 种题型评分算法（单选 / 多选 / 判断 / 填空 / 不定项 / 编程 / 简答）
  - 答案比对：等值 / 集合匹配 / 字符串模糊匹配
  - 主观题人工评分 + AI 辅助（GPT-4）
  - 错题本自动收录 + 掌握度动态算法
  - SHA-256 成绩签名 + KMS 注入改进
  - 编程题沙箱评测（Judge0 对接）
  - 阅卷质量评估指标
- 🆕 0001_README.md 索引更新：新增「🎓 考试专题（深度）」分组，考试模块开发人员阅读路径单独标记
- 📋 文档约定更新：专题分组规范（考试 0015-0019、其他专题 0020+ 预留）

### 计划
- 阅卷异步化（Redis Stream Worker）
- 数据库读写分离
- 答题 WebSocket 推送
- OpenTelemetry 链路追踪
- 弱网 PWA 离线答题

---

## [v1.0.0] - 2024-12-15

### 🎉 首次发布（GA）

### Added（新增）
- 用户 / 权限：超管 / 教师 / 学生三级 RBAC + JWT
- 题库：6 种题型 + Markdown 富文本 + Excel 批量导入
- 试卷：手动 / 随机 / 遗传算法 三种组卷策略
- 考试：限时 + 防作弊（切屏 / 全屏 / 复制粘贴 / 乱序）
- 阅卷：客观题自动 + 主观题人工 + 主观题 AI 辅助
- 统计：ECharts 可视化 + Excel 导出
- 深度收藏 / 错题本（核心）
- 后端：Hertz + Netpoll + GORM + MySQL + Redis
- 前端：Vue 3 + TypeScript + Vite + Pinia + Element Plus
- 部署：Docker Compose + K8s Manifest

### Security（安全）
- bcrypt(cost=12) 密码哈希
- JWT 双 Token（Access 2h + Refresh 7d）
- HMAC-SHA256 接口签名（关键接口）
- 限流中间件（基于 Redis 滑动窗口）
- TLS 1.3 + HSTS
- 审计日志（ems_audit_log）
- 数据脱敏（手机/身份证/邮箱）

### Performance（性能）
- 单实例 3,000 同时在线
- QPS：login 6,800 / answer 18,000
- P99 < 500ms

### Documentation（文档）
- 知识库 14 篇（架构/技术栈/功能/数据库/API/开发/部署/风险/性能/Runbook/安全/路线图/变更）

---

## [v0.9.0] - 2024-11-30（RC 候选版本）

### Added
- 完整的核心考试闭环
- 阅卷基础功能
- 错题本 V1

### Fixed
- Redis 答题进度偶尔丢失（增加 LocalStorage 双写）
- 遗传算法大题量下超时（增加超时中断）

### Changed
- 数据库表结构从 snake_case 统一为 ems_ 前缀

---

## [v0.5.0] - 2024-10-20（内测版）

### Added
- 登录 / 用户管理
- 题库基础功能
- 简单组卷

### Known Issues
- 并发 1000+ 时 MySQL 连接池偶发耗尽（v0.9 已修复）
- 考试同时开始存在白屏（v0.9 已修复）

---

## [v0.1.0] - 2024-08-01（项目初始化）

### Added
- 仓库初始化（README + 文档骨架）
- 后端 Hertz + GORM 脚手架
- 前端 Vue3 + Vite 脚手架
- Docker Compose MySQL + Redis

---

## 版本对比表

| 版本 | 时间 | 关键变化 | 在线人数 | QPS |
|------|------|----------|----------|-----|
| v0.1 | 2024-08 | 初始化 | - | - |
| v0.5 | 2024-10 | 内测 | 100 | 500 |
| v0.9 | 2024-11 | RC | 1,000 | 5,000 |
| **v1.0** | **2024-12** | **GA** | **3,000** | **18,000** |
| v1.1 | 2025-Q1 | 性能 | 10,000 | 30,000 |
| v2.0 | 2025-Q3 | AI | 30,000 | 50,000 |

---

## 升级指南

### v0.9 → v1.0

```bash
# 1. 备份
mysqldump -u root -p koala_exam > koala_v0.9_backup.sql

# 2. 数据库迁移（ems_ 前缀统一）
mysql -u root -p koala_exam < migrations/v0.9-to-v1.0.sql

# 3. 升级后端镜像
kubectl set image deployment/koala-exam-backend backend=koala-exam-backend:v1.0.0

# 4. 前端强刷缓存（vite build 文件名带 hash，无需）
```

### v1.0 → v1.1（计划）

- [ ] 启用 Redis Stream 阅卷队列（需部署 koala-grading-worker）
- [ ] 配置主从数据库（读写分离）
- [ ] 启用 OpenTelemetry Collector

---

## 贡献与规范

- 每次发布由 Release Manager 整理
- PR 必须关联 Issue
- 破坏性变更需在 [Unreleased] - 计划中 标注 BREAKING CHANGE

---

> **维护者**：Release Manager / 每两周一次小版本，每月一次次版本，每季度一次主版本