package editor

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

// D2 renders these shapes under another type name.
var aliases = map[string]string{"square": "rectangle", "circle": "oval"}

type tmpl struct{ Key, Label, Group, Template string }

func loadJSON(t *testing.T, name string, v any) {
	t.Helper()

	b, err := os.ReadFile("../../web/public/" + name)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

func TestEveryTemplateInserts(t *testing.T) {
	var ts []tmpl

	loadJSON(t, "templates.json", &ts)

	if len(ts) != 28 {
		t.Fatalf("got %d templates", len(ts))
	}

	for _, tp := range ts {
		for _, c := range []struct{ text, parent string }{{"", ""}, {"c", "c"}} {
			out, id, err := Insert(c.text, c.parent, tp.Template)
			if err != nil {
				t.Fatalf("%s: %v", tp.Key, err)
			}

			r := mustRender(t, out, RenderOpts{Seeds: []int64{1}})
			if len(r.Errors) != 0 {
				t.Fatalf("%s: %v\n%s", tp.Key, r.Errors, out)
			}

			s := byID(r, id)
			if s.ID == "" {
				t.Fatalf("%s: id %q not rendered\n%s", tp.Key, id, out)
			}

			m := regexp.MustCompile(`shape:\s*([\w-]+)`).FindStringSubmatch(tp.Template)
			if m == nil {
				continue
			}

			want := m[1]
			if a, ok := aliases[want]; ok {
				want = a
			}

			if s.Type != want {
				t.Fatalf("%s: type %q want %q", tp.Key, s.Type, want)
			}
		}
	}
}

func TestEveryExampleRenders(t *testing.T) {
	var ex []struct{ Name, Code string }

	loadJSON(t, "examples.json", &ex)

	if len(ex) < 7 || ex[0].Name != "Hello" {
		t.Fatalf("got %d examples", len(ex))
	}

	for _, e := range ex {
		if r := mustRender(t, e.Code, RenderOpts{Seeds: []int64{1}}); len(r.Errors) != 0 {
			t.Fatalf("%s: %v", e.Name, r.Errors)
		}
	}
}

func TestInsertUniqueKey(t *testing.T) {
	out, a, err := Insert("", "", "$: {shape: cylinder}")
	if err != nil {
		t.Fatal(err)
	}

	_, b, err := Insert(out, "", "$: {shape: cylinder}")
	if err != nil || a == b {
		t.Fatalf("%q %q %v", a, b, err)
	}
}

func TestInsertKeyFromInlineShape(t *testing.T) {
	_, id, err := Insert("", "", "$: {shape: rectangle}")
	if err != nil {
		t.Fatal(err)
	}

	if id != "rectangle" {
		t.Fatalf("id = %q, want rectangle", id)
	}
}
