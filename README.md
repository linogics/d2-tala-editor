# d2-tala-editor

Drag and drop editor for [D2](https://d2lang.com) diagrams with the TALA layout engine. It runs fully in the browser (Go compiled to WASM) and shows the D2 code side by side with the diagram. Edits on either side update the other.

Live: https://linogics.github.io/d2-tala-editor/

## Local dev

```
bash wasm/scripts/build.sh
cd web && npm ci && npm run dev
```

## Tests

```
cd wasm && go test ./...
cd web && npm run check && npx vitest run
```

## Performance

TALA layouts of large unpinned diagrams take several seconds in the browser. Pinning shapes (top and left) speeds them up a lot. See [docs/perf.md](docs/perf.md).

## License

This repository is BSD-3-Clause, see [LICENSE](LICENSE). The built WASM includes D2 and TALA, which are MPL-2.0 and used unmodified as a Go module at v0.9.0. See `licenses.html` on the live site for the full notices.
