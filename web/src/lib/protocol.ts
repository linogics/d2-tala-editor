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
  container: boolean
  x: number
  y: number
  w: number
  h: number
  pinned: boolean
  pinnable: boolean
}

export interface EdgeInfo {
  id: string
  src: string
  dst: string
  points: { x: number; y: number }[]
}

export interface RenderResult {
  svg: string
  shapes: ShapeInfo[]
  edges: EdgeInfo[]
  errors: Diag[]
}

export interface ThemeInfo {
  id: number
  name: string
  dark: boolean
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
  connect: { args: T & { src: string; dst: string }; result: Id }
  reconnect: { args: Id & { src: string | null; dst: string | null }; result: T }
  pin: { args: Id & { x: number; y: number }; result: T }
  unpin: { args: Id; result: T }
  pinAll: { args: T; result: T }
  reorder: { args: Before; result: T }
  reorderEdge: { args: Before; result: T }
  format: { args: T; result: T }
  completions: { args: T & { line: number; col: number }; result: CompletionItem[] }
  refRanges: { args: Id; result: Range[] }
  encode: { args: T; result: { value: string } }
  decode: { args: { value: string }; result: T }
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
