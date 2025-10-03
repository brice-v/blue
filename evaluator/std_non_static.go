//go:build !static
// +build !static

package evaluator

import (
	"blue/consts"
	"blue/lib"
	"blue/object"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() {
	_std_mods["ui"] = StdModFileAndBuiltins{File: lib.ReadStdFileToString("ui.b"), Builtins: _ui_builtin_map}
	_std_mods["gg"] = StdModFileAndBuiltins{File: lib.ReadStdFileToString("gg.b"), Builtins: _gg_builtin_map}
}

func setupBuiltinsWithEvaluator(name string, newE *Evaluator) {
	switch name {
	case "http":
		_http_builtin_map.Put("_handle", createHttpHandleBuiltin(newE, false))
		_http_builtin_map.Put("_handle_use", createHttpHandleBuiltin(newE, true))
		_http_builtin_map.Put("_handle_ws", createHttpHandleWSBuiltin(newE))
	case "ui":
		_ui_builtin_map.Put("_button", createUIButtonBuiltin(newE))
		_ui_builtin_map.Put("_check_box", createUICheckBoxBuiltin(newE))
		_ui_builtin_map.Put("_radio_group", createUIRadioBuiltin(newE))
		_ui_builtin_map.Put("_option_select", createUIOptionSelectBuiltin(newE))
		_ui_builtin_map.Put("_form", createUIFormBuiltin(newE))
		_ui_builtin_map.Put("_toolbar_action", createUIToolbarAction(newE))
	}
}

var _gg_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_init_window": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("init_window", len(args), 3, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("init_window", 1, object.INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("init_window", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newPositionalTypeError("init_window", 3, object.STRING_OBJ, args[2].Type())
			}
			width := int32(args[0].(*object.Integer).Value)
			height := int32(args[1].(*object.Integer).Value)
			title := args[2].(*object.Stringo).Value
			rl.InitWindow(width, height, title)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`init_window` initalizes the gg graphics window with the given width, height, and title",
			signature:   "init_window(width: int=800, height: int=600, title: str='gg - example app') -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "init_window() => null",
		}.String(),
	},
	"_close_window": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("close_window", len(args), 0, "")
			}
			rl.CloseWindow()
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`close_window` closes the gg graphics window",
			signature:   "close_window() -> null",
			errors:      "InvalidArgCount",
			example:     "close_window() => null",
		}.String(),
	},
	"_window_should_close": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("window_should_close", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.WindowShouldClose())
		},
		HelpStr: helpStrArgs{
			explanation: "`window_should_close` returns true if the window is closing with an 'ESC' press or 'X' close button on the window",
			signature:   "window_should_close() -> bool",
			errors:      "InvalidArgCount",
			example:     "window_should_close() => false",
		}.String(),
	},
	"_get_screen_width": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_screen_width", len(args), 0, "")
			}
			return &object.Integer{Value: int64(rl.GetScreenWidth())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_screen_width` gets the screen width as an int",
			signature:   "get_screen_width() -> int",
			errors:      "InvalidArgCount",
			example:     "get_screen_width() => 800",
		}.String(),
	},
	"_get_screen_height": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_screen_height", len(args), 0, "")
			}
			return &object.Integer{Value: int64(rl.GetScreenHeight())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_screen_height` gets the screen height as an int",
			signature:   "get_screen_height() -> int",
			errors:      "InvalidArgCount",
			example:     "get_screen_height() => 800",
		}.String(),
	},
	"_begin_drawing": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("begin_drawing", len(args), 0, "")
			}
			rl.BeginDrawing()
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`begin_drawing` sets up the drawing canvas to start drawing",
			signature:   "begin_drawing() -> null",
			errors:      "InvalidArgCount",
			example:     "begin_drawing() => null",
		}.String(),
	},
	"_end_drawing": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("end_drawing", len(args), 0, "")
			}
			rl.EndDrawing()
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`end_drawing` ends canvas drawing and swaps buffers (double buffering)",
			signature:   "end_drawing() -> null",
			errors:      "InvalidArgCount",
			example:     "end_drawing() => null",
		}.String(),
	},
	"_clear_background": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("clear_background", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("clear_background", 1, object.GO_OBJ, args[0].Type())
			}
			goObj, ok := args[0].(*object.GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("clear_background", 1, "rl.Color", args[0])
			}
			rl.ClearBackground(goObj.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`clear_background` sets the background color to the given color",
			signature:   "clear_background(color: GoObj[rl.Color]=color.white) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "clear_background() => null",
		}.String(),
	},
	"_color_map": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("color_map", len(args), 0, "")
			}
			mapObj := object.NewOrderedMap[string, object.Object]()
			lightGray := NewGoObj(rl.LightGray)
			gray := NewGoObj(rl.Gray)
			darkGray := NewGoObj(rl.DarkGray)
			yellow := NewGoObj(rl.Yellow)
			gold := NewGoObj(rl.Gold)
			orange := NewGoObj(rl.Orange)
			pink := NewGoObj(rl.Pink)
			red := NewGoObj(rl.Red)
			maroon := NewGoObj(rl.Maroon)
			green := NewGoObj(rl.Green)
			lime := NewGoObj(rl.Lime)
			darkGreen := NewGoObj(rl.DarkGreen)
			skyBlue := NewGoObj(rl.SkyBlue)
			blue := NewGoObj(rl.Blue)
			darkBlue := NewGoObj(rl.DarkBlue)
			purple := NewGoObj(rl.Purple)
			violet := NewGoObj(rl.Violet)
			darkPurple := NewGoObj(rl.DarkPurple)
			beige := NewGoObj(rl.Beige)
			brown := NewGoObj(rl.Brown)
			darkBrown := NewGoObj(rl.DarkBrown)
			white := NewGoObj(rl.White)
			black := NewGoObj(rl.Black)
			blank := NewGoObj(rl.Blank)
			magenta := NewGoObj(rl.Magenta)
			rayWhite := NewGoObj(rl.RayWhite)
			newColor := &object.Builtin{
				Fun: func(args ...object.Object) object.Object {
					if len(args) != 4 {
						return newInvalidArgCountError("new_color", len(args), 4, "")
					}
					if args[0].Type() != object.INTEGER_OBJ {
						return newPositionalTypeError("new_color", 1, object.INTEGER_OBJ, args[0].Type())
					}
					if args[1].Type() != object.INTEGER_OBJ {
						return newPositionalTypeError("new_color", 2, object.INTEGER_OBJ, args[1].Type())
					}
					if args[2].Type() != object.INTEGER_OBJ {
						return newPositionalTypeError("new_color", 3, object.INTEGER_OBJ, args[2].Type())
					}
					if args[3].Type() != object.INTEGER_OBJ {
						return newPositionalTypeError("new_color", 4, object.INTEGER_OBJ, args[3].Type())
					}
					return NewGoObj(rl.NewColor(
						uint8(args[0].(*object.Integer).Value),
						uint8(args[1].(*object.Integer).Value),
						uint8(args[2].(*object.Integer).Value),
						uint8(args[3].(*object.Integer).Value)))
				},
				HelpStr: helpStrArgs{
					explanation: "`new_color` returns a color based on the rgba values",
					signature:   "new_color(r: int(u8), g: int(u8), b: int(u8), a: int(u8)) -> GoObj[rl.Color]",
					errors:      "InvalidArgCount,PositionalType",
					example:     "new_color(230,230,240,1) => GoObj[rl.Color]",
				}.String(),
			}
			mapObj.Set("light_gray", lightGray)
			mapObj.Set("gray", gray)
			mapObj.Set("dark_gray", darkGray)
			mapObj.Set("yellow", yellow)
			mapObj.Set("gold", gold)
			mapObj.Set("orange", orange)
			mapObj.Set("pink", pink)
			mapObj.Set("red", red)
			mapObj.Set("maroon", maroon)
			mapObj.Set("green", green)
			mapObj.Set("lime", lime)
			mapObj.Set("dark_green", darkGreen)
			mapObj.Set("sky_blue", skyBlue)
			mapObj.Set("blue", blue)
			mapObj.Set("dark_blue", darkBlue)
			mapObj.Set("purple", purple)
			mapObj.Set("violet", violet)
			mapObj.Set("dark_purple", darkPurple)
			mapObj.Set("beige", beige)
			mapObj.Set("brown", brown)
			mapObj.Set("dark_brown", darkBrown)
			mapObj.Set("white", white)
			mapObj.Set("black", black)
			mapObj.Set("blank", blank)
			mapObj.Set("magenta", magenta)
			mapObj.Set("ray_white", rayWhite)
			mapObj.Set("new", newColor)
			return object.CreateMapObjectForGoMap(*mapObj)
		},
		HelpStr: helpStrArgs{
			explanation: "`color_map` returns a map with all the colors available as well as a function 'new' to generate a color from an rgba value",
			signature:   "color_map() -> map[str:GoObj[rl.Color]|fun(r,g,b,a)->GoObj[rl.Color]]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "color_map() => map[str:GoObj[rl.Color]|fun(r,g,b,a)->GoObj[rl.Color]]",
		}.String(),
	},
	"_draw_text": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 5 {
				return newInvalidArgCountError("draw_text", len(args), 5, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("draw_text", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_text", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_text", 3, object.INTEGER_OBJ, args[2].Type())
			}
			if args[3].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_text", 4, object.INTEGER_OBJ, args[3].Type())
			}
			if args[4].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_text", 5, object.GO_OBJ, args[4].Type())
			}
			goObj, ok := args[4].(*object.GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_text", 5, "rl.Color", args[4])
			}
			text := args[0].(*object.Stringo).Value
			posX := int32(args[1].(*object.Integer).Value)
			posY := int32(args[2].(*object.Integer).Value)
			fontSize := int32(args[3].(*object.Integer).Value)
			rl.DrawText(text, posX, posY, fontSize, goObj.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_text` draws text on the canvas with the given text at (x,y) with font_size, and color",
			signature:   "draw_text(text: str, pos_x: int=0, pos_y: int=0, font_size: int=20, text_color: GO_OBJ[rl.Color]=color.black) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "draw_text('Hello World!') => null",
		}.String(),
	},
	"_draw_texture": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("draw_texture", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture", 1, object.GO_OBJ, args[0].Type())
			}
			tex, ok := args[0].(*object.GoObj[rl.Texture2D])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture", 1, "rl.Texture2D", args[0])
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_texture", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_texture", 3, object.INTEGER_OBJ, args[2].Type())
			}
			if args[3].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture", 4, object.GO_OBJ, args[3].Type())
			}
			tint, ok := args[3].(*object.GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture", 4, "rl.Color", args[3])
			}
			posX := int32(args[1].(*object.Integer).Value)
			posY := int32(args[2].(*object.Integer).Value)
			rl.DrawTexture(tex.Value, posX, posY, tint.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_texture` draws the 2d texture on the canvas at (x,y) with given tint tint",
			signature:   "draw_texture(texture: GO_OBJ[rl.Texture2D], pos_x: int=0, pos_y: int=0, tint: GO_OBJ[rl.Color]=color.white) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "draw_texture(texture) => null",
		}.String(),
	},
	"_draw_texture_pro": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 6 {
				return newInvalidArgCountError("draw_texture_pro", len(args), 6, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 1, object.GO_OBJ, args[0].Type())
			}
			tex, ok := args[0].(*object.GoObj[rl.Texture2D])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 1, "rl.Texture2D", args[0])
			}
			if args[1].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 2, object.GO_OBJ, args[1].Type())
			}
			srcRect, ok := args[1].(*object.GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 2, "rl.Rectangle", args[1])
			}
			if args[2].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 3, object.GO_OBJ, args[2].Type())
			}
			dstRect, ok := args[2].(*object.GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 3, "rl.Rectangle", args[2])
			}
			if args[3].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 4, object.GO_OBJ, args[3].Type())
			}
			origin, ok := args[3].(*object.GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 4, "rl.Rectangle", args[3])
			}
			if args[4].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("draw_texture_pro", 5, object.FLOAT_OBJ, args[4].Type())
			}
			rotation := float32(args[4].(*object.Float).Value)
			if args[5].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_texture_pro", 6, object.GO_OBJ, args[5].Type())
			}
			tint, ok := args[5].(*object.GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 6, "rl.Color", args[5])
			}
			rl.DrawTexturePro(tex.Value, srcRect.Value, dstRect.Value, origin.Value, rotation, tint.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_texture_pro` draws a part of the 2d texture on the canvas with the given source_rec, dest_rec, origin, rotation, and tint",
			signature:   "draw_texture_pro(texture: GO_OBJ[rl.Texture2D], source_rec: GO_OBJ[rl.Rectangle]=Rectangle(), dest_rec: GO_OBJ[rl.Rectangle]=Rectangle(), origin: GO_OBJ[rl.Vector2]=Vector2(), rotation: float=0.0, tint=color.white) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "draw_texture_pro(texture) => null",
		}.String(),
	},
	"_draw_rectangle": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 5 {
				return newInvalidArgCountError("draw_rectangle", len(args), 5, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_rectangle", 1, object.INTEGER_OBJ, args[0].Type())
			}
			posx := int32(args[0].(*object.Integer).Value)
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_rectangle", 2, object.INTEGER_OBJ, args[1].Type())
			}
			posy := int32(args[1].(*object.Integer).Value)
			if args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_rectangle", 3, object.INTEGER_OBJ, args[2].Type())
			}
			width := int32(args[2].(*object.Integer).Value)
			if args[3].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("draw_rectangle", 4, object.INTEGER_OBJ, args[3].Type())
			}
			height := int32(args[3].(*object.Integer).Value)
			if args[4].Type() != object.GO_OBJ {
				return newPositionalTypeError("draw_rectangle", 5, object.GO_OBJ, args[4].Type())
			}
			color, ok := args[4].(*object.GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_texture_pro", 4, "rl.Color", args[4])
			}
			rl.DrawRectangle(posx, posy, width, height, color.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_rectangle` draws a rectangle at the given position with width and height",
			signature:   "draw_rectangle(posx: int, posy: int, width: int, height: int, color=color.black) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "draw_rectangle() (used as Rectangle().draw(color))=> null",
		}.String(),
	},
	"_set_target_fps": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_target_fps", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("set_target_fps", 1, object.INTEGER_OBJ, args[0].Type())
			}
			fps := int32(args[0].(*object.Integer).Value)
			rl.SetTargetFPS(fps)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_target_fps` sets the target fps to the given integer",
			signature:   "set_target_fps(fps: int) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_target_fps(60) => null",
		}.String(),
	},
	"_set_exit_key": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_exit_key", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("set_exit_key", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			rl.SetExitKey(key)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_exit_key` sets the exit key to the given key (integer)",
			signature:   "set_exit_key(key: int) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_exit_key(key.Q) => null",
		}.String(),
	},
	"_is_key_up": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_key_up", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("is_key_up", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			return nativeToBooleanObject(rl.IsKeyUp(key))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_key_up` returns true if the given key is up",
			signature:   "is_key_up(key: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_key_up(key.Q) => false",
		}.String(),
	},
	"_is_key_down": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_key_down", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("is_key_down", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			return nativeToBooleanObject(rl.IsKeyDown(key))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_key_down` returns true if the given key is down",
			signature:   "is_key_down(key: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_key_down(key.Q) => false",
		}.String(),
	},
	"_is_key_pressed": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_key_pressed", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("is_key_pressed", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			return nativeToBooleanObject(rl.IsKeyPressed(key))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_key_pressed` returns true if the given key is pressed",
			signature:   "is_key_pressed(key: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_key_pressed(key.Q) => false",
		}.String(),
	},
	"_is_key_released": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_key_released", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("is_key_released", 1, object.INTEGER_OBJ, args[0].Type())
			}
			key := int32(args[0].(*object.Integer).Value)
			return nativeToBooleanObject(rl.IsKeyReleased(key))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_key_released` returns true if the given key is released",
			signature:   "is_key_released(key: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_key_released(key.Q) => false",
		}.String(),
	},
	"_load_texture": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("load_texture", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("load_texture", 1, object.STRING_OBJ, args[0].Type())
			}
			fname := args[0].(*object.Stringo).Value
			if IsEmbed {
				s := fname
				if strings.HasPrefix(s, "./") {
					s = strings.TrimLeft(s, "./")
				}
				fileData, err := Files.ReadFile(consts.EMBED_FILES_PREFIX + s)
				if err != nil {
					ext := filepath.Ext(s)
					img := rl.LoadImageFromMemory(ext, fileData, int32(len(fileData)))
					img1 := rl.LoadTextureFromImage(img)
					return NewGoObj(img1)
				}
			}
			img := rl.LoadTexture(fname)
			return NewGoObj(img)
		},
		HelpStr: helpStrArgs{
			explanation: "`load_texture` loads a 2d texture image resource and returns an object referencing it",
			signature:   "load_texture(path: str) -> GoObj[rl.Texture2D]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "load_texture(key.Q) => GoObj[rl.Texture2D]",
		}.String(),
	},
	"_rectangle": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("Rectangle", len(args), 4, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Rectangle", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Rectangle", 2, object.FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Rectangle", 3, object.FLOAT_OBJ, args[2].Type())
			}
			if args[3].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Rectangle", 4, object.FLOAT_OBJ, args[3].Type())
			}
			x := float32(args[0].(*object.Float).Value)
			y := float32(args[1].(*object.Float).Value)
			width := float32(args[2].(*object.Float).Value)
			height := float32(args[3].(*object.Float).Value)
			rect := rl.NewRectangle(x, y, width, height)
			return NewGoObj(rect)
		},
		HelpStr: helpStrArgs{
			explanation: "`rectangle` returns a rectangle with position (x,y) and w,h",
			signature:   "rectangle(x: float=0.0, y: float=0.0, width: float=0.0, height: float=0.0) -> GoObj[rl.Rectangle]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "rectangle() => GoObj[rl.Rectangle]",
		}.String(),
	},
	"_rectangle_check_collision": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("rectangle_check_collision", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("rectangle_check_collision", 1, object.GO_OBJ, args[0].Type())
			}
			rec1, ok := args[0].(*object.GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("rectangle_check_collision", 1, "rl.Rectangle", args[0])
			}
			if args[1].Type() != object.GO_OBJ {
				return newPositionalTypeError("rectangle_check_collision", 2, object.GO_OBJ, args[1].Type())
			}
			rec2, ok := args[1].(*object.GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("rectangle_check_collision", 2, "rl.Rectangle", args[1])
			}
			return nativeToBooleanObject(rl.CheckCollisionRecs(rec1.Value, rec2.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`rectangle_check_collision` returns true if the 2 rectangles collide",
			signature:   "rectangle_check_collision(rec1: GoObj[rl.Rectangle], rec2: GoObj[rl.Rectangle]) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "rectangle_check_collision(rec1, rec2) => true",
		}.String(),
	},
	"_vector2": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("Vector2", len(args), 2, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector2", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector2", 2, object.FLOAT_OBJ, args[1].Type())
			}
			x := float32(args[0].(*object.Float).Value)
			y := float32(args[1].(*object.Float).Value)
			vec2 := rl.NewVector2(x, y)
			return NewGoObj(vec2)
		},
		HelpStr: helpStrArgs{
			explanation: "`vector2` returns a vector2 with x,y",
			signature:   "vector2(x: float=0.0, y: float=0.0) -> GoObj[rl.Vector2]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "vector2() => GoObj[rl.Vector2]",
		}.String(),
	},
	"_vector3": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("Vector3", len(args), 3, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector3", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector3", 2, object.FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector3", 3, object.FLOAT_OBJ, args[2].Type())
			}
			x := float32(args[0].(*object.Float).Value)
			y := float32(args[1].(*object.Float).Value)
			z := float32(args[2].(*object.Float).Value)
			vec3 := rl.NewVector3(x, y, z)
			return NewGoObj(vec3)
		},
		HelpStr: helpStrArgs{
			explanation: "`vector3` returns a vector3 with x,y,z",
			signature:   "vector3(x: float=0.0, y: float=0.0, z: float=0.0) -> GoObj[rl.Vector3]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "vector3() => GoObj[rl.Vector3]",
		}.String(),
	},
	"_vector4": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("Vector4", len(args), 4, "")
			}
			if args[0].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector4", 1, object.FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector4", 2, object.FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector4", 3, object.FLOAT_OBJ, args[2].Type())
			}
			if args[3].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Vector4", 4, object.FLOAT_OBJ, args[3].Type())
			}
			x := float32(args[0].(*object.Float).Value)
			y := float32(args[1].(*object.Float).Value)
			z := float32(args[2].(*object.Float).Value)
			w := float32(args[3].(*object.Float).Value)
			vec4 := rl.NewVector4(x, y, z, w)
			return NewGoObj(vec4)
		},
		HelpStr: helpStrArgs{
			explanation: "`vector4` returns a vector4 with x,y,z,w",
			signature:   "vector4(x: float=0.0, y: float=0.0, z: float=0.0, w: float=0.0) -> GoObj[rl.Vector4]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "vector4() => GoObj[rl.Vector4]",
		}.String(),
	},
	"_camera2d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("Camera2D", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera2D", 1, object.GO_OBJ, args[0].Type())
			}
			offset, ok := args[0].(*object.GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera2D", 1, "rl.Vector2", args[0])
			}
			if args[1].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera2D", 2, object.GO_OBJ, args[1].Type())
			}
			target, ok := args[1].(*object.GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera2D", 2, "rl.Vector2", args[1])
			}
			if args[2].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Camera2D", 3, object.FLOAT_OBJ, args[2].Type())
			}
			if args[3].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Camera2D", 4, object.FLOAT_OBJ, args[3].Type())
			}
			rotation := float32(args[2].(*object.Float).Value)
			zoom := float32(args[3].(*object.Float).Value)
			cam2d := rl.NewCamera2D(offset.Value, target.Value, rotation, zoom)
			return NewGoObj(cam2d)
		},
		HelpStr: helpStrArgs{
			explanation: "`camera2d` returns a 2D camera with offset, target, rotation, and zoom",
			signature:   "camera2d(offset: GoObj[rl.Vector2], target: GoObj[rl.Vector2], rotation: float=0.0, zoom: float=1.0) -> GoObj[rl.Camera2D]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "camera2d() => GoObj[rl.Camera2D]",
		}.String(),
	},
	"_begin_mode2d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("begin_mode2d", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("begin_mode2d", 1, object.GO_OBJ, args[0].Type())
			}
			cam, ok := args[0].(*object.GoObj[rl.Camera2D])
			if !ok {
				return newPositionalTypeErrorForGoObj("begin_mode2d", 1, "rl.Camera2D", args[0])
			}
			rl.BeginMode2D(cam.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`begin_mode2d` initializes 2d with a custom 2d camera",
			signature:   "begin_mode2d(cam: GoObj[rl.Camera2D]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "begin_mode2d(cam) => null",
		}.String(),
	},
	"_end_mode2d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("end_mode2d", len(args), 0, "")
			}
			rl.EndMode2D()
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`end_mode2d` ends 2d camera mode",
			signature:   "end_mode2d() -> null",
			errors:      "InvalidArgCount",
			example:     "end_mode2d() => null",
		}.String(),
	},
	"_camera3d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 5 {
				return newInvalidArgCountError("Camera3D", len(args), 5, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera3D", 1, object.GO_OBJ, args[0].Type())
			}
			position, ok := args[0].(*object.GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera3D", 1, "rl.Vector3", args[0])
			}
			if args[1].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera3D", 2, object.GO_OBJ, args[1].Type())
			}
			target, ok := args[1].(*object.GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera3D", 2, "rl.Vector3", args[1])
			}
			if args[2].Type() != object.GO_OBJ {
				return newPositionalTypeError("Camera3D", 3, object.GO_OBJ, args[2].Type())
			}
			up, ok := args[2].(*object.GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("Camera3D", 3, "rl.Vector3", args[2])
			}
			if args[3].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("Camera3D", 4, object.FLOAT_OBJ, args[3].Type())
			}
			if args[4].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("Camera3D", 5, object.INTEGER_OBJ, args[4].Type())
			}
			fovy := float32(args[3].(*object.Float).Value)
			projection := rl.CameraProjection(args[4].(*object.Integer).Value)
			cam3d := rl.NewCamera3D(position.Value, target.Value, up.Value, fovy, projection)
			return NewGoObj(cam3d)
		},
		HelpStr: helpStrArgs{
			explanation: "`camera3d` returns a 3D camera with position, target, up, fovy, and projection",
			signature:   "camera3d(position: GoObj[rl.Vector3], target: GoObj[rl.Vector3], up: GoObj[rl.Vector3], fovy: float=0.0, projection: int[CameraProjection.Perspective|Orthographic]=CameraProjection.Perspective) -> GoObj[rl.Camera3D]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "camera3d(Vector(z=0.0), Vector(z=0.0), Vector(z=0.0)) => GoObj[rl.Camera3D]",
		}.String(),
	},
	"_begin_mode3d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("begin_mode3d", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("begin_mode3d", 1, object.GO_OBJ, args[0].Type())
			}
			cam, ok := args[0].(*object.GoObj[rl.Camera3D])
			if !ok {
				return newPositionalTypeErrorForGoObj("begin_mode3d", 1, "rl.Camera3D", args[0])
			}
			rl.BeginMode3D(cam.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`begin_mode3d` begins 3d camera mode with the custom camera",
			signature:   "begin_mode3d(cam: GoObj[rl.Camera3D]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "begin_mode3d() => null",
		}.String(),
	},
	"_end_mode3d": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("end_mode3d", len(args), 0, "")
			}
			rl.EndMode3D()
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`end_mode3d` ends 3d camera mode",
			signature:   "end_mode3d() -> null",
			errors:      "InvalidArgCount",
			example:     "end_mode3d() => null",
		}.String(),
	},
	"_init_audio_device": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("init_audio_device", len(args), 0, "")
			}
			rl.InitAudioDevice()
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`init_audio_device` initalizes the audio device and context",
			signature:   "init_audio_device() -> null",
			errors:      "InvalidArgCount",
			example:     "init_audio_device() => null",
		}.String(),
	},
	"_close_audio_device": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("close_audio_device", len(args), 0, "")
			}
			rl.CloseAudioDevice()
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`close_audio_device` closes the audio device and context",
			signature:   "close_audio_device() -> null",
			errors:      "InvalidArgCount",
			example:     "close_audio_device() => null",
		}.String(),
	},
	"_load_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("load_music", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("load_music", 1, object.STRING_OBJ, args[0].Type())
			}
			fname := args[0].(*object.Stringo).Value
			if IsEmbed {
				s := fname
				if strings.HasPrefix(s, "./") {
					s = strings.TrimLeft(s, "./")
				}
				fileData, err := Files.ReadFile(consts.EMBED_FILES_PREFIX + s)
				if err != nil {
					ext := filepath.Ext(s)
					music := rl.LoadMusicStreamFromMemory(ext, fileData, int32(len(fileData)))
					return NewGoObj(music)
				}
			}
			music := rl.LoadMusicStream(fname)
			return NewGoObj(music)
		},
		HelpStr: helpStrArgs{
			explanation: "`load_music` loads the music stream from the given path and returns the music object resource reference",
			signature:   "load_music(path: str) -> GoObj[rl.Music]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "load_music() => GoObj[rl.Music]",
		}.String(),
	},
	"_update_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("update_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("update_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("update_music", 1, "rl.Music", args[0])
			}
			rl.UpdateMusicStream(music.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`update_music` updates the buffer for music streaming from the given music object",
			signature:   "update_music(music: GoObj[rl.Music]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "update_music(music) => null",
		}.String(),
	},
	"_play_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("play_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("play_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("play_music", 1, "rl.Music", args[0])
			}
			rl.PlayMusicStream(music.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`play_music` starts playing the music from the given music object",
			signature:   "play_music(music: GoObj[rl.Music]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "play_music(music) => null",
		}.String(),
	},
	"_stop_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("stop_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("stop_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("stop_music", 1, "rl.Music", args[0])
			}
			rl.StopMusicStream(music.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`stop_music` stops playing the music from the given music object",
			signature:   "stop_music(music: GoObj[rl.Music]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "stop_music(music) => null",
		}.String(),
	},
	"_resume_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("resume_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("resume_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("resume_music", 1, "rl.Music", args[0])
			}
			rl.ResumeMusicStream(music.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`resume_music` resumes playing the paused music from the given music object",
			signature:   "resume_music(music: GoObj[rl.Music]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "resume_music(music) => null",
		}.String(),
	},
	"_pause_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("pause_music", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("pause_music", 1, object.GO_OBJ, args[0].Type())
			}
			music, ok := args[0].(*object.GoObj[rl.Music])
			if !ok {
				return newPositionalTypeErrorForGoObj("pause_music", 1, "rl.Music", args[0])
			}
			rl.PauseMusicStream(music.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`pause_music` pauses the music from the given music object",
			signature:   "pause_music(music: GoObj[rl.Music]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "pause_music(music) => null",
		}.String(),
	},
	"_load_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("load_sound", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("load_sound", 1, object.STRING_OBJ, args[0].Type())
			}
			fname := args[0].(*object.Stringo).Value
			if IsEmbed {
				s := fname
				if strings.HasPrefix(s, "./") {
					s = strings.TrimLeft(s, "./")
				}
				fileData, err := Files.ReadFile(consts.EMBED_FILES_PREFIX + s)
				if err != nil {
					ext := filepath.Ext(s)
					wav := rl.LoadWaveFromMemory(ext, fileData, int32(len(fileData)))
					sound := rl.LoadSoundFromWave(wav)
					return NewGoObj(sound)
				}
			}
			sound := rl.LoadSound(fname)
			return NewGoObj(sound)
		},
		HelpStr: helpStrArgs{
			explanation: "`load_sound` loads the sound stream from the given path and returns the sound object resource reference",
			signature:   "load_sound(path: str) -> GoObj[rl.Sound]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "load_sound() => GoObj[rl.Sound]",
		}.String(),
	},
	"_play_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("play_sound", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("play_sound", 1, object.GO_OBJ, args[0].Type())
			}
			sound, ok := args[0].(*object.GoObj[rl.Sound])
			if !ok {
				return newPositionalTypeErrorForGoObj("play_sound", 1, "rl.Sound", args[0])
			}
			rl.PlaySound(sound.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`play_sound` starts playing the sound from the given sound object",
			signature:   "play_sound(sound: GoObj[rl.Sound]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "play_sound(sound) => null",
		}.String(),
	},
	"_stop_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("stop_sound", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("stop_sound", 1, object.GO_OBJ, args[0].Type())
			}
			sound, ok := args[0].(*object.GoObj[rl.Sound])
			if !ok {
				return newPositionalTypeErrorForGoObj("stop_sound", 1, "rl.Sound", args[0])
			}
			rl.StopSound(sound.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`stop_sound` stops playing the sound from the given sound object",
			signature:   "stop_sound(sound: GoObj[rl.Sound]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "stop_sound(sound) => null",
		}.String(),
	},
	"_resume_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("resume_sound", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("resume_sound", 1, object.GO_OBJ, args[0].Type())
			}
			sound, ok := args[0].(*object.GoObj[rl.Sound])
			if !ok {
				return newPositionalTypeErrorForGoObj("resume_sound", 1, "rl.Sound", args[0])
			}
			rl.ResumeSound(sound.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`resume_sound` resumes playing the paused sound from the given sound object",
			signature:   "resume_sound(sound: GoObj[rl.Sound]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "resume_sound(sound) => null",
		}.String(),
	},
	"_pause_sound": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("pause_sound", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("pause_sound", 1, object.GO_OBJ, args[0].Type())
			}
			sound, ok := args[0].(*object.GoObj[rl.Sound])
			if !ok {
				return newPositionalTypeErrorForGoObj("pause_sound", 1, "rl.Sound", args[0])
			}
			rl.PauseSound(sound.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`pause_sound` pauses the sound from the given sound object",
			signature:   "pause_sound(sound: GoObj[rl.Sound]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "pause_sound(sound) => null",
		}.String(),
	},
	"_unload": {
		Fun: func(args ...object.Object) object.Object {
			for i, arg := range args {
				// If the arg is a list go through the list and check every arg to remove
				if arg.Type() == object.LIST_OBJ {
					l := arg.(*object.List).Elements
					for _, e := range l {
						maybeErr := unloadFromRaylib(e, i)
						if isError(maybeErr) {
							return maybeErr
						}
					}
				} else {
					maybeErr := unloadFromRaylib(arg, i)
					if isError(maybeErr) {
						return maybeErr
					}
				}
			}
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`unload` unloads the given objects from the gg object",
			signature:   "unload(args...: GoObject[rl.Texture2D|rl.Music|rl.Sound]|list[GoObject[rl.Texture2D|rl.Music|rl.Sound]]) -> null",
			errors:      "CustomError",
			example:     "unload() => null",
		}.String(),
	},
})

func unloadFromRaylib(arg object.Object, pos int) object.Object {
	if arg.Type() != object.GO_OBJ {
		return newPositionalTypeError("unload", pos, object.GO_OBJ, arg.Type())
	}
	if tex, ok := arg.(*object.GoObj[rl.Texture2D]); ok {
		rl.UnloadTexture(tex.Value)
		return object.NULL
	} else if music, ok := arg.(*object.GoObj[rl.Music]); ok {
		rl.UnloadMusicStream(music.Value)
		return object.NULL
	} else if sound, ok := arg.(*object.GoObj[rl.Sound]); ok {
		rl.UnloadSound(sound.Value)
		return object.NULL
	}
	return newError("`unload` error: Failed to find gg object to unload, expected any GO_OBJ of [rl.Texture2D, rl.Music, rl.Sound]")
}

var _ui_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_new_app": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("new_app", len(args), 0, "")
			}
			app := app.New()
			return NewGoObj(app)
		},
		HelpStr: helpStrArgs{
			explanation: "`new_app` returns the base ui app object to be used for all other ui functions",
			signature:   "new_app() -> GoObj[fyne.App]",
			errors:      "InvalidArgCount",
			example:     "new_app() => GoObj[fyne.App]",
		}.String(),
	},
	"_window": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 5 {
				return newInvalidArgCountError("window", len(args), 5, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("window", 1, object.GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*object.GoObj[fyne.App])
			if !ok {
				return newPositionalTypeErrorForGoObj("window", 1, "fyne.App", args[0])
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("window", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("window", 3, object.INTEGER_OBJ, args[2].Type())
			}
			if args[3].Type() != object.STRING_OBJ {
				return newPositionalTypeError("window", 4, object.STRING_OBJ, args[3].Type())
			}
			if args[4].Type() != object.GO_OBJ {
				return newPositionalTypeError("window", 5, object.GO_OBJ, args[4].Type())
			}
			content, ok := args[4].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("window", 4, "fyne.CanvasObject", args[4])
			}
			width := args[1].(*object.Integer).Value
			height := args[2].(*object.Integer).Value
			title := args[3].(*object.Stringo).Value
			w := app.Value.NewWindow(title)
			w.Resize(fyne.Size{Width: float32(width), Height: float32(height)})
			w.SetContent(content.Value)
			w.ShowAndRun()
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`window` runs the window function on the given app to display the ui with the given content",
			signature:   "window(app: GoObj[fyne.App], width: int=400, height: int=400, title: str='blue ui window', content: GoObj[fyne.CanvasObject]=null) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "window(app) => null (side effect, shows ui window)",
		}.String(),
	},
	"_label": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("label", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("label", 1, object.STRING_OBJ, args[0].Type())
			}
			label := args[0].(*object.Stringo).Value
			l := widget.NewLabel(label)
			return NewGoObj[fyne.CanvasObject](l)
		},
		HelpStr: helpStrArgs{
			explanation: "`label` returns the label ui widget with the given STRING as the label",
			signature:   "label(title: str) -> GoObj[fyne.CanvasObject](Value: *widget.Label)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "label('Hello World') => GoObj[fyne.CanvasObject](Value: *widget.Label)",
		}.String(),
	},
	"_progress_bar": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("progress_bar", len(args), 1, "")
			}
			if args[0].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("progress_bar", 1, object.BOOLEAN_OBJ, args[0].Type())
			}
			isInfinite := args[0].(*object.Boolean).Value
			if isInfinite {
				return NewGoObj[fyne.CanvasObject](widget.NewProgressBarInfinite())
			}
			return NewGoObj[fyne.CanvasObject](widget.NewProgressBar())
		},
		HelpStr: helpStrArgs{
			explanation: "`progress_bar` returns the progress_bar ui widget with sets it to infinite if is_infinite is true",
			signature:   "progress_bar(is_infinite: bool=false) -> GoObj[fyne.CanvasObject](Value: *widget.ProgressBar)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "progress_bar() => GoObj[fyne.CanvasObject](Value: *widget.ProgressBar|Infinite)",
		}.String(),
	},
	"_progress_bar_set_value": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("progress_bar_set_value", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("progress_bar_set_value", 1, object.GO_OBJ, args[0].Type())
			}
			if args[1].Type() != object.FLOAT_OBJ {
				return newPositionalTypeError("progress_bar_set_value", 2, object.FLOAT_OBJ, args[1].Type())
			}
			value := args[1].(*object.Float).Value
			progressBar, ok := args[0].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("progress_bar_set_value", 1, "fyne.CanvasObject", args[0])
			}
			switch x := progressBar.Value.(type) {
			case *widget.ProgressBar:
				x.SetValue(value)
				return object.NULL
			default:
				return newError("`progress_bar_set_value` error: type mismatch. got=%T", x)
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`progress_bar_set_value` sets the float value of a progress bar widget",
			signature:   "progress_bar_set_value(pb: GoObj[fyne.CanvasObject](Value: *widget.ProgressBar), value: float) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "progress_bar_set_value(pb, 1.0) => null (side effect, refresh ui with updated progress bar)",
		}.String(),
	},
	"_toolbar": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) == 0 {
				return newInvalidArgCountError("toolbar", len(args), 1, "or more")
			}
			tis := []widget.ToolbarItem{}
			for i, arg := range args {
				if arg.Type() != object.GO_OBJ {
					return newPositionalTypeError("toolbar", i+1, object.GO_OBJ, arg.Type())
				}
				ti, ok := args[0].(*object.GoObj[widget.ToolbarItem])
				if !ok {
					return newPositionalTypeErrorForGoObj("toolbar", 1, "widget.ToolbarItem", args[0])
				}
				tis = append(tis, ti.Value)
			}
			return NewGoObj[fyne.CanvasObject](widget.NewToolbar(tis...))
		},
		HelpStr: helpStrArgs{
			explanation: "`toolbar.new()`: `toolbar` accepts a variable amount of widget.ToolbarItems to create a ui toolbar widget",
			signature:   "toolbar(args...: GoObj[widget.ToolbarItem]) -> GoObj[fyne.CanvasObject](Value: *widget.ToolBar)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "toolbar() => GoObj[fyne.CanvasObject](Value: *widget.ToolBar)",
		}.String(),
	},
	"_toolbar_spacer": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("toolbar_spacer", len(args), 0, "")
			}
			return NewGoObj[widget.ToolbarItem](widget.NewToolbarSpacer())
		},
		HelpStr: helpStrArgs{
			explanation: "`toolbar.spacer()`: `toolbar_spacer` returns a toolbar spacer widget",
			signature:   "toolbar_spacer() -> GoObj[widget.ToolbarItem](Value: *widget.ToolbarSpacer)",
			errors:      "InvalidArgCount",
			example:     "toolbar_spacer() => GoObj[widget.ToolbarItem](Value: *widget.ToolBarSpacer)",
		}.String(),
	},
	"_toolbar_separator": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("toolbar_separator", len(args), 0, "")
			}
			return NewGoObj[widget.ToolbarItem](widget.NewToolbarSeparator())
		},
		HelpStr: helpStrArgs{
			explanation: "`toolbar.separator()`: `toolbar_separator` returns a toolbar separator widget",
			signature:   "toolbar_separator() -> GoObj[widget.ToolbarItem](Value: *widget.ToolbarSeparator)",
			errors:      "InvalidArgCount",
			example:     "toolbar_separator() => GoObj[widget.ToolbarItem](Value: *widget.ToolbarSeparator)",
		}.String(),
	},
	"_row": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("row", len(args), 1, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("row", 1, object.LIST_OBJ, args[0].Type())
			}
			elements := args[0].(*object.List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != object.GO_OBJ {
					return newError("`row` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				o, ok := e.(*object.GoObj[fyne.CanvasObject])
				if !ok {
					return newPositionalTypeErrorForGoObj("row(children)", i+1, "fyne.CanvasObject", e)
				}
				canvasObjects[i] = o.Value
			}
			vbox := container.NewVBox(canvasObjects...)
			return NewGoObj[fyne.CanvasObject](vbox)
		},
		HelpStr: helpStrArgs{
			explanation: "`row` returns a ui object to align items given to it vertically",
			signature:   "row(elements: list[GoObject[fyne.CanvasObject]]=[]) -> GoObj[fyne.CanvasObject](Value: *fyne.Container)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "row(elems) => GoObj[fyne.CanvasObject](Value: *fyne.Container)",
		}.String(),
	},
	"_col": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("col", len(args), 1, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("col", 1, object.LIST_OBJ, args[0].Type())
			}
			elements := args[0].(*object.List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != object.GO_OBJ {
					return newError("`col` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				o, ok := e.(*object.GoObj[fyne.CanvasObject])
				if !ok {
					return newPositionalTypeErrorForGoObj("col", i+1, "fyne.CanvasObject", e)
				}
				canvasObjects[i] = o.Value
			}
			hbox := container.NewHBox(canvasObjects...)
			return NewGoObj[fyne.CanvasObject](hbox)
		},
		HelpStr: helpStrArgs{
			explanation: "`col` returns a ui object to align items given to it horizontally",
			signature:   "col(elements: list[GoObject[fyne.CanvasObject]]=[]) -> GoObj[fyne.CanvasObject](Value: *fyne.Container)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "col(elems) => GoObj[fyne.CanvasObject](Value: *fyne.Container)",
		}.String(),
	},
	"_grid": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("grid", len(args), 2, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("grid", 1, object.INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("grid", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.LIST_OBJ {
				return newPositionalTypeError("grid", 3, object.LIST_OBJ, args[2].Type())
			}
			rowsOrCols := int(args[0].(*object.Integer).Value)
			gridType := args[1].(*object.Stringo).Value
			if gridType != "COLS" && gridType != "ROWS" {
				return newError("`grid` error: type must be COLS or ROWS. got=%s", gridType)
			}
			elements := args[2].(*object.List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != object.GO_OBJ {
					return newError("`grid` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				o, ok := e.(*object.GoObj[fyne.CanvasObject])
				if !ok {
					return newPositionalTypeErrorForGoObj("grid", i+1, "fyne.CanvasObject", e)
				}
				canvasObjects[i] = o.Value
			}
			var grid *fyne.Container
			if gridType == "ROWS" {
				grid = container.NewGridWithRows(rowsOrCols, canvasObjects...)
			} else {
				grid = container.NewGridWithColumns(rowsOrCols, canvasObjects...)
			}
			return NewGoObj[fyne.CanvasObject](grid)
		},
		HelpStr: helpStrArgs{
			explanation: "`grid` returns a ui object to align items given to it in a grid based on the number of rowcols",
			signature:   "grid(rowcols: int, t: str('ROWS'|'COLS'), children: list[GoObject[fyne.CanvasObject]]=[]) -> GoObj[fyne.CanvasObject](Value: *fyne.Container)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "grid(elems) => GoObj[fyne.CanvasObject](Value: *fyne.Container)",
		}.String(),
	},
	"_entry": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("entry", len(args), 1, "")
			}
			if args[0].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("entry", 1, object.BOOLEAN_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("entry", 2, object.STRING_OBJ, args[1].Type())
			}
			isMultiline := args[0].(*object.Boolean).Value
			placeholderText := args[1].(*object.Stringo).Value
			var entry *widget.Entry
			if isMultiline {
				entry = widget.NewMultiLineEntry()
			} else {
				entry = widget.NewEntry()
			}
			entry.SetPlaceHolder(placeholderText)
			return NewGoObj[fyne.CanvasObject](entry)
		},
		HelpStr: helpStrArgs{
			explanation: "`entry` returns a ui entry widget object with placeholder text if given and its multiline if is_multiline is true",
			signature:   "entry(is_multiline: bool=false, placeholder: str='') -> GoObj[fyne.CanvasObject](Value: *widget.Entry)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "entry() => GoObj[fyne.CanvasObject](Value: *widget.Entry)",
		}.String(),
	},
	"_entry_get_text": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("entry_get_text", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("entry_get_text", 1, object.GO_OBJ, args[0].Type())
			}
			entry, ok := args[0].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("entry_get_text", 1, "fyne.CanvasObject", args[0])
			}
			switch x := entry.Value.(type) {
			case *widget.Entry:
				return &object.Stringo{Value: x.Text}
			default:
				return newError("`entry_get_text` error: entry id did not match entry. got=%T", x)
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`entry_get_text` returns the text that is currently present in the entry ui widget object",
			signature:   "entry_get_text(e: GoObj[fyne.CanvasObject](Value: *widget.Entry)) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "entry_get_text(e) => 'test'",
		}.String(),
	},
	"_entry_set_text": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("entry_set_text", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("entry_set_text", 1, object.GO_OBJ, args[0].Type())
			}
			entry, ok := args[0].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("entry_set_text", 1, "fyne.CanvasObject", args[0])
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("entry_set_text", 2, object.STRING_OBJ, args[1].Type())
			}
			value := args[1].(*object.Stringo).Value
			switch x := entry.Value.(type) {
			case *widget.Entry:
				x.SetText(value)
				return object.NULL
			default:
				return newError("`entry_set_text` error: entry id did not match entry. got=%T", x)
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`entry_set_text` sets the text of the entry ui widget object with the given string",
			signature:   "entry_set_text(e: GoObj[fyne.CanvasObject](Value: *widget.Entry), v: str) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "entry_set_text(e, 'test') => null (side effect, refresh ui with updated entry)",
		}.String(),
	},
	"_append_form": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("append_form", len(args), 3, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("append_form", 1, object.GO_OBJ, args[0].Type())
			}
			maybeForm, ok := args[0].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("append_form", 1, "fyne.CanvasObject", args[0])
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("append_form", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.GO_OBJ {
				return newPositionalTypeError("append_form", 3, object.GO_OBJ, args[2].Type())
			}
			w, ok := args[2].(*object.GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("append_form", 3, "fyne.CanvasObject", args[2])
			}
			var form *widget.Form
			switch x := maybeForm.Value.(type) {
			case *widget.Form:
				form = x
			default:
				return newError("`append_form` error: id used for form is not form. got=%T", x)
			}
			form.Append(args[1].(*object.Stringo).Value, w.Value)
			return object.NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`append_form` appends a label with the given string and a corresponding widget to the given form",
			signature:   "append_form(f: GoObj[fyne.CanvasObject](Value: *widget.Form), title: str, widget: GoObj[fyne.CanvasObject]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "append_form(f, 'test', w) => null (side effect, refresh ui form with updated label/widget)",
		}.String(),
	},
	"_icon_account": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_account", len(args), 0, "")
			}
			return NewGoObj(theme.AccountIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_account` returns the object of the icon_account resource",
			signature:   "icon_account() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_account() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_cancel": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_cancel", len(args), 0, "")
			}
			return NewGoObj(theme.CancelIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_cancel` returns the object of the icon_cancel resource",
			signature:   "icon_cancel() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_cancel() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_check_button_checked": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_check_button_checked", len(args), 0, "")
			}
			return NewGoObj(theme.CheckButtonCheckedIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_check_button_checked` returns the object of the icon_check_button_checked resource",
			signature:   "icon_check_button_checked() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_check_button_checked() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_check_button": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_check_button", len(args), 0, "")
			}
			return NewGoObj(theme.CheckButtonIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_check_button` returns the object of the icon_check_button resource",
			signature:   "icon_check_button() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_check_button() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_color_achromatic": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_color_achromatic", len(args), 0, "")
			}
			return NewGoObj(theme.ColorAchromaticIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_color_achromatic` returns the object of the icon_color_achromatic resource",
			signature:   "icon_color_achromatic() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_color_achromatic() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_color_chromatic": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_color_chromatic", len(args), 0, "")
			}
			return NewGoObj(theme.ColorChromaticIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_color_chromatic` returns the object of the icon_color_chromatic resource",
			signature:   "icon_color_chromatic() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_color_chromatic() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_color_palette": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_color_palette", len(args), 0, "")
			}
			return NewGoObj(theme.ColorPaletteIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_color_palette` returns the object of the icon_color_palette resource",
			signature:   "icon_color_palette() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_color_palette() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_computer": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_computer", len(args), 0, "")
			}
			return NewGoObj(theme.ComputerIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_computer` returns the object of the icon_computer resource",
			signature:   "icon_computer() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_computer() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_confirm": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_confirm", len(args), 0, "")
			}
			return NewGoObj(theme.ConfirmIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_confirm` returns the object of the icon_confirm resource",
			signature:   "icon_confirm() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_confirm() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_content_add": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_content_add", len(args), 0, "")
			}
			return NewGoObj(theme.ContentAddIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_content_add` returns the object of the icon_content_add resource",
			signature:   "icon_content_add() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_content_add() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_content_clear": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_content_clear", len(args), 0, "")
			}
			return NewGoObj(theme.ContentClearIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_content_clear` returns the object of the icon_content_clear resource",
			signature:   "icon_content_clear() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_content_clear() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_content_copy": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_content_copy", len(args), 0, "")
			}
			return NewGoObj(theme.ContentCopyIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_content_copy` returns the object of the icon_content_copy resource",
			signature:   "icon_content_copy() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_content_copy() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_content_cut": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_content_cut", len(args), 0, "")
			}
			return NewGoObj(theme.ContentCutIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_content_cut` returns the object of the icon_content_cut resource",
			signature:   "icon_content_cut() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_content_cut() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_content_paste": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_content_paste", len(args), 0, "")
			}
			return NewGoObj(theme.ContentPasteIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_content_paste` returns the object of the icon_content_paste resource",
			signature:   "icon_content_paste() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_content_paste() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_content_redo": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_content_redo", len(args), 0, "")
			}
			return NewGoObj(theme.ContentRedoIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_content_redo` returns the object of the icon_content_redo resource",
			signature:   "icon_content_redo() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_content_redo() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_content_remove": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_content_remove", len(args), 0, "")
			}
			return NewGoObj(theme.ContentRemoveIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_content_remove` returns the object of the icon_content_remove resource",
			signature:   "icon_content_remove() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_content_remove() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_content_undo": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_content_undo", len(args), 0, "")
			}
			return NewGoObj(theme.ContentUndoIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_content_undo` returns the object of the icon_content_undo resource",
			signature:   "icon_content_undo() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_content_undo() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_delete": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_delete", len(args), 0, "")
			}
			return NewGoObj(theme.DeleteIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_delete` returns the object of the icon_delete resource",
			signature:   "icon_delete() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_delete() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_document_create": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_document_create", len(args), 0, "")
			}
			return NewGoObj(theme.DocumentCreateIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_document_create` returns the object of the icon_document_create resource",
			signature:   "icon_document_create() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_document_create() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_document": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_document", len(args), 0, "")
			}
			return NewGoObj(theme.DocumentIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_document` returns the object of the icon_document resource",
			signature:   "icon_document() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_document() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_document_print": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_document_print", len(args), 0, "")
			}
			return NewGoObj(theme.DocumentPrintIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_document_print` returns the object of the icon_document_print resource",
			signature:   "icon_document_print() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_document_print() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_document_save": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_document_save", len(args), 0, "")
			}
			return NewGoObj(theme.DocumentSaveIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_document_save` returns the object of the icon_document_save resource",
			signature:   "icon_document_save() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_document_save() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_download": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_download", len(args), 0, "")
			}
			return NewGoObj(theme.DownloadIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_download` returns the object of the icon_download resource",
			signature:   "icon_download() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_download() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_error": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_error", len(args), 0, "")
			}
			return NewGoObj(theme.ErrorIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_error` returns the object of the icon_error resource",
			signature:   "icon_error() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_error() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_file_application": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_file_application", len(args), 0, "")
			}
			return NewGoObj(theme.FileApplicationIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_file_application` returns the object of the icon_file_application resource",
			signature:   "icon_file_application() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_file_application() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_file_audio": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_file_audio", len(args), 0, "")
			}
			return NewGoObj(theme.FileAudioIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_file_audio` returns the object of the icon_file_audio resource",
			signature:   "icon_file_audio() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_file_audio() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_file": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_file", len(args), 0, "")
			}
			return NewGoObj(theme.FileIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_file` returns the object of the icon_file resource",
			signature:   "icon_file() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_file() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_file_image": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_file_image", len(args), 0, "")
			}
			return NewGoObj(theme.FileImageIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_file_image` returns the object of the icon_file_image resource",
			signature:   "icon_file_image() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_file_image() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_file_text": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_file_text", len(args), 0, "")
			}
			return NewGoObj(theme.FileTextIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_file_text` returns the object of the icon_file_text resource",
			signature:   "icon_file_text() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_file_text() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_file_video": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_file_video", len(args), 0, "")
			}
			return NewGoObj(theme.FileVideoIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_file_video` returns the object of the icon_file_video resource",
			signature:   "icon_file_video() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_file_video() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_folder": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_folder", len(args), 0, "")
			}
			return NewGoObj(theme.FolderIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_folder` returns the object of the icon_folder resource",
			signature:   "icon_folder() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_folder() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_folder_new": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_folder_new", len(args), 0, "")
			}
			return NewGoObj(theme.FolderNewIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_folder_new` returns the object of the icon_folder_new resource",
			signature:   "icon_folder_new() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_folder_new() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_folder_open": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_folder_open", len(args), 0, "")
			}
			return NewGoObj(theme.FolderOpenIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_folder_open` returns the object of the icon_folder_open resource",
			signature:   "icon_folder_open() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_folder_open() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_grid": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_grid", len(args), 0, "")
			}
			return NewGoObj(theme.GridIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_grid` returns the object of the icon_grid resource",
			signature:   "icon_grid() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_grid() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_help": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_help", len(args), 0, "")
			}
			return NewGoObj(theme.HelpIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_help` returns the object of the icon_help resource",
			signature:   "icon_help() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_help() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_history": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_history", len(args), 0, "")
			}
			return NewGoObj(theme.HistoryIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_history` returns the object of the icon_history resource",
			signature:   "icon_history() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_history() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_home": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_home", len(args), 0, "")
			}
			return NewGoObj(theme.HomeIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_home` returns the object of the icon_home resource",
			signature:   "icon_home() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_home() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_info": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_info", len(args), 0, "")
			}
			return NewGoObj(theme.InfoIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_info` returns the object of the icon_info resource",
			signature:   "icon_info() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_info() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_list": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_list", len(args), 0, "")
			}
			return NewGoObj(theme.ListIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_list` returns the object of the icon_list resource",
			signature:   "icon_list() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_list() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_login": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_login", len(args), 0, "")
			}
			return NewGoObj(theme.LoginIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_login` returns the object of the icon_login resource",
			signature:   "icon_login() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_login() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_logout": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_logout", len(args), 0, "")
			}
			return NewGoObj(theme.LogoutIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_logout` returns the object of the icon_logout resource",
			signature:   "icon_logout() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_logout() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_mail_attachment": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_mail_attachment", len(args), 0, "")
			}
			return NewGoObj(theme.MailAttachmentIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_mail_attachment` returns the object of the icon_mail_attachment resource",
			signature:   "icon_mail_attachment() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_mail_attachment() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_mail_compose": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_mail_compose", len(args), 0, "")
			}
			return NewGoObj(theme.MailComposeIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_mail_compose` returns the object of the icon_mail_compose resource",
			signature:   "icon_mail_compose() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_mail_compose() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_mail_forward": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_mail_forward", len(args), 0, "")
			}
			return NewGoObj(theme.MailForwardIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_mail_forward` returns the object of the icon_mail_forward resource",
			signature:   "icon_mail_forward() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_mail_forward() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_mail_reply_all": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_mail_reply_all", len(args), 0, "")
			}
			return NewGoObj(theme.MailReplyAllIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_mail_reply_all` returns the object of the icon_mail_reply_all resource",
			signature:   "icon_mail_reply_all() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_mail_reply_all() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_mail_reply": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_mail_reply", len(args), 0, "")
			}
			return NewGoObj(theme.MailReplyIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_mail_reply` returns the object of the icon_mail_reply resource",
			signature:   "icon_mail_reply() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_mail_reply() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_mail_send": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_mail_send", len(args), 0, "")
			}
			return NewGoObj(theme.MailSendIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_mail_send` returns the object of the icon_mail_send resource",
			signature:   "icon_mail_send() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_mail_send() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_fast_forward": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_fast_forward", len(args), 0, "")
			}
			return NewGoObj(theme.MediaFastForwardIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_fast_forward` returns the object of the icon_media_fast_forward resource",
			signature:   "icon_media_fast_forward() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_fast_forward() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_fast_rewind": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_fast_rewind", len(args), 0, "")
			}
			return NewGoObj(theme.MediaFastRewindIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_fast_rewind` returns the object of the icon_media_fast_rewind resource",
			signature:   "icon_media_fast_rewind() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_fast_rewind() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_music": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_music", len(args), 0, "")
			}
			return NewGoObj(theme.MediaMusicIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_music` returns the object of the icon_media_music resource",
			signature:   "icon_media_music() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_music() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_pause": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_pause", len(args), 0, "")
			}
			return NewGoObj(theme.MediaPauseIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_pause` returns the object of the icon_media_pause resource",
			signature:   "icon_media_pause() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_pause() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_photo": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_photo", len(args), 0, "")
			}
			return NewGoObj(theme.MediaPhotoIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_photo` returns the object of the icon_media_photo resource",
			signature:   "icon_media_photo() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_photo() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_play": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_play", len(args), 0, "")
			}
			return NewGoObj(theme.MediaPlayIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_play` returns the object of the icon_media_play resource",
			signature:   "icon_media_play() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_play() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_record": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_record", len(args), 0, "")
			}
			return NewGoObj(theme.MediaRecordIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_record` returns the object of the icon_media_record resource",
			signature:   "icon_media_record() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_record() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_replay": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_replay", len(args), 0, "")
			}
			return NewGoObj(theme.MediaReplayIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_replay` returns the object of the icon_media_replay resource",
			signature:   "icon_media_replay() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_replay() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_skip_next": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_skip_next", len(args), 0, "")
			}
			return NewGoObj(theme.MediaSkipNextIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_skip_next` returns the object of the icon_media_skip_next resource",
			signature:   "icon_media_skip_next() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_skip_next() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_skip_previous": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_skip_previous", len(args), 0, "")
			}
			return NewGoObj(theme.MediaSkipPreviousIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_skip_previous` returns the object of the icon_media_skip_previous resource",
			signature:   "icon_media_skip_previous() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_skip_previous() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_stop": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_stop", len(args), 0, "")
			}
			return NewGoObj(theme.MediaStopIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_stop` returns the object of the icon_media_stop resource",
			signature:   "icon_media_stop() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_stop() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_media_video": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_media_video", len(args), 0, "")
			}
			return NewGoObj(theme.MediaVideoIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_media_video` returns the object of the icon_media_video resource",
			signature:   "icon_media_video() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_media_video() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_menu_drop_down": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_menu_drop_down", len(args), 0, "")
			}
			return NewGoObj(theme.MenuDropDownIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_menu_drop_down` returns the object of the icon_menu_drop_down resource",
			signature:   "icon_menu_drop_down() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_menu_drop_down() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_menu_drop_up": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_menu_drop_up", len(args), 0, "")
			}
			return NewGoObj(theme.MenuDropUpIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_menu_drop_up` returns the object of the icon_menu_drop_up resource",
			signature:   "icon_menu_drop_up() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_menu_drop_up() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_menu_expand": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_menu_expand", len(args), 0, "")
			}
			return NewGoObj(theme.MenuExpandIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_menu_expand` returns the object of the icon_menu_expand resource",
			signature:   "icon_menu_expand() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_menu_expand() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_menu": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_menu", len(args), 0, "")
			}
			return NewGoObj(theme.MenuIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_menu` returns the object of the icon_menu resource",
			signature:   "icon_menu() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_menu() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_more_horizontal": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_more_horizontal", len(args), 0, "")
			}
			return NewGoObj(theme.MoreHorizontalIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_more_horizontal` returns the object of the icon_more_horizontal resource",
			signature:   "icon_more_horizontal() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_more_horizontal() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_more_vertical": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_more_vertical", len(args), 0, "")
			}
			return NewGoObj(theme.MoreVerticalIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_more_vertical` returns the object of the icon_more_vertical resource",
			signature:   "icon_more_vertical() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_more_vertical() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_move_down": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_move_down", len(args), 0, "")
			}
			return NewGoObj(theme.MoveDownIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_move_down` returns the object of the icon_move_down resource",
			signature:   "icon_move_down() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_move_down() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_move_up": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_move_up", len(args), 0, "")
			}
			return NewGoObj(theme.MoveUpIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_move_up` returns the object of the icon_move_up resource",
			signature:   "icon_move_up() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_move_up() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_navigate_back": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_navigate_back", len(args), 0, "")
			}
			return NewGoObj(theme.NavigateBackIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_navigate_back` returns the object of the icon_navigate_back resource",
			signature:   "icon_navigate_back() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_navigate_back() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_navigate_next": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_navigate_next", len(args), 0, "")
			}
			return NewGoObj(theme.NavigateNextIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_navigate_next` returns the object of the icon_navigate_next resource",
			signature:   "icon_navigate_next() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_navigate_next() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_question": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_question", len(args), 0, "")
			}
			return NewGoObj(theme.QuestionIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_question` returns the object of the icon_question resource",
			signature:   "icon_question() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_question() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_radio_button_checked": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_radio_button_checked", len(args), 0, "")
			}
			return NewGoObj(theme.RadioButtonCheckedIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_radio_button_checked` returns the object of the icon_radio_button_checked resource",
			signature:   "icon_radio_button_checked() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_radio_button_checked() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_radio_button": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_radio_button", len(args), 0, "")
			}
			return NewGoObj(theme.RadioButtonIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_radio_button` returns the object of the icon_radio_button resource",
			signature:   "icon_radio_button() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_radio_button() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_search": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_search", len(args), 0, "")
			}
			return NewGoObj(theme.SearchIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_search` returns the object of the icon_search resource",
			signature:   "icon_search() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_search() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_search_replace": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_search_replace", len(args), 0, "")
			}
			return NewGoObj(theme.SearchReplaceIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_search_replace` returns the object of the icon_search_replace resource",
			signature:   "icon_search_replace() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_search_replace() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_settings": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_settings", len(args), 0, "")
			}
			return NewGoObj(theme.SettingsIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_settings` returns the object of the icon_settings resource",
			signature:   "icon_settings() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_settings() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_storage": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_storage", len(args), 0, "")
			}
			return NewGoObj(theme.StorageIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_storage` returns the object of the icon_storage resource",
			signature:   "icon_storage() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_storage() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_upload": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_upload", len(args), 0, "")
			}
			return NewGoObj(theme.UploadIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_upload` returns the object of the icon_upload resource",
			signature:   "icon_upload() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_upload() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_view_full_screen": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_view_full_screen", len(args), 0, "")
			}
			return NewGoObj(theme.ViewFullScreenIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_view_full_screen` returns the object of the icon_view_full_screen resource",
			signature:   "icon_view_full_screen() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_view_full_screen() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_view_refresh": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_view_refresh", len(args), 0, "")
			}
			return NewGoObj(theme.ViewRefreshIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_view_refresh` returns the object of the icon_view_refresh resource",
			signature:   "icon_view_refresh() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_view_refresh() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_view_restore": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_view_restore", len(args), 0, "")
			}
			return NewGoObj(theme.ViewRestoreIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_view_restore` returns the object of the icon_view_restore resource",
			signature:   "icon_view_restore() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_view_restore() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_visibility": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_visibility", len(args), 0, "")
			}
			return NewGoObj(theme.VisibilityIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_visibility` returns the object of the icon_visibility resource",
			signature:   "icon_visibility() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_visibility() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_visibility_off": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_visibility_off", len(args), 0, "")
			}
			return NewGoObj(theme.VisibilityOffIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_visibility_off` returns the object of the icon_visibility_off resource",
			signature:   "icon_visibility_off() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_visibility_off() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_volume_down": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_volume_down", len(args), 0, "")
			}
			return NewGoObj(theme.VolumeDownIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_volume_down` returns the object of the icon_volume_down resource",
			signature:   "icon_volume_down() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_volume_down() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_volume_mute": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_volume_mute", len(args), 0, "")
			}
			return NewGoObj(theme.VolumeMuteIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_volume_mute` returns the object of the icon_volume_mute resource",
			signature:   "icon_volume_mute() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_volume_mute() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_volume_up": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_volume_up", len(args), 0, "")
			}
			return NewGoObj(theme.VolumeUpIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_volume_up` returns the object of the icon_volume_up resource",
			signature:   "icon_volume_up() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_volume_up() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_warning": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_warning", len(args), 0, "")
			}
			return NewGoObj(theme.WarningIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_warning` returns the object of the icon_warning resource",
			signature:   "icon_warning() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_warning() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_zoom_fit": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_zoom_fit", len(args), 0, "")
			}
			return NewGoObj(theme.ZoomFitIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_zoom_fit` returns the object of the icon_zoom_fit resource",
			signature:   "icon_zoom_fit() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_zoom_fit() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_zoom_in": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_zoom_in", len(args), 0, "")
			}
			return NewGoObj(theme.ZoomInIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_zoom_in` returns the object of the icon_zoom_in resource",
			signature:   "icon_zoom_in() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_zoom_in() -> GoObj[fyne.Resouce]",
		}.String(),
	},
	"_icon_zoom_out": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("icon_zoom_out", len(args), 0, "")
			}
			return NewGoObj(theme.ZoomOutIcon())
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_zoom_out` returns the object of the icon_zoom_out resource",
			signature:   "icon_zoom_out() -> GoObj[fyne.Resource]",
			errors:      "InvalidArgCount",
			example:     "icon_zoom_out() -> GoObj[fyne.Resouce]",
		}.String(),
	},
})
