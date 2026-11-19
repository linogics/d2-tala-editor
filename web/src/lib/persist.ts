const KEY = 'd2-editor:code'
const PREFIX = '#d2='

export async function loadInitial(
  location: Location,
  storage: Storage | null,
  decode: (s: string) => Promise<string>,
  firstExample: string,
): Promise<string> {
  try {
    if (location.hash.startsWith(PREFIX)) return await decode(location.hash.slice(PREFIX.length))
  } catch {
    // Bad share link, fall through.
  }

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

export function shareUrl(location: Location, value: string): string {
  return location.href.split('#')[0] + PREFIX + value
}
