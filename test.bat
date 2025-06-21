@echo off
echo =================================
echo    BLOCKCHAIN TEST SCRIPT
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

REM Test CLI demo first
echo.
echo =================================
echo    TESTING CLI DEMO (Alice-Bob)
echo =================================
go run cmd/main.go
if %errorlevel% neq 0 (
    echo CLI demo failed!
    pause
    exit /b 1
)

echo.
echo =================================
echo    TESTING SINGLE NODE
echo =================================
echo Starting single node (Press Ctrl+C to stop)...
set NODE_ID=node1
set IS_LEADER=true
set PEERS=
blockchain-node.exe

pause
