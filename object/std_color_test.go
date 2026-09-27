package object

import (
	c "image/color"
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
	expected := []string{"_style", "_normal", "_bold", "_italic", "_underlined", "_new", "_color_map"}
	for _, name := range expected {
		if !seen[name] {
			t.Errorf("expected color builtin %q to be registered", name)
		}
	}
}

func TestStyleConstants(t *testing.T) {
	constants := map[string]color.Color{
		"_normal":     color.Normal,
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
	normal := colorBuiltinFn(t, "_normal")().(*Integer).Value
	bold := colorBuiltinFn(t, "_bold")().(*Integer).Value
	red := NewGoObj(Red)
	white := NewGoObj(White)

	res := styleFn(in(bold), red, white)
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

	withIntBg := styleFn(in(normal), red, in(int64(color.White)))
	if _, ok := withIntBg.(*Map); !ok {
		t.Errorf("_style with an integer background should return a MAP, got %T", withIntBg)
	}

	unknownRGB := NewGoObj(c.RGBA{R: 1, G: 2, B: 3, A: 255})
	unknown := styleFn(in(bold), unknownRGB, in(int64(-997)))
	if _, ok := unknown.(*Map); !ok {
		t.Errorf("_style with unknown colors should still return a MAP, got %T", unknown)
	}

	runBuiltinTestsFor(t, ColorBuiltins, "_style", []builtinTestCase{
		{name: "text not int", args: []Object{&Stringo{Value: "x"}, red, white}, err: "PositionalTypeError"},
		{name: "fg not color", args: []Object{in(bold), in(1), white}, err: "PositionalTypeError"},
		{name: "bg wrong type", args: []Object{in(bold), red, &Stringo{Value: "w"}}, err: "PositionalTypeError"},
		{name: "two args", args: []Object{in(bold), red}, err: "InvalidArgCountError"},
	})
}

func TestNewColorBuiltin(t *testing.T) {
	newFn := colorBuiltinFn(t, "_new")
	res := newFn(in(230), in(41), in(55), in(255))
	got, ok := res.(*GoObj[c.RGBA])
	if !ok {
		t.Fatalf("_new returned %T, want *GoObj[c.RGBA]", res)
	}
	if got.Value != Red {
		t.Errorf("_new(230, 41, 55, 255) = %v, want %v", got.Value, Red)
	}

	runBuiltinTestsFor(t, ColorBuiltins, "_new", []builtinTestCase{
		{name: "too few args", args: []Object{in(1), in(2), in(3)}, err: "InvalidArgCountError"},
		{name: "too many args", args: []Object{in(1), in(2), in(3), in(4), in(5)}, err: "InvalidArgCountError"},
		{name: "first not int", args: []Object{&Float{Value: 1}, in(2), in(3), in(4)}, err: "PositionalTypeError"},
		{name: "second not int", args: []Object{in(1), &Float{Value: 2}, in(3), in(4)}, err: "PositionalTypeError"},
		{name: "third not int", args: []Object{in(1), in(2), &Float{Value: 3}, in(4)}, err: "PositionalTypeError"},
		{name: "fourth not int", args: []Object{in(1), in(2), in(3), &Float{Value: 4}}, err: "PositionalTypeError"},
	})
}

func TestColorMapBuiltin(t *testing.T) {
	mapFn := colorBuiltinFn(t, "_color_map")
	res := mapFn()
	m, ok := res.(*Map)
	if !ok {
		t.Fatalf("_color_map returned %T, want *Map", res)
	}

	expected := map[string]c.RGBA{
		"beige":       Beige,
		"black":       Black,
		"blank":       Blank,
		"blue":        Blue,
		"brown":       Brown,
		"cyan":        Cyan,
		"dark_blue":   DarkBlue,
		"dark_brown":  DarkBrown,
		"dark_gray":   DarkGray,
		"dark_green":  DarkGreen,
		"dark_grey":   DarkGray,
		"dark_purple": DarkPurple,
		"gold":        Gold,
		"gray":        Gray,
		"green":       Green,
		"grey":        Gray,
		"light_gray":  LightGray,
		"light_grey":  LightGray,
		"lime":        Lime,
		"magenta":     Magenta,
		"maroon":      Maroon,
		"orange":      Orange,
		"pink":        Pink,
		"purple":      Purple,
		"ray_white":   RayWhite,
		"red":         Red,
		"sky_blue":    SkyBlue,
		"violet":      Violet,
		"white":       White,
		"yellow":      Yellow,
	}
	if m.Pairs.Len() != len(expected) {
		t.Errorf("_color_map has %d entries, want %d", m.Pairs.Len(), len(expected))
	}
	for name, want := range expected {
		val := mapGetString(t, m, name)
		goObj, ok := val.(*GoObj[c.RGBA])
		if !ok {
			t.Errorf("_color_map[%q] = %T, want *GoObj[c.RGBA]", name, val)
			continue
		}
		if goObj.Value != want {
			t.Errorf("_color_map[%q] = %v, want %v", name, goObj.Value, want)
		}
	}

	runBuiltinTestsFor(t, ColorBuiltins, "_color_map", []builtinTestCase{
		{name: "takes no args", args: []Object{in(1)}, err: "InvalidArgCountError"},
	})
}
