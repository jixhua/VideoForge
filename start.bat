@echo off
REM VideoForge 快速启动脚本 (Windows)

echo ========================================
echo VideoForge 视频处理服务
echo ========================================
echo.

REM 检查配置文件
if not exist "config.json" (
    echo [错误] 配置文件 config.json 不存在
    pause
    exit /b 1
)

REM 创建必要的目录
if not exist "output" mkdir output
if not exist "videos" mkdir videos
if not exist "web" (
    echo [错误] web 目录不存在，请确保前端文件完整
    pause
    exit /b 1
)

echo [启动] 正在启动 VideoForge 服务...
echo.
echo 访问地址: http://localhost:8080
echo 按 Ctrl+C 停止服务
echo.
echo ========================================
echo.

go run main.go
