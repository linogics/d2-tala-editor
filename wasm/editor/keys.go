package editor

import (
	"errors"
	"slices"
	"strings"

	"github.com/d2lang/d2/d2ast"
	"github.com/d2lang/d2/d2format"
	"github.com/d2lang/d2/d2parser"
)

// The oracle refuses vars keys and appends duplicate class keys, so these edit the AST directly.

// ClassInfo is one entry of the top level classes map.
type ClassInfo struct {
	Name  string            `json:"name"`
	Attrs map[string]string `json:"attrs"`
}

// keyPath returns the plain segments of a map key, or nil for edges and other keys.
func keyPath(k *d2ast.Key) []string {
	if k == nil || k.Key == nil || len(k.Edges) > 0 {
		return nil
	}

	return k.Key.StringIDA()
}

// prefixOf reports whether a is a case-insensitive prefix of b, like D2 key matching.
func prefixOf(a, b []string) bool {
	return len(a) <= len(b) && slices.EqualFunc(a, b[:len(a)], strings.EqualFold)
}

// keyEdit parses text, applies fn to its AST and returns the formatted result.
// A result D2 refuses, for example a bad override color, becomes an error naming the problem.
func keyEdit(text string, fn func(m *d2ast.Map) error) (string, error) {
	if _, err := compile(text); err != nil {
		return "", err
	}

	m, err := d2parser.Parse("", strings.NewReader(text), nil)
	if err != nil {
		return "", err
	}

	if err := fn(m); err != nil {
		return "", err
	}

	out := d2format.Format(m)
	if _, err := compile(out); err != nil {
		if d, ok := BrokenDiag(err); ok {
			return "", errors.New(d.Message)
		}

		return "", err
	}

	return out, nil
}

// SetKey sets the scalar at path, or deletes the subtree at path when value is nil.
// An empty value writes a bare key, such as an empty class "classes: {x}".
func SetKey(text string, path []string, value *string) (string, error) {
	if len(path) == 0 {
		return "", errors.New("empty key path")
	}

	return keyEdit(text, func(m *d2ast.Map) error {
		if value == nil {
			deleteKey(m, path)

			return nil
		}

		leaves := findLeaves(m, path)
		if len(leaves) == 0 {
			return insertKey(m, path, *value)
		}
		// Later keys win in D2, so the last one keeps the value and the rest go.
		for _, l := range leaves[:len(leaves)-1] {
			l.m.Nodes = slices.DeleteFunc(l.m.Nodes, func(n d2ast.MapNodeBox) bool { return n.MapKey == l.k })
		}

		last := leaves[len(leaves)-1].k
		last.Value = d2ast.ValueBox{}
		last.Primary = d2ast.ScalarBox{}

		if *value != "" {
			last.Value = d2ast.MakeValueBox(d2ast.RawString(*value, false))
		}

		return nil
	})
}

type leaf struct {
	m *d2ast.Map
	k *d2ast.Key
}

// findLeaves returns every key whose full path is exactly path.
func findLeaves(m *d2ast.Map, path []string) []leaf {
	var out []leaf

	for _, n := range m.Nodes {
		ps := keyPath(n.MapKey)
		if ps == nil || !prefixOf(ps, path) {
			continue
		}

		if len(ps) == len(path) {
			out = append(out, leaf{m, n.MapKey})
		} else if n.MapKey.Value.Map != nil {
			out = append(out, findLeaves(n.MapKey.Value.Map, path[len(ps):])...)
		}
	}

	return out
}

// deleteKey removes every key at or below path and drops maps it leaves empty.
func deleteKey(m *d2ast.Map, path []string) {
	m.Nodes = slices.DeleteFunc(m.Nodes, func(n d2ast.MapNodeBox) bool {
		ps := keyPath(n.MapKey)
		if ps == nil {
			return false
		}

		if prefixOf(path, ps) {
			return true
		}

		if sub := n.MapKey.Value.Map; prefixOf(ps, path) && sub != nil {
			deleteKey(sub, path[len(ps):])

			return len(sub.Nodes) == 0
		}

		return false
	})
}

