#!/bin/bash
# PODX Uninstaller for Linux/macOS
# Usage: curl -fsSL https://raw.githubusercontent.com/dwirx/podx/main/uninstall.sh | bash

set -e

BINARY_NAME="podx"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/podx"

echo "🗑️  Uninstalling PODX..."

# Remove binary
if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
    sudo rm -f "${INSTALL_DIR}/${BINARY_NAME}"
    echo "✓ Removed binary: ${INSTALL_DIR}/${BINARY_NAME}"
else
    echo "⚠️  Binary not found: ${INSTALL_DIR}/${BINARY_NAME}"
fi

# Ask about config
if [ -d "$CONFIG_DIR" ]; then
    echo ""
    read -p "Remove config directory ($CONFIG_DIR)? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "$CONFIG_DIR"
        echo "✓ Removed config: $CONFIG_DIR"
    else
        echo "⚠️  Kept config: $CONFIG_DIR"
    fi
fi

echo ""
echo "✅ PODX uninstalled successfully!"
