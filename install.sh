#!/bin/bash

# ORGM Linux Installer
# This script downloads and installs the ORGM CLI tool

set -e

echo "🚀 Installing ORGM CLI for Linux..."

# Variables
INSTALL_DIR="$HOME/.local/bin"
BINARY_URL="https://raw.githubusercontent.com/osmargm1202/orgm/main/orgm"
BINARY_PATH="$INSTALL_DIR/orgm"
WAILS_BINARY_URL="https://raw.githubusercontent.com/osmargm1202/orgm/main/apps/prop/build/bin/orgm-prop"
WAILS_BINARY_PATH="$INSTALL_DIR/orgm-prop"

# Create installation directory if it doesn't exist
echo "📁 Creating installation directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

# Download the binary
echo "📥 Downloading ORGM binary..."
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$BINARY_URL" -o "$BINARY_PATH"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$BINARY_URL" -O "$BINARY_PATH"
else
    echo "❌ Error: Neither curl nor wget is available. Please install one of them."
    exit 1
fi

# Download the Wails binary
echo "📥 Downloading ORGM orgm-prop binary..."
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$WAILS_BINARY_URL" -o "$WAILS_BINARY_PATH"
elif command -v wget >/dev/null 2>&1; then
    wget -q "$WAILS_BINARY_URL" -O "$WAILS_BINARY_PATH"
else
    echo "❌ Error: Neither curl nor wget is available. Please install one of them."
    exit 1
fi

# Make it executable
echo "🔧 Setting executable permissions..."
chmod +x "$BINARY_PATH"
chmod +x "$WAILS_BINARY_PATH"

# Check if ~/.local/bin is in PATH
echo "🔍 Checking PATH configuration..."
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo "⚠️  ~/.local/bin is not in your PATH"
    echo "💡 Adding ~/.local/bin to your PATH in shell profiles..."
    
    # Add to various shell profiles
    for profile in ~/.bashrc ~/.zshrc ~/.profile ~/.bash_profile; do
        if [[ -f "$profile" ]]; then
            if ! grep -q 'export PATH=$PATH:$HOME/.local/bin' "$profile"; then
                echo 'export PATH=$PATH:$HOME/.local/bin' >> "$profile"
                echo "   ✅ Updated $profile"
            fi
        fi
    done
    
    echo ""
    echo "🔄 To use orgm immediately, run one of these:"
    echo "   export PATH=\$PATH:\$HOME/.local/bin"
    echo "   source ~/.bashrc  (or ~/.zshrc, depending on your shell)"
    echo "   Open a new terminal"
else
    echo "✅ ~/.local/bin is already in your PATH"
fi

# Test installation
echo ""
echo "🧪 Testing installation..."
if "$BINARY_PATH" version >/dev/null 2>&1; then
    echo "✅ ORGM CLI installed successfully!"
    echo "📍 Installed at: $BINARY_PATH"
    echo "📍 orgm-prop binary at: $WAILS_BINARY_PATH"
    echo ""
    echo "🎉 You can now use 'orgm' command!"
    echo "💡 Try: orgm --help"
    echo "💡 Try: orgm prop orgm-prop (for GUI interface)"
else
    echo "⚠️  Installation completed but unable to verify. Try running:"
    echo "   $BINARY_PATH version"
fi

echo ""
echo "📚 To update ORGM in the future, run: orgm update"
