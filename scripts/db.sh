#!/bin/bash
# 数据库管理脚本
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

CMD=${1:-help}

case $CMD in
  fresh)
    echo "🔥 重置数据库（删表 + 建表 + 种子）..."
    cd koala-exam-backend && go run ./cmd/migrate -op fresh && cd ..
    ;;
  reset)
    echo "🔄 重置表结构（删表 + 建表）..."
    cd koala-exam-backend && go run ./cmd/migrate -op reset && cd ..
    ;;
  seed)
    echo "🌱 写入测试数据..."
    cd koala-exam-backend && go run ./cmd/migrate -op seed && cd ..
    ;;
  init)
    echo "📜 直接执行 SQL 初始化（不需要 Go）..."
    CID=$(docker compose ps -q mysql)
    docker exec -i $CID mysql -uroot -p123456 -e "CREATE DATABASE IF NOT EXISTS KoalaExam DEFAULT CHARSET utf8mb4"
    docker exec -i $CID mysql -uroot -p123456 KoalaExam < koala-exam-backend/migrations/init/01_schema.sql
    docker exec -i $CID mysql -uroot -p123456 KoalaExam < koala-exam-backend/migrations/init/02_seed.sql
    echo "✓ 初始化完成"
    ;;
  shell)
    CID=$(docker compose ps -q mysql)
    docker exec -it $CID mysql -uroot -p123456 KoalaExam
    ;;
  help|*)
    echo "用法: ./scripts/db.sh <命令>"
    echo ""
    echo "命令:"
    echo "  fresh    重置数据库（含测试数据）"
    echo "  reset    重置表结构（不含数据）"
    echo "  seed     仅写入测试数据"
    echo "  init     直接执行 SQL 初始化（不需要 Go）"
    echo "  shell    进入 MySQL Shell"
    ;;
esac
