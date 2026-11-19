import type { Engine } from './engine'
import { EngineError, type ArgsOf, type Diag, type Op, type RenderResult, type ResultOf } from './protocol'
import { createRenderLoop, FAST_MODE } from './renderLoop'

export interface Selection {
  kind: 'shape' | 'edge'
  id: string
}

export function createAppState(engine: Engine) {
  const s = $state({
    code: '',
    render: null as RenderResult | null,
    errors: [] as Diag[],
    selection: null as Selection | null,
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
    // Replaces the document through the editor, so undo restores the old code.
    replaceCode(text: string) {
      sink(text)
    },
    registerCodeSink(fn: (text: string) => void) {
      sink = fn
    },
    select(sel: Selection | null) {
      s.selection = sel
    },
    dispose: () => loop.dispose(),
  })
}

export type AppState = ReturnType<typeof createAppState>
