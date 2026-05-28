#!/usr/bin/env bash
# Install mockapi — detects OS/arch, downloads the right binary, places it in /usr/local/bin.
# Usage: curl -sSL https://raw.githubusercontent.com/aakash-a-dev/blackboard/main/blackboard-cli/scripts/install.sh | bash

set -euo pipefail

REPO="aakash-a-dev/blackboard"
BINARY="mockapi"
INSTALL_DIR="/usr/local/bin"

# ── detect OS ──────────────────────────────────────────────────────────────
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)
    echo "error: unsupported OS: $OS"
    exit 1
    ;;
esac

# ── detect arch ────────────────────────────────────────────────────────────
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)          ARCH="amd64" ;;
  arm64 | aarch64) ARCH="arm64" ;;
  *)
    echo "error: unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# ── fetch latest version tag ───────────────────────────────────────────────
echo "Fetching latest mockapi release..."
VERSION="$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"

if [[ -z "$VERSION" ]]; then
  echo "error: could not determine latest version"
  exit 1
fi

echo "Installing mockapi ${VERSION} (${OS}/${ARCH})..."

# ── download & extract ─────────────────────────────────────────────────────
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

ARCHIVE="${BINARY}_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

curl -sSfL "$URL" -o "$TMP/$ARCHIVE"
tar -xzf "$TMP/$ARCHIVE" -C "$TMP"

# ── install ────────────────────────────────────────────────────────────────
if [[ -w "$INSTALL_DIR" ]]; then
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
else
  sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"

echo ""
echo "  mockapi ${VERSION} installed → $INSTALL_DIR/$BINARY"
echo "  Run: mockapi init"
