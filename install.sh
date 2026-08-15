#!/usr/bin/env bash

set -euo pipefail

echo "Installing Macaron..."

SOURCE_DIR="/tmp/macaron-source-$(date +%s)"

TARGET_USER="${SUDO_USER:-$USER}"
if [ -n "${SUDO_USER:-}" ]; then
  TARGET_HOME=$(dscl . -read "/Users/$TARGET_USER" NFSHomeDirectory | awk '{print $2}')
else
  TARGET_HOME="$HOME"
fi

git clone --depth 1 https://github.com/clemenzi/macaron.git "$SOURCE_DIR"
cd "$SOURCE_DIR"

cp macaron /usr/local/bin/
chmod +x /usr/local/bin/macaron

mkdir -p "$TARGET_HOME/.config/macaron/services"
if [ -n "${SUDO_USER:-}" ]; then
  chown -R "$TARGET_USER" "$TARGET_HOME/.config/macaron"
fi

echo "Macaron installed successfully!"
