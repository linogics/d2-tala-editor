import { describe, expect, it } from 'vitest'
import { resolveConnect, resolveEdgeDrop, resolveShapeDrop } from './dragIntent'
import type { EdgeInfo, ShapeInfo } from './protocol'

const sh = (id: string, parent: string, container: string, x: number, y: number, w = 40, h = 40, pinnable = true): ShapeInfo => ({
  id, parent, type: 'rectangle', container, x, y, w, h, pinned: false, pinnable, attrs: {},
})
const grid = [
  sh('g', '', 'grid', 0, 0, 200, 100),
  sh('g.a', 'g', '', 0, 0, 100, 50, false),
  sh('g.b', 'g', '', 100, 0, 100, 50, false),
  sh('g.c', 'g', '', 0, 50, 100, 50, false),
  sh('g.d', 'g', '', 100, 50, 100, 50, false),
]
const seq = [
  sh('s', '', 'sequence', 0, 0, 300, 100),
  sh('s.a', 's', '', 0, 0, 60, 30, false),
  sh('s.b', 's', '', 100, 0, 60, 30, false),
  sh('s.c', 's', '', 200, 0, 60, 30, false),
]
const z = { x: 0, y: 0 }

describe('dragIntent', () => {
  it('grid child drag resolves to reorder', () => {
    expect(resolveShapeDrop(grid, grid[1], { x: 20, y: 70 }, z)).toEqual({ kind: 'reorder', id: 'g.a', before: 'g.c' })
  })

  it('grid child drag to last slot gives empty before', () => {
    expect(resolveShapeDrop(grid, grid[1], { x: 190, y: 90 }, z)).toEqual({ kind: 'reorder', id: 'g.a', before: '' })
  })

  it('sequence actor drag resolves to reorder by X', () => {
    expect(resolveShapeDrop(seq, seq[1], { x: 130, y: 10 }, z)).toEqual({ kind: 'reorder', id: 's.a', before: 's.c' })
  })

  it('message drag resolves to reorderEdge', () => {
    const e = (id: string, y: number): EdgeInfo => ({ id, src: 's.a', dst: 's.b', points: [[0, y], [10, y]], attrs: {} })
    const es = [e('m1', 10), e('m2', 20), e('m3', 30)]
    expect(resolveEdgeDrop(es, es[0], { x: 0, y: 25 })).toEqual({ kind: 'reorderEdge', id: 'm1', before: 'm3' })
    expect(resolveEdgeDrop(es, es[0], { x: 0, y: 99 })).toEqual({ kind: 'reorderEdge', id: 'm1', before: '' })
  })

  it('drop into other container resolves to move', () => {
    const s = [sh('g', '', '', 0, 0, 100, 100), sh('y', '', '', 200, 0, 40, 40)]
    expect(resolveShapeDrop(s, s[1], { x: 50, y: 50 }, z)).toEqual({ kind: 'move', id: 'y', parent: 'g' })
  })

  it('drop on empty space from container moves to root', () => {
    const s = [sh('g', '', '', 0, 0, 100, 100), sh('g.x', 'g', '', 10, 10, 20, 20)]
    expect(resolveShapeDrop(s, s[1], { x: 500, y: 500 }, z)).toEqual({ kind: 'move', id: 'g.x', parent: '' })
  })

  it('plain drag resolves to pin', () => {
    const s = [sh('a', '', '', 0, 0, 40, 40)]
    expect(resolveShapeDrop(s, s[0], { x: 300, y: 300 }, { x: 5, y: 6 })).toEqual({ kind: 'pin', id: 'a', x: 5, y: 6 })
  })

  it('connect onto shape', () => {
    const s = [sh('a', '', '', 0, 0), sh('b', '', '', 100, 0)]
    expect(resolveConnect(s, 'a', { x: 110, y: 10 })).toEqual({ kind: 'connect', src: 'a', dst: 'b' })
  })

  it('connect onto empty space is none', () => {
    expect(resolveConnect([sh('a', '', '', 0, 0)], 'a', { x: 500, y: 500 })).toEqual({ kind: 'none' })
  })
})
