import type { Engine } from './engine'
import { EngineError, type ArgsOf, type Diag, type Op, type RenderResult, type ResultOf } from './protocol'
import { createRenderLoop, FAST_MODE } from './renderLoop'

export interface Selection {
  kind: 'shape' | 'edge'
  id: string
}

// One canvas edit of a group action, the text is filled in by applyEdits.
export type Step = { [O in Op]: [O, Omit<ArgsOf<O>, 'text'>] }[Op]

export function createAppState(engine: Engine) {
  const s = $state({
    code: '',
    render: null as RenderResult | null,
    errors: [] as Diag[],
    // Several shapes, or a single edge. Empty means nothing is selected.
    selection: [] as Selection[],
    notice: null as string | null,
    busy: false,
    // Magnet: shows the grid and snaps dragged shapes to it.
    snap: false,
    grid: 20,
  })
  let sink: (text: string) => void = () => {}

  const loop = createRenderLoop(
    engine,
    { debounceMs: 300, idleFullMs: 1500, fastMode: FAST_MODE },
    (r) => {
      s.errors = r.errors
      // Keep the last good render while the code has errors.
      if (!r.errors.length) s.render = r
    },
    (b) => (s.busy = b),
  )

  return Object.assign(s, {
    engine,
    setCodeFromEditor(text: string) {
      s.code = text
      loop.textChanged(text)
    },
    async applyEdit<O extends Op>(op: O, args: Omit<ArgsOf<O>, 'text'>): Promise<ResultOf<O> | null> {
      try {
        const r = await engine.call(op, { ...args, text: s.code } as ArgsOf<O>)
        const text = (r as { text?: unknown }).text
        s.notice = null
        if (typeof text === 'string') sink(text)

        return r
      } catch (e) {
        if (!(e instanceof EngineError)) throw e
        s.notice = e.code === 'broken_text' ? 'Fix the code error first' : e.message

        return null
      }
    },
    // Runs the steps on the text one after another and pushes the result once, so the group is one undo step.
    async applyEdits(steps: Step[]): Promise<boolean> {
      let text = s.code
      try {
        for (const [op, args] of steps) {
          const r = await engine.call(op, { ...args, text } as ArgsOf<typeof op>)
          const t = (r as { text?: unknown }).text
          if (typeof t === 'string') text = t
        }
      } catch (e) {
        if (!(e instanceof EngineError)) throw e
        s.notice = e.code === 'broken_text' ? 'Fix the code error first' : e.message

        return false
      }
      s.notice = null
      if (text !== s.code) sink(text)

      return true
    },
    // Replaces the document through the editor, so undo restores the old code.
    replaceCode(text: string) {
      sink(text)
    },
    registerCodeSink(fn: (text: string) => void) {
      sink = fn
    },
    select(sel: Selection | null) {
      s.selection = sel ? [sel] : []
    },
    selectShapes(ids: string[]) {
      s.selection = ids.map((id) => ({ kind: 'shape' as const, id }))
    },
    // Adds or removes a shape. An edge in the selection is dropped, since groups hold shapes only.
    toggle(id: string) {
      const shapes = s.selection.filter((x) => x.kind === 'shape')
      s.selection = shapes.some((x) => x.id === id) ? shapes.filter((x) => x.id !== id) : [...shapes, { kind: 'shape', id }]
    },
    dispose: () => loop.dispose(),
  })
}

export type AppState = ReturnType<typeof createAppState>
