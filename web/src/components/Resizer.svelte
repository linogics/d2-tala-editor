<script lang="ts">
  // A drag handle on the left edge of a column. Dragging left widens it, double click resets it.
  let { width = $bindable(), min, max, label }: { width: number | null; min: number; max: () => number; label: string } = $props()

  let el: HTMLDivElement
  let drag: { x: number; w: number } | null = null

  // The column right of the handle, measured so a default % width starts from its real size.
  const current = () => (el.nextElementSibling as HTMLElement).getBoundingClientRect().width
  const clamp = (w: number) => Math.round(Math.max(min, Math.min(max(), w)))

  function down(e: PointerEvent) {
    if (e.button !== 0) return
    el.setPointerCapture(e.pointerId)
    drag = { x: e.clientX, w: current() }
  }

  function move(e: PointerEvent) {
    if (drag) width = clamp(drag.w - (e.clientX - drag.x))
  }

  function key(e: KeyboardEvent) {
    const d = { ArrowLeft: 16, ArrowRight: -16 }[e.key]
    if (!d) return
    e.preventDefault()
    width = clamp(current() + d)
  }
</script>

<!-- A focusable separator is an interactive ARIA widget, Svelte does not know that role. -->
<!-- svelte-ignore a11y_no_noninteractive_tabindex, a11y_no_noninteractive_element_interactions -->
<div
  bind:this={el}
  class="resizer"
  role="separator"
  aria-orientation="vertical"
  aria-label={label}
  aria-valuenow={width ?? undefined}
  aria-valuemin={min}
  title="Drag to resize, double click to reset"
  tabindex="0"
  onpointerdown={down}
  onpointermove={move}
  onpointerup={() => (drag = null)}
  ondblclick={() => (width = null)}
  onkeydown={key}
></div>

<style>
  .resizer { width: 6px; margin-left: -3px; z-index: 1; cursor: col-resize; touch-action: none; }
  .resizer:hover, .resizer:focus-visible { background: var(--primary-soft); outline: none; }
</style>
