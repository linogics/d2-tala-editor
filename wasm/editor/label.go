package editor

import (
	"strings"

	"github.com/d2lang/d2/d2ast"
	"github.com/d2lang/d2/d2graph"
)

// labelKey returns the key that writes the label, or nil when the label is the id.
// The compiler sets MapKey even without a label, so only a key that holds a scalar counts.
func labelKey(s d2graph.Scalar) *d2ast.Key {
	k := s.MapKey
	if k == nil || (k.Primary.Unbox() == nil && k.Value.ScalarBox().Unbox() == nil) {
		return nil
	}

	return k
}

// labelOf finds the label attr ("label" or "<end>-arrowhead.label") of the shape or edge id.
func labelOf(g *d2graph.Graph, id, attr string) *d2graph.Scalar {
	if o := lookup(g, id); o != nil {
		if attr == "label" {
			return &o.Label
		}

		return nil
	}

	for _, e := range g.Edges {
		if e.AbsID() != id {
			continue
		}

		switch attr {
		case "label":
			return &e.Label
		case "source-arrowhead.label":
			if e.SrcArrowhead != nil {
				return &e.SrcArrowhead.Label
			}
		case "target-arrowhead.label":
			if e.DstArrowhead != nil {
				return &e.DstArrowhead.Label
			}
		}
	}

	return nil
}

// quoteLabel rewrites an unquoted label of id as a double-quoted string.
func quoteLabel(g *d2graph.Graph, id, attr string) {
	s := labelOf(g, id, attr)
	if s == nil {
		return
	}

	k := labelKey(*s)
	if k == nil {
		return
	}

	if q := quoted(k.Primary.Unbox()); q != nil {
		k.Primary = d2ast.MakeValueBox(q).ScalarBox()
	} else if q := quoted(k.Value.ScalarBox().Unbox()); q != nil {
		k.Value = d2ast.MakeValueBox(q)
	}
}

// quoted returns s as a double-quoted string, or nil for block, double-quoted and null values.
func quoted(s d2ast.Scalar) *d2ast.DoubleQuotedString {
	switch u := s.(type) {
	case *d2ast.UnquotedString:
		q := &d2ast.DoubleQuotedString{Range: u.Range, Value: u.Value}
		// The raw form was escaped for an unquoted string, the printer escapes again from String.
		for i := range q.Value {
			q.Value[i].StringRaw = nil
		}

		return q
	case *d2ast.SingleQuotedString, *d2ast.Number, *d2ast.Boolean:
		return d2ast.FlatDoubleQuotedString(s.ScalarString())
	}

	return nil
}

func isLabelAttr(attr string) bool {
	return attr == "label" || strings.HasSuffix(attr, "-arrowhead.label")
}
