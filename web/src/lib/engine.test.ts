import { describe, expect, it } from 'vitest'
import { createEngine } from './engine'
import { EngineError } from './protocol'

function fakeWorker() {
  const w: any = { postMessage: (m: any) => (w.sent = m) }

  return w
}

describe('engine', () => {
  it('resolves ok result', async () => {
    const w = fakeWorker()
    const e = createEngine(w)
    w.onmessage({ data: { type: 'ready' } })
    const p = e.call('format', { text: 'a' })
    await Promise.resolve()
    w.onmessage({ data: { id: w.sent.id, json: '{"ok":true,"result":{"text":"a\\n"}}' } })
    expect(await p).toEqual({ text: 'a\n' })
    expect(w.sent.op).toBe('format')
  })

  it('rejects with EngineError code and diag', async () => {
    const w = fakeWorker()
    const e = createEngine(w)
    w.onmessage({ data: { type: 'ready' } })
    const p = e.call('format', { text: 'a' })
    await Promise.resolve()
    const diag = { line: 1, col: 2, message: 'bad' }
    w.onmessage({
      data: { id: w.sent.id, json: JSON.stringify({ ok: false, error: { code: 'broken_text', message: 'm', diag } }) },
    })
    const err = await p.catch((x) => x)
    expect(err).toBeInstanceOf(EngineError)
    expect(err.code).toBe('broken_text')
    expect(err.diag).toEqual(diag)
  })

  it('call waits for ready', async () => {
    const w = fakeWorker()
    const e = createEngine(w)
    const p = e.call('format', { text: 'a' })
    await new Promise((r) => setTimeout(r, 10))
    expect(w.sent).toBeUndefined()
    w.onmessage({ data: { type: 'ready' } })
    await e.ready
    await Promise.resolve()
    expect(w.sent.op).toBe('format')
    w.onmessage({ data: { id: w.sent.id, json: '{"ok":true,"result":{"text":"x"}}' } })
    expect(await p).toEqual({ text: 'x' })
  })
})
