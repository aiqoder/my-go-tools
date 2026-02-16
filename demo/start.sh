#!/bin/bash

# Demo 启动脚本

echo "启动 Vue + Gin 静态服务器 Demo..."

# 切换到 server 目录
cd "$(dirname "$0")/server"

# 启动服务
./server
