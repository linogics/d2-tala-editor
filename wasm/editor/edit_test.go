package editor

import (
	"errors"
	"strings"
	"testing"
)

func must(t *testing.T) func(string, error) string {
	return func(s string, err error) string {
		t.Helper()

		if err != nil {
			t.Fatal(err)
		}

		return s
	}
}

func TestCreateRoot(t *testing.T) {
	out, id, err := Create("", "", "cylinder")
	if err != nil || id != "cylinder" || !strings.Contains(out, "shape: cylinder") {
		t.Fatalf("%q %q %v", out, id, err)
	}

	if _, err := compile(out); err != nil {
		t.Fatal(err)
	}
}

func TestCreateInParent(t *testing.T) {
	out, id, err := Create("c", "c", "")
	if err != nil || id != "c.rectangle" {
		t.Fatalf("%q %q %v", out, id, err)
	}
}

func TestSetAndDeleteAttr(t *testing.T) {
	in := "a -> b"
	want, _ := Format(in)
	v := "#ff0000"

	out, err := Set(in, "a", "style.fill", &v)
	if err != nil || !strings.Contains(out, "#ff0000") {
		t.Fatalf("%q %v", out, err)
	}

	out, err = Set(out, "a", "style.fill", nil)
	if err != nil || out != want {
		t.Fatalf("%q vs %q %v", out, want, err)
	}
}

func TestClearLabel(t *testing.T) {
	cases := []struct{ in, id, want string }{
		{"a -> b: hi", "(a -> b)[0]", "a -> b"},
		{"a -> b: hi {style.stroke: red}", "(a -> b)[0]", "a -> b: {style.stroke: red}"},
		{"a -> b\n(a -> b)[0].label: hi", "(a -> b)[0]", "a -> b"},
		{"a: A", "a", "a"},
		{"a: A {style.fill: red}", "a", "a: {style.fill: red}"},
	}
	for _, c := range cases {
		out, err := Set(c.in, c.id, "label", nil)
		want, _ := Format(c.want)

		if err != nil || out != want {
			t.Errorf("%q: got %q want %q %v", c.in, out, want, err)
		}
	}
}

func TestQuotedIDs(t *testing.T) {
	text, id, err := Create("", "", "")
	if err != nil {
		t.Fatal(err)
	}

	text, pid, err := Rename(text, id, "my node")
	if err != nil {
		t.Fatal(err)
	}

	text, cid, err := Create(text, pid, "")
	if err != nil {
		t.Fatal(err)
	}

	text, xid, err := Rename(text, cid, "a.b")
	if err != nil {
		t.Fatal(err)
	}

	text, _, err = Create(text+"\nCafé", "", "")
	if err != nil {
		t.Fatal(err)
	}

	text, _, err = Connect(text, "Café", xid)
	if err != nil {
		t.Fatal(err)
	}

	r := mustRender(t, text, RenderOpts{Seeds: []int64{1}})
	if len(r.Errors) != 0 {
		t.Fatalf("%v\n%s", r.Errors, text)
	}

	if byID(r, pid).W == 0 || byID(r, xid).W == 0 {
		t.Fatalf("ids %q %q not in shapes %v\n%s", pid, xid, r.Shapes, text)
	}
}

func TestMoveReparent(t *testing.T) {
	out, id, err := Move("a: {x}\nb", "a.x", "b")
	if err != nil || id != "b.x" {
		t.Fatalf("%q %q %v", out, id, err)
	}

	g, _ := compile(out)
	if lookup(g, "b.x") == nil || lookup(g, "a.x") != nil {
		t.Fatal(out)
	}
}

func TestRenameUpdatesEdges(t *testing.T) {
	out, id, err := Rename("a -> b", "a", "z")
	if err != nil || id != "z" || !strings.Contains(out, "z -> b") {
		t.Fatalf("%q %q %v", out, id, err)
	}
}

