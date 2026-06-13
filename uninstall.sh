#!/bin/sh
# uninstall.sh - removes pong-ball and ALL its data.
# POSIX sh — no bash required.
# Usage:  curl -fsSL https://raw.githubusercontent.com/subhadeeproy3902/pong-ball/main/uninstall.sh | sh
set -u

ESC=$(printf '\033')
CYAN="${ESC}[0;36m"; GREEN="${ESC}[0;32m"; YELLOW="${ESC}[1;33m"; RESET="${ESC}[0m"
info()    { printf '%s[pong-ball]%s %s\n' "$CYAN" "$RESET" "$*"; }
success() { printf '%s[pong-ball]%s %s\n' "$GREEN" "$RESET" "$*"; }

printf '\n'
printf '%sThis will permanently remove pong-ball and ALL its data:%s\n' "$YELLOW" "$RESET"
echo "  - the pong-ball binary (from your PATH)"
echo "  - saved scores & settings  (~/.pong-ball)"
echo "  - cached sound files       (\${TMPDIR:-/tmp}/pong-ball-sfx)"
echo
printf "Remove everything? (Y/N) "
# Read from the terminal even when run via 'curl ... | bash'.
read -r ans </dev/tty

case "$ans" in
  [Yy]*)
    BIN="$(command -v pong-ball 2>/dev/null || true)"
    if [ -n "$BIN" ]; then
      if [ -w "$BIN" ] || [ -w "$(dirname "$BIN")" ]; then
        rm -f "$BIN"
      else
        info "Requesting sudo to remove $BIN…"
        sudo rm -f "$BIN"
      fi
      info "removed $BIN"
    fi

    rm -rf "$HOME/.pong-ball"
    rm -rf "${TMPDIR:-/tmp}/pong-ball-sfx"

    echo
    success "pong-ball uninstalled. Thanks for playing!"
    echo "  (Installed via Homebrew? 'brew uninstall pong-ball' clears the formula receipt too.)"
    ;;
  *)
    info "Aborted - nothing was removed."
    ;;
esac
