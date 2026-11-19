package editor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// apiError is the error payload sent to the UI.
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Diag    *Diag  `json:"diag"`
}

type textArgs struct {
	Text string `json:"text"`
}

type idArgs struct {
	Text string `json:"text"`
	ID   string `json:"id"`
}

type beforeArgs struct {
	Text   string `json:"text"`
	ID     string `json:"id"`
	Before string `json:"before"`
}

type textID struct {
	Text string `json:"text"`
	ID   string `json:"id"`
}

type textOnly struct {
	Text string `json:"text"`
}

type handler func(args []byte) (any, error)

// op decodes the args into A and runs fn.
func op[A any](fn func(a A) (any, error)) handler {
	return func(args []byte) (any, error) {
		var a A

		if err := json.Unmarshal(args, &a); err != nil {
			return nil, fmt.Errorf("%w: %w", errBadArgs, err)
		}

		return fn(a)
	}
}

var errBadArgs = errors.New("bad args")

// textRes wraps an edit that returns only text.
func textRes(text string, err error) (any, error) {
	if err != nil {
		return nil, err
	}

	return textOnly{text}, nil
}

// textIDRes wraps an edit that returns text and an id.
func textIDRes(text, id string, err error) (any, error) {
	if err != nil {
		return nil, err
	}

	return textID{text, id}, nil
}

// list turns a nil slice into an empty one so it marshals as [].
func list[T any](s []T, err error) (any, error) {
	if err != nil {
		return nil, err
	}

	if s == nil {
		s = []T{}
	}

	return s, nil
}

var handlers = map[string]handler{
	"render": op(func(a struct {
		Text  string  `json:"text"`
		Seeds []int64 `json:"seeds"`
	}) (any, error) {
		return Render(context.Background(), a.Text, RenderOpts{Seeds: a.Seeds})
	}),
	"create": op(func(a struct{ Text, Parent, Shape string }) (any, error) {
		t, id, err := Create(a.Text, a.Parent, a.Shape)

		return textIDRes(t, id, err)
	}),
	"insert": op(func(a struct{ Text, Parent, Template string }) (any, error) {
		t, id, err := Insert(a.Text, a.Parent, a.Template)

		return textIDRes(t, id, err)
	}),
	"set": op(func(a struct {
		Text, ID, Attr string
		Value          *string
	}) (any, error) {
		return textRes(Set(a.Text, a.ID, a.Attr, a.Value))
	}),
	"delete": op(func(a idArgs) (any, error) { return textRes(Delete(a.Text, a.ID)) }),
	"move": op(func(a struct{ Text, ID, Parent string }) (any, error) {
		t, id, err := Move(a.Text, a.ID, a.Parent)

		return textIDRes(t, id, err)
	}),
	"rename": op(func(a struct{ Text, ID, Name string }) (any, error) {
		t, id, err := Rename(a.Text, a.ID, a.Name)

		return textIDRes(t, id, err)
	}),
	"connect": op(func(a struct{ Text, Src, Dst string }) (any, error) {
		t, id, err := Connect(a.Text, a.Src, a.Dst)

		return textIDRes(t, id, err)
	}),
	"reconnect": op(func(a struct {
		Text, ID string
		Src, Dst *string
	}) (any, error) {
		return textRes(Reconnect(a.Text, a.ID, a.Src, a.Dst))
	}),
	"pin": op(func(a struct {
		Text, ID string
		X, Y     int
	}) (any, error) {
		return textRes(Pin(a.Text, a.ID, a.X, a.Y))
	}),
	"unpin":  op(func(a idArgs) (any, error) { return textRes(Unpin(a.Text, a.ID)) }),
	"pinAll": op(func(a textArgs) (any, error) { return textRes(PinAll(context.Background(), a.Text)) }),
	"reorder": op(func(a beforeArgs) (any, error) {
		return textRes(Reorder(a.Text, a.ID, a.Before))
	}),
	"reorderEdge": op(func(a beforeArgs) (any, error) {
		return textRes(ReorderEdge(a.Text, a.ID, a.Before))
	}),
	"format": op(func(a textArgs) (any, error) { return textRes(Format(a.Text)) }),
	"completions": op(func(a struct {
		Text      string
		Line, Col int
	}) (any, error) {
		return list(Completions(a.Text, a.Line, a.Col))
	}),
	"refRanges": op(func(a idArgs) (any, error) { return list(RefRanges(a.Text, a.ID)) }),
	"setKey": op(func(a struct {
		Text  string
		Path  []string
		Value *string
	}) (any, error) {
		return textRes(SetKey(a.Text, a.Path, a.Value))
	}),
	"renameClass": op(func(a struct{ Text, From, To string }) (any, error) {
		return textRes(RenameClass(a.Text, a.From, a.To))
	}),
	"themes": op(func(struct{}) (any, error) { return list(Themes(), nil) }),
}

// Dispatch runs one op with JSON args and returns a JSON envelope.
func Dispatch(opName string, args []byte) (out []byte) {
	defer func() {
		if r := recover(); r != nil {
			out = envelope(nil, fmt.Errorf("panic: %v", r))
		}
	}()

	h, ok := handlers[opName]
	if !ok {
		return envelope(nil, fmt.Errorf("%w: %s", errUnknownOp, opName))
	}

	res, err := h(args)

	return envelope(res, err)
}

var errUnknownOp = errors.New("unknown op")

func envelope(res any, err error) []byte {
	var v any = struct {
		OK     bool `json:"ok"`
		Result any  `json:"result"`
	}{true, res}

	if err != nil {
		v = struct {
			OK    bool     `json:"ok"`
			Error apiError `json:"error"`
		}{false, toAPIError(err)}
	}

	b, mErr := json.Marshal(v)
	if mErr != nil {
		return []byte(`{"ok":false,"error":{"code":"internal","message":"marshal failed","diag":null}}`)
	}

	return b
}

func toAPIError(err error) apiError {
	e := apiError{Code: "internal", Message: err.Error()}

	switch {
	case errors.Is(err, ErrBrokenText):
		e.Code = "broken_text"

		if d, ok := BrokenDiag(err); ok {
			e.Diag = &d
		}
	case errors.Is(err, ErrNotPinnable):
		e.Code = "not_pinnable"
	case errors.Is(err, ErrNotReorderable):
		e.Code = "not_reorderable"
	case errors.Is(err, errBadArgs):
		e.Code = "bad_args"
	case errors.Is(err, errUnknownOp):
		e.Code = "unknown_op"
	}

	return e
}
