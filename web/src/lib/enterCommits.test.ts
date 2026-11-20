// @vitest-environment node
import { describe, expect, it, vi } from 'vitest'
import { enterCommits } from './fields'

const press = (key: string, shiftKey = false) => {
  const e = { key, shiftKey, isComposing: false, preventDefault: vi.fn(), currentTarget: { blur: vi.fn() } }
  enterCommits(e as unknown as KeyboardEvent)
  return { kept: e.preventDefault.mock.calls.length === 0, blurred: e.currentTarget.blur.mock.calls.length > 0 }
}

describe('enterCommits', () => {
  it('Enter applies, Shift+Enter and other keys type', () => {
    expect(press('Enter')).toEqual({ kept: false, blurred: true })
    expect(press('Enter', true)).toEqual({ kept: true, blurred: false })
    expect(press('a')).toEqual({ kept: true, blurred: false })
  })
})
