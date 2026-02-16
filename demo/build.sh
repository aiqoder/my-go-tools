#!/bin/bash

# Demo 构建脚本

echo "构建 Demo..."

# 构建前端
echo ">>> 构建前端..."
cd "$(dirname "$0")/frontend"
npm run build

# 构建后端
echo ">>> 构建后端..."
cd "$(dirname "$0")/server"
go build -o server .

echo ">>> 构建完成！"
echo "运行 ./start.sh 启动服务"
