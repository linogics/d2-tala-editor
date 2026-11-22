<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import type { EdgeInfo, RenderResult, ShapeInfo } from '../lib/protocol'
  import { hitTest, screenToDiagram, snapTo } from '../lib/coords'
  import { templates } from '../lib/templates'
  import { resolveConnect, resolveEdgeDrop, resolveShapeDrop, type DragIntent } from '../lib/dragIntent'
  import { clampDelta, enclosed, movers } from '../lib/group'

  let { state: app }: { state: AppState } = $props()

  type Ghost = { id: string; x: number; y: number; w: number; h: number }

  let box: HTMLDivElement
  let overlay: SVGSVGElement | undefined = $state()
  let view = $state({ x: 20, y: 20, k: 1 })
  let ghosts: Ghost[] = $state([])
  // Set after a drop, the ghosts stay until a render other than this one arrives.
  let dropRender: RenderResult | null = null
  // Space held turns a left drag into a pan.
  let space = $state(false)

  // D2 nests the diagram svg inside a pixel-space outer svg, its viewBox holds the diagram coords.
  const vb = $derived(app.render?.svg.match(/<svg[^>]*d2-svg[^>]*viewBox="([-\d.\s]+)"/)?.[1].trim().split(/\s+/).map(Number))
  const size = $derived(vb ? { w: vb[2], h: vb[3] } : { w: 0, h: 0 })

  $effect(() => {
    if (app.render !== dropRender) {
      dropRender = null
      ghosts = []
    }
  })

  // Zooms to k and keeps the box point (cx, cy) still. Defaults to the box center.
  function zoomTo(k: number, cx = box.clientWidth / 2, cy = box.clientHeight / 2) {
    k = Math.min(8, Math.max(0.1, k))
    view = { x: cx - ((cx - view.x) * k) / view.k, y: cy - ((cy - view.y) * k) / view.k, k }
  }

  // Centers the diagram in the box, scaled down to fit but never above 100%.
  function fit() {
    if (!size.w || !size.h) return
    const pad = 32
    const w = box.clientWidth
    const h = box.clientHeight
    const k = Math.min(1, Math.max(0.1, Math.min((w - 2 * pad) / size.w, (h - 2 * pad) / size.h)))
    view = { x: (w - size.w * k) / 2, y: (h - size.h * k) / 2, k }
  }

  // The first diagram opens fitted, later renders keep the view the user set.
  let fitted = false
  $effect(() => {
    if (vb && !fitted) {
      fitted = true
      fit()
    }
  })

  $effect(() => {
    const el = box
    const onWheel = (e: WheelEvent) => {
      e.preventDefault()
      const r = el.getBoundingClientRect()
      zoomTo(view.k * Math.exp(-e.deltaY * 0.002), e.clientX - r.left, e.clientY - r.top)
    }
    el.addEventListener('wheel', onWheel, { passive: false })

    return () => el.removeEventListener('wheel', onWheel)
  })

  type Drag =
    | { kind: 'pan'; sx: number; sy: number; vx: number; vy: number }
    | { kind: 'shape'; shape: ShapeInfo; start: Pt; moved: boolean }
    | { kind: 'group'; shape: ShapeInfo; group: ShapeInfo[]; start: Pt; moved: boolean }
    | { kind: 'box'; add: boolean; start: Pt; moved: boolean }
    | { kind: 'edge'; edge: EdgeInfo; start: Pt; moved: boolean }
    | { kind: 'connect'; shape: ShapeInfo; start: Pt; moved: boolean }
  type Pt = { x: number; y: number }
  let drag: Drag | null = $state(null)
  // Live drop preview, null when no drag is active.
  let live: DragIntent | null = $state(null)
  let cur: Pt | null = $state(null)

  const shapes = $derived(app.render?.shapes ?? [])
  const byId = (id: string) => shapes.find((s) => s.id === id)
  const selIds = $derived(app.selection.filter((s) => s.kind === 'shape').map((s) => s.id))
  const selShape = $derived(app.selection.length === 1 && selIds.length === 1 ? byId(selIds[0]) : undefined)

  // Messages of a sequence diagram are the edges whose ends share a sequence parent.
  function seqMessages(ed: EdgeInfo): EdgeInfo[] | null {
    const par = (id: string) => byId(id)?.parent
    const p = par(ed.src)
    if (p === undefined || p !== par(ed.dst) || byId(p)?.container !== 'sequence') return null

    return (app.render?.edges ?? []).filter((e) => par(e.src) === p && par(e.dst) === p)
  }

  // Where the dragged shape's corner lands, on the grid when the magnet is on.
  function ghostAt(d: { shape: ShapeInfo; start: Pt }, p: Pt): Pt {
    const x = d.shape.x + p.x - d.start.x
    const y = d.shape.y + p.y - d.start.y

    return app.snap ? { x: snapTo(x, app.grid), y: snapTo(y, app.grid) } : { x, y }
  }

  // The grabbed shape snaps, the others keep their offset, and the group stops at 0 as a whole.
  function groupGhosts(d: Extract<Drag, { kind: 'group' }>, p: Pt): Ghost[] {
    const at = ghostAt(d, p)
    const dd = clampDelta(d.group, { x: at.x - d.shape.x, y: at.y - d.shape.y })

    return d.group.map((s) => ({ id: s.id, x: s.x + dd.x, y: s.y + dd.y, w: s.w, h: s.h }))
  }

  function intentFor(d: Drag, p: Pt): DragIntent {
    if (d.kind === 'shape') return resolveShapeDrop(shapes, d.shape, p, ghostAt(d, p))
    if (d.kind === 'edge') return resolveEdgeDrop(seqMessages(d.edge) ?? [], d.edge, p)
    if (d.kind === 'connect') return resolveConnect(shapes, d.shape.id, p)

    return { kind: 'none' }
  }

  function down(e: PointerEvent) {
    if (ghosts.length) return
    if (e.button === 1 || (e.button === 0 && space)) {
      e.preventDefault()
      box.setPointerCapture(e.pointerId)
      drag = { kind: 'pan', sx: e.clientX, sy: e.clientY, vx: view.x, vy: view.y }

      return
    }
    if (e.button !== 0 || !overlay) return
    const p = screenToDiagram(overlay, e.clientX, e.clientY)
    const shape = hitTest(shapes, p)
    const mod = e.shiftKey || e.ctrlKey || e.metaKey
    if (shape && mod) return app.toggle(shape.id)
    box.setPointerCapture(e.pointerId)
    if (!shape) {
      drag = { kind: 'box', add: mod, start: p, moved: false }
    } else if (selIds.length > 1 && selIds.includes(shape.id)) {
      drag = { kind: 'group', shape, group: movers(shapes, selIds), start: p, moved: false }
    } else {
      app.select({ kind: 'shape', id: shape.id })
      drag = { kind: 'shape', shape, start: p, moved: false }
    }
  }

  function downHandle(e: PointerEvent, shape: ShapeInfo) {
    if (e.button !== 0 || ghosts.length || space) return
    e.stopPropagation()
    box.setPointerCapture(e.pointerId)
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
      ghosts = [{ id: shape.id, ...ghostAt(drag, p), w: shape.w, h: shape.h }]
    } else if (drag.kind === 'group') {
      ghosts = groupGhosts(drag, p)
    }
  }

  async function up(e: PointerEvent) {
    const d = drag
    const end = cur
    drag = null
    live = null
    cur = null
    if (!d || d.kind === 'pan') return
    if (d.kind === 'box') {
      if (!d.moved || !end) return void (d.add || app.select(null))
      const ids = enclosed(shapes, d.start, end)
      app.selectShapes(d.add ? [...new Set([...selIds, ...ids])] : ids)

      return
    }
    if (d.kind === 'group') {
      if (!d.moved) return app.select({ kind: 'shape', id: d.shape.id })
      if (await pinGhosts()) dropRender = app.render
      else ghosts = []

      return
    }
    if (!d.moved) return
    const intent = intentFor(d, screenToDiagram(overlay!, e.clientX, e.clientY))
    const g = ghosts[0] ?? null
    const r = await run(intent, g, e.altKey)
    if (r && g) dropRender = app.render
    else ghosts = []
  }

  const rnd = (n: number) => Math.max(0, Math.round(n))

  // Pins every ghost where it is, as one undo step.
  function pinGhosts(): Promise<boolean> {
    return app.applyEdits(ghosts.map((g) => ['pin', { id: g.id, x: rnd(g.x), y: rnd(g.y) }]))
  }

  // Alt held at release connects with a plain line instead of an arrow.
  async function run(i: DragIntent, g: Ghost | null, alt = false): Promise<unknown> {
    switch (i.kind) {
      case 'pin':
        return app.applyEdit('pin', { id: i.id, x: rnd(i.x), y: rnd(i.y) })
      case 'reorder':
        return app.applyEdit('reorder', i)
      case 'reorderEdge':
        return app.applyEdit('reorderEdge', i)
      case 'connect':
        return app.applyEdit('connect', { src: i.src, dst: i.dst, op: alt ? '--' : '->' })
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
      await app.applyEdit('pin', { id: r.id, x: rnd(p.x), y: rnd(p.y) })
    }
    app.select({ kind: 'shape', id: r.id })
  }

  function downEdge(e: PointerEvent, ed: EdgeInfo) {
    if (space || e.button === 1) return
    e.stopPropagation()
    app.select({ kind: 'edge', id: ed.id })
    if (e.button !== 0 || ghosts.length || !seqMessages(ed)) return
    box.setPointerCapture(e.pointerId)
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

  // Arrow keys move the ghosts right away and pin once the keys pause, so a burst is one layout and one undo step.
  function key(e: KeyboardEvent) {
    const t = e.target as HTMLElement
    if (t.closest?.('input, textarea, select, .cm-editor') || t.isContentEditable) return
    if (e.key === ' ' && !t.closest?.('button')) {
      e.preventDefault()
      space = true

      return
    }
    if (e.key === 'Escape') return app.select(null)
    const mod = e.ctrlKey || e.metaKey
    if (mod && e.key.toLowerCase() === 'a') {
      e.preventDefault()

      return app.selectShapes(shapes.map((s) => s.id))
    }
    // The code editor holds the history, so canvas undo goes through it.
    if (mod && (e.key.toLowerCase() === 'y' || (e.key.toLowerCase() === 'z' && e.shiftKey))) {
      e.preventDefault()

      return app.redo()
    }
    if (mod && e.key.toLowerCase() === 'z') {
      e.preventDefault()

      return app.undo()
    }
    if (!mod && !e.altKey && e.key === '0') return fit()
    if (!mod && !e.altKey && e.key === '1') return zoomTo(1)
    const st = steps[e.key]
    if (!st || drag || e.ctrlKey || e.metaKey || e.altKey) return
    const group = movers(shapes, selIds)
    if (!group.length) return
    e.preventDefault()
    const n = e.shiftKey ? app.grid : 1
    const base = group.map((s) => ghosts.find((g) => g.id === s.id) ?? s)
    const d = clampDelta(base, { x: st[0] * n, y: st[1] * n })
    ghosts = base.map((b) => ({ id: b.id, x: b.x + d.x, y: b.y + d.y, w: b.w, h: b.h }))
    clearTimeout(nudgeTimer)
    nudgeTimer = setTimeout(async () => {
      if (!ghosts.length) return
      if (await pinGhosts()) dropRender = app.render
      else ghosts = []
    }, 300)
  }

  $effect(() => () => clearTimeout(nudgeTimer))

  const isSel = (kind: string, id: string) => app.selection.some((s) => s.kind === kind && s.id === id)
</script>

<svelte:window onkeydown={key} onkeyup={(e) => e.key === ' ' && (space = false)} onblur={() => (space = false)} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="box"
  class:grid={app.snap}
  class:panning={space || drag?.kind === 'pan'}
  style:--g="{app.grid * view.k}px"
  style:--gx="{view.x - (vb?.[0] ?? 0) * view.k}px"
  style:--gy="{view.y - (vb?.[1] ?? 0) * view.k}px"
  bind:this={box}
  ondragover={(e) => e.preventDefault()}
  ondrop={dropTemplate}
  onpointerdown={down}
  onpointermove={move}
  onpointerup={up}
  onpointercancel={up}
>
  <div class="stage" style:width="{size.w}px" style:height="{size.h}px" style:transform="translate({view.x}px, {view.y}px) scale({view.k})">
    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
    {@html app.render?.svg ?? ''}
    {#if vb}
      <svg class="overlay" bind:this={overlay} viewBox={vb.join(' ')}>
        {#if app.snap}
          <defs>
            <pattern id="grid" width={app.grid} height={app.grid} patternUnits="userSpaceOnUse">
              <path d="M{app.grid} 0H0V{app.grid}" fill="none" stroke="var(--grid-line)" stroke-width={1 / view.k} shape-rendering="crispEdges" />
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
        {#if selShape && !ghosts.length}
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
        {#if drag?.kind === 'box' && cur}
          <rect
            class="marquee"
            x={Math.min(drag.start.x, cur.x)}
            y={Math.min(drag.start.y, cur.y)}
            width={Math.abs(cur.x - drag.start.x)}
            height={Math.abs(cur.y - drag.start.y)}
            stroke-width={1 / view.k}
          />
        {/if}
        {#each ghosts as g (g.id)}
          <rect class="ghost" x={g.x} y={g.y} width={g.w} height={g.h} />
        {/each}
      </svg>
    {/if}
  </div>
  {#if vb}
    <!-- Diagram top 0 and left 0, shapes can not be pinned beyond these lines. -->
    <div class="axis v"></div>
    <div class="axis h"></div>
  {/if}
  <!-- Controls sit on the canvas, so a press on them must not start a canvas drag. -->
  <div class="zoom" role="group" aria-label="Zoom" onpointerdown={(e) => e.stopPropagation()}>
    <button class="icon" title="Fit diagram (0)" aria-label="Fit diagram" onclick={fit}>
      <svg viewBox="0 0 24 24" width="15" height="15" aria-hidden="true"><path d="M4 9V4h5M15 4h5v5M20 15v5h-5M9 20H4v-5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" /></svg>
    </button>
    <button class="pct" title="Zoom to 100% (1)" aria-label="Zoom {Math.round(view.k * 100)}%, reset to 100%" onclick={() => zoomTo(1)}>{Math.round(view.k * 100)}%</button>
  </div>
  {#if app.busy}<div class="badge" role="status">Laying out…</div>{/if}
  {#if app.errors.length}
    {@const err = app.errors[0]}
    <div role="alert">
      <button class="banner" title="Show in code" onpointerdown={(e) => e.stopPropagation()} onclick={() => app.reveal(err.line, err.col)}>
        Line {err.line + 1}: {err.message.replace(/^\d+:\d+:\s*/, '')}{#if app.errors.length > 1}<span class="more">+{app.errors.length - 1} more</span>{/if}
      </button>
    </div>
  {/if}
</div>

<style>
  .box {
    position: relative;
    height: 100%;
    overflow: hidden;
    border-right: 1px solid var(--border);
    touch-action: none;
    user-select: none;
    background-color: var(--canvas);
    background-image: radial-gradient(var(--canvas-dot) 1px, transparent 1.4px);
    background-size: 18px 18px;
  }
  .panning, .panning :global(*) { cursor: grab !important; }
  .grid {
    background-image: linear-gradient(to right, var(--grid-line) 1px, transparent 1px), linear-gradient(to bottom, var(--grid-line) 1px, transparent 1px);
    background-size: var(--g) var(--g);
    background-position: var(--gx) var(--gy);
  }
  .axis { position: absolute; pointer-events: none; border: 0 solid var(--secondary); }
  .axis.v { top: 0; bottom: 0; left: var(--gx); border-left-width: 1px; }
  .axis.h { left: 0; right: 0; top: var(--gy); border-top-width: 1px; }
  .gridfill { fill: url(#grid); pointer-events: none; }
  .stage { position: absolute; transform-origin: 0 0; box-shadow: var(--shadow); }
  .stage :global(> svg:not(.overlay)) { width: 100%; height: 100%; display: block; }
  .overlay { position: absolute; inset: 0; width: 100%; height: 100%; overflow: visible; }
  rect, polyline { fill: transparent; }
  polyline { stroke: transparent; stroke-width: 12; fill: none; cursor: pointer; }
  .grab { cursor: grab; }
  .sel { stroke: var(--secondary); stroke-width: 2; }
  polyline.sel { stroke: var(--secondary); stroke-opacity: 0.35; }
  .target { stroke: var(--primary); stroke-width: 3; fill: var(--primary-soft); }
  .handle { fill: var(--secondary); stroke: var(--surface); stroke-width: 1.5; cursor: crosshair; }
  .marker { stroke: var(--primary); stroke-width: 3; pointer-events: none; }
  .marquee { fill: var(--secondary-soft); stroke: var(--secondary); pointer-events: none; }
  .ghost { fill: var(--secondary-soft); stroke: var(--secondary); stroke-dasharray: 4 3; stroke-width: 2; pointer-events: none; }
  .zoom { position: absolute; left: 10px; bottom: 10px; display: flex; gap: 4px; }
  .zoom button { background: var(--surface); box-shadow: var(--shadow); }
  .pct { min-width: 4.5em; justify-content: center; font-family: var(--mono); font-size: 0.8125rem; font-variant-numeric: tabular-nums; }
  .badge { position: absolute; bottom: 10px; right: 10px; padding: 3px 10px; background: var(--warn-bg); color: var(--warn-fg); border-radius: 999px; font-size: 0.8125rem; box-shadow: var(--shadow); }
  .banner { position: absolute; left: 10px; right: 10px; top: 10px; display: flex; justify-content: space-between; gap: 12px; text-align: left; padding: 8px 12px; background: var(--danger-bg); color: var(--danger); border: 1px solid var(--danger); border-radius: var(--radius); box-shadow: var(--shadow); font-family: var(--mono); font-size: 0.8125rem; }
  .banner:hover { border-color: var(--danger); text-decoration: underline; }
  .more { flex: none; opacity: 0.8; }
</style>
