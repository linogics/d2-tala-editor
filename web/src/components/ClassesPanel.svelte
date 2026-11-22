<script lang="ts">
  import type { AppState } from '../lib/state.svelte'
  import { classFields, groupsOf, headFields, type Field } from '../lib/fields'
  import FieldInput from './FieldInput.svelte'
  import Group from './Group.svelte'

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
    const el = e.currentTarget as HTMLInputElement
    const to = el.value.trim()
    // A blank or refused name puts the current one back.
    if (to && to !== cls!.name && (await app.applyEdit('renameClass', { from: cls!.name, to }))) picked = to
    else el.value = cls!.name
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
    if (await app.applyEdit('deleteClass', { name: cls!.name })) picked = ''
  }
</script>


<div class="form">
  <label>New class <input placeholder="name, then Enter" onchange={add} /></label>
  {#if classes.length}
    <div class="list" role="listbox" aria-label="Classes">
      {#each classes as c (c.name)}
        <button class="mono" role="option" aria-selected={c === cls} class:on={c === cls} onclick={() => (picked = c.name)}>{c.name}</button>
      {/each}
    </div>
  {:else}
    <p>No classes yet. A class is a reusable set of styles for shapes and edges.</p>
  {/if}
  {#if cls}
    {#key cls.name}
      {#snippet input(fd: Field)}
        <FieldInput field={fd} value={cls!.attrs[fd.attr]} placeholder={fd.kind === 'color' ? 'unset' : undefined} onset={(v) => set(fd.attr, v)} />
      {/snippet}
      {#each groupsOf(classFields) as g (g)}
        {#if g === 'Arrows'}
          <Group name={g} plain>
            {#each ['source', 'target'] as end (end)}
              <div class="head" role="group" aria-label="{end === 'source' ? 'Source' : 'Target'} head">
                <span class="end">{end} head</span>
                {#each headFields(end) as fd (fd.attr)}{@render input(fd)}{/each}
              </div>
            {/each}
          </Group>
        {:else}
          <Group name={g}>
            {#if g === 'General'}<label>Name <input class="mono" value={cls.name} onchange={rename} /></label>{/if}
            {#each classFields.filter((x) => x.group === g) as fd (fd.attr)}{@render input(fd)}{/each}
          </Group>
        {/if}
      {/each}
      <button class="danger" onclick={del}>Delete class</button>
    {/key}
  {/if}
</div>

<style>
  p { color: var(--muted); margin: 0 0 12px; }
  .list { display: flex; flex-wrap: wrap; gap: 4px; margin-bottom: 12px; }
  .list button { padding: 2px 8px; font-size: 0.8125rem; }
  .danger { margin-top: 8px; }
</style>
