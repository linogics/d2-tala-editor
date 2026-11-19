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

func graph(n int) string {
	r := rand.New(rand.NewSource(42))
	var b strings.Builder
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "n%d\n", i)
	}
	for i := 0; i < n*13/10; i++ {
		fmt.Fprintf(&b, "n%d -> n%d\n", r.Intn(n), r.Intn(n))
	}
	return b.String()
}

func TestTiming(t *testing.T) {
	for _, n := range []int{10, 50, 150} {
		src := graph(n)
		for _, seeds := range [][]int64{{1}, {1, 2, 3}} {
			start := time.Now()
			if _, err := editor.Render(context.Background(), src, editor.RenderOpts{Seeds: seeds}); err != nil {
				t.Fatal(err)
			}
			fmt.Printf("nodes=%d seeds=%v ms=%d\n", n, seeds, time.Since(start).Milliseconds())
		}
	}
}
