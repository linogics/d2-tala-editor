// @vitest-environment jsdom
import { afterEach, describe, expect, it, vi } from 'vitest'
import { currentMode, currentPref, setMode } from './mode'

describe('mode', () => {
  const root = document.createElement('html')
  afterEach(() => vi.unstubAllGlobals())

  it('explicit choice wins, system follows the OS', () => {
    vi.stubGlobal('matchMedia', () => ({ matches: true }))
    setMode('light', root)
    expect([currentPref(root), currentMode(root)]).toEqual(['light', 'light'])
    expect(localStorage.getItem('d2-editor:mode')).toBe('light')

    setMode('system', root)
    expect(root.dataset.mode).toBeUndefined()
    expect([currentPref(root), currentMode(root)]).toEqual(['system', 'dark'])
    expect(localStorage.getItem('d2-editor:mode')).toBeNull()
  })
})
