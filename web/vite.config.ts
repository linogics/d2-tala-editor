import { defineConfig } from 'vitest/config'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  base: './',
  plugins: [svelte()],
  // CI sets VERSION to the release tag.
  define: { __VERSION__: JSON.stringify(process.env.VERSION || 'dev') },
  // Svelte must resolve its browser build for mount() in jsdom tests.
  ...(process.env.VITEST ? { resolve: { conditions: ['browser'] } } : {}),
})
