const KEY = 'd2-editor:code'

export function loadInitial(storage: Storage | null, firstExample: string): string {
  try {
    const saved = storage?.getItem(KEY)
    if (saved != null) return saved
  } catch {
    // Storage blocked, fall through.
  }

  return firstExample
}

export function saveCode(storage: Storage | null, text: string): void {
  try {
    storage?.setItem(KEY, text)
  } catch {
    // Quota or blocked storage, autosave is best effort.
  }
}

const GRID_KEY = 'd2-editor:grid'

export function loadGrid(storage: Storage | null): { snap: boolean; grid: number } {
  try {
    const v = JSON.parse(storage?.getItem(GRID_KEY) ?? 'null')
    if (typeof v?.snap === 'boolean' && v.grid >= 2) return { snap: v.snap, grid: v.grid }
  } catch {
    // Bad or blocked storage, use defaults.
  }

  return { snap: false, grid: 20 }
}

export function saveGrid(storage: Storage | null, v: { snap: boolean; grid: number }): void {
  try {
    storage?.setItem(GRID_KEY, JSON.stringify(v))
  } catch {
    // Best effort.
  }
}
