<script lang="ts">
  import { createEngine, createWorker } from './lib/engine'
  import { createAppState } from './lib/state.svelte'
  import { currentMode } from './lib/mode'
  import { loadCols, loadGrid, loadInitial, saveCode, saveCols, saveGrid } from './lib/persist'
  import TopBar from './components/TopBar.svelte'
  import Palette from './components/Palette.svelte'
  import Canvas from './components/Canvas.svelte'
  import Side from './components/Side.svelte'
  import CodePane from './components/CodePane.svelte'
  import Resizer from './components/Resizer.svelte'

  const app = createAppState(createEngine(createWorker()))
  let examples = $state<{ name: string; code: string }[]>([])
  let ready = $state(false)
  let loaded = $state(0)
  let total = $state(0)
  const storage = (() => {
    try {
      return localStorage
    } catch {
      return null
    }
  })()

  Object.assign(app, loadGrid(storage), { mode: currentMode() })
  $effect(() => saveGrid(storage, { snap: app.snap, grid: app.grid }))
  // While the mode follows the OS, an OS switch updates the app mode too.
  $effect(() => {
    const mq = matchMedia('(prefers-color-scheme: dark)')
    const sync = () => (app.mode = currentMode())
    mq.addEventListener('change', sync)

    return () => mq.removeEventListener('change', sync)
  })

  const cols = $state(loadCols(storage))
  $effect(() => saveCols(storage, { side: cols.side, code: cols.code }))
  // Palette, canvas, side and code, the canvas keeps at least 200px.
  const PALETTE = 168
  const CANVAS = 200
  const codeW = () => cols.code ?? 0.34 * innerWidth
  const sideW = () => cols.side ?? 248

  app.engine.onProgress((l, t) => ((loaded = l), (total = t)))
  app.engine.ready.then(() => (ready = true))

  // Without examples the editor still opens, with the saved code or an empty one.
  const init = fetch('./examples.json')
    .then((r) => (r.ok ? r.json() : []))
    .catch(() => [])
    .then((ex: { name: string; code: string }[]) => {
      examples = ex
      app.setCodeFromEditor(loadInitial(storage, ex[0]?.code ?? ''))
    })

  $effect(() => {
    const code = app.code
    const t = setTimeout(() => saveCode(storage, code), 500)

    return () => clearTimeout(t)
  })

  const mb = (n: number) => (n / 1048576).toFixed(1)
</script>

{#if !ready}
  <div class="overlay">
    <strong class="brand">d2-tala-editor</strong>
    <div class="track" role="progressbar" aria-label="Loading engine" aria-valuemin={0} aria-valuemax={total || undefined} aria-valuenow={loaded}><div class="fill" style:width="{total ? (100 * loaded) / total : 0}%"></div></div>
    <div class="sub">Loading engine {mb(loaded)} / {mb(total)} MB</div>
  </div>
{/if}

{#await init then}
  <div class="app">
    <TopBar {app} {examples} />
    <div class="cols" style:grid-template-columns="{PALETTE}px minmax({CANVAS}px, 1fr) 0 {sideW()}px 0 {cols.code ? `${cols.code}px` : '34%'}">
      <Palette />
      <Canvas state={app} />
      <Resizer bind:width={cols.side} min={180} max={() => innerWidth - PALETTE - CANVAS - codeW()} label="Resize properties column" />
      <Side state={app} />
      <Resizer bind:width={cols.code} min={200} max={() => innerWidth - PALETTE - CANVAS - sideW()} label="Resize code column" />
      <CodePane state={app} />
    </div>
  </div>
{/await}

<style>
  .app { display: grid; grid-template-rows: auto 1fr; height: 100vh; }
  .cols { display: grid; min-height: 0; }
  .overlay { position: fixed; inset: 0; z-index: 10; background: var(--bg); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 14px; }
  .brand { font-family: var(--mono); font-size: 1.4375rem; font-weight: 700; color: var(--primary); }
  .track { width: 260px; height: 6px; border-radius: 999px; background: var(--surface-2); overflow: hidden; }
  .fill { height: 100%; background: var(--primary); transition: width 0.2s; }
  .sub { color: var(--muted); font-size: 0.8125rem; font-variant-numeric: tabular-nums; }
</style>
