//go:build js && wasm

// Command editor exposes the D2 editor engine to JavaScript.
package main

import (
	"syscall/js"

	"github.com/linogics/d2-tala-editor/wasm/editor"
)

func main() {
	api := js.Global().Get("Object").New()
	api.Set("call", js.FuncOf(func(_ js.Value, a []js.Value) any {
		return string(editor.Dispatch(a[0].String(), []byte(a[1].String())))
	}))
	js.Global().Set("d2editor", api)

	if ready := js.Global().Get("onD2EditorReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}

	select {}
}
