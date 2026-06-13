#!/bin/sh
# install.sh — one-line installer for pong-ball
# POSIX sh — no bash required.
# Usage:  curl -fsSL https://raw.githubusercontent.com/subhadeeproy3902/pong-ball/main/install.sh | sh
set -eu

REPO="subhadeeproy3902/pong-ball"
BINARY="pong-ball"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# ── colour helpers ──────────────────────────────────────────────────────────
# Real ESC chars (not literal \033) so plain printf '%s' renders them anywhere.
ESC=$(printf '\033')
RED="${ESC}[0;31m"; GREEN="${ESC}[0;32m"; CYAN="${ESC}[0;36m"; BOLD="${ESC}[1m"; RESET="${ESC}[0m"
info()    { printf '%s[pong-ball]%s %s\n' "$CYAN" "$RESET" "$*"; }
success() { printf '%s[pong-ball]%s %s\n' "$GREEN" "$RESET" "$*"; }
error()   { printf '%s[pong-ball] ERROR:%s %s\n' "$RED" "$RESET" "$*" >&2; exit 1; }

# ── detect OS ───────────────────────────────────────────────────────────────
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
  linux)   OS="linux"   ;;
  darwin)  OS="darwin"  ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)       error "Unsupported OS: $OS" ;;
esac

# ── detect arch ─────────────────────────────────────────────────────────────
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)             error "Unsupported architecture: $ARCH" ;;
esac

# ── fetch latest tag ────────────────────────────────────────────────────────
info "Fetching latest release…"
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\(.*\)".*/\1/')
[ -z "$LATEST" ] && error "Could not determine latest version. Check your internet connection."
info "Latest version: ${BOLD}${LATEST}${RESET}"

# ── build download URL ───────────────────────────────────────────────────────
EXT="tar.gz"
[ "$OS" = "windows" ] && EXT="zip"
FILENAME="${BINARY}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

# ── download ────────────────────────────────────────────────────────────────
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

info "Downloading ${FILENAME}…"
curl -fsSL "$URL" -o "$TMP/$FILENAME" \
  || error "Download failed. URL: $URL"

# ── extract ─────────────────────────────────────────────────────────────────
info "Extracting…"
if [ "$EXT" = "zip" ]; then
  command -v unzip >/dev/null 2>&1 || error "unzip not found — install it first"
  unzip -q "$TMP/$FILENAME" -d "$TMP"
else
  tar -xzf "$TMP/$FILENAME" -C "$TMP"
fi

# ── install ─────────────────────────────────────────────────────────────────
BIN_NAME="$BINARY"
[ "$OS" = "windows" ] && BIN_NAME="${BINARY}.exe"
BIN_PATH="$TMP/$BIN_NAME"
[ ! -f "$BIN_PATH" ] && BIN_PATH=$(find "$TMP" -name "$BIN_NAME" -type f | head -1)
[ -z "$BIN_PATH" ] && error "Binary not found in archive"

chmod +x "$BIN_PATH"

DEST="$INSTALL_DIR/$BIN_NAME"
if [ -w "$INSTALL_DIR" ]; then
  mv "$BIN_PATH" "$DEST"
else
  info "Requesting sudo to install to $INSTALL_DIR…"
  sudo mv "$BIN_PATH" "$DEST"
fi

# ── verify ──────────────────────────────────────────────────────────────────
INSTALLED_VER="$("$DEST" version 2>/dev/null | head -1 || echo '?')"
success "Installed ${BOLD}${BINARY}${RESET} → ${DEST}"
success "Version: ${INSTALLED_VER}"
printf '\n'
printf '  Run %spong-ball%s to play!\n' "$CYAN" "$RESET"
printf '  Run %spong-ball --help%s for all commands.\n' "$CYAN" "$RESET"