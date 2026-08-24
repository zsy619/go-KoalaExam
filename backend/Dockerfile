# KoalaExam 后端 Dockerfile（多阶段构建）
# Stage 1: 构建
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o koala-exam ./cmd/hertz

# Stage 2: 运行
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app

COPY --from=builder /build/koala-exam /app/koala-exam
COPY --from=builder /build/configs /app/configs
COPY --from=builder /build/migrations /app/migrations

RUN mkdir -p /app/logs /app/uploads
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s CMD wget --spider http://localhost:8080/health || exit 1
ENTRYPOINT ["/app/koala-exam"]
