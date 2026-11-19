export type Mode = 'light' | 'dark'

const KEY = 'd2-editor:mode'

// The explicit choice wins, the OS preference is the fallback.
export function currentMode(root: HTMLElement = document.documentElement): Mode {
  const m = root.dataset.mode
  if (m === 'light' || m === 'dark') return m

  return matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function setMode(m: Mode, root: HTMLElement = document.documentElement): void {
  root.dataset.mode = m
  try {
    localStorage.setItem(KEY, m)
  } catch {
    // Blocked storage, the choice lasts for this page only.
  }
}
