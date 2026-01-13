@echo off
REM Build ChronicleCore with CGO enabled

echo Building with CGO enabled...

set CGO_ENABLED=1
set CC=C:\Users\josh\AppData\Local\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin\gcc.exe
set GOOS=windows
set GOARCH=amd64

"C:\Program Files\Go\bin\go.exe" build -tags="sqlite_omit_load_extension" -ldflags="-s -w" -o chroniclecore.exe ./cmd/server

if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!
echo Verifying CGO is enabled...
"C:\Program Files\Go\bin\go.exe" version -m chroniclecore.exe | findstr CGO

pause
