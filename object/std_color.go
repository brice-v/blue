package object

import "github.com/gookit/color"

var ColorBuiltins = NewBuiltinSliceType{
	{Name: "_style", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("style", len(args), 3, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("style", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("style", 2, INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != INTEGER_OBJ {
				return newPositionalTypeError("style", 3, INTEGER_OBJ, args[2].Type())
			}
			arg1, arg2, arg3 := args[0].(*Integer).Value, args[1].(*Integer).Value, args[2].(*Integer).Value
			textStyle := color.Color(arg1)
			fgActualColor := color.Color(arg2)
			fgColor := fgActualColor.ToFg()
			bgActualColor := color.Color(arg3)
			bgColor := bgActualColor.ToBg()
			textStyleName := textStyle.Name()
			fgColorName := fgColor.Name()
			bgColorName := bgColor.Name()
			fgActualColorName := fgActualColor.Name()
			bgActualColorName := bgActualColor.Name()
			s := color.New()
			unknown := "unknown"
			if textStyleName != unknown {
				s.Add(textStyle)
			}
			if fgColorName != unknown || fgActualColorName != unknown {
				s.Add(fgColor)
			}
			if bgColorName != unknown || bgActualColorName != unknown {
				s.Add(bgColor)
			}
			return CreateBasicMapObjectForGoObj("color", NewGoObj(s))
		},
		HelpStr: helpStrArgs{
			explanation: "`style` returns an object to be used in printing that affects the stylized output",
			signature:   "style(text: int=normal, fg_color: int=normal, bg_color: int=normal) -> {t: 'color', v: GoObj[color.Style]}",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "style(fg_color=magenta, bg_color=white) => color style object",
		}.String(),
	}},
	{Name: "_normal", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("normal", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Normal)}
		},
		HelpStr: helpStrArgs{
			explanation: "`normal` returns the int version of the normal color",
			signature:   "normal() -> int",
			errors:      "InvalidArgCount",
			example:     "normal() -> int",
		}.String(),
	}},
	{Name: "_red", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("red", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Red)}
		},
		HelpStr: helpStrArgs{
			explanation: "`red` returns the int version of the red color",
			signature:   "red() -> int",
			errors:      "InvalidArgCount",
			example:     "red() -> int",
		}.String(),
	}},
	{Name: "_cyan", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("cyan", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Cyan)}
		},
		HelpStr: helpStrArgs{
			explanation: "`cyan` returns the int version of the cyan color",
			signature:   "cyan() -> int",
			errors:      "InvalidArgCount",
			example:     "cyan() -> int",
		}.String(),
	}},
	{Name: "_gray", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("gray", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Gray)}
		},
		HelpStr: helpStrArgs{
			explanation: "`gray` returns the int version of the gray color",
			signature:   "gray() -> int",
			errors:      "InvalidArgCount",
			example:     "gray() -> int",
		}.String(),
	}},
	{Name: "_blue", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("blue", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Blue)}
		},
		HelpStr: helpStrArgs{
			explanation: "`blue` returns the int version of the blue color",
			signature:   "blue() -> int",
			errors:      "InvalidArgCount",
			example:     "blue() -> int",
		}.String(),
	}},
	{Name: "_black", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("black", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Black)}
		},
		HelpStr: helpStrArgs{
			explanation: "`black` returns the int version of the black color",
			signature:   "black() -> int",
			errors:      "InvalidArgCount",
			example:     "black() -> int",
		}.String(),
	}},
	{Name: "_green", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("green", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Green)}
		},
		HelpStr: helpStrArgs{
			explanation: "`green` returns the int version of the green color",
			signature:   "green() -> int",
			errors:      "InvalidArgCount",
			example:     "green() -> int",
		}.String(),
	}},
	{Name: "_white", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("white", len(args), 0, "")
			}
			return &Integer{Value: int64(color.White)}
		},
		HelpStr: helpStrArgs{
			explanation: "`white` returns the int version of the white color",
			signature:   "white() -> int",
			errors:      "InvalidArgCount",
			example:     "white() -> int",
		}.String(),
	}},
	{Name: "_yellow", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("yellow", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Yellow)}
		},
		HelpStr: helpStrArgs{
			explanation: "`yellow` returns the int version of the yellow color",
			signature:   "yellow() -> int",
			errors:      "InvalidArgCount",
			example:     "yellow() -> int",
		}.String(),
	}},
	{Name: "_magenta", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("magenta", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Magenta)}
		},
		HelpStr: helpStrArgs{
			explanation: "`magenta` returns the int version of the magenta color",
			signature:   "magenta() -> int",
			errors:      "InvalidArgCount",
			example:     "magenta() -> int",
		}.String(),
	}},
	{Name: "_bold", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("bold", len(args), 0, "")
			}
			return &Integer{Value: int64(color.Bold)}
		},
		HelpStr: helpStrArgs{
			explanation: "`bold` returns the int version of the bold color",
			signature:   "bold() -> int",
			errors:      "InvalidArgCount",
			example:     "bold() -> int",
		}.String(),
	}},
	{Name: "_italic", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("italic", len(args), 0, "")
			}
			return &Integer{Value: int64(color.OpItalic)}
		},
		HelpStr: helpStrArgs{
			explanation: "`italic` returns the int version of the italic color",
			signature:   "italic() -> int",
			errors:      "InvalidArgCount",
			example:     "italic() -> int",
		}.String(),
	}},
	{Name: "_underlined", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("underlined", len(args), 0, "")
			}
			return &Integer{Value: int64(color.OpUnderscore)}
		},
		HelpStr: helpStrArgs{
			explanation: "`underlined` returns the int version of the underlined color",
			signature:   "underlined() -> int",
			errors:      "InvalidArgCount",
			example:     "underlined() -> int",
		}.String(),
	}},
}
