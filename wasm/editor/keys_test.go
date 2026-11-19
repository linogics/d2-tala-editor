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
