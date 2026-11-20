<script lang="ts">
  import { enterCommits, isHex, type Field } from '../lib/fields'

  // onset gets the raw value, "" for an unchecked box. A falsy result means the edit was refused.
  let { field: fd, value: v, placeholder, base, disabled, onset }: { field: Field; value?: string; placeholder?: string; base?: string; disabled?: boolean; onset: (v: string) => unknown } = $props()

  // A refused edit leaves the value unchanged, so the control is put back to show it.
  async function send(el: HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement, value: string) {
    if (await onset(value)) return
    if (el instanceof HTMLInputElement && el.type === 'checkbox') el.checked = v === 'true'
    else if (el.type !== 'color') el.value = v ?? ''
  }

  const val = (e: Event) => {
    const el = e.currentTarget as HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement
    return send(el, el.value)
  }
  // A value outside the list, typed in the code, stays selectable.
  const opts = $derived(v && !fd.options?.includes(v) ? [v, ...(fd.options ?? [])] : (fd.options ?? []))
</script>

{#if fd.kind === 'bool'}
  <label><input type="checkbox" {disabled} checked={v === 'true'} onchange={(e) => send(e.currentTarget, e.currentTarget.checked ? 'true' : '')} /> {fd.name}</label>
{:else if fd.kind === 'color'}
  <label>
    {fd.name}
    <span class="color">
      <input value={v ?? ''} {placeholder} {disabled} onchange={val} />
      <input type="color" {disabled} aria-label="{fd.name} picker" value={isHex(v) ? v : isHex(base) ? base : '#000000'} onchange={val} />
    </span>
  </label>
{:else if fd.kind === 'select'}
  <label>
    {fd.name}
    <select value={v ?? ''} {disabled} onchange={val}>
      <option value="">{placeholder ?? '(unset)'}</option>
      {#each opts as o (o)}<option value={o}>{o}</option>{/each}
    </select>
  </label>
{:else if fd.kind === 'text'}
  <label>
    {fd.name}
    <textarea rows="1" value={v ?? ''} {placeholder} {disabled} onkeydown={enterCommits} onchange={val}></textarea>
  </label>
{:else}
  <label>
    {fd.name}
    <input type={fd.kind} min={fd.min} max={fd.max} step={fd.step} value={v ?? ''} {placeholder} {disabled} onchange={val} />
  </label>
{/if}
