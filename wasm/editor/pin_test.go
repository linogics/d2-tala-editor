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

func TestPinClampsNegative(t *testing.T) {
	out := must(t)(Pin("a", "a", -50, -10))
	_, err := compile(out)
	if err != nil || !strings.Contains(out, "top: 0") || !strings.Contains(out, "left: 0") {
		t.Fatalf("%q %v", out, err)
	}
	out = must(t)(Pin("c: {x}\ny", "c.x", -500, -500))
	if _, err := compile(out); err != nil {
		t.Fatalf("%q %v", out, err)
	}
}

func TestPinAllNestedKeepsPositions(t *testing.T) {
	const in = "c: {x -> y}\nz -> c"
	before := mustRender(t, in, RenderOpts{})
	out, err := PinAll(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := compile(out); err != nil {
		t.Fatalf("%q %v", out, err)
	}
	after := mustRender(t, out, RenderOpts{})
	for _, id := range []string{"c.x", "c.y"} {
		b, a := byID(before, id), byID(after, id)
		if abs(a.X-b.X) > 2 || abs(a.Y-b.Y) > 2 {
			t.Errorf("%s moved %d,%d to %d,%d\n%s", id, b.X, b.Y, a.X, a.Y, out)
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
