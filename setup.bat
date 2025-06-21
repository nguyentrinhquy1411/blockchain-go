@echo off
echo ====================================================
echo Setup: Blockchain Project
echo ====================================================

echo Installing dependencies...
go mod tidy

echo Building CLI...
go build -o blockchain.exe .\cmd\main.go

echo Building node server...
go build -o blockchain-node.exe .\cmd\node

echo Setup completed!
echo.
echo Available commands:
echo   blockchain.exe demo          - Run Alice and Bob demo
echo   blockchain.exe test          - Run full system test
echo   blockchain.exe help          - Show all commands
echo.
echo Docker commands:
echo   docker-compose up -d         - Start 3-node consensus
echo   .\test-consensus.bat         - Test consensus mechanism
echo.
pause
