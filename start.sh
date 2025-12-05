#!/bin/bash
# VideoForge 快速启动脚本 (Linux/Mac)

echo "========================================"
echo "VideoForge 视频处理服务"
echo "========================================"
echo ""

# 检查配置文件
if [ ! -f "config.json" ]; then
    echo "[错误] 配置文件 config.json 不存在"
    exit 1
fi

# 创建必要的目录
mkdir -p output
mkdir -p videos

if [ ! -d "web" ]; then
    echo "[错误] web 目录不存在，请确保前端文件完整"
    exit 1
fi

echo "[启动] 正在启动 VideoForge 服务..."
echo ""
echo "访问地址: http://localhost:8080"
echo "按 Ctrl+C 停止服务"
echo ""
echo "========================================"
echo ""

go run main.go
