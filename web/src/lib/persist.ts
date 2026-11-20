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

const COLS_KEY = 'd2-editor:cols'

// Column widths in px, null keeps the default width.
export type Cols = { side: number | null; code: number | null }

export function loadCols(storage: Storage | null): Cols {
  const ok = (n: unknown) => (typeof n === 'number' && n > 0 ? n : null)
  try {
    const v = JSON.parse(storage?.getItem(COLS_KEY) ?? 'null')

    return { side: ok(v?.side), code: ok(v?.code) }
  } catch {
    return { side: null, code: null }
  }
}

export function saveCols(storage: Storage | null, v: Cols): void {
  try {
    storage?.setItem(COLS_KEY, JSON.stringify(v))
  } catch {
    // Best effort.
  }
}

const FOLD_KEY = 'd2-editor:folded'

// Side form groups the user folded, by group name.
export function loadFolded(storage: Storage | null): Record<string, boolean> {
  try {
    const v = JSON.parse(storage?.getItem(FOLD_KEY) ?? 'null')
    if (v && typeof v === 'object' && !Array.isArray(v)) return Object.fromEntries(Object.entries(v).filter(([, b]) => b === true)) as Record<string, boolean>
  } catch {
    // Bad or blocked storage, all open.
  }

  return {}
}

export function saveFolded(storage: Storage | null, v: Record<string, boolean>): void {
  try {
    storage?.setItem(FOLD_KEY, JSON.stringify(v))
  } catch {
    // Best effort.
  }
}
