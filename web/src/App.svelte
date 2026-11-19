<script lang="ts">
  import { createEngine, createWorker } from './lib/engine'
  import { createAppState } from './lib/state.svelte'
  import TopBar from './components/TopBar.svelte'
  import Palette from './components/Palette.svelte'
  import Canvas from './components/Canvas.svelte'
  import Properties from './components/Properties.svelte'
  import CodePane from './components/CodePane.svelte'

  const state = createAppState(createEngine(createWorker()))
  // Task 11 replaces this with persistence.
  const init = fetch('./examples.json')
    .then((r) => r.json())
    .then((ex: { code: string }[]) => state.setCodeFromEditor(ex[0].code))
</script>

{#await init then}
  <div class="app">
    <TopBar />
    <div class="cols">
      <Palette />
      <Canvas {state} />
      <Properties />
      <CodePane {state} />
    </div>
  </div>
{/await}

<style>
  :global(body) { margin: 0; }
  .app { display: grid; grid-template-rows: auto 1fr; height: 100vh; }
  .cols { display: grid; grid-template-columns: 140px 1fr 220px 34%; min-height: 0; }
</style>
