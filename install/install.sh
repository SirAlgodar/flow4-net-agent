#!/usr/bin/env bash
set -e

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux*)   TARGET="linux-amd64" ;;
  darwin*)  TARGET="darwin-amd64" ;;
  *) echo "unsupported os: $OS"; exit 1 ;;
esac

INSTALL_DIR="$HOME/.flow4network/agent"
mkdir -p "$INSTALL_DIR"

URL="https://example.com/flow4-net-agent-$TARGET.tar.gz"

curl -fsSL "$URL" | tar -xz -C "$INSTALL_DIR"

echo "installed to $INSTALL_DIR"
