// @vitest-environment jsdom
// Drives the real Properties and Classes panels against the real wasm engine, the way a user clicks.
import { existsSync, readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { afterEach, beforeAll, describe, expect, it } from 'vitest'
import { flushSync, mount, unmount } from 'svelte'
import Properties from './Properties.svelte'
import ClassesPanel from './ClassesPanel.svelte'
import { createAppState, type AppState } from '../lib/state.svelte'
import { EngineError, type RenderResult } from '../lib/protocol'
import type { Engine } from '../lib/engine'
import { classFields, shapeFieldsFor, type Field } from '../lib/fields'
import { templates } from '../lib/templates'

const wasm = resolve(import.meta.dirname, '../../public/editor.wasm')
const glue = resolve(import.meta.dirname, '../../public/wasm_exec.js')

let raw: (op: string, args: object) => any
const engine: Engine = {
  ready: Promise.resolve(),
  onProgress() {},
  async call(op, args) {
    const r = raw(op, args)
    if (!r.ok) throw new EngineError(r.error.code, r.error.message, r.error.diag)

    return r.result
  },
}

let app: AppState
let comp: ReturnType<typeof mount> | null = null
let host: HTMLElement

const tick = () => new Promise((r) => setTimeout(r, 0))

// Renders the current code into the state, like the render loop after an edit.
function refresh() {
  const r: RenderResult = raw('render', { text: app.code, seeds: [1] }).result
  app.errors = r.errors
  if (!r.errors.length) app.render = r
  flushSync()

  return r
}

function open(Comp: typeof Properties | typeof ClassesPanel, text: string) {
  app = createAppState(engine)
  app.registerCodeSink((t) => (app.code = t))
  app.code = text
  refresh()
  host = document.createElement('div')
  document.body.append(host)
  comp = mount(Comp, { target: host, props: { state: app } })
  flushSync()
}

afterEach(() => {
  if (comp) unmount(comp)
  comp = null
  host?.remove()
  app?.dispose()
})

// Fires the change a user makes, waits for the edit and renders the result.
async function change(el: Element | null | undefined, value: string | boolean) {
  expect(el, 'control missing').toBeTruthy()
  const input = el as HTMLInputElement
  if (typeof value === 'boolean') input.checked = value
  else input.value = value
  input.dispatchEvent(new Event('change', { bubbles: true }))
  for (let i = 0; i < 4; i++) await tick()

  return refresh()
}

async function click(el: Element | null | undefined) {
  expect(el, 'button missing').toBeTruthy()
  ;(el as HTMLElement).click()
  for (let i = 0; i < 4; i++) await tick()

  return refresh()
}

// The UI must stay usable: no notice and code D2 accepts.
function ok(what: string) {
  const errs = refresh().errors.map((e) => e.message)
  expect({ what, notice: app.notice, errs }).toEqual({ what, notice: null, errs: [] })
}

const label = (name: string) => Array.from(host.querySelectorAll('label')).find((l) => l.textContent?.trim().startsWith(name))
const control = (fd: Field) => label(fd.name)?.querySelector('input:not([type=color]), select, textarea') as HTMLInputElement | null
// Arrowhead fields live in a per end group as "Shape" and "Label".
const head = (end: string, name: string) =>
  Array.from(host.querySelectorAll(`[aria-label="${end === 'source' ? 'Source' : 'Target'} head"] label`)).find((l) => l.textContent?.trim().startsWith(name))
const headCtl = (end: string, name: string) => head(end, name)?.querySelector('input, select, textarea') as HTMLInputElement | HTMLSelectElement | null
// Arrowhead fields by end box, the rest by label.
const ctl = (fd: Field) => {
  const m = fd.attr.match(/^(source|target)-arrowhead\.(shape|label)$/)
  return m ? headCtl(m[1], m[2] === 'shape' ? 'Shape' : 'Label') : control(fd)
}
const button = (text: string) => Array.from(host.querySelectorAll('button')).find((b) => b.textContent?.trim() === text)
const shape = (id: string) => app.render!.shapes.find((s) => s.id === id)!
const cls = (name: string) => app.render!.classes.find((c) => c.name === name)

const sample = (f: Field) =>
  f.kind === 'number'
    ? String(f.max ?? (f.min ?? 0) + 3)
    : f.kind === 'select'
      ? f.options![2]
      : { text: 'x\ny', color: '#ff0000', bool: true, url: 'https://example.com/a.svg' }[f.kind]

const insert = (template: string) => raw('insert', { text: '', parent: '', template }).result as { text: string; id: string }

describe.skipIf(!existsSync(wasm))('ui flows (skipped: web/public/editor.wasm missing, run wasm/scripts/build.sh)', () => {
  beforeAll(async () => {
    const g = globalThis as any
    ;(0, eval)(readFileSync(glue, 'utf8'))
    const go = new g.Go()
    const ready = new Promise<void>((r) => (g.onD2EditorReady = r))
    const { instance } = await WebAssembly.instantiate(readFileSync(wasm), go.importObject)
    void go.run(instance)
    await ready
    raw = (op, args) => JSON.parse(g.d2editor.call(op, JSON.stringify(args)))
  }, 60000)

  describe('properties on every shape', () => {
    it.each(templates.map((t) => [t.key, t.template]))('%s: set, toggle and clear every field', async (key, template) => {
      const { text, id } = insert(template)
      open(Properties, text)
      app.select({ kind: 'shape', id })
      flushSync()
      const type = shape(id).type

      for (const fd of shapeFieldsFor([type])) {
        const v = sample(fd)
        await change(control(fd), v)
        ok(`${key} ${fd.attr}=${v}`)
        if (fd.attr !== 'top' && fd.attr !== 'left') expect(shape(id).attrs[fd.attr], `${key} ${fd.attr}`).toBe(v === true ? 'true' : v)
        if (key === 'image' && fd.attr === 'icon') {
          // An image needs its icon: the edit is refused and the field shows the kept value.
          await change(control(fd), '')
          expect(app.notice).toBe('image shape must include an "icon" field')
          expect(control(fd)?.value).toBe(v)
          continue
        }
        await change(control(fd), typeof v === 'boolean' ? false : '')
        ok(`${key} clear ${fd.attr}`)
        expect(shape(id)?.attrs[fd.attr], `${key} clear ${fd.attr}`).toBeUndefined()
      }
    })

    it.each(templates.map((t) => [t.key, t.template]))('%s: label, pin, rename and delete', async (key, template) => {
      const { text, id } = insert(template)
      open(Properties, text)
      app.select({ kind: 'shape', id })
      flushSync()

      // D2 refuses line breaks in sql_table labels.
      const multi = key !== 'sql_table'
      await change(label('Label')?.querySelector('textarea'), multi ? 'Hello\nworld' : 'Hello')
      if (multi) expect(app.code).toContain('"Hello\\nworld"')
      ok(`${key} label`)
      await change(label('Label')?.querySelector('textarea'), '')
      ok(`${key} clear label`)

      const pin = label('Pinned')?.querySelector('input') as HTMLInputElement
      if (!pin.disabled) {
        await change(pin, true)
        ok(`${key} pin`)
        expect(shape(id).pinned).toBe(true)
        await change(label('Pinned')?.querySelector('input'), false)
        ok(`${key} unpin`)
        expect(shape(id).pinned).toBe(false)
      }

      await change(label('Id')?.querySelector('input'), 'renamed')
      ok(`${key} rename`)
      expect(app.selection).toEqual([{ kind: 'shape', id: 'renamed' }])
      flushSync()

      await click(button('Delete'))
      ok(`${key} delete`)
      // D2 keeps the children of a deleted container.
      expect(app.render!.shapes.map((s) => s.id)).not.toContain('renamed')
      expect(app.selection).toEqual([])
    })

    it.each(templates.map((t) => [t.key, t.template]))('%s: change to every other shape with styles set', async (key, template) => {
      const { text, id } = insert(template)
      const styled = raw('set', { text, id, attr: 'style.fill', value: '#00ff00' }).result.text
      open(Properties, styled)
      app.select({ kind: 'shape', id })
      flushSync()
      for (const to of ['rectangle', 'cylinder', 'circle', 'hexagon', 'oval']) {
        const fd = shapeFieldsFor([shape(id).type]).find((x) => x.kind === 'bool' && x.attr === 'style.3d') ?? shapeFieldsFor([shape(id).type]).find((x) => x.attr === 'style.double-border')
        if (fd) await change(control(fd), true)
        await change(label('Shape')?.querySelector('select'), to)
        ok(`${key} -> ${to}`)
        expect(shape(id).type).toBe(to)
      }
    })
  })

  describe('containers', () => {
    it('deletes a container with its children, Shift keeps them', async () => {
      for (const [shift, want] of [[false, ['x']], [true, ['b', 'x']]] as const) {
        open(Properties, 'x -> a.b\na: {b}')
        app.select({ kind: 'shape', id: 'a' })
        flushSync()
        button('Delete')!.dispatchEvent(new MouseEvent('click', { bubbles: true, shiftKey: shift }))
        for (let i = 0; i < 4; i++) await tick()
        ok(`shift=${shift}`)
        expect(refresh().shapes.map((s) => s.id).sort()).toEqual(want)
        unmount(comp!)
        comp = null
        host.remove()
        app.dispose()
      }
    })
  })

  describe('classes panel', () => {
    async function add(name: string) {
      return change(label('New class')?.querySelector('input'), name)
    }

    it('creates, picks and deletes a class', async () => {
      open(ClassesPanel, 'a')
      expect(host.textContent).toContain('No classes yet')
      await add('box')
      ok('add')
      expect(cls('box')).toEqual({ name: 'box', attrs: {} })
      expect(label('Name')?.querySelector('input')?.value).toBe('box')
      await click(button('Delete class'))
      ok('delete')
      expect(app.render!.classes).toEqual([])
      expect(label('Name')).toBeUndefined()
    })

    it('ignores blank and duplicate names', async () => {
      open(ClassesPanel, 'a')
      await add('   ')
      ok('blank')
      expect(app.render!.classes).toEqual([])
      await add('Box')
      await add('box')
      await add(' BOX ')
      ok('dup')
      expect(app.render!.classes.map((c) => c.name)).toEqual(['Box'])
    })

    it('refuses the reserved name classes with a notice', async () => {
      open(ClassesPanel, 'a')
      await add('classes')
      expect(app.notice).toBeTruthy()
      expect(app.code).toBe('a')
    })

    it.each(['my class', 'a.b', 'ünï', 'x-y_z', '1', 'with"quote', 'style', 'label'])('handles name %j', async (name) => {
      open(ClassesPanel, 'a')
      await add(name)
      ok(`add ${name}`)
      expect(app.render!.classes.map((c) => c.name)).toEqual([name])
      const fill = classFields.find((f) => f.attr === 'style.fill')!
      await change(control(fill), '#123456')
      ok(`fill ${name}`)
      expect(cls(name)?.attrs['style.fill']).toBe('#123456')
      await change(label('Name')?.querySelector('input'), name + '2')
      ok(`rename ${name}`)
      expect(app.render!.classes.map((c) => c.name)).toEqual([name + '2'])
      await click(button('Delete class'))
      ok(`delete ${name}`)
      expect(app.render!.classes).toEqual([])
    })

    it('sets, toggles and clears every class field', async () => {
      open(ClassesPanel, 'a')
      await add('c')
      for (const fd of classFields) {
        const v = sample(fd)
        await change(ctl(fd), v)
        ok(`set ${fd.attr}`)
        expect(cls('c')?.attrs[fd.attr], fd.attr).toBe(v === true ? 'true' : v)
        await change(ctl(fd), typeof v === 'boolean' ? false : '')
        ok(`clear ${fd.attr}`)
        expect(cls('c'), `class kept after clearing ${fd.attr}`).toEqual({ name: 'c', attrs: {} })
      }
    })

    it('keeps other attrs when clearing one', async () => {
      open(ClassesPanel, 'a')
      await add('c')
      const [fill, stroke] = ['style.fill', 'style.stroke'].map((a) => classFields.find((f) => f.attr === a)!)
      await change(control(fill), '#111111')
      await change(control(stroke), '#222222')
      await change(control(fill), '')
      ok('clear one')
      expect(cls('c')?.attrs).toEqual({ 'style.stroke': '#222222' })
    })

    it('renames to a case variant and refuses an existing name', async () => {
      open(ClassesPanel, 'classes: {a; b}\nx.class: a')
      await click(Array.from(host.querySelectorAll('[role=option]')).find((b) => b.textContent === 'a'))
      await change(label('Name')?.querySelector('input'), 'A')
      ok('case rename')
      expect(app.render!.classes.map((c) => c.name).sort()).toEqual(['A', 'b'])
      expect(shape('x').attrs.class).toBe('A')
      await change(label('Name')?.querySelector('input'), 'b')
      expect(app.notice).toBe('class "b" already exists')
      expect(label('Name')?.querySelector('input')?.value).toBe('A')
      await change(label('Name')?.querySelector('input'), '  ')
      expect(label('Name')?.querySelector('input')?.value).toBe('A')
      expect(app.render!.classes.map((c) => c.name).sort()).toEqual(['A', 'b'])
    })

    it('deleting a used class strips it from shapes and edges', async () => {
      open(ClassesPanel, 'classes: {a: {style.fill: red}; b}\nx.class: a\ny.class: [a; b]\nx -> y: {class: a}\nz: {class: [b; a]}')
      await click(Array.from(host.querySelectorAll('[role=option]')).find((b) => b.textContent === 'a'))
      await click(button('Delete class'))
      ok('delete used')
      expect(app.code).not.toMatch(/\ba\b(?!\s*->)/)
      expect(shape('x').attrs.class).toBeUndefined()
      expect(shape('y').attrs.class).toBe('b')
      expect(shape('z').attrs.class).toBe('b')
      expect(app.render!.edges[0].attrs.class).toBeUndefined()
    })
  })

  describe('assigning classes in properties', () => {
    const chips = () => Array.from(host.querySelectorAll('.chip')).map((c) => c.firstChild?.textContent)
    const addSel = () => host.querySelector('select[aria-label="Add class"]') as HTMLSelectElement

    it.each(templates.map((t) => [t.key, t.template]))('%s: add two, remove one, remove all', async (key, template) => {
      const { text, id } = insert(template)
      open(Properties, `classes: {p: {style.opacity: 0.5}; q: {style.font-size: 20}}\n${text}`)
      app.select({ kind: 'shape', id })
      flushSync()
      expect(host.textContent).not.toContain('none defined')

      await change(addSel(), 'p')
      ok(`${key} add p`)
      expect(shape(id).attrs.class).toBe('p')
      await change(addSel(), 'q')
      ok(`${key} add q`)
      expect(shape(id).attrs.class).toBe('p;q')
      expect(chips()).toEqual(['p', 'q'])
      expect(addSel()).toBeNull()

      await click(host.querySelector('[aria-label="Remove class p"]'))
      ok(`${key} remove p`)
      expect(shape(id).attrs.class).toBe('q')
      await click(host.querySelector('[aria-label="Remove class q"]'))
      ok(`${key} remove q`)
      expect(shape(id).attrs.class).toBeUndefined()
      expect(chips()).toEqual([])
    })

    it('class on an edge', async () => {
      open(Properties, 'classes: {p: {style.stroke: red}; q: {style.animated: true}}\na -> b')
      app.select({ kind: 'edge', id: '(a -> b)[0]' })
      flushSync()
      await change(addSel(), 'p')
      await change(addSel(), 'q')
      ok('edge add')
      expect(app.render!.edges[0].attrs.class).toBe('p;q')
      await click(host.querySelector('[aria-label="Remove class q"]'))
      await click(host.querySelector('[aria-label="Remove class p"]'))
      ok('edge remove')
      expect(app.render!.edges[0].attrs.class).toBeUndefined()
    })

    it('class to a group of shapes with mixed classes', async () => {
      open(Properties, 'classes: {p; q}\na.class: p\nb.class: [p; q]\nc')
      app.selectShapes(['a', 'b', 'c'])
      flushSync()
      expect(host.textContent).toContain('3 shapes selected')
      expect(addSel()).toBeTruthy()
      await change(addSel(), 'q')
      ok('group add')
      expect(['a', 'b', 'c'].map((x) => shape(x).attrs.class)).toEqual(['q', 'q', 'q'])
    })

    it('shows none defined when no classes exist', async () => {
      open(Properties, 'a')
      app.select({ kind: 'shape', id: 'a' })
      flushSync()
      expect(host.textContent).toContain('none defined')
    })

    it('3d class on a cylinder', async () => {
      open(Properties, 'classes: {c: {style.3d: true}}\na: {shape: cylinder}')
      app.select({ kind: 'shape', id: 'a' })
      flushSync()
      await change(addSel(), 'c')
      ok('3d class on cylinder')
    })
  })

  describe('edges', () => {
    it('set, toggle and clear every edge field', async () => {
      const { edgeFields } = await import('../lib/fields')
      // Both ends have a head, so every arrowhead field is enabled.
      open(Properties, 'a <-> b')
      app.select({ kind: 'edge', id: '(a <-> b)[0]' })
      flushSync()
      for (const fd of edgeFields) {
        const v = sample(fd)
        await change(ctl(fd), v)
        ok(`edge ${fd.attr}=${v}`)
        expect(app.render!.edges[0].attrs[fd.attr], fd.attr).toBe(v === true ? 'true' : v)
        await change(ctl(fd), typeof v === 'boolean' ? false : '')
        ok(`edge clear ${fd.attr}`)
        expect(app.render!.edges[0].attrs[fd.attr], fd.attr).toBeUndefined()
      }
      await change(label('Label')?.querySelector('textarea'), 'go')
      ok('edge label')
      await click(button('Delete'))
      ok('edge delete')
      expect(app.render!.edges).toEqual([])
    })
  })

  describe('edge heads', () => {
    it('toggles pick the op and keep the edge selected', async () => {
      open(Properties, 'a -> b\n(a -> b)[0].target-arrowhead.shape: diamond')
      app.select({ kind: 'edge', id: '(a -> b)[0]' })
      flushSync()
      const box = (end: string) => headCtl(end, `${end} head`)
      const shapeSel = (end: string) => headCtl(end, 'Shape') as HTMLSelectElement
      expect(shapeSel('target').value).toBe('diamond')
      expect(shapeSel('target').disabled).toBe(false)
      expect(shapeSel('source').disabled).toBe(true)
      expect((headCtl('source', 'Label') as HTMLInputElement).disabled).toBe(false)

      await change(box('target'), false)
      ok('line')
      expect(app.code).toContain('a -- b')
      expect(app.selection).toEqual([{ kind: 'edge', id: '(a -- b)[0]' }])
      // The disabled shape key stays, so the head comes back as it was.
      expect(shapeSel('target').disabled).toBe(true)
      expect(shapeSel('target').value).toBe('diamond')
      expect(app.render!.edges[0].attrs['target-arrowhead.shape']).toBe('diamond')

      await change(box('source'), true)
      ok('back')
      expect(app.code).toContain('a <- b')
      await change(box('target'), true)
      expect(app.code).toContain('a <-> b')
      expect(shapeSel('target').disabled).toBe(false)
      expect(shapeSel('target').value).toBe('diamond')
    })
  })

  describe('near', () => {
    it('offers positions and shapes outside its own branch', async () => {
      open(Properties, 'a\nb\nc: {d}\nc.near: a')
      app.select({ kind: 'shape', id: 'c' })
      flushSync()
      const sel = () => label('Near')!.querySelector('select')!
      const opts = Array.from(sel().options).map((o) => o.value)
      expect(opts).toEqual(['', 'top-left', 'top-center', 'top-right', 'center-left', 'center-right', 'bottom-left', 'bottom-center', 'bottom-right', 'a', 'b'])
      expect(sel().value).toBe('a')

      await change(sel(), 'b')
      ok('near b')
      expect(shape('c').attrs.near).toBe('b')
    })
  })

  describe('groups', () => {
    it('sets a field on many shapes of different types', async () => {
      open(Properties, 'a\nb: {shape: circle}\nc: {shape: cylinder}')
      app.selectShapes(['a', 'b', 'c'])
      flushSync()
      const types = ['rectangle', 'circle', 'cylinder']
      expect(shapeFieldsFor(types).some((f) => f.attr === 'style.3d')).toBe(false)
      const shadow = shapeFieldsFor(types).find((f) => f.attr === 'style.shadow')!
      await change(control(shadow), true)
      ok('group shadow')
      expect(['a', 'b', 'c'].map((x) => shape(x).attrs['style.shadow'])).toEqual(['true', 'true', 'true'])
      await change(label('Shape')?.querySelector('select'), 'hexagon')
      ok('group shape')
      expect(['a', 'b', 'c'].map((x) => shape(x).type)).toEqual(['hexagon', 'hexagon', 'hexagon'])
      await click(button('Delete'))
      ok('group delete')
      expect(app.render!.shapes).toEqual([])
    })

    it('deletes a container with its selected child', async () => {
      open(Properties, 'p: {c}\nq')
      app.selectShapes(['p', 'p.c'])
      flushSync()
      await click(button('Delete'))
      ok('container delete')
      expect(app.render!.shapes.map((s) => s.id)).toEqual(['q'])
    })
  })
})
