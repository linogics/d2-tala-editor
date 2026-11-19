<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import { templates } from '../lib/templates'
  import { roots } from '../lib/group'
  import { arrowheads, edgeFields, fieldSteps, nears, shapeFieldsFor } from '../lib/fields'
  import FieldInput from './FieldInput.svelte'

  let { state: app }: { state: AppState } = $props()

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
  const fields = $derived(
    many.length
      ? shapeFieldsFor(many.map((s) => s.type)).filter((x) => x.attr !== 'top' && x.attr !== 'left')
      : shape
        ? shapeFieldsFor([shape.type])
        : edge
          ? edgeFields
          : [],
  )
  const type = $derived(many.length ? (many.every((s) => s.type === many[0].type) ? many[0].type : '') : shape?.type)
  const types = $derived(type && !templates.some((t) => t.key === type) ? [type, ...templates.map((t) => t.key)] : templates.map((t) => t.key))
  const ids = $derived(app.selection.map((x) => x.id))
  // Members disagree on this attr.
  const mixed = (attr: string) => many.some((s) => s.attrs[attr] !== many[0].attrs[attr])

  const setVal = (attr: string, value: string) => {
    const v = value.trim() || null
    if (shape && !many.length && (attr === 'top' || attr === 'left')) return app.applyEdits(fieldSteps(shape, attr, v))
    if (ids.length === 1) return app.applyEdit('set', { id: ids[0], attr, value: v })

    return app.applyEdits(ids.map((id) => ['set', { id, attr, value: v }]))
  }
  // Classes come back as "a;b", one is written as a scalar and several as a D2 array.
  const assigned = $derived(attrs.class ? attrs.class.split(';') : [])
  const available = $derived((app.render?.classes ?? []).map((c) => c.name).filter((n) => !assigned.includes(n)))
  const setClasses = (list: string[]) => setVal('class', list.length > 1 ? `[${list.join('; ')}]` : (list[0] ?? ''))

  const set = (attr: string, e: Event) => setVal(attr, (e.currentTarget as HTMLInputElement | HTMLSelectElement).value)

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

<div class="form">
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
      {#if shape || edge}
        <div class="classes" role="group" aria-label="Classes">
          <span class="lbl">Classes</span>
          {#each assigned as c (c)}
            <span class="chip">{c}<button aria-label="Remove class {c}" onclick={() => setClasses(assigned.filter((x) => x !== c))}>×</button></span>
          {/each}
          {#if available.length}
            <select aria-label="Add class" value="" onchange={(e) => setClasses([...assigned, e.currentTarget.value])}>
              <option value="" disabled>Add class</option>
              {#each available as n (n)}<option value={n}>{n}</option>{/each}
            </select>
          {:else if !assigned.length}
            <span class="none">none defined, see the Classes tab</span>
          {/if}
        </div>
      {/if}
      {#if shape && !many.length}
        <label title={shape.pinnable ? '' : 'Not pinnable here'}>
          <input type="checkbox" checked={shape.pinned} disabled={!shape.pinnable} onchange={pin} /> Pinned
        </label>
      {/if}
      {#each fields as fd (fd.attr)}
        <FieldInput field={fd} value={attrs[fd.attr]} placeholder={mixed(fd.attr) ? 'mixed' : fd.kind === 'color' ? 'unset' : undefined} onset={(v) => setVal(fd.attr, v)} />
      {/each}
      <button class="danger" onclick={del}>Delete</button>
    {/key}
  {/if}
</div>

<style>
  p { color: var(--muted); margin: 0 0 12px; }
  button { margin-top: 8px; }
  .classes { display: flex; flex-wrap: wrap; align-items: center; gap: 4px; margin-bottom: 10px; }
  .lbl { width: 100%; font-size: 12px; font-weight: 500; color: var(--muted); }
  .chip { display: inline-flex; align-items: center; gap: 2px; padding: 1px 2px 1px 8px; border-radius: 999px; background: var(--secondary-soft); color: var(--secondary); font-size: 12px; }
  .chip button { margin: 0; padding: 0 5px; border: 0; background: none; color: inherit; line-height: 1.4; }
  .none { font-size: 12px; color: var(--muted); }
  .classes select { margin-top: 0; }
</style>