func TestConnectAndReconnect(t *testing.T) {
	out, eid, err := Connect("a\nb\nc", "a", "b")
	if err != nil || !strings.Contains(out, "a -> b") {
		t.Fatalf("%q %v", out, err)
	}

	dst := "c"

	out, err = Reconnect(out, eid, nil, &dst)
	if err != nil || !strings.Contains(out, "a -> c") {
		t.Fatalf("%q %v", out, err)
	}
}

func TestDeleteRemovesEdges(t *testing.T) {
	out, err := Delete("a -> b\nb -> c", "b")
	if err != nil || strings.Contains(out, "->") {
		t.Fatalf("%q %v", out, err)
	}
}

func TestEditOnBrokenTextFails(t *testing.T) {
	const bad = "a: {"

	s := "x"

	ops := map[string]func() error{
		"create":    func() error { _, _, e := Create(bad, "", ""); return e },
		"set":       func() error { _, e := Set(bad, "a", "shape", &s); return e },
		"delete":    func() error { _, e := Delete(bad, "a"); return e },
		"move":      func() error { _, _, e := Move(bad, "a", "b"); return e },
		"rename":    func() error { _, _, e := Rename(bad, "a", "b"); return e },
		"connect":   func() error { _, _, e := Connect(bad, "a", "b"); return e },
		"reconnect": func() error { _, e := Reconnect(bad, "(a -> b)[0]", &s, &s); return e },
		"format":    func() error { _, e := Format(bad); return e },
		"pin":       func() error { _, e := Pin(bad, "a", 1, 2); return e },
		"unpin":     func() error { _, e := Unpin(bad, "a"); return e },
	}
	for name, f := range ops {
		err := f()
		if !errors.Is(err, ErrBrokenText) {
			t.Errorf("%s: %v", name, err)
		}

		if _, ok := BrokenDiag(err); !ok {
			t.Errorf("%s: no diag", name)
		}
	}
}

func TestSetShapeDropsRefusedStyles(t *testing.T) {
	in := "a: {style: {3d: true; double-border: true; shadow: true}}"

	out := must(t)(Set(in, "a", "shape", ptr("cylinder")))
	if strings.Contains(out, "3d") || strings.Contains(out, "double-border") || !strings.Contains(out, "shadow") {
		t.Fatal(out)
	}

	out = must(t)(Set(in, "a", "shape", ptr("square")))
	if !strings.Contains(out, "3d") || !strings.Contains(out, "double-border") {
		t.Fatal(out)
	}
}

func TestRefusalIsShort(t *testing.T) {
	_, err := Set("a: {shape: image; icon: https://x.com/a.svg}", "a", "icon", nil)
	if err == nil || err.Error() != `image shape must include an "icon" field` {
		t.Fatalf("got %v", err)
	}
}

func TestSetReplacesIconAndNear(t *testing.T) {
	v := "https://b.com/y.svg"

	out, err := Set("a: {shape: image; icon: https://a.com/x.svg}", "a", "icon", &v)
	if err != nil || strings.Count(out, "icon:") != 1 || !strings.Contains(out, v) {
		t.Fatalf("icon: %q %v", out, err)
	}

	n := "top-right"

	out, err = Set("a.near: top-left", "a", "near", &n)
	if err != nil || out != "a.near: top-right\n" {
		t.Fatalf("near: %q %v", out, err)
	}

	// The oracle writes the new value into the earlier "c" map, the old flat key must go.
	b := "b"

	out, err = Set("a\nb\nc\nc.near: a", "c", "near", &b)
	if err != nil || strings.Count(out, "near") != 1 || !strings.Contains(out, "near: b") {
		t.Fatalf("near object: %q %v", out, err)
	}
}

func TestShapeChangeDropsConstraints(t *testing.T) {
	r := "rectangle"

	out, err := Set("t: {shape: sql_table; id: int {constraint: primary_key}; n: varchar}", "t", "shape", &r)
	if err != nil || out != "t: {shape: rectangle; id: int; n: varchar}\n" {
		t.Fatalf("%q %v", out, err)
	}
}
