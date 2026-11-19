package editor

import (
	"errors"

	"github.com/d2lang/d2/d2ast"
	"github.com/d2lang/d2/d2graph"
)

var ErrNotReorderable = errors.New("parent is not a grid or sequence diagram")

// Reorder moves child id of a grid or sequence diagram before sibling beforeID
// ("" means last). D2 orders by first mention, so plain declarations for all
// children are written at the top of the container block.
// shortcut: only the container's own block is handled, children declared as
// "g.a" elsewhere keep their first mention.
func Reorder(text, id, beforeID string) (string, error) {
	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		o, b := lookup(g, id), lookup(g, beforeID)
		if o == nil || o.Parent == nil || (beforeID != "" && (b == nil || b.Parent != o.Parent)) {
			return nil, ErrNotReorderable
		}

		p := o.Parent
		if !p.IsGridDiagram() && !p.IsSequenceDiagram() {
			return nil, ErrNotReorderable
		}

		var m *d2ast.Map

		for _, r := range p.References {
			if r.Key == r.MapKey.Key && r.KeyPathIndex == len(r.Key.Path)-1 && r.MapKey.Value.Map != nil {
				m = r.MapKey.Value.Map
				break
			}
		}

		if m == nil {
			return nil, ErrNotReorderable
		}

		var order []*d2graph.Object

		for _, c := range p.ChildrenArray {
			if c != o {
				if c == b {
					order = append(order, o)
				}

				order = append(order, c)
			}
		}

		if b == nil {
			order = append(order, o)
		}

		same := true
		for i, c := range p.ChildrenArray {
			same = same && order[i] == c
		}

		if same {
			return g, nil
		}

		names := map[string]bool{}

		var decls []d2ast.MapNodeBox

		for _, c := range order {
			names[c.IDVal] = true
			decls = append(decls, d2ast.MakeMapNodeBox(&d2ast.Key{Key: d2ast.MakeKeyPath([]string{c.IDVal})}))
		}

		rest := decls

		for _, n := range m.Nodes {
			if k := n.MapKey; k != nil && isBareDecl(k) && names[k.Key.Path[0].Unbox().ScalarString()] {
				continue
			}

			rest = append(rest, n)
		}

		m.Nodes = rest

		return g, nil
	})
}

func isBareDecl(k *d2ast.Key) bool {
	return k.Key != nil && len(k.Key.Path) == 1 && len(k.Edges) == 0 && k.EdgeIndex == nil && k.EdgeKey == nil &&
		!k.Ampersand && !k.NotAmpersand && k.Primary.Unbox() == nil && k.Value.Unbox() == nil
}

// ReorderEdge moves the sequence message edgeID before beforeEdgeID ("" means last).
// shortcut: both keys must sit in the same map, chained edges move as one line.
func ReorderEdge(text, edgeID, beforeEdgeID string) (string, error) {
	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		find := func(id string) *d2ast.Key {
			for _, e := range g.Edges {
				if e.AbsID() == id && len(e.References) > 0 {
					return e.References[0].MapKey
				}
			}

			return nil
		}

		var e *d2graph.Edge

		for _, x := range g.Edges {
			if x.AbsID() == edgeID {
				e = x
			}
		}

		k, b := find(edgeID), find(beforeEdgeID)
		if e == nil || k == nil || e.Src.OuterSequenceDiagram() == nil || e.Dst.OuterSequenceDiagram() == nil || (beforeEdgeID != "" && b == nil) {
			return nil, ErrNotReorderable
		}

		m := mapHolding(g.AST, k)
		if m == nil || (b != nil && mapHolding(g.AST, b) != m) {
			return nil, ErrNotReorderable
		}

		var (
			nodes []d2ast.MapNodeBox
			moved d2ast.MapNodeBox
		)

		for _, n := range m.Nodes {
			if n.MapKey == k {
				moved = n
			} else {
				nodes = append(nodes, n)
			}
		}

		at := len(nodes)
		for i, n := range nodes {
			if b != nil && n.MapKey == b {
				at = i
			}
		}

		m.Nodes = append(nodes[:at], append([]d2ast.MapNodeBox{moved}, nodes[at:]...)...)

		return g, nil
	})
}

// mapHolding returns the map in the tree whose direct nodes contain k.
func mapHolding(m *d2ast.Map, k *d2ast.Key) *d2ast.Map {
	for _, n := range m.Nodes {
		if n.MapKey == k {
			return m
		}

		if n.MapKey != nil && n.MapKey.Value.Map != nil {
			if r := mapHolding(n.MapKey.Value.Map, k); r != nil {
				return r
			}
		}
	}

	return nil
}
