// @vitest-environment node
import { existsSync, readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const wasm = fileURLToPath(new URL('../../public/editor.wasm', import.meta.url))
const glue = fileURLToPath(new URL('../../public/wasm_exec.js', import.meta.url))

describe.skipIf(!existsSync(wasm))('real wasm (skipped: web/public/editor.wasm missing, run wasm/scripts/build.sh)', () => {
  it('renders and reports unknown ops', async () => {
    const g = globalThis as any
    ;(0, eval)(readFileSync(glue, 'utf8'))
    const go = new g.Go()
    const ready = new Promise<void>((r) => (g.onD2EditorReady = r))
    const { instance } = await WebAssembly.instantiate(readFileSync(wasm), go.importObject)
    void go.run(instance)
    await ready

    const r = JSON.parse(g.d2editor.call('render', JSON.stringify({ text: 'a -> b', seeds: [1] })))
    expect(r.ok).toBe(true)
    expect(r.result.shapes).toHaveLength(2)

    const refs = JSON.parse(g.d2editor.call('refRangesAll', JSON.stringify({ text: 'a -> b', ids: ['a', 'zzz'] })))
    expect(refs.ok).toBe(true)
    expect(refs.result.a).toHaveLength(1)
    expect(refs.result.zzz).toEqual([])

    const bad = JSON.parse(g.d2editor.call('nope', '{}'))
    expect(bad.ok).toBe(false)
    expect(bad.error.code).toBe('unknown_op')
  }, 60000)
})
