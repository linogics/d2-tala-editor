package editor

import (
	"strings"
	"unicode/utf16"

	"github.com/d2lang/d2/d2ast"
	"github.com/d2lang/d2/d2ir"
	"github.com/d2lang/d2/d2lsp"
	"github.com/d2lang/d2/d2parser"
	"github.com/d2lang/d2/d2themes"
	"github.com/d2lang/d2/d2themes/d2themescatalog"
	"github.com/d2lang/d2/lib/memfs"
)

// CompletionItem is d2lsp's completion item.
type CompletionItem = d2lsp.CompletionItem

// Range is a 0-based span with UTF-16 columns.
type Range struct {
	FromLine int `json:"fromLine"`
	FromCol  int `json:"fromCol"`
	ToLine   int `json:"toLine"`
	ToCol    int `json:"toCol"`
}

type ThemeInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Dark bool   `json:"dark"`
	// Colors maps override slots such as "B1" to the theme's own color.
	Colors map[string]string `json:"colors"`
}

func Completions(text string, line, col int) ([]CompletionItem, error) {
	return d2lsp.GetCompletionItemsUTF16(text, line, col)
}

func RefRanges(text, id string) ([]Range, error) {
	all, err := RefRangesAll(text, []string{id})

	return all[id], err
}

// RefRangesAll compiles the text once and returns the ref ranges of every id.
// An id that is not found maps to an empty list.
func RefRangesAll(text string, ids []string) (map[string][]Range, error) {
	ast, err := d2parser.Parse("index.d2", strings.NewReader(text), nil)
	if err != nil {
		return nil, err
	}

	fs, err := memfs.New(map[string]string{"index.d2": text})
	if err != nil {
		return nil, err
	}

	root, _, err := d2ir.Compile(ast, &d2ir.CompileOptions{FS: fs})
	if err != nil {
		return nil, err
	}

	// d2ast reports byte columns, so convert them to UTF-16.
	lines := strings.Split(text, "\n")
	col := func(line, c int) int {
		return len(utf16.Encode([]rune(lines[line][:c])))
	}

	out := make(map[string][]Range, len(ids))
	for _, id := range ids {
		rs := []Range{}
		for _, r := range refsOf(root, id) {
			rs = append(rs, Range{
				r.Start.Line, col(r.Start.Line, r.Start.Column),
				r.End.Line, col(r.End.Line, r.End.Column),
			})
		}

		out[id] = rs
	}

	return out, nil
}

// refsOf mirrors d2lsp.GetRefRanges on an already compiled map.
func refsOf(m *d2ir.Map, key string) []d2ast.Range {
	mk, err := d2parser.ParseMapKey(key)
	if err != nil || (mk.Key == nil && len(mk.Edges) == 0) {
		return nil
	}

	var f *d2ir.Field

	if mk.Key != nil {
		for _, p := range mk.Key.Path {
			if m == nil {
				return nil
			}

			if f = m.GetField(p.Unbox()); f == nil {
				return nil
			}

			m = f.Map()
		}
	}

	var out []d2ast.Range

	if len(mk.Edges) > 0 {
		if m == nil {
			return nil
		}

		for _, eid := range d2ir.NewEdgeIDs(mk) {
			for _, e := range m.GetEdges(eid, nil, nil) {
				for _, ref := range e.References {
					out = append(out, ref.AST().GetRange())
				}
			}
		}

		return out
	}

	for _, ref := range f.References {
		out = append(out, ref.AST().GetRange())
	}

	return out
}

func Themes() []ThemeInfo {
	var out []ThemeInfo

	for _, t := range d2themescatalog.LightCatalog {
		out = append(out, ThemeInfo{t.ID, t.Name, false, colors(t)})
	}

	for _, t := range d2themescatalog.DarkCatalog {
		out = append(out, ThemeInfo{t.ID, t.Name, true, colors(t)})
	}

	return out
}

func colors(t d2themes.Theme) map[string]string {
	c, n := t.Colors, t.Colors.Neutrals

	return map[string]string{
		"N1": n.N1, "N2": n.N2, "N3": n.N3, "N4": n.N4, "N5": n.N5, "N6": n.N6, "N7": n.N7,
		"B1": c.B1, "B2": c.B2, "B3": c.B3, "B4": c.B4, "B5": c.B5, "B6": c.B6,
		"AA2": c.AA2, "AA4": c.AA4, "AA5": c.AA5, "AB4": c.AB4, "AB5": c.AB5,
	}
}
