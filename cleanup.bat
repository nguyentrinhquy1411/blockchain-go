@echo off
echo ====================================================
echo Blockchain Project Cleanup
echo ====================================================

echo Removing demo files...
del /q alice_key.json 2>nul
del /q bob_key.json 2>nul
del /q key.json 2>nul
rmdir /s /q demo_blockchain 2>nul
rmdir /s /q blockchain_data 2>nul
rmdir /s /q data 2>nul

echo Stopping Docker containers...
docker-compose down --remove-orphans -v 2>nul

echo Removing build artifacts...
del /q blockchain.exe 2>nul
del /q blockchain-node.exe 2>nul

echo Cleanup completed!
echo   - Demo files removed
echo   - Docker containers stopped
echo   - Build artifacts cleaned
echo.
echo To rebuild: .\setup.bat
pause
