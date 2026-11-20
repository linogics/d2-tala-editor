import { describe, expect, it } from 'vitest'
import { loadCols, loadFolded, loadGrid, loadInitial, saveCode, saveCols, saveGrid } from './persist'

const store = (v: string | null) => ({ getItem: () => v }) as unknown as Storage

describe('persist', () => {
  it('storage wins over example', () => {
    expect(loadInitial(store('saved'), 'ex')).toBe('saved')
  })
  it('falls back to example when storage empty', () => {
    expect(loadInitial(store(null), 'ex')).toBe('ex')
  })
  it('falls back when storage throws', () => {
    const s = {
      getItem() {
        throw new Error('blocked')
      },
    } as unknown as Storage
    expect(loadInitial(s, 'ex')).toBe('ex')
  })
  it('saveCode swallows quota error', () => {
    const s = {
      setItem() {
        throw new Error('quota')
      },
    } as unknown as Storage
    expect(() => saveCode(s, 'a')).not.toThrow()
  })
})

describe('grid settings', () => {
  it('round trips through storage', () => {
    const m = new Map<string, string>()
    const s = { getItem: (k: string) => m.get(k) ?? null, setItem: (k: string, v: string) => m.set(k, v) } as unknown as Storage
    expect(loadGrid(s)).toEqual({ snap: false, grid: 20 })
    saveGrid(s, { snap: true, grid: 8 })
    expect(loadGrid(s)).toEqual({ snap: true, grid: 8 })
  })
  it('ignores garbage', () => {
    expect(loadGrid(store('{"snap":1,"grid":0}'))).toEqual({ snap: false, grid: 20 })
    expect(loadGrid(store('nope'))).toEqual({ snap: false, grid: 20 })
  })
})

describe('column widths', () => {
  it('round trips and drops garbage', () => {
    const m = new Map<string, string>()
    const s = { getItem: (k: string) => m.get(k) ?? null, setItem: (k: string, v: string) => m.set(k, v) } as unknown as Storage
    expect(loadCols(s)).toEqual({ side: null, code: null })
    saveCols(s, { side: 300, code: null })
    expect(loadCols(s)).toEqual({ side: 300, code: null })
    expect(loadCols(store('{"side":-4,"code":"x"}'))).toEqual({ side: null, code: null })
    expect(loadCols(store('nope'))).toEqual({ side: null, code: null })
  })
  it('folded keeps only true flags and survives junk', () => {
    expect(loadFolded(store('{"Font":true,"Fill":false,"x":1}'))).toEqual({ Font: true })
    expect(loadFolded(store('[1]'))).toEqual({})
    expect(loadFolded(store('nope'))).toEqual({})
  })
})
