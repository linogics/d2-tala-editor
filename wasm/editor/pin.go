package editor

import (
	"context"
	"strconv"

	"github.com/d2lang/d2/d2format"

	"github.com/d2lang/d2/d2graph"
	"github.com/d2lang/d2/d2oracle"
)

func lookup(g *d2graph.Graph, id string) *d2graph.Object {
	for _, o := range g.Objects {
		if o.AbsID() == id {
			return o
		}
	}

	return nil
}

func setPos(g *d2graph.Graph, id string, x, y int) (*d2graph.Graph, error) {
	o := lookup(g, id)
	if o == nil || !pinnable(o) {
		return nil, ErrNotPinnable
	}
	// The compiler rejects negative top and left.
	top, left := strconv.Itoa(max(y, 0)), strconv.Itoa(max(x, 0))

	g, err := d2oracle.Set(g, nil, id+".top", nil, &top)
	if err != nil {
		return nil, err
	}

	return d2oracle.Set(g, nil, id+".left", nil, &left)
}

// Pin sets top and left so the shape's corner lands at x, y. TALA reads nested
// top and left relative to the parent's content origin, so nested shapes are
// converted using one probe render with the shape pinned at 0, 0.
func Pin(text, id string, x, y int) (string, error) {
	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		o := lookup(g, id)
		if o == nil || !pinnable(o) {
			return nil, ErrNotPinnable
		}

		if o.Parent != nil {
			g, err := setPos(g, id, 0, 0)
			if err != nil {
				return nil, err
			}

			r, err := Render(context.Background(), d2format.Format(g.AST), RenderOpts{Seeds: []int64{1}})
			if err != nil {
				return nil, err
			}

			c := shapeByID(r, id) // at rel 0,0 its position is the parent's content origin

			return setPos(g, id, x-c.X, y-c.Y)
		}

		return setPos(g, id, x, y)
	})
}

func shapeByID(r RenderResult, id string) ShapeInfo {
	for _, s := range r.Shapes {
		if s.ID == id {
			return s
		}
	}

	return ShapeInfo{}
}

func Unpin(text, id string) (string, error) {
	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		g, err := d2oracle.Delete(g, nil, id+".top")
		if err != nil {
			return nil, err
		}

		return d2oracle.Delete(g, nil, id+".left")
	})
}

// PinAll renders once and pins every pinnable shape where it was laid out.
// Nested shapes are pinned at 0, 0 first, then converted to offsets from their
// parent's position using a probe render.
func PinAll(ctx context.Context, text string) (string, error) {
	if _, err := compile(text); err != nil {
		return "", err
	}

	r, err := Render(ctx, text, RenderOpts{})
	if err != nil {
		return "", err
	}

	return edit(text, func(g *d2graph.Graph) (*d2graph.Graph, error) {
		var nested []ShapeInfo

		for _, s := range r.Shapes {
			if !s.Pinnable {
				continue
			}

			x, y := s.X, s.Y
			if s.Parent != "" {
				nested = append(nested, s)
				x, y = 0, 0
			}

			var err error
			if g, err = setPos(g, s.ID, x, y); err != nil {
				return nil, err
			}
		}

		if len(nested) == 0 {
			return g, nil
		}

		probe, err := Render(ctx, d2format.Format(g.AST), RenderOpts{Seeds: []int64{1}})
		if err != nil {
			return nil, err
		}

		for _, s := range nested {
			c, p, pa := shapeByID(probe, s.ID), shapeByID(probe, s.Parent), shapeByID(r, s.Parent)
			if g, err = setPos(g, s.ID, s.X-pa.X-(c.X-p.X), s.Y-pa.Y-(c.Y-p.Y)); err != nil {
				return nil, err
			}
		}

		return g, nil
	})
}
