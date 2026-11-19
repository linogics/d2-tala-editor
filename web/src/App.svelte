<script lang="ts">
  import { createEngine, createWorker } from './lib/engine'

  const engine = createEngine(createWorker())
  let ready = $state(false)
  engine.ready.then(() => (ready = true))
</script>

<p>{ready ? 'ready' : 'loading'}</p>
