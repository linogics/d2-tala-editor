import raw from '../../public/templates.json'

export interface Template {
  key: string
  label: string
  group: 'shape' | 'special'
  template: string
}

export const templates = raw as Template[]
