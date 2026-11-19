<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import { templates } from '../lib/templates'
  import { roots } from '../lib/group'

  let { state: app }: { state: AppState } = $props()

  type Kind = 'text' | 'number' | 'color' | 'bool' | 'url'
  type Field = { attr: string; name: string; kind: Kind; list?: string; min?: number; max?: number; step?: number }

  const f = (attr: string, name: string, kind: Kind = 'text', more: Partial<Field> = {}): Field => ({ attr, name, kind, ...more })
  const stroke = [
    f('style.stroke', 'Stroke', 'color'),
    f('style.stroke-width', 'Stroke width', 'number', { min: 0, max: 15 }),
    f('style.stroke-dash', 'Stroke dash', 'number', { min: 0, max: 10 }),
    f('style.opacity', 'Opacity', 'number', { min: 0, max: 1, step: 0.1 }),
    f('style.font-size', 'Font size', 'number', { min: 1 }),
    f('style.font-color', 'Font color', 'color'),
  ]
  const shapeFields = [
    f('top', 'Top', 'number', { min: 0 }),
    f('left', 'Left', 'number', { min: 0 }),
    f('width', 'Width', 'number', { min: 1 }),
    f('height', 'Height', 'number', { min: 1 }),
    f('style.fill', 'Fill', 'color'),
    ...stroke,
    f('style.border-radius', 'Border radius', 'number', { min: 0 }),
    ...['shadow', '3d', 'multiple', 'bold', 'italic', 'underline', 'double-border'].map((k) => f(`style.${k}`, k, 'bool')),
    f('icon', 'Icon URL', 'url'),
    f('tooltip', 'Tooltip'),
    f('link', 'Link', 'url'),
    f('near', 'Near', 'text', { list: 'near' }),
  ]
  const edgeFields = [
    ...stroke,
    ...['source', 'target'].flatMap((end) => [
      f(`${end}-arrowhead.shape`, `${end} arrowhead`, 'text', { list: 'arrowheads' }),
      f(`${end}-arrowhead.label`, `${end} arrowhead label`),
    ]),
    ...['animated', 'bold', 'italic'].map((k) => f(`style.${k}`, k, 'bool')),
  ]
  const nears = ['top-left', 'top-center', 'top-right', 'center-left', 'center-right', 'bottom-left', 'bottom-center', 'bottom-right']
  const arrowheads = ['triangle', 'arrow', 'diamond', 'circle', 'box', 'cross', 'cf-one', 'cf-one-required', 'cf-many', 'cf-many-required']

  // Single item, or null for none and for a group of shapes.
  const sel = $derived(app.selection.length === 1 ? app.selection[0] : null)
  const many = $derived(app.selection.length > 1 ? app.selection.map((x) => app.render?.shapes.find((s) => s.id === x.id)).filter((s) => !!s) : [])
  const shape = $derived(sel?.kind === 'shape' ? app.render?.shapes.find((s) => s.id === sel.id) : many[0])
  const edge = $derived(sel?.kind === 'edge' ? app.render?.edges.find((e) => e.id === sel.id) : undefined)
  // A group shows a value only where every member agrees.
  const attrs = $derived(
    many.length
      ? Object.fromEntries(Object.entries(many[0].attrs).filter(([k, v]) => many.every((s) => s.attrs[k] === v)))
      : (shape?.attrs ?? edge?.attrs ?? {}),
  )
  const fields = $derived(many.length ? shapeFields.filter((x) => x.attr !== 'top' && x.attr !== 'left') : shape ? shapeFields : edge ? edgeFields : [])
  const type = $derived(many.length ? (many.every((s) => s.type === many[0].type) ? many[0].type : '') : shape?.type)
  const types = $derived(type && !templates.some((t) => t.key === type) ? [type, ...templates.map((t) => t.key)] : templates.map((t) => t.key))
  const ids = $derived(app.selection.map((x) => x.id))
  // Members disagree on this attr.
  const mixed = (attr: string) => many.some((s) => s.attrs[attr] !== many[0].attrs[attr])

  const setVal = (attr: string, value: string) => {
    const v = value.trim() || null
    if (ids.length === 1) return app.applyEdit('set', { id: ids[0], attr, value: v })

    return app.applyEdits(ids.map((id) => ['set', { id, attr, value: v }]))
  }
  const set = (attr: string, e: Event) => setVal(attr, (e.currentTarget as HTMLInputElement | HTMLSelectElement).value)
  const isHex = (v: string | undefined) => /^#[0-9a-f]{6}$/i.test(v ?? '')

  async function rename(e: Event) {
    const name = (e.currentTarget as HTMLInputElement).value.trim()
    if (!name || name === sel!.id.split('.').pop()) return
    const r = await app.applyEdit('rename', { id: sel!.id, name })
    if (r) app.select({ kind: 'shape', id: r.id })
  }

  function pin(e: Event) {
    const on = (e.currentTarget as HTMLInputElement).checked
    app.applyEdit(on ? 'pin' : 'unpin', { id: shape!.id, x: Math.round(shape!.x), y: Math.round(shape!.y) } as never)
  }

  async function del() {
    if (ids.length === 1) {
      if (await app.applyEdit('delete', { id: ids[0] })) app.select(null)

      return
    }
    // Children of a selected container go with it, deleting them first would fail.
    if (await app.applyEdits(roots(ids).map((id) => ['delete', { id }]))) app.select(null)
  }

  function key(e: KeyboardEvent) {
    if (e.key !== 'Delete' || !ids.length) return
    const t = e.target as HTMLElement
    if (t.closest?.('input, textarea, select, .cm-editor') || t.isContentEditable) return
    del()
  }
