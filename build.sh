#!/bin/bash
# VideoForge 编译脚本 (Linux/Mac)

echo "========================================"
echo "VideoForge 视频处理服务 - 编译脚本"
echo "========================================"
echo ""

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "[错误] 未找到 Go 环境，请先安装 Go 1.21+"
    exit 1
fi

echo "[1/3] 检查 Go 版本..."
go version

echo ""
echo "[2/3] 下载依赖..."
go mod download

echo ""
echo "[3/3] 编译程序..."
go build -ldflags="-s -w" -o videoforge

if [ $? -eq 0 ]; then
    echo ""
    echo "========================================"
    echo "✓ 编译成功！"
    echo "========================================"
    echo ""
    echo "可执行文件: videoforge"
    echo ""
    echo "运行方式:"
    echo "  ./videoforge"
    echo ""
    echo "然后访问: http://localhost:8080"
    echo "========================================"
    
    # 添加执行权限
    chmod +x videoforge
else
    echo ""
    echo "[错误] 编译失败"
    exit 1
fi
