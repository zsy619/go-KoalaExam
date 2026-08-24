#!/bin/bash
export PATH="/opt/homebrew/bin:/opt/homebrew/Cellar/node/26.3.0/bin:/usr/local/bin:/usr/bin:/bin"
cd "/Volumes/E/JYW/创意项目/go-KoalaExam/frontend"
exec /opt/homebrew/bin/node ./node_modules/vite/bin/vite.js > /tmp/vite.log 2>&1