</script>

<svelte:window onkeydown={key} />

<datalist id="near">{#each nears as n (n)}<option value={n}></option>{/each}</datalist>
<datalist id="arrowheads">{#each arrowheads as n (n)}<option value={n}></option>{/each}</datalist>

<aside class="box">
  <div class="panel-head">Properties</div>
  <div class="body">
    {#if !ids.length}
      <p>Drag shapes from the left, or type in the code pane</p>
      <button class="primary" onclick={() => app.applyEdit('pinAll', {})}>Pin all</button>
    {:else}
      {#key ids.join('\n')}
        {#if many.length}
          <p>{many.length} shapes selected</p>
        {:else}
          {#if shape}
            <label>Id <input value={sel!.id.split('.').pop()} onchange={rename} /></label>
          {/if}
          <label>Label <input value={attrs.label ?? ''} onchange={(e) => set('label', e)} /></label>
        {/if}
        {#if shape}
          <label>
            Shape
            <select value={type} onchange={(e) => set('shape', e)}>
              {#if !type}<option value="" disabled>mixed</option>{/if}
              {#each types as t (t)}<option value={t}>{t}</option>{/each}
            </select>
          </label>
        {/if}
        {#if shape && !many.length}
          <label title={shape.pinnable ? '' : 'Not pinnable here'}>
            <input type="checkbox" checked={shape.pinned} disabled={!shape.pinnable} onchange={pin} /> Pinned
          </label>
        {/if}
        {#each fields as fd (fd.attr)}
          {@const v = attrs[fd.attr]}
          {#if fd.kind === 'bool'}
            <label><input type="checkbox" checked={v === 'true'} onchange={(e) => setVal(fd.attr, e.currentTarget.checked ? 'true' : '')} /> {fd.name}</label>
          {:else if fd.kind === 'color'}
            <label>
              {fd.name}
              <span class="color">
                <input value={v ?? ''} placeholder={mixed(fd.attr) ? 'mixed' : 'unset'} onchange={(e) => set(fd.attr, e)} />
                <input type="color" aria-label="{fd.name} picker" value={isHex(v) ? v : '#000000'} onchange={(e) => set(fd.attr, e)} />
              </span>
            </label>
          {:else}
            <label>
              {fd.name}
              <input type={fd.kind} list={fd.list} min={fd.min} max={fd.max} step={fd.step} value={v ?? ''} placeholder={mixed(fd.attr) ? 'mixed' : undefined} onchange={(e) => set(fd.attr, e)} />
            </label>
          {/if}
        {/each}
        <button class="danger" onclick={del}>Delete</button>
      {/key}
    {/if}
  </div>
</aside>

<style>
  .box { display: flex; flex-direction: column; min-height: 0; background: var(--surface); border-right: 1px solid var(--border); }
  .body { padding: 12px; overflow-y: auto; }
  label { display: block; margin-bottom: 10px; font-size: 12px; font-weight: 500; color: var(--muted); }
  label:has(> input[type='checkbox']) { display: flex; align-items: center; gap: 6px; color: var(--text); font-weight: 400; margin-bottom: 6px; }
  input:not([type='checkbox']), select { display: block; width: 100%; box-sizing: border-box; margin-top: 4px; color: var(--text); }
  .color { display: flex; gap: 4px; }
  .color input { margin-top: 4px; }
  .color input[type='color'] { width: 34px; flex: none; padding: 2px; }
  p { color: var(--muted); margin: 0 0 12px; }
  button { margin-top: 8px; }
</style>
