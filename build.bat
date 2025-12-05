@echo off
REM VideoForge 编译脚本 (Windows)

echo ========================================
echo VideoForge 视频处理服务 - 编译脚本
echo ========================================
echo.

REM 检查 Go 环境
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [错误] 未找到 Go 环境，请先安装 Go 1.21+
    exit /b 1
)

echo [1/3] 检查 Go 版本...
go version

echo.
echo [2/3] 下载依赖...
go mod download

echo.
echo [3/3] 编译程序...
go build -ldflags="-s -w" -o videoforge.exe

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo ? 编译成功！
    echo ========================================
    echo.
    echo 可执行文件: videoforge.exe
    echo.
    echo 运行方式:
    echo   videoforge.exe
    echo.
    echo 然后访问: http://localhost:8080
    echo ========================================
) else (
    echo.
    echo [错误] 编译失败
    exit /b 1
)

pause
