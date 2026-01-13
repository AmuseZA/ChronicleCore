@echo off
REM Build script for ChronicleCore backend on Windows

echo Building ChronicleCore backend...
echo.

REM Set GCC path (installed via winget)
set GCC_PATH=C:\Users\josh\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin\gcc.exe
set GO_PATH=C:\Program Files\Go\bin\go.exe

REM Check if GCC exists
if not exist "%GCC_PATH%" (
    echo ERROR: GCC not found at %GCC_PATH%
    echo Please install via: winget install BrechtSanders.WinLibs.POSIX.UCRT
    exit /b 1
)

REM Check if Go exists
if not exist "%GO_PATH%" (
    echo ERROR: Go not found at %GO_PATH%
    echo Please install from: https://go.dev/dl/
    exit /b 1
)

REM Build with CGO enabled
echo Using GCC: %GCC_PATH%
echo Using Go: %GO_PATH%
echo.

set CGO_ENABLED=1
set CC=%GCC_PATH%

"%GO_PATH%" build -o chroniclecore.exe ./cmd/server

if errorlevel 1 (
    echo.
    echo Build failed!
    exit /b 1
)

echo.
echo Build successful! Binary: chroniclecore.exe
dir chroniclecore.exe
echo.
echo To run: .\chroniclecore.exe
