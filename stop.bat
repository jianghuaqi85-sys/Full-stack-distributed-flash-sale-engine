@echo off
chcp 65001 >nul
title 分布式票务抢购系统 - 停止服务

echo ========================================
echo   分布式票务抢购系统 - 停止服务
echo ========================================
echo.

echo 正在停止服务...

:: 通过窗口标题停止本项目的进程（不会影响其他 Go/Node 项目）
taskkill /fi "WINDOWTITLE eq 票务后端*" /f >nul 2>nul
taskkill /fi "WINDOWTITLE eq 票务前端*" /f >nul 2>nul

:: 兜底：通过端口查找并杀掉占用 8080 和 3000 的进程
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080 " ^| findstr "LISTENING"') do (
    taskkill /f /pid %%a >nul 2>nul
)
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3000 " ^| findstr "LISTENING"') do (
    taskkill /f /pid %%a >nul 2>nul
)

echo 服务已停止！
echo.
pause
