@echo off
echo ====================================================
echo Node Auto-Recovery Test
echo ====================================================

echo Starting 3-node blockchain network...
docker-compose up -d

echo Waiting for nodes to initialize...
timeout /t 10 /nobreak > nul

echo Initial node status:
docker-compose ps

echo.
echo Simulating node2 crash (SIGKILL)...
echo This simulates real failure, not graceful shutdown
docker kill blockchain-node2

echo Waiting 15 seconds to observe auto-recovery...
timeout /t 15 /nobreak > nul

echo Status after crash (should auto-restart):
docker-compose ps

echo.
echo Logs from node2 recovery:
docker logs blockchain-node2 --tail=20

echo.
echo Testing with process kill inside container...
echo Killing the Go process, not the container
docker exec blockchain-node3 pkill blockchain-node

echo Waiting 10 seconds for container restart...
timeout /t 10 /nobreak > nul

echo Final status (node3 should be restarted):
docker-compose ps

echo.
echo Logs from node3 recovery:
docker logs blockchain-node3 --tail=20

echo.
echo Auto-recovery test completed!
echo Containers should automatically restart after process kills
echo To stop all nodes: docker-compose down
pause
