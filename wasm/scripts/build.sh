#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
out=../web/public
mkdir -p "$out"

GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o "$out/editor.wasm" ./cmd/editor
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$out/wasm_exec.js"

echo "raw:  $(stat -c %s "$out/editor.wasm") bytes"
echo "gzip: $(gzip -9 -c "$out/editor.wasm" | wc -c) bytes"

d2dir="$(go list -m -f '{{.Dir}}' github.com/d2lang/d2)"
mkdir -p "$out/licenses"
install -m 644 "$d2dir/THIRD_PARTY_NOTICES.txt" "$out/licenses/THIRD_PARTY_NOTICES.txt"
install -m 644 "$d2dir/LICENSE.txt" "$out/licenses/D2-MPL-2.0.txt"
install -m 644 ../LICENSE "$out/licenses/LICENSE-BSD-3.txt"
