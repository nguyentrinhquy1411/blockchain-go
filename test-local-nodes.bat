@echo off
echo =================================
echo   BLOCKCHAIN 3-NODE LOCAL TEST
echo =================================

REM Build project
echo Building project...
go build -o blockchain-node.exe cmd/node/main.go
if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

REM Clean old data
echo Cleaning old data...
if exist data rmdir /s /q data
mkdir data

echo.
echo Starting 3 local nodes for consensus test...
echo - Node1 (Leader): Port 50051
echo - Node2 (Follower): Port 50052  
echo - Node3 (Follower): Port 50053
echo.
echo Press Ctrl+C to stop all nodes
echo.

REM Start node1 (leader) on port 50051
start "Node1-Leader" cmd /c "set NODE_ID=node1 && set IS_LEADER=true && set PEERS=localhost:50052,localhost:50053 && set PORT=50051 && blockchain-node.exe"

REM Wait a bit
timeout /t 3 /nobreak >nul

REM Start node2 (follower) on port 50052
start "Node2-Follower" cmd /c "set NODE_ID=node2 && set IS_LEADER=false && set PEERS=localhost:50051 && set PORT=50052 && blockchain-node.exe"

REM Wait a bit
timeout /t 3 /nobreak >nul

REM Start node3 (follower) on port 50053
start "Node3-Follower" cmd /c "set NODE_ID=node3 && set IS_LEADER=false && set PEERS=localhost:50051 && set PORT=50053 && blockchain-node.exe"

echo.
echo All 3 nodes started in separate windows!
echo Check each terminal window for logs.
echo.
echo You can now:
echo 1. Watch consensus in action in the logs
echo 2. Test sending transactions via gRPC
echo 3. Close individual windows to test node recovery
echo.
echo Press any key to exit this script (nodes will keep running)...
pause
