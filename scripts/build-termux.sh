#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$ROOT_DIR/dist/termux"
BIN_DIR="$OUTPUT_DIR/bin"
BUNDLE_DIR="$OUTPUT_DIR/bundle"
CONTENT_DIR="$OUTPUT_DIR/content"
CONTENT_SOURCE_DIR="$ROOT_DIR/packages/shared-content"
GO_MOD_CACHE="$ROOT_DIR/.cache/go-mod"
GO_BUILD_CACHE="$ROOT_DIR/.cache/go-build"

echo "Building Termux artifacts..."

mkdir -p "$BIN_DIR" "$BUNDLE_DIR" "$CONTENT_DIR" "$GO_MOD_CACHE" "$GO_BUILD_CACHE"

bun install --frozen-lockfile

echo "Compiling AI gateway standalone executable..."
bun build \
  --compile \
  --target=bun-linux-arm64 \
  --outfile "$BIN_DIR/ai-gateway-linux-arm64" \
  "$ROOT_DIR/apps/ai-gateway/index.ts"

echo "Bundling AI gateway fallback for Bun on Termux..."
bun build \
  --target=bun \
  --minify \
  --outfile "$BUNDLE_DIR/ai-gateway.js" \
  "$ROOT_DIR/apps/ai-gateway/index.ts"

echo "Compiling TUI server for linux/arm64..."
(
  cd "$ROOT_DIR/apps/tui-server"
  GOMODCACHE="$GO_MOD_CACHE" \
    GOCACHE="$GO_BUILD_CACHE" \
    GOOS=linux \
    GOARCH=arm64 \
    go build -o "$BIN_DIR/tui-server-linux-arm64" .
)

echo "Copying shared content..."
cp "$CONTENT_SOURCE_DIR/bio.md" "$CONTENT_DIR/bio.md"
cp "$CONTENT_SOURCE_DIR/projects.json" "$CONTENT_DIR/projects.json"
cp "$CONTENT_SOURCE_DIR/resume.json" "$CONTENT_DIR/resume.json"
cp "$CONTENT_SOURCE_DIR/theme.json" "$CONTENT_DIR/theme.json"

cat >"$OUTPUT_DIR/run-termux.sh" <<'EOF'
#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

export CONTENT_PATH="${CONTENT_PATH:-$SCRIPT_DIR/content}"
export AI_GATEWAY_PORT="${AI_GATEWAY_PORT:-3001}"
export SSH_PORT="${SSH_PORT:-2222}"
export AI_GATEWAY_URL="${AI_GATEWAY_URL:-http://127.0.0.1:${AI_GATEWAY_PORT}}"

cleanup() {
  if [[ -n "${AI_GATEWAY_PID:-}" ]]; then
    kill "$AI_GATEWAY_PID" 2>/dev/null || true
  fi
}

trap cleanup EXIT INT TERM

if command -v bun >/dev/null 2>&1; then
  bun "$SCRIPT_DIR/bundle/ai-gateway.js" &
else
  "$SCRIPT_DIR/bin/ai-gateway-linux-arm64" &
fi

AI_GATEWAY_PID=$!
"$SCRIPT_DIR/bin/tui-server-linux-arm64"
EOF

cat >"$OUTPUT_DIR/README.termux.md" <<'EOF'
# Termux build output

This folder contains:

- `bin/tui-server-linux-arm64`: Go SSH server compiled for linux/arm64
- `bin/ai-gateway-linux-arm64`: Bun standalone executable for linux/arm64
- `bundle/ai-gateway.js`: Bun bundle fallback for Termux environments where the standalone executable is not compatible
- `content/`: shared runtime content required by both services
- `run-termux.sh`: starts both services with the correct `CONTENT_PATH`

## Recommended on Termux

1. Install Bun in Termux if available in your setup.
2. Copy this folder to your device.
3. Run `chmod +x ./run-termux.sh ./bin/tui-server-linux-arm64 ./bin/ai-gateway-linux-arm64`.
4. Start the services with `./run-termux.sh`.

The launcher prefers the bundled `bun` runtime path when `bun` is installed because Bun standalone linux binaries are not guaranteed to run on Android's libc without a compatibility layer.
EOF

chmod +x \
  "$OUTPUT_DIR/run-termux.sh" \
  "$BIN_DIR/tui-server-linux-arm64" \
  "$BIN_DIR/ai-gateway-linux-arm64"

echo ""
echo "Termux artifacts written to $OUTPUT_DIR"
