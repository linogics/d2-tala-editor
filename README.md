# d2-tala-editor

[![ci](https://github.com/linogics/d2-tala-editor/actions/workflows/ci.yml/badge.svg)](https://github.com/linogics/d2-tala-editor/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/linogics/d2-tala-editor)](https://github.com/linogics/d2-tala-editor/releases/latest)
[![license](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)

Drag and drop editor for [D2](https://d2lang.com) diagrams with the TALA layout engine. It runs fully in the browser (Go compiled to WebAssembly) and shows the D2 code next to the diagram. Edits on either side update the other.

**Open the editor: https://d2-tala-editor.linogics.io**

<!-- TODO: record docs/assets/demo.gif, then replace this line with:
![Dragging a shape on the canvas rewrites the D2 code](docs/assets/demo.gif) -->

## Features

- **Two-way editing.** Type D2 and the diagram re-renders. Drag, connect or edit on the canvas and the code is rewritten for you.
- **TALA layout** in the browser, with a fast preview while you type and a full layout once you pause.
- **Pinning.** Dropped shapes get `top` and `left`, so TALA keeps them in place. "Pin all" freezes the whole layout.
- **Canvas tools.** Palette with shapes and special objects (SQL tables, classes, code, grids, sequence diagrams), move into and out of containers, reorder grid cells and sequence messages, connect with a drag handle, multi-select, arrow key nudging, grid snapping.
- **Properties, classes and themes.** Edit styles of one shape or a whole selection, manage D2 classes, pick a theme, override theme colors for light and dark mode, toggle sketch style.
- **Code editor** with D2 highlighting, completions, inline errors and highlighting of the selected shape's references.
- **Export** to `.d2` and `.svg`. Work is autosaved in your browser.

Your diagrams never leave your browser. There is no backend.

## Documentation

- [User guide](docs/user-guide.md): everything the editor can do, including shortcuts
- [Architecture](docs/architecture.md): how the Go engine, the worker and the Svelte app fit together
- [Development](docs/development.md): build, test, release
- [Performance](docs/perf.md): TALA timings in WASM and why pinning matters

## Quick start for developers

Needs Go (version from `wasm/go.mod`) and Node 22.

```sh
bash wasm/scripts/build.sh          # builds web/public/editor.wasm and license files
cd web && npm ci && npm run dev     # http://localhost:5173
```

Tests:

```sh
cd wasm && go test ./...
cd web && npm run check && npx vitest run
```

See [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

## Limitations

- The first load downloads a WASM file of about 35 MB, which the browser then caches.
- TALA layouts of large diagrams with no pinned shapes take several seconds. Pin shapes to speed them up, see [docs/perf.md](docs/perf.md).

## License

This repository is BSD-3-Clause, see [LICENSE](LICENSE). The built WASM includes D2 and TALA, which are MPL-2.0 and used unmodified as a Go module at v0.9.0. See [licenses.html](https://d2-tala-editor.linogics.io/licenses.html) on the live site for the full notices.

This project is not affiliated with Terrastruct, the authors of D2 and TALA.
