<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import { classFields } from '../lib/fields'
  import { templates } from '../lib/templates'
  import FieldInput from './FieldInput.svelte'

  let { state: app }: { state: AppState } = $props()

  const classes = $derived(app.render?.classes ?? [])
  let picked = $state('')
  const cls = $derived(classes.find((c) => c.name.toLowerCase() === picked.toLowerCase()))

  async function add(e: Event) {
    const input = e.currentTarget as HTMLInputElement
    const name = input.value.trim()
    if (!name) return
    if (!classes.some((c) => c.name.toLowerCase() === name.toLowerCase())) {
      if (!(await app.applyEdit('setKey', { path: ['classes', name], value: '' }))) return
    }
    input.value = ''
    picked = name
  }

  async function rename(e: Event) {
    const to = (e.currentTarget as HTMLInputElement).value.trim()
    if (!to || to === cls!.name) return
    if (await app.applyEdit('renameClass', { from: cls!.name, to })) picked = to
  }

  function set(attr: string, v: string) {
    const name = cls!.name
    const value = v.trim() || null
    const path = ['classes', name, ...attr.split('.')]
    // Clearing the last attr would drop the class, so it is written back as a bare key.
    if (!value && Object.keys(cls!.attrs).every((k) => k === attr)) {
      return app.applyEdits([
        ['setKey', { path, value: null }],
        ['setKey', { path: ['classes', name], value: '' }],
      ])
    }

    return app.applyEdit('setKey', { path, value })
  }

  async function del() {
    if (await app.applyEdit('setKey', { path: ['classes', cls!.name], value: null })) picked = ''
  }
</script>

<datalist id="shapes">{#each templates as t (t.key)}<option value={t.key}></option>{/each}</datalist>

<div class="form">
  <label>New class <input placeholder="name, then Enter" onchange={add} /></label>
  {#if classes.length}
    <div class="list" role="listbox" aria-label="Classes">
      {#each classes as c (c.name)}
        <button role="option" aria-selected={c === cls} class:on={c === cls} onclick={() => (picked = c.name)}>{c.name}</button>
      {/each}
    </div>
  {:else}
    <p>No classes yet. A class is a reusable set of styles for shapes and edges.</p>
  {/if}
  {#if cls}
    {#key cls.name}
      <label>Name <input value={cls.name} onchange={rename} /></label>
      {#each classFields as fd (fd.attr)}
        <FieldInput field={fd} value={cls.attrs[fd.attr]} placeholder={fd.kind === 'color' ? 'unset' : undefined} onset={(v) => set(fd.attr, v)} />
      {/each}
      <button class="danger" onclick={del}>Delete class</button>
    {/key}
  {/if}
</div>

<style>
  p { color: var(--muted); margin: 0 0 12px; }
  .list { display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 12px; }
  .list button { padding: 2px 8px; font-size: 12px; }
  .danger { margin-top: 8px; }
</style>
