#!/bin/bash
set -e

# Configuration
REPO="isai-arellano/kolyn-cli"
BINARY="kolyn"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Installing Kolyn CLI...${NC}"

# Detect OS and Arch
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)  OS="Linux" ;;
    Darwin) OS="Darwin" ;;
    MINGW*|MSYS*|CYGWIN*) OS="Windows" ;;
    *) echo -e "${RED}Unsupported OS: $OS${NC}"; exit 1 ;;
esac

case "$ARCH" in
    x86_64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

# Construct release filename based on GoReleaser naming
FILENAME="kolyn_${OS}_${ARCH}.tar.gz"
if [ "$OS" = "Windows" ]; then
    FILENAME="kolyn_${OS}_${ARCH}.zip"
fi

echo -e "Detected: ${OS} ${ARCH}"

# Get latest release URL
LATEST_URL="https://github.com/$REPO/releases/latest/download/$FILENAME"

# Download
echo -e "Downloading from $LATEST_URL..."
TMP_DIR=$(mktemp -d)
curl -sL -o "$TMP_DIR/$FILENAME" "$LATEST_URL"

# Extract
echo -e "Extracting..."
if [ "$OS" = "Windows" ]; then
    unzip -q "$TMP_DIR/$FILENAME" -d "$TMP_DIR"
else
    tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"
fi

# Install
echo -e "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
else
    sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
fi

# Cleanup
rm -rf "$TMP_DIR"

echo -e "${GREEN}Kolyn installed successfully!${NC}"
echo -e "Run 'kolyn --help' to get started."
