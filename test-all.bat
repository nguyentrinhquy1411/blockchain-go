@echo off
echo ====================================================
echo Blockchain Full System Test
echo ====================================================

echo Building project...
call setup.bat

echo.
echo ====================================================
echo Test 1: Core Blockchain Features
echo ====================================================
blockchain.exe test

echo.
echo ====================================================
echo Test 2: Individual Component Tests
echo ====================================================

echo Testing Alice wallet creation...
blockchain.exe create-alice

echo Testing Bob wallet creation...
blockchain.exe create-bob

echo Testing Alice to Bob transaction...
blockchain.exe alice-to-bob 25.5

echo.
echo ====================================================
echo Test 3: 3-Node Consensus Test
echo ====================================================
echo Starting Docker containers for consensus test...
echo Press Ctrl+C to stop the consensus test when ready
timeout /t 3 > nul
docker-compose up

echo.
echo All tests completed!
pause
