package editor

import (
	"errors"
	"fmt"
	"strings"

	"github.com/d2lang/d2/d2compiler"
	"github.com/d2lang/d2/d2format"
	"github.com/d2lang/d2/d2graph"
	"github.com/d2lang/d2/d2oracle"
	"github.com/d2lang/d2/d2parser"
)

var (
	ErrBrokenText  = errors.New("text does not compile")
	ErrNotPinnable = errors.New("shape cannot be pinned")
)

// brokenError wraps ErrBrokenText and carries the first diagnostic.
type brokenError struct{ diag Diag }

func (e *brokenError) Error() string {
	return fmt.Sprintf("%v: %d:%d: %s", ErrBrokenText, e.diag.Line, e.diag.Col, e.diag.Message)
}
func (e *brokenError) Unwrap() error { return ErrBrokenText }

// BrokenDiag recovers the first diagnostic from an ErrBrokenText error.
func BrokenDiag(err error) (Diag, bool) {
	var be *brokenError
	if errors.As(err, &be) {
		return be.diag, true
	}
	return Diag{}, false
}

func compile(text string) (*d2graph.Graph, error) {
	g, _, err := d2compiler.Compile("", strings.NewReader(text), &d2compiler.CompileOptions{UTF16Pos: true})
	if err != nil {
		d := Diag{Message: err.Error()}
		var pe *d2parser.ParseError
		if errors.As(err, &pe) && len(pe.Errors) > 0 {
			e := pe.Errors[0]
			d = Diag{e.Range.Start.Line, e.Range.Start.Column, e.Message}
		}
		return nil, &brokenError{d}
	}
	return g, nil
}

// edit compiles text, applies fn to the graph and formats the result.
func edit(text string, fn func(g *d2graph.Graph) (*d2graph.Graph, error)) (string, error) {
	g, err := compile(text)
	if err != nil {
		return "", err
	}
	if g, err = fn(g); err != nil {
		return "", err
	}
	return d2format.Format(g.AST), nil
}

func join(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}

func Create(text, parent, shape string) (string, string, error) {
	base := shape
	if base == "" {
		base = "rectangle"
	}
	var id string
	out, err := edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		var err error
		g, id, err = d2oracle.Create(g, nil, join(parent, base))
		if err != nil || shape == "" {
			return g, err
		}
		return d2oracle.Set(g, nil, id+".shape", nil, &shape)
	})
	return out, id, err
}

// Set sets attr (relative to id) to value, or deletes it when value is nil.
func Set(text, id, attr string, value *string) (string, error) {
	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		if value == nil {
			return d2oracle.Delete(g, nil, join(id, attr))
		}
		return d2oracle.Set(g, nil, join(id, attr), nil, value)
	})
}

func Delete(text, id string) (string, error) {
	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		return d2oracle.Delete(g, nil, id)
	})
}

func Move(text, id, newParent string) (string, string, error) {
	nk, err := d2parser.ParseKey(id)
	if err != nil || len(nk.Path) == 0 {
		return "", "", fmt.Errorf("bad id %q", id)
	}
	newID := join(newParent, nk.Path[len(nk.Path)-1].Unbox().ScalarString())
	out, err := edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		return d2oracle.Move(g, nil, id, newID, true)
	})
	return out, newID, err
}

func Rename(text, id, name string) (string, string, error) {
	var nid string
	out, err := edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		var err error
		g, nid, err = d2oracle.Rename(g, nil, id, name)
		return g, err
	})
	return out, nid, err
}

func Connect(text, src, dst string) (string, string, error) {
	var eid string
	out, err := edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		var err error
		g, eid, err = d2oracle.Create(g, nil, src+" -> "+dst)
		return g, err
	})
	return out, eid, err
}

func Reconnect(text, edgeID string, src, dst *string) (string, error) {
	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		return d2oracle.ReconnectEdge(g, nil, edgeID, src, dst)
	})
}

func Format(text string) (string, error) {
	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) { return g, nil })
}
