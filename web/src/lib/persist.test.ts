import { describe, expect, it } from 'vitest'
import { loadGrid, loadInitial, saveCode, saveGrid, shareUrl } from './persist'

const loc = (hash: string, href = 'http://x/p/' + hash) => ({ hash, href }) as Location
const store = (v: string | null) => ({ getItem: () => v }) as unknown as Storage
const dec = async (s: string) => 'decoded:' + s
const bad = async () => {
  throw new Error('bad')
}

describe('persist', () => {
  it('hash wins over storage', async () => {
    expect(await loadInitial(loc('#d2=abc'), store('saved'), dec, 'ex')).toBe('decoded:abc')
  })
  it('storage wins over example', async () => {
    expect(await loadInitial(loc(''), store('saved'), dec, 'ex')).toBe('saved')
  })
  it('falls back to example when storage empty', async () => {
    expect(await loadInitial(loc(''), store(null), dec, 'ex')).toBe('ex')
  })
  it('falls back when storage throws', async () => {
    const s = {
      getItem() {
        throw new Error('blocked')
      },
    } as unknown as Storage
    expect(await loadInitial(loc(''), s, dec, 'ex')).toBe('ex')
  })
  it('falls back on bad share link', async () => {
    expect(await loadInitial(loc('#d2=zz'), store('saved'), bad, 'ex')).toBe('saved')
  })
  it('shareUrl keeps path and replaces hash', () => {
    expect(shareUrl(loc('#d2=old'), 'new')).toBe('http://x/p/#d2=new')
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