// insertKey appends path: value inside the deepest map the text already has on that path.
func insertKey(m *d2ast.Map, path []string, value string) error {
	for i := len(m.Nodes) - 1; i >= 0; i-- {
		k := m.Nodes[i].MapKey

		ps := keyPath(k)
		if ps == nil || len(ps) >= len(path) || !prefixOf(ps, path) {
			continue
		}

		if k.Value.Map != nil {
			return insertKey(k.Value.Map, path[len(ps):], value)
		}
		// A bare key such as an empty class is replaced by the full key appended below.
		if k.Value.Unbox() == nil && k.Primary.Unbox() == nil {
			m.Nodes = slices.Delete(m.Nodes, i, i+1)
		}
	}

	// Built as text and parsed, since d2format drops the braces of maps that have no source range.
	s := d2format.Format(d2ast.MakeKeyPath(path[len(path)-1:]))
	if value != "" {
		s += ": " + d2format.Format(d2ast.RawString(value, false))
	}

	for i := len(path) - 2; i >= 0; i-- {
		s = "{\n" + s + "\n}"
		s = d2format.Format(d2ast.MakeKeyPath(path[i:i+1])) + ": " + s
	}

	sub, err := d2parser.Parse("", strings.NewReader(s), nil)
	if err != nil {
		return err
	}

	m.Nodes = append(m.Nodes, sub.Nodes...)

	return nil
}

// RenameClass renames a class definition and every class reference to it.
func RenameClass(text, from, to string) (string, error) {
	if to == "" {
		return "", errors.New("empty class name")
	}

	return keyEdit(text, func(m *d2ast.Map) error {
		renameDef(m, []string{"classes", from}, to)
		d2ast.Walk(m, func(n d2ast.Node) bool {
			k, ok := n.(*d2ast.Key)
			if !ok || !isClassKey(k) {
				return true
			}

			if a := k.Value.Array; a != nil {
				for i, el := range a.Nodes {
					if s, ok := el.Unbox().(d2ast.String); ok && strings.EqualFold(s.ScalarString(), from) {
						a.Nodes[i] = d2ast.MakeArrayNodeBox(d2ast.RawString(to, false))
					}
				}
			} else if s := k.Value.ScalarBox().Unbox(); s != nil && strings.EqualFold(s.ScalarString(), from) {
				k.Value = d2ast.MakeValueBox(d2ast.RawString(to, false))
			}

			return true
		})

		return nil
	})
}

// isClassKey reports whether k assigns classes, as in "x.class: a" or "(a -> b)[0].class: a".
func isClassKey(k *d2ast.Key) bool {
	p := k.EdgeKey
	if p == nil {
		p = k.Key
	}

	return p != nil && len(p.Path) > 0 && strings.EqualFold(p.Path[len(p.Path)-1].Unbox().ScalarString(), "class")
}

// renameDef renames the segment of every key that ends at path, dotted or nested.
func renameDef(m *d2ast.Map, path []string, to string) {
	for _, n := range m.Nodes {
		ps := keyPath(n.MapKey)
		switch {
		case ps == nil:
		case prefixOf(path, ps):
			n.MapKey.Key.Path[len(path)-1] = d2ast.RawStringBox(to, true)
		case prefixOf(ps, path) && n.MapKey.Value.Map != nil:
			renameDef(n.MapKey.Value.Map, path[len(ps):], to)
		}
	}
}

// classes lists the top level class definitions with their attributes flattened to paths.
func classes(m *d2ast.Map) []ClassInfo {
	var out []ClassInfo

	byName := map[string]int{}

	var walk func(m *d2ast.Map, prefix []string)

	walk = func(m *d2ast.Map, prefix []string) {
		for _, n := range m.Nodes {
			ps := keyPath(n.MapKey)
			if ps == nil {
				continue
			}

			p := append(slices.Clone(prefix), ps...)
			if !strings.EqualFold(p[0], "classes") {
				continue
			}

			if len(p) >= 2 {
				name := strings.ToLower(p[1])

				i, ok := byName[name]
				if !ok {
					i = len(out)
					byName[name] = i

					out = append(out, ClassInfo{Name: p[1], Attrs: map[string]string{}})
				}

				if s := n.MapKey.Value.ScalarBox().Unbox(); len(p) > 2 && s != nil {
					out[i].Attrs[strings.Join(p[2:], ".")] = s.ScalarString()
				}
			}

			if sub := n.MapKey.Value.Map; sub != nil {
				walk(sub, p)
			}
		}
	}
	walk(m, nil)

	return out
}

// setClassList sets attr "class" on id to a D2 array, which the oracle would write as a quoted string.
func setClassList(text, id, list string) (string, error) {
	arr, err := d2parser.ParseValue(list)
	if err != nil {
		return "", err
	}

	const placeholder = "d2te-class-placeholder"

	out, err := Set(text, id, "class", ptr(placeholder))
	if err != nil {
		return "", err
	}

	return keyEdit(out, func(m *d2ast.Map) error {
		d2ast.Walk(m, func(n d2ast.Node) bool {
			if k, ok := n.(*d2ast.Key); ok && isClassKey(k) {
				if s := k.Value.ScalarBox().Unbox(); s != nil && s.ScalarString() == placeholder {
					k.Value = d2ast.MakeValueBox(arr)
				}
			}

			return true
		})

		return nil
	})
}

func ptr[T any](v T) *T { return &v }
