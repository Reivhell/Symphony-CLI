#!/bin/sh
# Symphony CLI installer — downloads a release binary from GitHub and verifies SHA256.
# Usage:
#   curl -sSL https://raw.githubusercontent.com/Reivhell/symphony/main/install.sh | sh
# Options:
#   SYMPHONY_VERSION=v1.2.3  pin a release tag (default: latest)
#   sh install.sh --verify-only   download + verify checksums; do not install

set -eu

REPO="${SYMPHONY_GITHUB_REPO:-Reivhell/Symphony-CLI}"
BINARY_NAME="symphony"
VERIFY_ONLY=false

for arg in "$@"; do
  case "$arg" in
    --verify-only) VERIFY_ONLY=true ;;
  esac
done

die() {
  echo "symphony-install: $*" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || die "curl is required"
command -v shasum >/dev/null 2>&1 || command -v sha256sum >/dev/null 2>&1 || die "sha256sum or shasum is required"

# --- detect OS / arch ---
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64 | amd64) ARCH=amd64 ;;
  aarch64 | arm64) ARCH=arm64 ;;
  *) die "unsupported architecture: $ARCH (need amd64 or arm64)" ;;
esac

case "$OS" in
  linux) GOOS=linux ;;
  darwin) GOOS=darwin ;;
  *) die "unsupported OS: $OS (need Linux or macOS)" ;;
esac

TMPDIR="${TMPDIR:-/tmp}"
WORKDIR=$(mktemp -d "$TMPDIR/symphony-install.XXXXXX")
trap 'rm -rf "$WORKDIR"' EXIT INT TERM

API_LATEST="https://api.github.com/repos/${REPO}/releases/latest"
if [ -n "${SYMPHONY_VERSION:-}" ]; then
  TAG="$SYMPHONY_VERSION"
  case "$TAG" in
    v*) ;;
    *) TAG="v${TAG}" ;;
  esac
else
  JSON=$(curl -fsSL "$API_LATEST") || die "failed to fetch release metadata from GitHub"
  TAG=$(printf '%s\n' "$JSON" | grep -o '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' | head -n1 | sed 's/.*"\(v[^"]*\)".*/\1/')
  [ -n "$TAG" ] || die "could not parse latest release tag"
fi

VER="${TAG#v}"
ARCHIVE_BASE="symphony_${VER}_${GOOS}_${ARCH}"
case "$GOOS" in
  windows) die "use Windows builds from the release page or WSL";;
esac
ARCHIVE="${ARCHIVE_BASE}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE}"
SUMS_URL="https://github.com/${REPO}/releases/download/${TAG}/checksums.txt"

echo "Downloading ${ARCHIVE} (${TAG})..."
curl -fsSL "$DOWNLOAD_URL" -o "$WORKDIR/$ARCHIVE" || die "failed to download archive"
curl -fsSL "$SUMS_URL" -o "$WORKDIR/checksums.txt" || die "failed to download checksums.txt"

cd "$WORKDIR"
if command -v sha256sum >/dev/null 2>&1; then
  grep " ${ARCHIVE}$" checksums.txt | sha256sum -c - || die "SHA256 verification failed"
else
  EXPECT=$(grep " ${ARCHIVE}$" checksums.txt | awk '{print $1}')
  ACTUAL=$(shasum -a 256 "$ARCHIVE" | awk '{print $1}')
  [ "$EXPECT" = "$ACTUAL" ] || die "SHA256 mismatch (expected $EXPECT, got $ACTUAL)"
fi

if "$VERIFY_ONLY"; then
  echo "Verify-only: checksum OK for $ARCHIVE"
  exit 0
fi

tar -xzf "$ARCHIVE"
BIN_PATH="$WORKDIR/$BINARY_NAME"
if [ ! -f "$BIN_PATH" ]; then
  BIN_PATH=$(find "$WORKDIR" -type f -name "$BINARY_NAME" ! -name "*.tar.gz" ! -name "*.txt" 2>/dev/null | head -n1)
fi
[ -n "$BIN_PATH" ] && [ -f "$BIN_PATH" ] || die "could not find $BINARY_NAME inside archive"

INSTALL_DIR=""
if [ -w /usr/local/bin ] 2>/dev/null; then
  INSTALL_DIR=/usr/local/bin
elif [ -w "$HOME/bin" ] 2>/dev/null; then
  INSTALL_DIR="$HOME/bin"
else
  INSTALL_DIR="${XDG_BIN_HOME:-$HOME/.local/bin}"
  mkdir -p "$INSTALL_DIR"
fi

DEST="$INSTALL_DIR/$BINARY_NAME"
echo "Installing to $DEST ..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$BIN_PATH" "$DEST"
  chmod 755 "$DEST" || true
else
  echo "Need write access to $INSTALL_DIR (try: mkdir -p $INSTALL_DIR && retry, or use sudo)" >&2
  sudo mv "$BIN_PATH" "$DEST"
  sudo chmod 755 "$DEST" || true
fi

if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
  echo "Warning: $INSTALL_DIR is not on your PATH. Add: export PATH=\"$INSTALL_DIR:\$PATH\"" >&2
fi

"$BINARY_NAME" version || die "installation verification failed (symphony version)"

echo "Symphony installed successfully."
