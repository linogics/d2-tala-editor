import { defineConfig } from 'vitest/config'
import { svelte } from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  base: './',
  plugins: [svelte()],
  // Svelte must resolve its browser build for mount() in jsdom tests.
  ...(process.env.VITEST ? { resolve: { conditions: ['browser'] } } : {}),
})
