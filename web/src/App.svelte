<script lang="ts">
  import { createEngine } from './lib/engine'

  const engine = createEngine(
    new Worker(new URL('./lib/worker.ts', import.meta.url), { type: 'classic' }),
  )
  let ready = $state(false)
  engine.ready.then(() => (ready = true))
</script>

<p>{ready ? 'ready' : 'loading'}</p>
