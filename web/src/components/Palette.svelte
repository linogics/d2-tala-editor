<script lang="ts">
  import { templates } from '../lib/templates'

  const groups = [
    { name: 'Shapes', key: 'shape' },
    { name: 'Special objects', key: 'special' },
  ]

  function start(e: DragEvent, key: string) {
    e.dataTransfer!.setData('application/x-d2-template', key)
    e.dataTransfer!.effectAllowed = 'copy'
  }
</script>

<aside class="box">
  <div class="panel-head">Palette</div>
  <div class="list">
    {#each groups as g (g.key)}
      <h4>{g.name}</h4>
      {#each templates.filter((t) => t.group === g.key) as t (t.key)}
        <div class="item" draggable="true" role="listitem" data-template={t.key} ondragstart={(e) => start(e, t.key)}>
          <svg class="icon" class:special={g.key === 'special'} viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
            {#if g.key === 'shape'}
              <rect x="2" y="3" width="12" height="10" rx="2" />
            {:else}
              <rect x="2" y="2.5" width="12" height="11" rx="1.5" /><path d="M2 6.5h12M6 6.5v7" />
            {/if}
          </svg>
          {t.label}
        </div>
      {/each}
    {/each}
  </div>
</aside>

<style>
  .box { display: flex; flex-direction: column; min-height: 0; background: var(--surface); border-right: 1px solid var(--border); }
  .list { padding: 4px 8px 12px; overflow-y: auto; }
  h4 { margin: 12px 6px 4px; font-size: 0.8125rem; font-weight: 600; color: var(--muted); }
  .item { display: flex; align-items: center; gap: 8px; padding: 5px 8px; cursor: grab; border-radius: 6px; border: 1px solid transparent; transition: background 0.12s, border-color 0.12s; }
  .item:hover { background: var(--primary-soft); border-color: var(--primary-soft); }
  .item:active { cursor: grabbing; }
  .icon { flex: none; fill: none; stroke: var(--primary); stroke-width: 1.5; }
  .icon.special { stroke: var(--secondary); }
</style>
