# 🛠 技术栈详解

> KoalaExam 后端 + 前端 + 基础设施完整技术栈，含选型理由、对比分析与最佳实践。

---

## 1. 后端技术栈

### 1.1 总体一览

| 层级 | 组件 | 版本 | 选型理由 |
|------|------|------|----------|
| 语言 | Go | 1.21+ | 高并发、静态编译、生态成熟 |
| HTTP 框架 | Hertz | v0.7.0 | 字节跳动开源、基于 Netpoll、QPS 高 |
| 网络库 | Netpoll | v0.6.0 | 自研 Reactor 模型、Linux epoll 高效 |
| ORM | GORM | v1.25.4 | Go 生态最流行、自动迁移、关联加载 |
| 数据库 | MySQL | 8.0 | 主流稳定、支持 JSON / 窗口函数 |
| 缓存 | Redis | 7.x | 高性能 KV、Hash/Stream 丰富 |
| 认证 | JWT (HS256) | jwt/v5 | 无状态、前端友好、跨域可用 |
| 配置 | Viper | v1.18 | YAML / ENV / 远程配置 / 热加载 |
| 日志 | Uber Zap | v1.26 | uber 开源、结构化、高性能 |
| Excel | Excelize | v2.8 | 纯 Go 实现、读写兼容 Excel 2007+ |
| 加密 | bcrypt + AES | - | 密码慢哈希 + 字段加密 |
| 签名 | SHA-256 | - | 成绩防篡改 |
| 链路追踪 | OpenTelemetry | latest | CNCF 标准、跨语言 |

### 1.2 Hertz vs 主流框架对比

| 框架 | 基准 QPS | 生态 | 学习成本 | 适用场景 |
|------|---------|------|----------|----------|
| **Hertz** | ~50,000 | 字节内部 + 社区 | 中 | 高并发 API |
| Gin | ~40,000 | 非常丰富 | 低 | 通用 |
| Echo | ~35,000 | 中等 | 低 | 通用 |
| Fiber | ~80,000 | 中等 | 中（API 类 Express） | 极致性能 |
| go-zero | ~30,000 | 中等 | 中（DSL） | 微服务 |

**选 Hertz 的核心理由**：

1. **Netpoll 网络库**：Linux 下基于 epoll 的 Reactor 模型，P99 比 stdlib 降低 30%+
2. **字节大规模验证**：抖音/今日头条内部日均万亿级调用
3. **中间件链完善**：Recovery / CORS / JWT / 限流 / Tracer 即插即用
4. **未来扩展**：Thrift / Protobuf IDL 生成，proto-gen 友好
5. **渐进式**：标准库 net/http 接口兼容，可平滑切换

### 1.3 Netpoll 网络库原理

```
Go 原生 net/http：
  每个连接 → 一个 goroutine → epoll (Go runtime)

Hertz + Netpoll：
  N 个连接 → 1 个 goroutine (Reactor) → epoll + 主动监听

收益：
  - goroutine 数降低 50%+
  - 调度开销降低 30%+
  - 长连接 / 高并发场景显著优势

开关配置：
  config.yaml:
    app:
      enable_netpoll: true   # 默认 true；故障时可临时关闭降级
```

### 1.4 GORM 关键实践

```go
// 1. 自动迁移（首次启动）
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &entity.User{},
        &entity.Exam{},
        &entity.ExamRecord{},
        // ...
    )
}

// 2. 连接池调优
func NewDB(cfg *Config) (*gorm.DB, error) {
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(200)
    sqlDB.SetMaxIdleConns(50)
    sqlDB.SetConnMaxLifetime(time.Hour)
    sqlDB.SetConnMaxIdleTime(10 * time.Minute)
    return db, nil
}

// 3. 事务封装
func (s *ExamApp) SubmitExam(...) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Save(record).Error; err != nil {
            return err
        }
        if err := tx.Create(&answerLog).Error; err != nil {
            return err
        }
        return nil
    })
}

// 4. 软删除
// gorm.DeletedAt → 自动生成 deleted_at 字段 + 软删除条件
// 查询需显式 Unscoped() 才能查到
```

### 1.5 替代方案对比

