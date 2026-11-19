package editor

import (
	"errors"
	"testing"
)

var fast = RenderOpts{Seeds: []int64{1}}

func xOrder(t *testing.T, text string, ids ...string) {
	t.Helper()
	r := mustRender(t, text, fast)
	for i := 1; i < len(ids); i++ {
		if byID(r, ids[i-1]).X >= byID(r, ids[i]).X {
			t.Fatalf("%s not left of %s:\n%s", ids[i-1], ids[i], text)
		}
	}
}

func TestReorderGridCells(t *testing.T) {
	out := must(t)(Reorder("g: {grid-columns: 3\na\nb\nc}", "g.c", "g.a"))
	xOrder(t, out, "g.c", "g.a", "g.b")
}

func TestReorderGridLast(t *testing.T) {
	out := must(t)(Reorder("g: {grid-columns: 3\na\nb\nc}", "g.a", ""))
	xOrder(t, out, "g.b", "g.c", "g.a")
}

func TestReorderGridFirstMentionInEdge(t *testing.T) {
	out := must(t)(Reorder("g: {grid-columns: 2\na -> b\nc}", "g.b", "g.a"))
	xOrder(t, out, "g.b", "g.a")
	g, err := compile(out)
	if err != nil || len(g.Edges) != 1 {
		t.Fatalf("edge lost:\n%s", out)
	}
}

func TestReorderSequenceActors(t *testing.T) {
	out := must(t)(Reorder("s: {shape: sequence_diagram\nalice -> bob\nbob -> carol}", "s.carol", "s.alice"))
	xOrder(t, out, "s.carol", "s.alice", "s.bob")
}

func TestReorderEdgeMessages(t *testing.T) {
	const in = "s: {shape: sequence_diagram\nalice -> bob\nbob -> carol}"
	g, err := compile(in)
	if err != nil {
		t.Fatal(err)
	}
	e0, e1 := g.Edges[0].AbsID(), g.Edges[1].AbsID()
	out := must(t)(ReorderEdge(in, e1, e0))
	r := mustRender(t, out, fast)
	var y0, y1 float64
	for _, e := range r.Edges {
		switch e.ID {
		case e0:
			y0 = e.Points[0][1]
		case e1:
			y1 = e.Points[0][1]
		}
	}
	if y1 >= y0 {
		t.Fatalf("not swapped (%v vs %v):\n%s", y1, y0, out)
	}
}

func TestReorderNotReorderable(t *testing.T) {
	if _, err := Reorder("c: {a\nb}", "c.b", "c.a"); !errors.Is(err, ErrNotReorderable) {
		t.Fatal(err)
	}
}

func TestReorderIdempotentText(t *testing.T) {
	const in = "g: {grid-columns: 3\na\nb\nc}"
	want := must(t)(Format(in))
	if got := must(t)(Reorder(in, "g.b", "g.c")); got != want {
		t.Fatalf("%q vs %q", got, want)
	}
}

func TestReorderQuotedIDs(t *testing.T) {
	r := mustRender(t, "g: {grid-columns: 2\n\"a.b\"\n\"my node\"}", fast)
	var ids []string
	for _, s := range r.Shapes {
		if s.Parent == "g" {
			ids = append(ids, s.ID)
		}
	}
	if len(ids) != 2 {
		t.Fatalf("%v", ids)
	}
	out := must(t)(Reorder("g: {grid-columns: 2\n\"a.b\"\n\"my node\"}", ids[1], ids[0]))
	xOrder(t, out, ids[1], ids[0])
}

func TestReorderEdgeFromSpan(t *testing.T) {
	const in = "s: {shape: sequence_diagram\nalice.t -> bob: one\nalice -> bob: two}"
	g, err := compile(in)
	if err != nil {
		t.Fatal(err)
	}
	var first, second string
	for _, e := range g.Edges {
		switch e.Label.Value {
		case "one":
			first = e.AbsID()
		case "two":
			second = e.AbsID()
		}
	}
	out := must(t)(ReorderEdge(in, second, first))
	r := mustRender(t, out, fast)
	var y1, y2 float64
	for _, e := range r.Edges {
		switch e.ID {
		case first:
			y1 = e.Points[0][1]
		case second:
			y2 = e.Points[0][1]
		}
	}
	if y2 >= y1 {
		t.Fatalf("not swapped (%v vs %v):\n%s", y2, y1, out)
	}
}
