import { describe, expect, it } from 'vitest'
import { themeIdOf, withTheme } from './theme'

describe('theme', () => {
  it('ignores dark-theme-id', () => expect(themeIdOf('vars: {d2-config: {dark-theme-id: 200\ntheme-id: 3}}')).toBe(3))
  it('reads default 0', () => expect(themeIdOf('a -> b')).toBe(0))
  it('appends a vars block', () => {
    const out = withTheme('a -> b', 200)
    expect(out).toContain('theme-id: 200')
    expect(themeIdOf(out)).toBe(200)
  })
  it('replaces an existing id', () => expect(withTheme(withTheme('a', 1), 4)).toBe(withTheme('a', 4)))
})
