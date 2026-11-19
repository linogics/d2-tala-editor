import { StreamLanguage, type StringStream } from '@codemirror/language'

const KEYWORDS = new Set([
  'shape', 'style', 'label', 'icon', 'near', 'top', 'left', 'vars',
  'grid-rows', 'grid-columns', 'direction', 'classes',
])

interface S {
  close: string | null
}

function token(stream: StringStream, state: S): string | null {
  if (state.close) {
    while (!stream.eol()) {
      if (stream.match(state.close)) {
        state.close = null
        break
      }
      stream.next()
    }

    return 'string'
  }

  if (stream.eatSpace()) return null
  if (stream.match('#')) {
    stream.skipToEnd()

    return 'comment'
  }

  const open = stream.match(/^(\|+)(`*)[A-Za-z]*/) as RegExpMatchArray | null
  if (open) {
    state.close = open[2] + open[1]

    return 'string'
  }

  const q = stream.peek()
  if (q === '"' || q === "'") {
    stream.next()
    while (!stream.eol()) {
      const c = stream.next()
      if (c === '\\') stream.next()
      else if (c === q) break
    }

    return 'string'
  }

  if (stream.match(/^(<->|->|<-|--)/)) return 'operator'

  if (stream.match(/^[A-Za-z_]\w*(-\w+)*/)) return KEYWORDS.has(stream.current()) ? 'keyword' : null
  if (stream.match(/^[{}[\]:;.,]/)) return 'punctuation'

  stream.next()

  return null
}

export const d2Language = StreamLanguage.define<S>({
  startState: () => ({ close: null }),
  copyState: (s) => ({ ...s }),
  token,
})
