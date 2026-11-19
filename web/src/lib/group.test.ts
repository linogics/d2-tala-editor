import { describe, expect, it } from 'vitest'
import { clampDelta, enclosed, movers, roots } from './group'
import type { ShapeInfo } from './protocol'

const sh = (id: string, x: number, y: number, over: Partial<ShapeInfo> = {}): ShapeInfo => ({
  id, parent: '', type: 'rectangle', container: '', x, y, w: 10, h: 10, pinned: false, pinnable: true, attrs: {}, ...over,
})

describe('group helpers', () => {
  it('roots drops descendants of selected containers', () => {
    expect(roots(['a', 'a.b', 'ab', 'c.d'])).toEqual(['a', 'ab', 'c.d'])
  })

  it('movers keeps pinnable roots only', () => {
    const shapes = [sh('a', 0, 0), sh('a.b', 1, 1), sh('g', 0, 0, { pinnable: false })]
    expect(movers(shapes, ['a', 'a.b', 'g']).map((s) => s.id)).toEqual(['a'])
  })

  it('clampDelta stops the whole group at 0', () => {
    expect(clampDelta([sh('a', 5, 30), sh('b', 40, 8)], { x: -20, y: -20 })).toEqual({ x: -5, y: -8 })
    expect(clampDelta([sh('a', 5, 30)], { x: 3, y: -4 })).toEqual({ x: 3, y: -4 })
  })

  it('enclosed needs the full shape inside the box, any drag direction', () => {
    const shapes = [sh('in', 10, 10), sh('half', 25, 10)]
    expect(enclosed(shapes, { x: 30, y: 30 }, { x: 5, y: 5 })).toEqual(['in'])
  })
})
