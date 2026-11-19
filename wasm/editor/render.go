// Package editor holds the D2 engine behind the browser editor.
package editor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/d2lang/d2/d2graph"
	"github.com/d2lang/d2/d2layouts/d2talalayout"
	"github.com/d2lang/d2/d2lib"
	"github.com/d2lang/d2/d2parser"
	"github.com/d2lang/d2/d2renderers/d2svg"
	"github.com/d2lang/d2/d2target"
	"github.com/d2lang/d2/lib/log"
	"github.com/d2lang/d2/lib/textmeasure"
	"github.com/d2lang/util-go/go2"
)

// RenderOpts configures a render. Nil Seeds means the TALA defaults.
type RenderOpts struct {
	Seeds []int64
}

type RenderResult struct {
	SVG    string      `json:"svg"`
	Shapes []ShapeInfo `json:"shapes"`
	Edges  []EdgeInfo  `json:"edges"`
	Errors []Diag      `json:"errors"`
}

// ShapeInfo describes one laid out shape. Container describes the shape itself.
type ShapeInfo struct {
	ID        string `json:"id"`
	Parent    string `json:"parent"`
	Type      string `json:"type"`
	Container string `json:"container"` // "grid", "sequence", "hierarchy" or ""
	X         int    `json:"x"`
	Y         int    `json:"y"`
	W         int    `json:"w"`
	H         int    `json:"h"`
	Pinned    bool   `json:"pinned"`
	Pinnable  bool   `json:"pinnable"`
}

type EdgeInfo struct {
	ID     string       `json:"id"`
	Src    string       `json:"src"`
	Dst    string       `json:"dst"`
	Points [][2]float64 `json:"points"`
}

// Diag is a diagnostic. Line is 0-based and Col is a UTF-16 column.
type Diag struct {
	Line    int    `json:"line"`
	Col     int    `json:"col"`
	Message string `json:"message"`
}

var (
	rulerOnce sync.Once
	ruler     *textmeasure.Ruler
	rulerErr  error
)

func sharedRuler() (*textmeasure.Ruler, error) {
	rulerOnce.Do(func() { ruler, rulerErr = textmeasure.NewRuler() })
	return ruler, rulerErr
}

func isHierarchy(obj *d2graph.Object) bool {
	return obj != nil && strings.EqualFold(obj.Shape.Value, d2target.ShapeHierarchy)
}

// pinnable mirrors validatePositionsCompatibility in d2compiler.
func pinnable(obj *d2graph.Object) bool {
	p := obj.Parent
	if p == nil {
		return true
	}

	return !isHierarchy(p) && obj.OuterSequenceDiagram() == nil && p.GridColumns == nil && p.GridRows == nil
}

func container(obj *d2graph.Object) string {
	switch {
	case obj.IsGridDiagram():
		return "grid"
	case obj.IsSequenceDiagram():
		return "sequence"
	case isHierarchy(obj):
		return "hierarchy"
	}

	return ""
}

// Render compiles text, lays it out with TALA and renders it. Compile errors
// go in the result. The returned error is only for internal failures.
func Render(ctx context.Context, text string, opts RenderOpts) (RenderResult, error) {
	res := RenderResult{Shapes: []ShapeInfo{}, Edges: []EdgeInfo{}, Errors: []Diag{}}

	r, err := sharedRuler()
	if err != nil {
		return res, err
	}

	ctx = log.WithDefault(ctx) // D2 layout code requires a logger in ctx

	var talaOpts *d2talalayout.Options
	if opts.Seeds != nil {
		talaOpts = &d2talalayout.Options{Seeds: opts.Seeds}
	}

	d, g, err := d2lib.Compile(ctx, text, &d2lib.CompileOptions{
		UTF16Pos: true,
		Ruler:    r,
		Layout:   go2.Pointer("tala"),
		LayoutResolver: func(engine string) (d2graph.LayoutGraph, error) {
			if engine != "tala" {
				return nil, fmt.Errorf("unsupported layout engine %q", engine)
			}

			return func(ctx context.Context, g *d2graph.Graph) error {
				return d2talalayout.Layout(ctx, g, talaOpts)
			}, nil
		},
		RouterResolver: func(string) (d2graph.RouteEdges, error) { return d2talalayout.RouteEdges, nil },
	}, nil)
	if err != nil {
		var pe *d2parser.ParseError
		if errors.As(err, &pe) {
			for _, e := range pe.Errors {
				res.Errors = append(res.Errors, Diag{e.Range.Start.Line, e.Range.Start.Column, e.Message})
			}
		} else {
			res.Errors = append(res.Errors, Diag{Message: err.Error()})
		}

		return res, nil
	}

	svg, err := d2svg.Render(d, &d2svg.RenderOpts{
		ThemeID:     d.Config.ThemeID,
		DarkThemeID: d.Config.DarkThemeID,
		Sketch:      d.Config.Sketch,
	})
	if err != nil {
		return res, err
	}

	res.SVG = string(svg)

	objs := make(map[string]*d2graph.Object, len(g.Objects))
	for _, o := range g.Objects {
		objs[o.AbsID()] = o
	}

	for _, s := range d.Shapes {
		si := ShapeInfo{ID: s.ID, Type: s.Type, X: s.Pos.X, Y: s.Pos.Y, W: s.Width, H: s.Height}
		if o := objs[s.ID]; o != nil {
			if o.Parent != nil {
				si.Parent = o.Parent.AbsID()
			}

			si.Container = container(o)
			si.Pinned = o.Top != nil
			si.Pinnable = pinnable(o)
		}

		res.Shapes = append(res.Shapes, si)
	}

	for _, c := range d.Connections {
		ei := EdgeInfo{ID: c.ID, Src: c.Src, Dst: c.Dst, Points: [][2]float64{}}
		for _, p := range c.Route {
			ei.Points = append(ei.Points, [2]float64{p.X, p.Y})
		}

		res.Edges = append(res.Edges, ei)
	}

	return res, nil
}
