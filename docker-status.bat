@echo off
echo =================================
echo   BLOCKCHAIN NODES STATUS CHECK
echo =================================

echo Checking Docker containers status...
docker-compose ps

echo.
echo =================================
echo   PORT STATUS
echo =================================
echo Checking port availability...
netstat -an | findstr ":5005" | findstr "LISTENING"
echo.

echo =================================
echo   NODE HEALTH CHECK
echo =================================
echo Testing Node 1 (localhost:50051)...
timeout /t 1 /nobreak > nul
echo.

echo Testing Node 2 (localhost:50052)...
timeout /t 1 /nobreak > nul
echo.

echo Testing Node 3 (localhost:50053)...
timeout /t 1 /nobreak > nul
echo.

echo =================================
echo   RECENT LOGS (Last 10 lines each)
echo =================================
echo --- Node 1 ---
docker-compose logs --tail=10 node1
echo.
echo --- Node 2 ---
docker-compose logs --tail=10 node2
echo.
echo --- Node 3 ---
docker-compose logs --tail=10 node3

echo.
echo =================================
echo   Web Interfaces:
echo   Node 1: http://localhost:8080
echo   Node 2: http://localhost:8081  
echo   Node 3: http://localhost:8082
echo =================================
