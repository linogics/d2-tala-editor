import type { ShapeInfo } from './protocol'
import type { Step } from './state.svelte'

export type Kind = 'text' | 'number' | 'color' | 'bool' | 'url'
export type Field = { attr: string; name: string; kind: Kind; list?: string; min?: number; max?: number; step?: number }

const f = (attr: string, name: string, kind: Kind = 'text', more: Partial<Field> = {}): Field => ({ attr, name, kind, ...more })

const stroke = [
  f('style.stroke', 'Stroke', 'color'),
  f('style.stroke-width', 'Stroke width', 'number', { min: 0, max: 15 }),
  f('style.stroke-dash', 'Stroke dash', 'number', { min: 0, max: 10 }),
  f('style.opacity', 'Opacity', 'number', { min: 0, max: 1, step: 0.1 }),
  f('style.font-size', 'Font size', 'number', { min: 8, max: 100 }),
  f('style.font-color', 'Font color', 'color'),
]

export const shapeFields = [
  f('top', 'Top', 'number', { min: 0 }),
  f('left', 'Left', 'number', { min: 0 }),
  f('width', 'Width', 'number', { min: 1 }),
  f('height', 'Height', 'number', { min: 1 }),
  f('style.fill', 'Fill', 'color'),
  ...stroke,
  f('style.border-radius', 'Border radius', 'number', { min: 0 }),
  ...['shadow', '3d', 'multiple', 'bold', 'italic', 'underline', 'double-border'].map((k) => f(`style.${k}`, k, 'bool')),
  f('icon', 'Icon URL', 'url'),
  f('tooltip', 'Tooltip'),
  f('link', 'Link', 'url'),
  f('near', 'Near', 'text', { list: 'near' }),
]

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

const arrowFields = ['source', 'target'].flatMap((end) => [
  f(`${end}-arrowhead.shape`, `${end} arrowhead`, 'text', { list: 'arrowheads' }),
  f(`${end}-arrowhead.label`, `${end} arrowhead label`),
])

export const edgeFields = [...stroke, ...arrowFields, ...['animated', 'bold', 'italic'].map((k) => f(`style.${k}`, k, 'bool'))]

// A class may style shapes and edges, so it offers both sets without the per object position.
export const classFields = [
  f('label', 'Label'),
  f('shape', 'Shape', 'text', { list: 'shapes' }),
  ...shapeFields.filter((x) => !['top', 'left', 'near'].includes(x.attr)),
  ...arrowFields,
  f('style.animated', 'animated', 'bool'),
]

export const nears = ['top-left', 'top-center', 'top-right', 'center-left', 'center-right', 'bottom-left', 'bottom-center', 'bottom-right']
export const arrowheads = ['triangle', 'arrow', 'diamond', 'circle', 'box', 'cross', 'cf-one', 'cf-one-required', 'cf-many', 'cf-many-required']

export const isHex = (v: string | undefined) => /^#[0-9a-f]{6}$/i.test(v ?? '')
