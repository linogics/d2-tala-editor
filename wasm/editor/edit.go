package editor

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/d2lang/d2/d2ast"
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
		return "", refusal(err)
	}

	return d2format.Format(g.AST), nil
}

// refusal shortens an oracle "failed to recompile" error, which quotes the whole
// rewritten text, to the compiler message without its position in that text.
func refusal(err error) error {
	msg := err.Error()
	if !strings.Contains(msg, "failed to recompile") {
		return err
	}

	last := msg[strings.LastIndex(msg, "\n")+1:]

	return errors.New(recompilePos.ReplaceAllString(last, ""))
}

var recompilePos = regexp.MustCompile(`^\d+:\d+: `)

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
// A class value in array form ("[a; b]") is written as a D2 array.
func Set(text, id, attr string, value *string) (string, error) {
	if attr == "class" {
		// The oracle neither deletes a shape's class nor replaces an array one, so old keys go first.
		t, err := clearClass(text, id)
		if err != nil || value == nil {
			return t, err
		}

		text = t
		if strings.HasPrefix(*value, "[") {
			return setClassList(text, id, *value)
		}
	}

	if attr == "shape" && value != nil {
		t, err := dropStyles(text, id, *value)
		if err != nil {
			return "", err
		}

		text = t
	}

	if value == nil && slices.Contains(gridKeys, attr) {
		return dropKey(text, id, attr)
	}

	out, err := edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		if value == nil && attr == "label" {
			// Delete misses the inline form ("a: A"), Set with nil clears it.
			g, err := d2oracle.Set(g, nil, id, nil, nil)
			if err != nil {
				return nil, err
			}

			return d2oracle.Delete(g, nil, join(id, attr))
		}

		if value == nil {
			return d2oracle.Delete(g, nil, join(id, attr))
		}

		return d2oracle.Set(g, nil, join(id, attr), nil, value)
	})
	if err != nil || value == nil {
		return out, err
	}

	return dedupe(out, id, attr, *value)
}

// gridKeys are the keys d2oracle.Delete leaves in place.
var gridKeys = []string{"grid-rows", "grid-columns", "grid-gap", "vertical-gap", "horizontal-gap"}

// dropKey removes every id.attr key straight from the AST.
func dropKey(text, id, attr string) (string, error) {
	mk, err := d2parser.ParseMapKey(id)
	if err != nil {
		return "", err
	}

	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		for _, l := range findLeaves(g.AST, append(keyPath(mk), attr)) {
			l.m.Nodes = slices.DeleteFunc(l.m.Nodes, func(n d2ast.MapNodeBox) bool { return n.MapKey == l.k })
		}

		return g, nil
	})
}

// dedupe keeps only the key at id.attr that holds value. The oracle adds a second
// "icon" or "near" key instead of replacing the first, sometimes into an earlier map.
func dedupe(text, id, attr, value string) (string, error) {
	mk, err := d2parser.ParseMapKey(id)
	if err != nil || len(mk.Edges) > 0 {
		return text, nil
	}

	m, err := d2parser.Parse("", strings.NewReader(text), nil)
	if err != nil {
		return "", err
	}

	leaves := findLeaves(m, append(keyPath(mk), strings.Split(attr, ".")...))
	if len(leaves) < 2 {
		return text, nil
	}

	keep := len(leaves) - 1
	for i, l := range leaves {
		if l.k.Value.ScalarBox().Unbox() != nil && l.k.Value.ScalarBox().Unbox().ScalarString() == value {
			keep = i
		}
	}

	for i, l := range leaves {
		if i != keep {
			l.m.Nodes = slices.DeleteFunc(l.m.Nodes, func(n d2ast.MapNodeBox) bool { return n.MapKey == l.k })
		}
	}

	return d2format.Format(m), nil
}

// styleShapes lists the only shapes D2 accepts each style on, see d2compiler validateKey.
var styleShapes = map[string][]string{
	"3d":            {"rectangle", "square", "hexagon"},
	"double-border": {"rectangle", "square", "circle", "oval"},
}

