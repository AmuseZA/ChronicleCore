# ChronicleCore Packaging Guide

This guide details the single, official method for building the production installer (`.exe`) for ChronicleCore.

## Prerequisites

- **Python 3.8+**: Installed and added to PATH.
- **Node.js 18+**: For building the frontend.
- **Go 1.21+**: For building the backend.
- **GCC**: Required for `go-sqlite3` (e.g., MinGW or TDM-GCC).
- **Inno Setup 6**: Required for compiling the installer.

## Build Process (Official)

The build process is automated by two PowerShell scripts.

### 1. Preparation

This step builds the backend, builds the frontend, downloads the portable Python environment, and stages all files into the `dist_installer` directory.

Run this from the project root:

```powershell
.\prepare_installer.ps1 -Version 1.5.0
```

**What this does:**
- Cleans previous builds.
- Compiles `apps/chroniclecore-core` -> `chroniclecore.exe`.
- Builds `apps/chroniclecore-ui` -> `apps/chroniclecore-ui/build`.
- Downloads Python 3.11 (embedded) and installs all `requirements.txt` dependencies.
- Copies everything to `dist_installer/`.

### 2. Packaging

This step takes the staged files and compiles them into a single executable installer using Inno Setup.

Run this immediately after preparation:

```powershell
.\build_inno_installer.ps1 -Version 1.5.0
```

**What this does:**
- Updates the version in `chroniclecore.iss`.
- Compiles the installer using ISCC.
- Outputs the final file to `installer_output/`.

## Output

The final installer will be located at:
`installer_output/ChronicleCore_Setup_v1.5.0.exe`

## Troubleshooting

- **"ISCC not found"**: Ensure Inno Setup 6 is installed. The script looks in default locations.
- **"Backend not found"**: Ensure `prepare_installer.ps1` ran successfully first.
- **"Python download failed"**: Check your internet connection; the script downloads Python from python.org.
