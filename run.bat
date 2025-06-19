@echo off
REM Blockchain CLI Helper Script

if "%1"=="" (
    echo Usage: run.bat [command]
    echo.
    echo Available commands:
    echo   build       - Build the CLI
    echo   demo        - Run Alice Bob demo
    echo   create      - Create new wallet
    echo   help        - Show CLI help
    echo   clean       - Clean build files
    goto :EOF
)

if "%1"=="build" (
    echo Building CLI...
    go build -o cli.exe ./cmd/main.go
    if %ERRORLEVEL% EQU 0 (
        echo ✅ Build successful!
    ) else (
        echo ❌ Build failed!
    )
    goto :EOF
)

if "%1"=="demo" (
    echo Running Alice & Bob Demo...
    .\cli.exe demo
    goto :EOF
)

if "%1"=="create" (
    echo Creating new wallet...
    .\cli.exe create
    goto :EOF
)

if "%1"=="help" (
    .\cli.exe help
    goto :EOF
)

if "%1"=="clean" (
    echo Cleaning build files...
    if exist cli.exe del cli.exe
    if exist user_key.json del user_key.json
    if exist blockchain_data rmdir /s /q blockchain_data
    if exist demo_blockchain rmdir /s /q demo_blockchain
    echo ✅ Cleaned!
    goto :EOF
)

echo Unknown command: %1
echo Run 'run.bat' without arguments to see available commands.
