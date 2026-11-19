#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
out=../web/public
mkdir -p "$out"

GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o "$out/editor.wasm" ./cmd/editor
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$out/wasm_exec.js"

echo "raw:  $(stat -c %s "$out/editor.wasm") bytes"
echo "gzip: $(gzip -9 -c "$out/editor.wasm" | wc -c) bytes"
