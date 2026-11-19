<script lang="ts">
  import { onMount } from 'svelte'
  import { basicSetup } from 'codemirror'
  import { EditorState, StateEffect, StateField, Transaction } from '@codemirror/state'
  import { Decoration, EditorView, type DecorationSet } from '@codemirror/view'
  import { isolateHistory } from '@codemirror/commands'
  import { autocompletion, type CompletionContext } from '@codemirror/autocomplete'
  import { setDiagnostics } from '@codemirror/lint'
  import type { AppState } from '../lib/state.svelte'
  import type { Range } from '../lib/protocol'
  import { d2Language } from '../lib/d2lang'

  let { state }: { state: AppState } = $props()
  let host: HTMLDivElement
  let view: EditorView

  const setRefs = StateEffect.define<DecorationSet>()
  const refField = StateField.define<DecorationSet>({
    create: () => Decoration.none,
    update(v, tr) {
      v = v.map(tr.changes)
      for (const e of tr.effects) if (e.is(setRefs)) v = e.value

      return v
    },
    provide: (f) => EditorView.decorations.from(f),
  })

  const pos = (doc: EditorState['doc'], line: number, col: number) => {
    const l = doc.line(Math.min(Math.max(line, 0), doc.lines - 1) + 1)

    return Math.min(l.from + col, l.to)
  }

  async function complete(ctx: CompletionContext) {
    const word = ctx.matchBefore(/[\w.-]*/)
    if (!word || (word.from === word.to && !ctx.explicit)) return null
    const line = ctx.state.doc.lineAt(ctx.pos)
    const items = await state.engine.call('completions', {
      text: ctx.state.doc.toString(),
      line: line.number - 1,
      col: ctx.pos - line.from,
    })

    return {
      from: word.from,
      options: items.map((i) => ({ label: i.Label, detail: i.Detail, apply: i.InsertText, type: i.Kind.toLowerCase() })),
    }
  }

  // Click inside the refs of a shape selects it. The innermost (shortest) range wins.
  async function selectAt(at: number) {
    const text = view.state.doc.toString()
    const hits: { id: string; len: number }[] = []

    await Promise.all(
      (state.render?.shapes ?? []).map(async (sh) => {
        const ranges = await state.engine.call('refRanges', { id: sh.id, text }).catch(() => [] as Range[])
        for (const r of ranges) {
          const from = pos(view.state.doc, r.fromLine, r.fromCol)
          const to = pos(view.state.doc, r.toLine, r.toCol)
          if (at >= from && at <= to) hits.push({ id: sh.id, len: to - from })
        }
      }),
    )
    const best = hits.sort((a, b) => a.len - b.len)[0]
    if (best) state.select({ kind: 'shape', id: best.id })
  }

  onMount(() => {
    view = new EditorView({
      parent: host,
      state: EditorState.create({
        doc: state.code,
        extensions: [
          basicSetup,
          d2Language,
          refField,
          autocompletion({ override: [complete] }),
          EditorView.theme({ '&': { height: '100%' }, '.cm-ref': { backgroundColor: 'rgba(80,140,255,.25)' } }),
          EditorView.updateListener.of((u) => {
            if (u.docChanged) state.setCodeFromEditor(u.state.doc.toString())
            else if (u.selectionSet && u.transactions.some((t) => t.isUserEvent('select.pointer')))
              void selectAt(u.state.selection.main.head)
          }),
        ],
      }),
    })
    // A canvas edit is one undo step, isolated from surrounding typing.
    state.registerCodeSink((text) =>
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: text },
        annotations: [isolateHistory.of('full'), Transaction.userEvent.of('input.canvas')],
      }),
    )

    return () => view.destroy()
  })

  $effect(() => {
    const doc = view.state.doc
    const diags = state.errors.map((e) => {
      const from = pos(doc, e.line, e.col)

      return { from, to: Math.min(from + 1, doc.length), severity: 'error' as const, message: e.message }
    })
    view.dispatch(setDiagnostics(view.state, diags))
  })

  $effect(() => {
    const sel = state.selection
    const text = state.code
    if (!sel) return void view.dispatch({ effects: setRefs.of(Decoration.none) })
    state.engine
      .call('refRanges', { id: sel.id, text })
      .then((ranges) => {
        const doc = view.state.doc
        const marks = ranges
          .map((r) => [pos(doc, r.fromLine, r.fromCol), pos(doc, r.toLine, r.toCol)])
          .filter(([f, t]) => t > f)
          .sort((a, b) => a[0] - b[0])
          .map(([f, t]) => Decoration.mark({ class: 'cm-ref' }).range(f, t))
        view.dispatch({ effects: setRefs.of(Decoration.set(marks)) })
      })
      .catch(() => {})
  })
</script>

<div class="pane" bind:this={host}></div>

<style>
  .pane {
    height: 100%;
    overflow: hidden;
  }
</style>
