<script lang="ts">
  import type { Snippet } from 'svelte'
  import { fold, folded } from '../lib/folds.svelte'

  // plain drops the box, for content that brings its own boxes.
  let { name, plain = false, children }: { name: string; plain?: boolean; children: Snippet } = $props()
</script>

<details class="group" open={!folded[name]} ontoggle={(e) => fold(name, !e.currentTarget.open)}>
  <summary>{name}</summary>
  <div class:box={!plain}>{@render children()}</div>
</details>

<style>
  .group { margin-bottom: 12px; }
  summary { cursor: pointer; margin-bottom: 6px; font-size: 11px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--muted); }
  .box { border: 1px solid var(--border); border-radius: 6px; padding: 8px 8px 0; }
</style>
