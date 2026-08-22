package object

import (
	"testing"

	"github.com/gookit/color"
)

func colorBuiltinFn(t *testing.T, name string) BuiltinFunction {
	t.Helper()
	for _, b := range ColorBuiltins {
		if b.Name == name {
			if b.Fun == nil {
				t.Fatalf("color builtin %q has nil Fun", name)
			}
			return b.Fun
		}
	}
	t.Fatalf("color builtin %q not found", name)
	return nil
}

func TestColorRegistry(t *testing.T) {
	seen := make(map[string]bool)
	for _, b := range ColorBuiltins {
		if b.Name == "" || b.HelpStr == "" {
			t.Fatalf("color builtin %q missing Name or HelpStr", b.Name)
		}
		if seen[b.Name] {
			t.Errorf("duplicate color builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.Fun == nil {
			t.Errorf("color builtin %q has nil Fun", b.Name)
		}
	}
}

func TestColorConstants(t *testing.T) {
	constants := map[string]color.Color{
		"_normal":     color.Normal,
		"_red":        color.Red,
		"_cyan":       color.Cyan,
		"_gray":       color.Gray,
		"_blue":       color.Blue,
		"_black":      color.Black,
		"_green":      color.Green,
		"_white":      color.White,
		"_yellow":     color.Yellow,
		"_magenta":    color.Magenta,
		"_bold":       color.Bold,
		"_italic":     color.OpItalic,
		"_underlined": color.OpUnderscore,
	}
	for name, want := range constants {
		fn := colorBuiltinFn(t, name)
		res := fn()
		got, ok := res.(*Integer)
		if !ok {
			t.Errorf("%s() returned %T, want *Integer", name, res)
			continue
		}
		if got.Value != int64(want) {
			t.Errorf("%s() = %d, want %d", name, got.Value, int64(want))
		}
		res2 := fn(in(1))
		if _, isErr := res2.(*Error); !isErr {
			t.Errorf("%s(1) should error (takes no args), got %v", name, res2.Inspect())
		}
	}
}

func TestStyleBuiltin(t *testing.T) {
	styleFn := colorBuiltinFn(t, "_style")
	red := colorBuiltinFn(t, "_red")().(*Integer).Value
	white := colorBuiltinFn(t, "_white")().(*Integer).Value
	bold := colorBuiltinFn(t, "_bold")().(*Integer).Value

	res := styleFn(in(bold), in(red), in(white))
	m, ok := res.(*Map)
	if !ok {
		t.Fatalf("_style returned %T, want *Map", res)
	}
	typeVal := mapGetString(t, m, "t")
	ts, ok := typeVal.(*Stringo)
	if !ok || ts.Value != "color" {
		t.Errorf("_style 't' field = %#v, want STRING 'color'", typeVal)
	}
	if _, ok := mapGetString(t, m, "v").(*GoObj[color.Style]); !ok {
		t.Errorf("_style 'v' field should be GoObj[color.Style], got %T", mapGetString(t, m, "v"))
	}

	runBuiltinTestsFor(t, ColorBuiltins, "_style", []builtinTestCase{
		{name: "text not int", args: []Object{&Stringo{Value: "x"}, in(red), in(white)}, err: "PositionalTypeError"},
		{name: "fg not int", args: []Object{in(bold), &Float{Value: 31}, in(white)}, err: "PositionalTypeError"},
		{name: "bg not int", args: []Object{in(bold), in(red), &Stringo{Value: "w"}}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{in(bold), in(red)}, err: "InvalidArgCountError"},
	})

	unknown := styleFn(in(-999), in(-998), in(-997))
	if _, ok := unknown.(*Map); !ok {
		t.Errorf("_style with unknown colors should still return a MAP, got %T", unknown)
	}
}
