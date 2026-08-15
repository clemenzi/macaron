#!/usr/bin/env bash

set -euo pipefail

echo "Installing Macaron..."

SOURCE_DIR="/tmp/macaron-source-$(date +%s)"

git clone --depth 1 https://github.com/clemenzi/macaron.git "$SOURCE_DIR"
cd "$SOURCE_DIR"

cp macaron /usr/local/bin/
chmod +x /usr/local/bin/macaron

mkdir "$HOME/.config/macaron"
mkdir "$HOME/.config/macaron/services"

echo "Macaron installed successfully!"
