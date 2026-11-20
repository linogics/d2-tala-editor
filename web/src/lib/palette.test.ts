// @vitest-environment jsdom
import { describe, expect, it, vi } from 'vitest'
import { flushSync, mount } from 'svelte'
import Palette from '../components/Palette.svelte'
import Properties from '../components/Properties.svelte'

function shape(over = {}) {
  return { id: 'a', parent: '', type: 'rectangle', container: '', x: 1, y: 2, w: 10, h: 10, pinned: false, pinnable: true, ...over }
}

function props(over = {}, shapes = [shape(over)], classes: { name: string; attrs: object }[] = []) {
  const st = {
    selection: shapes.map((s) => ({ kind: 'shape', id: s.id })),
    render: { shapes, edges: [], classes },
    applyEdit: vi.fn(async () => ({ id: 'a', text: '' })),
    applyEdits: vi.fn(async () => true),
    select: vi.fn(),
  }
  const target = document.createElement('div')
  document.body.appendChild(target)
  mount(Properties, { target, props: { state: st as any } })
  flushSync()

  return { st, target }
}

function change(el: Element, value: string) {
  ;(el as HTMLInputElement).value = value
  el.dispatchEvent(new Event('change', { bubbles: true }))
}

describe('palette and properties', () => {
  it('palette lists 19 shapes and 9 special objects', () => {
    const target = document.createElement('div')
    mount(Palette, { target })
    flushSync()
    expect(target.querySelectorAll('.item').length).toBe(28)
    expect(target.querySelectorAll('h4').length).toBe(2)
  })

  it('palette sorts each group by name', () => {
    const target = document.createElement('div')
    mount(Palette, { target })
    flushSync()
    const names = (g: string) => Array.from(target.querySelectorAll('.item')).filter((x) => !!x.querySelector(g === 'special' ? '.special' : '.icon:not(.special)')).map((x) => x.textContent!.trim())
    for (const g of ['shape', 'special']) expect(names(g)).toEqual([...names(g)].sort((a, b) => a.localeCompare(b)))
  })

  it('fill change sends set style.fill', () => {
    const { st, target } = props()
    change(target.querySelector('input[type=color]')!, '#ff0000')
    expect(st.applyEdit).toHaveBeenCalledWith('set', { id: 'a', attr: 'style.fill', value: '#ff0000' })
  })

  it('empty label sends null', () => {
    const { st, target } = props()
    change(target.querySelectorAll('input')[1], '')
    expect(st.applyEdit).toHaveBeenCalledWith('set', { id: 'a', attr: 'label', value: null })
  })

  it('pin toggle disabled for grid child', () => {
    const { target } = props({ pinnable: false })
    expect((target.querySelector('input[type=checkbox]') as HTMLInputElement).disabled).toBe(true)
  })

  it('group edit sets the attr on every shape and marks mixed values', () => {
    const a = shape({ attrs: { 'style.fill': '#ff0000', 'style.opacity': '0.5' } })
    const b = shape({ id: 'b', attrs: { 'style.fill': '#00ff00', 'style.opacity': '0.5' } })
    const { st, target } = props({}, [a, b])
    expect(target.textContent).toContain('2 shapes selected')
    expect(target.querySelector('input[placeholder=mixed]')).not.toBeNull()
    expect(target.textContent).not.toContain('Top')
    change(target.querySelector('input[type=color]')!, '#0000ff')
    expect(st.applyEdits).toHaveBeenCalledWith([
      ['set', { id: 'a', attr: 'style.fill', value: '#0000ff' }],
      ['set', { id: 'b', attr: 'style.fill', value: '#0000ff' }],
    ])
  })
})

describe('class chips', () => {
  it('adding a second class writes an array, removing one writes a scalar', () => {
    const over = { attrs: { class: 'x' } }
    const { st, target } = props(over, [shape(over)], [{ name: 'x', attrs: {} }, { name: 'y', attrs: {} }])
    change(target.querySelector('select[aria-label="Add class"]')!, 'y')
    expect(st.applyEdit).toHaveBeenCalledWith('set', { id: 'a', attr: 'class', value: '[x; y]' })
    ;(target.querySelector('button[aria-label="Remove class x"]') as HTMLButtonElement).click()
    expect(st.applyEdit).toHaveBeenCalledWith('set', { id: 'a', attr: 'class', value: null })
  })
})
