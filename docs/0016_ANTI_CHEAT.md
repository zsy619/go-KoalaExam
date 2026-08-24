# 🛡 防作弊专题（Anti-Cheat）

> 围绕考试公平性的全链路防作弊体系：前端拦截、后端审计、AI 监考（v2.0）。
> 本文聚焦 **检测、记录、告警、处置** 四个维度，并提供完整的代码示例与配置项。

---

## 1. 设计原则

防作弊体系的 5 条原则：

1. **检测 > 阻止**：无法 100% 阻止作弊（任何客户端方案都可绕过），但能 100% 检测到
2. **不影响正常作答**：检测逻辑必须低开销、不阻塞渲染
3. **可审计**：所有可疑事件必须留痕（学员 / 教师 / CSO 三方可查）
4. **可配置**：不同考试场景（家庭作业 / 期末考试）应有不同严格度
5. **隐私合规**：视频流、人脸数据需明确告知 + 加密存储

---

## 2. 防作弊措施全景

| 类别 | 措施 | 实现层 | 绕过难度 | 成本 |
|------|------|--------|----------|------|
| **身份验证** | JWT + 设备指纹 | 后端 | 高 | 低 |
| **全屏监控** | requestFullscreen + 失焦告警 | 前端 | 中 | 低 |
| **切屏检测** | visibilitychange | 前端 | 中 | 低 |
| **复制粘贴** | copy/paste/contextmenu 拦截 | 前端 | 高 | 低 |
| **题目乱序** | rand.Shuffle | 后端 | - | 低 |
| **选项乱序** | rand.Shuffle | 后端 | - | 低 |
| **审计日志** | visibilitychange → Redis → MySQL | 前后端 | - | 中 |
| **IP 绑定** | 异常 IP 告警 | 后端 | 低 | 中 |
| **答题时长** | Elapsed 字段分析 | 后端 | 中 | 低 |
| **人脸识别** | 活体检测 + 抓拍 | 前端 + AI | 极高 | 高 |
| **屏幕录制** | getDisplayMedia | 前端 | 中 | 中 |
| **AI 异常行为** | 鼠标轨迹 / 答题模式 | AI | 极高 | 高 |
| **服务端答案校验** | 学员端永远拿不到 answer | 后端 | - | - |
| **多端互踢** | DeviceID 黑名单 | 后端 | 高 | 中 |

---

## 3. 前端拦截措施详解

### 3.1 全屏监控

```ts
// composables/useFullscreen.ts
export function useFullscreen(enabled: boolean, onExit: () => void) {
  function enter() {
    if (!enabled) return
    if (document.fullscreenElement) return
    document.documentElement.requestFullscreen().catch(err => {
      console.warn('Fullscreen denied:', err)
    })
  }

  function onChange() {
    if (!enabled) return
    if (!document.fullscreenElement) {
      onExit()  // 触发告警
    }
  }

  onMounted(() => {
    enter()
    document.addEventListener('fullscreenchange', onChange)
  })

  onBeforeUnmount(() => {
    document.removeEventListener('fullscreenchange', onChange)
  })

  return { enter }
}
```

**绕过手段**：用户可在浏览器设置关闭全屏、DevTools 修改 DOM / 调用 exitFullscreen。

**应对**：每次退出都记录审计日志 + 切屏计数 +1。

### 3.2 切屏检测

