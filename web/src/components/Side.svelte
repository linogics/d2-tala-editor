<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import Properties from './Properties.svelte'
  import ThemePanel from './ThemePanel.svelte'
  import ClassesPanel from './ClassesPanel.svelte'

  let { state: app }: { state: AppState } = $props()

  const tabs = ['Properties', 'Classes', 'Theme'] as const
  let tab = $state<(typeof tabs)[number]>('Properties')
</script>

<aside class="box">
  <div class="panel-head tabs" role="tablist">
    {#each tabs as t (t)}
      <button role="tab" aria-selected={tab === t} class:on={tab === t} onclick={() => (tab = t)}>{t}</button>
    {/each}
  </div>
  <!-- Properties stays mounted for its Delete key handler. -->
  <div class="pane" hidden={tab !== 'Properties'}><Properties state={app} /></div>
  {#if tab === 'Classes'}<div class="pane"><ClassesPanel state={app} /></div>{/if}
  {#if tab === 'Theme'}<div class="pane"><ThemePanel state={app} /></div>{/if}
</aside>

<style>
  .box { display: flex; flex-direction: column; min-height: 0; background: var(--surface); border-right: 1px solid var(--border); }
  .tabs { gap: 4px; justify-content: flex-start; }
  .tabs button { padding: 3px 8px; font-size: 12px; }
  .pane { flex: 1; min-height: 0; overflow-y: auto; }
  .pane[hidden] { display: none; }
</style>
