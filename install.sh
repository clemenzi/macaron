#!/usr/bin/env bash

set -euo pipefail

echo "🔄 Downloading latest Macaron..."

TARGET_USER="${SUDO_USER:-$USER}"
if [ -n "${SUDO_USER:-}" ]; then
  TARGET_HOME=$(dscl . -read "/Users/$TARGET_USER" NFSHomeDirectory | awk '{print $2}')
else
  TARGET_HOME="$HOME"
fi

CONTENT="$(curl -s "https://raw.githubusercontent.com/clemenzi/macaron/main/macaron")"

echo "$CONTENT" > /usr/local/bin/macaron
chmod +x /usr/local/bin/macaron

mkdir -p "$TARGET_HOME/.config/macaron/services"
if [ -n "${SUDO_USER:-}" ]; then
  chown -R "$TARGET_USER" "$TARGET_HOME/.config/macaron"
fi

echo "✅ Macaron installed successfully!"
