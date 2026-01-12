#!/bin/bash
# PODX Installer for Linux/macOS
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash -s -- --beta
#   curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/install.sh | bash -s -- --version v1.0.0

set -e

REPO="dwirx/podx"
INSTALL_DIR="${PODX_INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="podx"
USE_BETA=false
SPECIFIC_VERSION=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_color() {
    printf "${1}${2}${NC}\n"
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --beta|-b)
            USE_BETA=true
            shift
            ;;
        --version|-v)
            SPECIFIC_VERSION="$2"
            shift 2
            ;;
        --dir|-d)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help|-h)
            echo "PODX Installer"
            echo ""
            echo "Usage: install.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --beta, -b          Install beta/dev version"
            echo "  --version, -v VER   Install specific version (e.g., v1.0.0)"
            echo "  --dir, -d DIR       Install to specific directory"
            echo "  --help, -h          Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Detect OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$OS" in
    linux)  OS="linux" ;;
    darwin) OS="darwin" ;;
    msys*|mingw*|cygwin*)
        print_color $RED "❌ Please use install.ps1 for Windows"
        exit 1
        ;;
    *)
        print_color $RED "❌ Unsupported OS: $OS"
        exit 1
        ;;
esac

case "$ARCH" in
    x86_64|amd64)   ARCH="amd64" ;;
    arm64|aarch64)  ARCH="arm64" ;;
    armv7l)         ARCH="arm" ;;
    *)
        print_color $RED "❌ Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

ASSET_NAME="podx-${OS}-${ARCH}"

print_color $CYAN "🔐 PODX Installer"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "   Platform: $OS/$ARCH"
echo "   Install:  $INSTALL_DIR"

if [ "$USE_BETA" = true ]; then
    print_color $YELLOW "   Channel:  BETA (development)"
else
    echo "   Channel:  Stable"
fi
echo ""

# Determine which release to download
if [ -n "$SPECIFIC_VERSION" ]; then
    RELEASE_URL="https://api.github.com/repos/${REPO}/releases/tags/${SPECIFIC_VERSION}"
    print_color $CYAN "📦 Fetching version: $SPECIFIC_VERSION"
elif [ "$USE_BETA" = true ]; then
    RELEASE_URL="https://api.github.com/repos/${REPO}/releases/tags/beta"
    print_color $YELLOW "📦 Fetching beta release..."
else
    RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
    print_color $CYAN "📦 Fetching latest stable release..."
fi

# Get release info
RELEASE_INFO=$(curl -fsSL "$RELEASE_URL" 2>/dev/null) || {
    print_color $RED "❌ Failed to fetch release info"
    if [ "$USE_BETA" = true ]; then
        echo "   Beta release may not exist yet. Try stable version."
    fi
    exit 1
}

# Extract download URL
DOWNLOAD_URL=$(echo "$RELEASE_INFO" | grep "browser_download_url.*${ASSET_NAME}\"" | head -1 | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    print_color $RED "❌ Could not find release for ${ASSET_NAME}"
    echo "   Available assets:"
    echo "$RELEASE_INFO" | grep "browser_download_url" | cut -d '"' -f 4 | sed 's/^/   - /'
    exit 1
fi

# Extract version
VERSION=$(echo "$RELEASE_INFO" | grep '"tag_name"' | head -1 | cut -d '"' -f 4)
print_color $GREEN "   Version: $VERSION"
echo ""

# Download
TMP_FILE=$(mktemp)
print_color $CYAN "⬇️  Downloading..."
echo "   URL: $DOWNLOAD_URL"

if command -v wget &> /dev/null; then
    wget -q --show-progress -O "$TMP_FILE" "$DOWNLOAD_URL" 2>&1 || curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"
else
    curl -fsSL --progress-bar "$DOWNLOAD_URL" -o "$TMP_FILE"
fi

# Verify download
if [ ! -s "$TMP_FILE" ]; then
    print_color $RED "❌ Download failed or file is empty"
    rm -f "$TMP_FILE"
    exit 1
fi

echo ""

# Check if we need sudo
NEED_SUDO=false
if [ ! -w "$INSTALL_DIR" ]; then
    NEED_SUDO=true
fi

# Create install directory if needed
if [ ! -d "$INSTALL_DIR" ]; then
    if [ "$NEED_SUDO" = true ]; then
        sudo mkdir -p "$INSTALL_DIR"
    else
        mkdir -p "$INSTALL_DIR"
    fi
fi

# Install
chmod +x "$TMP_FILE"
print_color $CYAN "📁 Installing to $INSTALL_DIR..."

if [ "$NEED_SUDO" = true ]; then
    sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
else
    mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
fi

# Verify installation
if [ ! -x "${INSTALL_DIR}/${BINARY_NAME}" ]; then
    print_color $RED "❌ Installation failed"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
print_color $GREEN "✅ PODX installed successfully!"
echo ""
echo "   Location: ${INSTALL_DIR}/${BINARY_NAME}"
echo "   Version:  $("${INSTALL_DIR}/${BINARY_NAME}" version 2>/dev/null | head -1 || echo "$VERSION")"
echo ""
print_color $CYAN "🚀 Quick Start:"
echo "   podx                  # Open interactive TUI"
echo "   podx keygen -t age    # Generate encryption key"
echo "   podx init             # Initialize project"
echo "   podx encrypt-all      # Encrypt all secrets"
echo ""

if [ "$USE_BETA" = true ]; then
    print_color $YELLOW "⚠️  You installed a BETA version. Report bugs at:"
    echo "   https://github.com/${REPO}/issues"
    echo ""
fi

# Check if in PATH
if ! command -v podx &> /dev/null; then
    print_color $YELLOW "⚠️  ${INSTALL_DIR} is not in your PATH"
    echo "   Add this to your shell profile (~/.bashrc, ~/.zshrc):"
    echo "   export PATH=\"\$PATH:${INSTALL_DIR}\""
fi
