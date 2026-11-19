/// <reference lib="webworker" />
declare const d2editor: { call(op: string, args: string): string }

async function load(base: string) {
  importScripts(base + 'wasm_exec.js')

  const res = await fetch(base + 'editor.wasm')
  const total = Number(res.headers.get('content-length')) || 0
  const reader = res.body!.getReader()
  const chunks: Uint8Array[] = []
  let loaded = 0

  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    chunks.push(value)
    loaded += value.length
    postMessage({ type: 'progress', loaded, total })
  }

  const go = new (self as any).Go()
  const { instance } = await WebAssembly.instantiate(await new Blob(chunks as BlobPart[]).arrayBuffer(), go.importObject)
  ;(self as any).onD2EditorReady = () => postMessage({ type: 'ready' })
  go.run(instance)
}

onmessage = (e: MessageEvent<{ type?: string; base?: string; id: number; op: string; args: unknown }>) => {
  const m = e.data

  if (m.type === 'init') {
    void load(m.base!)

    return
  }

  let json: string

  try {
    json = d2editor.call(m.op, JSON.stringify(m.args))
  } catch (err) {
    json = JSON.stringify({ ok: false, error: { code: 'internal', message: String(err), diag: null } })
  }

  postMessage({ id: m.id, json })
}
