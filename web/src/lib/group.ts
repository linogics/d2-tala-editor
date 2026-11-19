import type { ShapeInfo } from './protocol'

interface Pt {
  x: number
  y: number
}

// Selected ids without those whose ancestor is selected too, since they follow it.
export function roots(ids: string[]): string[] {
  return ids.filter((id) => !ids.some((o) => o !== id && id.startsWith(o + '.')))
}

// Shapes a group drag pins: pinnable roots of the selection.
export function movers(shapes: ShapeInfo[], ids: string[]): ShapeInfo[] {
  const r = new Set(roots(ids))

  return shapes.filter((s) => r.has(s.id) && s.pinnable)
}

// Limits a group offset so no member goes left of or above 0.
export function clampDelta(group: Pt[], d: Pt): Pt {
  if (!group.length) return d
  const minX = Math.min(...group.map((s) => s.x))
  const minY = Math.min(...group.map((s) => s.y))

  return { x: Math.max(d.x, -minX), y: Math.max(d.y, -minY) }
}

// Ids of shapes lying fully inside the box spanned by a and b.
export function enclosed(shapes: ShapeInfo[], a: Pt, b: Pt): string[] {
  const x0 = Math.min(a.x, b.x)
  const x1 = Math.max(a.x, b.x)
  const y0 = Math.min(a.y, b.y)
  const y1 = Math.max(a.y, b.y)

  return shapes.filter((s) => s.x >= x0 && s.y >= y0 && s.x + s.w <= x1 && s.y + s.h <= y1).map((s) => s.id)
}
