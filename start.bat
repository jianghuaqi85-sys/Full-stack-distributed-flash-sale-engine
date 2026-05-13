@echo off
chcp 65001 >nul
title 分布式票务抢购系统 - 一键启动

echo ========================================
echo   分布式票务抢购系统 - 一键启动脚本
echo ========================================
echo.

:: 检查 Go 环境
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 Go，请先安装 Go
    pause
    exit /b 1
)

:: 检查 Node.js 环境
where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] 未找到 Node.js，请先安装 Node.js
    pause
    exit /b 1
)

:: 检查端口占用
netstat -ano | findstr ":8080 " | findstr "LISTENING" >nul 2>nul
if %errorlevel% equ 0 (
    echo [警告] 端口 8080 已被占用，可能已有后端服务在运行
    echo         如需重启，请先运行 stop.bat
    echo.
)

netstat -ano | findstr ":3000 " | findstr "LISTENING" >nul 2>nul
if %errorlevel% equ 0 (
    echo [警告] 端口 3000 已被占用，可能已有前端服务在运行
    echo         如需重启，请先运行 stop.bat
    echo.
)

cd /d "%~dp0"

echo [1/4] 检查依赖...
go mod tidy >nul 2>nul
if %errorlevel% neq 0 (
    echo [错误] go mod tidy 失败，请检查网络连接
    pause
    exit /b 1
)

echo [2/4] 启动后端服务...
start "票务后端" cmd /k "cd /d %~dp0 && go run ./cmd/api/main.go && pause"

:: 等待后端启动
echo 等待后端启动...
timeout /t 5 /nobreak >nul

:: 验证后端是否启动
curl -s http://localhost:8080/health >nul 2>nul
if %errorlevel% neq 0 (
    echo [警告] 后端可能未成功启动，请检查新窗口中的错误信息
) else (
    echo [OK] 后端启动成功
)

echo [3/4] 启动前端服务...
start "票务前端" cmd /k "cd /d %~dp0\web && npm run dev"

:: 等待前端启动
echo 等待前端启动...
timeout /t 3 /nobreak >nul

echo [4/4] 打开浏览器...
start http://localhost:3000

echo.
echo ========================================
echo   启动完成！
echo ========================================
echo.
echo   后端地址: http://localhost:8080
echo   前端地址: http://localhost:3000
echo.
echo   提示: 运行 stop.bat 可停止所有服务
echo.
echo   按任意键关闭此窗口...
echo ========================================
pause >nul
