<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import type { ThemeInfo } from '../lib/protocol'
  import { themeIdOf, withTheme } from '../lib/theme'
  import { currentMode, currentPref, setMode, type ModePref } from '../lib/mode'

  let { app, examples }: { app: AppState; examples: { name: string; code: string }[] } = $props()

  let themes = $state<ThemeInfo[]>([])
  $effect(() => void app.engine.call('themes', {}).then((t) => (themes = t)))

  const themeId = $derived(themeIdOf(app.code))

  // Cycles system, light, dark. The label names what a click switches to.
  const next: Record<ModePref, ModePref> = { system: 'light', light: 'dark', dark: 'system' }
  const modeLabel: Record<ModePref, string> = { system: 'Match system', light: 'Light mode', dark: 'Dark mode' }
  let pref = $state(currentPref())
  function toggleMode() {
    pref = next[pref]
    setMode(pref)
    app.mode = currentMode()
  }

  const sketch = $derived(!!app.render?.config?.sketch)
  const toggleSketch = () => app.applyEdit('setKey', { path: ['vars', 'd2-config', 'sketch'], value: sketch ? null : 'true' })

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
    // Revoking in the same task can cancel the download in Firefox and Safari.
    setTimeout(() => URL.revokeObjectURL(a.href), 1000)
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
    <button class="icon" class:on={sketch} aria-pressed={sketch} title="Sketch style" aria-label="Sketch style" onclick={toggleSketch}>
      <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><path d="M4 20l1.2-4.6L16.5 4.1a2 2 0 0 1 2.8 0l.6.6a2 2 0 0 1 0 2.8L8.6 18.8z M14.5 6l3.5 3.5" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" /><path d="M13 20c2-1 4 1 7-1" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" /></svg>
    </button>
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
      class="grid mono"
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

  <div class="status" role="status" aria-live="polite">
    {#if app.notice}<span class="notice">{app.notice}</span>{/if}
  </div>

  <div class="group">
    <button title="Export .d2" onclick={() => download('diagram.d2', app.code, 'text/plain')}>{@render dl()} .d2</button>
    <button class="primary" title="Export .svg" onclick={exportSvg}>{@render dl()} .svg</button>
  </div>

  <button class="icon" title="{modeLabel[pref]} (click for {modeLabel[next[pref]].toLowerCase()})" aria-label="Color mode: {modeLabel[pref]}" onclick={toggleMode}>
    {#if pref === 'system'}
      <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="12" rx="2" /><path d="M8 20h8M12 16v4" /></g></svg>
    {:else if pref === 'light'}
      <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M2 12h2M20 12h2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" /></g></svg>
    {:else}
      <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" /></svg>
    {/if}
  </button>
  <span class="ver" title="Version">{__VERSION__}</span>
  <a class="ico" href="https://github.com/linogics/d2-tala-editor/blob/main/docs/user-guide.md" target="_blank" rel="noopener" title="Help (user guide)" aria-label="Help (user guide)">
    <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9" /><path d="M9.5 9.5a2.5 2.5 0 0 1 4.9.7c0 1.7-2.4 2.3-2.4 3.8" /><path d="M12 17h.01" /></g></svg>
  </a>
  <a class="ico" href="licenses.html" title="Licenses" aria-label="Licenses">
    <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v18M8 21h8M5 7h14M5 7l-3 7a3 3 0 0 0 6 0zM19 7l-3 7a3 3 0 0 0 6 0z" /></g></svg>
  </a>
  <a class="ico" href="https://github.com/linogics/d2-tala-editor" target="_blank" rel="noopener" title="GitHub repository" aria-label="GitHub repository">
    <svg viewBox="0 0 16 16" width="16" height="16" aria-hidden="true"><path fill="currentColor" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z" /></svg>
  </a>
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
    font-family: var(--mono);
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
  .grid { width: 4.5em; }
  a, .ver { color: var(--muted); font-size: 12px; }
  a:hover { color: var(--text); }
  .ico { display: inline-flex; }
</style>
