// @vitest-environment node
import { existsSync, readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { beforeAll, describe, expect, it } from 'vitest'
import { edgeFields, fieldSteps, shapeFieldsFor, type Field } from './fields'
import { templates } from './templates'

const wasm = fileURLToPath(new URL('../../public/editor.wasm', import.meta.url))
const glue = fileURLToPath(new URL('../../public/wasm_exec.js', import.meta.url))

// Values to try per field, numbers at both ends of their range.
const samples = (f: Field): string[] =>
  f.kind === 'number'
    ? [f.min, f.max].filter((n) => n != null).map(String)
    : f.kind === 'select'
      ? f.options!
      : [{ text: 'x', color: '#ff0000', bool: 'true', url: 'https://example.com/a.svg' }[f.kind]]

// Every option the Properties panel offers must be accepted by D2 on every shape it offers it for.
describe.skipIf(!existsSync(wasm))('every field on every shape (skipped: web/public/editor.wasm missing, run wasm/scripts/build.sh)', () => {
  let call: (op: string, args: object) => any

  beforeAll(async () => {
    const g = globalThis as any
    ;(0, eval)(readFileSync(glue, 'utf8'))
    const go = new g.Go()
    const ready = new Promise<void>((r) => (g.onD2EditorReady = r))
    const { instance } = await WebAssembly.instantiate(readFileSync(wasm), go.importObject)
    void go.run(instance)
    await ready
    call = (op, args) => JSON.parse(g.d2editor.call(op, JSON.stringify(args)))
  }, 60000)

  it.each(templates.map((t) => [t.key, t.template]))('%s', (key, template) => {
    const ins = call('insert', { text: '', parent: '', template })
    expect(ins.ok, `${key}: ${JSON.stringify(ins.error)}`).toBe(true)
    const { text, id } = ins.result
    const shape = call('render', { text, seeds: [1] }).result.shapes.find((s: any) => s.id === id)

    const bad = shapeFieldsFor([shape.type]).flatMap((f) =>
      samples(f).flatMap((v) => {
        let t = text
        for (const [op, args] of fieldSteps(shape, f.attr, v)) {
          const r = call(op, { ...args, text: t })
          if (!r.ok) return [`${f.attr}=${v}: ${r.error.message}`]
          t = r.result.text
        }
        const errs = call('render', { text: t, seeds: [1] }).result.errors ?? []

        return errs.map((e: any) => `${f.attr}=${v}: ${e.message}`)
      }),
    )
    expect(bad).toEqual([])
  })

  it('edge fields', () => {
    const bad = edgeFields.flatMap((f) =>
      samples(f).flatMap((v) => {
        const r = call('set', { text: 'a -> b', id: '(a -> b)[0]', attr: f.attr, value: v })

        return r.ok ? [] : [`${f.attr}=${v}: ${r.error.message}`]
      }),
    )
    expect(bad).toEqual([])
  })

  it('shape change drops refused styles', () => {
    const r = call('set', { text: 'a: {style.3d: true; style.double-border: true}', id: 'a', attr: 'shape', value: 'cylinder' })
    expect(r.ok, JSON.stringify(r.error)).toBe(true)
  })
})
