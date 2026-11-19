<script lang="ts">
  import { isHex, type Field } from '../lib/fields'

  // onset gets the raw value, "" for an unchecked box.
  let { field: fd, value: v, placeholder, base, onset }: { field: Field; value?: string; placeholder?: string; base?: string; onset: (v: string) => void } = $props()

  const val = (e: Event) => onset((e.currentTarget as HTMLInputElement).value)
</script>

{#if fd.kind === 'bool'}
  <label><input type="checkbox" checked={v === 'true'} onchange={(e) => onset(e.currentTarget.checked ? 'true' : '')} /> {fd.name}</label>
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
