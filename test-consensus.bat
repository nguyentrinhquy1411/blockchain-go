@echo off
echo =================================
echo   BLOCKCHAIN 3-NODE DOCKER TEST
echo =================================

REM Check if ports are available
echo Checking for port conflicts...
netstat -an | findstr ":50051 " > nul
if %errorlevel% equ 0 (
    echo WARNING: Port 50051 is in use. Attempting cleanup...
    for /f "tokens=5" %%i in ('netstat -ano ^| findstr ":50051 "') do (
        echo Killing process %%i on port 50051...
        taskkill /PID %%i /F 2>nul
    )
)

netstat -an | findstr ":50052 " > nul
if %errorlevel% equ 0 (
    echo WARNING: Port 50052 is in use. Attempting cleanup...
    for /f "tokens=5" %%i in ('netstat -ano ^| findstr ":50052 "') do (
        echo Killing process %%i on port 50052...
        taskkill /PID %%i /F 2>nul
    )
)

netstat -an | findstr ":50053 " > nul
if %errorlevel% equ 0 (
    echo WARNING: Port 50053 is in use. Attempting cleanup...
    for /f "tokens=5" %%i in ('netstat -ano ^| findstr ":50053 "') do (
        echo Killing process %%i on port 50053...
        taskkill /PID %%i /F 2>nul
    )
)

REM Clean up any existing containers
echo Stopping any existing containers...
docker-compose down --remove-orphans -v 2>nul

echo Waiting for cleanup to complete...
timeout /t 3 /nobreak > nul

REM Build and start 3 nodes with Docker Compose
echo Building and starting 3 Docker nodes...
echo - Node1 (Leader): localhost:50051 - Web: http://localhost:8080
echo - Node2 (Follower): localhost:50052 - Web: http://localhost:8081
echo - Node3 (Follower): localhost:50053 - Web: http://localhost:8082
echo.

docker-compose up --build -d

if %errorlevel% neq 0 (
    echo Docker compose failed to start!
    echo Make sure Docker Desktop is running and ports are available
    echo You can run 'docker-cleanup.bat' to clean up resources
    pause
    exit /b 1
)

echo.
echo =================================
echo   3 NODES STARTED SUCCESSFULLY
echo =================================
echo.
echo Checking node status...
timeout /t 5 /nobreak >nul

echo.
echo === NODE 1 (LEADER) LOGS ===
docker-compose logs --tail=20 node1

echo.
echo === NODE 2 (FOLLOWER) LOGS ===
docker-compose logs --tail=20 node2

echo.
echo === NODE 3 (FOLLOWER) LOGS ===
docker-compose logs --tail=20 node3

echo.
echo =================================
echo   NODES ARE RUNNING
echo =================================
echo You can:
echo 1. Check logs: docker-compose logs -f
echo 2. Stop nodes: docker-compose down
echo 3. Test consensus by watching the logs
echo.
echo Press any key to view live logs (Ctrl+C to stop)...
pause >nul

echo.
echo Showing live logs from all nodes...
docker-compose logs -f
