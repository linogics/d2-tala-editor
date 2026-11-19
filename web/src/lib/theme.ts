const RE = /theme-id:\s*(\d+)/

export const themeIdOf = (code: string): number => Number(RE.exec(code)?.[1] ?? 0)

// The oracle cannot set vars keys, so the theme is edited as text.
// shortcut: appends a second vars block when d2-config exists without theme-id, D2 merges them.
export function withTheme(code: string, id: number): string {
  if (RE.test(code)) return code.replace(RE, `theme-id: ${id}`)

  return `${code.replace(/\n*$/, '\n')}vars: {\n  d2-config: {\n    theme-id: ${id}\n  }\n}\n`
}
