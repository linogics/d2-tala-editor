package editor

import (
	"regexp"
	"strings"

	"github.com/d2lang/d2/d2graph"
	"github.com/d2lang/d2/d2oracle"
)

var shapeLine = regexp.MustCompile(`(?m)^\s*shape:\s*([\w-]+)`)

// Insert adds template under parent. Every "$" in template becomes the new
// object's id, which is unique like in Create.
func Insert(text, parent, template string) (string, string, error) {
	base := "text"
	if m := shapeLine.FindStringSubmatch(template); m != nil {
		base = m[1]
	}

	var id string

	// The graph only picks the unique id, the template text is appended below.
	if _, err := edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		var err error

		g, id, err = d2oracle.Create(g, nil, join(parent, base))

		return g, err
	}); err != nil {
		return "", "", err
	}

	out, err := Format(text + "\n" + strings.ReplaceAll(template, "$", id))

	return out, id, err
}
