#!/usr/bin/env bash

set -euo pipefail

INSTALL_PATH="/usr/local/bin/macaron"

status() { printf '%s\n' "$*"; }

if [ -e "$INSTALL_PATH" ]; then
  IS_UPDATE=true
  status "🔄 Updating Macaron"
else
  IS_UPDATE=false
  status "🔄 Downloading latest Macaron"
fi

TARGET_USER="${SUDO_USER:-$USER}"
if [ -n "${SUDO_USER:-}" ]; then
  TARGET_HOME=$(dscl . -read "/Users/$TARGET_USER" NFSHomeDirectory | awk '{print $2}')
else
  TARGET_HOME="$HOME"
fi

case "$(uname -m)" in
  arm64) ARCH="arm64" ;;
  x86_64) ARCH="amd64" ;;
  *)
    echo "Unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

TEMP_FILE=$(mktemp)
trap 'rm -f "$TEMP_FILE"' EXIT

curl -fsSL "https://github.com/clemenzi/macaron/releases/latest/download/macaron_darwin_$ARCH" -o "$TEMP_FILE"
install -m 755 "$TEMP_FILE" "$INSTALL_PATH"

mkdir -p "$TARGET_HOME/.config/macaron/services" "$TARGET_HOME/.config/macaron/services-disabled"
if [ -n "${SUDO_USER:-}" ]; then
  chown -R "$TARGET_USER" "$TARGET_HOME/.config/macaron"
fi

if [ "$IS_UPDATE" = true ]; then
  status "✅ Macaron updated successfully"
else
  status "✅ Macaron installed successfully"
fi
