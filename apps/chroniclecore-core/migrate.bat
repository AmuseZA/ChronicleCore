@echo off
REM Migration script for ChronicleCore backend

echo Running migration...

REM Set GCC path (installed via winget)
set GCC_PATH=C:\Users\josh\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin\gcc.exe
set GO_PATH=C:\Program Files\Go\bin\go.exe

if not exist "%GCC_PATH%" (
    echo ERROR: GCC not found at %GCC_PATH%
    exit /b 1
)

set CGO_ENABLED=1
set CC=%GCC_PATH%

"%GO_PATH%" run cmd/migrate_temp/main.go

if errorlevel 1 (
    echo.
    echo Migration failed!
    exit /b 1
)

echo.
echo Migration successful!
