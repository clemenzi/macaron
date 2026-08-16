#!/usr/bin/env bash

set -euo pipefail

INSTALL_PATH="/usr/local/bin/macaron"

if [ -t 1 ] && [ -z "${NO_COLOR:-}" ]; then
  GRAY=$'\033[90m'
  GREEN=$'\033[32m'
  RESET=$'\033[0m'
else
  GRAY=""
  GREEN=""
  RESET=""
fi

log() { printf '%s%s%s\n' "$GRAY" "$*" "$RESET"; }
log_success() { printf '%s%s%s\n' "$GREEN" "$*" "$RESET"; }

if [ -e "$INSTALL_PATH" ]; then
  IS_UPDATE=true
  log "🔄 Updating Macaron..."
else
  IS_UPDATE=false
  log "🔄 Downloading latest Macaron..."
fi

TARGET_USER="${SUDO_USER:-$USER}"
if [ -n "${SUDO_USER:-}" ]; then
  TARGET_HOME=$(dscl . -read "/Users/$TARGET_USER" NFSHomeDirectory | awk '{print $2}')
else
  TARGET_HOME="$HOME"
fi

TEMP_FILE=$(mktemp)
trap 'rm -f "$TEMP_FILE"' EXIT

curl -fsSL "https://raw.githubusercontent.com/clemenzi/macaron/main/macaron" -o "$TEMP_FILE"
install -m 755 "$TEMP_FILE" "$INSTALL_PATH"

mkdir -p "$TARGET_HOME/.config/macaron/services" "$TARGET_HOME/.config/macaron/services-disabled"
if [ -n "${SUDO_USER:-}" ]; then
  chown -R "$TARGET_USER" "$TARGET_HOME/.config/macaron"
fi

if [ "$IS_UPDATE" = true ]; then
  log_success "✅ Macaron updated successfully!"
else
  log_success "✅ Macaron installed successfully!"
fi