| 组件 | 选型 | 备选 | 选择理由 |
|------|------|------|----------|
| ORM | GORM | ent, sqlx | GORM 生态最成熟；ent 类型安全但学习成本高 |
| 数据库 | MySQL | PostgreSQL | 团队熟悉 + 云厂商兼容好；PG 在 JSON / 全文检索更好 |
| 缓存 | Redis | Memcached, KeyDB | Redis 数据结构丰富（Hash/Stream）；KeyDB 单线程瓶颈 |
| 配置 | Viper | koanf, envconfig | Viper 功能完整 + 热加载；生态最广 |
| 日志 | Zap | zerolog, logrus | Zap 性能最佳 + 结构化字段 |
| JWT | golang-jwt | jwx | 标准 HS256/ES256 都支持；足够 |

---

## 2. 前端技术栈

### 2.1 总体一览

| 层级 | 组件 | 版本 | 选型理由 |
|------|------|------|----------|
| 框架 | Vue 3 (Composition API) | 3.4+ | 逻辑复用 + TS 友好 + 组合式 |
| 语言 | TypeScript | 5.0+ | 类型安全、IDE 提示 |
| 构建 | Vite | 5.0+ | 启动快（ESM + esbuild）、HMR 毫秒级 |
| UI | Element Plus | 2.4+ | 企业级组件丰富、中文友好 |
| 状态 | Pinia | 2.0+ | Vue 3 官方推荐、TS 友好 |
| 路由 | Vue Router | 4.0+ | 嵌套路由 + 守卫 + 动态导入 |
| HTTP | Axios | 1.0+ | 拦截器 + 取消请求 + 自动 JSON |
| 图表 | ECharts | 5.0+ | 折线/柱状/雷达/饼图、大数据量 |
| 自动导入 | unplugin-auto-import | latest | 按需引入 Element Plus / Vue API |
| 工具 | Day.js + Sass | - | 日期 / 样式预处理 |
| 测试 | Vitest + Vue Test Utils | - | 与 Vite 同源、速度快 |
| 规范 | ESLint + Prettier | - | 代码风格统一 |

### 2.2 Vue 3 vs Vue 2 / React

| 维度 | Vue 3 | Vue 2 | React 18 |
|------|-------|-------|----------|
| API 风格 | 组合式（函数式）| Options（对象式）| Hooks（函数式）|
| TS 友好度 | 优秀 | 一般 | 优秀 |
| 学习曲线 | 平缓 | 最平缓 | 中等 |
| 性能 | 优秀（Proxy）| 一般 | 优秀 |
| 国内生态 | 丰富 | 丰富 | 较弱 |
| Element Plus | 支持 | 仅 Element UI | Ant Design |

**选 Vue 3 的核心理由**：

1. **Element Plus**：是国内最成熟的企业级组件库（KPI 场景全覆盖）
2. **Pinia**：类型推导比 Vuex 好太多
3. **组合式 API**：`useExamTimer` / `useAntiCheat` 等可复用逻辑比 mixin 清晰
4. **国内开发者友好**：社区活跃、中文文档丰富

### 2.3 Vite vs Webpack

| 维度 | Vite 5 | Webpack 5 |
|------|--------|-----------|
| 启动时间 | < 1s（ESM）| 10s+ |
| HMR | 毫秒级 | 秒级 |
| 生产构建 | Rollup（轻）| Webpack（重）|
| 配置复杂度 | 简单 | 复杂 |

**Vite 优势**：dev server 基于浏览器原生 ESM，无需打包，秒级冷启动。

### 2.4 状态管理（Pinia vs Vuex）

```ts
// Pinia 优势：
// 1. 完整 TS 类型推导
// 2. 无 mutation（直接修改 state）
// 3. 支持 Composition API（setup store）

// 例：考试全局态
export const useExamStore = defineStore('exam', () => {
  const currentRecord = ref<ExamRecord | null>(null)
  const answers = ref<Map<number, Answer>>(new Map())

  function saveAnswer(qid: number, ans: Answer) {
    answers.value.set(qid, ans)
  }

  const progress = computed(() => answers.value.size)

  return { currentRecord, answers, progress, saveAnswer }
})
```

