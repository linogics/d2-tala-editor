package editor

import (
	"errors"
	"strings"
	"testing"
)

func TestSetArrows(t *testing.T) {
	cases := []struct {
		name, text, id string
		src, dst       bool
		want, wantID   string
	}{
		{"line", "a -> b", "(a -> b)[0]", false, false, "a -- b", "(a -- b)[0]"},
		{"back", "a -> b", "(a -> b)[0]", true, false, "a <- b", "(a <- b)[0]"},
		{"both", "a -- b", "(a -- b)[0]", true, true, "a <-> b", "(a <-> b)[0]"},
		{"chain", "a -> b -> c", "(b -> c)[0]", false, false, "a -> b -- c", "(b -- c)[0]"},
		{"styled", "a -> b -> c\n(b -> c)[0].style.stroke: red", "(b -> c)[0]", false, false, "(b -- c)[0].style.stroke: red", "(b -- c)[0]"},
		{"nested", "x: {a -> b}", "x.(a -> b)[0]", false, false, "x: {a -- b}", "x.(a -- b)[0]"},
		// The new op meets an earlier a -- b, so the moved edge and its indexed key become [1].
		{
			"renumber", "a -- b\na -> b -> c\n(a -> b)[0].style.stroke: red", "(a -> b)[0]", false, false,
			"(a -- b)[1].style.stroke: red", "(a -- b)[1]",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, id, err := SetArrows(c.text, c.id, c.src, c.dst)
			if err != nil || !strings.Contains(out, c.want) || id != c.wantID {
				t.Fatalf("%q %q %v", out, id, err)
			}
		})
	}
}

func TestSetArrowsFlippedRef(t *testing.T) {
	out, _, err := SetArrows("a -> b\n(b <- a)[0].style.stroke: red", "(a -> b)[0]", true, false)
	if err != nil || !strings.Contains(out, "a <- b") || !strings.Contains(out, "(b -> a)[0]") {
		t.Fatalf("%q %v", out, err)
	}
}

func TestSetArrowsRefusesGlob(t *testing.T) {
	_, _, err := SetArrows("a -> b\n(a -> b)[*].style.stroke: red", "(a -> b)[0]", false, false)
	if !errors.Is(err, ErrGlobEdge) {
		t.Fatal(err)
	}
}

func TestConnectOp(t *testing.T) {
	out, id, err := Connect("a\nb", "a", "b", "--")
	if err != nil || !strings.Contains(out, "a -- b") || id != "(a -- b)[0]" {
		t.Fatalf("%q %q %v", out, id, err)
	}

	if _, _, err := Connect("a\nb", "a", "b", "=>"); !errors.Is(err, errBadArgs) {
		t.Fatal(err)
	}
}
