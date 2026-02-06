#!/bin/sh
set -e

REPO="HabibPro1999/shiplog"
BINARY="shiplog"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  arm64|aarch64)  ARCH="arm64" ;;
  *)              echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release"
  exit 1
fi

# Download
URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY}_${OS}_${ARCH}.tar.gz"
echo "Downloading $BINARY $LATEST for ${OS}/${ARCH}..."
TMPDIR=$(mktemp -d)
curl -fsSL "$URL" -o "$TMPDIR/archive.tar.gz"
tar -xzf "$TMPDIR/archive.tar.gz" -C "$TMPDIR"

# Install
mkdir -p "$INSTALL_DIR"
mv "$TMPDIR/$BINARY" "$INSTALL_DIR/$BINARY"
chmod +x "$INSTALL_DIR/$BINARY"
rm -rf "$TMPDIR"

echo "Installed $BINARY to $INSTALL_DIR/$BINARY"

# Check PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Note: Add $INSTALL_DIR to your PATH if not already done" ;;
esac
