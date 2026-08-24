#!/bin/bash
# 停止 KoalaExam 服务
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

[ -f .backend.pid ] && kill $(cat .backend.pid) 2>/dev/null && rm -f .backend.pid && echo "✓ 后端已停止"
[ -f .frontend.pid ] && kill $(cat .frontend.pid) 2>/dev/null && rm -f .frontend.pid && echo "✓ 前端已停止"

docker compose down
echo "✓ 基础设施已停止"
