<script lang="ts">
  import { isHex, type Field } from '../lib/fields'

  // onset gets the raw value, "" for an unchecked box. A falsy result means the edit was refused.
  let { field: fd, value: v, placeholder, base, onset }: { field: Field; value?: string; placeholder?: string; base?: string; onset: (v: string) => unknown } = $props()

  // A refused edit leaves the value unchanged, so the control is put back to show it.
  async function send(el: HTMLInputElement, value: string) {
    if (await onset(value)) return
    if (el.type === 'checkbox') el.checked = v === 'true'
    else if (el.type !== 'color') el.value = v ?? ''
  }

  const val = (e: Event) => send(e.currentTarget as HTMLInputElement, (e.currentTarget as HTMLInputElement).value)
</script>

{#if fd.kind === 'bool'}
  <label><input type="checkbox" checked={v === 'true'} onchange={(e) => send(e.currentTarget, e.currentTarget.checked ? 'true' : '')} /> {fd.name}</label>
{:else if fd.kind === 'color'}
  <label>
    {fd.name}
    <span class="color">
      <input value={v ?? ''} {placeholder} onchange={val} />
      <input type="color" aria-label="{fd.name} picker" value={isHex(v) ? v : isHex(base) ? base : '#000000'} onchange={val} />
    </span>
  </label>
{:else}
  <label>
    {fd.name}
    <input type={fd.kind} list={fd.list} min={fd.min} max={fd.max} step={fd.step} value={v ?? ''} {placeholder} onchange={val} />
  </label>
{/if}
