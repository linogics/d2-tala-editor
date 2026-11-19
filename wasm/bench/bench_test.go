//go:build js && wasm

package bench

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/linogics/d2-tala-editor/wasm/editor"
)

// graph builds a random tree plus n/5 cross links between distinct nodes.
func graph(n int) string {
	r := rand.New(rand.NewSource(42))

	var b strings.Builder
	for i := 1; i < n; i++ {
		fmt.Fprintf(&b, "n%d -> n%d\n", r.Intn(i), i)
	}

	for i := 0; i < n/5; i++ {
		a, c := r.Intn(n), r.Intn(n)
		for c == a {
			c = r.Intn(n)
		}

		fmt.Fprintf(&b, "n%d -> n%d\n", a, c)
	}

	return b.String()
}

func timed(t *testing.T, label, src string, seeds []int64) editor.RenderResult {
	start := time.Now()

	res, err := editor.Render(context.Background(), src, editor.RenderOpts{Seeds: seeds})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Errors) > 0 {
		t.Fatalf("%s: %v", label, res.Errors)
	}

	fmt.Printf("%s seeds=%v ms=%d\n", label, seeds, time.Since(start).Milliseconds())

	return res
}

func TestTiming(t *testing.T) {
	timed(t, "warmup", "a", nil)

	for _, n := range []int{10, 30, 50} {
		src := graph(n)
		for _, seeds := range [][]int64{{1}, {1, 2, 3}} {
			timed(t, fmt.Sprintf("nodes=%d", n), src, seeds)
		}
	}

	src := graph(50)

	res := timed(t, "layout50", src, []int64{1})
	for _, s := range res.Shapes {
		src += fmt.Sprintf("%s: {top: %d\nleft: %d}\n", s.ID, s.Y, s.X)
	}

	for _, seeds := range [][]int64{{1}, {1, 2, 3}} {
		timed(t, "pinned50", src, seeds)
	}
}
