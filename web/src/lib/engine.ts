import { EngineError, type ArgsOf, type Op, type ResultOf } from './protocol'

export interface Engine {
  call<O extends Op>(op: O, args: ArgsOf<O>): Promise<ResultOf<O>>
  ready: Promise<void>
  onProgress(cb: (loaded: number, total: number) => void): void
}

export function createWorker(): Worker {
  const w = new Worker(new URL('./worker.ts', import.meta.url), { type: 'classic' })
  w.postMessage({ type: 'init', base: new URL(import.meta.env.BASE_URL, document.baseURI).href })

  return w
}

export function createEngine(worker: Worker): Engine {
  let nextId = 1
  const pending = new Map<number, (json: string) => void>()
  const progress: ((loaded: number, total: number) => void)[] = []
  let resolveReady!: () => void
  const ready = new Promise<void>((r) => (resolveReady = r))

  worker.onmessage = (e: MessageEvent) => {
    const m = e.data
    if (m.type === 'ready') resolveReady()
    else if (m.type === 'progress') progress.forEach((cb) => cb(m.loaded, m.total))
    else pending.get(m.id)?.(m.json), pending.delete(m.id)
  }

  return {
    ready,
    onProgress: (cb) => void progress.push(cb),
    async call<O extends Op>(op: O, args: ArgsOf<O>) {
      await ready
      const id = nextId++

      return new Promise<ResultOf<O>>((resolve, reject) => {
        pending.set(id, (json) => {
          const r = JSON.parse(json)
          if (r.ok) resolve(r.result)
          else reject(new EngineError(r.error.code, r.error.message, r.error.diag))
        })
        worker.postMessage({ id, op, args })
      })
    },
  }
}
