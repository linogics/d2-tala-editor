import type { ShapeInfo } from './protocol'
import type { Step } from './state.svelte'
import { templates } from './templates'

export type Kind = 'text' | 'number' | 'color' | 'bool' | 'url' | 'select'
export type Field = { attr: string; name: string; kind: Kind; group: string; options?: string[]; min?: number; max?: number; step?: number }

export const nears = ['top-left', 'top-center', 'top-right', 'center-left', 'center-right', 'bottom-left', 'bottom-center', 'bottom-right']
export const arrowheads = ['triangle', 'arrow', 'diamond', 'circle', 'box', 'cross', 'cf-one', 'cf-one-required', 'cf-many', 'cf-many-required']

const f = (group: string, attr: string, name: string, kind: Kind = 'text', more: Partial<Field> = {}): Field => ({ attr, name, kind, group, ...more })
const bools = (group: string, keys: string[]) => keys.map((k) => f(group, `style.${k}`, k, 'bool'))
const opacity = (group: string) => f(group, 'style.opacity', 'Opacity', 'number', { min: 0, max: 1, step: 0.1 })

const stroke = [
  f('Stroke', 'style.stroke', 'Stroke', 'color'),
  f('Stroke', 'style.stroke-width', 'Stroke width', 'number', { min: 0, max: 15 }),
  f('Stroke', 'style.stroke-dash', 'Stroke dash', 'number', { min: 0, max: 10 }),
]
const font = [f('Font', 'style.font-color', 'Font color', 'color'), f('Font', 'style.font-size', 'Font size', 'number', { min: 8, max: 100 })]

// Field order sets the order of the groups in the side forms.
export const shapeFields = [
  f('Position', 'top', 'Top', 'number', { min: 0 }),
  f('Position', 'left', 'Left', 'number', { min: 0 }),
  f('Position', 'width', 'Width', 'number', { min: 1 }),
  f('Position', 'height', 'Height', 'number', { min: 1 }),
  f('Position', 'near', 'Near', 'select', { options: nears }),
  f('Grid', 'grid-rows', 'Rows', 'number', { min: 1 }),
  f('Grid', 'grid-columns', 'Columns', 'number', { min: 1 }),
  f('Grid', 'grid-gap', 'Gap', 'number', { min: 0 }),
  f('Grid', 'vertical-gap', 'Vertical gap', 'number', { min: 0 }),
  f('Grid', 'horizontal-gap', 'Horizontal gap', 'number', { min: 0 }),
  f('Fill', 'style.fill', 'Fill', 'color'),
  opacity('Fill'),
  f('Fill', 'style.border-radius', 'Border radius', 'number', { min: 0 }),
  ...stroke,
  ...font,
  ...bools('Font', ['bold', 'italic', 'underline']),
  ...bools('Effects', ['shadow', '3d', 'multiple', 'double-border']),
  f('Extras', 'icon', 'Icon URL', 'url'),
  f('Extras', 'tooltip', 'Tooltip'),
  f('Extras', 'link', 'Link', 'url'),
]

// The groups of fields, in order of first use.
export const groupsOf = (fields: Field[]) => [...new Set(fields.map((x) => x.group))]

// D2 refuses these styles on other shapes, see d2compiler validateKey. Mirrors styleShapes in wasm/editor/edit.go.
const only: Record<string, string[]> = {
  'style.3d': ['rectangle', 'square', 'hexagon'],
  'style.double-border': ['rectangle', 'square', 'circle', 'oval'],
}

// The shape fields every one of types accepts.
export const shapeFieldsFor = (types: string[]) => shapeFields.filter((x) => types.every((t) => only[x.attr]?.includes(t) ?? true))

// The edits that set attr of shape s. Top and left only work as a pair, so setting one
// pins the shape first to keep the other where it is, and clearing one unpins it.
export function fieldSteps(s: Pick<ShapeInfo, 'id' | 'x' | 'y' | 'attrs'>, attr: string, value: string | null): Step[] {
  const other = { top: 'left', left: 'top' }[attr]
  if (!other) return [['set', { id: s.id, attr, value }]]
  if (value === null) return [['unpin', { id: s.id }]]
  const pin: Step[] = s.attrs[other] == null ? [['pin', { id: s.id, x: Math.round(s.x), y: Math.round(s.y) }]] : []

  return [...pin, ['set', { id: s.id, attr, value }]]
}

// One arrow end, shown in its own box.
export const headFields = (end: string): Field[] => [
  f('Arrows', `${end}-arrowhead.shape`, 'Shape', 'select', { options: arrowheads }),
  f('Arrows', `${end}-arrowhead.label`, 'Label'),
]
const arrowFields = ['source', 'target'].flatMap(headFields)

export const edgeFields = [...stroke, opacity('Stroke'), ...font, ...bools('Font', ['bold', 'italic']), ...arrowFields, ...bools('Effects', ['animated'])]

// A class may style shapes and edges, so it offers both sets without the per object position.
export const classFields = [
  f('General', 'label', 'Label'),
  f('General', 'shape', 'Shape', 'select', { options: templates.map((t) => t.key) }),
  ...arrowFields,
  ...shapeFields.filter((x) => !['top', 'left', 'near'].includes(x.attr)).map((x) => (x.group === 'Position' ? { ...x, group: 'Size' } : x)),
  ...bools('Effects', ['animated']),
]

// Text fields are textareas: Enter applies the value, Shift+Enter adds a line break.
export function enterCommits(e: KeyboardEvent) {
  if (e.key !== 'Enter' || e.shiftKey || e.isComposing) return
  e.preventDefault()
  ;(e.currentTarget as HTMLElement).blur()
}

export const isHex = (v: string | undefined) => /^#[0-9a-f]{6}$/i.test(v ?? '')
