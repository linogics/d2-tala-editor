# TALA timing

Measured with `wasm/bench` (`TestTiming`) under `GOOS=js GOARCH=wasm`, run through `go_js_wasm_exec` on Node 22.
Graphs have random edges (math/rand seed 42, about 1.3 edges per node). Each row is one render, so the 10 node row includes warmup (font and ruler setup).

| Nodes | Seeds | Time (ms) |
|-------|-------|-----------|
| 10 | [1] | 1572 |
| 10 | [1,2,3] | 996 |
| 50 | [1] | 6486 |
| 50 | [1,2,3] | 11747 |
| 150 | [1] | did not finish in about 9 min (test killed by the 10 min go test timeout) |
| 150 | [1,2,3] | not reached |

A native (non-WASM) run of the same test also did not finish 150 nodes within 500 s, so the slowness is in TALA itself and not only the WASM build.

## Decision

Fast mode default: ON. 50 nodes with `[1,2,3]` took 11747 ms, which is more than 1000 ms.
Even fast mode (`[1]`) takes 6486 ms at 50 nodes, so Task 6 should treat large graphs as slow and keep rendering off the UI thread.
