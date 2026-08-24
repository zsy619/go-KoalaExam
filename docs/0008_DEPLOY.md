# 🚀 部署指南

> KoalaExam 的本地开发、Docker Compose、Kubernetes、生产环境部署完整流程。

---

## 1. 本地开发

### 1.1 一键启动

```bash
# 1. 启动 MySQL + Redis
docker compose up -d

# 2. 启动后端（热重载）
cd koala-exam-backend
make air

# 3. 启动前端
cd ../koala-exam-frontend
pnpm dev
```

### 1.2 服务地址

| 服务 | 地址 | 默认账号 |
|------|------|----------|
| 前端 | http://localhost:5173 | - |
| 后端 | http://localhost:8080 | - |
| Swagger | http://localhost:8080/swagger/index.html | - |
| MySQL | localhost:3306 | koala / koala123 |
| Redis | localhost:6379 | koala123 |

---

## 2. Docker Compose 全栈部署

### 2.1 docker-compose.yml

项目根目录的 `docker-compose.yml`：

```yaml
version: 3.8

services:
  mysql:
    image: mysql:8.0
    container_name: koala-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_DATABASE: koala_exam
      MYSQL_USER: koala
      MYSQL_PASSWORD: koala123
      TZ: Asia/Shanghai
    ports: ["3306:3306"]
    volumes:
      - ./docker-data/mysql:/var/lib/mysql
      - ./koala-exam-backend/migrations/init:/docker-entrypoint-initdb.d

  redis:
    image: redis:7-alpine
    container_name: koala-redis
    restart: unless-stopped
    ports: ["6379:6379"]
    volumes:
      - ./docker-data/redis:/data
    command: redis-server --appendonly yes --requirepass koala123

  # 后端（生产可选；开发用 air 热重载）
  backend:
    build: ./koala-exam-backend
    image: koala-exam-backend:latest
    container_name: koala-backend
    depends_on: [mysql, redis]
    environment:
      - GIN_MODE=release
    ports: ["8080:8080"]

  # 前端
  frontend:
    build: ./koala-exam-frontend
    image: koala-exam-frontend:latest
    container_name: koala-frontend
    ports: ["80:80"]
    depends_on: [backend]

networks:
  default:
    name: koala-net
```

### 2.2 启动

```bash
# 启动全部
docker compose up -d

# 仅启动基础设施
docker compose up -d mysql redis

# 查看日志
docker compose logs -f backend

# 停止
docker compose down

# 停止并清理数据
docker compose down -v
```

---

## 3. 后端 Dockerfile

`koala-exam-backend/Dockerfile`：

```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS builder
WORKDIR /app

# 缓存依赖
COPY go.mod go.sum ./
RUN go mod download

# 编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/koala-exam \
    ./cmd/hertz

# 运行时镜像
FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/bin/koala-exam /app/koala-exam
COPY --from=builder /app/configs /app/configs
EXPOSE 8080
CMD ["/app/koala-exam"]
```

---

## 4. 前端 Dockerfile + Nginx

`koala-exam-frontend/Dockerfile`：

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM nginx:1.25-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

`nginx.conf`：

```nginx
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location ~* \.(js|css|png|jpg|svg|woff2?)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    gzip on;
    gzip_types text/css application/javascript application/json;
}
```

---

## 5. Kubernetes 部署

### 5.1 后端 Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: koala-exam-backend
  namespace: koala
  labels: { app: koala-exam-backend }
spec:
  replicas: 3
  selector:
    matchLabels: { app: koala-exam-backend }
  template:
    metadata:
      labels: { app: koala-exam-backend }
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9100"
    spec:
      containers:
        - name: backend
          image: registry.example.com/koala-exam-backend:v1.0.0
          ports:
            - { containerPort: 8080 }
            - { containerPort: 9100, name: metrics }
          env:
            - name: APP_ENV
              value: "prod"
            - name: MYSQL_HOST
              value: "mysql.koala.svc.cluster.local"
            - name: REDIS_HOST
              value: "redis.koala.svc.cluster.local"
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef: { name: koala-secrets, key: jwt-secret }
          resources:
            requests: { cpu: 200m, memory: 256Mi }
            limits: { cpu: 1000m, memory: 1Gi }
          readinessProbe:
            httpGet: { path: /health, port: 8080 }
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet: { path: /health, port: 8080 }
            initialDelaySeconds: 30
            periodSeconds: 30
```

### 5.2 Service + Ingress

```yaml
apiVersion: v1
kind: Service
metadata:
  name: koala-exam-backend
  namespace: koala
spec:
  selector: { app: koala-exam-backend }
  ports:
    - { port: 8080, targetPort: 8080 }
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: koala-exam
  namespace: koala
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts: [api.koala.example.com]
      secretName: koala-tls
  rules:
    - host: api.koala.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend: { service: { name: koala-exam-backend, port: { number: 8080 } } }
```

### 5.3 HPA（自动伸缩）

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: koala-exam-backend
  namespace: koala
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: koala-exam-backend
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target: { type: Utilization, averageUtilization: 70 }
```

---

## 6. 生产环境配置

### 6.1 环境变量（强制）

