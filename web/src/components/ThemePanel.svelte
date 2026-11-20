<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import type { ThemeInfo } from '../lib/protocol'
  import type { Field } from '../lib/fields'
  import FieldInput from './FieldInput.svelte'

  let { state: app }: { state: AppState } = $props()

  let themes = $state<ThemeInfo[]>([])
  $effect(() => void app.engine.call('themes', {}).then((t) => (themes = t)))

  const groups: [string, string[]][] = [
    ['Neutrals (text, background)', ['N1', 'N2', 'N3', 'N4', 'N5', 'N6', 'N7']],
    ['Base (borders, lines, fills)', ['B1', 'B2', 'B3', 'B4', 'B5', 'B6']],
    ['Alternative (special shapes)', ['AA2', 'AA4', 'AA5', 'AB4', 'AB5']],
  ]
  // What each slot paints by default, taken from D2 v0.9.0 (d2graph GetFill/GetStroke, d2exporter, d2svg).
  const roles: Record<string, string> = {
    N1: 'Text and labels',
    N2: 'Muted text, SQL column types',
    N3: 'Markdown neutral',
    N4: 'Diamond fill',
    N5: 'Queue, parallelogram, hexagon fill',
    N6: 'Markdown subtle background',
    N7: 'Background, cloud and callout fill',
    B1: 'Borders and connection lines',
    B2: 'Dashed lines, SQL and class field names',
    B3: 'Person fill',
    B4: 'Container fill',
    B5: 'Nested container fill, sequence actors',
    B6: 'Shape fill',
    AA2: 'SQL constraints, class methods',
    AA4: 'Cylinder, stored data, package fill',
    AA5: 'Nested cylinder, stored data, package fill',
    AB4: 'Step, page, document fill',
    AB5: 'Nested step, page, document fill',
  }
  const field = (slot: string): Field => ({ attr: slot, name: `${roles[slot]} (${slot})`, kind: 'color', group: 'Theme' })

  const dark = $derived(app.mode === 'dark')
  const key = $derived(dark ? 'dark-theme-overrides' : 'theme-overrides')
  const cfg = $derived(app.render?.config)
  const overrides = $derived((dark ? cfg?.darkThemeOverrides : cfg?.themeOverrides) ?? {})
  const themeId = $derived((dark ? cfg?.darkThemeID : cfg?.themeID) ?? 0)
  const base = $derived(themes.find((t) => t.id === themeId)?.colors ?? {})
  const any = $derived(Object.values(overrides).some((v) => v != null))

  const set = (slot: string, v: string) => app.applyEdit('setKey', { path: ['vars', 'd2-config', key, slot], value: v.trim() || null })
  const reset = () => app.applyEdit('setKey', { path: ['vars', 'd2-config', key], value: null })
</script>

<div class="form">
  <p>Editing {dark ? 'dark' : 'light'} mode colors. Switch the app mode to edit the other set.</p>
  {#if dark && cfg?.darkThemeID == null}
    <p class="warn">Dark overrides only apply when the diagram sets a dark theme (dark-theme-id).</p>
  {/if}
  {#each groups as [name, slots] (name)}
    <h4>{name}</h4>
    {#each slots as slot (slot)}
      <FieldInput field={field(slot)} value={overrides[slot.toLowerCase()] ?? undefined} placeholder={base[slot] ?? 'unset'} base={base[slot]} onset={(v) => set(slot, v)} />
    {/each}
  {/each}
  <button disabled={!any} onclick={reset}>Reset all</button>
</div>

<style>
  p { color: var(--muted); margin: 0 0 12px; }
  .warn { color: var(--warn-fg); background: var(--warn-bg); padding: 6px 8px; border-radius: 6px; }
  h4 { margin: 16px 0 8px; padding-top: 12px; border-top: 1px solid var(--border); font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; color: var(--text); }
</style>
