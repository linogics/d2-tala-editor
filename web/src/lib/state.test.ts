// @vitest-environment jsdom
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushSync, mount, unmount } from 'svelte'
import { EditorView } from '@codemirror/view'
import { undo } from '@codemirror/commands'
import CodePane from '../components/CodePane.svelte'
import { createAppState } from './state.svelte'
import { EngineError, type RenderResult } from './protocol'

const good: RenderResult = { svg: '<svg/>', shapes: [], edges: [], errors: [] }
const bad: RenderResult = { svg: '', shapes: [], edges: [], errors: [{ line: 0, col: 3, message: 'oops' }] }

function fake(handler: (op: string, args: any) => any) {
  return { call: vi.fn(async (op: string, args: any) => handler(op, args)), ready: Promise.resolve(), onProgress() {} } as any
}

beforeEach(() => vi.useFakeTimers())
afterEach(() => vi.useRealTimers())

describe('app state', () => {
  it('keeps last good render on error', async () => {
    const results = [good, bad]
    const st = createAppState(fake(() => results.shift()))
    st.setCodeFromEditor('a')
    await vi.advanceTimersByTimeAsync(300)
    st.setCodeFromEditor('a:{')
    await vi.advanceTimersByTimeAsync(300)
    expect(st.render).toEqual(good)
    expect(st.errors).toEqual(bad.errors)
    st.dispose()
  })

  it('applyEdit sends current text and pushes to sink', async () => {
    const eng = fake(() => ({ id: 'x', text: 'x\n' }))
    const st = createAppState(eng)
    const sink = vi.fn()
    st.registerCodeSink(sink)
    st.code = 'a'
    const r = await st.applyEdit('create', { parent: '', shape: 'rectangle' })
    expect(eng.call).toHaveBeenCalledWith('create', { parent: '', shape: 'rectangle', text: 'a' })
    expect(sink).toHaveBeenCalledWith('x\n')
    expect(r?.id).toBe('x')
    st.dispose()
  })

  it('applyEdit on broken text sets notice and keeps code', async () => {
    const st = createAppState(fake(() => Promise.reject(new EngineError('broken_text', 'bad'))))
    const sink = vi.fn()
    st.registerCodeSink(sink)
    st.code = 'a:{'
    expect(await st.applyEdit('format', {})).toBeNull()
    expect(st.notice).toBe('Fix the code error first')
    expect(st.code).toBe('a:{')
    expect(sink).not.toHaveBeenCalled()
    st.dispose()
  })

  it('undo restores text before a canvas edit', async () => {
    const st = createAppState(fake((op) => (op === 'render' ? good : { text: 'a\nb\n' })))
    st.code = 'a\n'
    const target = document.body.appendChild(document.createElement('div'))
    const app = mount(CodePane, { target, props: { state: st } })
    flushSync()
    const view = EditorView.findFromDOM(target)!
    await st.applyEdit('format', {})
    expect(st.code).toBe('a\nb\n')
    expect(undo(view)).toBe(true)
    expect(view.state.doc.toString()).toBe('a\n')
    expect(st.code).toBe('a\n')
    unmount(app)
    st.dispose()
  })
})
