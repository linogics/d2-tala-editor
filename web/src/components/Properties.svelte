<script lang="ts">
  import type { AppState, Step } from '../lib/state.svelte'
  import { templates } from '../lib/templates'
  import { edgeFields, enterCommits, fieldSteps, groupsOf, headFields, nears, shapeFieldsFor, type Field } from '../lib/fields'
  import FieldInput from './FieldInput.svelte'
  import Group from './Group.svelte'

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
          ? edgeFields.filter((x) => !x.attr.includes('-arrowhead.'))
          : [],
  )
  // D2 refuses near on the shape itself, its ancestors and its descendants.
  const related = (a: string, b: string) => a === b || a.startsWith(b + '.') || b.startsWith(a + '.')
  const nearTargets = $derived((app.render?.shapes ?? []).map((s) => s.id).filter((id) => !ids.some((x) => related(x, id))))
  const withTargets = (fd: Field) => (fd.attr === 'near' ? { ...fd, options: [...nears, ...nearTargets] } : fd)
  const type = $derived(many.length ? (many.every((s) => s.type === many[0].type) ? many[0].type : '') : shape?.type)
  const types = $derived(type && !templates.some((t) => t.key === type) ? [type, ...templates.map((t) => t.key)] : templates.map((t) => t.key))
  const ids = $derived(app.selection.map((x) => x.id))
  // Members disagree on this attr.
  const mixed = (attr: string) => many.some((s) => s.attrs[attr] !== many[0].attrs[attr])

  const gridAttrs = ['grid-rows', 'grid-columns']
  // The Grid group opens on grids and stays folded elsewhere.
  const isGrid = $derived(gridAttrs.some((a) => attrs[a] != null))

  const setVal = (attr: string, value: string) => {
    const v = value.trim() || null
    if (shape && !many.length && (attr === 'top' || attr === 'left')) return app.applyEdits(fieldSteps(shape, attr, v))
    // D2 refuses top and left on grid cells, so making a grid unpins its direct children in the same step.
    const unpin = (id: string): Step[] =>
      v && gridAttrs.includes(attr) ? (app.render?.shapes ?? []).filter((s) => s.parent === id && s.pinned).map((s) => ['unpin', { id: s.id }]) : []
    const steps = ids.flatMap((id): Step[] => [...unpin(id), ['set', { id, attr, value: v }]])
    if (steps.length === 1) return app.applyEdit('set', { id: ids[0], attr, value: v })

    return app.applyEdits(steps)
  }
  // Classes come back as "a;b", one is written as a scalar and several as a D2 array.
  const assigned = $derived(attrs.class ? attrs.class.split(';') : [])
  const available = $derived((app.render?.classes ?? []).map((c) => c.name).filter((n) => !assigned.includes(n)))
  const setClasses = (list: string[]) => setVal('class', list.length > 1 ? `[${list.join('; ')}]` : (list[0] ?? ''))

  const set = (attr: string, e: Event) => setVal(attr, (e.currentTarget as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement).value)

  // An end without a head draws no arrowhead shape, its label still shows.
  const ends = $derived(edge ? [{ end: 'source', name: 'Source', on: edge.srcArrow }, { end: 'target', name: 'Target', on: edge.dstArrow }] : [])

  // The heads pick the op, so the edge gets a new id.
  async function heads(el: HTMLInputElement, src: boolean, dst: boolean) {
    const r = await app.applyEdit('setArrows', { id: edge!.id, src, dst })
    if (r) app.select({ kind: 'edge', id: r.id })
    else el.checked = !el.checked
  }

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

  // A container goes with everything inside it. keep (Shift) lifts its children out instead.
  async function del(keep = false) {
    // Deepest first, so with keep a deleted container does not lift its selected children out.
    const order = keep
      ? [...ids].sort((a, b) => b.split('.').length - a.split('.').length)
      : ids.filter((id) => !ids.some((p) => id.startsWith(p + '.')))
    if (await app.applyEdits(order.map((id) => ['delete', { id, keepChildren: keep }]))) app.select(null)
  }

  function key(e: KeyboardEvent) {
    if (e.key !== 'Delete' || !ids.length) return
    const t = e.target as HTMLElement
    if (t.closest?.('input, textarea, select, .cm-editor') || t.isContentEditable) return
    del(e.shiftKey)
  }
</script>

<svelte:window onkeydown={key} />


<div class="form">
  {#if !ids.length}
    <p>Drag shapes from the left, or type in the code pane</p>
    <button class="primary" onclick={() => app.applyEdit('pinAll', {})}>Pin all</button>
  {:else}
    {#key ids.join('\n')}
      <Group name="General">
        {#if many.length}
          <p>{many.length} shapes selected</p>
        {:else}
          {#if shape}
            <label>Id <input value={sel!.id.split('.').pop()} onchange={rename} /></label>
          {/if}
          <label>Label <textarea rows="1" value={attrs.label ?? ''} onkeydown={enterCommits} onchange={(e) => set('label', e)}></textarea></label>
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
      </Group>
      {#if edge}
        <Group name="Arrows" plain>
          {#each ends as { end, name, on } (end)}
            <div class="head" role="group" aria-label="{name} head">
              <label>
                <input type="checkbox" checked={on} onchange={(e) => heads(e.currentTarget, end === 'source' ? e.currentTarget.checked : edge.srcArrow, end === 'target' ? e.currentTarget.checked : edge.dstArrow)} />
                {end} head
              </label>
              {#each headFields(end) as fd (fd.attr)}
                <FieldInput field={fd} value={attrs[fd.attr]} disabled={fd.kind === 'select' && !on} onset={(v) => setVal(fd.attr, v)} />
              {/each}
            </div>
          {/each}
        </Group>
      {/if}
      {#each groupsOf(fields) as g (g)}
        <Group name={g} open={g === 'Grid' ? isGrid : undefined}>
          {#if g === 'Position' && shape && !many.length}
            <label title={shape.pinnable ? '' : 'Not pinnable here'}>
              <input type="checkbox" checked={shape.pinned} disabled={!shape.pinnable} onchange={pin} /> Pinned
            </label>
          {/if}
          {#each fields.filter((x) => x.group === g) as fd (fd.attr)}
            <FieldInput field={withTargets(fd)} value={attrs[fd.attr]} placeholder={mixed(fd.attr) ? 'mixed' : fd.kind === 'color' ? 'unset' : undefined} onset={(v) => setVal(fd.attr, v)} />
          {/each}
        </Group>
      {/each}
      <button class="danger" title="Shift keeps the children" onclick={(e) => del(e.shiftKey)}>Delete</button>
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
