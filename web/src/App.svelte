<script lang="ts">
  import { createEngine, createWorker } from './lib/engine'
  import { createAppState } from './lib/state.svelte'
  import { loadGrid, loadInitial, saveCode, saveGrid } from './lib/persist'
  import TopBar from './components/TopBar.svelte'
  import Palette from './components/Palette.svelte'
  import Canvas from './components/Canvas.svelte'
  import Properties from './components/Properties.svelte'
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

  Object.assign(app, loadGrid(storage))
  $effect(() => saveGrid(storage, { snap: app.snap, grid: app.grid }))

  app.engine.onProgress((l, t) => ((loaded = l), (total = t)))
  app.engine.ready.then(() => (ready = true))

  const init = fetch('./examples.json')
    .then((r) => r.json())
    .then(async (ex: { name: string; code: string }[]) => {
      examples = ex
      const text = await loadInitial(
        location,
        storage,
        async (value) => (await app.engine.call('decode', { value })).text,
        ex[0].code,
      )
      app.setCodeFromEditor(text)
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
    <div>Loading engine {mb(loaded)} / {mb(total)} MB</div>
    <progress value={loaded} max={total || 1}></progress>
  </div>
{/if}

{#await init then}
  <div class="app">
    <TopBar {app} {examples} />
    <div class="cols">
      <Palette />
      <Canvas state={app} />
      <Properties state={app} />
      <CodePane state={app} />
    </div>
  </div>
{/await}

<style>
  :global(body) { margin: 0; }
  .app { display: grid; grid-template-rows: auto 1fr; height: 100vh; }
  .cols { display: grid; grid-template-columns: 140px 1fr 220px 34%; min-height: 0; }
  .overlay { position: fixed; inset: 0; z-index: 10; background: #fff; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; }
</style>
