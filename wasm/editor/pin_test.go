package editor

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestPinRoundTrip(t *testing.T) {
	out := must(t)(Pin("a\nb\na -> b", "a", 300, 120))
	a := byID(mustRender(t, out, RenderOpts{Seeds: []int64{1}}), "a")
	if a.X != 300 || a.Y != 120 {
		t.Fatalf("%q at %d,%d", out, a.X, a.Y)
	}
}

func TestPinNested(t *testing.T) {
	out := must(t)(Pin("c: {x}\ny", "c.x", 400, 260))
	t.Log(out)
	x := byID(mustRender(t, out, RenderOpts{Seeds: []int64{1}}), "c.x")
	if x.X != 400 || x.Y != 260 {
		t.Fatalf("%q at %d,%d", out, x.X, x.Y)
	}
}

func TestPinRefused(t *testing.T) {
	for _, in := range []string{
		"g: {grid-columns: 2\na\nb}",
		"s: {shape: sequence_diagram\na -> b}",
	} {
		id := "g.a"
		if strings.HasPrefix(in, "s") {
			id = "s.a"
		}
		if _, err := Pin(in, id, 1, 2); !errors.Is(err, ErrNotPinnable) {
			t.Errorf("%s: %v", id, err)
		}
	}
}

func TestUnpin(t *testing.T) {
	want := must(t)(Format("a\nb"))
	out := must(t)(Pin("a\nb", "a", 10, 20))
	if got := must(t)(Unpin(out, "a")); got != want {
		t.Fatalf("%q vs %q", got, want)
	}
}

func TestPinAllSkipsGridChildren(t *testing.T) {
	out, err := PinAll(context.Background(), "g: {grid-columns: 2\na\nb}\ny")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out, "left:") != 2 || strings.Contains(out, "a: {\n    top") {
		t.Fatalf("%s", out)
	}
	g, _ := compile(out)
	if o := lookup(g, "g.a"); o == nil || o.Top != nil {
		t.Fatalf("grid child pinned:\n%s", out)
	}
}
