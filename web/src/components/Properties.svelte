<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import { templates } from '../lib/templates'

  let { state: app }: { state: AppState } = $props()

  const sel = $derived(app.selection)
  const shape = $derived(sel?.kind === 'shape' ? app.render?.shapes.find((s) => s.id === sel.id) : undefined)
  const types = $derived(shape && !templates.some((t) => t.key === shape.type) ? [shape.type, ...templates.map((t) => t.key)] : templates.map((t) => t.key))

  const val = (e: Event) => (e.currentTarget as HTMLInputElement | HTMLSelectElement).value
  const set = (attr: string, e: Event) => app.applyEdit('set', { id: sel!.id, attr, value: val(e) || null })

  async function rename(e: Event) {
    const name = val(e).trim()
    if (!name || name === sel!.id.split('.').pop()) return
    const r = await app.applyEdit('rename', { id: sel!.id, name })
    if (r) app.select({ kind: 'shape', id: r.id })
  }

  function pin(e: Event) {
    const on = (e.currentTarget as HTMLInputElement).checked
    app.applyEdit(on ? 'pin' : 'unpin', { id: shape!.id, x: Math.round(shape!.x), y: Math.round(shape!.y) } as never)
  }

  async function del() {
    if (await app.applyEdit('delete', { id: sel!.id })) app.select(null)
  }

  function key(e: KeyboardEvent) {
    if (e.key !== 'Delete' || !sel) return
    const t = e.target as HTMLElement
    if (t.closest?.('input, textarea, select, .cm-editor') || t.isContentEditable) return
    del()
  }
</script>

<svelte:window onkeydown={key} />

<div class="box">
  {#if !sel}
    <p>Drag shapes from the left, or type in the code pane</p>
    <button onclick={() => app.applyEdit('pinAll', {})}>Pin all</button>
  {:else}
    {#key sel.id}
      {#if shape}
        <label>Id <input value={sel.id.split('.').pop()} onchange={rename} /></label>
      {/if}
      <label>Label <input onchange={(e) => set('label', e)} /></label>
      {#if shape}
        <label>
          Shape
          <select value={shape.type} onchange={(e) => set('shape', e)}>
            {#each types as t (t)}<option value={t}>{t}</option>{/each}
          </select>
        </label>
        <label>Fill <input type="color" onchange={(e) => set('style.fill', e)} /></label>
        <label>Stroke <input type="color" onchange={(e) => set('style.stroke', e)} /></label>
        <label>Font size <input type="number" min="1" onchange={(e) => set('style.font-size', e)} /></label>
        <label>Icon URL <input type="url" onchange={(e) => set('icon', e)} /></label>
        <label title={shape.pinnable ? '' : 'Not pinnable here'}>
          <input type="checkbox" checked={shape.pinned} disabled={!shape.pinnable} onchange={pin} /> Pinned
        </label>
      {/if}
      <button onclick={del}>Delete</button>
    {/key}
  {/if}
</div>

<style>
  .box { padding: 8px; border-right: 1px solid #ccc; height: 100%; box-sizing: border-box; overflow-y: auto; font: 13px sans-serif; }
  label { display: block; margin-bottom: 6px; }
  input:not([type='checkbox']), select { display: block; width: 100%; box-sizing: border-box; }
  p { color: #666; }
</style>
