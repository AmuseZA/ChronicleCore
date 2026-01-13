#!/bin/bash
# Build script for ChronicleCore backend on Windows (Git Bash)

echo "Building ChronicleCore backend..."
echo

# Set paths
GCC_PATH="C:/Users/josh/AppData/Local/Microsoft/WinGet/Packages/BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe/mingw64/bin/gcc.exe"
GO_PATH="/c/Program Files/Go/bin/go.exe"

# Check if GCC exists
if [ ! -f "$GCC_PATH" ]; then
    echo "ERROR: GCC not found at $GCC_PATH"
    echo "Please install via: winget install BrechtSanders.WinLibs.POSIX.UCRT"
    exit 1
fi

# Check if Go exists
if [ ! -f "$GO_PATH" ]; then
    echo "ERROR: Go not found at $GO_PATH"
    echo "Please install from: https://go.dev/dl/"
    exit 1
fi

# Build with CGO enabled
echo "Using GCC: $GCC_PATH"
echo "Using Go: $GO_PATH"
echo

CGO_ENABLED=1 CC="$GCC_PATH" "$GO_PATH" build -o chroniclecore.exe ./cmd/server

if [ $? -eq 0 ]; then
    echo
    echo "Build successful! Binary: chroniclecore.exe"
    ls -lh chroniclecore.exe
    echo
    echo "To run: ./chroniclecore.exe"
else
    echo
    echo "Build failed!"
    exit 1
fi
