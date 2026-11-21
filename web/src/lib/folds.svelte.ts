import { loadFolded, saveFolded } from './persist'

const storage = (() => {
  try {
    return localStorage
  } catch {
    return null
  }
})()

// Folded side form groups, shared by every panel so one fold applies everywhere.
export const folded = $state(loadFolded(storage))

export function fold(name: string, closed: boolean) {
  if (!!folded[name] === closed) return
  if (closed) folded[name] = true
  else delete folded[name]
  saveFolded(storage, { ...folded })
}