// dropStyles removes the styles and column constraints of shape id that D2 refuses on shape.
func dropStyles(text, id, shape string) (string, error) {
	mk, err := d2parser.ParseMapKey(id)
	if err != nil {
		return "", err
	}

	if len(mk.Edges) > 0 {
		return text, nil
	}

	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		obj, ok := g.Root.HasChild(d2graph.Key(mk.Key))
		if !ok {
			return g, nil
		}

		set := map[string]bool{"3d": obj.Style.ThreeDee != nil, "double-border": obj.Style.DoubleBorder != nil}
		for style, shapes := range styleShapes {
			if !set[style] || slices.Contains(shapes, strings.ToLower(shape)) {
				continue
			}

			if g, err = d2oracle.Delete(g, nil, join(id, "style."+style)); err != nil {
				return nil, err
			}
		}

		if obj.SQLTable == nil || strings.EqualFold(shape, "sql_table") {
			return g, nil
		}
		// Columns keep their name and type, only sql_table accepts a constraint.
		for _, c := range obj.SQLTable.Columns {
			if len(c.Constraint) == 0 {
				continue
			}

			// The oracle does not see columns as objects, so the key goes from the AST.
			for _, l := range findLeaves(g.AST, append(keyPath(mk), c.Name.Label, "constraint")) {
				l.m.Nodes = slices.DeleteFunc(l.m.Nodes, func(n d2ast.MapNodeBox) bool { return n.MapKey == l.k })
			}
		}

		return g, nil
	})
}

// clearClass removes every class key of the shape or edge id.
func clearClass(text, id string) (string, error) {
	mk, err := d2parser.ParseMapKey(id)
	if err != nil {
		return "", err
	}

	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		if len(mk.Edges) > 0 {
			return d2oracle.Delete(g, nil, join(id, "class"))
		}

		obj, ok := g.Root.HasChild(d2graph.Key(mk.Key))
		if !ok {
			return g, nil
		}

		all := func(string) bool { return true }

		for _, ref := range obj.References {
			if ref.InEdge() {
				continue
			}

			last := len(ref.Key.Path) - 1
			switch {
			case ref.KeyPathIndex == last && ref.MapKey.Value.Map != nil:
				// "a: {class: x}", only the direct class keys belong to a.
				m := ref.MapKey.Value.Map
				m.Nodes = slices.DeleteFunc(m.Nodes, func(n d2ast.MapNodeBox) bool {
					return len(keyPath(n.MapKey)) == 1 && isClassKey(n.MapKey) && dropClass(n.MapKey, all)
				})
			case ref.KeyPathIndex == last-1 && isClassKey(ref.MapKey):
				// "a.class: x", cut back to "a".
				dropClass(ref.MapKey, all)
			}
		}

		return g, nil
	})
}

// Delete removes id. Without keepChildren its descendants go too, deepest first,
// since the oracle lifts the children of a deleted container into its parent.
func Delete(text, id string, keepChildren bool) (string, error) {
	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		if !keepChildren {
			var kids []string
			for _, o := range g.Objects {
				if strings.HasPrefix(o.AbsID(), id+".") {
					kids = append(kids, o.AbsID())
				}
			}
			slices.SortFunc(kids, func(a, b string) int { return strings.Count(b, ".") - strings.Count(a, ".") })

			for _, k := range kids {
				var err error
				if g, err = d2oracle.Delete(g, nil, k); err != nil {
					return nil, err
				}
			}
		}

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

// Connect adds the edge "src op dst", op is one of the D2 edge ops.
func Connect(text, src, dst, op string) (string, string, error) {
	if _, ok := arrowOps[op]; !ok {
		return "", "", fmt.Errorf("%w: unknown edge op %q", errBadArgs, op)
	}

	var eid string

	out, err := edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		var err error

		g, eid, err = d2oracle.Create(g, nil, src+" "+op+" "+dst)

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