```ts
// composables/useAntiCheat.ts
export function useAntiCheat(options: {
  enabled: boolean
  maxTabSwitches: number       // 默认 3
  onViolation: (type: string, payload?: any) => void
  onForceSubmit: () => void
}) {
  const tabSwitchCount = ref(0)

  function onVisibilityChange() {
    if (!options.enabled || !document.hidden) return

    tabSwitchCount.value++
    options.onViolation('tab_switch', {
      from: 'visible', to: 'hidden',
      count: tabSwitchCount.value,
    })

    // 达到上限自动收卷
    if (tabSwitchCount.value >= options.maxTabSwitches) {
      options.onForceSubmit()
    }
  }

  function preventCopy(e: ClipboardEvent) {
    if (!options.enabled) return
    e.preventDefault()
    options.onViolation('copy_attempt', { event: e.type })
  }

  function preventContextMenu(e: MouseEvent) {
    if (!options.enabled) return
    e.preventDefault()
    options.onViolation('context_menu')
  }

  onMounted(() => {
    document.addEventListener('visibilitychange', onVisibilityChange)
    document.addEventListener('copy', preventCopy)
    document.addEventListener('paste', preventCopy)
    document.addEventListener('cut', preventCopy)
    document.addEventListener('contextmenu', preventContextMenu)
  })

  onBeforeUnmount(() => {
    document.removeEventListener('visibilitychange', onVisibilityChange)
    document.removeEventListener('copy', preventCopy)
    // ... 其他清理
  })

  return { tabSwitchCount }
}
```

**绕过手段**：禁用 JavaScript / 浏览器扩展拦截事件 / DevTools 修改 listener。

**应对**：检测不到事件 → 心跳检测 + 服务端时长异常检测（学员每 10s 上报 elapsed，若 30s 没动作 → 可疑）。

### 3.3 复制粘贴拦截

| 事件 | 默认行为 | 防作弊拦截 |
|------|----------|------------|
| Ctrl+C / copy | 复制选中 | e.preventDefault() + 审计 |
| Ctrl+V / paste | 粘贴 | e.preventDefault() + 审计 |
| 右键 | 弹出菜单 | e.preventDefault() + 审计 |
| 拖拽 | 拖动选中 | dragstart 拦截 |

**注意**：不能完全阻止（用户可在 URL bar 手输、截图、OCR）。但能阻止 90% 的快捷键作弊。

### 3.4 屏幕键盘检测

```ts
// 检测开发者工具开启（启发式）
function isDevToolsOpen(): boolean {
  const threshold = 160
  return (
    window.outerWidth - window.innerWidth > threshold ||
    window.outerHeight - window.innerHeight > threshold
  )
}

setInterval(() => {
  if (isDevToolsOpen()) {
    onViolation('devtools_open')
  }
}, 5000)
```

**绕过**：用户可调整窗口大小 / 用独立窗口的 DevTools。

**应对**：仅作为辅助信号，不作主要判定。

---

## 4. 后端审计与判定

### 4.1 审计事件接口

```go
// application/dto/exam_dto.go
type AuditReq struct {
    RecordID int64                  `json:"record_id" binding:"required"`
    Events   map[string]interface{} `json:"events" binding:"required"`
}

// 调用方（前端）：
// POST /api/v1/exams/audit
// { record_id, events: { type: "tab_switch", payload: BLANK } }
```

### 4.2 后端审计处理

```go
// application/exam/exam_app.go
func (a *ExamApp) AuditEvent(ctx context.Context, req *dto.AuditReq) error {
    rec, err := a.recordRepo.GetByID(ctx, req.RecordID)
    if err != nil { return err }

    // 1. 累加切屏次数（高频事件单独计数）
    if ev, ok := req.Events["type"]; ok {
        switch ev {
        case "tab_switch":
            rec.TabSwitchCnt++
            // 超限标记
            if rec.TabSwitchCnt >= 3 {
                rec.Status = consts.RecordStatusSuspicious  // 异常态
            }
        case "copy_attempt":
            a.stats.RecordCheatingAttempt(rec.UserID, "copy")
        case "devtools_open":
            a.stats.RecordCheatingAttempt(rec.UserID, "devtools")
        }
    }

    // 2. 追加审计日志
    var audit []map[string]interface{}
    if rec.AuditLog != "" {
        _ = json.Unmarshal([]byte(rec.AuditLog), &audit)
    }
    audit = append(audit, map[string]interface{}{
        "type":      req.Events["type"],
        "payload":   req.Events["payload"],
        "timestamp": time.Now().Unix(),
        "ip":        ctx.Value("client_ip"),
    })

    // ⚠️ 性能问题：每次 UPDATE 整个 record，写放大
    b, _ := json.Marshal(audit)
    rec.AuditLog = string(b)
    return a.recordRepo.Update(ctx, rec)
}
```

