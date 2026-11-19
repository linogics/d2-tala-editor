package editor

import (
	"strings"
	"testing"
)

func mustKey(t *testing.T, text string, path []string, v *string) string {
	t.Helper()

	out, err := SetKey(text, path, v)
	if err != nil {
		t.Fatal(err)
	}

	return out
}

func TestSetKeyVars(t *testing.T) {
	out := mustKey(t, "a -> b\n", []string{"vars", "d2-config", "sketch"}, ptr("true"))
	out = mustKey(t, out, []string{"vars", "d2-config", "theme-overrides", "B1"}, ptr("#ff0000"))

	r := mustRender(t, out, RenderOpts{})
	if r.Config == nil || r.Config.Sketch == nil || !*r.Config.Sketch {
		t.Fatalf("sketch not set:\n%s", out)
	}

	if o := r.Config.ThemeOverrides; o == nil || o.B1 == nil || *o.B1 != "#ff0000" {
		t.Fatalf("override not set:\n%s", out)
	}

	if !strings.Contains(strings.ToLower(r.SVG), "#ff0000") {
		t.Fatal("override missing from svg")
	}

	if strings.Count(out, "vars") != 1 {
		t.Fatalf("second vars block:\n%s", out)
	}

	out = mustKey(t, out, []string{"vars", "d2-config", "sketch"}, nil)

	out = mustKey(t, out, []string{"vars", "d2-config", "theme-overrides"}, nil)
	if strings.Contains(out, "vars") {
		t.Fatalf("empty maps kept:\n%s", out)
	}
}

func TestSetKeyReplacesDotted(t *testing.T) {
	out := mustKey(t, "classes.x.style.fill: red\nclasses: {x.style.fill: blue}\n", []string{"classes", "x", "style", "fill"}, ptr("green"))
	if strings.Count(out, "fill") != 1 || !strings.Contains(out, "green") {
		t.Fatalf("got:\n%s", out)
	}
}

func TestSetKeyRejectsBadValue(t *testing.T) {
	if _, err := SetKey("a\n", []string{"vars", "d2-config", "theme-overrides", "B1"}, ptr("nope")); err == nil {
		t.Fatal("bad color accepted")
	}
}

func TestClassesAndRename(t *testing.T) {
	text := "classes: {x: {style.fill: red}; y}\na.class: x\nb: {class: [x; y]}\na -> b: {class: x}\n"

	cs := mustRender(t, text, RenderOpts{}).Classes
	if len(cs) != 2 || cs[0].Name != "x" || cs[0].Attrs["style.fill"] != "red" || cs[1].Name != "y" {
		t.Fatalf("classes %+v", cs)
	}

	out, err := RenameClass(text, "x", "z")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(out, "x") || strings.Count(out, "z") != 4 {
		t.Fatalf("got:\n%s", out)
	}

	out = mustKey(t, out, []string{"classes", "z"}, nil)
	if strings.Contains(out, "fill") {
		t.Fatalf("class kept:\n%s", out)
	}
}

func TestSetClassList(t *testing.T) {
	text := "classes: {x; y}\na -> b\n"

	out, err := Set(text, "(a -> b)[0]", "class", ptr("[x; y]"))
	if err != nil {
		t.Fatal(err)
	}

	r := mustRender(t, out, RenderOpts{})
	if got := r.Edges[0].Attrs["class"]; got != "x;y" {
		t.Fatalf("class %q:\n%s", got, out)
	}
}

func TestSetKeyBareClass(t *testing.T) {
	out := mustKey(t, "a\n", []string{"classes", "x"}, ptr(""))

	cs := mustRender(t, out, RenderOpts{}).Classes
	if len(cs) != 1 || cs[0].Name != "x" {
		t.Fatalf("classes %+v:\n%s", cs, out)
	}

	out = mustKey(t, out, []string{"classes", "x", "style", "fill"}, ptr("red"))
	if strings.Count(out, "x") != 1 {
		t.Fatalf("bare key kept:\n%s", out)
	}
}

func TestSetClassRemoves(t *testing.T) {
	cases := []struct {
		name, text, id string
		value          *string
		want           string
	}{
		{"dotted", "classes: {x}\na.class: x\n", "a", nil, ""},
		{"map", "classes: {x}\na: {class: x}\n", "a", nil, ""},
		{"nested", "classes: {x}\np: {a.class: x}\n", "p.a", nil, ""},
		{"array to one", "classes: {x; y}\na.class: [x; y]\n", "a", ptr("y"), "y"},
		{"array to none", "classes: {x; y}\na.class: [x; y]\n", "a", nil, ""},
		{"edge", "classes: {x}\na -> b: {class: x}\n", "(a -> b)[0]", nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := Set(c.text, c.id, "class", c.value)
			if err != nil {
				t.Fatal(err)
			}

			r := mustRender(t, out, RenderOpts{})
			got := ""

			for _, s := range r.Shapes {
				if s.ID == c.id {
					got = s.Attrs["class"]
				}
			}

			if len(r.Edges) > 0 && c.id == "(a -> b)[0]" {
				got = r.Edges[0].Attrs["class"]
			}

			if got != c.want {
				t.Fatalf("class %q, want %q:\n%s", got, c.want, out)
			}
		})
	}
}

func TestDeleteClassDropsRefs(t *testing.T) {
	text := "classes: {x; y}\na.class: x\nb: {class: [x; y]}\nc -> d: {class: x}\n"

	out, err := DeleteClass(text, "x")
	if err != nil {
		t.Fatal(err)
	}

	want := "classes: {y}\na\nb: {class: [y]}\nc -> d\n"
	if out != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
}

func TestRenameClassRefusesExisting(t *testing.T) {
	if _, err := RenameClass("classes: {a; b}\nx.class: a", "a", "B"); err == nil {
		t.Fatal("rename onto existing class must fail")
	}

	out, err := RenameClass("classes: {a; b}\nx.class: a", "a", "A")
	if err != nil || out != "classes: {A; b}\nx.class: A\n" {
		t.Fatalf("case rename: %q %v", out, err)
	}
}
