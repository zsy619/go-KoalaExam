#!/bin/bash
export HOME=/Users/zhushuyan
export USER=zhushuyan
export PATH="/usr/local/go/bin:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin"
export APP_ENV=dev
cd "/Volumes/E/JYW/创意项目/go-KoalaExam/backend"
exec /usr/local/go/bin/go run ./cmd/hertz > /tmp/koala-backend.log 2>&1
