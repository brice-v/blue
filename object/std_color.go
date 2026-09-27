package object

import (
	c "image/color"

	"github.com/gookit/color"
)

var ColorBuiltins = []*Builtin{
	{
		Name: "_style",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("style", len(args), 3, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("style", 1, INTEGER_OBJ, args[0].Type())
			}
			foregroundColor, ok := args[1].(*GoObj[c.RGBA])
			if !ok {
				return newPositionalTypeError("style", 2, "COLOR", args[1].Type())
			}
			textStyle := color.Color(args[0].(*Integer).Value)
			fgActualColor := color.RGB(foregroundColor.Value.R, foregroundColor.Value.G, foregroundColor.Value.B).Color()
			fgColor := fgActualColor.ToFg()
			backgroundColor, ok := args[2].(*GoObj[c.RGBA])
			if !ok {
				if args[2].Type() != INTEGER_OBJ {
					return newPositionalTypeError("style", 3, "INTEGER or COLOR", args[2].Type())
				}
			}
			var bgActualColor color.Color
			if ok {
				bgActualColor = color.RGB(backgroundColor.Value.R, backgroundColor.Value.G, backgroundColor.Value.B).Color()
			} else {
				bgActualColor = color.Color(args[2].(*Integer).Value)
			}
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
	},
	{
		Name: "_normal",
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
	},
	{
		Name: "_bold",
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
	},
	{
		Name: "_italic",
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
	},
	{
		Name: "_underlined",
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
	},
	{
		Name: "_new",
		Fun: func(args ...Object) Object {
			if len(args) != 4 {
				return newInvalidArgCountError("new", len(args), 4, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("new", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("new", 2, INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != INTEGER_OBJ {
				return newPositionalTypeError("new", 3, INTEGER_OBJ, args[2].Type())
			}
			if args[3].Type() != INTEGER_OBJ {
				return newPositionalTypeError("new", 4, INTEGER_OBJ, args[3].Type())
			}
			return NewGoObj(NewColor(
				uint8(args[0].(*Integer).Value),
				uint8(args[1].(*Integer).Value),
				uint8(args[2].(*Integer).Value),
				uint8(args[3].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`new` returns a color based on the rgba values",
			signature:   "new(r: int(u8), g: int(u8), b: int(u8), a: int(u8)) -> GoObj[rl.Color]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "new(230,230,240,1) => GoObj[rl.Color]",
		}.String(),
	},
	{
		Name: "_color_map",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("color_map", len(args), 0, "")
			}
			beige := NewGoObj(Beige)
			black := NewGoObj(Black)
			blank := NewGoObj(Blank)
			blue := NewGoObj(Blue)
			brown := NewGoObj(Brown)
			cyan := NewGoObj(Cyan)
			darkBlue := NewGoObj(DarkBlue)
			darkBrown := NewGoObj(DarkBrown)
			darkGray := NewGoObj(DarkGray)
			darkGreen := NewGoObj(DarkGreen)
			darkPurple := NewGoObj(DarkPurple)
			gold := NewGoObj(Gold)
			gray := NewGoObj(Gray)
			green := NewGoObj(Green)
			lightGray := NewGoObj(LightGray)
			lime := NewGoObj(Lime)
			magenta := NewGoObj(Magenta)
			mapObj := NewOrderedMap[string, Object]()
			maroon := NewGoObj(Maroon)
			orange := NewGoObj(Orange)
			pink := NewGoObj(Pink)
			purple := NewGoObj(Purple)
			rayWhite := NewGoObj(RayWhite)
			red := NewGoObj(Red)
			skyBlue := NewGoObj(SkyBlue)
			violet := NewGoObj(Violet)
			white := NewGoObj(White)
			yellow := NewGoObj(Yellow)

			mapObj.Set("beige", beige)
			mapObj.Set("black", black)
			mapObj.Set("blank", blank)
			mapObj.Set("blue", blue)
			mapObj.Set("brown", brown)
			mapObj.Set("cyan", cyan)
			mapObj.Set("dark_blue", darkBlue)
			mapObj.Set("dark_brown", darkBrown)
			mapObj.Set("dark_gray", darkGray)
			mapObj.Set("dark_green", darkGreen)
			mapObj.Set("dark_grey", darkGray)
			mapObj.Set("dark_purple", darkPurple)
			mapObj.Set("gold", gold)
			mapObj.Set("gray", gray)
			mapObj.Set("green", green)
			mapObj.Set("grey", gray)
			mapObj.Set("light_gray", lightGray)
			mapObj.Set("light_grey", lightGray)
			mapObj.Set("lime", lime)
			mapObj.Set("magenta", magenta)
			mapObj.Set("maroon", maroon)
			mapObj.Set("orange", orange)
			mapObj.Set("pink", pink)
			mapObj.Set("purple", purple)
			mapObj.Set("ray_white", rayWhite)
			mapObj.Set("red", red)
			mapObj.Set("sky_blue", skyBlue)
			mapObj.Set("violet", violet)
			mapObj.Set("white", white)
			mapObj.Set("yellow", yellow)
			return CreateMapObjectForGoMap(*mapObj)
		},
		HelpStr: helpStrArgs{
			explanation: "`color_map` returns a map with all the colors available as well as a function 'new' to generate a color from an rgba value",
			signature:   "color_map() -> map[str:GoObj[Color]|fun(r,g,b,a)->GoObj[Color]]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "color_map() => map[str:GoObj[Color]|fun(r,g,b,a)->GoObj[Color]]",
		}.String(),
	},
}

func NewColor(r, g, b, a uint8) c.RGBA {
	return c.RGBA{r, g, b, a}
}

// From Raylib
var (
	Beige      = NewColor(211, 176, 131, 255)
	Black      = NewColor(0, 0, 0, 255)
	Blank      = NewColor(0, 0, 0, 0)
	Blue       = NewColor(0, 121, 241, 255)
	Brown      = NewColor(127, 106, 79, 255)
	Cyan       = NewColor(0, 255, 255, 255)
	DarkBlue   = NewColor(0, 82, 172, 255)
	DarkBrown  = NewColor(76, 63, 47, 255)
	DarkGray   = NewColor(80, 80, 80, 255)
	DarkGreen  = NewColor(0, 117, 44, 255)
	DarkPurple = NewColor(112, 31, 126, 255)
	Gold       = NewColor(255, 203, 0, 255)
	Gray       = NewColor(130, 130, 130, 255)
	Green      = NewColor(0, 228, 48, 255)
	LightGray  = NewColor(200, 200, 200, 255)
	Lime       = NewColor(0, 158, 47, 255)
	Magenta    = NewColor(255, 0, 255, 255)
	Maroon     = NewColor(190, 33, 55, 255)
	Orange     = NewColor(255, 161, 0, 255)
	Pink       = NewColor(255, 109, 194, 255)
	Purple     = NewColor(200, 122, 255, 255)
	RayWhite   = NewColor(245, 245, 245, 255)
	Red        = NewColor(230, 41, 55, 255)
	SkyBlue    = NewColor(102, 191, 255, 255)
	Violet     = NewColor(135, 60, 190, 255)
	White      = NewColor(255, 255, 255, 255)
	Yellow     = NewColor(253, 249, 0, 255)
)
