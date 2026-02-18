package object

import (
	"blue/consts"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var GgBuiltins = NewBuiltinSliceType{
	{
		Name: "_init_window",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 3 {
					return newInvalidArgCountError("init_window", len(args), 3, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("init_window", 1, INTEGER_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("init_window", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != STRING_OBJ {
					return newPositionalTypeError("init_window", 3, STRING_OBJ, args[2].Type())
				}
				width := int32(args[0].(*Integer).Value)
				height := int32(args[1].(*Integer).Value)
				title := args[2].(*Stringo).Value
				rl.InitWindow(width, height, title)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`init_window` initalizes the gg graphics window with the given width, height, and title",
				signature:   "init_window(width: int=800, height: int=600, title: str='gg - example app') -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "init_window() => null",
			}.String(),
		},
	},
	{
		Name: "_close_window",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("close_window", len(args), 0, "")
				}
				rl.CloseWindow()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`close_window` closes the gg graphics window",
				signature:   "close_window() -> null",
				errors:      "InvalidArgCount",
				example:     "close_window() => null",
			}.String(),
		},
	},
	{
		Name: "_window_should_close",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
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
	},
	{
		Name: "_get_screen_width",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("get_screen_width", len(args), 0, "")
				}
				return &Integer{Value: int64(rl.GetScreenWidth())}
			},
			HelpStr: helpStrArgs{
				explanation: "`get_screen_width` gets the screen width as an int",
				signature:   "get_screen_width() -> int",
				errors:      "InvalidArgCount",
				example:     "get_screen_width() => 800",
			}.String(),
		},
	},
	{
		Name: "_get_screen_height",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("get_screen_height", len(args), 0, "")
				}
				return &Integer{Value: int64(rl.GetScreenHeight())}
			},
			HelpStr: helpStrArgs{
				explanation: "`get_screen_height` gets the screen height as an int",
				signature:   "get_screen_height() -> int",
				errors:      "InvalidArgCount",
				example:     "get_screen_height() => 800",
			}.String(),
		},
	},
	{
		Name: "_begin_drawing",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("begin_drawing", len(args), 0, "")
				}
				rl.BeginDrawing()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`begin_drawing` sets up the drawing canvas to start drawing",
				signature:   "begin_drawing() -> null",
				errors:      "InvalidArgCount",
				example:     "begin_drawing() => null",
			}.String(),
		},
	},
	{
		Name: "_end_drawing",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("end_drawing", len(args), 0, "")
				}
				rl.EndDrawing()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`end_drawing` ends canvas drawing and swaps buffers (double buffering)",
				signature:   "end_drawing() -> null",
				errors:      "InvalidArgCount",
				example:     "end_drawing() => null",
			}.String(),
		},
	},
	{
		Name: "_clear_background",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("clear_background", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("clear_background", 1, GO_OBJ, args[0].Type())
				}
				goObj, ok := args[0].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("clear_background", 1, "rl.Color", args[0])
				}
				rl.ClearBackground(goObj.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`clear_background` sets the background color to the given color",
				signature:   "clear_background(color: GoObj[rl.Color]=color.white) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "clear_background() => null",
			}.String(),
		},
	},
	{
		Name: "_color_map",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("color_map", len(args), 0, "")
				}
				mapObj := NewOrderedMap[string, Object]()
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
				newColor := &Builtin{
					Fun: func(args ...Object) Object {
						if len(args) != 4 {
							return newInvalidArgCountError("new_color", len(args), 4, "")
						}
						if args[0].Type() != INTEGER_OBJ {
							return newPositionalTypeError("new_color", 1, INTEGER_OBJ, args[0].Type())
						}
						if args[1].Type() != INTEGER_OBJ {
							return newPositionalTypeError("new_color", 2, INTEGER_OBJ, args[1].Type())
						}
						if args[2].Type() != INTEGER_OBJ {
							return newPositionalTypeError("new_color", 3, INTEGER_OBJ, args[2].Type())
						}
						if args[3].Type() != INTEGER_OBJ {
							return newPositionalTypeError("new_color", 4, INTEGER_OBJ, args[3].Type())
						}
						return NewGoObj(rl.NewColor(
							uint8(args[0].(*Integer).Value),
							uint8(args[1].(*Integer).Value),
							uint8(args[2].(*Integer).Value),
							uint8(args[3].(*Integer).Value)))
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
				return CreateMapObjectForGoMap(*mapObj)
			},
			HelpStr: helpStrArgs{
				explanation: "`color_map` returns a map with all the colors available as well as a function 'new' to generate a color from an rgba value",
				signature:   "color_map() -> map[str:GoObj[rl.Color]|fun(r,g,b,a)->GoObj[rl.Color]]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "color_map() => map[str:GoObj[rl.Color]|fun(r,g,b,a)->GoObj[rl.Color]]",
			}.String(),
		},
	},
	{
		Name: "_draw_text",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 5 {
					return newInvalidArgCountError("draw_text", len(args), 5, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("draw_text", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_text", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_text", 3, INTEGER_OBJ, args[2].Type())
				}
				if args[3].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_text", 4, INTEGER_OBJ, args[3].Type())
				}
				if args[4].Type() != GO_OBJ {
					return newPositionalTypeError("draw_text", 5, GO_OBJ, args[4].Type())
				}
				goObj, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_text", 5, "rl.Color", args[4])
				}
				text := args[0].(*Stringo).Value
				posX := int32(args[1].(*Integer).Value)
				posY := int32(args[2].(*Integer).Value)
				fontSize := int32(args[3].(*Integer).Value)
				rl.DrawText(text, posX, posY, fontSize, goObj.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`draw_text` draws text on the canvas with the given text at (x,y) with font_size, and color",
				signature:   "draw_text(text: str, pos_x: int=0, pos_y: int=0, font_size: int=20, text_color: GO_OBJ[rl.Color]=color.black) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "draw_text('Hello World!') => null",
			}.String(),
		},
	},
	{
		Name: "_draw_texture",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 4 {
					return newInvalidArgCountError("draw_texture", len(args), 4, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("draw_texture", 1, GO_OBJ, args[0].Type())
				}
				tex, ok := args[0].(*GoObj[rl.Texture2D])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_texture", 1, "rl.Texture2D", args[0])
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_texture", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_texture", 3, INTEGER_OBJ, args[2].Type())
				}
				if args[3].Type() != GO_OBJ {
					return newPositionalTypeError("draw_texture", 4, GO_OBJ, args[3].Type())
				}
				tint, ok := args[3].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_texture", 4, "rl.Color", args[3])
				}
				posX := int32(args[1].(*Integer).Value)
				posY := int32(args[2].(*Integer).Value)
				rl.DrawTexture(tex.Value, posX, posY, tint.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`draw_texture` draws the 2d texture on the canvas at (x,y) with given tint tint",
				signature:   "draw_texture(texture: GO_OBJ[rl.Texture2D], pos_x: int=0, pos_y: int=0, tint: GO_OBJ[rl.Color]=color.white) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "draw_texture(texture) => null",
			}.String(),
		},
	},
	{
		Name: "_draw_texture_pro",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 6 {
					return newInvalidArgCountError("draw_texture_pro", len(args), 6, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("draw_texture_pro", 1, GO_OBJ, args[0].Type())
				}
				tex, ok := args[0].(*GoObj[rl.Texture2D])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_texture_pro", 1, "rl.Texture2D", args[0])
				}
				if args[1].Type() != GO_OBJ {
					return newPositionalTypeError("draw_texture_pro", 2, GO_OBJ, args[1].Type())
				}
				srcRect, ok := args[1].(*GoObj[rl.Rectangle])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_texture_pro", 2, "rl.Rectangle", args[1])
				}
				if args[2].Type() != GO_OBJ {
					return newPositionalTypeError("draw_texture_pro", 3, GO_OBJ, args[2].Type())
				}
				dstRect, ok := args[2].(*GoObj[rl.Rectangle])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_texture_pro", 3, "rl.Rectangle", args[2])
				}
				if args[3].Type() != GO_OBJ {
					return newPositionalTypeError("draw_texture_pro", 4, GO_OBJ, args[3].Type())
				}
				origin, ok := args[3].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_texture_pro", 4, "rl.Rectangle", args[3])
				}
				if args[4].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_texture_pro", 5, FLOAT_OBJ, args[4].Type())
				}
				rotation := float32(args[4].(*Float).Value)
				if args[5].Type() != GO_OBJ {
					return newPositionalTypeError("draw_texture_pro", 6, GO_OBJ, args[5].Type())
				}
				tint, ok := args[5].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_texture_pro", 6, "rl.Color", args[5])
				}
				rl.DrawTexturePro(tex.Value, srcRect.Value, dstRect.Value, origin.Value, rotation, tint.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`draw_texture_pro` draws a part of the 2d texture on the canvas with the given source_rec, dest_rec, origin, rotation, and tint",
				signature:   "draw_texture_pro(texture: GO_OBJ[rl.Texture2D], source_rec: GO_OBJ[rl.Rectangle]=Rectangle(), dest_rec: GO_OBJ[rl.Rectangle]=Rectangle(), origin: GO_OBJ[rl.Vector2]=Vector2(), rotation: float=0.0, tint=color.white) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "draw_texture_pro(texture) => null",
			}.String(),
		},
	},
	{
		Name: "_draw_rectangle",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 5 {
					return newInvalidArgCountError("draw_rectangle", len(args), 5, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle", 1, INTEGER_OBJ, args[0].Type())
				}
				posx := int32(args[0].(*Integer).Value)
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle", 2, INTEGER_OBJ, args[1].Type())
				}
				posy := int32(args[1].(*Integer).Value)
				if args[2].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle", 3, INTEGER_OBJ, args[2].Type())
				}
				width := int32(args[2].(*Integer).Value)
				if args[3].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle", 4, INTEGER_OBJ, args[3].Type())
				}
				height := int32(args[3].(*Integer).Value)
				if args[4].Type() != GO_OBJ {
					return newPositionalTypeError("draw_rectangle", 5, GO_OBJ, args[4].Type())
				}
				color, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_texture_pro", 4, "rl.Color", args[4])
				}
				rl.DrawRectangle(posx, posy, width, height, color.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`draw_rectangle` draws a rectangle at the given position with width and height",
				signature:   "draw_rectangle(posx: int, posy: int, width: int, height: int, color=color.black) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "draw_rectangle() (used as Rectangle().draw(color))=> null",
			}.String(),
		},
	},
	{
		Name: "_set_target_fps",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("set_target_fps", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("set_target_fps", 1, INTEGER_OBJ, args[0].Type())
				}
				fps := int32(args[0].(*Integer).Value)
				rl.SetTargetFPS(fps)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`set_target_fps` sets the target fps to the given integer",
				signature:   "set_target_fps(fps: int) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "set_target_fps(60) => null",
			}.String(),
		},
	},
	{
		Name: "_set_exit_key",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("set_exit_key", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("set_exit_key", 1, INTEGER_OBJ, args[0].Type())
				}
				key := int32(args[0].(*Integer).Value)
				rl.SetExitKey(key)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`set_exit_key` sets the exit key to the given key (integer)",
				signature:   "set_exit_key(key: int) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "set_exit_key(key.Q) => null",
			}.String(),
		},
	},
	{
		Name: "_is_key_up",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("is_key_up", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("is_key_up", 1, INTEGER_OBJ, args[0].Type())
				}
				key := int32(args[0].(*Integer).Value)
				return nativeToBooleanObject(rl.IsKeyUp(key))
			},
			HelpStr: helpStrArgs{
				explanation: "`is_key_up` returns true if the given key is up",
				signature:   "is_key_up(key: int) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "is_key_up(key.Q) => false",
			}.String(),
		},
	},
	{
		Name: "_is_key_down",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("is_key_down", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("is_key_down", 1, INTEGER_OBJ, args[0].Type())
				}
				key := int32(args[0].(*Integer).Value)
				return nativeToBooleanObject(rl.IsKeyDown(key))
			},
			HelpStr: helpStrArgs{
				explanation: "`is_key_down` returns true if the given key is down",
				signature:   "is_key_down(key: int) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "is_key_down(key.Q) => false",
			}.String(),
		},
	},
	{
		Name: "_is_key_pressed",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("is_key_pressed", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("is_key_pressed", 1, INTEGER_OBJ, args[0].Type())
				}
				key := int32(args[0].(*Integer).Value)
				return nativeToBooleanObject(rl.IsKeyPressed(key))
			},
			HelpStr: helpStrArgs{
				explanation: "`is_key_pressed` returns true if the given key is pressed",
				signature:   "is_key_pressed(key: int) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "is_key_pressed(key.Q) => false",
			}.String(),
		},
	},
	{
		Name: "_is_key_released",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("is_key_released", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("is_key_released", 1, INTEGER_OBJ, args[0].Type())
				}
				key := int32(args[0].(*Integer).Value)
				return nativeToBooleanObject(rl.IsKeyReleased(key))
			},
			HelpStr: helpStrArgs{
				explanation: "`is_key_released` returns true if the given key is released",
				signature:   "is_key_released(key: int) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "is_key_released(key.Q) => false",
			}.String(),
		},
	},
	{
		Name: "_load_texture",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("load_texture", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("load_texture", 1, STRING_OBJ, args[0].Type())
				}
				fname := args[0].(*Stringo).Value
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
	},
	{
		Name: "_rectangle",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 4 {
					return newInvalidArgCountError("Rectangle", len(args), 4, "")
				}
				if args[0].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Rectangle", 1, FLOAT_OBJ, args[0].Type())
				}
				if args[1].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Rectangle", 2, FLOAT_OBJ, args[1].Type())
				}
				if args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Rectangle", 3, FLOAT_OBJ, args[2].Type())
				}
				if args[3].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Rectangle", 4, FLOAT_OBJ, args[3].Type())
				}
				x := float32(args[0].(*Float).Value)
				y := float32(args[1].(*Float).Value)
				width := float32(args[2].(*Float).Value)
				height := float32(args[3].(*Float).Value)
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
	},
	{
		Name: "_rectangle_check_collision",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("rectangle_check_collision", len(args), 4, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("rectangle_check_collision", 1, GO_OBJ, args[0].Type())
				}
				rec1, ok := args[0].(*GoObj[rl.Rectangle])
				if !ok {
					return newPositionalTypeErrorForGoObj("rectangle_check_collision", 1, "rl.Rectangle", args[0])
				}
				if args[1].Type() != GO_OBJ {
					return newPositionalTypeError("rectangle_check_collision", 2, GO_OBJ, args[1].Type())
				}
				rec2, ok := args[1].(*GoObj[rl.Rectangle])
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
	},
	{
		Name: "_vector2",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("Vector2", len(args), 2, "")
				}
				if args[0].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Vector2", 1, FLOAT_OBJ, args[0].Type())
				}
				if args[1].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Vector2", 2, FLOAT_OBJ, args[1].Type())
				}
				x := float32(args[0].(*Float).Value)
				y := float32(args[1].(*Float).Value)
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
	},
	{
		Name: "_vector3",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 3 {
					return newInvalidArgCountError("Vector3", len(args), 3, "")
				}
				if args[0].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Vector3", 1, FLOAT_OBJ, args[0].Type())
				}
				if args[1].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Vector3", 2, FLOAT_OBJ, args[1].Type())
				}
				if args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Vector3", 3, FLOAT_OBJ, args[2].Type())
				}
				x := float32(args[0].(*Float).Value)
				y := float32(args[1].(*Float).Value)
				z := float32(args[2].(*Float).Value)
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
	},
	{
		Name: "_vector4",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 4 {
					return newInvalidArgCountError("Vector4", len(args), 4, "")
				}
				if args[0].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Vector4", 1, FLOAT_OBJ, args[0].Type())
				}
				if args[1].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Vector4", 2, FLOAT_OBJ, args[1].Type())
				}
				if args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Vector4", 3, FLOAT_OBJ, args[2].Type())
				}
				if args[3].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Vector4", 4, FLOAT_OBJ, args[3].Type())
				}
				x := float32(args[0].(*Float).Value)
				y := float32(args[1].(*Float).Value)
				z := float32(args[2].(*Float).Value)
				w := float32(args[3].(*Float).Value)
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
	},
	{
		Name: "_camera2d",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 4 {
					return newInvalidArgCountError("Camera2D", len(args), 4, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("Camera2D", 1, GO_OBJ, args[0].Type())
				}
				offset, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("Camera2D", 1, "rl.Vector2", args[0])
				}
				if args[1].Type() != GO_OBJ {
					return newPositionalTypeError("Camera2D", 2, GO_OBJ, args[1].Type())
				}
				target, ok := args[1].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("Camera2D", 2, "rl.Vector2", args[1])
				}
				if args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Camera2D", 3, FLOAT_OBJ, args[2].Type())
				}
				if args[3].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Camera2D", 4, FLOAT_OBJ, args[3].Type())
				}
				rotation := float32(args[2].(*Float).Value)
				zoom := float32(args[3].(*Float).Value)
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
	},
	{
		Name: "_begin_mode2d",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("begin_mode2d", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("begin_mode2d", 1, GO_OBJ, args[0].Type())
				}
				cam, ok := args[0].(*GoObj[rl.Camera2D])
				if !ok {
					return newPositionalTypeErrorForGoObj("begin_mode2d", 1, "rl.Camera2D", args[0])
				}
				rl.BeginMode2D(cam.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`begin_mode2d` initializes 2d with a custom 2d camera",
				signature:   "begin_mode2d(cam: GoObj[rl.Camera2D]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "begin_mode2d(cam) => null",
			}.String(),
		},
	},
	{
		Name: "_end_mode2d",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("end_mode2d", len(args), 0, "")
				}
				rl.EndMode2D()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`end_mode2d` ends 2d camera mode",
				signature:   "end_mode2d() -> null",
				errors:      "InvalidArgCount",
				example:     "end_mode2d() => null",
			}.String(),
		},
	},
	{
		Name: "_camera3d",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 5 {
					return newInvalidArgCountError("Camera3D", len(args), 5, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("Camera3D", 1, GO_OBJ, args[0].Type())
				}
				position, ok := args[0].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("Camera3D", 1, "rl.Vector3", args[0])
				}
				if args[1].Type() != GO_OBJ {
					return newPositionalTypeError("Camera3D", 2, GO_OBJ, args[1].Type())
				}
				target, ok := args[1].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("Camera3D", 2, "rl.Vector3", args[1])
				}
				if args[2].Type() != GO_OBJ {
					return newPositionalTypeError("Camera3D", 3, GO_OBJ, args[2].Type())
				}
				up, ok := args[2].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("Camera3D", 3, "rl.Vector3", args[2])
				}
				if args[3].Type() != FLOAT_OBJ {
					return newPositionalTypeError("Camera3D", 4, FLOAT_OBJ, args[3].Type())
				}
				if args[4].Type() != INTEGER_OBJ {
					return newPositionalTypeError("Camera3D", 5, INTEGER_OBJ, args[4].Type())
				}
				fovy := float32(args[3].(*Float).Value)
				projection := rl.CameraProjection(args[4].(*Integer).Value)
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
	},
	{
		Name: "_begin_mode3d",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("begin_mode3d", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("begin_mode3d", 1, GO_OBJ, args[0].Type())
				}
				cam, ok := args[0].(*GoObj[rl.Camera3D])
				if !ok {
					return newPositionalTypeErrorForGoObj("begin_mode3d", 1, "rl.Camera3D", args[0])
				}
				rl.BeginMode3D(cam.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`begin_mode3d` begins 3d camera mode with the custom camera",
				signature:   "begin_mode3d(cam: GoObj[rl.Camera3D]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "begin_mode3d() => null",
			}.String(),
		},
	},
	{
		Name: "_end_mode3d",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("end_mode3d", len(args), 0, "")
				}
				rl.EndMode3D()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`end_mode3d` ends 3d camera mode",
				signature:   "end_mode3d() -> null",
				errors:      "InvalidArgCount",
				example:     "end_mode3d() => null",
			}.String(),
		},
	},
	{
		Name: "_init_audio_device",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("init_audio_device", len(args), 0, "")
				}
				rl.InitAudioDevice()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`init_audio_device` initalizes the audio device and context",
				signature:   "init_audio_device() -> null",
				errors:      "InvalidArgCount",
				example:     "init_audio_device() => null",
			}.String(),
		},
	},
	{
		Name: "_close_audio_device",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("close_audio_device", len(args), 0, "")
				}
				rl.CloseAudioDevice()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`close_audio_device` closes the audio device and context",
				signature:   "close_audio_device() -> null",
				errors:      "InvalidArgCount",
				example:     "close_audio_device() => null",
			}.String(),
		},
	},
	{
		Name: "_load_music",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("load_music", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("load_music", 1, STRING_OBJ, args[0].Type())
				}
				fname := args[0].(*Stringo).Value
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
	},
	{
		Name: "_update_music",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("update_music", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("update_music", 1, GO_OBJ, args[0].Type())
				}
				music, ok := args[0].(*GoObj[rl.Music])
				if !ok {
					return newPositionalTypeErrorForGoObj("update_music", 1, "rl.Music", args[0])
				}
				rl.UpdateMusicStream(music.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`update_music` updates the buffer for music streaming from the given music object",
				signature:   "update_music(music: GoObj[rl.Music]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "update_music(music) => null",
			}.String(),
		},
	},
	{
		Name: "_play_music",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("play_music", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("play_music", 1, GO_OBJ, args[0].Type())
				}
				music, ok := args[0].(*GoObj[rl.Music])
				if !ok {
					return newPositionalTypeErrorForGoObj("play_music", 1, "rl.Music", args[0])
				}
				rl.PlayMusicStream(music.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`play_music` starts playing the music from the given music object",
				signature:   "play_music(music: GoObj[rl.Music]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "play_music(music) => null",
			}.String(),
		},
	},
	{
		Name: "_stop_music",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("stop_music", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("stop_music", 1, GO_OBJ, args[0].Type())
				}
				music, ok := args[0].(*GoObj[rl.Music])
				if !ok {
					return newPositionalTypeErrorForGoObj("stop_music", 1, "rl.Music", args[0])
				}
				rl.StopMusicStream(music.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`stop_music` stops playing the music from the given music object",
				signature:   "stop_music(music: GoObj[rl.Music]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "stop_music(music) => null",
			}.String(),
		},
	},
	{
		Name: "_resume_music",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("resume_music", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("resume_music", 1, GO_OBJ, args[0].Type())
				}
				music, ok := args[0].(*GoObj[rl.Music])
				if !ok {
					return newPositionalTypeErrorForGoObj("resume_music", 1, "rl.Music", args[0])
				}
				rl.ResumeMusicStream(music.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`resume_music` resumes playing the paused music from the given music object",
				signature:   "resume_music(music: GoObj[rl.Music]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "resume_music(music) => null",
			}.String(),
		},
	},
	{
		Name: "_pause_music",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("pause_music", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("pause_music", 1, GO_OBJ, args[0].Type())
				}
				music, ok := args[0].(*GoObj[rl.Music])
				if !ok {
					return newPositionalTypeErrorForGoObj("pause_music", 1, "rl.Music", args[0])
				}
				rl.PauseMusicStream(music.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`pause_music` pauses the music from the given music object",
				signature:   "pause_music(music: GoObj[rl.Music]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "pause_music(music) => null",
			}.String(),
		},
	},
	{
		Name: "_load_sound",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("load_sound", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("load_sound", 1, STRING_OBJ, args[0].Type())
				}
				fname := args[0].(*Stringo).Value
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
	},
	{
		Name: "_play_sound",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("play_sound", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("play_sound", 1, GO_OBJ, args[0].Type())
				}
				sound, ok := args[0].(*GoObj[rl.Sound])
				if !ok {
					return newPositionalTypeErrorForGoObj("play_sound", 1, "rl.Sound", args[0])
				}
				rl.PlaySound(sound.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`play_sound` starts playing the sound from the given sound object",
				signature:   "play_sound(sound: GoObj[rl.Sound]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "play_sound(sound) => null",
			}.String(),
		},
	},
	{
		Name: "_stop_sound",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("stop_sound", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("stop_sound", 1, GO_OBJ, args[0].Type())
				}
				sound, ok := args[0].(*GoObj[rl.Sound])
				if !ok {
					return newPositionalTypeErrorForGoObj("stop_sound", 1, "rl.Sound", args[0])
				}
				rl.StopSound(sound.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`stop_sound` stops playing the sound from the given sound object",
				signature:   "stop_sound(sound: GoObj[rl.Sound]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "stop_sound(sound) => null",
			}.String(),
		},
	},
	{
		Name: "_resume_sound",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("resume_sound", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("resume_sound", 1, GO_OBJ, args[0].Type())
				}
				sound, ok := args[0].(*GoObj[rl.Sound])
				if !ok {
					return newPositionalTypeErrorForGoObj("resume_sound", 1, "rl.Sound", args[0])
				}
				rl.ResumeSound(sound.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`resume_sound` resumes playing the paused sound from the given sound object",
				signature:   "resume_sound(sound: GoObj[rl.Sound]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "resume_sound(sound) => null",
			}.String(),
		},
	},
	{
		Name: "_pause_sound",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("pause_sound", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("pause_sound", 1, GO_OBJ, args[0].Type())
				}
				sound, ok := args[0].(*GoObj[rl.Sound])
				if !ok {
					return newPositionalTypeErrorForGoObj("pause_sound", 1, "rl.Sound", args[0])
				}
				rl.PauseSound(sound.Value)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`pause_sound` pauses the sound from the given sound object",
				signature:   "pause_sound(sound: GoObj[rl.Sound]) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "pause_sound(sound) => null",
			}.String(),
		},
	},
	{
		Name: "_unload",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				for i, arg := range args {
					// If the arg is a list go through the list and check every arg to remove
					if arg.Type() == LIST_OBJ {
						l := arg.(*List).Elements
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
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`unload` unloads the given objects from the gg object",
				signature:   "unload(args...: GoObject[rl.Texture2D|rl.Music|rl.Sound]|list[GoObject[rl.Texture2D|rl.Music|rl.Sound]]) -> null",
				errors:      "CustomError",
				example:     "unload() => null",
			}.String(),
		},
	},
}

func unloadFromRaylib(arg Object, pos int) Object {
	if arg.Type() != GO_OBJ {
		return newPositionalTypeError("unload", pos, GO_OBJ, arg.Type())
	}
	if tex, ok := arg.(*GoObj[rl.Texture2D]); ok {
		rl.UnloadTexture(tex.Value)
		return NULL
	} else if music, ok := arg.(*GoObj[rl.Music]); ok {
		rl.UnloadMusicStream(music.Value)
		return NULL
	} else if sound, ok := arg.(*GoObj[rl.Sound]); ok {
		rl.UnloadSound(sound.Value)
		return NULL
	}
	return newError("`unload` error: Failed to find gg object to unload, expected any GO_OBJ of [rl.Texture2D, rl.Music, rl.Sound]")
}