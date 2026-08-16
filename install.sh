#!/usr/bin/env bash

set -euo pipefail

INSTALL_PATH="/usr/local/bin/macaron"

if [ -e "$INSTALL_PATH" ]; then
  IS_UPDATE=true
  echo "🔄 Updating Macaron..."
else
  IS_UPDATE=false
  echo "🔄 Downloading latest Macaron..."
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

mkdir -p "$TARGET_HOME/.config/macaron/services"
if [ -n "${SUDO_USER:-}" ]; then
  chown -R "$TARGET_USER" "$TARGET_HOME/.config/macaron"
fi

if [ "$IS_UPDATE" = true ]; then
  echo "✅ Macaron updated successfully!"
else
  echo "✅ Macaron installed successfully!"
fi
