#!/bin/bash
# KoalaExam 一键启动脚本
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "=========================================="
echo "  🐨 KoalaExam 一键启动"
echo "=========================================="
echo ""

# 1. 检查依赖
command -v docker >/dev/null 2>&1 || { echo "❌ 需要 Docker"; exit 1; }
command -v go >/dev/null 2>&1 || { echo "❌ 需要 Go 1.21+"; exit 1; }
command -v node >/dev/null 2>&1 || { echo "❌ 需要 Node.js 20+"; exit 1; }

# 2. 启动基础设施
echo "[1/5] 启动 MySQL + Redis ..."
docker compose up -d mysql redis
echo "等待数据库就绪 ..."
sleep 8

# 3. 初始化后端
echo "[2/5] 安装后端依赖 ..."
cd koala-exam-backend
go mod tidy || true
echo "[3/5] 创建数据库 + 建表 + 种子数据 ..."
go run ./cmd/migrate -op fresh || {
  echo "⚠  自动迁移失败，尝试手动执行 SQL ..."
  CID=$(docker compose ps -q mysql)
  docker exec -i $CID mysql -uroot -p123456 KoalaExam < migrations/init/01_schema.sql
  docker exec -i $CID mysql -uroot -p123456 KoalaExam < migrations/init/02_seed.sql
}

# 4. 启动后端（后台）
mkdir -p ../logs
echo "[4/5] 启动后端服务（端口 8080）..."
nohup go run ./cmd/hertz > ../logs/backend.log 2>&1 &
echo $! > ../.backend.pid
cd ..

# 5. 启动前端
echo "[5/5] 安装前端依赖 + 启动（端口 5173）..."
cd koala-exam-frontend
[ -d node_modules ] || (pnpm install 2>/dev/null || npm install)
nohup pnpm dev > ../logs/frontend.log 2>&1 &
echo $! > ../.frontend.pid
cd ..

sleep 5
echo ""
echo "=========================================="
echo "  ✅ 启动完成"
echo "=========================================="
echo ""
echo "📍 前端地址:    http://localhost:5173"
echo "📍 后端 API:    http://localhost:8080/api/v1"
echo "📍 健康检查:    http://localhost:8080/health"
echo "📍 MySQL:       localhost:3306 (root / 123456)"
echo "📍 Redis:       localhost:6379"
echo ""
echo "🔑 默认账号："
echo "   超管: admin / koala123"
echo "   教师: teacher / koala123"
echo "   学员: student / koala123"
echo ""
echo "📋 查看日志: tail -f logs/backend.log logs/frontend.log"
echo "🛑 停止服务: ./scripts/stop.sh"
