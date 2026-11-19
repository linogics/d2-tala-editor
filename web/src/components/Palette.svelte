<script lang="ts">
  import { templates } from '../lib/templates'

  const groups = [
    { name: 'Shapes', key: 'shape' },
    { name: 'Special objects', key: 'special' },
  ]
  // Generic glyph, special objects get a distinct one.
  const glyph = (g: string) => (g === 'shape' ? '□' : '▤')

  function start(e: DragEvent, key: string) {
    e.dataTransfer!.setData('application/x-d2-template', key)
    e.dataTransfer!.effectAllowed = 'copy'
  }
</script>

<div class="box">
  {#each groups as g (g.key)}
    <h4>{g.name}</h4>
    {#each templates.filter((t) => t.group === g.key) as t (t.key)}
      <div class="item" draggable="true" role="listitem" data-template={t.key} ondragstart={(e) => start(e, t.key)}>
        <span class="icon">{glyph(t.group)}</span>{t.label}
      </div>
    {/each}
  {/each}
</div>

<style>
  .box { padding: 8px; border-right: 1px solid #ccc; height: 100%; box-sizing: border-box; overflow-y: auto; font: 13px sans-serif; }
  h4 { margin: 8px 0 4px; font-size: 12px; color: #666; }
  .item { padding: 3px 6px; cursor: grab; border-radius: 4px; }
  .item:hover { background: #eef2ff; }
  .icon { display: inline-block; width: 1.4em; color: #2563eb; }
</style>