### 2.5 自动导入

```ts
// vite.config.ts
plugins: [
  AutoImport({
    resolvers: [ElementPlusResolver()],
    dts: 'src/auto-imports.d.ts',
  }),
  Components({
    resolvers: [ElementPlusResolver()],
    dts: 'src/components.d.ts',
  }),
]

// 使用：无需 import { ElButton } from 'element-plus'  直接 <el-button>
```

---

## 3. 基础设施

### 3.1 数据库 / 缓存 / 队列

| 组件 | 用途 | 版本 | 选型理由 |
|------|------|------|----------|
| MySQL 8.0 | 主数据库 | 8.0.32 | 主流稳定、JSON / 窗口函数 |
| Redis 7 | 缓存 + 队列 + 限流 | 7.2 | 多数据结构、Stream 做轻量 MQ |

### 3.2 部署与编排

| 组件 | 用途 |
|------|------|
| Docker | 镜像构建 |
| Docker Compose | 本地 / 小规模 |
| Kubernetes | 生产 / 大规模 |
| Helm | K8s 包管理 |
| Nginx | 反向代理 + TLS |

### 3.3 可观测性

| 组件 | 用途 |
|------|------|
| Prometheus | 指标采集 |
| Grafana | 可视化 + 告警 |
| Loki | 日志聚合 |
| Promtail | 日志采集 |
| Jaeger | 链路追踪 |
| OpenTelemetry | 跨语言追踪标准 |

### 3.4 CI/CD

| 组件 | 用途 |
|------|------|
| GitHub Actions / GitLab CI | 流水线 |
| Docker Registry | 镜像仓库 |
| ArgoCD | K8s GitOps |

---

## 4. 技术决策记录（ADR 摘要）

### ADR-001: 选择 Hertz 而非 Gin

- **状态**：已采纳
- **决策**：使用 Hertz + Netpoll
- **理由**：Netpoll 高并发性能 + 字节大规模验证
- **权衡**：社区资源少于 Gin

### ADR-002: 选择 Vue 3 而非 React

- **状态**：已采纳
- **决策**：Vue 3 + Element Plus
- **理由**：国内生态丰富 + 团队熟悉
- **权衡**：国际化能力 React 略优

### ADR-003: 阅卷异步化（Redis Stream vs Kafka）

- **状态**：已采纳
- **决策**：v1.0 用 Redis Stream
- **理由**：减少中间件依赖、运维简单
- **后续**：日均任务 > 1kw 时切换 Kafka

### ADR-004: 数据库 MySQL vs PostgreSQL

- **状态**：已采纳
- **决策**：MySQL 8.0
- **理由**：团队熟悉 + 云厂商兼容好
- **权衡**：JSON / 全文检索 PG 更强（已用 JSON 字段够用）

---

## 5. 依赖管理

### 5.1 Go (go.mod)

```bash
# 添加依赖
go get github.com/xxx/yyy@latest

# 整理
go mod tidy

# 升级
go get -u github.com/xxx/yyy

# 漏洞扫描
govulncheck ./...
```

### 5.2 Node (package.json)

```bash
# 安装
pnpm install

# 添加
pnpm add xxx

# 升级
pnpm update

# 安全审计
pnpm audit
pnpm audit --fix
```

### 5.3 依赖锁定

- Go: `go.mod` + `go.sum` 必须提交
- Node: `pnpm-lock.yaml` 必须提交
- Docker: 基础镜像 tag 必须精确（不写 latest）

---

## 6. 技术债与待优化

### 短期（v1.1）
- [ ] Wire 依赖注入（替代手动构造）
- [ ] Hertz 启用 HTTP/3
- [ ] WebSocket 替代轮询

### 中期（v2.0）
- [ ] 微服务化（按 domain 拆分）
- [ ] LLM 微调（教育专用模型）

### 长期（v3.0）
- [ ] Service Mesh（Istio）
- [ ] WASM 边缘函数

---

> **相关文档**：[0002 架构](0002_ARCHITECTURE.md) · [0009 风险](0009_TECH_RISK.md) · [0010 性能](0010_PERFORMANCE.md)