### 4.3 v1.1 改进：审计拆表

```sql
-- 新表：ke_exam_audit_log
CREATE TABLE ke_exam_audit_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    record_id BIGINT NOT NULL,
    event_type VARCHAR(32) NOT NULL,
    payload JSON,
    ip VARCHAR(45),
    user_agent VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    KEY idx_record (record_id, created_at)
) ENGINE=InnoDB;

-- 写入改为批量 INSERT
INSERT INTO ke_exam_audit_log (record_id, event_type, payload) VALUES
  (?, 'tab_switch', ?),
  (?, 'copy_attempt', ?),
  ...;

-- 主表只保留汇总
ALTER TABLE ke_exam_record DROP COLUMN audit_log;
ALTER TABLE ke_exam_record ADD COLUMN audit_summary JSON;  -- { tab_switch: 3, copy: 1, ... }
```

**收益**：审计写入吞吐提升 100x，主表记录保持精简。

---

## 5. 防作弊判定规则

### 5.1 可疑模式检测（v2.0 计划）

| 模式 | 触发条件 | 严重度 | 处置 |
|------|----------|--------|------|
| 秒答全对 | 总用时 < duration * 10% 且正确率 > 95% | 高 | 标记可疑 + 人工复审 |
| 答题速度异常 | 平均答题时长 < 5 秒 | 中 | 标记待观察 |
| 切屏集中爆发 | 30s 内切屏 > 5 次 | 高 | 自动收卷 + 教师告警 |
| 多 IP 切换 | 考试中 IP 变化 > 3 次 | 极高 | 强制收卷 + 报告 |
| 同一答案 | 同一 exam 内多人答案完全一致 | 中 | 关联分析（可能团伙） |
| 答案匹配已知题库 | 答案与公开题库匹配 | 中 | 加分可疑度 |

### 5.2 实现示例：异常答题时长检测

```go
// application/exam/cheat_detector.go
func (a *ExamApp) DetectAnomalies(rec *entity.ExamRecord) []Anomaly {
    anomalies := []Anomaly{}

    // 1. 秒答全对
    if rec.TotalScore >= rec.PassScore * 0.95 {
        elapsedMin := float64(rec.Duration) / 60
        if elapsedMin < float64(rec.Exam.Duration) * 0.1 {
            anomalies = append(anomalies, Anomaly{
                Type: "instant_correct",
                Severity: "high",
                Detail: fmt.Sprintf("用时 %.1f 分钟达到 %.1f 分", elapsedMin, rec.TotalScore),
            })
        }
    }

    // 2. 切屏集中爆发（基于 audit_log）
    // ...（略）

    return anomalies
}
```

---

## 6. 教师监考大屏（v1.2 计划）

### 6.1 功能

```
┌─────────────────────────────────────────────────────────┐
│  实时监考（XX 场考试进行中）                              │
├─────────────────────────────────────────────────────────┤
│  [场次A] 60 学员在线 / 30 已交卷 / 2 异常              │
│    ├─ 张三：进行中 / 切屏 2 次 / 答题进度 60%          │
│    ├─ 李四：进行中 / 切屏 0 次 / 答题进度 80%          │
│    ├─ 王五：⚠️ 异常 / 切屏 5 次（已自动收卷）          │
│    └─ ...                                                │
│                                                         │
│  [场次B] 100 学员在线 / 90 已交卷 / 0 异常              │
│    └─ ...                                                │
│                                                         │
│  [操作] [批量强制收卷] [延长考试时间] [导出异常报告]    │
└─────────────────────────────────────────────────────────┘
```

### 6.2 后端 SSE 推送

