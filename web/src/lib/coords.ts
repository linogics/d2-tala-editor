import type { ShapeInfo } from './protocol'

interface Pt {
  x: number
  y: number
}

// Maps client pixels into the user space of the given svg element.
export function screenToDiagram(svg: SVGSVGElement, clientX: number, clientY: number): Pt {
  const m = svg.getScreenCTM()!.inverse()

  return { x: m.a * clientX + m.c * clientY + m.e, y: m.b * clientX + m.d * clientY + m.f }
}

// Smallest shape containing p wins, so children beat their containers.
export function hitTest(shapes: ShapeInfo[], p: Pt): ShapeInfo | null {
  let best: ShapeInfo | null = null
  for (const s of shapes) {
    if (p.x < s.x || p.x > s.x + s.w || p.y < s.y || p.y > s.y + s.h) continue
    if (!best || s.w * s.h < best.w * best.h) best = s
  }

  return best
}
