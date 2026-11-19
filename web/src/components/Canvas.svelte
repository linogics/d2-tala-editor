<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import type { EdgeInfo, RenderResult, ShapeInfo } from '../lib/protocol'
  import { hitTest, screenToDiagram, snapTo } from '../lib/coords'
  import { templates } from '../lib/templates'
  import { resolveConnect, resolveEdgeDrop, resolveShapeDrop, type DragIntent } from '../lib/dragIntent'

  let { state: app }: { state: AppState } = $props()

  let box: HTMLDivElement
  let overlay: SVGSVGElement | undefined = $state()
  let view = $state({ x: 20, y: 20, k: 1 })
  let ghost: { id: string; x: number; y: number; w: number; h: number } | null = $state(null)
  // Set after a drop, the ghost stays until a render other than this one arrives.
  let dropRender: RenderResult | null = null

  // D2 nests the diagram svg inside a pixel-space outer svg, its viewBox holds the diagram coords.
  const vb = $derived(app.render?.svg.match(/<svg[^>]*d2-svg[^>]*viewBox="([-\d.\s]+)"/)?.[1].trim().split(/\s+/).map(Number))
  const size = $derived(vb ? { w: vb[2], h: vb[3] } : { w: 0, h: 0 })

  $effect(() => {
    if (app.render !== dropRender) {
      dropRender = null
      ghost = null
    }
  })

  $effect(() => {
    const el = box
    const onWheel = (e: WheelEvent) => {
      e.preventDefault()
      const r = el.getBoundingClientRect()
      const cx = e.clientX - r.left
      const cy = e.clientY - r.top
      const k = Math.min(8, Math.max(0.1, view.k * Math.exp(-e.deltaY * 0.002)))
      view = { x: cx - ((cx - view.x) * k) / view.k, y: cy - ((cy - view.y) * k) / view.k, k }
    }
    el.addEventListener('wheel', onWheel, { passive: false })

    return () => el.removeEventListener('wheel', onWheel)
  })

  type Drag =
    | { kind: 'pan'; sx: number; sy: number; vx: number; vy: number }
    | { kind: 'shape'; shape: ShapeInfo; start: Pt; moved: boolean }
    | { kind: 'edge'; edge: EdgeInfo; start: Pt; moved: boolean }
    | { kind: 'connect'; shape: ShapeInfo; start: Pt; moved: boolean }
  type Pt = { x: number; y: number }
  let drag: Drag | null = $state(null)
  // Live drop preview, null when no drag is active.
  let live: DragIntent | null = $state(null)
  let cur: Pt | null = $state(null)

  const shapes = $derived(app.render?.shapes ?? [])
  const byId = (id: string) => shapes.find((s) => s.id === id)
  const selShape = $derived(app.selection?.kind === 'shape' ? byId(app.selection.id) : undefined)

  // Messages of a sequence diagram are the edges whose ends share a sequence parent.
  function seqMessages(ed: EdgeInfo): EdgeInfo[] | null {
    const par = (id: string) => byId(id)?.parent
    const p = par(ed.src)
    if (p === undefined || p !== par(ed.dst) || byId(p)?.container !== 'sequence') return null

    return (app.render?.edges ?? []).filter((e) => par(e.src) === p && par(e.dst) === p)
  }

  // Where the dragged shape's corner lands, on the grid when the magnet is on.
  function ghostAt(d: Extract<Drag, { kind: 'shape' }>, p: Pt): Pt {
    const x = d.shape.x + p.x - d.start.x
    const y = d.shape.y + p.y - d.start.y

    return app.snap ? { x: snapTo(x, app.grid), y: snapTo(y, app.grid) } : { x, y }
  }

  function intentFor(d: Drag, p: Pt): DragIntent {
    if (d.kind === 'shape') return resolveShapeDrop(shapes, d.shape, p, ghostAt(d, p))
    if (d.kind === 'edge') return resolveEdgeDrop(seqMessages(d.edge) ?? [], d.edge, p)
    if (d.kind === 'connect') return resolveConnect(shapes, d.shape.id, p)

    return { kind: 'none' }
  }

  function down(e: PointerEvent) {
    if (e.button !== 0 || ghost) return
    const p = screenToDiagram(overlay!, e.clientX, e.clientY)
    const shape = hitTest(shapes, p)
    overlay!.setPointerCapture(e.pointerId)
    if (shape) {
      app.select({ kind: 'shape', id: shape.id })
      drag = { kind: 'shape', shape, start: p, moved: false }
    } else {
      drag = { kind: 'pan', sx: e.clientX, sy: e.clientY, vx: view.x, vy: view.y }
    }
  }

  function downHandle(e: PointerEvent, shape: ShapeInfo) {
    if (e.button !== 0 || ghost) return
    e.stopPropagation()
    overlay!.setPointerCapture(e.pointerId)
    drag = { kind: 'connect', shape, start: screenToDiagram(overlay!, e.clientX, e.clientY), moved: false }
  }

  function move(e: PointerEvent) {
    if (!drag) return
    if (drag.kind === 'pan') {
      view = { ...view, x: drag.vx + e.clientX - drag.sx, y: drag.vy + e.clientY - drag.sy }

      return
    }
    const p = screenToDiagram(overlay!, e.clientX, e.clientY)
    const { start } = drag
    // A few pixels of slack keeps plain clicks from becoming drags.
    if (!drag.moved && Math.hypot(p.x - start.x, p.y - start.y) * view.k < 4) return
    drag.moved = true
    cur = p
    live = intentFor(drag, p)
    if (drag.kind === 'shape') {
      const { shape } = drag
      ghost = { id: shape.id, ...ghostAt(drag, p), w: shape.w, h: shape.h }
    }
  }

  async function up(e: PointerEvent) {
    const d = drag
    drag = null
    live = null
    cur = null
    if (!d || d.kind === 'pan' || !d.moved) return
    const intent = intentFor(d, screenToDiagram(overlay!, e.clientX, e.clientY))
    const g = ghost
    const r = await run(intent, g)
    if (r && g) dropRender = app.render
    else ghost = null
  }

  async function run(i: DragIntent, g: typeof ghost): Promise<unknown> {
    const rnd = (n: number) => Math.max(0, Math.round(n))
    switch (i.kind) {
      case 'pin':
        return app.applyEdit('pin', { id: i.id, x: rnd(i.x), y: rnd(i.y) })
      case 'reorder':
        return app.applyEdit('reorder', i)
      case 'reorderEdge':
        return app.applyEdit('reorderEdge', i)
      case 'connect':
        return app.applyEdit('connect', { src: i.src, dst: i.dst })
      case 'move': {
        const r = await app.applyEdit('move', { id: i.id, parent: i.parent })
        if (!r || !g || byId(i.parent)?.container) return r
        // The shape got a new id, pin it where the ghost was dropped.
        const name = i.id.split('.').pop()!
        const id = r.id || (i.parent ? `${i.parent}.${name}` : name)

        return (await app.applyEdit('pin', { id, x: rnd(g.x), y: rnd(g.y) })) ?? r
      }
      default:
        return null
    }
  }

  // Palette drop: insert into the container under the cursor, pin at the drop point, select.
  async function dropTemplate(e: DragEvent) {
    const key = e.dataTransfer?.getData('application/x-d2-template')
    const t = templates.find((x) => x.key === key)
    if (!t || !overlay) return
    e.preventDefault()
    const p = screenToDiagram(overlay, e.clientX, e.clientY)
    const parent = hitTest(shapes, p)
    const r = await app.applyEdit('insert', { parent: parent?.id ?? '', template: t.template })
    if (!r) return
    if (parent?.container !== 'grid' && parent?.container !== 'sequence') {
      await app.applyEdit('pin', { id: r.id, x: Math.max(0, Math.round(p.x)), y: Math.max(0, Math.round(p.y)) })
    }
    app.select({ kind: 'shape', id: r.id })
  }

  function downEdge(e: PointerEvent, ed: EdgeInfo) {
    e.stopPropagation()
    app.select({ kind: 'edge', id: ed.id })
    if (e.button !== 0 || ghost || !seqMessages(ed)) return
    overlay!.setPointerCapture(e.pointerId)
    drag = { kind: 'edge', edge: ed, start: screenToDiagram(overlay!, e.clientX, e.clientY), moved: false }
  }

  // Marker line for a reorder, placed before the target sibling or after the last one.
  function marker(i: Extract<DragIntent, { kind: 'reorder' }>) {
    const sibs = shapes.filter((s) => s.parent === byId(i.id)?.parent && s.id !== i.id)
    const b = byId(i.before)
    const last = sibs[sibs.length - 1]
    const ref = b ?? last
    if (!ref) return null
    const x = b ? ref.x : ref.x + ref.w

    return { x1: x, y1: ref.y, x2: x, y2: ref.y + ref.h }
  }

  const steps: Record<string, [number, number]> = { ArrowLeft: [-1, 0], ArrowRight: [1, 0], ArrowUp: [0, -1], ArrowDown: [0, 1] }
  let nudgeTimer: ReturnType<typeof setTimeout> | undefined

  // Arrow keys move the ghost right away and pin once the keys pause, so a burst is one layout and one undo step.
  function key(e: KeyboardEvent) {
    const st = steps[e.key]
    const t = e.target as HTMLElement
    if (!st || drag || !selShape?.pinnable || e.ctrlKey || e.metaKey || e.altKey) return
    if (t.closest?.('input, textarea, select, .cm-editor') || t.isContentEditable) return
    e.preventDefault()
    const s = selShape
    const n = e.shiftKey ? app.grid : 1
    const base = ghost?.id === s.id ? ghost : s
    ghost = { id: s.id, x: Math.max(0, base.x + st[0] * n), y: Math.max(0, base.y + st[1] * n), w: s.w, h: s.h }
    clearTimeout(nudgeTimer)
    nudgeTimer = setTimeout(async () => {
      const g = ghost
      if (!g) return
      if (await app.applyEdit('pin', { id: g.id, x: Math.round(g.x), y: Math.round(g.y) })) dropRender = app.render
      else ghost = null
    }, 300)
  }

  $effect(() => () => clearTimeout(nudgeTimer))

  const isSel = (kind: string, id: string) => app.selection?.kind === kind && app.selection.id === id
