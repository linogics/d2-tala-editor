/// <reference lib="webworker" />
declare const d2editor: { call(op: string, args: string): string }

// Base is the worker script dir in dev and the app root in build, so use the page-relative root.
const base = new URL('.', self.location.href).href.replace(/assets\/$/, '').replace(/src\/lib\/$/, '')

importScripts(base + 'wasm_exec.js')

async function load() {
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

onmessage = (e: MessageEvent<{ id: number; op: string; args: unknown }>) => {
  const { id, op, args } = e.data
  postMessage({ id, json: d2editor.call(op, JSON.stringify(args)) })
}

load()
