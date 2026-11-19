# TALA timing

Measured with `wasm/bench` (`TestTiming`) under `GOOS=js GOARCH=wasm`, run through `go_js_wasm_exec` on Node 22.
Graphs are a random tree (each node links from an earlier random node) plus n/5 cross links between distinct nodes, math/rand seed 42.
One warmup render of `a` (435 ms) runs before timing. Each row is one render.

| Case | Seeds | Time (ms) |
|------|-------|-----------|
| 10 nodes | [1] | 2525 |
| 10 nodes | [1,2,3] | 3781 |
| 30 nodes | [1] | 6587 |
| 30 nodes | [1,2,3] | 18895 |
| 50 nodes | [1] | 27264 |
| 50 nodes | [1,2,3] | 67428 |
| pinned 50 (all shapes have top and left) | [1] | 1497 |
| pinned 50 (all shapes have top and left) | [1,2,3] | 4229 |

Native (non-WASM) reference from the controller: free 50 [1] = 4971 ms, pinned 50 [1] = 272 ms.

## Decision

Fast mode default: ON. 50 nodes with `[1,2,3]` took 67428 ms, far more than 1000 ms.

Free layout is slow (seconds at 30 to 50 nodes, even with a single seed), but fully pinned diagrams are fast. The editor therefore relies on pinning (pin on drop, pin all), and Task 6 should keep rendering off the UI thread.

## WASM build size

Raw 36667557 bytes (35.0 MB), gzip -9 9980831 bytes (9.5 MB). Under the 15 MB gzip budget.
