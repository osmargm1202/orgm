@echo off
setlocal enabledelayedexpansion

REM ORGM Windows Installer
REM This script downloads and installs the ORGM CLI tool

echo 🚀 Installing ORGM CLI for Windows...

REM Variables
set "INSTALL_DIR=%USERPROFILE%\.config\orgm"
set "BINARY_URL=https://raw.githubusercontent.com/osmargm1202/orgm/main/orgm.exe"
set "BINARY_PATH=%INSTALL_DIR%\orgm.exe"
set "WAILS_BINARY_URL=https://raw.githubusercontent.com/osmargm1202/orgm/main/orgm-prop.exe"
set "WAILS_BINARY_PATH=%INSTALL_DIR%\orgm-prop.exe"

REM Create installation directory if it doesn't exist
echo 📁 Creating installation directory: %INSTALL_DIR%
if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
)

REM Download the binary using PowerShell
echo 📥 Downloading ORGM binary...
powershell -Command "try { Invoke-WebRequest -Uri '%BINARY_URL%' -OutFile '%BINARY_PATH%' -UseBasicParsing } catch { Write-Host 'Error downloading file: ' $_.Exception.Message; exit 1 }"

if not exist "%BINARY_PATH%" (
    echo ❌ Error: Failed to download ORGM binary
    exit /b 1
)

REM Download the Wails binary
echo 📥 Downloading ORGM Wails binary...
powershell -Command "try { Invoke-WebRequest -Uri '%WAILS_BINARY_URL%' -OutFile '%WAILS_BINARY_PATH%' -UseBasicParsing } catch { Write-Host 'Error downloading Wails file: ' $_.Exception.Message; exit 1 }"

if not exist "%WAILS_BINARY_PATH%" (
    echo ❌ Error: Failed to download ORGM Wails binary
    exit /b 1
)

REM Check if the installation directory is in PATH
echo 🔍 Checking PATH configuration...
echo %PATH% | findstr /C:"%INSTALL_DIR%" >nul
if errorlevel 1 (
    echo ⚠️  %INSTALL_DIR% is not in your PATH
    echo 💡 Adding %INSTALL_DIR% to your PATH...
    
    REM Add to user PATH environment variable
    for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v PATH 2^>nul') do set "CurrentPath=%%B"
    if not defined CurrentPath set "CurrentPath="
    
    REM Check if the path is already there
    echo !CurrentPath! | findstr /C:"%INSTALL_DIR%" >nul
    if errorlevel 1 (
        if defined CurrentPath (
            set "NewPath=!CurrentPath!;%INSTALL_DIR%"
        ) else (
            set "NewPath=%INSTALL_DIR%"
        )
        reg add "HKCU\Environment" /v PATH /t REG_EXPAND_SZ /d "!NewPath!" /f >nul
        echo ✅ Added to PATH registry
        echo 🔄 Refreshing environment...
        REM Notify system of environment change
        powershell -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User'), 'User')"
    ) else (
        echo ✅ %INSTALL_DIR% is already in PATH registry
    )
    
    echo ""
    echo 🔄 To use orgm immediately, either:
    echo    1. Open a new Command Prompt or PowerShell window
    echo    2. Or use the full path: %BINARY_PATH%
) else (
    echo ✅ %INSTALL_DIR% is already in your PATH
)

REM Test installation
echo.
echo 🧪 Testing installation...
"%BINARY_PATH%" version >nul 2>&1
if errorlevel 1 (
    echo ⚠️  Installation completed but unable to verify. Try running:
    echo    %BINARY_PATH% version
) else (
    echo ✅ ORGM CLI installed successfully!
    echo 📍 Installed at: %BINARY_PATH%
    echo 📍 Wails binary at: %WAILS_BINARY_PATH%
    echo.
    echo 🎉 You can now use 'orgm' command in new terminals!
    echo 💡 Try: orgm --help
    echo 💡 Try: orgm prop wails (for GUI interface)
)

echo.
echo 📚 To update ORGM in the future, run: orgm update

pause
