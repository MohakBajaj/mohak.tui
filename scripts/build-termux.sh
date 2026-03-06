#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$ROOT_DIR/dist/termux"
BIN_PATH="$OUTPUT_DIR/tui-server-linux-arm64"
GO_MOD_CACHE="$ROOT_DIR/.cache/go-mod"
GO_BUILD_CACHE="$ROOT_DIR/.cache/go-build"

echo "Building Termux artifacts..."

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR" "$GO_MOD_CACHE" "$GO_BUILD_CACHE"

bun install --frozen-lockfile

echo "Compiling TUI server for linux/arm64..."
(
  cd "$ROOT_DIR/apps/tui-server"
  GOMODCACHE="$GO_MOD_CACHE" \
    GOCACHE="$GO_BUILD_CACHE" \
    GOOS=linux \
    GOARCH=arm64 \
    go build -o "$BIN_PATH" .
)

cat >"$OUTPUT_DIR/run-termux.sh" <<'EOF'
#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -f "$SCRIPT_DIR/.env" ]]; then
  set -a
  source "$SCRIPT_DIR/.env"
  set +a
fi

"$SCRIPT_DIR/tui-server-linux-arm64"
EOF

cat >"$OUTPUT_DIR/README.termux.md" <<'EOF'
# Termux build output

This folder contains:

- `tui-server-linux-arm64`: single self-contained Go binary for linux/arm64
- `run-termux.sh`: convenience launcher for Termux

## Recommended on Termux

1. Copy this folder to your device.
2. Run `chmod +x ./run-termux.sh ./tui-server-linux-arm64`.
3. Start the service with `./run-termux.sh`.
4. Connect with `ssh -p 2222 localhost`.

The portfolio content is embedded into the Go binary, so no extra runtime files are required.
EOF

chmod +x \
  "$OUTPUT_DIR/run-termux.sh" \
  "$BIN_PATH"

echo ""
echo "Termux artifacts written to $OUTPUT_DIR"
