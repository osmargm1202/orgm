#!/bin/bash

# ORGM Universal Installer
# This script detects the OS and runs the appropriate installer

set -e

echo "🚀 ORGM CLI Universal Installer"
echo "================================"

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     MACHINE=Linux;;
    Darwin*)    MACHINE=Mac;;
    CYGWIN*)    MACHINE=Cygwin;;
    MINGW*)     MACHINE=MinGw;;
    MSYS_NT*)   MACHINE=Windows;;
    *)          MACHINE="UNKNOWN:${OS}"
esac

echo "🔍 Detected OS: $MACHINE"

if [ "$MACHINE" = "Linux" ] || [ "$MACHINE" = "Mac" ]; then
    echo "📥 Downloading Linux installer..."
    curl -fsSL "https://raw.githubusercontent.com/osmargm1202/orgm/master/install.sh" | bash
elif [ "$MACHINE" = "Windows" ] || [ "$MACHINE" = "Cygwin" ] || [ "$MACHINE" = "MinGw" ]; then
    echo "⚠️  For Windows, please run:"
    echo "   Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/osmargm1202/orgm/master/install.bat' -OutFile 'install.bat' && .\\install.bat && del install.bat"
    echo ""
    echo "   Or download install.bat manually from:"
    echo "   https://raw.githubusercontent.com/osmargm1202/orgm/master/install.bat"
else
    echo "❌ Unsupported operating system: $MACHINE"
    echo "📚 Please visit https://github.com/osmargm1202/orgm for manual installation instructions"
    exit 1
fi
