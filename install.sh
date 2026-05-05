#!/bin/bash
set -e

echo "🍊 Squeezing Pulp into your system..."

# Define repo
REPO="Baseplayer23893/Pulp"

# Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then ARCH="amd64"; fi
if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then ARCH="arm64"; fi

# Get the latest version tag from GitHub
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    echo "❌ Could not find latest release. Check your repo settings!"
    exit 1
fi

# Construct filename based on your GitHub Actions matrix
# Format: pulp-linux-amd64
FILENAME="pulp-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then FILENAME="${FILENAME}.exe"; fi

URL="https://github.com/$REPO/releases/download/${LATEST_TAG}/${FILENAME}"

echo "📥 Downloading Pulp ${LATEST_TAG} for ${OS}/${ARCH}..."
curl -L "$URL" -o pulp

# Make executable and move to local bin
chmod +x pulp
sudo mv pulp /usr/local/bin/pulp

echo "✅ Pulp is ready! Run 'pulp' to start squeezing."