package editor

import (
	"errors"
	"strings"

	"github.com/d2lang/d2/d2graph"
)

var ErrGlobEdge = errors.New("edge is set through a glob, edit its op in the code")

// arrowOps maps the D2 edge ops to their src and dst heads.
var arrowOps = map[string][2]bool{"->": {false, true}, "<-": {true, false}, "<->": {true, true}, "--": {false, false}}

func astArrows(src, dst bool) (string, string) {
	var s, d string
	if src {
		s = "<"
	}

	if dst {
		d = ">"
	}

	return s, d
}

// edgeKey groups edges that share an index, a flipped edge counts as the same one.
func edgeKey(src, dst string, sa, da bool) [4]any {
	if src > dst {
		src, dst, sa, da = dst, src, da, sa
	}

	return [4]any{src, dst, sa, da}
}

// SetArrows sets whether edgeID has a head at its src and dst end, which picks
// its op among "->", "<-", "<->" and "--". It returns the text and the new edge id.
func SetArrows(text, edgeID string, src, dst bool) (string, string, error) {
	pos := -1

	out, err := edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		var e *d2graph.Edge

		for i, x := range g.Edges {
			if x.AbsID() == edgeID {
				e, pos = x, i
			}
		}

		if e == nil {
			return nil, errors.New("edge not found")
		}

		for _, ref := range e.References {
			if ix := ref.MapKey.EdgeIndex; ix != nil && ix.Glob {
				return nil, ErrGlobEdge
			}
		}

		for _, ref := range e.References {
			s, d := src, dst
			if refFlipped(ref, e) {
				s, d = d, s
			}

			ref.Edge.SrcArrow, ref.Edge.DstArrow = astArrows(s, d)
		}

		// Indexes count the edges that share ends and op, so changing the op renumbers both groups.
		heads := func(x *d2graph.Edge) (bool, bool) {
			if x == e {
				return src, dst
			}

			return x.SrcArrow, x.DstArrow
		}
		next := map[[4]any]int{}

		for _, x := range g.Edges {
			sa, da := heads(x)
			k := edgeKey(x.Src.AbsID(), x.Dst.AbsID(), sa, da)
			n := next[k]
			next[k]++

			for _, ref := range x.References {
				if ix := ref.MapKey.EdgeIndex; ix != nil && ix.Int != nil {
					*ix.Int = n
				}
			}
		}

		return g, nil
	})
	if err != nil {
		return "", "", err
	}

	g, err := compile(out)
	if err != nil {
		return "", "", err
	}

	return out, g.Edges[pos].AbsID(), nil
}

// refFlipped reports whether ref writes e with its ends swapped, as in "b <- a" for a -> b.
func refFlipped(ref d2graph.EdgeReference, e *d2graph.Edge) bool {
	obj, ida, err := d2graph.ResolveUnderscoreKey(d2graph.Key(ref.Edge.Src), ref.ScopeObj)
	if err != nil {
		return false
	}

	src := strings.Join(ida, ".")
	if obj.Parent != nil {
		src = obj.AbsID() + "." + src
	}

	return src != e.Src.AbsID() && strings.EqualFold(src, e.Dst.AbsID())
}
