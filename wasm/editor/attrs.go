package editor

import (
	"strings"

	"github.com/d2lang/d2/d2format"
	"github.com/d2lang/d2/d2graph"
)

// attrs returns the D2 attributes set on a shape or edge, keyed by their D2 path.
func attrs(a *d2graph.Attributes) map[string]string {
	m := map[string]string{"label": a.Label.Value}
	s := a.Style

	for k, v := range map[string]*d2graph.Scalar{
		"top": a.Top, "left": a.Left, "width": a.WidthAttr, "height": a.HeightAttr,
		"tooltip": a.Tooltip, "link": a.Link,
		"style.fill": s.Fill, "style.stroke": s.Stroke, "style.stroke-width": s.StrokeWidth,
		"style.stroke-dash": s.StrokeDash, "style.border-radius": s.BorderRadius, "style.opacity": s.Opacity,
		"style.font-size": s.FontSize, "style.font-color": s.FontColor, "style.shadow": s.Shadow,
		"style.3d": s.ThreeDee, "style.multiple": s.Multiple, "style.bold": s.Bold, "style.italic": s.Italic,
		"style.underline": s.Underline, "style.double-border": s.DoubleBorder, "style.animated": s.Animated,
	} {
		// Layout fills defaults such as font-size without a MapKey, only keep what the source sets.
		if v != nil && v.MapKey != nil {
			m[k] = v.Value
		}
	}

	if len(a.Classes) > 0 {
		m["class"] = strings.Join(a.Classes, ";")
	}

	if a.Icon != nil {
		m["icon"] = a.Icon.String()
	}

	if a.NearKey != nil {
		m["near"] = d2format.Format(a.NearKey)
	}

	return m
}

func edgeAttrs(e *d2graph.Edge) map[string]string {
	m := attrs(&e.Attributes)

	for k, h := range map[string]*d2graph.Attributes{"source-arrowhead": e.SrcArrowhead, "target-arrowhead": e.DstArrowhead} {
		if h == nil {
			continue
		}

		if h.Shape.Value != "" {
			m[k+".shape"] = h.Shape.Value
		}

		if h.Label.Value != "" {
			m[k+".label"] = h.Label.Value
		}
	}

	return m
}
