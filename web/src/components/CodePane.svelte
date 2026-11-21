<script lang="ts">
  import { onMount } from 'svelte'
  import { basicSetup } from 'codemirror'
  import { EditorState, StateEffect, StateField, Transaction } from '@codemirror/state'
  import { Decoration, EditorView, type DecorationSet } from '@codemirror/view'
  import { isolateHistory, redo, undo } from '@codemirror/commands'
  import { autocompletion, type CompletionContext } from '@codemirror/autocomplete'
  import { setDiagnostics } from '@codemirror/lint'
  import type { AppState } from '../lib/state.svelte'
  import type { Range } from '../lib/protocol'
  import { d2Language } from '../lib/d2lang'
  import { HighlightStyle, syntaxHighlighting } from '@codemirror/language'
  import { tags } from '@lezer/highlight'

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

  // Colors come from the CSS tokens, so the editor follows light and dark mode without a reconfigure.
  const theme = EditorView.theme({
    '&': { height: '100%', backgroundColor: 'var(--surface)', color: 'var(--text)' },
    '.cm-scroller': { fontFamily: 'var(--mono)', fontSize: '12.5px', lineHeight: '1.6' },
    '.cm-content': { caretColor: 'var(--primary)' },
    '.cm-cursor, .cm-dropCursor': { borderLeftColor: 'var(--primary)', borderLeftWidth: '2px' },
    '&.cm-focused': { outline: 'none' },
    '&.cm-focused > .cm-scroller > .cm-selectionLayer .cm-selectionBackground, .cm-selectionBackground, ::selection': {
      backgroundColor: 'var(--selection) !important',
    },
    '.cm-gutters': { backgroundColor: 'var(--surface)', color: 'var(--muted)', border: 'none' },
    '.cm-activeLine': { backgroundColor: 'var(--primary-soft)' },
    '.cm-activeLineGutter': { backgroundColor: 'transparent', color: 'var(--text)' },
    '.cm-matchingBracket, &.cm-focused .cm-matchingBracket': { backgroundColor: 'var(--secondary-soft)', outline: 'none' },
    '.cm-foldPlaceholder': { backgroundColor: 'var(--surface-2)', border: '1px solid var(--border)', color: 'var(--muted)' },
    '.cm-tooltip': { backgroundColor: 'var(--surface)', border: '1px solid var(--border)', borderRadius: '8px', boxShadow: 'var(--shadow)', overflow: 'hidden' },
    '.cm-tooltip-autocomplete > ul > li[aria-selected]': { backgroundColor: 'var(--primary)', color: 'var(--primary-fg)' },
    '.cm-panels': { backgroundColor: 'var(--surface-2)', color: 'var(--text)' },
    '.cm-searchMatch': { backgroundColor: 'var(--secondary-soft)' },
    '.cm-ref': { backgroundColor: 'var(--secondary-soft)', borderRadius: '3px' },
  })
  const highlight = HighlightStyle.define([
    { tag: tags.keyword, color: 'var(--syn-keyword)', fontWeight: '600' },
    { tag: tags.string, color: 'var(--syn-string)' },
    { tag: tags.comment, color: 'var(--syn-comment)', fontStyle: 'italic' },
    { tag: tags.operator, color: 'var(--syn-operator)', fontWeight: '600' },
    { tag: tags.punctuation, color: 'var(--muted)' },
  ])

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
    const ids = (state.render?.shapes ?? []).map((sh) => sh.id)
    const all = await state.engine.call('refRangesAll', { ids, text }).catch(() => ({}) as Record<string, Range[]>)
    const hits: { id: string; len: number }[] = []
    for (const [id, ranges] of Object.entries(all)) {
      for (const r of ranges) {
        const from = pos(view.state.doc, r.fromLine, r.fromCol)
        const to = pos(view.state.doc, r.toLine, r.toCol)
        if (at >= from && at <= to) hits.push({ id, len: to - from })
      }
    }
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
          theme,
          syntaxHighlighting(highlight),
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

    state.registerEditor({
      undo: () => undo(view),
      redo: () => redo(view),
      reveal(line, col) {
        const at = pos(view.state.doc, line, col)
        view.dispatch({ selection: { anchor: at }, effects: EditorView.scrollIntoView(at, { y: 'center' }) })
        view.focus()
      },
    })

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
    if (!sel.length) return void view.dispatch({ effects: setRefs.of(Decoration.none) })
    state.engine
      .call('refRangesAll', { ids: sel.map((x) => x.id), text })
      .then((all) => Object.values(all).flat())
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

<section class="pane">
  <div class="panel-head">Code</div>
  <div class="host" bind:this={host}></div>
</section>

<style>
  .pane { display: flex; flex-direction: column; min-height: 0; background: var(--surface); }
  .host { flex: 1; min-height: 0; overflow: hidden; }
</style>
