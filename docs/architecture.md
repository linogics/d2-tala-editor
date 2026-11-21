# Architecture

The app is a static site. All work happens in the browser: a Go engine compiled to WebAssembly runs in a web worker, and a Svelte app talks to it with JSON messages.

```
Svelte app (main thread)                  Web worker
---------------------------               ------------------------------
AppState ── engine.call(op, args) ──────▶ worker.ts
   ▲          postMessage({id,op,args})     d2editor.call(op, json)
   │                                         │
   └──── {ok, result} or {ok:false,error} ◀─ editor.Dispatch (Go, WASM)
                                               D2 parser, oracle, TALA
```

## Go engine (`wasm/`)

- `cmd/editor` is the WASM entry point. It registers `d2editor.call(op, argsJSON)` on the global object and calls `onD2EditorReady`.
- `editor/dispatch.go` maps op names to handlers and wraps every result in an envelope `{ok, result}` or `{ok: false, error: {code, message, diag}}`. Error codes are `broken_text`, `not_pinnable`, `not_reorderable`, `bad_args`, `unknown_op` and `internal`. Panics are recovered into an `internal` error.
- Every edit op takes the full D2 text and returns the new text. The engine keeps no state between calls.
- Edits go through D2's `d2oracle` where it has an operation (create, set, delete, move, rename, connect). Edits it does not cover (pin, reorder, generic keys, class lists) change the D2 AST directly and format it with `d2format`. Each edit compiles the result and fails with `broken_text` if the input does not compile.
- `editor/render.go` compiles, lays out with TALA and renders SVG. Besides the SVG it returns shape and edge geometry, attributes, classes and the diagram config, which the canvas uses for hit testing and the panels.
- `deps_test.go` makes sure no other layout engine (ELK, dagre) ends up in the WASM build.

### Ops

| Op | Purpose |
|----|---------|
| `render` | Layout and SVG. Optional `seeds` for the fast preview |
| `create`, `insert` | New shape, from a type or from a palette template |
| `set`, `setKey` | Set or delete an attribute of a shape, or any key path such as `vars.d2-config.sketch` |
| `delete`, `move`, `rename` | Structural edits, references are updated |
| `connect`, `reconnect` | Edges |
| `pin`, `unpin`, `pinAll` | Write or remove `top` and `left` |
| `reorder`, `reorderEdge` | Move a declaration or a sequence message before another |
| `renameClass` | Rename a class and its uses |
| `format`, `completions`, `refRanges`, `refRangesAll` | Code editor support |
| `themes` | Theme list with colors |

## Worker (`web/src/lib/worker.ts`)

Loads `wasm_exec.js`, streams `editor.wasm` with progress messages (the loading bar), starts the Go runtime and forwards calls. Calls are synchronous inside the worker, so they run one at a time.

## Svelte app (`web/src/`)

- `lib/engine.ts` turns worker messages into typed promises. Op names and argument types live in `lib/protocol.ts`.
- `lib/state.svelte.ts` holds the app state: code, last render, errors, selection, notice, grid and mode. `applyEdit` runs one op on the current code. `applyEdits` chains several ops and pushes the final text once, so a group action is one undo step.
- The code editor (CodeMirror) is the only writer of `code`. Canvas edits send new text through a sink into the editor, which keeps undo history intact.
- `lib/renderLoop.ts` debounces text changes (300 ms), renders a fast one-seed preview, then a full render after 1.5 s of idle. Results from older requests are dropped. A broken render keeps the last good diagram.
- `components/Canvas.svelte` draws the SVG and an overlay for selection, handles and drop markers. `lib/dragIntent.ts` decides what a drop means (pin, move, reorder, connect).
- `public/templates.json` holds the palette items and `public/examples.json` the examples.

## Build output

`wasm/scripts/build.sh` writes `editor.wasm`, `wasm_exec.js` and the license files into `web/public/`. Vite builds `web/dist/` with a relative base path, so the site works from any path. The release workflow publishes `web/dist/` to GitHub Pages.
