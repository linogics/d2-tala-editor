<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import type { RenderResult, ShapeInfo } from '../lib/protocol'
  import { hitTest, screenToDiagram } from '../lib/coords'

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
    | { kind: 'shape'; shape: ShapeInfo; start: { x: number; y: number }; moved: boolean }
  let drag: Drag | null = null

  function down(e: PointerEvent) {
    if (e.button !== 0 || ghost) return
    const p = screenToDiagram(overlay!, e.clientX, e.clientY)
    const shape = hitTest(app.render?.shapes ?? [], p)
    overlay!.setPointerCapture(e.pointerId)
    if (shape) {
      app.select({ kind: 'shape', id: shape.id })
      drag = shape.pinnable ? { kind: 'shape', shape, start: p, moved: false } : null
    } else {
      drag = { kind: 'pan', sx: e.clientX, sy: e.clientY, vx: view.x, vy: view.y }
    }
  }

  function move(e: PointerEvent) {
    if (!drag) return
    if (drag.kind === 'pan') {
      view = { ...view, x: drag.vx + e.clientX - drag.sx, y: drag.vy + e.clientY - drag.sy }

      return
    }
    const p = screenToDiagram(overlay!, e.clientX, e.clientY)
    const { shape, start } = drag
    // A few pixels of slack keeps plain clicks from becoming drags.
    if (!drag.moved && Math.hypot(p.x - start.x, p.y - start.y) * view.k < 4) return
    drag.moved = true
    ghost = { id: shape.id, x: shape.x + p.x - start.x, y: shape.y + p.y - start.y, w: shape.w, h: shape.h }
  }

  async function up() {
    const d = drag
    drag = null
    if (d?.kind === 'shape' && !d.moved) return
    if (d?.kind !== 'shape' || !ghost) return
    const g = ghost
    const r = await app.applyEdit('pin', { id: g.id, x: Math.max(0, Math.round(g.x)), y: Math.max(0, Math.round(g.y)) })
    if (r) dropRender = app.render
    else ghost = null
  }

  function selectEdge(e: PointerEvent, id: string) {
    e.stopPropagation()
    app.select({ kind: 'edge', id })
  }

  const isSel = (kind: string, id: string) => app.selection?.kind === kind && app.selection.id === id
</script>

<div class="box" bind:this={box}>
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
        {#each app.render?.shapes ?? [] as s (s.id)}
          <rect
            data-shape={s.id}
            x={s.x} y={s.y} width={s.w} height={s.h}
            class:sel={isSel('shape', s.id)}
            class:grab={s.pinnable}
          />
        {/each}
        {#each app.render?.edges ?? [] as ed (ed.id)}
          <polyline
            data-edge={ed.id}
            points={ed.points.map((p) => `${p.x},${p.y}`).join(' ')}
            class:sel={isSel('edge', ed.id)}
            onpointerdown={(e) => selectEdge(e, ed.id)}
          />
        {/each}
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
  .stage { position: absolute; transform-origin: 0 0; }
  .stage :global(> svg:not(.overlay)) { width: 100%; height: 100%; display: block; }
  .overlay { position: absolute; inset: 0; width: 100%; height: 100%; }
  rect, polyline { fill: transparent; }
  polyline { stroke: transparent; stroke-width: 12; fill: none; cursor: pointer; }
  .grab { cursor: grab; }
  .sel { stroke: #2563eb; stroke-width: 2; }
  polyline.sel { stroke: #2563eb55; }
  .ghost { fill: #2563eb22; stroke: #2563eb; stroke-dasharray: 4 3; stroke-width: 2; pointer-events: none; }
  .badge { position: absolute; top: 8px; right: 8px; padding: 2px 8px; background: #fef3c7; border-radius: 4px; font: 12px sans-serif; }
  .banner { position: absolute; left: 0; right: 0; top: 0; padding: 6px 10px; background: #fee2e2; color: #991b1b; font: 13px sans-serif; }
</style>
