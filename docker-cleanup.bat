@echo off
echo === Docker Blockchain Cleanup ===

echo Stopping all blockchain containers...
docker-compose down -v 2>nul

echo Removing blockchain containers if they exist...
docker rm -f blockchain-node1 blockchain-node2 blockchain-node3 2>nul

echo Removing blockchain images if they exist...
docker rmi blockchain-go-node1 blockchain-go-node2 blockchain-go-node3 2>nul

echo Removing blockchain networks...
docker network rm blockchain-go_blockchain-network 2>nul

echo Pruning unused Docker resources...
docker system prune -f

echo === Cleanup Complete ===
