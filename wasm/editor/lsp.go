package editor

import (
	"strings"
	"unicode/utf16"

	"github.com/d2lang/d2/d2lsp"
	"github.com/d2lang/d2/d2themes/d2themescatalog"
	"github.com/d2lang/d2/lib/urlenc"
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

func Encode(text string) (string, error) { return urlenc.Encode(text) }

func Decode(s string) (string, error) { return urlenc.Decode(s) }

func Themes() []ThemeInfo {
	var out []ThemeInfo

	for _, t := range d2themescatalog.LightCatalog {
		out = append(out, ThemeInfo{t.ID, t.Name, false})
	}

	for _, t := range d2themescatalog.DarkCatalog {
		out = append(out, ThemeInfo{t.ID, t.Name, true})
	}

	return out
}
