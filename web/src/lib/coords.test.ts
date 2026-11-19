import { describe, expect, it } from 'vitest'
import { hitTest, screenToDiagram, snapTo } from './coords'
import type { ShapeInfo } from './protocol'

const sh = (id: string, x: number, y: number, w: number, h: number): ShapeInfo => ({
  id, parent: '', type: 'rectangle', container: "", x, y, w, h, pinned: false, pinnable: true, attrs: {},
})

describe('coords', () => {
  const shapes = [sh('outer', 0, 0, 100, 100), sh('inner', 10, 10, 20, 20)]

  it('hitTest returns deepest shape', () => {
    expect(hitTest(shapes, { x: 15, y: 15 })?.id).toBe('inner')
    expect(hitTest(shapes, { x: 50, y: 50 })?.id).toBe('outer')
  })

  it('hitTest returns null on empty space', () => {
    expect(hitTest(shapes, { x: 200, y: 200 })).toBeNull()
  })

  it('screenToDiagram applies inverse CTM', () => {
    // Forward matrix: scale 2, translate 10. Inverse: x = (cx - 10) / 2.
    const svg = {
      getScreenCTM: () => ({ inverse: () => ({ a: 0.5, b: 0, c: 0, d: 0.5, e: -5, f: -5 }) }),
    } as unknown as SVGSVGElement

    expect(screenToDiagram(svg, 30, 50)).toEqual({ x: 10, y: 20 })
  })
})

describe('snapTo', () => {
  it('rounds to the nearest grid line', () => {
    expect(snapTo(29, 20)).toBe(20)
    expect(snapTo(31, 20)).toBe(40)
    expect(snapTo(0, 7)).toBe(0)
  })
})
