package editor

import "testing"

func TestSetLabelQuotes(t *testing.T) {
	cases := []struct{ in, id, attr, value, want string }{
		{"a", "a", "label", "My Label", `a: "My Label"`},
		{"a: {shape: circle}", "a", "label", "My Label", `a: "My Label" {shape: circle}`},
		{"a -> b", "(a -> b)[0]", "label", "go", `a -> b: "go"`},
		{"a -> b", "(a -> b)[0]", "target-arrowhead.label", "1", `a -> b: {target-arrowhead.label: "1"}`},
		{"a", "a", "label", `say "hi"`, `a: "say \"hi\""`},
		{"a\nb: Hand", "a", "label", "X", "a: \"X\"\nb: Hand"},
	}
	for _, c := range cases {
		v := c.value
		out, err := Set(c.in, c.id, c.attr, &v)
		want, _ := Format(c.want)

		if err != nil || out != want {
			t.Errorf("%q %s=%q: got %q want %q %v", c.in, c.attr, c.value, out, want, err)
		}
	}
}

func TestRenameKeepsLabel(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Foo", `bar: "Foo"`},
		{"Foo: {shape: circle}", `bar: "Foo" {shape: circle}`},
		{"Foo: My Label", "bar: My Label"},
		{"Foo: |md\n# hi\n|", "bar: |md\n# hi\n|"},
		{"classes: {c: {label: K}}\nFoo.class: c", "classes: {c: {label: K}}\nbar.class: c"},
	}
	for _, c := range cases {
		out, id, err := Rename(c.in, "Foo", "bar")
		want, _ := Format(c.want)

		if err != nil || out != want || id != "bar" {
			t.Errorf("%q: got %q %q want %q %v", c.in, out, id, want, err)
		}
	}
}

func TestAttrsOnlyWrittenLabel(t *testing.T) {
	g, err := compile("a: X\nb\nc: {shape: circle}")
	if err != nil {
		t.Fatal(err)
	}

	for id, want := range map[string]string{"a": "X", "b": "", "c": ""} {
		if got, ok := attrs(&lookup(g, id).Attributes)["label"]; got != want || ok != (want != "") {
			t.Errorf("%s: label %q %v want %q", id, got, ok, want)
		}
	}
}
