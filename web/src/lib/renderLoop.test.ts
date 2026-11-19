import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createRenderLoop } from './renderLoop'
import type { RenderResult } from './protocol'

const res = (svg: string): RenderResult => ({ svg, shapes: [], edges: [], errors: [], config: null, classes: [] })

function setup(fastMode: boolean) {
  const calls: { args: any; resolve: (r: RenderResult) => void }[] = []
  const engine = {
    call: (_op: string, args: any) => new Promise<RenderResult>((resolve) => calls.push({ args, resolve })),
  } as any
  const onResult = vi.fn()
  const onBusy = vi.fn()
  const loop = createRenderLoop(engine, { debounceMs: 300, idleFullMs: 1500, fastMode }, onResult, onBusy)

  return { calls, onResult, onBusy, loop }
}

beforeEach(() => vi.useFakeTimers())
afterEach(() => vi.useRealTimers())

describe('renderLoop', () => {
  it('debounces 300 ms', async () => {
    const { calls, loop } = setup(false)
    loop.textChanged('a')
    await vi.advanceTimersByTimeAsync(100)
    loop.textChanged('ab')
    await vi.advanceTimersByTimeAsync(100)
    loop.textChanged('abc')
    await vi.advanceTimersByTimeAsync(299)
    expect(calls).toHaveLength(0)
    await vi.advanceTimersByTimeAsync(1)
    expect(calls).toHaveLength(1)
    expect(calls[0].args.text).toBe('abc')
  })

  it('drops stale results', async () => {
    const { calls, onResult, loop } = setup(false)
    loop.textChanged('a')
    await vi.advanceTimersByTimeAsync(300)
    calls[0].resolve(res('1'))
    await vi.advanceTimersByTimeAsync(0)
    loop.textChanged('b')
    await vi.advanceTimersByTimeAsync(300)
    calls[1].resolve(res('2'))
    await vi.advanceTimersByTimeAsync(0)
    expect(onResult.mock.calls.map((c) => c[0].svg)).toEqual(['1', '2'])
  })

  it('one render in flight', async () => {
    const { calls, loop } = setup(false)
    loop.textChanged('a')
    await vi.advanceTimersByTimeAsync(300)
    loop.textChanged('b')
    await vi.advanceTimersByTimeAsync(300)
    expect(calls).toHaveLength(1)
    calls[0].resolve(res('1'))
    await vi.advanceTimersByTimeAsync(0)
    expect(calls).toHaveLength(2)
    expect(calls[1].args.text).toBe('b')
  })

  it('pending coalesces to newest', async () => {
    const { calls, onBusy, loop } = setup(false)
    loop.textChanged('a')
    await vi.advanceTimersByTimeAsync(300)
    for (const t of ['b', 'c', 'd']) {
      loop.textChanged(t)
      await vi.advanceTimersByTimeAsync(300)
    }
    calls[0].resolve(res('1'))
    await vi.advanceTimersByTimeAsync(0)
    expect(calls).toHaveLength(2)
    expect(calls[1].args.text).toBe('d')
    calls[1].resolve(res('2'))
    await vi.advanceTimersByTimeAsync(0)
    expect(onBusy.mock.calls.map((c) => c[0])).toEqual([true, true, false])
  })

  it('fast then full', async () => {
    const { calls, onResult, loop } = setup(true)
    loop.textChanged('a')
    await vi.advanceTimersByTimeAsync(300)
    expect(calls[0].args).toEqual({ text: 'a', seeds: [1] })
    calls[0].resolve(res('f'))
    await vi.advanceTimersByTimeAsync(1500)
    expect(calls).toHaveLength(2)
    expect(calls[1].args).toEqual({ text: 'a' })
    calls[1].resolve(res('F'))
    await vi.advanceTimersByTimeAsync(0)
    expect(onResult.mock.calls.map((c) => c[1])).toEqual([false, true])
  })

  it('no full rerender when fastMode off', async () => {
    const { calls, loop } = setup(false)
    loop.textChanged('a')
    await vi.advanceTimersByTimeAsync(300)
    expect(calls[0].args).toEqual({ text: 'a' })
    calls[0].resolve(res('1'))
    await vi.advanceTimersByTimeAsync(5000)
    expect(calls).toHaveLength(1)
  })
})
