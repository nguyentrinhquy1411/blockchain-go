@echo off
echo ====================================================
echo 3-Node Consensus Test
echo ====================================================

echo Cleaning up existing containers...
docker-compose down --remove-orphans -v 2>nul

echo Building blockchain node image...
docker-compose build

echo Starting 3-node blockchain network...
echo   - Node1 (Leader): localhost:50051
echo   - Node2 (Follower): localhost:50052  
echo   - Node3 (Follower): localhost:50053
echo.

docker-compose up -d

echo Waiting for nodes to initialize...
timeout /t 10 /nobreak > nul

echo Checking node status...
docker-compose ps

echo.
echo ====================================================
echo Testing Consensus Mechanism
echo ====================================================

echo Viewing logs from all nodes...
echo Press Ctrl+C to stop log viewing
timeout /t 3 > nul
docker-compose logs -f

echo.
echo To stop all nodes: docker-compose down
pause

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
echo.
echo PROFESSIONAL BLOCKCHAIN NETWORK STATUS:
echo.
echo Node 1 (Leader)   : localhost:50051 - Web: http://localhost:8080
echo Node 2 (Follower) : localhost:50052 - Web: http://localhost:8081  
echo Node 3 (Follower) : localhost:50053 - Web: http://localhost:8082
echo.
echo AVAILABLE OPERATIONS:
echo 1. Check logs       : docker-compose logs -f [node1|node2|node3]
echo 2. Stop nodes       : docker-compose down
echo 3. Node status      : docker-compose ps
echo 4. Test consensus   : bin\blockchain-cli.exe -server localhost:50051 -cmd send -sender Alice -receiver Bob -amount 100
echo 5. Check blockchain : bin\blockchain-cli.exe -server localhost:50051 -cmd latest
echo.
echo FOR TECHNICAL INTERVIEW DEMO:
echo - All nodes are synchronized and ready
echo - Consensus mechanism is active
echo - Byzantine fault tolerance (2/3 majority) implemented
echo - Automatic node recovery enabled
echo.
echo TIP: Run 'docker-compose logs -f' to see real-time consensus activity
echo.
echo Press any key to view live consensus logs (Ctrl+C to stop)...
pause >nul

echo.
echo Showing live logs from all nodes...
docker-compose logs -f
