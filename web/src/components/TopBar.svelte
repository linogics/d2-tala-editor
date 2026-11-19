<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import type { ThemeInfo } from '../lib/protocol'
  import { shareUrl } from '../lib/persist'
  import { themeIdOf, withTheme } from '../lib/theme'

  let { app, examples }: { app: AppState; examples: { name: string; code: string }[] } = $props()

  let themes = $state<ThemeInfo[]>([])
  $effect(() => void app.engine.call('themes', {}).then((t) => (themes = t)))

  const themeId = $derived(themeIdOf(app.code))

  function setTheme(e: Event) {
    app.replaceCode(withTheme(app.code, Number((e.target as HTMLSelectElement).value)))
  }

  function pickExample(e: Event) {
    const sel = e.target as HTMLSelectElement
    const ex = examples[Number(sel.value)]
    sel.value = ''
    if (ex) app.replaceCode(ex.code)
  }

  function download(name: string, text: string, type: string) {
    const a = document.createElement('a')
    a.href = URL.createObjectURL(new Blob([text], { type }))
    a.download = name
    a.click()
    URL.revokeObjectURL(a.href)
  }

  async function share() {
    try {
      const { value } = await app.engine.call('encode', { text: app.code })
      const url = shareUrl(location, value)
      await navigator.clipboard.writeText(url)
      app.notice = 'Share link copied'
    } catch {
      app.notice = 'Could not copy the share link'
    }
  }

  async function exportSvg() {
    app.busy = true
    try {
      // No seeds so the file matches the d2 CLI, not the fast preview.
      const r = await app.engine.call('render', { text: app.code })
      if (r.errors.length) app.notice = 'Fix the code error first'
      else download('diagram.svg', r.svg, 'image/svg+xml')
    } finally {
      app.busy = false
    }
  }
</script>

<div class="bar">
  <strong>D2 editor</strong>
  <select aria-label="Theme" value={themeId} onchange={setTheme}>
    <optgroup label="Light">
      {#each themes.filter((t) => !t.dark) as t (t.id)}<option value={t.id}>{t.name}</option>{/each}
    </optgroup>
    <optgroup label="Dark">
      {#each themes.filter((t) => t.dark) as t (t.id)}<option value={t.id}>{t.name}</option>{/each}
    </optgroup>
  </select>
  <select aria-label="Examples" value="" onchange={pickExample}>
    <option value="" disabled>Examples</option>
    {#each examples as ex, i (i)}<option value={i}>{ex.name}</option>{/each}
  </select>
  <button onclick={share}>Share</button>
  <button onclick={() => download('diagram.d2', app.code, 'text/plain')}>Export .d2</button>
  <button onclick={exportSvg}>Export .svg</button>
  {#if app.busy}<span class="badge">working</span>{/if}
  {#if app.notice}<span class="notice">{app.notice}</span>{/if}
  <a href="licenses.html">Licenses</a>
</div>

<style>
  .bar { display: flex; align-items: center; gap: 8px; padding: 6px 10px; border-bottom: 1px solid #ccc; }
  .notice { color: #a00; }
  .badge { color: #666; }
  a { margin-left: auto; }
</style>