```bash
APP_ENV=prod
MYSQL_HOST=mysql.example.com
MYSQL_PORT=3306
MYSQL_USER=koala_user
MYSQL_PASSWORD=<STRONG_PASSWORD>     # 从 Vault/KMS 注入
MYSQL_DB=koala_exam
REDIS_HOST=redis.example.com
REDIS_PASSWORD=<STRONG_PASSWORD>
JWT_SECRET=<32+ chars random>          # 从 KMS 注入
APP_LOG_LEVEL=info
APP_ENABLE_NETPOLL=true
```

### 6.2 安全配置 Checklist

- [ ] JWT_SECRET 使用 32+ 字符强随机串（从 KMS 注入）
- [ ] 数据库密码使用 Vault / KMS 管理
- [ ] 启用 HTTPS（cert-manager / Let's Encrypt）
- [ ] Redis 启用密码 + 仅监听内网
- [ ] 数据库账号按需最小权限
- [ ] 启用日志收集（ELK / Loki）
- [ ] WAF 规则启用（SQL/XSS / CC）
- [ ] HSTS / CSP / X-Frame-Options 头启用

详见 [0012_SECURITY.md](0012_SECURITY.md)。

### 6.3 性能配置

```yaml
mysql:
  max_open_conns: 200
  max_idle_conns: 50
  conn_max_lifetime: 3600
  slow_threshold: 200ms

redis:
  pool_size: 100
  min_idle_conns: 10

log:
  level: info
  compress: true

ratelimit:
  enabled: true
  qps: 200
  burst: 400

app:
  mode: release
  enable_netpoll: true
```

---

## 7. CI/CD（GitHub Actions）

`.github/workflows/deploy.yml`：

```yaml
name: deploy
on:
  push: { branches: [main] }
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: 1.21 }
      - name: Test backend
        run: cd koala-exam-backend && go test ./...
      - name: Lint
        run: cd koala-exam-backend && golangci-lint run

  build-and-push:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Login to Registry
        uses: docker/login-action@v3
        with:
          registry: registry.example.com
          username: ${{ secrets.REGISTRY_USER }}
          password: ${{ secrets.REGISTRY_PASS }}
      - name: Build & Push Backend
        run: |
          docker build -t registry.example.com/koala-exam-backend:${{ github.sha }} ./koala-exam-backend
          docker push registry.example.com/koala-exam-backend:${{ github.sha }}
      - name: Build & Push Frontend
        run: |
          docker build -t registry.example.com/koala-exam-frontend:${{ github.sha }} ./koala-exam-frontend
          docker push registry.example.com/koala-exam-frontend:${{ github.sha }}

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to K8s
        uses: azure/setup-kubectl@v4
      - name: Set backend image
        run: kubectl set image deployment/koala-exam-backend -n koala backend=registry.example.com/koala-exam-backend:${{ github.sha }}
      - name: Set frontend image
        run: kubectl set image deployment/koala-exam-frontend -n koala frontend=registry.example.com/koala-exam-frontend:${{ github.sha }}
      - name: Wait for rollout
        run: kubectl rollout status deployment/koala-exam-backend -n koala
```

---

## 8. 数据库迁移（生产）

```bash
# 1. 备份
mysqldump -h mysql.prod -u backup -p koala_exam > backup_$(date +%F).sql

# 2. 应用迁移
kubectl exec -it deploy/koala-exam-backend -n koala -- /app/koala-exam migrate up

# 3. 验证
kubectl exec -it deploy/koala-exam-backend -n koala -- /app/koala-exam migrate status
```

---

## 9. 监控与告警

### 9.1 监控接入

- **APM**：Prometheus + Grafana（采集 Hertz metrics）
- **日志**：ELK / Loki（Zap 结构化日志）
- **告警**：AlertManager（错误率 > 1%, P99 > 500ms）
- **追踪**：Jaeger / OpenTelemetry（Hertz tracer）

### 9.2 关键告警规则

```yaml
# alertmanager/rules/koala.yaml
groups:
  - name: koala_alerts
    rules:
      - alert: HighErrorRate
        expr: sum(rate(koala_http_requests_total{status=~"5.."}[5m])) / sum(rate(koala_http_requests_total[5m])) > 0.01
        for: 5m
        annotations: { summary: "高错误率 > 1%" }
      - alert: HighLatency
        expr: histogram_quantile(0.99, rate(koala_http_request_duration_seconds_bucket[5m])) > 0.5
        for: 3m
        annotations: { summary: "P99 延迟 > 500ms" }
```

详见 [0010_PERFORMANCE.md](0010_PERFORMANCE.md)。

---

## 10. 回滚流程

```bash
# 1. 立即回滚（保留数据）
kubectl rollout undo deployment/koala-exam-backend -n koala
kubectl rollout undo deployment/koala-exam-frontend -n koala

# 2. 数据库回滚（需提前备份）
mysql -h mysql.prod -u root -p koala_exam < backup_2024-12-15.sql

# 3. Redis 清理（如有脏数据）
redis-cli -h redis.prod FLUSHDB   # 仅限缓存可清空

# 4. 通知 + 记录（事后写 RCA）
```

---

> 相关文档：[0007 开发](0007_DEV_GUIDE.md) · [0010 性能](0010_PERFORMANCE.md) · [0012 安全](0012_SECURITY.md)