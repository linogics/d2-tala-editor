<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import type { ThemeInfo } from '../lib/protocol'
  import type { Field } from '../lib/fields'
  import FieldInput from './FieldInput.svelte'

  let { state: app }: { state: AppState } = $props()

  let themes = $state<ThemeInfo[]>([])
  $effect(() => void app.engine.call('themes', {}).then((t) => (themes = t)))

  const groups: [string, string[]][] = [
    ['Neutrals (text, borders, background)', ['N1', 'N2', 'N3', 'N4', 'N5', 'N6', 'N7']],
    ['Base (shape fills by depth)', ['B1', 'B2', 'B3', 'B4', 'B5', 'B6']],
    ['Alternative', ['AA2', 'AA4', 'AA5', 'AB4', 'AB5']],
  ]
  const field = (slot: string): Field => ({ attr: slot, name: slot, kind: 'color' })

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
  h4 { margin: 14px 0 8px; font-size: 11px; text-transform: uppercase; letter-spacing: 0.04em; color: var(--muted); }
</style>
