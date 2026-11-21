// @vitest-environment jsdom
import { afterEach, expect, it } from 'vitest'
import { flushSync, mount, unmount } from 'svelte'
import Resizer from './Resizer.svelte'

let comp: ReturnType<typeof mount> | null = null
afterEach(() => {
  if (comp) unmount(comp)
  document.body.innerHTML = ''
})

// Mounts the handle before a 300px column and returns the bound width.
function setup(start: number | null) {
  let width = start
  const host = document.body.appendChild(document.createElement('div'))
  comp = mount(Resizer, {
    target: host,
    props: {
      get width() {
        return width
      },
      set width(v) {
        width = v
      },
      min: 100,
      max: () => 500,
      label: 'Resize',
    },
  })
  const col = host.appendChild(document.createElement('div'))
  col.getBoundingClientRect = () => ({ width: 300 }) as DOMRect
  const el = host.querySelector('[role=separator]') as HTMLElement
  el.setPointerCapture = () => {}
  flushSync()

  return { el, get: () => width }
}

const ptr = (el: HTMLElement, type: string, clientX: number) => el.dispatchEvent(new MouseEvent(type, { clientX, button: 0, bubbles: true }))

it('dragging left widens the column within its bounds', () => {
  const { el, get } = setup(null)
  ptr(el, 'pointerdown', 400)
  ptr(el, 'pointermove', 350)
  expect(get()).toBe(350)
  ptr(el, 'pointermove', 0)
  expect(get()).toBe(500)
  ptr(el, 'pointermove', 700)
  expect(get()).toBe(100)
  ptr(el, 'pointerup', 700)
  ptr(el, 'pointermove', 400)
  expect(get()).toBe(100)
})

it('arrow keys step and double click resets', () => {
  const { el, get } = setup(300)
  el.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowLeft', bubbles: true }))
  expect(get()).toBe(316)
  el.dispatchEvent(new MouseEvent('dblclick', { bubbles: true }))
  expect(get()).toBeNull()
})
