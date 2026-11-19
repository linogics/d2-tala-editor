<script lang="ts">
  import type { AppState } from '../lib/state.svelte'

  let { state }: { state: AppState } = $props()
</script>

<div class="box">
  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
  {@html state.render?.svg ?? ''}
</div>

<style>
  .box { height: 100%; overflow: auto; border-right: 1px solid #ccc; }
</style>