```go
// interfaces/handler/admin_monitor.go
func (h *MonitorHandler) Stream(c *app.RequestContext) {
    // 设置响应头
    c.SetHeader("Content-Type", "text/event-stream")
    c.SetHeader("Cache-Control", "no-cache")
    c.SetHeader("Connection", "keep-alive")

    // 订阅 Redis Pub/Sub 频道
    pubsub := h.rdb.Subscribe(c, "koala:monitor")
    defer pubsub.Close()

    for {
        select {
        case msg := <-pubsub.Channel():
            fmt.Fprintf(c.Writer, "data: %s\n\n", msg.Payload)
            c.Writer.Flush()
        case <-c.Done():
            return
        }
    }
}

// 推送：
// data: {"event":"tab_switch","user_id":123,"count":3}
```

### 6.3 前端 EventSource 订阅

```ts
const es = new EventSource('/api/v1/admin/monitor/stream', {
  withCredentials: true,
})

es.onmessage = (e) => {
  const data = JSON.parse(e.data)
  if (data.event === 'tab_switch') {
    ElNotification.warning(`学员 ${data.user_id} 切屏 ${data.count} 次`)
  }
}
```

---

## 7. AI 智能监考（v2.0 计划）

### 7.1 人脸识别

```ts
// composables/useFaceCheck.ts
async function startFaceCheck(stream: MediaStream) {
  // 每 30 秒抓拍一次
  setInterval(async () => {
    const img = captureFrame(stream)
    const result = await api.verifyFace({ image: img.toDataURL() })
    if (!result.matched) {
      onViolation('face_mismatch', { score: result.score })
    }
    if (result.liveness < 0.8) {
      onViolation('no_liveness')
    }
  }, 30_000)
}
```

### 7.2 屏幕录制

```ts
const stream = await navigator.mediaDevices.getDisplayMedia({ video: true })
const recorder = new MediaRecorder(stream, { mimeType: 'video/webm' })
recorder.ondataavailable = e => uploadChunk(e.data)  // 分片上传 OSS
recorder.start(10_000)  // 每 10s 一个 chunk
```

### 7.3 鼠标轨迹分析（v2.0）

```ts
// 采样鼠标移动
const trail = []
document.addEventListener('mousemove', e => {
  trail.push({ x: e.clientX, y: e.clientY, t: Date.now() })
})

// 提交时上传 → AI 分析
// 异常模式：直线运动（机器人）、长时间静止（无人）
```

---

## 8. 安全措施配套

### 8.1 防服务端答案泄露

| 层 | 措施 |
|----|------|
| 学员端 ToResp | showAnswer=false |
| 网络传输 | TLS 1.3 加密 |
| API 响应 | 仅返回题干 + 选项（不含 answer / analysis） |
| 教师端 | 阅卷权限 + 审计 |
| 数据库 | correct_answer 字段 AES-256-GCM 加密（v1.1 计划） |

### 8.2 防重放攻击

- 接口签名（HMAC-SHA256，详见 0012）
- 防重放 nonce
- 时间戳（5 分钟内有效）

### 8.3 防 IP 异常

```go
// middleware/ip_check.go
func IPAnomalyCheck() app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        userID := ctx.Value("user_id").(int64)
        ip := c.ClientIP()

        // 1. 缓存历史 IP
        key := fmt.Sprintf("koala:user:ips:%d", userID)
        ips, _ := rdb.SAdd(ctx, key, ip).Result()
        if ips > 3 {
            // 2. 多 IP 告警
            metrics.RecordIPChange(userID, ips)
            if ips > 5 {
                logger.Warn("user multi-IP detected",
                    zap.Int64("user_id", userID),
                    zap.String("new_ip", ip))
                // 可选：强制收卷
            }
        }
        // 30 天后清理
        rdb.Expire(ctx, key, 30 * 24 * time.Hour)
    }
}
```

---

## 9. 配置项（ke_exam 表新增字段）

