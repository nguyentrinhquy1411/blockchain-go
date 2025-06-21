@echo off
echo Generating protobuf files...

protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/blockchain.proto

if %errorlevel% neq 0 (
    echo Error: Failed to generate protobuf files
    echo Make sure you have protoc and protoc-gen-go installed:
    echo   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    echo   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    pause
    exit /b 1
)

echo Protobuf files generated successfully!
pause
