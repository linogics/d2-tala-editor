package editor

import "testing"

func TestCompletionsShape(t *testing.T) {
	text := "a: {shape: "

	items, err := Completions(text, 0, len(text))
	if err != nil {
		t.Fatal(err)
	}

	for _, it := range items {
		if it.Label == "cylinder" {
			return
		}
	}

	t.Fatalf("no cylinder in %v", items)
}

func TestRefRangesUTF16(t *testing.T) {
	rs, err := RefRanges("é -> b\nb: x", "b")
	if err != nil {
		t.Fatal(err)
	}

	var l0, l1 bool

	for _, r := range rs {
		l0 = l0 || (r.FromLine == 0 && r.FromCol == 5)
		l1 = l1 || r.FromLine == 1
	}

	if !l0 || !l1 {
		t.Fatalf("ranges %v", rs)
	}
}

func TestThemesHasDefault(t *testing.T) {
	var zero, dark bool

	for _, th := range Themes() {
		zero = zero || th.ID == 0
		dark = dark || th.Dark
	}

	if !zero || !dark {
		t.Fatal("missing default or dark theme")
	}
}
