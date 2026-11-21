import raw from '../../public/templates.json'

export interface Template {
  key: string
  label: string
  group: 'shape' | 'special'
  template: string
}

// A-Z by label, for the palette and the shape pickers.
export const templates = [...(raw as Template[])].sort((a, b) => a.label.localeCompare(b.label))
