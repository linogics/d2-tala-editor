<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import type { ThemeInfo } from '../lib/protocol'
  import { themeIdOf, withTheme } from '../lib/theme'
  import { currentMode, setMode, type Mode } from '../lib/mode'

  let { app, examples }: { app: AppState; examples: { name: string; code: string }[] } = $props()

  let themes = $state<ThemeInfo[]>([])
  $effect(() => void app.engine.call('themes', {}).then((t) => (themes = t)))

  const themeId = $derived(themeIdOf(app.code))

  let mode = $state<Mode>(currentMode())
  function toggleMode() {
    mode = mode === 'dark' ? 'light' : 'dark'
    setMode(mode)
  }

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

{#snippet dl()}
  <svg viewBox="0 0 24 24" width="15" height="15" aria-hidden="true"><path d="M12 4v11m0 0-4-4m4 4 4-4M5 19h14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" /></svg>
{/snippet}

<header class="bar">
  <strong class="brand">d2-tala-editor</strong>

  <div class="group">
    <select aria-label="Diagram theme" title="Diagram theme" value={themeId} onchange={setTheme}>
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
  </div>

  <div class="group">
    <button class="icon" class:on={app.snap} aria-pressed={app.snap} title="Grid and snap" aria-label="Grid and snap" onclick={() => (app.snap = !app.snap)}>
      <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><path d="M6 3v8a6 6 0 0 0 12 0V3h-4v8a2 2 0 0 1-4 0V3z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" /><path d="M6 7h4M14 7h4" stroke="currentColor" stroke-width="1.8" /></svg>
    </button>
    <input
      class="grid"
      type="number"
      min="2"
      aria-label="Grid size"
      title="Grid size"
      value={app.grid}
      onchange={(e) => {
        const n = Math.round(Number((e.target as HTMLInputElement).value))
        if (n >= 2) app.grid = n
        else (e.target as HTMLInputElement).value = String(app.grid)
      }}
    />
  </div>

  <div class="status">
    {#if app.busy}<span class="badge">Working…</span>{/if}
    {#if app.notice}<span class="notice">{app.notice}</span>{/if}
  </div>

  <div class="group">
    <button title="Export .d2" onclick={() => download('diagram.d2', app.code, 'text/plain')}>{@render dl()} .d2</button>
    <button class="primary" title="Export .svg" onclick={exportSvg}>{@render dl()} .svg</button>
  </div>

  <button class="icon" title={mode === 'dark' ? 'Light mode' : 'Dark mode'} aria-label={mode === 'dark' ? 'Light mode' : 'Dark mode'} onclick={toggleMode}>
    {#if mode === 'dark'}
      <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M2 12h2M20 12h2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" /></g></svg>
    {:else}
      <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" /></svg>
    {/if}
  </button>
  <a href="licenses.html">Licenses</a>
</header>

<style>
  .bar {
    display: flex;
    align-items: center;
    gap: 12px;
    height: 48px;
    padding: 0 14px;
    background: var(--surface);
    border-bottom: 1px solid var(--border);
  }
  .brand {
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: var(--primary);
    margin-right: 4px;
  }
  .group { display: flex; align-items: center; gap: 6px; }
  .group + .group { padding-left: 12px; border-left: 1px solid var(--border); }
  .status { flex: 1; display: flex; justify-content: flex-end; gap: 8px; min-width: 0; }
  .notice { color: var(--danger); background: var(--danger-bg); padding: 3px 10px; border-radius: 999px; font-size: 12px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .badge { color: var(--warn-fg); background: var(--warn-bg); padding: 3px 10px; border-radius: 999px; font-size: 12px; }
  .grid { width: 4.5em; }
  a { color: var(--muted); font-size: 12px; }
  a:hover { color: var(--text); }
</style>
