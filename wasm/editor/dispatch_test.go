package editor

import (
	"encoding/json"
	"strings"
	"testing"
)

type env struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
	Error  struct {
		Code string `json:"code"`
		Diag *Diag  `json:"diag"`
	} `json:"error"`
}

func call(t *testing.T, op, args string) env {
	t.Helper()

	var e env
	if err := json.Unmarshal(Dispatch(op, []byte(args)), &e); err != nil {
		t.Fatal(err)
	}

	return e
}

func TestDispatchRender(t *testing.T) {
	e := call(t, "render", `{"text":"a -> b","seeds":[1]}`)
	if !e.OK {
		t.Fatalf("not ok: %+v", e)
	}

	var r RenderResult
	if err := json.Unmarshal(e.Result, &r); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(r.SVG, "<svg") || len(r.Shapes) != 2 {
		t.Fatalf("bad result: %d shapes", len(r.Shapes))
	}
}

func TestDispatchUnknownOp(t *testing.T) {
	if e := call(t, "nope", `{}`); e.OK || e.Error.Code != "unknown_op" {
		t.Fatalf("got %+v", e)
	}
}

func TestDispatchBadArgs(t *testing.T) {
	if e := call(t, "format", `{`); e.OK || e.Error.Code != "bad_args" {
		t.Fatalf("got %+v", e)
	}
}

func TestDispatchBrokenText(t *testing.T) {
	e := call(t, "format", "{\"text\":\"a\\n b: {\"}")
	if e.OK || e.Error.Code != "broken_text" || e.Error.Diag == nil {
		t.Fatalf("got %+v", e)
	}
}

func TestDispatchNotPinnable(t *testing.T) {
	e := call(t, "pin", `{"text":"g: {grid-rows: 1\nc}","id":"g.c","x":1,"y":1}`)
	if e.OK || e.Error.Code != "not_pinnable" {
		t.Fatalf("got %+v", e)
	}
}

func TestDispatchEmptyListsAreArrays(t *testing.T) {
	if e := call(t, "refRanges", `{"text":"a","id":"zzz"}`); !e.OK || string(e.Result) != "[]" {
		t.Fatalf("got %s", e.Result)
	}
}

func TestDispatchEveryOp(t *testing.T) {
	const two = `a\nb`

	cases := map[string]string{
		"render":      `{"text":"a -> b","seeds":[1]}`,
		"create":      `{"text":"","parent":"","shape":"rectangle"}`,
		"insert":      `{"text":"","parent":"","template":"$: {shape: circle}"}`,
		"set":         `{"text":"a","id":"a","attr":"style.fill","value":"red"}`,
		"delete":      `{"text":"` + two + `","id":"a"}`,
		"move":        `{"text":"` + two + `","id":"a","parent":"b"}`,
		"rename":      `{"text":"a","id":"a","name":"z"}`,
		"connect":     `{"text":"` + two + `","src":"a","dst":"b"}`,
		"reconnect":   `{"text":"a -> b\nc","id":"(a -> b)[0]","dst":"c"}`,
		"pin":         `{"text":"a","id":"a","x":10,"y":20}`,
		"unpin":       `{"text":"a: {top: 1\nleft: 1}","id":"a"}`,
		"pinAll":      `{"text":"a: {top: 1\nleft: 1}"}`,
		"reorder":     `{"text":"g: {grid-rows: 1\na\nb}","id":"g.b","before":"g.a"}`,
		"reorderEdge": `{"text":"shape: sequence_diagram\na -> b: 1\nb -> a: 2","id":"(a -> b)[0]","before":"(b -> a)[0]"}`,
		"format":      `{"text":"a:   {b}"}`,
		"completions": `{"text":"a: {","line":0,"col":4}`,
		"refRanges":   `{"text":"a -> b","id":"a"}`,
		"themes":      `{}`,
	}

	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			if e := call(t, name, args); !e.OK {
				t.Fatalf("%s: %+v", name, e)
			}
		})
	}
}
