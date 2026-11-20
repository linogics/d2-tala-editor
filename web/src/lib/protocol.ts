export interface Diag {
  line: number
  col: number
  message: string
}

export interface Range {
  fromLine: number
  fromCol: number
  toLine: number
  toCol: number
}

export interface ShapeInfo {
  id: string
  parent: string
  type: string
  container: string
  x: number
  y: number
  w: number
  h: number
  pinned: boolean
  pinnable: boolean
  // D2 attributes set on the shape, keyed by path such as "style.fill".
  attrs: Record<string, string>
}

export interface EdgeInfo {
  id: string
  src: string
  dst: string
  points: [number, number][]
  attrs: Record<string, string>
  srcArrow: boolean
  dstArrow: boolean
}

export interface RenderResult {
  svg: string
  shapes: ShapeInfo[]
  edges: EdgeInfo[]
  errors: Diag[]
  // Compiled vars.d2-config. Override keys are lowercase, such as "b1".
  config: {
    sketch: boolean | null
    themeID: number | null
    darkThemeID: number | null
    themeOverrides?: Record<string, string | null>
    darkThemeOverrides?: Record<string, string | null>
  } | null
  classes: ClassInfo[]
}

export interface ClassInfo {
  name: string
  attrs: Record<string, string>
}

export interface ThemeInfo {
  id: number
  name: string
  dark: boolean
  // Override slot, such as "B1", to the theme's own color.
  colors: Record<string, string>
}

export interface CompletionItem {
  Label: string
  Kind: string
  Detail: string
  InsertText: string
}

type T = { text: string }
type Id = T & { id: string }
type Before = Id & { before: string }

export interface Ops {
  render: { args: T & { seeds?: number[] }; result: RenderResult }
  create: { args: T & { parent: string; shape: string }; result: Id }
  insert: { args: T & { parent: string; template: string }; result: Id }
  set: { args: Id & { attr: string; value: string | null }; result: T }
  delete: { args: Id; result: T }
  move: { args: Id & { parent: string }; result: Id }
  rename: { args: Id & { name: string }; result: Id }
  connect: { args: T & { src: string; dst: string; op?: '->' | '<-' | '<->' | '--' }; result: Id }
  // Picks the op by which ends have a head, the edge id changes with it.
  setArrows: { args: Id & { src: boolean; dst: boolean }; result: Id }
  reconnect: { args: Id & { src: string | null; dst: string | null }; result: T }
  pin: { args: Id & { x: number; y: number }; result: T }
  unpin: { args: Id; result: T }
  pinAll: { args: T; result: T }
  reorder: { args: Before; result: T }
  reorderEdge: { args: Before; result: T }
  format: { args: T; result: T }
  completions: { args: T & { line: number; col: number }; result: CompletionItem[] }
  refRanges: { args: Id; result: Range[] }
  // Sets the scalar at a vars or classes path, null deletes it, "" writes a bare key.
  setKey: { args: T & { path: string[]; value: string | null }; result: T }
  renameClass: { args: T & { from: string; to: string }; result: T }
  // Deletes a class definition and drops it from every shape and edge using it.
  deleteClass: { args: T & { name: string }; result: T }
  themes: { args: Record<string, never>; result: ThemeInfo[] }
}

export type Op = keyof Ops
export type ArgsOf<O extends Op> = Ops[O]['args']
export type ResultOf<O extends Op> = Ops[O]['result']

export class EngineError extends Error {
  code: string
  diag?: Diag

  constructor(code: string, message: string, diag?: Diag | null) {
    super(message)
    this.code = code
    this.diag = diag ?? undefined
  }
}
