package editor

import (
	"strings"
	"unicode/utf16"

	"github.com/d2lang/d2/d2lsp"
	"github.com/d2lang/d2/d2themes"
	"github.com/d2lang/d2/d2themes/d2themescatalog"
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
	rs, _, err := d2lsp.GetRefRanges("index.d2", map[string]string{"index.d2": text}, nil, id)
	if err != nil {
		return nil, err
	}

	// GetRefRanges reports byte columns, so convert them to UTF-16.
	lines := strings.Split(text, "\n")
	col := func(line, c int) int {
		return len(utf16.Encode([]rune(lines[line][:c])))
	}

	out := make([]Range, 0, len(rs))
	for _, r := range rs {
		out = append(out, Range{
			r.Start.Line, col(r.Start.Line, r.Start.Column),
			r.End.Line, col(r.End.Line, r.End.Column),
		})
	}

	return out, nil
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
