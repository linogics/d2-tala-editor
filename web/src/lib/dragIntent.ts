import { hitTest } from './coords'
import type { EdgeInfo, ShapeInfo } from './protocol'

interface Pt {
  x: number
  y: number
}

export type DragIntent =
  | { kind: 'pin'; id: string; x: number; y: number }
  | { kind: 'move'; id: string; parent: string }
  | { kind: 'reorder'; id: string; before: string }
  | { kind: 'reorderEdge'; id: string; before: string }
  | { kind: 'connect'; src: string; dst: string }
  | { kind: 'none' }

const inside = (s: ShapeInfo, p: Pt) => p.x >= s.x && p.x <= s.x + s.w && p.y >= s.y && p.y <= s.y + s.h

// True when sibling s comes after the drop point in reading order.
function after(s: ShapeInfo, p: Pt, grid: boolean): boolean {
  const cx = s.x + s.w / 2
  const cy = s.y + s.h / 2
  if (!grid) return cx > p.x
  if (Math.abs(cy - p.y) <= s.h / 2) return cx > p.x

  return cy > p.y
}

export function resolveShapeDrop(shapes: ShapeInfo[], dragged: ShapeInfo, drop: Pt, ghost: Pt): DragIntent {
  const parent = shapes.find((s) => s.id === dragged.parent)
  if (parent && (parent.container === 'grid' || parent.container === 'sequence') && inside(parent, drop)) {
    const sib = shapes.find((s) => s.parent === parent.id && s.id !== dragged.id && after(s, drop, parent.container === 'grid'))

    return { kind: 'reorder', id: dragged.id, before: sib?.id ?? '' }
  }

  const others = shapes.filter((s) => s.id !== dragged.id && !s.id.startsWith(dragged.id + '.'))
  const target = hitTest(others, drop)
  if ((target?.id ?? '') !== dragged.parent) return { kind: 'move', id: dragged.id, parent: target?.id ?? '' }
  if (dragged.pinnable) return { kind: 'pin', id: dragged.id, x: ghost.x, y: ghost.y }

  return { kind: 'none' }
}

const midY = (e: EdgeInfo) => (e.points[0][1] + e.points[e.points.length - 1][1]) / 2

// Callers pass only the messages of one sequence diagram, in declaration order.
export function resolveEdgeDrop(edges: EdgeInfo[], dragged: EdgeInfo, drop: Pt): DragIntent {
  const next = edges.find((e) => e.id !== dragged.id && midY(e) > drop.y)

  return { kind: 'reorderEdge', id: dragged.id, before: next?.id ?? '' }
}

export function resolveConnect(shapes: ShapeInfo[], src: string, drop: Pt): DragIntent {
  const t = hitTest(shapes, drop)

  return t && t.id !== src ? { kind: 'connect', src, dst: t.id } : { kind: 'none' }
}
