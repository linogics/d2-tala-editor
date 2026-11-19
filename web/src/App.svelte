<script lang="ts">
  import { createEngine, createWorker } from './lib/engine'
  import { createAppState } from './lib/state.svelte'
  import { currentMode } from './lib/mode'
  import { loadGrid, loadInitial, saveCode, saveGrid } from './lib/persist'
  import TopBar from './components/TopBar.svelte'
  import Palette from './components/Palette.svelte'
  import Canvas from './components/Canvas.svelte'
  import Side from './components/Side.svelte'
  import CodePane from './components/CodePane.svelte'

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

  app.engine.onProgress((l, t) => ((loaded = l), (total = t)))
  app.engine.ready.then(() => (ready = true))

  const init = fetch('./examples.json')
    .then((r) => r.json())
    .then((ex: { name: string; code: string }[]) => {
      examples = ex
      app.setCodeFromEditor(loadInitial(storage, ex[0].code))
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
    <div class="track"><div class="fill" style:width="{total ? (100 * loaded) / total : 0}%"></div></div>
    <div class="sub">Loading engine {mb(loaded)} / {mb(total)} MB</div>
  </div>
{/if}

{#await init then}
  <div class="app">
    <TopBar {app} {examples} />
    <div class="cols">
      <Palette />
      <Canvas state={app} />
      <Side state={app} />
      <CodePane state={app} />
    </div>
  </div>
{/await}

<style>
  .app { display: grid; grid-template-rows: auto 1fr; height: 100vh; }
  .cols { display: grid; grid-template-columns: 168px 1fr 248px 34%; min-height: 0; }
  .overlay { position: fixed; inset: 0; z-index: 10; background: var(--bg); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 14px; }
  .brand { font-size: 22px; font-weight: 700; color: var(--primary); }
  .track { width: 260px; height: 6px; border-radius: 999px; background: var(--surface-2); overflow: hidden; }
  .fill { height: 100%; background: var(--primary); transition: width 0.2s; }
  .sub { color: var(--muted); font-size: 12px; font-variant-numeric: tabular-nums; }
</style>
