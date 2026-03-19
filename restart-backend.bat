@echo off
setlocal

echo [1/3] 正在停止旧的 3001 端口进程...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3001" ^| findstr "LISTENING"') do (
  taskkill /PID %%a /F >nul 2>nul
)

echo [2/3] 切换到后端目录...
cd /d "C:\Users\fengx\0\demo\backend"

echo [3/3] 启动后端（最新代码）...
echo ----------------------------------------
npm run dev

endlocal
