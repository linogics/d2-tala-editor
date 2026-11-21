export type Mode = 'light' | 'dark'
// What the user picked. system follows the OS.
export type ModePref = Mode | 'system'

const KEY = 'd2-editor:mode'

export function currentPref(root: HTMLElement = document.documentElement): ModePref {
  const m = root.dataset.mode

  return m === 'light' || m === 'dark' ? m : 'system'
}

// The explicit choice wins, the OS preference is the fallback.
export function currentMode(root: HTMLElement = document.documentElement): Mode {
  const p = currentPref(root)
  if (p !== 'system') return p

  return matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export function setMode(p: ModePref, root: HTMLElement = document.documentElement): void {
  if (p === 'system') delete root.dataset.mode
  else root.dataset.mode = p
  try {
    if (p === 'system') localStorage.removeItem(KEY)
    else localStorage.setItem(KEY, p)
  } catch {
    // Blocked storage, the choice lasts for this page only.
  }
}