</script>

<svelte:window onkeydown={key} />

<div class="box" class:grid={app.snap} style:--g="{app.grid * view.k}px" style:--gx="{view.x - (vb?.[0] ?? 0) * view.k}px" style:--gy="{view.y - (vb?.[1] ?? 0) * view.k}px" bind:this={box} role="presentation" ondragover={(e) => e.preventDefault()} ondrop={dropTemplate}>
  <div class="stage" style:width="{size.w}px" style:height="{size.h}px" style:transform="translate({view.x}px, {view.y}px) scale({view.k})">
    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
    {@html app.render?.svg ?? ''}
    {#if vb}
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <svg
        class="overlay"
        bind:this={overlay}
        viewBox={vb.join(' ')}
        onpointerdown={down}
        onpointermove={move}
        onpointerup={up}
        onpointercancel={up}
      >
        {#if app.snap}
          <defs>
            <pattern id="grid" width={app.grid} height={app.grid} patternUnits="userSpaceOnUse">
              <path d="M{app.grid} 0H0V{app.grid}" fill="none" stroke="#8886" stroke-width={1 / view.k} shape-rendering="crispEdges" />
            </pattern>
          </defs>
          <rect class="gridfill" x={vb[0]} y={vb[1]} width={vb[2]} height={vb[3]} />
        {/if}
        {#each app.render?.shapes ?? [] as s (s.id)}
          <rect
            data-shape={s.id}
            x={s.x} y={s.y} width={s.w} height={s.h}
            class:sel={isSel('shape', s.id)}
            class:grab={s.pinnable}
            class:target={live?.kind === 'move' && live.parent === s.id}
          />
        {/each}
        {#each app.render?.edges ?? [] as ed (ed.id)}
          <polyline
            data-edge={ed.id}
            points={ed.points.map(([x, y]) => `${x},${y}`).join(' ')}
            class:sel={isSel('edge', ed.id)}
            onpointerdown={(e) => downEdge(e, ed)}
          />
        {/each}
        {#if selShape && !ghost}
          <circle class="handle" cx={selShape.x + selShape.w} cy={selShape.y + selShape.h / 2} r="6" onpointerdown={(e) => downHandle(e, selShape)} />
        {/if}
        {#if live?.kind === 'reorder'}
          {@const m = marker(live)}
          {#if m}<line class="marker" {...m} />{/if}
        {:else if live?.kind === 'reorderEdge' && cur}
          <line class="marker" x1={0} y1={cur.y} x2={size.w} y2={cur.y} />
        {:else if live?.kind !== undefined && drag?.kind === 'connect' && cur}
          <line class="marker" x1={drag.shape.x + drag.shape.w} y1={drag.shape.y + drag.shape.h / 2} x2={cur.x} y2={cur.y} />
        {/if}
        {#if ghost}
          <rect class="ghost" x={ghost.x} y={ghost.y} width={ghost.w} height={ghost.h} />
        {/if}
      </svg>
    {/if}
  </div>
  {#if app.busy}<div class="badge">Laying out…</div>{/if}
  {#if app.errors.length}
    <div class="banner">Line {app.errors[0].line + 1}: {app.errors[0].message}</div>
  {/if}
</div>

<style>
  .box { position: relative; height: 100%; overflow: hidden; border-right: 1px solid #ccc; touch-action: none; user-select: none; }
  .grid {
    background-image: linear-gradient(to right, #8886 1px, transparent 1px), linear-gradient(to bottom, #8886 1px, transparent 1px);
    background-size: var(--g) var(--g);
    background-position: var(--gx) var(--gy);
  }
  .gridfill { fill: url(#grid); pointer-events: none; }
  .stage { position: absolute; transform-origin: 0 0; }
  .stage :global(> svg:not(.overlay)) { width: 100%; height: 100%; display: block; }
  .overlay { position: absolute; inset: 0; width: 100%; height: 100%; }
  rect, polyline { fill: transparent; }
  polyline { stroke: transparent; stroke-width: 12; fill: none; cursor: pointer; }
  .grab { cursor: grab; }
  .sel { stroke: #2563eb; stroke-width: 2; }
  polyline.sel { stroke: #2563eb55; }
  .target { stroke: #16a34a; stroke-width: 3; fill: #16a34a18; }
  .handle { fill: #2563eb; stroke: #fff; stroke-width: 1.5; cursor: crosshair; }
  .marker { stroke: #2563eb; stroke-width: 3; pointer-events: none; }
  .ghost { fill: #2563eb22; stroke: #2563eb; stroke-dasharray: 4 3; stroke-width: 2; pointer-events: none; }
  .badge { position: absolute; top: 8px; right: 8px; padding: 2px 8px; background: #fef3c7; border-radius: 4px; font: 12px sans-serif; }
  .banner { position: absolute; left: 0; right: 0; top: 0; padding: 6px 10px; background: #fee2e2; color: #991b1b; font: 13px sans-serif; }
</style>
