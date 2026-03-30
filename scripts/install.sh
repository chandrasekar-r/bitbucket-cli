#!/bin/sh
# bb — Bitbucket Cloud CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/chandrasekar-r/bitbucket-cli/main/scripts/install.sh | sh
set -eu

REPO="chandrasekar-r/bitbucket-cli"
BINARY="bb"
INSTALL_DIR="/usr/local/bin"

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux)  echo "linux" ;;
    *)      echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64)  echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
  esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

# macOS uses a universal binary (darwin_all) — covers both Intel and Apple Silicon
if [ "$OS" = "darwin" ]; then
  ARCH="all"
fi

# Fetch latest release tag from GitHub API
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "Error: could not determine latest version" >&2
  exit 1
fi

ARCHIVE="${BINARY}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Downloading bb ${VERSION} for ${OS}/${ARCH}..."
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/$ARCHIVE"
tar -xzf "$TMP/$ARCHIVE" -C "$TMP" "$BINARY"
install -m 755 "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"

echo "bb ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
bb version