```sql
-- v1.1 计划迁移
ALTER TABLE ke_exam ADD COLUMN
    max_tab_switch_count INT DEFAULT 3,
    auto_submit_on_violation BOOLEAN DEFAULT false,
    enable_fullscreen BOOLEAN DEFAULT true,
    enable_copy_paste_block BOOLEAN DEFAULT true,
    enable_face_check BOOLEAN DEFAULT false,         -- v2.0
    enable_screen_record BOOLEAN DEFAULT false;       -- v2.0
```

---

## 10. 防作弊与隐私合规

### 10.1 知情同意

```html
<!-- 学员端进入考试前显示 -->
<div class="consent-dialog">
  <h3>考试环境检测</h3>
  <p>为维护考试公平，本场考试将启用以下检测：</p>
  <ul>
    <li>✅ 全屏监控（防止切换窗口）</li>
    <li>✅ 人脸识别活体检测（仅本场考试使用）</li>
    <li>✅ 屏幕录制（考试结束后 30 天删除）</li>
  </ul>
  <p>所有数据加密存储，仅超管 / 教师可查阅。</p>
  <el-button @click="startExam">同意并开始</el-button>
  <el-button @click="cancel">退出</el-button>
</div>
```

### 10.2 数据保留策略

```
- 审计日志（tab_switch / copy_attempt）：保留 180 天
- 屏幕录制：考试后 30 天删除（除非涉及作弊申诉）
- 人脸抓拍：仅本场考试使用，不持久化
- 鼠标轨迹：实时分析，不持久化
```

### 10.3 合规要点

- 个人信息保护法：检测前必须明确告知
- GDPR（海外）：用户可请求导出 / 删除自己的审计数据
- 等保 2.0：审计日志 180 天保留 + 防篡改
- 学校规定：未成年学员需要家长同意（K12 场景）

---

## 11. 防作弊效果评估

### 11.1 指标

```
作弊率 = 确认作弊人数 / 总考试人数
误判率 = 误判作弊人数 / 总告警人数
发现率 = 发现作弊人数 / 实际作弊人数（需抽样调研）

目标：
  作弊率 < 1%
  误判率 < 5%
  发现率 > 80%
```

### 11.2 攻防对抗

| 攻击手段 | 防御 | 评估 |
|----------|------|------|
| 另一台设备搜索答案 | 答题时长 + 切屏检测 | 80% 检出 |
| 双人协同答题 | 答题模式分析 + 教师人工 | 70% 检出 |
| 拍照搜题 | 禁止复制 + 切屏 | 60% 检出（需 AI 辅助） |
| 代考 | 人脸识别 | 95% 检出（v2.0） |
| 题库泄露 | 试卷快照 + 题目乱序 | 95% 防御 |

---

## 12. 已知技术债

### v1.0 待改进

- [ ] AuditEvent 拆表（写放大严重）
- [ ] audit_log 体积过大（单条可达 MB）
- [ ] 全局 rand → crypto/rand（确定性随机）
- [ ] 防作弊规则引擎化（当前硬编码）

### v1.1 计划

- [ ] AI 异常答题模式检测
- [ ] 教师监考大屏（SSE 推送）
- [ ] 多端互踢（DeviceID）
- [ ] IP 异常自动处置

### v2.0 计划

- [ ] 人脸识别 + 活体检测
- [ ] 屏幕录制（OSS 分片上传）
- [ ] 鼠标轨迹 AI 分析
- [ ] 异常行为实时拦截

---

## 13. 相关文档

- [0015_EXAM_LIFECYCLE.md](0015_EXAM_LIFECYCLE.md) - 考试状态机与流转
- [0017_GRADING.md](0017_GRADING.md) - 阅卷流水线（错题自动收录）
- [0012_SECURITY.md](0012_SECURITY.md) - 安全合规
- [0009_TECH_RISK.md](0009_TECH_RISK.md) - 作弊风险评估
- [0011_RUNBOOK.md](0011_RUNBOOK.md) - 监考异常应急流程

---

> 维护者：考试组 + 安全组 / 月度评审
