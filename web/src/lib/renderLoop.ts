import type { Engine } from './engine'
import type { RenderResult } from './protocol'

// Decision from docs/perf.md: fast mode is on.
export const FAST_MODE = true

export interface RenderLoopOpts {
  debounceMs: number
  idleFullMs: number
  fastMode: boolean
}

export function createRenderLoop(
  engine: Pick<Engine, 'call'>,
  opts: RenderLoopOpts,
  onResult: (r: RenderResult, full: boolean) => void,
  onBusy: (busy: boolean) => void = () => {},
) {
  let seq = 0
  let delivered = 0
  let running = false
  let pending: { text: string; full: boolean } | null = null
  let debounce: ReturnType<typeof setTimeout> | undefined
  let idle: ReturnType<typeof setTimeout> | undefined
  let lastText = ''
  let disposed = false

  async function run(text: string, full: boolean) {
    running = true
    onBusy(true)
    const n = ++seq

    try {
      const r = await engine.call('render', opts.fastMode && !full ? { text, seeds: [1] } : { text })
      if (!disposed && n > delivered) {
        delivered = n
        onResult(r, full || !opts.fastMode)
      }
    } catch {
      // A failed render leaves the previous result on screen.
    }

    running = false

    if (pending && !disposed) {
      const p = pending
      pending = null
      void run(p.text, p.full)
    } else {
      onBusy(false)
    }
  }

  function request(text: string, full: boolean) {
    if (running) pending = { text, full }
    else void run(text, full)
  }

  return {
    textChanged(text: string) {
      lastText = text
      clearTimeout(debounce)
      clearTimeout(idle)
      debounce = setTimeout(() => {
        request(lastText, false)

        if (opts.fastMode) idle = setTimeout(() => request(lastText, true), opts.idleFullMs)
      }, opts.debounceMs)
    },
    dispose() {
      disposed = true
      clearTimeout(debounce)
      clearTimeout(idle)
    },
  }
}
