#!/bin/bash
# musicr installation script for macOS and Linux

set -e

VERSION="${1:-latest}"
OS=$(uname -s)
ARCH=$(uname -m)

# Normalize architecture names
if [ "$ARCH" = "x86_64" ]; then
    ARCH="x86_64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

# Determine download URL based on OS
case "$OS" in
    Darwin)
        # macOS
        if [ "$ARCH" = "arm64" ]; then
            BINARY_NAME="musicr_${VERSION}_darwin_arm64"
        else
            BINARY_NAME="musicr_${VERSION}_darwin_x86_64"
        fi
        ;;
    Linux)
        BINARY_NAME="musicr_${VERSION}_linux_x86_64"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

echo "Downloading musicr ($OS / $ARCH)..."

# Download URL from GitHub releases
GITHUB_REPO="yourusername/musicr"
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/${BINARY_NAME}.tar.gz"

# Create temporary directory
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

# Download the binary
if ! curl -fL "$DOWNLOAD_URL" -o "$TMPDIR/musicr.tar.gz"; then
    echo "Failed to download musicr from $DOWNLOAD_URL"
    exit 1
fi

# Extract binary
tar -xzf "$TMPDIR/musicr.tar.gz" -C "$TMPDIR"

# Determine installation directory
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

# Install binary
if ! install -m 755 "$TMPDIR/musicr" "$INSTALL_DIR/musicr"; then
    echo "Failed to install musicr to $INSTALL_DIR"
    exit 1
fi

echo "✓ musicr installed to $INSTALL_DIR/musicr"
echo ""
echo "Installation Notes:"
echo "  1. Ensure $INSTALL_DIR is in your PATH"
echo "  2. Install mpv:"

if [ "$OS" = "Darwin" ]; then
    echo "     brew install mpv"
else
    echo "     apt install mpv (Debian/Ubuntu)"
    echo "     pacman -S mpv (Arch Linux)"
    echo "     dnf install mpv (Fedora)"
fi

echo ""
echo "  3. Run: musicr 'your search query'"
echo ""
echo "For more info: https://github.com/${GITHUB_REPO}"
