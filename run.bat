@echo off
REM Blockchain CLI Helper Script

if "%1"=="" (
    echo Usage: run.bat [command]
    echo.
    echo Available commands:
    echo   build         - Build the CLI
    echo   demo          - Run Alice Bob demo
    echo   create        - Create new wallet
    echo   create-alice  - Create Alice's wallet
    echo   create-bob    - Create Bob's wallet
    echo   alice-to-bob  - Send money from Alice to Bob
    echo   send          - Send money to address
    echo   init          - Initialize blockchain
    echo   help          - Show CLI help
    echo   clean         - Clean build files
    goto :EOF
)

if "%1"=="build" (
    echo Building CLI...
    go build -o cli.exe ./cmd/main.go
    if %ERRORLEVEL% EQU 0 (
        echo Build successful!
    ) else (
        echo Build failed!
    )
    goto :EOF
)

if "%1"=="create" (
    echo Creating new wallet...
    if not exist cli.exe (
        echo Building CLI first...
        go build -o cli.exe ./cmd/main.go
    )
    .\cli.exe create
    goto :EOF
)

if "%1"=="create-alice" (
    echo Creating Alice's wallet...
    if not exist cli.exe (
        echo Building CLI first...
        go build -o cli.exe ./cmd/main.go
    )
    .\cli.exe create-alice
    goto :EOF
)

if "%1"=="create-bob" (
    echo Creating Bob's wallet...
    if not exist cli.exe (
        echo Building CLI first...
        go build -o cli.exe ./cmd/main.go
    )
    .\cli.exe create-bob
    goto :EOF
)

if "%1"=="alice-to-bob" (
    if "%2"=="" (
        echo Usage: run.bat alice-to-bob [amount]
        echo Example: run.bat alice-to-bob 75.5
        goto :EOF
    )
    echo Alice sending %2 coins to Bob...
    if not exist cli.exe (
        echo Building CLI first...
        go build -o cli.exe ./cmd/main.go
    )
    .\cli.exe alice-to-bob %2
    goto :EOF
)

if "%1"=="send" (
    if "%3"=="" (
        echo Usage: run.bat send [address] [amount]
        echo Example: run.bat send 437c6e08e2fc87d08d056b8db9fc174fe003560d 50.0
        goto :EOF
    )
    echo Sending %3 coins to %2...
    if not exist cli.exe (
        echo Building CLI first...
        go build -o cli.exe ./cmd/main.go
    )
    .\cli.exe send %2 %3
    goto :EOF
)

if "%1"=="init" (
    echo Initializing blockchain...
    if not exist cli.exe (
        echo Building CLI first...
        go build -o cli.exe ./cmd/main.go
    )
    .\cli.exe init
    goto :EOF
)

if "%1"=="demo" (
    echo Running Alice and Bob Demo...
    if not exist cli.exe (
        echo Building CLI first...
        go build -o cli.exe ./cmd/main.go
    )
    .\cli.exe demo
    goto :EOF
)

if "%1"=="help" (
    if not exist cli.exe (
        echo Building CLI first...
        go build -o cli.exe ./cmd/main.go
    )
    .\cli.exe help
    goto :EOF
)

if "%1"=="clean" (
    echo Cleaning build files...
    if exist cli.exe del cli.exe
    if exist user_key.json del user_key.json
    if exist alice_key.json del alice_key.json
    if exist bob_key.json del bob_key.json
    if exist blockchain_data rmdir /s /q blockchain_data
    if exist demo_blockchain rmdir /s /q demo_blockchain
    echo Cleaned!
    goto :EOF
)

echo Unknown command: %1
echo Run 'run.bat' without arguments to see available commands.
