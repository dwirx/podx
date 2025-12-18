#!/bin/bash
# PODX Installer for Linux/macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash

set -e

REPO="dwirx/podx"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="podx"

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    *)      echo "❌ Unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)      echo "❌ Unsupported architecture: $ARCH"; exit 1 ;;
esac

ASSET_NAME="podx-${OS}-${ARCH}"

echo "🔐 Installing PODX..."
echo "   OS: $OS, Arch: $ARCH"

# Get latest release
LATEST_URL=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep "browser_download_url.*${ASSET_NAME}" | cut -d '"' -f 4)

if [ -z "$LATEST_URL" ]; then
    echo "❌ Could not find release for ${ASSET_NAME}"
    exit 1
fi

# Download
TMP_FILE=$(mktemp)
echo "⬇️  Downloading from: $LATEST_URL"
curl -fsSL "$LATEST_URL" -o "$TMP_FILE"

# Install
chmod +x "$TMP_FILE"
sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"

echo ""
echo "✅ PODX installed successfully!"
echo "   Location: ${INSTALL_DIR}/${BINARY_NAME}"
echo ""
echo "🚀 Quick Start:"
echo "   podx keygen -t age    # Generate key"
echo "   podx init             # Init project"
echo "   podx encrypt-all      # Encrypt secrets"
echo ""
podx version
