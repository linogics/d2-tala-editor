package editor

import (
	"context"
	"strings"
	"testing"
)

func byID(r RenderResult, id string) ShapeInfo {
	for _, s := range r.Shapes {
		if s.ID == id {
			return s
		}
	}

	return ShapeInfo{}
}

func mustRender(t *testing.T, text string, o RenderOpts) RenderResult {
	t.Helper()

	r, err := Render(context.Background(), text, o)
	if err != nil {
		t.Fatal(err)
	}

	return r
}

func TestRenderSimple(t *testing.T) {
	r := mustRender(t, "a -> b", RenderOpts{})
	if len(r.Shapes) != 2 || len(r.Edges) != 1 || len(r.Errors) != 0 {
		t.Fatalf("got %d shapes %d edges %v", len(r.Shapes), len(r.Edges), r.Errors)
	}

	if r.Edges[0].ID != "(a -> b)[0]" {
		t.Fatalf("edge id %q", r.Edges[0].ID)
	}

	if !strings.Contains(r.SVG, "<svg") {
		t.Fatal("no svg")
	}
}

func TestRenderUsesTala(t *testing.T) {
	r := mustRender(t, "a: {top: 100\nleft: 200}\nb\na -> b", RenderOpts{})
	if a := byID(r, "a"); a.X != 200 || a.Y != 100 || !a.Pinned {
		t.Fatalf("a = %+v errors %v", a, r.Errors)
	}
}

func TestRenderCompileError(t *testing.T) {
	r := mustRender(t, "a -> \nb: {", RenderOpts{})
	// D2 reports the dangling edge on line 0 and the unterminated map on line 1.
	if len(r.Errors) < 2 || r.Errors[1].Line != 1 || r.SVG != "" {
		t.Fatalf("got %+v svg=%q", r.Errors, r.SVG)
	}
}

func TestRenderUTF16Col(t *testing.T) {
	r := mustRender(t, "é: {shape: nope}", RenderOpts{})
	if len(r.Errors) < 1 {
		t.Fatal("no error")
	}
	// "é: {shape: " is 11 UTF-16 units, so the value starts at col 11 (12 if é counted as 2).
	if r.Errors[0].Col != 11 {
		t.Fatalf("col %d, errors %+v", r.Errors[0].Col, r.Errors)
	}
}

func TestRenderContainerKinds(t *testing.T) {
	r := mustRender(t, "g: {grid-columns: 2\nx\ny}\ns: {shape: sequence_diagram\nx -> y}", RenderOpts{})
	if len(r.Errors) != 0 {
		t.Fatal(r.Errors)
	}

	for id, want := range map[string]ShapeInfo{
		"g": {Container: "grid", Pinnable: true}, "s": {Container: "sequence", Pinnable: true},
		"g.x": {}, "s.x": {},
	} {
		got := byID(r, id)
		if got.ID == "" || got.Container != want.Container || got.Pinnable != want.Pinnable {
			t.Errorf("%s = %+v, want container=%q pinnable=%v", id, got, want.Container, want.Pinnable)
		}
	}
}

func TestRenderFastSeeds(t *testing.T) {
	r := mustRender(t, "a -> b -> c", RenderOpts{Seeds: []int64{1}})
	if len(r.Errors) != 0 || len(r.Shapes) != 3 {
		t.Fatalf("%+v", r)
	}
}

func TestRenderSkipsLifelines(t *testing.T) {
	r, err := Render(t.Context(), "s: {shape: sequence_diagram\na -> b}", RenderOpts{Seeds: []int64{1}})
	if err != nil {
		t.Fatal(err)
	}

	if len(r.Edges) != 1 || r.Edges[0].ID != "s.(a -> b)[0]" {
		t.Fatalf("edges = %+v", r.Edges)
	}
}

func TestRenderAttrs(t *testing.T) {
	r := mustRender(t, "a: {style.fill: red; top: 10; left: 20; near: top-center}\nb\na -> b: {style.animated: true; target-arrowhead.shape: diamond}", RenderOpts{})

	a := byID(r, "a").Attrs
	if a["style.fill"] != "red" || a["top"] != "10" || a["left"] != "20" || a["label"] != "" {
		t.Fatalf("a attrs %v errors %v", a, r.Errors)
	}

	if b := byID(r, "b").Attrs; len(b) != 0 {
		t.Fatalf("b attrs %v, want none", b)
	}

	e := r.Edges[0].Attrs
	if e["style.animated"] != "true" || e["target-arrowhead.shape"] != "diamond" {
		t.Fatalf("edge attrs %v", e)
	}
}
