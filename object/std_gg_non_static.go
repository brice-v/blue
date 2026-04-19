//go:build !static
// +build !static

package object

import (
	"blue/consts"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var GgBuiltins = []*Builtin{
	{
		Name: "_init_window",
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
	{
		Name: "_close_window",
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
	{
		Name: "_window_should_close",
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
	{
		Name: "_is_window_ready",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("is_window_ready", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.IsWindowReady())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_window_ready` returns true if the window is ready",
			signature:   "is_window_ready() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_window_ready() => false",
		}.String(),
	},
	{
		Name: "_is_window_fullscreen",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("is_window_fullscreen", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.IsWindowFullscreen())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_window_fullscreen` returns true if the window is fullscreen",
			signature:   "is_window_fullscreen() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_window_fullscreen() => false",
		}.String(),
	},
	{
		Name: "_is_window_hidden",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("is_window_hidden", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.IsWindowFullscreen())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_window_hidden` returns true if the window is hidden",
			signature:   "is_window_hidden() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_window_hidden() => false",
		}.String(),
	},
	{
		Name: "_is_window_maximized",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("is_window_maximized", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.IsWindowMaximized())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_window_maximized` returns true if the window is maximized",
			signature:   "is_window_maximized() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_window_maximized() => false",
		}.String(),
	},
	{
		Name: "_is_window_minimized",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("is_window_minimized", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.IsWindowMinimized())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_window_minimized` returns true if the window is minimized",
			signature:   "is_window_minimized() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_window_minimized() => false",
		}.String(),
	},
	{
		Name: "_is_window_focused",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("is_window_focused", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.IsWindowFocused())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_window_focused` returns true if the window is focused",
			signature:   "is_window_focused() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_window_focused() => false",
		}.String(),
	},
	{
		Name: "_is_window_resized",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("is_window_resized", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.IsWindowResized())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_window_resized` returns true if the window is resized",
			signature:   "is_window_resized() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_window_resized() => false",
		}.String(),
	},
	{
		Name: "_toggle_window_fullscreen",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("toggle_window_fullscreen", len(args), 0, "")
			}
			rl.ToggleFullscreen()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`toggle_window_fullscreen` toggles the window to be fullscreen or not",
			signature:   "toggle_window_fullscreen() -> null",
			errors:      "InvalidArgCount",
			example:     "toggle_window_fullscreen() => null",
		}.String(),
	},
	{
		Name: "_maximize_window",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("maximize_window", len(args), 0, "")
			}
			rl.MaximizeWindow()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`maximize_window` maximizes the window",
			signature:   "maximize_window() -> null",
			errors:      "InvalidArgCount",
			example:     "maximize_window() => null",
		}.String(),
	},
	{
		Name: "_minimize_window",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("minimize_window", len(args), 0, "")
			}
			rl.MinimizeWindow()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`minimize_window` maximizes the window",
			signature:   "minimize_window() -> null",
			errors:      "InvalidArgCount",
			example:     "minimize_window() => null",
		}.String(),
	},
	{
		Name: "_restore_window",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("restore_window", len(args), 0, "")
			}
			rl.RestoreWindow()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`restore_window` restores the window",
			signature:   "restore_window() -> null",
			errors:      "InvalidArgCount",
			example:     "restore_window() => null",
		}.String(),
	},
	{
		Name: "_set_window_icon",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_window_icon", len(args), 1, "")
			}
			if args[0].Type() == LIST_OBJ {
				elemList := args[0].(*List).Elements
				icons := make([]rl.Image, len(elemList))
				for i, e := range elemList {
					iconImage, ok := e.(*GoObj[rl.Image])
					if !ok {
						return newPositionalTypeErrorForGoObj("set_window_icon", i+1, "rl.Image", e)
					}
					icons[i] = iconImage.Value
				}
				rl.SetWindowIcons(icons, int32(len(elemList)))
			} else {
				icon, ok := args[0].(*GoObj[rl.Image])
				if !ok {
					return newPositionalTypeErrorForGoObj("set_window_icon", 1, "rl.Image", args[0])
				}
				rl.SetWindowIcon(icon.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_window_icon` sets the icon or icons for the window",
			signature:   "set_window_icon(icon: GoObj[rl.Image]|list[GoObj[rl.Image]]) -> null",
			errors:      "InvalidArgCount,PositionalTypeError",
			example:     "set_window_icon(icon) => null",
		}.String(),
	},
	{
		Name: "_set_window_title",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_window_title", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("set_window_title", 1, STRING_OBJ, args[0].Type())
			}
			rl.SetWindowTitle(args[0].(*Stringo).Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_window_title` sets the windows title to the given string",
			signature:   "set_window_title() -> null",
			errors:      "InvalidArgCount,PositionalTypeError",
			example:     "set_window_title() => null",
		}.String(),
	},
	{
		Name: "_set_window_position",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("set_window_position", len(args), 2, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_window_position", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_window_position", 2, INTEGER_OBJ, args[1].Type())
			}
			x := int(args[0].(*Integer).Value)
			y := int(args[1].(*Integer).Value)
			rl.SetWindowPosition(x, y)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_window_position` sets the window position to x and y",
			signature:   "set_window_position(x: int, y: int) -> null",
			errors:      "InvalidArgCount,PositionalTypeError",
			example:     "set_window_position(100, 200) => null",
		}.String(),
	},
	{
		Name: "_set_window_monitor",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_window_monitor", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_window_monitor", 1, INTEGER_OBJ, args[0].Type())
			}
			rl.SetWindowMonitor(int(args[0].(*Integer).Value))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_window_monitor` sets the monitor for the current window",
			signature:   "set_window_monitor(monitor: int) -> null",
			errors:      "InvalidArgCount,PositionalTypeError",
			example:     "set_window_monitor(0) => null",
		}.String(),
	},
	{
		Name: "_set_window_min_size",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("set_window_min_size", len(args), 2, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_window_min_size", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_window_min_size", 2, INTEGER_OBJ, args[1].Type())
			}
			w := int(args[0].(*Integer).Value)
			h := int(args[1].(*Integer).Value)
			rl.SetWindowMinSize(w, h)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_window_min_size` sets the minimum window size for resizable windows",
			signature:   "set_window_min_size(w: int, h: int) -> null",
			errors:      "InvalidArgCount,PositionalTypeError",
			example:     "set_window_min_size(200, 200) => null",
		}.String(),
	},
	{
		Name: "_set_window_size",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("set_window_size", len(args), 2, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_window_size", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_window_size", 2, INTEGER_OBJ, args[1].Type())
			}
			w := int(args[0].(*Integer).Value)
			h := int(args[1].(*Integer).Value)
			rl.SetWindowSize(w, h)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_window_size` sets the window size",
			signature:   "set_window_size(w: int, h: int) -> null",
			errors:      "InvalidArgCount,PositionalTypeError",
			example:     "set_window_size(800, 600) => null",
		}.String(),
	},
	{
		Name: "_set_window_opacity",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_window_opacity", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("set_window_opacity", 1, FLOAT_OBJ, args[0].Type())
			}
			rl.SetWindowOpacity(float32(args[0].(*Float).Value))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_window_opacity` sets the window opacity (value from 0.0-1.0)",
			signature:   "set_window_opacity(o: float) -> null",
			errors:      "InvalidArgCount,PositionalTypeError",
			example:     "set_window_opacity(0.6) => null",
		}.String(),
	},
	{
		Name: "_set_window_focused",
		Fun: func(args ...Object) Object {
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_window_focused` ",
			signature:   "set_window_focused() -> null",
			errors:      "InvalidArgCount,PositionalTypeError",
			example:     "set_window_focused() => null",
		}.String(),
	},
	{
		Name: "_get_screen_width",
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
	{
		Name: "_get_screen_height",
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
	{
		Name: "_get_monitor_count",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_monitor_count", len(args), 0, "")
			}
			return &Integer{Value: int64(rl.GetMonitorCount())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_monitor_count` gets the number of monitors as an int",
			signature:   "get_monitor_count() -> int",
			errors:      "InvalidArgCount",
			example:     "get_monitor_count() => 3",
		}.String(),
	},
	{
		Name: "_get_current_monitor",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_current_monitor", len(args), 0, "")
			}
			return &Integer{Value: int64(rl.GetCurrentMonitor())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_current_monitor` gets the current connected monitor as an int",
			signature:   "get_current_monitor() -> int",
			errors:      "InvalidArgCount",
			example:     "get_current_monitor() => 0",
		}.String(),
	},
	{
		Name: "_get_monitor_position",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("get_monitor_position", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("get_monitor_position", 1, INTEGER_OBJ, args[0].Type())
			}
			monitor := int(args[0].(*Integer).Value)
			return NewGoObj(rl.GetMonitorPosition(monitor))
		},
		HelpStr: helpStrArgs{
			explanation: "`get_monitor_position` gets given monitors position",
			signature:   "get_monitor_position(monitor: int) -> rl.Vector2",
			errors:      "InvalidArgCount,PositionalType",
			example:     "get_monitor_position(0) => rl.Vector2[300,100]",
		}.String(),
	},
	{
		Name: "_get_monitor_width",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("get_monitor_width", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("get_monitor_width", 1, INTEGER_OBJ, args[0].Type())
			}
			monitor := int(args[0].(*Integer).Value)
			return &Integer{Value: int64(rl.GetMonitorWidth(monitor))}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_monitor_width` gets the width of the given monitor as an int",
			signature:   "get_monitor_width(monitor: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "get_monitor_width(0) => 800",
		}.String(),
	},
	{
		Name: "_get_monitor_height",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("get_monitor_height", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("get_monitor_height", 1, INTEGER_OBJ, args[0].Type())
			}
			monitor := int(args[0].(*Integer).Value)
			return &Integer{Value: int64(rl.GetMonitorHeight(monitor))}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_monitor_height` gets the height of the given monitor as an int",
			signature:   "get_monitor_height(monitor: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "get_monitor_height(0) => 800",
		}.String(),
	},
	{
		Name: "_get_monitor_physical_width",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("get_monitor_physical_width", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("get_monitor_physical_width", 1, INTEGER_OBJ, args[0].Type())
			}
			monitor := int(args[0].(*Integer).Value)
			return &Integer{Value: int64(rl.GetMonitorPhysicalWidth(monitor))}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_monitor_physical_width` gets the physical width of the given monitor as an int",
			signature:   "get_monitor_physical_width(monitor: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "get_monitor_physical_width(0) => 800",
		}.String(),
	},
	{
		Name: "_get_monitor_physical_height",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("get_monitor_physical_height", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("get_monitor_physical_height", 1, INTEGER_OBJ, args[0].Type())
			}
			monitor := int(args[0].(*Integer).Value)
			return &Integer{Value: int64(rl.GetMonitorHeight(monitor))}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_monitor_physical_height` gets the physical height of the given monitor as an int",
			signature:   "get_monitor_physical_height(monitor: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "get_monitor_physical_height(0) => 800",
		}.String(),
	},
	{
		Name: "_get_monitor_refresh_rate",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("get_monitor_refresh_rate", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("get_monitor_refresh_rate", 1, INTEGER_OBJ, args[0].Type())
			}
			monitor := int(args[0].(*Integer).Value)
			return &Integer{Value: int64(rl.GetMonitorRefreshRate(monitor))}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_monitor_refresh_rate` gets the refresh rate of the given monitor as an int",
			signature:   "get_monitor_refresh_rate(monitor: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "get_monitor_refresh_rate(0) => 800",
		}.String(),
	},
	{
		Name: "_get_monitor_name",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("get_monitor_name", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("get_monitor_name", 1, INTEGER_OBJ, args[0].Type())
			}
			monitor := int(args[0].(*Integer).Value)
			return &Stringo{Value: rl.GetMonitorName(monitor)}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_monitor_name` gets the refresh rate of the given monitor as an int",
			signature:   "get_monitor_name(monitor: int) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "get_monitor_name(0) => 'Hello'",
		}.String(),
	},
	{
		Name: "_get_clipboard_text",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_clipboard_text", len(args), 0, "")
			}
			return &Stringo{Value: rl.GetClipboardText()}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_clipboard_text` gets the clipboard text as a string",
			signature:   "get_clipboard_text() -> str",
			errors:      "InvalidArgCount",
			example:     "get_clipboard_text() => 'Hello'",
		}.String(),
	},
	{
		Name: "_set_clipboard_text",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_clipboard_text", len(args), 0, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("set_clipboard_text", 1, STRING_OBJ, args[0].Type())
			}
			rl.SetClipboardText(args[0].(*Stringo).Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_clipboard_text` sets the clipboard text with the given string",
			signature:   "set_clipboard_text(data: str) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_clipboard_text('Hello') => null",
		}.String(),
	},
	{
		Name: "_show_cursor",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("show_cursor", len(args), 0, "")
			}
			rl.ShowCursor()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`show_cursor` shows the cursor",
			signature:   "show_cursor() -> null",
			errors:      "InvalidArgCount",
			example:     "show_cursor() => null",
		}.String(),
	},
	{
		Name: "_hide_cursor",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("hide_cursor", len(args), 0, "")
			}
			rl.HideCursor()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`hide_cursor` hides the cursor",
			signature:   "hide_cursor() -> null",
			errors:      "InvalidArgCount",
			example:     "hide_cursor() => null",
		}.String(),
	},
	{
		Name: "_is_cursor_hidden",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("is_cursor_hidden", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.IsCursorHidden())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_cursor_hidden` returns true if the cursor is hidden",
			signature:   "is_cursor_hidden() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_cursor_hidden() => false",
		}.String(),
	},
	{
		Name: "_enable_cursor",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("enable_cursor", len(args), 0, "")
			}
			rl.EnableCursor()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`enable_cursor` enables the cursor",
			signature:   "enable_cursor() -> null",
			errors:      "InvalidArgCount",
			example:     "enable_cursor() => null",
		}.String(),
	},
	{
		Name: "_disable_cursor",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("disable_cursor", len(args), 0, "")
			}
			rl.DisableCursor()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`disable_cursor` disables the cursor",
			signature:   "disable_cursor() -> null",
			errors:      "InvalidArgCount",
			example:     "disable_cursor() => null",
		}.String(),
	},
	{
		Name: "_is_cursor_on_screen",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("is_cursor_on_screen", len(args), 0, "")
			}
			return nativeToBooleanObject(rl.IsCursorOnScreen())
		},
		HelpStr: helpStrArgs{
			explanation: "`is_cursor_on_screen` returns true if the cursor is on screen",
			signature:   "is_cursor_on_screen() -> bool",
			errors:      "InvalidArgCount",
			example:     "is_cursor_on_screen() => false",
		}.String(),
	},
	{
		Name: "_begin_drawing",
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
	{
		Name: "_end_drawing",
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
	{
		Name: "_clear_background",
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
	{
		Name: "_color_map",
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
				Name: "new_color",
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
			mapObj.Set("light_grey", lightGray)
			mapObj.Set("gray", gray)
			mapObj.Set("grey", gray)
			mapObj.Set("dark_gray", darkGray)
			mapObj.Set("dark_grey", darkGray)
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
	{
		Name: "_draw_text",
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
	{
		Name: "_draw_texture",
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
	{
		Name: "_draw_texture_pro",
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
	{
		Name: "_draw_pixel",
		Fun: func(args ...Object) Object {
			if len(args) != 3 && len(args) != 2 {
				return newInvalidArgCountError("draw_pixel", len(args), 3, "or 2")
			}
			if len(args) == 3 {
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_pixel", 1, INTEGER_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_pixel", 2, INTEGER_OBJ, args[1].Type())
				}
				color, ok := args[2].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_pixel", 3, "rl.Color", args[2])
				}
				rl.DrawPixel(int32(args[0].(*Integer).Value), int32(args[1].(*Integer).Value), color.Value)
			} else {
				pos, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_pixel", 1, "rl.Vector2", args[0])
				}
				color, ok := args[1].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_pixel", 2, "rl.Color", args[1])
				}
				rl.DrawPixelV(pos.Value, color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_pixel` draw a pixel",
			signature: "draw_pixel(pox_x: int, pos_y: int, color: color) -> null\n" +
				"// Draw a pixel using geometry [Can be slow, use with care]\n" +
				"void DrawPixel(int posX, int posY, Color color);\n" +
				"// Draw a pixel using geometry (Vector version) [Can be slow, use with care]\n" +
				"void DrawPixelV(Vector2 position, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_pixel(0, 0, color.red) => null",
		}.String(),
	},
	{
		Name: "_draw_line",
		Fun: func(args ...Object) Object {
			if len(args) != 5 && len(args) != 3 && len(args) != 4 {
				return newInvalidArgCountError("draw_line", len(args), 5, "or 4 or 3")
			}
			if len(args) == 5 {
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_line", 1, INTEGER_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_line", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_line", 3, INTEGER_OBJ, args[2].Type())
				}
				if args[3].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_line", 4, INTEGER_OBJ, args[3].Type())
				}
				color, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line", 5, "rl.Color", args[4])
				}
				rl.DrawLine(int32(args[0].(*Integer).Value), int32(args[1].(*Integer).Value), int32(args[2].(*Integer).Value), int32(args[3].(*Integer).Value), color.Value)
			} else if len(args) == 4 {
				startPos, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line", 1, "rl.Vector2", args[0])
				}
				endPos, ok := args[1].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line", 2, "rl.Vector2", args[1])
				}
				if args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_line", 3, FLOAT_OBJ, args[2].Type())
				}
				color, ok := args[3].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line", 4, "rl.Vector2", args[3])
				}
				rl.DrawLineEx(startPos.Value, endPos.Value, float32(args[2].(*Float).Value), color.Value)
			} else {
				startPos, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line", 1, "rl.Vector2", args[0])
				}
				endPos, ok := args[1].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line", 2, "rl.Vector2", args[1])
				}
				color, ok := args[2].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line", 3, "rl.Vector2", args[2])
				}
				rl.DrawLineV(startPos.Value, endPos.Value, color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_line` draw a line",
			signature: "draw_line(start_pos_x: int, start_pos_y: int, end_pos_x: int, end_pos_y: int) -> null\n" +
				"// Draw a line\n" +
				"void DrawLine(int startPosX, int startPosY, int endPosX, int endPosY, Color color);\n" +
				"// Draw a line (using gl lines)\n" +
				"void DrawLineV(Vector2 startPos, Vector2 endPos, Color color);\n" +
				"// Draw a line (using triangles/quads)\n" +
				"void DrawLineEx(Vector2 startPos, Vector2 endPos, float thick, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_line() => (see signature for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_line_strip",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("draw_line_strip", len(args), 2, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("draw_line_strip", 1, LIST_OBJ, args[0].Type())
			}
			points := make([]rl.Vector2, len(args[0].(*List).Elements))
			for i, e := range args[0].(*List).Elements {
				point, ok := e.(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_strip", 1, "list[rl.Vector2]", e)
				}
				points[i] = point.Value
			}
			color, ok := args[1].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_line_strip", 2, "rl.Color", args[1])
			}
			rl.DrawLineStrip(points, int32(len(points)), color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_line_strip` draws a line sequence",
			signature: "draw_line_strip(points: list[rl.Vector2], color: color) -> null\n" +
				"// Draw lines sequence (using gl lines)\n" +
				"void DrawLineStrip(const Vector2 *points, int pointCount, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_line_strip() => (see signature for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_line_bezier",
		Fun: func(args ...Object) Object {
			if len(args) != 4 && len(args) != 5 && len(args) != 6 {
				return newInvalidArgCountError("draw_line_bezier", len(args), 4, "or 5 or 6")
			}
			if len(args) == 4 {
				startPos, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 1, "rl.Vector2", args[0])
				}
				endPos, ok := args[1].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 2, "rl.Vector2", args[1])
				}
				if args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_line_bezier", 3, FLOAT_OBJ, args[2].Type())
				}
				color, ok := args[3].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 4, "rl.Color", args[3])
				}
				rl.DrawLineBezier(startPos.Value, endPos.Value, float32(args[2].(*Float).Value), color.Value)
			} else if len(args) == 5 {
				startPos, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 1, "rl.Vector2", args[0])
				}
				endPos, ok := args[1].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 2, "rl.Vector2", args[1])
				}
				controlPos, ok := args[2].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 3, "rl.Vector2", args[2])
				}
				if args[3].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_line_bezier", 4, FLOAT_OBJ, args[3].Type())
				}
				color, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 5, "rl.Color", args[4])
				}
				rl.DrawLineBezierQuad(startPos.Value, endPos.Value, controlPos.Value, float32(args[3].(*Float).Value), color.Value)
			} else {
				startPos, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 1, "rl.Vector2", args[0])
				}
				endPos, ok := args[1].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 2, "rl.Vector2", args[1])
				}
				startControlPos, ok := args[2].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 3, "rl.Vector2", args[2])
				}
				endControlPos, ok := args[3].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 4, "rl.Vector2", args[3])
				}
				if args[4].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_line_bezier", 5, FLOAT_OBJ, args[4].Type())
				}
				color, ok := args[5].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_line_bezier", 6, "rl.Color", args[5])
				}
				rl.DrawLineBezierCubic(startPos.Value, endPos.Value, startControlPos.Value, endControlPos.Value, float32(args[4].(*Float).Value), color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_line_bezier` draws a line with cubic bezier curves in-out",
			signature: "draw_line_bezier(start_pos: rl.Vector2, end_pos: rl.Vector2, thick: float, color: color) -> null\n" +
				"Cubic and Quad also available",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_line_bezier() => (see signature for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_circle",
		Fun: func(args ...Object) Object {
			if len(args) != 5 && len(args) != 4 && len(args) != 3 {
				return newInvalidArgCountError("draw_circle", len(args), 3, "or 4 or 5")
			}
			if len(args) == 3 {
				center, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_circle", 1, "rl.Vector2", args[0])
				}
				if args[1].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_circle", 2, FLOAT_OBJ, args[1].Type())
				}
				color, ok := args[2].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_circle", 3, "rl.Color", args[2])
				}
				rl.DrawCircleV(center.Value, float32(args[1].(*Float).Value), color.Value)
			} else if len(args) == 4 {
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_circle", 1, INTEGER_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_circle", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_circle", 3, FLOAT_OBJ, args[2].Type())
				}
				color, ok := args[3].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_circle", 4, "rl.Color", args[3])
				}
				rl.DrawCircle(int32(args[0].(*Integer).Value), int32(args[1].(*Integer).Value), float32(args[2].(*Float).Value), color.Value)
			} else {
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_circle", 1, INTEGER_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_circle", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_circle", 3, FLOAT_OBJ, args[2].Type())
				}
				colorInner, ok := args[3].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_circle", 4, "rl.Color", args[3])
				}
				colorOuter, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_circle", 5, "rl.Color", args[4])
				}
				rl.DrawCircleGradient(int32(args[0].(*Integer).Value), int32(args[1].(*Integer).Value), float32(args[2].(*Float).Value), colorInner.Value, colorOuter.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_circle` draws a colored fill circle",
			signature: "draw_circle(center_x: int, center_y: int, radius: float, color: color) -> null\n" +
				"// Draw a color-filled circle\n" +
				"void DrawCircle(int centerX, int centerY, float radius, Color color);\n" +
				"// Draw a gradient-filled circle\n" +
				"void DrawCircleGradient(int centerX, int centerY, float radius, Color inner, Color outer);\n" +
				"// Draw a color-filled circle (Vector version)\n" +
				"void DrawCircleV(Vector2 center, float radius, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_circle() => (see signature for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_circle_sector",
		Fun: func(args ...Object) Object {
			if len(args) != 7 {
				return newInvalidArgCountError("draw_circle_sector", len(args), 7, "")
			}
			center, ok := args[0].(*GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_circle_sector", 1, "rl.Vector2", args[0])
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_circle_sector", 2, FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_circle_sector", 3, FLOAT_OBJ, args[2].Type())
			}
			if args[3].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_circle_sector", 4, FLOAT_OBJ, args[3].Type())
			}
			if args[4].Type() != INTEGER_OBJ {
				return newPositionalTypeError("draw_circle_sector", 5, INTEGER_OBJ, args[4].Type())
			}
			color, ok := args[5].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_circle_sector", 6, "rl.Color", args[5])
			}
			if args[6].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("draw_circle_sector", 7, BOOLEAN_OBJ, args[7].Type())
			}
			if args[6].(*Boolean).Value {
				rl.DrawCircleSectorLines(center.Value, float32(args[1].(*Float).Value), float32(args[2].(*Float).Value), float32(args[3].(*Float).Value), int32(args[4].(*Integer).Value), color.Value)
			} else {
				rl.DrawCircleSector(center.Value, float32(args[1].(*Float).Value), float32(args[2].(*Float).Value), float32(args[3].(*Float).Value), int32(args[4].(*Integer).Value), color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_circle_sector` draws a piece of a circle",
			signature: "draw_circle_sector(center: GoObj[rl.Vector2], radius: float, start_angle: float, end_angle: float, segments: int, color: color, with_lines: bool=false) -> null\n" +
				"// Draw a piece of a circle\n" +
				"void DrawCircleSector(Vector2 center, float radius, float startAngle, float endAngle, int segments, Color color);\n" +
				"// Draw circle sector outline\n" +
				"void DrawCircleSectorLines(Vector2 center, float radius, float startAngle, float endAngle, int segments, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_circle_sector() => (see signature for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_circle_lines",
		Fun: func(args ...Object) Object {
			if len(args) != 4 {
				return newInvalidArgCountError("draw_circle_lines", len(args), 3, "or 4")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("draw_circle_lines", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("draw_circle_lines", 2, INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_circle_lines", 3, FLOAT_OBJ, args[2].Type())
			}
			color, ok := args[3].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_circle_lines", 4, "rl.Color", args[3])
			}
			rl.DrawCircleLines(int32(args[0].(*Integer).Value), int32(args[1].(*Integer).Value), float32(args[2].(*Float).Value), color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_circle_lines` draws circle outline",
			signature: "draw_circle_lines(center_x: int, center_y: int, radius: float, color: color) -> null\n" +
				"// Draw circle outline\n" +
				"void DrawCircleLines(int centerX, int centerY, float radius, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_circle_lines() => (see signature for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_rectangle",
		Fun: func(args ...Object) Object {
			if len(args) != 5 && len(args) != 4 && len(args) != 3 && len(args) != 2 {
				return newInvalidArgCountError("draw_rectangle", len(args), 5, "or 4 or 3 or 2")
			}
			if len(args) == 5 {
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
					return newPositionalTypeErrorForGoObj("draw_rectangle", 4, "rl.Color", args[4])
				}
				rl.DrawRectangle(posx, posy, width, height, color.Value)
			} else if len(args) == 4 {
				rec, ok := args[0].(*GoObj[rl.Rectangle])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle", 1, "rl.Rectangle", args[0])
				}
				origin, ok := args[1].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle", 2, "rl.Vector2", args[1])
				}
				if args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_rectangle", 3, FLOAT_OBJ, args[2].Type())
				}
				color, ok := args[3].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle", 4, "rl.Color", args[3])
				}
				rl.DrawRectanglePro(rec.Value, origin.Value, float32(args[2].(*Float).Value), color.Value)
			} else if len(args) == 3 {
				position, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle", 1, "rl.Vector2", args[0])
				}
				size, ok := args[1].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle", 2, "rl.Vector2", args[1])
				}
				color, ok := args[2].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle", 3, "rl.Color", args[2])
				}
				rl.DrawRectangleV(position.Value, size.Value, color.Value)
			} else if len(args) == 2 {
				rec, ok := args[0].(*GoObj[rl.Rectangle])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle", 1, "rl.Rectangle", args[0])
				}
				color, ok := args[1].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle", 2, "rl.Color", args[1])
				}
				rl.DrawRectangleRec(rec.Value, color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_rectangle` draws a rectangle at the given position with width and height\n" +
				"// Draw a color-filled rectangle\n" +
				"void DrawRectangle(int posX, int posY, int width, int height, Color color);\n" +
				"// Draw a color-filled rectangle (Vector version)\n" +
				"void DrawRectangleV(Vector2 position, Vector2 size, Color color);\n" +
				"// Draw a color-filled rectangle\n" +
				"void DrawRectangleRec(Rectangle rec, Color color);\n" +
				"// Draw a color-filled rectangle with pro parameters\n" +
				"void DrawRectanglePro(Rectangle rec, Vector2 origin, float rotation, Color color);",
			signature: "draw_rectangle(posx: int, posy: int, width: int, height: int, color=color.black) -> null",
			errors:    "InvalidArgCount,PositionalType",
			example:   "draw_rectangle() (used as Rectangle().draw(color))=> null",
		}.String(),
	},
	{
		Name: "_draw_rectangle_gradient",
		Fun: func(args ...Object) Object {
			if len(args) != 7 {
				return newInvalidArgCountError("draw_rectangle_gradient", len(args), 7, "")
			}
			if args[5] == NULL {
				rec, ok := args[0].(*GoObj[rl.Rectangle])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_gradient", 1, "rl.Rectangle", args[0])
				}
				color1, ok := args[1].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_gradient", 2, "rl.Color", args[1])
				}
				color2, ok := args[2].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_gradient", 3, "rl.Color", args[2])
				}
				color3, ok := args[3].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_gradient", 4, "rl.Color", args[3])
				}
				color4, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_gradient", 5, "rl.Color", args[4])
				}
				rl.DrawRectangleGradientEx(rec.Value, color1.Value, color2.Value, color3.Value, color4.Value)
			} else {
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle_gradient", 1, INTEGER_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle_gradient", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle_gradient", 3, INTEGER_OBJ, args[2].Type())
				}
				if args[3].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle_gradient", 4, INTEGER_OBJ, args[3].Type())
				}
				colorLeftOrTop, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_gradient", 5, "rl.Color", args[4])
				}
				colorRightOrBottom, ok := args[5].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_gradient", 6, "rl.Color", args[5])
				}
				if args[6].Type() != BOOLEAN_OBJ {
					return newPositionalTypeError("draw_rectangle_gradient", 7, BOOLEAN_OBJ, args[6].Type())
				}
				if args[6].(*Boolean).Value {
					// Draw vertical
					rl.DrawRectangleGradientV(int32(args[0].(*Integer).Value), int32(args[1].(*Integer).Value), int32(args[2].(*Integer).Value), int32(args[3].(*Integer).Value), colorLeftOrTop.Value, colorRightOrBottom.Value)
				} else {
					// Draw horizontal
					rl.DrawRectangleGradientH(int32(args[0].(*Integer).Value), int32(args[1].(*Integer).Value), int32(args[2].(*Integer).Value), int32(args[3].(*Integer).Value), colorLeftOrTop.Value, colorRightOrBottom.Value)
				}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_rectangle_gradient` draws a horizontal or vertical gradient filled rectangle or with custom vertex colors\n" +
				"// Draw a vertical-gradient-filled rectangle\n" +
				"void DrawRectangleGradientV(int posX, int posY, int width, int height, Color top, Color bottom);\n" +
				"// Draw a horizontal-gradient-filled rectangle\n" +
				"void DrawRectangleGradientH(int posX, int posY, int width, int height, Color left, Color right);\n" +
				"// Draw a gradient-filled rectangle with custom vertex colors\n" +
				"void DrawRectangleGradientEx(Rectangle rec, Color topLeft, Color bottomLeft, Color topRight, Color bottomRight);\n",
			signature: "draw_rectangle_gradient(posx: int, posy: int, width: int, height: int, color_left_or_top: color, color_right_or_bottom: color) -> null",
			errors:    "InvalidArgCount,PositionalType",
			example:   "draw_rectangle_gradient() (see explanation for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_rectangle_lines",
		Fun: func(args ...Object) Object {
			if len(args) != 5 && len(args) != 3 {
				return newInvalidArgCountError("draw_rectangle_lines", len(args), 5, "or 3")
			}
			if len(args) == 5 {
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle_lines", 1, INTEGER_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle_lines", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle_lines", 3, INTEGER_OBJ, args[2].Type())
				}
				if args[3].Type() != INTEGER_OBJ {
					return newPositionalTypeError("draw_rectangle_lines", 4, INTEGER_OBJ, args[3].Type())
				}
				color, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_lines", 5, "rl.Color", args[4])
				}
				rl.DrawRectangleLines(int32(args[0].(*Integer).Value), int32(args[1].(*Integer).Value), int32(args[2].(*Integer).Value), int32(args[3].(*Integer).Value), color.Value)
			} else {
				rec, ok := args[0].(*GoObj[rl.Rectangle])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_lines", 1, "rl.Rectangle", args[0])
				}
				if args[1].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_rectangle_lines", 2, FLOAT_OBJ, args[1].Type())
				}
				color, ok := args[2].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_rectangle_lines", 3, "rl.Color", args[2])
				}
				rl.DrawRectangleLinesEx(rec.Value, float32(args[1].(*Float).Value), color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_rectangle_lines` draws a rectangle outline\n" +
				"// Draw rectangle outline\n" +
				"void DrawRectangleLines(int posX, int posY, int width, int height, Color color);\n" +
				"// Draw rectangle outline with extended parameters\n" +
				"void DrawRectangleLinesEx(Rectangle rec, float lineThick, Color color);",
			signature: "draw_rectangle_lines(posx: int, posy: int, width: int, height: int, color: color) -> null",
			errors:    "InvalidArgCount,PositionalType",
			example:   "draw_rectangle_lines() (see explanation for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_rectangle_rounded",
		Fun: func(args ...Object) Object {
			if len(args) != 4 {
				return newInvalidArgCountError("draw_rectangle_rounded", len(args), 4, "")
			}
			rec, ok := args[0].(*GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_rectangle_rounded", 1, "rl.Rectangle", args[0])
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_rectangle_rounded", 2, FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != INTEGER_OBJ {
				return newPositionalTypeError("draw_rectangle_rounded", 3, INTEGER_OBJ, args[2].Type())
			}
			color, ok := args[3].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_rectangle_rounded", 4, "rl.Rectangle", args[3])
			}
			rl.DrawRectangleRounded(rec.Value, float32(args[1].(*Float).Value), int32(args[2].(*Integer).Value), color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_rectangle_rounded` draws a rectangle outline\n" +
				"// Draw rectangle with rounded edges\n" +
				"void DrawRectangleRounded(Rectangle rec, float roundness, int segments, Color color);",
			signature: "draw_rectangle_rounded(rec: Rectangle, roundness: float, segments: int, color: color) -> null",
			errors:    "InvalidArgCount,PositionalType",
			example:   "draw_rectangle_rounded() (see explanation for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_rectangle_rounded_lines",
		Fun: func(args ...Object) Object {
			if len(args) != 5 {
				return newInvalidArgCountError("draw_rectangle_rounded_lines", len(args), 5, "")
			}
			rec, ok := args[0].(*GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_rectangle_rounded_lines", 1, "rl.Rectangle", args[0])
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_rectangle_rounded_lines", 2, FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_rectangle_rounded_lines", 3, FLOAT_OBJ, args[2].Type())
			}
			if args[3].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_rectangle_rounded_lines", 4, FLOAT_OBJ, args[3].Type())
			}
			color, ok := args[4].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_rectangle_rounded_lines", 5, "rl.Color", args[4])
			}
			rl.DrawRectangleRoundedLines(rec.Value, float32(args[1].(*Float).Value), float32(args[2].(*Float).Value), float32(args[3].(*Float).Value), color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_rectangle_rounded_lines` draws a rectangle with rounded edge outline\n" +
				"// Draw rectangle with rounded edges outline\n" +
				"void DrawRectangleRoundedLinesEx(Rectangle rec, float roundness, int segments, float lineThick, Color color);",
			signature: "draw_rectangle_rounded_lines(rec: Rectangle, roundness: float, segments: float, line_thick: float, color: color) -> null",
			errors:    "InvalidArgCount,PositionalType",
			example:   "draw_rectangle_rounded_lines() (see explanation for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_triangle",
		Fun: func(args ...Object) Object {
			if len(args) != 5 {
				return newInvalidArgCountError("draw_triangle", len(args), 5, "")
			}
			v1, ok := args[0].(*GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_triangle", 1, "rl.Vector2", args[0])
			}
			v2, ok := args[1].(*GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_triangle", 2, "rl.Vector2", args[1])
			}
			v3, ok := args[2].(*GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_triangle", 3, "rl.Vector2", args[2])
			}
			color, ok := args[3].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_triangle", 4, "rl.Color", args[3])
			}
			if args[4].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("draw_triangle", 5, BOOLEAN_OBJ, args[4].Type())
			}
			useLines := args[4].(*Boolean).Value
			if useLines {
				rl.DrawTriangleLines(v1.Value, v2.Value, v3.Value, color.Value)
			} else {
				rl.DrawTriangle(v1.Value, v2.Value, v3.Value, color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_triangle` draws a triangle with lines if specified\n" +
				"// Draw a color-filled triangle (vertex in counter-clockwise order!)\n" +
				"void DrawTriangle(Vector2 v1, Vector2 v2, Vector2 v3, Color color);\n" +
				"// Draw triangle outline (vertex in counter-clockwise order!)\n" +
				"void DrawTriangleLines(Vector2 v1, Vector2 v2, Vector2 v3, Color color);",
			signature: "draw_triangle(rec: Rectangle, roundness: float, segments: float, line_thick: float, color: color) -> null",
			errors:    "InvalidArgCount,PositionalType",
			example:   "draw_triangle() (see explanation for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_triangle_fan",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("draw_triangle_fan", len(args), 2, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("draw_triangle_fan", 1, LIST_OBJ, args[0].Type())
			}
			points := make([]rl.Vector2, len(args[0].(*List).Elements))
			for i, e := range args[0].(*List).Elements {
				point, ok := e.(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_triangle_fan", 1, "rl.Vector2", e)
				}
				points[i] = point.Value
			}
			color, ok := args[1].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_triangle_fan", 2, "rl.Color", args[1])
			}
			rl.DrawTriangleFan(points, color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_triangle_fan` draws a triangle as a fan defined by points",
			signature:   "draw_triangle_fan(points: list[rl.Vector2], color: color) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "draw_triangle_fan() (see explanation for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_triangle_strip",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("draw_triangle_strip", len(args), 2, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("draw_triangle_strip", 1, LIST_OBJ, args[0].Type())
			}
			points := make([]rl.Vector2, len(args[0].(*List).Elements))
			for i, e := range args[0].(*List).Elements {
				point, ok := e.(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_triangle_strip", 1, "rl.Vector2", e)
				}
				points[i] = point.Value
			}
			color, ok := args[1].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_triangle_strip", 2, "rl.Color", args[1])
			}
			rl.DrawTriangleStrip(points, color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_triangle_strip` draws a triangle as a strip defined by points",
			signature:   "draw_triangle_strip(points: list[rl.Vector2], color: color) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "draw_triangle_strip() (see explanation for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_poly",
		Fun: func(args ...Object) Object {
			if len(args) != 7 {
				return newInvalidArgCountError("draw_poly", len(args), 7, "")
			}
			center, ok := args[0].(*GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_poly", 1, "rl.Vector2", args[0])
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("draw_poly", 2, INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_poly", 3, FLOAT_OBJ, args[2].Type())
			}
			if args[3].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_poly", 4, FLOAT_OBJ, args[3].Type())
			}
			if args[5].Type() != NULL_OBJ {
				if args[4].Type() != FLOAT_OBJ {
					return newPositionalTypeError("draw_poly", 5, FLOAT_OBJ, args[4].Type())
				}
				color, ok := args[5].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_poly", 6, "rl.Color", args[5])
				}
				rl.DrawPolyLinesEx(center.Value, int32(args[1].(*Integer).Value), float32(args[2].(*Float).Value), float32(args[3].(*Float).Value), float32(args[4].(*Float).Value), color.Value)
				return NULL
			}
			color, ok := args[4].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_poly", 5, "rl.Color", args[4])
			}
			if args[6].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("draw_poly", 7, BOOLEAN_OBJ, args[6].Type())
			}
			useLines := args[6].(*Boolean).Value
			if useLines {
				rl.DrawPolyLines(center.Value, int32(args[1].(*Integer).Value), float32(args[2].(*Float).Value), float32(args[3].(*Float).Value), color.Value)
			} else {
				rl.DrawPoly(center.Value, int32(args[1].(*Integer).Value), float32(args[2].(*Float).Value), float32(args[3].(*Float).Value), color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_poly` draws a polygon" +
				"// Draw a regular polygon (Vector version)\n" +
				"void DrawPoly(Vector2 center, int sides, float radius, float rotation, Color color);\n" +
				"// Draw a polygon outline of n sides\n" +
				"void DrawPolyLines(Vector2 center, int sides, float radius, float rotation, Color color);\n" +
				"// Draw a polygon outline of n sides with extended parameters\n" +
				"void DrawPolyLinesEx(Vector2 center, int sides, float radius, float rotation, float lineThick, Color color);",
			signature: "draw_poly() (see explanation for some signatures) -> null",
			errors:    "InvalidArgCount,PositionalType",
			example:   "draw_poly() (see explanation for some signatures)=> null",
		}.String(),
	},
	{
		Name: "_draw_line_3d",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("draw_line_3d", len(args), 3, "")
			}
			startPos, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_line_3d", 1, "rl.Vector3", args[0])
			}
			endPos, ok := args[1].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_line_3d", 2, "rl.Vector3", args[1])
			}
			color, ok := args[2].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_line_3d", 3, "rl.Color", args[2])
			}
			rl.DrawLine3D(startPos.Value, endPos.Value, color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_line_3d` draws a line in 3d world space",
			signature: "draw_line_3d(start_pos: vector3, end_pos: vector3, color: color)\n" +
				"// Draw a line in 3D world space\n" +
				"void DrawLine3D(Vector3 startPos, Vector3 endPos, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_line_3d()",
		}.String(),
	},
	{
		Name: "_draw_point_3d",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("draw_point_3d", len(args), 2, "")
			}
			pos, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_point_3d", 1, "rl.Vector3", args[0])
			}
			color, ok := args[1].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_point_3d", 2, "rl.Color", args[1])
			}
			rl.DrawPoint3D(pos.Value, color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_point_3d` draws a point in 3d world space",
			signature: "draw_point_3d(pos: vector3, color: color)\n" +
				"// Draw a point in 3D space, actually a small line\n" +
				"void DrawPoint3D(Vector3 position, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_point_3d()",
		}.String(),
	},
	{
		Name: "_draw_circle_3d",
		Fun: func(args ...Object) Object {
			if len(args) != 5 {
				return newInvalidArgCountError("draw_circle_3d", len(args), 5, "")
			}
			center, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_circle_3d", 1, "rl.Vector3", args[0])
			}
			radius, ok := args[1].(*Float)
			if !ok {
				return newPositionalTypeError("draw_circle_3d", 2, FLOAT_OBJ, args[1].Type())
			}
			rotationAxis, ok := args[2].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_circle_3d", 3, "rl.Vector3", args[2])
			}
			rotationAngle, ok := args[3].(*Float)
			if !ok {
				return newPositionalTypeError("draw_circle_3d", 4, FLOAT_OBJ, args[3].Type())
			}
			color, ok := args[4].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_circle_3d", 5, "rl.Color", args[4])
			}
			rl.DrawCircle3D(center.Value, float32(radius.Value), rotationAxis.Value, float32(rotationAngle.Value), color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_circle_3d` draws a circle in 3d world space",
			signature: "draw_circle_3d(center: vector3, radius: float, rotation_axis: vector3, rotation_angle: float, color: color)\n" +
				"// Draw a circle in 3D world space\n" +
				"void DrawCircle3D(Vector3 center, float radius, Vector3 rotationAxis, float rotationAngle, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_circle_3d()",
		}.String(),
	},
	{
		Name: "_draw_cube_wires",
		Fun: func(args ...Object) Object {
			if len(args) != 3 && len(args) != 5 {
				return newInvalidArgCountError("draw_cube_wires", len(args), 3, "or 5")
			}
			if len(args) == 3 {
				position, ok := args[0].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube_wires", 1, "rl.Vector3", args[0])
				}
				size, ok := args[1].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube_wires", 2, "rl.Vector3", args[1])
				}
				color, ok := args[2].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube_wires", 3, "rl.Color", args[2])
				}
				rl.DrawCubeWiresV(position.Value, size.Value, color.Value)
			} else {
				position, ok := args[0].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube_wires", 1, "rl.Vector3", args[0])
				}
				width, ok := args[1].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cube_wires", 2, FLOAT_OBJ, args[1].Type())
				}
				height, ok := args[2].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cube_wires", 3, FLOAT_OBJ, args[2].Type())
				}
				length, ok := args[3].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cube_wires", 4, FLOAT_OBJ, args[3].Type())
				}
				color, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube_wires", 5, "rl.Color", args[4])
				}
				rl.DrawCubeWires(position.Value, float32(width.Value), float32(height.Value), float32(length.Value), color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_cube_wires` draws cube with wires",
			signature: "draw_cube_wires(pos: vector3, width: float, height: float, length: float, color: color)\n" +
				"// Draw cube wires\n" +
				"void DrawCubeWires(Vector3 position, float width, float height, float length, Color color);\n" +
				"// Draw cube wires (Vector version)\n" +
				"void DrawCubeWiresV(Vector3 position, Vector3 size, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_cube_wires()",
		}.String(),
	},
	{
		Name: "_draw_cube",
		Fun: func(args ...Object) Object {
			if len(args) != 3 && len(args) != 5 {
				return newInvalidArgCountError("draw_cube", len(args), 3, "or 5")
			}
			if len(args) == 3 {
				position, ok := args[0].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube", 1, "rl.Vector3", args[0])
				}
				size, ok := args[1].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube", 2, "rl.Vector3", args[1])
				}
				color, ok := args[2].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube", 3, "rl.Color", args[2])
				}
				rl.DrawCubeV(position.Value, size.Value, color.Value)
			} else {
				position, ok := args[0].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube", 1, "rl.Vector3", args[0])
				}
				width, ok := args[1].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cube", 2, FLOAT_OBJ, args[1].Type())
				}
				height, ok := args[2].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cube", 3, FLOAT_OBJ, args[2].Type())
				}
				length, ok := args[3].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cube", 4, FLOAT_OBJ, args[3].Type())
				}
				color, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cube", 5, "rl.Color", args[4])
				}
				rl.DrawCube(position.Value, float32(width.Value), float32(height.Value), float32(length.Value), color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_cube` draws cube",
			signature: "draw_cube(pos: vector3, width: float, height: float, length: float, color: color)\n" +
				"// Draw cube\n" +
				"void DrawCube(Vector3 position, float width, float height, float length, Color color);\n" +
				"// Draw cube (Vector version)\n" +
				"void DrawCubeV(Vector3 position, Vector3 size, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_cube()",
		}.String(),
	},
	{
		Name: "_draw_sphere_wires",
		Fun: func(args ...Object) Object {
			if len(args) != 5 {
				return newInvalidArgCountError("draw_sphere_wires", len(args), 5, "")
			}
			centerPos, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_sphere_wires", 1, "rl.Vector3", args[0])
			}
			radius, ok := args[1].(*Float)
			if !ok {
				return newPositionalTypeError("draw_sphere_wires", 2, FLOAT_OBJ, args[1].Type())
			}
			rings, ok := args[2].(*Integer)
			if !ok {
				return newPositionalTypeError("draw_sphere_wires", 3, INTEGER_OBJ, args[2].Type())
			}
			slices, ok := args[3].(*Integer)
			if !ok {
				return newPositionalTypeError("draw_sphere_wires", 4, INTEGER_OBJ, args[3].Type())
			}
			color, ok := args[4].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_sphere_wires", 5, "rl.Vector3", args[4])
			}
			rl.DrawSphereWires(centerPos.Value, float32(radius.Value), int32(rings.Value), int32(slices.Value), color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_sphere_wires` draws sphere with wires",
			signature: "draw_sphere_wires(center_pos: vector3, radius: float, rings: int, slices: int32, color: color)\n" +
				"// Draw sphere wires\n" +
				"void DrawSphereWires(Vector3 centerPos, float radius, int rings, int slices, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_sphere_wires()",
		}.String(),
	},
	{
		Name: "_draw_sphere",
		Fun: func(args ...Object) Object {
			if len(args) != 3 && len(args) != 5 {
				return newInvalidArgCountError("draw_sphere", len(args), 3, "or 5")
			}
			if len(args) == 3 {
				centerPos, ok := args[0].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_sphere", 1, "rl.Vector3", args[0])
				}
				radius, ok := args[1].(*Float)
				if !ok {
					return newPositionalTypeError("draw_sphere", 2, FLOAT_OBJ, args[1].Type())
				}
				color, ok := args[2].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_sphere", 3, "rl.Color", args[2])
				}
				rl.DrawSphere(centerPos.Value, float32(radius.Value), color.Value)
			} else {
				centerPos, ok := args[0].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_sphere", 1, "rl.Vector3", args[0])
				}
				radius, ok := args[1].(*Float)
				if !ok {
					return newPositionalTypeError("draw_sphere", 2, FLOAT_OBJ, args[1].Type())
				}
				rings, ok := args[2].(*Integer)
				if !ok {
					return newPositionalTypeError("draw_sphere", 3, INTEGER_OBJ, args[2].Type())
				}
				slices, ok := args[3].(*Integer)
				if !ok {
					return newPositionalTypeError("draw_sphere", 4, INTEGER_OBJ, args[3].Type())
				}
				color, ok := args[4].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_sphere", 5, "rl.Color", args[4])
				}
				rl.DrawSphereEx(centerPos.Value, float32(radius.Value), int32(rings.Value), int32(slices.Value), color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_sphere` draws sphere",
			signature: "draw_sphere(center_pos: vector3, radius: float, color: color)\n" +
				"// Draw sphere\n" +
				"void DrawSphere(Vector3 centerPos, float radius, Color color);\n" +
				"// Draw sphere with extended parameters\n" +
				"void DrawSphereEx(Vector3 centerPos, float radius, int rings, int slices, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_sphere()",
		}.String(),
	},
	{
		Name: "_draw_cylinder_wires",
		Fun: func(args ...Object) Object {
			if len(args) != 6 {
				return newInvalidArgCountError("draw_cylinder_wires", len(args), 6, "")
			}
			position, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_cylinder_wires", 1, "rl.Vector3", args[0])
			}
			if args[1].Type() == FLOAT_OBJ {
				radiusTop := args[1].(*Float)
				radiusBot, ok := args[2].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cylinder_wires", 3, FLOAT_OBJ, args[2].Type())
				}
				height, ok := args[3].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cylinder_wires", 4, FLOAT_OBJ, args[3].Type())
				}
				slices, ok := args[4].(*Integer)
				if !ok {
					return newPositionalTypeError("draw_cylinder_wires", 5, INTEGER_OBJ, args[4].Type())
				}
				color, ok := args[5].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cylinder_wires", 6, "rl.Color", args[5])
				}
				rl.DrawCylinderWires(position.Value, float32(radiusTop.Value), float32(radiusBot.Value), float32(height.Value), int32(slices.Value), color.Value)
			} else {
				endPos, ok := args[1].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cylinder_wires", 2, "rl.Vector3", args[1])
				}
				startRadius, ok := args[2].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cylinder_wires", 3, FLOAT_OBJ, args[2].Type())
				}
				endRadius, ok := args[3].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cylinder_wires", 4, FLOAT_OBJ, args[3].Type())
				}
				sides, ok := args[4].(*Integer)
				if !ok {
					return newPositionalTypeError("draw_cylinder_wires", 5, INTEGER_OBJ, args[4].Type())
				}
				color, ok := args[5].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cylinder_wires", 6, "rl.Color", args[5])
				}
				rl.DrawCylinderWiresEx(position.Value, endPos.Value, float32(startRadius.Value), float32(endRadius.Value), int32(sides.Value), color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_cylinder_wires` draws cylinder with wires",
			signature: "draw_cylinder_wires(pos: vector3, radius_top: float, radius_bot: float, height: float, slices_sides: int, color: color)\n" +
				"// Draw a cylinder/cone wires\n" +
				"void DrawCylinderWires(Vector3 position, float radiusTop, float radiusBottom, float height, int slices, Color color);\n" +
				"// Draw a cylinder wires with base at startPos and top at endPos\n" +
				"void DrawCylinderWiresEx(Vector3 startPos, Vector3 endPos, float startRadius, float endRadius, int sides, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_cylinder_wires()",
		}.String(),
	},
	{
		Name: "_draw_cylinder",
		Fun: func(args ...Object) Object {
			if len(args) != 6 {
				return newInvalidArgCountError("draw_cylinder", len(args), 6, "")
			}
			position, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_cylinder", 1, "rl.Vector3", args[0])
			}
			if args[1].Type() == FLOAT_OBJ {
				radiusTop := args[1].(*Float)
				radiusBot, ok := args[2].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cylinder", 3, FLOAT_OBJ, args[2].Type())
				}
				height, ok := args[3].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cylinder", 4, FLOAT_OBJ, args[3].Type())
				}
				slices, ok := args[4].(*Integer)
				if !ok {
					return newPositionalTypeError("draw_cylinder", 5, INTEGER_OBJ, args[4].Type())
				}
				color, ok := args[5].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cylinder", 6, "rl.Color", args[5])
				}
				rl.DrawCylinder(position.Value, float32(radiusTop.Value), float32(radiusBot.Value), float32(height.Value), int32(slices.Value), color.Value)
			} else {
				endPos, ok := args[1].(*GoObj[rl.Vector3])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cylinder", 2, "rl.Vector3", args[1])
				}
				startRadius, ok := args[2].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cylinder", 3, FLOAT_OBJ, args[2].Type())
				}
				endRadius, ok := args[3].(*Float)
				if !ok {
					return newPositionalTypeError("draw_cylinder", 4, FLOAT_OBJ, args[3].Type())
				}
				sides, ok := args[4].(*Integer)
				if !ok {
					return newPositionalTypeError("draw_cylinder", 5, INTEGER_OBJ, args[4].Type())
				}
				color, ok := args[5].(*GoObj[rl.Color])
				if !ok {
					return newPositionalTypeErrorForGoObj("draw_cylinder", 6, "rl.Color", args[5])
				}
				rl.DrawCylinderEx(position.Value, endPos.Value, float32(startRadius.Value), float32(endRadius.Value), int32(sides.Value), color.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_cylinder` draws cylinder",
			signature: "draw_cylinder(pos: vector3, radius_top: float, radius_bot: float, height: float, slices_sides: int, color: color)\n" +
				"// Draw a cylinder/cone\n" +
				"void DrawCylinder(Vector3 position, float radiusTop, float radiusBottom, float height, int slices, Color color);\n" +
				"// Draw a cylinder with base at startPos and top at endPos\n" +
				"void DrawCylinderEx(Vector3 startPos, Vector3 endPos, float startRadius, float endRadius, int sides, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_cylinder()",
		}.String(),
	},
	{
		Name: "_draw_capsule_wires",
		Fun: func(args ...Object) Object {
			if len(args) != 6 {
				return newInvalidArgCountError("draw_capsule_wires", len(args), 6, "")
			}
			startPos, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_capsule_wires", 1, "rl.Vector3", args[0])
			}
			endPos, ok := args[1].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_capsule_wires", 2, "rl.Vector3", args[1])
			}
			radius, ok := args[2].(*Float)
			if !ok {
				return newPositionalTypeError("draw_capsule_wires", 3, FLOAT_OBJ, args[2].Type())
			}
			slices, ok := args[3].(*Integer)
			if !ok {
				return newPositionalTypeError("draw_capsule_wires", 4, INTEGER_OBJ, args[3].Type())
			}
			rings, ok := args[4].(*Integer)
			if !ok {
				return newPositionalTypeError("draw_capsule_wires", 5, INTEGER_OBJ, args[4].Type())
			}
			color, ok := args[5].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_capsule_wires", 6, "rl.Color", args[5])
			}
			rl.DrawCapsuleWires(startPos.Value, endPos.Value, float32(radius.Value), int32(slices.Value), int32(rings.Value), color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_capsule_wires` draws capsule with wires",
			signature: "draw_capsule_wires(start_pos: vector3, end_pos: vector3, radius: float, slices: int, rings: int, color: color)\n" +
				"// Draw capsule wireframe with the center of its sphere caps at startPos and endPos\n" +
				"void DrawCapsuleWires(Vector3 startPos, Vector3 endPos, float radius, int slices, int rings, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_capsule_wires()",
		}.String(),
	},
	{
		Name: "_draw_capsule",
		Fun: func(args ...Object) Object {
			if len(args) != 6 {
				return newInvalidArgCountError("draw_capsule", len(args), 6, "")
			}
			startPos, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_capsule", 1, "rl.Vector3", args[0])
			}
			endPos, ok := args[1].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_capsule", 2, "rl.Vector3", args[1])
			}
			radius, ok := args[2].(*Float)
			if !ok {
				return newPositionalTypeError("draw_capsule", 3, FLOAT_OBJ, args[2].Type())
			}
			slices, ok := args[3].(*Integer)
			if !ok {
				return newPositionalTypeError("draw_capsule", 4, INTEGER_OBJ, args[3].Type())
			}
			rings, ok := args[4].(*Integer)
			if !ok {
				return newPositionalTypeError("draw_capsule", 5, INTEGER_OBJ, args[4].Type())
			}
			color, ok := args[5].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_capsule", 6, "rl.Color", args[5])
			}
			rl.DrawCapsule(startPos.Value, endPos.Value, float32(radius.Value), int32(slices.Value), int32(rings.Value), color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_capsule` draws capsule",
			signature: "draw_capsule(start_pos: vector3, end_pos: vector3, radius: float, slices: int, rings: int, color: color)\n" +
				"// Draw a capsule with the center of its sphere caps at startPos and endPos\n" +
				"void DrawCapsule(Vector3 startPos, Vector3 endPos, float radius, int slices, int rings, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_capsule()",
		}.String(),
	},
	{
		Name: "_draw_plane",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("draw_plane", len(args), 3, "")
			}
			position, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_plane", 1, "rl.Vector3", args[0])
			}
			size, ok := args[1].(*GoObj[rl.Vector2])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_plane", 2, "rl.Vector2", args[1])
			}
			color, ok := args[2].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_plane", 3, "rl.Color", args[2])
			}
			rl.DrawPlane(position.Value, size.Value, color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_plane` draws a plane XZ",
			signature: "draw_plane(center_pos: vector3, size: vector2, color: color)\n" +
				"// Draw a plane XZ\n" +
				"void DrawPlane(Vector3 centerPos, Vector2 size, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_plane()",
		}.String(),
	},
	{
		Name: "_draw_ray",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("draw_ray", len(args), 2, "")
			}
			ray, ok := args[0].(*GoObj[rl.Ray])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_ray", 1, "rl.Ray", args[0])
			}
			color, ok := args[1].(*GoObj[rl.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("draw_ray", 2, "rl.Color", args[1])
			}
			rl.DrawRay(ray.Value, color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_ray` draws a ray line",
			signature: "draw_ray(ray: ray, color: color)\n" +
				"// Draw a ray line\n" +
				"void DrawRay(Ray ray, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_ray()",
		}.String(),
	},
	{
		Name: "_draw_grid",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("draw_grid", len(args), 2, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("draw_grid", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("draw_grid", 2, FLOAT_OBJ, args[1].Type())
			}
			rl.DrawGrid(int32(args[0].(*Integer).Value), float32(args[1].(*Float).Value))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_grid` draws a grid centered at 0,0,0",
			signature: "draw_grid(slices: int, spacing: float)\n" +
				"// Draw a grid (centered at (0, 0, 0))\n" +
				"void DrawGrid(int slices, float spacing);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_grid()",
		}.String(),
	},
	{
		Name: "_set_target_fps",
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
	{
		Name: "_get_fps",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_fps", len(args), 0, "")
			}
			return &Integer{Value: int64(rl.GetFPS())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_fps` returns current fps as an integer",
			signature:   "get_fps() -> int",
			errors:      "InvalidArgCount",
			example:     "get_fps() => 60",
		}.String(),
	},
	{
		Name: "_get_frame_time",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_frame_time", len(args), 0, "")
			}
			return &Float{Value: float64(rl.GetFrameTime())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_frame_time` returns the time in seconds (float) for the last frame drawn (delta time)",
			signature:   "get_frame_time() -> float",
			errors:      "InvalidArgCount",
			example:     "get_frame_time() => 16.67",
		}.String(),
	},
	{
		Name: "_get_time",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_time", len(args), 0, "")
			}
			return &Float{Value: rl.GetTime()}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_time` returns the time elapsed (float) since init_window was called",
			signature:   "get_time() -> float",
			errors:      "InvalidArgCount",
			example:     "get_time() => 100.1",
		}.String(),
	},
	{
		Name: "_set_exit_key",
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
	{
		Name: "_is_key_up",
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
	{
		Name: "_is_key_down",
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
	{
		Name: "_is_key_pressed",
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
	{
		Name: "_is_key_released",
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
	{
		Name: "_is_mouse_button_pressed",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_mouse_button_pressed", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("is_mouse_button_pressed", 1, INTEGER_OBJ, args[0].Type())
			}
			return nativeToBooleanObject(rl.IsMouseButtonPressed(int32(args[0].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_mouse_button_pressed` returns true if mouse button has been pressed once",
			signature:   "is_mouse_button_pressed(button: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_mouse_button_pressed(0) => true",
		}.String(),
	},
	{
		Name: "_is_mouse_button_down",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_mouse_button_down", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("is_mouse_button_down", 1, INTEGER_OBJ, args[0].Type())
			}
			return nativeToBooleanObject(rl.IsMouseButtonDown(int32(args[0].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_mouse_button_down` returns true if mouse button is being pressed",
			signature:   "is_mouse_button_down(button: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_mouse_button_down(0) => true",
		}.String(),
	},
	{
		Name: "_is_mouse_button_released",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_mouse_button_released", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("is_mouse_button_released", 1, INTEGER_OBJ, args[0].Type())
			}
			return nativeToBooleanObject(rl.IsMouseButtonReleased(int32(args[0].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_mouse_button_released` returns true if mouse button has been released once",
			signature:   "is_mouse_button_released(button: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_mouse_button_released(0) => true",
		}.String(),
	},
	{
		Name: "_is_mouse_button_up",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_mouse_button_up", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("is_mouse_button_up", 1, INTEGER_OBJ, args[0].Type())
			}
			return nativeToBooleanObject(rl.IsMouseButtonUp(int32(args[0].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_mouse_button_up` returns true is mouse button is not being pressed",
			signature:   "is_mouse_button_up(button: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_mouse_button_up(0) => true",
		}.String(),
	},
	{
		Name: "_get_mouse_x",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_mouse_x", len(args), 0, "")
			}
			return &Integer{Value: int64(rl.GetMouseX())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_mouse_x` returns mouse x position as int",
			signature:   "get_mouse_x() -> int",
			errors:      "InvalidArgCount",
			example:     "get_mouse_x() => 200",
		}.String(),
	},
	{
		Name: "_get_mouse_y",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_mouse_y", len(args), 0, "")
			}
			return &Integer{Value: int64(rl.GetMouseY())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_mouse_y` returns mouse y position as int",
			signature:   "get_mouse_y() -> int",
			errors:      "InvalidArgCount",
			example:     "get_mouse_y() => 200",
		}.String(),
	},
	{
		Name: "_get_mouse_position",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_mouse_position", len(args), 0, "")
			}
			return NewGoObj(rl.GetMousePosition())
		},
		HelpStr: helpStrArgs{
			explanation: "`get_mouse_position` returns mouse position as GoObj[rl.Vector2] (x,y)",
			signature:   "get_mouse_position() -> GoObj[rl.Vector2]",
			errors:      "InvalidArgCount",
			example:     "get_mouse_position() => GoObj[rl.Vector2]",
		}.String(),
	},
	{
		Name: "_get_mouse_delta",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_mouse_delta", len(args), 0, "")
			}
			return NewGoObj(rl.GetMouseDelta())
		},
		HelpStr: helpStrArgs{
			explanation: "`get_mouse_delta` returns mouse delta between frames as GoObj[rl.Vector2]",
			signature:   "get_mouse_delta() -> GoObj[rl.Vector2]",
			errors:      "InvalidArgCount",
			example:     "get_mouse_delta() => GoObj[rl.Vector2]",
		}.String(),
	},
	{
		Name: "_set_mouse_position",
		Fun: func(args ...Object) Object {
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("set_mouse_position", len(args), 1, "or 2")
			}
			if len(args) == 0 {
				position, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("set_mouse_position", 1, "rl.Vector2", args[0])
				}
				rl.SetMousePosition(int(position.Value.X), int(position.Value.Y))
				return NULL
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_mouse_position", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_mouse_position", 2, INTEGER_OBJ, args[1].Type())
			}
			rl.SetMousePosition(int(args[0].(*Integer).Value), int(args[1].(*Integer).Value))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_mouse_position` sets the x,y mouse position. A GoObj[rl.Vector2] can be used but float precision is not kept",
			signature:   "set_mouse_position(position_or_x: GoObj[rl.Vector2]|int, y: int=null) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_mouse_position(100, 100) => null",
		}.String(),
	},
	{
		Name: "_set_mouse_offset",
		Fun: func(args ...Object) Object {
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("set_mouse_offset", len(args), 1, "or 2")
			}
			if len(args) == 0 {
				position, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("set_mouse_offset", 1, "rl.Vector2", args[0])
				}
				rl.SetMouseOffset(int(position.Value.X), int(position.Value.Y))
				return NULL
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_mouse_offset", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_mouse_offset", 2, INTEGER_OBJ, args[1].Type())
			}
			rl.SetMouseOffset(int(args[0].(*Integer).Value), int(args[1].(*Integer).Value))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_mouse_offset` sets the x,y mouse offset. A GoObj[rl.Vector2] can be used but float precision is not kept",
			signature:   "set_mouse_offset(offset_pos_or_offset_x: GoObj[rl.Vector2]|int, offset_y: int=null) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_mouse_offset(100, 100) => null",
		}.String(),
	},
	{
		Name: "_set_mouse_scale",
		Fun: func(args ...Object) Object {
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("set_mouse_scale", len(args), 1, "or 2")
			}
			if len(args) == 0 {
				position, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("set_mouse_scale", 1, "rl.Vector2", args[0])
				}
				rl.SetMouseScale(position.Value.X, position.Value.Y)
				return NULL
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("set_mouse_scale", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("set_mouse_scale", 2, FLOAT_OBJ, args[1].Type())
			}
			rl.SetMouseScale(float32(args[0].(*Float).Value), float32(args[1].(*Float).Value))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_mouse_scale` sets the x,y mouse scaling. A GoObj[rl.Vector2] can be used",
			signature:   "set_mouse_scale(scale_or_scale_x: GoObj[rl.Vector2]|float, scale_y: float=null) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_mouse_scale(1.0, 1.0) => null",
		}.String(),
	},
	{
		Name: "_get_mouse_wheel_move",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_mouse_wheel_move", len(args), 0, "")
			}
			return &Float{Value: float64(rl.GetMouseWheelMove())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_mouse_wheel_move` returns the mouse wheel movement for x or y, whichever is larger, as a float",
			signature:   "get_mouse_wheel_move() -> float",
			errors:      "InvalidArgCount",
			example:     "get_mouse_wheel_move() => 1.0",
		}.String(),
	},
	{
		Name: "_get_mouse_wheel_move_v",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_mouse_wheel_move_v", len(args), 0, "")
			}
			return NewGoObj(rl.GetMouseWheelMoveV())
		},
		HelpStr: helpStrArgs{
			explanation: "`get_mouse_wheel_move_v` returns the mouse wheel movement for x and y as GoObj[rl.Vector2]",
			signature:   "get_mouse_wheel_move_v() -> GoObj[rl.Vector2]",
			errors:      "InvalidArgCount",
			example:     "get_mouse_wheel_move_v() => GoObj[rl.Vector2]",
		}.String(),
	},
	{
		Name: "_set_mouse_cursor",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_mouse_cursor", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("set_mouse_cursor", 1, INTEGER_OBJ, args[0].Type())
			}
			rl.SetMouseCursor(int32(args[0].(*Integer).Value))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_mouse_cursor` sets the mouse cursor",
			signature:   "set_mouse_cursor(cursor: int) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_mouse_cursor(1) => null",
		}.String(),
	},
	{
		Name: "_load_texture",
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
	{
		Name: "_rectangle",
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
	{
		Name: "_rectangle_check_collision",
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
	{
		Name: "_check_collision",
		Fun: func(args ...Object) Object {
			argLen := len(args)
			if argLen != 4 && argLen != 3 && argLen != 2 {
				return newInvalidArgCountError("check_collision", argLen, 2, "or 3 or 4")
			}
			switch argLen {
			case 2:
				if point, ok := args[0].(*GoObj[rl.Vector2]); ok {
					rec, ok := args[1].(*GoObj[rl.Rectangle])
					if !ok {
						return newPositionalTypeErrorForGoObj("check_collision", 2, "rl.Rectangle", args[1])
					}
					return nativeToBooleanObject(rl.CheckCollisionPointRec(point.Value, rec.Value))
				} else if rec, ok := args[0].(*GoObj[rl.Rectangle]); ok {
					rec1, ok := args[1].(*GoObj[rl.Rectangle])
					if !ok {
						return newPositionalTypeErrorForGoObj("check_collision", 2, "rl.Rectangle", args[1])
					}
					return nativeToBooleanObject(rl.CheckCollisionRecs(rec.Value, rec1.Value))
				} else if bb, ok := args[0].(*GoObj[rl.BoundingBox]); ok {
					bb1, ok := args[1].(*GoObj[rl.BoundingBox])
					if !ok {
						return newPositionalTypeErrorForGoObj("check_collision", 2, "rl.BoundingBox", args[1])
					}
					return nativeToBooleanObject(rl.CheckCollisionBoxes(bb.Value, bb1.Value))
				} else {
					return newPositionalTypeErrorForGoObj("check_collision", 1, "rl.Vector2 or rl.Rectangle or rl.BoundingBox", args[0])
				}
			case 3:
				if pointOrCenter, ok := args[0].(*GoObj[rl.Vector2]); ok {
					if args[1].Type() == FLOAT_OBJ {
						rec, ok := args[2].(*GoObj[rl.Rectangle])
						if !ok {
							return newPositionalTypeErrorForGoObj("check_collision", 3, "rl.Rectangle", args[2])
						}
						return nativeToBooleanObject(rl.CheckCollisionCircleRec(pointOrCenter.Value, float32(args[1].(*Float).Value), rec.Value))
					} else if args[1].Type() == LIST_OBJ {
						l := args[1].(*List).Elements
						points := make([]rl.Vector2, len(l))
						for i, e := range l {
							point, ok := e.(*GoObj[rl.Vector2])
							if !ok {
								return newPositionalTypeErrorForGoObj("check_collision", 2, "list[rl.Vector2]", e)
							}
							points[i] = point.Value
						}
						if args[2].Type() != INTEGER_OBJ {
							return newPositionalTypeError("check_collision", 3, INTEGER_OBJ, args[2].Type())
						}
						return nativeToBooleanObject(rl.CheckCollisionPointPoly(pointOrCenter.Value, points, int32(args[2].(*Integer).Value)))
					} else if args[1].Type() == GO_OBJ {
						center, ok := args[1].(*GoObj[rl.Vector2])
						if !ok {
							return newPositionalTypeErrorForGoObj("check_collision", 2, "rl.Vector2", args[1])
						}
						if args[2].Type() != FLOAT_OBJ {
							return newPositionalTypeError("check_collision", 3, FLOAT_OBJ, args[2].Type())
						}
						return nativeToBooleanObject(rl.CheckCollisionPointCircle(pointOrCenter.Value, center.Value, float32(args[2].(*Float).Value)))
					} else {
						return newPositionalTypeErrorForGoObj("check_collision", 2, "float or rl.Vector2 or list[rl.Vector2]", args[1])
					}
				} else if bb, ok := args[0].(*GoObj[rl.BoundingBox]); ok {
					center, ok := args[1].(*GoObj[rl.Vector3])
					if !ok {
						return newPositionalTypeErrorForGoObj("check_collision", 2, "rl.Vector3", args[1])
					}
					err := checkArgType("check_collision", 3, FLOAT_OBJ, args)
					if err != nil {
						return err
					}
					return nativeToBooleanObject(rl.CheckCollisionBoxSphere(bb.Value, center.Value, float32(args[2].(*Float).Value)))
				} else {
					return newPositionalTypeErrorForGoObj("check_collision", 1, "rl.Vector2 or rl.BoundingBox", args[0])
				}
			case 4:
				if pointOrCenter, ok := args[0].(*GoObj[rl.Vector2]); ok {
					if args[1].Type() == FLOAT_OBJ {
						center, ok := args[2].(*GoObj[rl.Vector2])
						if !ok {
							return newPositionalTypeErrorForGoObj("check_collision", 3, "rl.Vector2", args[2])
						}
						if args[3].Type() != FLOAT_OBJ {
							return newPositionalTypeError("check_collision", 4, FLOAT_OBJ, args[3].Type())
						}
						return nativeToBooleanObject(rl.CheckCollisionCircles(pointOrCenter.Value, float32(args[1].(*Float).Value), center.Value, float32(args[3].(*Float).Value)))
					} else {
						point2, ok := args[1].(*GoObj[rl.Vector2])
						if !ok {
							return newPositionalTypeErrorForGoObj("check_collision", 2, "rl.Vector2", args[1])
						}
						point3, ok := args[2].(*GoObj[rl.Vector2])
						if !ok {
							return newPositionalTypeErrorForGoObj("check_collision", 3, "rl.Vector2", args[2])
						}
						if point4, ok := args[3].(*GoObj[rl.Vector2]); ok {
							return nativeToBooleanObject(rl.CheckCollisionPointTriangle(pointOrCenter.Value, point2.Value, point3.Value, point4.Value))
						} else if args[3].Type() != INTEGER_OBJ {
							return newPositionalTypeError("check_collision", 4, INTEGER_OBJ, args[3].Type())
						}
						return nativeToBooleanObject(rl.CheckCollisionPointLine(pointOrCenter.Value, point2.Value, point3.Value, int32(args[3].(*Integer).Value)))
					}
				} else if center1, ok := args[0].(*GoObj[rl.Vector3]); ok {
					err := checkArgType("check_collision", 2, FLOAT_OBJ, args)
					if err != nil {
						return err
					}
					center2, err := checkGoObjType[rl.Vector3]("check_collision", 3, "rl.Vector3", args)
					if err != nil {
						return err
					}
					err = checkArgType("check_collision", 4, FLOAT_OBJ, args)
					if err != nil {
						return err
					}
					return nativeToBooleanObject(rl.CheckCollisionSpheres(center1.Value, float32(args[1].(*Float).Value), center2.Value, float32(args[3].(*Float).Value)))
				} else {
					return newPositionalTypeErrorForGoObj("check_collision", 1, "rl.Vector2", args[0])
				}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`check_collision` returns true if the objects collide",
			signature: "check_collision() -> bool\n" +
				"// Check collision between two rectangles\n" +
				"bool CheckCollisionRecs(Rectangle rec1, Rectangle rec2);\n" +
				"// Check collision between two circles\n" +
				"bool CheckCollisionCircles(Vector2 center1, float radius1, Vector2 center2, float radius2);\n" +
				"// Check collision between circle and rectangle\n" +
				"bool CheckCollisionCircleRec(Vector2 center, float radius, Rectangle rec);\n" +
				"// Check if circle collides with a line created betweeen two points [p1] and [p2\n" +
				"bool CheckCollisionCircleLine(Vector2 center, float radius, Vector2 p1, Vector2 p2);\n" +
				"// Check if point is inside rectangle\n" +
				"bool CheckCollisionPointRec(Vector2 point, Rectangle rec);\n" +
				"// Check if point is inside circle\n" +
				"bool CheckCollisionPointCircle(Vector2 point, Vector2 center, float radius);\n" +
				"// Check if point belongs to line created between two points [p1] and [p2] with\n" +
				"bool CheckCollisionPointLine(Vector2 point, Vector2 p1, Vector2 p2, int threshold);\n" +
				"// Check if point is within a polygon described by array of vertices\n" +
				"bool CheckCollisionPointPoly(Vector2 point, const Vector2 *points, int pointCount);\n" +
				"// Check collision between two spheres\n" +
				"bool CheckCollisionSpheres(Vector3 center1, float radius1, Vector3 center2, float radius2);\n" +
				"// Check collision between two bounding boxes\n" +
				"bool CheckCollisionBoxes(BoundingBox box1, BoundingBox box2);\n" +
				"// Check collision between box and sphere\n" +
				"bool CheckCollisionBoxSphere(BoundingBox box, Vector3 center, float radius);",
			errors:  "InvalidArgCount,PositionalType",
			example: "check_collision() => (see signature for examples)=>true",
		}.String(),
	},
	{
		Name: "_get_collision",
		Fun: func(args ...Object) Object {
			err := checkArgsCount("get_collision", []int{2, 3, 4, 5}, args)
			if err != nil {
				return err
			}
			argLen := len(args)
			if rec1, ok := args[0].(*GoObj[rl.Rectangle]); ok {
				rec2, ok := args[1].(*GoObj[rl.Rectangle])
				if !ok {
					return newPositionalTypeErrorForGoObj("get_collision", 2, "rl.Rectangle", args[1])
				}
				return NewGoObj(rl.GetCollisionRec(rec1.Value, rec2.Value))
			} else if ray, ok := args[0].(*GoObj[rl.Ray]); ok {
				switch argLen {
				case 2:
					bb, err := checkGoObjType[rl.BoundingBox]("get_collision", 2, "rl.BoundingBox", args)
					if err != nil {
						return err
					}
					return NewGoObj(rl.GetRayCollisionBox(ray.Value, bb.Value))
				case 3:
					if center, ok := args[1].(*GoObj[rl.Vector3]); ok {
						err = checkArgType("get_collision", 3, FLOAT_OBJ, args)
						if err != nil {
							return err
						}
						return NewGoObj(rl.GetRayCollisionSphere(ray.Value, center.Value, float32(args[2].(*Float).Value)))
					} else if mesh, ok := args[1].(*GoObj[rl.Mesh]); ok {
						transform, err := checkGoObjType[rl.Matrix]("get_collision", 3, "rl.Matrix", args)
						if err != nil {
							return err
						}
						return NewGoObj(rl.GetRayCollisionMesh(ray.Value, mesh.Value, transform.Value))
					} else {
						return newPositionalTypeErrorForGoObj("get_collision", 2, "rl.Vector3 or rl.Mesh", args[1])
					}
				case 4:
					p1, err := checkGoObjType[rl.Vector3]("get_collision", 2, "rl.Vector3", args)
					if err != nil {
						return err
					}
					p2, err := checkGoObjType[rl.Vector3]("get_collision", 3, "rl.Vector3", args)
					if err != nil {
						return err
					}
					p3, err := checkGoObjType[rl.Vector3]("get_collision", 4, "rl.Vector3", args)
					if err != nil {
						return err
					}
					return NewGoObj(rl.GetRayCollisionTriangle(ray.Value, p1.Value, p2.Value, p3.Value))
				case 5:
					p1, err := checkGoObjType[rl.Vector3]("get_collision", 2, "rl.Vector3", args)
					if err != nil {
						return err
					}
					p2, err := checkGoObjType[rl.Vector3]("get_collision", 3, "rl.Vector3", args)
					if err != nil {
						return err
					}
					p3, err := checkGoObjType[rl.Vector3]("get_collision", 4, "rl.Vector3", args)
					if err != nil {
						return err
					}
					p4, err := checkGoObjType[rl.Vector3]("get_collision", 5, "rl.Vector3", args)
					if err != nil {
						return err
					}
					return NewGoObj(rl.GetRayCollisionQuad(ray.Value, p1.Value, p2.Value, p3.Value, p4.Value))
				}
			} else {
				return newPositionalTypeErrorForGoObj("get_collision", 1, "rl.Rectangle or rl.Ray", args[0])
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`get_collision` returns the collision rectangle or ray collision",
			signature: "get_collision() -> rl.Rectangle|rl.RayCollision\n" +
				"// Get collision rectangle for two rectangles collision\n" +
				"Rectangle GetCollisionRec(Rectangle rec1, Rectangle rec2);\n" +
				"// Get collision info between ray and sphere\n" +
				"RayCollision GetRayCollisionSphere(Ray ray, Vector3 center, float radius);\n" +
				"// Get collision info between ray and box\n" +
				"RayCollision GetRayCollisionBox(Ray ray, BoundingBox box);\n" +
				"// Get collision info between ray and mesh\n" +
				"RayCollision GetRayCollisionMesh(Ray ray, Mesh mesh, Matrix transform);\n" +
				"// Get collision info between ray and triangle\n" +
				"RayCollision GetRayCollisionTriangle(Ray ray, Vector3 p1, Vector3 p2, Vector3 p3);\n" +
				"// Get collision info between ray and quad\n" +
				"RayCollision GetRayCollisionQuad(Ray ray, Vector3 p1, Vector3 p2, Vector3 p3, Vector3 p4);",
			errors:  "InvalidArgCount,PositionalType",
			example: "get_collision() => (see signature for examples)=>rl.Rectangle|rl.RayCollision",
		}.String(),
	},
	{
		Name: "_vector2",
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
	{
		Name: "_vector3",
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
	{
		Name: "_vector4",
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
	{
		Name: "_ray",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("ray", len(args), 2, "")
			}
			pos, ok := args[0].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("ray", 1, "rl.Vector3", args[0])
			}
			dir, ok := args[1].(*GoObj[rl.Vector3])
			if !ok {
				return newPositionalTypeErrorForGoObj("ray", 2, "rl.Vector3", args[1])
			}
			return NewGoObj(rl.NewRay(pos.Value, dir.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`ray` returns a ray with position and direction",
			signature:   "ray(pos: vector3, dir: vector3) -> GoObj[rl.Ray]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "ray() => GoObj[rl.Ray]",
		}.String(),
	},
	{
		Name: "_camera2d",
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
	{
		Name: "_begin_mode2d",
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
	{
		Name: "_end_mode2d",
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
	{
		Name: "_camera3d",
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
	{
		Name: "_begin_mode3d",
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
	{
		Name: "_end_mode3d",
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
	{
		Name: "_matrix",
		Fun: func(args ...Object) Object {
			err := checkArgsCount("matrix", []int{16, 1}, args)
			if err != nil {
				return err
			}
			if len(args) == 1 || args[0].Type() == LIST_OBJ {
				elems := args[0].(*List).Elements
				if len(elems) != 4 {
					return newInvalidArgCountError("matrix", len(elems), 4, "")
				}
				matrix := rl.Matrix{}
				for i, e := range elems {
					if e.Type() != LIST_OBJ {
						return newPositionalTypeError("matrxi", 1, LIST_OBJ, e.Type())
					}
					elems2 := e.(*List).Elements
					if len(elems2) != 4 {
						return newInvalidArgCountError("matrix", len(elems2), 4, "")
					}
					for _, ee := range elems2 {
						if ee.Type() != FLOAT_OBJ {
							return newPositionalTypeError("matrix", 1, FLOAT_OBJ, ee.Type())
						}
					}
					switch i {
					case 0:
						// M0, M1, M2, M3
						matrix.M0 = float32(elems[i].(*List).Elements[0].(*Float).Value)
						matrix.M1 = float32(elems[i].(*List).Elements[1].(*Float).Value)
						matrix.M2 = float32(elems[i].(*List).Elements[2].(*Float).Value)
						matrix.M3 = float32(elems[i].(*List).Elements[3].(*Float).Value)
					case 1:
						// M4, M5, M6, M7
						matrix.M4 = float32(elems[i].(*List).Elements[4].(*Float).Value)
						matrix.M5 = float32(elems[i].(*List).Elements[5].(*Float).Value)
						matrix.M6 = float32(elems[i].(*List).Elements[6].(*Float).Value)
						matrix.M7 = float32(elems[i].(*List).Elements[7].(*Float).Value)
					case 2:
						// M8, M9, M10, M11
						matrix.M8 = float32(elems[i].(*List).Elements[8].(*Float).Value)
						matrix.M9 = float32(elems[i].(*List).Elements[9].(*Float).Value)
						matrix.M10 = float32(elems[i].(*List).Elements[10].(*Float).Value)
						matrix.M11 = float32(elems[i].(*List).Elements[11].(*Float).Value)
					case 3:
						// M12, M13, M14, M15
						matrix.M12 = float32(elems[i].(*List).Elements[12].(*Float).Value)
						matrix.M13 = float32(elems[i].(*List).Elements[13].(*Float).Value)
						matrix.M14 = float32(elems[i].(*List).Elements[14].(*Float).Value)
						matrix.M15 = float32(elems[i].(*List).Elements[15].(*Float).Value)
					}
				}
				return NewGoObj(matrix)
			} else {
				m0, ok := args[0].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 1, FLOAT_OBJ, args[0].Type())
				}
				m4, ok := args[1].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 2, FLOAT_OBJ, args[1].Type())
				}
				m8, ok := args[2].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 3, FLOAT_OBJ, args[2].Type())
				}
				m12, ok := args[3].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 4, FLOAT_OBJ, args[3].Type())
				}
				m1, ok := args[4].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 5, FLOAT_OBJ, args[4].Type())
				}
				m5, ok := args[5].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 6, FLOAT_OBJ, args[5].Type())
				}
				m9, ok := args[6].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 7, FLOAT_OBJ, args[6].Type())
				}
				m13, ok := args[7].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 8, FLOAT_OBJ, args[7].Type())
				}
				m2, ok := args[8].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 9, FLOAT_OBJ, args[8].Type())
				}
				m6, ok := args[9].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 10, FLOAT_OBJ, args[9].Type())
				}
				m10, ok := args[10].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 11, FLOAT_OBJ, args[10].Type())
				}
				m14, ok := args[11].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 12, FLOAT_OBJ, args[11].Type())
				}
				m3, ok := args[12].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 13, FLOAT_OBJ, args[12].Type())
				}
				m7, ok := args[13].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 14, FLOAT_OBJ, args[13].Type())
				}
				m11, ok := args[14].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 15, FLOAT_OBJ, args[14].Type())
				}
				m15, ok := args[15].(*Float)
				if !ok {
					return newPositionalTypeError("matrix", 16, FLOAT_OBJ, args[15].Type())
				}
				return NewGoObj(rl.NewMatrix(float32(m0.Value), float32(m4.Value), float32(m8.Value), float32(m12.Value),
					float32(m1.Value), float32(m5.Value), float32(m9.Value), float32(m13.Value),
					float32(m2.Value), float32(m6.Value), float32(m10.Value), float32(m14.Value),
					float32(m3.Value), float32(m7.Value), float32(m11.Value), float32(m15.Value)))
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`matrix` returns a new matrix",
			signature:   "matrix(m0: float=0.0, m4: float=0.0, m8: float=0.0, m12: float=0.0, m1: float=0.0, m5: float=0.0, m9: float=0.0, m13: float=0.0, m2: float=0.0, m6: float=0.0, m10: float=0.0, m14: float=0.0, m3: float=0.0, m7: float=0.0, m11: float=0.0, m15: float=0.0) -> rl.Matrix",
			errors:      "InvalidArgCount,PositionalType",
			example:     "matrix() => rl.Matrix",
		}.String(),
	},
	{
		Name: "_init_audio_device",
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
	{
		Name: "_close_audio_device",
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
	{
		Name: "_load_music",
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
	{
		Name: "_update_music",
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
	{
		Name: "_play_music",
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
	{
		Name: "_stop_music",
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
	{
		Name: "_resume_music",
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
	{
		Name: "_pause_music",
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
	{
		Name: "_load_sound",
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
	{
		Name: "_play_sound",
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
	{
		Name: "_stop_sound",
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
	{
		Name: "_resume_sound",
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
	{
		Name: "_pause_sound",
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
	{
		Name: "_load_model",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("load_model", 1, args); err != nil {
				return err
			}
			if args[0].Type() == STRING_OBJ {
				return NewGoObj(rl.LoadModel(args[0].(*Stringo).Value))
			}
			mesh, err := checkGoObjType[rl.Mesh]("load_model", 1, "rl.Mesh", args)
			if err != nil {
				return err
			}
			rl.LoadModelFromMesh(mesh.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`load_model` loads a model",
			signature: "load_model(filename_or_mesh: str|rl.Mesh) -> rl.Model\n" +
				"// Load model from files (meshes and materials)\n" +
				"Model LoadModel(const char *fileName);\n" +
				"// Load model from generated mesh (default material)\n" +
				"Model LoadModelFromMesh(Mesh mesh);",
			errors:  "InvalidArgCount,PositionalType",
			example: "load_model()",
		}.String(),
	},
	{
		Name: "_is_model_ready",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("is_model_ready", 1, args); err != nil {
				return err
			}
			model, err := checkGoObjType[rl.Model]("is_model_ready", 1, "rl.Model", args)
			if err != nil {
				return err
			}
			return nativeToBooleanObject(rl.IsModelReady(model.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_model_ready` returns true if the model is ready",
			signature: "is_model_ready(model: rl.Model) -> bool\n" +
				"// Check if a model is valid (loaded in GPU, VAO/VBOs)\n" +
				"bool IsModelValid(Model model);",
			errors:  "InvalidArgCount,PositionalType",
			example: "is_model_ready()",
		}.String(),
	},
	{
		Name: "_get_bounding_box",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("get_bounding_box", 1, args); err != nil {
				return err
			}
			model, err := checkGoObjType[rl.Model]("get_bounding_box", 1, "rl.Model", args)
			if err != nil {
				mesh, err := checkGoObjType[rl.Mesh]("get_bounding_box", 1, "rl.Mesh", args)
				if err != nil {
					return err
				}
				return NewGoObj(rl.GetMeshBoundingBox(mesh.Value))
			} else {
				return NewGoObj(rl.GetModelBoundingBox(model.Value))
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_bounding_box` returns the model/mesh's bounding box",
			signature: "get_bounding_box(model: rl.Model|rl.Mesh) -> rl.BoundingBox\n" +
				"// Compute model bounding box limits (considers all meshes)\n" +
				"BoundingBox GetModelBoundingBox(Model model);" +
				"// Compute mesh bounding box limits\n" +
				"BoundingBox GetMeshBoundingBox(Mesh mesh);",
			errors:  "InvalidArgCount,PositionalType",
			example: "get_bounding_box()",
		}.String(),
	},
	{
		Name: "_draw_model",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("draw_model", 7, args); err != nil {
				return err
			}
			if err := checkArgType("draw_model", 7, BOOLEAN_OBJ, args); err != nil {
				return err
			}
			model, err := checkGoObjType[rl.Model]("draw_model", 1, "rl.Model", args)
			if err != nil {
				return err
			}
			position, err := checkGoObjType[rl.Vector3]("draw_model", 2, "rl.Vector3", args)
			if err != nil {
				return err
			}
			withWires := args[6].(*Boolean).Value
			if args[5].Type() == NULL_OBJ {
				rotationAxis, err := checkGoObjType[rl.Vector3]("draw_model", 3, "rl.Vector3", args)
				if err != nil {
					return err
				}
				if err := checkArgType("draw_model", 4, FLOAT_OBJ, args); err != nil {
					return err
				}
				scale, err := checkGoObjType[rl.Vector3]("draw_model", 5, "rl.Vector3", args)
				if err != nil {
					return err
				}
				tint, err := checkGoObjType[rl.Color]("draw_model", 6, "rl.Color", args)
				if err != nil {
					return err
				}
				if withWires {
					rl.DrawModelWiresEx(model.Value, position.Value, rotationAxis.Value, float32(args[3].(*Float).Value), scale.Value, tint.Value)
				} else {
					rl.DrawModelEx(model.Value, position.Value, rotationAxis.Value, float32(args[3].(*Float).Value), scale.Value, tint.Value)
				}
			} else {
				if err := checkArgType("draw_model", 3, FLOAT_OBJ, args); err != nil {
					return err
				}
				tint, err := checkGoObjType[rl.Color]("draw_model", 4, "rl.Color", args)
				if err != nil {
					return err
				}
				if withWires {
					rl.DrawModelWires(model.Value, position.Value, float32(args[2].(*Float).Value), tint.Value)
				} else {
					rl.DrawModel(model.Value, position.Value, float32(args[2].(*Float).Value), tint.Value)
				}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_model` draws the model",
			signature: "draw_model(model: rl.Model) -> null\n" +
				"// Draw a model (with texture if set)\n" +
				"void DrawModel(Model model, Vector3 position, float scale, Color tint);\n" +
				"// Draw a model with extended parameters\n" +
				"void DrawModelEx(Model model, Vector3 position, Vector3 rotationAxis, float rotationAngle, Vector3 scale, Color tint);\n" +
				"// Draw a model wires (with texture if set)\n" +
				"void DrawModelWires(Model model, Vector3 position, float scale, Color tint);\n" +
				"// Draw a model wires (with texture if set) with extended parameters\n" +
				"void DrawModelWiresEx(Model model, Vector3 position, Vector3 rotationAxis, float rotationAngle, Vector3 scale, Color tint);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_model()",
		}.String(),
	},
	{
		Name: "_draw_bounding_box",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("draw_bounding_box", 2, args); err != nil {
				return err
			}
			boundingBox, err := checkGoObjType[rl.BoundingBox]("draw_bounding_box", 1, "rl.BoundingBox", args)
			if err != nil {
				return err
			}
			color, err := checkGoObjType[rl.Color]("draw_bounding_box", 1, "rl.Color", args)
			if err != nil {
				return err
			}
			rl.DrawBoundingBox(boundingBox.Value, color.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_bounding_box` draws a bounding box",
			signature: "draw_bounding_box(bounding_box: rl.BoundingBox, color: color) -> null\n" +
				"// Draw bounding box (wires)\n" +
				"void DrawBoundingBox(BoundingBox box, Color color);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_bounding_box()",
		}.String(),
	},
	{
		Name: "_draw_billboard",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("draw_billboard", []int{5, 6, 9}, args); err != nil {
				return err
			}
			camera, err := checkGoObjType[rl.Camera]("draw_billboard", 1, "rl.Camera", args)
			if err != nil {
				return err
			}
			texture, err := checkGoObjType[rl.Texture2D]("draw_billboard", 2, "rl.Texture2D", args)
			if err != nil {
				return err
			}
			if len(args) == 5 {
				position, err := checkGoObjType[rl.Vector3]("draw_billboard", 3, "rl.Vector3", args)
				if err != nil {
					return err
				}
				err = checkArgType("draw_billboard", 4, FLOAT_OBJ, args)
				if err != nil {
					return err
				}
				tint, err := checkGoObjType[rl.Color]("draw_billboard", 5, "rl.Color", args)
				if err != nil {
					return err
				}
				rl.DrawBillboard(camera.Value, texture.Value, position.Value, float32(args[3].(*Float).Value), tint.Value)
			} else if len(args) == 6 {
				source, err := checkGoObjType[rl.Rectangle]("draw_billboard", 3, "rl.Rectangle", args)
				if err != nil {
					return err
				}
				position, err := checkGoObjType[rl.Vector3]("draw_billboard", 4, "rl.Vector3", args)
				if err != nil {
					return err
				}
				size, err := checkGoObjType[rl.Vector2]("draw_billboard", 5, "rl.Vector2", args)
				if err != nil {
					return err
				}
				tint, err := checkGoObjType[rl.Color]("draw_billboard", 6, "rl.Color", args)
				if err != nil {
					return err
				}
				rl.DrawBillboardRec(camera.Value, texture.Value, source.Value, position.Value, size.Value, tint.Value)
			} else {
				source, err := checkGoObjType[rl.Rectangle]("draw_billboard", 3, "rl.Rectangle", args)
				if err != nil {
					return err
				}
				position, err := checkGoObjType[rl.Vector3]("draw_billboard", 4, "rl.Vector3", args)
				if err != nil {
					return err
				}
				up, err := checkGoObjType[rl.Vector3]("draw_billboard", 5, "rl.Vector3", args)
				if err != nil {
					return err
				}
				size, err := checkGoObjType[rl.Vector2]("draw_billboard", 6, "rl.Vector2", args)
				if err != nil {
					return err
				}
				origin, err := checkGoObjType[rl.Vector2]("draw_billboard", 7, "rl.Vector2", args)
				if err != nil {
					return err
				}
				err = checkArgType("draw_billboard", 7, FLOAT_OBJ, args)
				if err != nil {
					return err
				}
				tint, err := checkGoObjType[rl.Color]("draw_billboard", 8, "rl.Color", args)
				if err != nil {
					return err
				}
				rl.DrawBillboardPro(camera.Value, texture.Value, source.Value, position.Value, up.Value, origin.Value, size.Value, float32(args[7].(*Float).Value), tint.Value)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_billboard` draws a bounding box",
			signature: "draw_billboard(camera: rl.Camera, texture: rl.Texture2D, position: rl.Vector3, scale: float, tint: color) -> null\n" +
				"// Draw a billboard texture\n" +
				"void DrawBillboard(Camera camera, Texture2D texture, Vector3 position, float scale, Color tint);\n" +
				"// Draw a billboard texture defined by source\n" +
				"void DrawBillboardRec(Camera camera, Texture2D texture, Rectangle source, Vector3 position, Vector2 size, Color tint);\n" +
				"// Draw a billboard texture defined by source and rotation\n" +
				"void DrawBillboardPro(Camera camera, Texture2D texture, Rectangle source, Vector3 position, Vector3 up, Vector2 size, Vector2 origin, float rotation, Color tint);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_billboard()",
		}.String(),
	},
	{
		Name: "_upload_mesh",
		Fun: func(args ...Object) Object {
			err := checkArgCount("upload_mesh", 2, args)
			if err != nil {
				return err
			}
			mesh, err := checkGoObjType[rl.Mesh]("upload_mesh", 1, "rl.Mesh", args)
			if err != nil {
				return err
			}
			err = checkArgType("upload_mesh", 2, BOOLEAN_OBJ, args)
			if err != nil {
				return err
			}
			rl.UploadMesh(&mesh.Value, args[1].(*Boolean).Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`upload_mesh` uploads mesh vertex data",
			signature: "upload_mesh(mesh: rl.Mesh, dynamic: bool) -> null\n" +
				"// Upload mesh vertex data in GPU and provide VAO/VBO ids\n" +
				"void UploadMesh(Mesh *mesh, bool dynamic);",
			errors:  "InvalidArgCount,PositionalType",
			example: "upload_mesh()",
		}.String(),
	},
	{
		Name: "_draw_mesh",
		Fun: func(args ...Object) Object {
			err := checkArgsCount("draw_mesh", []int{3, 4}, args)
			if err != nil {
				return err
			}
			mesh, err := checkGoObjType[rl.Mesh]("draw_mesh", 1, "rl.Mesh", args)
			if err != nil {
				return err
			}
			material, err := checkGoObjType[rl.Material]("draw_mesh", 2, "rl.Material", args)
			if err != nil {
				return err
			}
			if len(args) == 3 {
				transform, err := checkGoObjType[rl.Matrix]("draw_mesh", 3, "rl.Matrix", args)
				if err != nil {
					return err
				}
				rl.DrawMesh(mesh.Value, material.Value, transform.Value)
			} else {
				err = checkArgType("draw_mesh", 3, LIST_OBJ, args)
				if err != nil {
					return err
				}
				elems := args[2].(*List).Elements
				transforms := make([]rl.Matrix, len(elems))
				for i, e := range elems {
					transform, ok := e.(*GoObj[rl.Matrix])
					if !ok {
						return newPositionalTypeErrorForGoObj("draw_mesh", 3, "list[rl.Matrix]", e)
					}
					transforms[i] = transform.Value
				}
				err = checkArgType("draw_mesh", 4, INTEGER_OBJ, args)
				if err != nil {
					return err
				}
				rl.DrawMeshInstanced(mesh.Value, material.Value, transforms, int(args[3].(*Integer).Value))
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`draw_mesh` draws the mesh or multiple mesh instances with material and different transforms",
			signature: "draw_mesh(mesh: rl.Mesh, material: rl.Material, transform: rl.Matrix) -> null\n" +
				"// Draw a 3d mesh with material and transform\n" +
				"void DrawMesh(Mesh mesh, Material material, Matrix transform);\n" +
				"// Draw multiple mesh instances with material and different transforms\n" +
				"void DrawMeshInstanced(Mesh mesh, Material material, const Matrix *transforms, int instances);",
			errors:  "InvalidArgCount,PositionalType",
			example: "draw_mesh()",
		}.String(),
	},
	{
		Name: "_export_mesh",
		Fun: func(args ...Object) Object {
			err := checkArgCount("export_mesh", 2, args)
			if err != nil {
				return err
			}
			mesh, err := checkGoObjType[rl.Mesh]("export_mesh", 1, "rl.Mesh", args)
			if err != nil {
				return err
			}
			err = checkArgType("export_mesh", 2, STRING_OBJ, args)
			if err != nil {
				return err
			}
			rl.ExportMesh(mesh.Value, args[1].(*Stringo).Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`export_mesh` exports the mesh as an obj file",
			signature: "export_mesh(mesh: rl.Mesh, filename: str) -> null\n" +
				"// Export mesh data to file, returns true on success\n" +
				"bool ExportMesh(Mesh mesh, const char *fileName);",
			errors:  "InvalidArgCount,PositionalType",
			example: "export_mesh()",
		}.String(),
	},
	{
		Name: "_gen_mesh_poly",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_poly", 2, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_poly", 1, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_poly", 2, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshPoly(int(args[0].(*Integer).Value), float32(args[1].(*Float).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_poly` generates poly mesh",
			signature: "gen_mesh_poly(sides: int, radius: float) -> rl.Mesh\n" +
				"// Generate polygonal mesh\n" +
				"Mesh GenMeshPoly(int sides, float radius);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_poly()",
		}.String(),
	},
	{
		Name: "_gen_mesh_plane",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_plane", 4, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_plane", 1, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_plane", 2, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_plane", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_plane", 4, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshPlane(float32(args[0].(*Float).Value), float32(args[1].(*Float).Value), int(args[2].(*Integer).Value), int(args[3].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_plane` generates plane mesh",
			signature: "gen_mesh_plane(width: float, height: float, res_x: int, res_z: int) -> rl.Mesh\n" +
				"// Generate plane mesh (with subdivisions)\n" +
				"Mesh GenMeshPlane(float width, float length, int resX, int resZ);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_plane()",
		}.String(),
	},
	{
		Name: "_gen_mesh_cube",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_cube", 3, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cube", 1, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cube", 2, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cube", 3, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshCube(float32(args[0].(*Float).Value), float32(args[1].(*Float).Value), float32(args[2].(*Float).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_cube` generates cube mesh",
			signature: "gen_mesh_cube(width: float, height: float, length: float) -> rl.Mesh\n" +
				"// Generate cuboid mesh\n" +
				"Mesh GenMeshCube(float width, float height, float length);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_cube()",
		}.String(),
	},
	{
		Name: "_gen_mesh_sphere",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_sphere", 3, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_sphere", 1, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_sphere", 2, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_sphere", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshSphere(float32(args[0].(*Float).Value), int(args[1].(*Integer).Value), int(args[2].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_sphere` generates sphere mesh",
			signature: "gen_mesh_sphere(radius: float, rings: int, slices: int) -> rl.Mesh\n" +
				"// Generate sphere mesh (standard sphere)\n" +
				"Mesh GenMeshSphere(float radius, int rings, int slices);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_sphere()",
		}.String(),
	},
	{
		Name: "_gen_mesh_hemi_sphere",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_hemi_shere", 3, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_hemi_shere", 1, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_hemi_shere", 2, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_hemi_shere", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshHemiSphere(float32(args[0].(*Float).Value), int(args[1].(*Integer).Value), int(args[2].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_hemi_sphere` generates half-sphere mesh (no bottom)",
			signature: "gen_mesh_hemi_sphere(radius: float, rings: int, slices: int) -> rl.Mesh\n" +
				"// Generate half-sphere mesh (no bottom cap)\n" +
				"Mesh GenMeshHemiSphere(float radius, int rings, int slices);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_hemi_sphere()",
		}.String(),
	},
	{
		Name: "_gen_mesh_cylinder",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_cylinder", 3, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cylinder", 1, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cylinder", 2, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cylinder", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshCylinder(float32(args[0].(*Float).Value), float32(args[1].(*Integer).Value), int(args[2].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_cylinder` generates cylinder mesh",
			signature: "gen_mesh_cylinder(radius: float, height: float, slices: int) -> rl.Mesh\n" +
				"// Generate cylinder mesh\n" +
				"Mesh GenMeshCylinder(float radius, float height, int slices);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_cylinder()",
		}.String(),
	},
	{
		Name: "_gen_mesh_cone",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_cone", 3, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cone", 1, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cone", 2, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cone", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshCone(float32(args[0].(*Float).Value), float32(args[1].(*Integer).Value), int(args[2].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_cone` generates cone/pyramid mesh",
			signature: "gen_mesh_cone(radius: float, height: float, slices: int) -> rl.Mesh\n" +
				"// Generate cone/pyramid mesh\n" +
				"Mesh GenMeshCone(float radius, float height, int slices);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_cone()",
		}.String(),
	},
	{
		Name: "_gen_mesh_torus",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_torus", 4, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cone", 1, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cone", 2, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cone", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_cone", 4, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshTorus(float32(args[0].(*Float).Value), float32(args[1].(*Float).Value), int(args[2].(*Integer).Value), int(args[3].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_torus` generates torus mesh",
			signature: "gen_mesh_torus(radius: float, size: float, rad_seg: int, sides: int) -> rl.Mesh\n" +
				"// Generate torus mesh\n" +
				"Mesh GenMeshTorus(float radius, float size, int radSeg, int sides);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_torus()",
		}.String(),
	},
	{
		Name: "_gen_mesh_knot",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_knot", 4, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_knot", 1, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_knot", 2, FLOAT_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_knot", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("gen_mesh_knot", 4, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshKnot(float32(args[0].(*Float).Value), float32(args[1].(*Float).Value), int(args[2].(*Integer).Value), int(args[3].(*Integer).Value)))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_knot` generates trefoil knot mesh",
			signature: "gen_mesh_knot(radius: float, size: float, rad_seg: int, sides: int) -> rl.Mesh\n" +
				"// Generate trefoil knot mesh\n" +
				"Mesh GenMeshKnot(float radius, float size, int radSeg, int sides);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_knot()",
		}.String(),
	},
	{
		Name: "_gen_mesh_heightmap",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_heightmap", 2, args)
			if err != nil {
				return err
			}
			heightMap, err := checkGoObjType[rl.Image]("gen_mesh_heightmap", 1, "rl.Image", args)
			if err != nil {
				return err
			}
			size, err := checkGoObjType[rl.Vector3]("gen_mesh_heightmap", 2, "rl.Vector3", args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshHeightmap(heightMap.Value, size.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_heightmap` generates heightmap mesh from image data",
			signature: "gen_mesh_heightmap(heightmap: rl.Image, size: rl.Vector3) -> rl.Mesh\n" +
				"// Generate heightmap mesh from image data\n" +
				"Mesh GenMeshHeightmap(Image heightmap, Vector3 size);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_heightmap()",
		}.String(),
	},
	{
		Name: "_gen_mesh_cubicmap",
		Fun: func(args ...Object) Object {
			err := checkArgCount("gen_mesh_cubicmap", 2, args)
			if err != nil {
				return err
			}
			cubicMap, err := checkGoObjType[rl.Image]("gen_mesh_cubicmap", 1, "rl.Image", args)
			if err != nil {
				return err
			}
			size, err := checkGoObjType[rl.Vector3]("gen_mesh_cubicmap", 2, "rl.Vector3", args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.GenMeshCubicmap(cubicMap.Value, size.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`gen_mesh_cubicmap` generates cube based map mesh from image data",
			signature: "gen_mesh_cubicmap(cubicmap: rl.Image, size: rl.Vector3) -> rl.Mesh\n" +
				"// Generate cubes-based map mesh from image data\n" +
				"Mesh GenMeshCubicmap(Image cubicmap, Vector3 cubeSize);",
			errors:  "InvalidArgCount,PositionalType",
			example: "gen_mesh_cubicmap()",
		}.String(),
	},
	{
		Name: "_load_materials",
		Fun: func(args ...Object) Object {
			err := checkArgCount("load_materials", 1, args)
			if err != nil {
				return err
			}
			err = checkArgType("load_materials", 1, STRING_OBJ, args)
			if err != nil {
				return err
			}
			materials := rl.LoadMaterials(args[0].(*Stringo).Value)
			materialsList := make([]Object, len(materials))
			for i, e := range materials {
				materialsList[i] = NewGoObj(e)
			}
			return &List{Elements: materialsList}
		},
		HelpStr: helpStrArgs{
			explanation: "`load_materials` loads materials from model file (.MTL)",
			signature: "load_materials(filename: str) -> list[rl.Material]\n" +
				"// Load materials from model file\n" +
				"Material *LoadMaterials(const char *fileName, int *materialCount);",
			errors:  "InvalidArgCount,PositionalType",
			example: "load_materials()",
		}.String(),
	},
	{
		Name: "_load_material_default",
		Fun: func(args ...Object) Object {
			err := checkArgCount("load_material_default", 0, args)
			if err != nil {
				return err
			}
			return NewGoObj(rl.LoadMaterialDefault())
		},
		HelpStr: helpStrArgs{
			explanation: "`load_material_default` loads default material (supports: DIFFUSE, SPECULAR, NORMAL maps)",
			signature: "load_material_default() -> rl.Material\n" +
				"// Load default material (Supports: DIFFUSE, SPECULAR, NORMAL maps)\n" +
				"Material LoadMaterialDefault(void);",
			errors:  "InvalidArgCount,PositionalType",
			example: "load_material_default()",
		}.String(),
	},
	{
		Name: "_is_material_ready",
		Fun: func(args ...Object) Object {
			err := checkArgCount("is_material_ready", 1, args)
			if err != nil {
				return err
			}
			material, err := checkGoObjType[rl.Material]("is_material_ready", 1, "rl.Material", args)
			if err != nil {
				return err
			}
			return nativeToBooleanObject(rl.IsMaterialReady(material.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_material_ready` returns true if material is ready",
			signature:   "is_material_ready(material: rl.Material) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_material_ready()",
		}.String(),
	},
	{
		Name: "_set_material_texture",
		Fun: func(args ...Object) Object {
			err := checkArgCount("set_material_texture", 3, args)
			if err != nil {
				return err
			}
			material, err := checkGoObjType[rl.Material]("set_material_texture", 1, "rl.Material", args)
			if err != nil {
				return err
			}
			err = checkArgType("set_material_texture", 2, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			texture, err := checkGoObjType[rl.Texture2D]("set_material_texture", 3, "rl.Texture2D", args)
			if err != nil {
				return err
			}
			rl.SetMaterialTexture(&material.Value, int32(args[1].(*Integer).Value), texture.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_material_texture` sets texture for a material map type",
			signature: "set_material_texture(material: rl.Material, map_type: int, texture: rl.Texture2D) -> null\n" +
				"// Set texture for a material map type (MATERIAL_MAP_DIFFUSE, MATERIAL_MAP_SPECULAR...)\n" +
				"void SetMaterialTexture(Material *material, int mapType, Texture2D texture);",
			errors:  "InvalidArgCount,PositionalType",
			example: "set_material_texture()",
		}.String(),
	},
	{
		Name: "_set_model_mesh_material",
		Fun: func(args ...Object) Object {
			err := checkArgCount("set_model_mesh_material", 3, args)
			if err != nil {
				return err
			}
			model, err := checkGoObjType[rl.Model]("set_model_mesh_material", 1, "rl.Model", args)
			if err != nil {
				return err
			}
			err = checkArgType("set_model_mesh_material", 2, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			err = checkArgType("set_model_mesh_material", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			rl.SetModelMeshMaterial(&model.Value, int32(args[1].(*Integer).Value), int32(args[2].(*Integer).Value))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_model_mesh_material` sets material for mesh",
			signature: "set_model_mesh_material(model: rl.Model, mesh_id: int, material_id: int) -> null\n" +
				"// Set material for a mesh\n" +
				"void SetModelMeshMaterial(Model *model, int meshId, int materialId);",
			errors:  "InvalidArgCount,PositionalType",
			example: "set_model_mesh_material()",
		}.String(),
	},
	{
		Name: "_load_model_animations",
		Fun: func(args ...Object) Object {
			err := checkArgCount("load_model_animations", 1, args)
			if err != nil {
				return err
			}
			err = checkArgType("load_model_animation", 1, STRING_OBJ, args)
			if err != nil {
				return err
			}
			modelAnimations := rl.LoadModelAnimations(args[0].(*Stringo).Value)
			modelAnimationsList := make([]Object, len(modelAnimations))
			for i, e := range modelAnimations {
				modelAnimationsList[i] = NewGoObj(e)
			}
			return &List{Elements: modelAnimationsList}
		},
		HelpStr: helpStrArgs{
			explanation: "`load_model_animations` loads model animations from file",
			signature: "load_model_animations(filename: str) -> list[rl.ModelAnimation]\n" +
				"// Load model animations from file\n" +
				"ModelAnimation *LoadModelAnimations(const char *fileName, int *animCount);",
			errors:  "InvalidArgCount,PositionalType",
			example: "load_model_animations()",
		}.String(),
	},
	{
		Name: "_update_model_animation",
		Fun: func(args ...Object) Object {
			err := checkArgCount("update_model_animation", 3, args)
			if err != nil {
				return err
			}
			model, err := checkGoObjType[rl.Model]("update_model_animation", 1, "rl.Model", args)
			if err != nil {
				return err
			}
			modelAnimation, err := checkGoObjType[rl.ModelAnimation]("update_model_animation", 2, "rl.ModelAnimation", args)
			if err != nil {
				return err
			}
			err = checkArgType("update_model_animation", 3, INTEGER_OBJ, args)
			if err != nil {
				return err
			}
			rl.UpdateModelAnimation(model.Value, modelAnimation.Value, int32(args[2].(*Integer).Value))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`update_model_animations` updates model animation pose",
			signature: "update_model_animations(model: rl.Model, model_animation: rl.ModelAnimation, frame: int) -> null\n" +
				"// Update model animation pose (CPU)\n" +
				"void UpdateModelAnimation(Model model, ModelAnimation anim, int frame);",
			errors:  "InvalidArgCount,PositionalType",
			example: "update_model_animations()",
		}.String(),
	},
	{
		Name: "_is_model_animation_valid",
		Fun: func(args ...Object) Object {
			err := checkArgCount("is_model_animation_valid", 2, args)
			if err != nil {
				return err
			}
			model, err := checkGoObjType[rl.Model]("is_model_animation_valid", 1, "rl.Model", args)
			if err != nil {
				return err
			}
			modelAnimation, err := checkGoObjType[rl.ModelAnimation]("is_model_animation_valid", 2, "rl.ModelAnimation", args)
			if err != nil {
				return err
			}
			return nativeToBooleanObject(rl.IsModelAnimationValid(model.Value, modelAnimation.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_model_animation_valid` returns true if model animation skeleon matches",
			signature: "is_model_animation_valid(model: rl.Model, model_animation: rl.ModelAnimation) -> bool\n" +
				"// Check model animation skeleton match\n" +
				"bool IsModelAnimationValid(Model model, ModelAnimation anim);",
			errors:  "InvalidArgCount,PositionalType",
			example: "is_model_animation_valid()",
		}.String(),
	},
	{
		Name: "_unload",
		Fun: func(args ...Object) Object {
			for i, arg := range args {
				// If the arg is a list go through the list and check every arg to remove
				if arg.Type() == LIST_OBJ {
					l := arg.(*List).Elements
					for _, e := range l {
						if err := unloadFromRaylib(e, i); err != nil {
							return err
						}
					}
				} else {
					if err := unloadFromRaylib(arg, i); err != nil {
						return err
					}
				}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`unload` unloads the given objects from the gg object",
			signature:   "unload(args...: rl.any|list[rl.any]) -> null",
			errors:      "CustomError",
			example:     "unload() => null",
		}.String(),
	},
}

func unloadFromRaylib(arg Object, pos int) Object {
	if arg.Type() != GO_OBJ {
		return newPositionalTypeError("unload", pos, GO_OBJ, arg.Type())
	}
	if tex, ok := arg.(*GoObj[rl.Texture2D]); ok {
		rl.UnloadTexture(tex.Value)
		return nil
	} else if music, ok := arg.(*GoObj[rl.Music]); ok {
		rl.UnloadMusicStream(music.Value)
		return nil
	} else if sound, ok := arg.(*GoObj[rl.Sound]); ok {
		rl.UnloadSound(sound.Value)
		return nil
	} else if model, ok := arg.(*GoObj[rl.Model]); ok {
		rl.UnloadModel(model.Value)
		return nil
	} else if mesh, ok := arg.(*GoObj[rl.Mesh]); ok {
		rl.UnloadMesh(&mesh.Value)
		return nil
	} else if material, ok := arg.(*GoObj[rl.Material]); ok {
		rl.UnloadMaterial(material.Value)
		return nil
	} else if modelAnimation, ok := arg.(*GoObj[rl.ModelAnimation]); ok {
		// Note: there is a version for []rl.ModelAnimation that may be more efficient
		rl.UnloadModelAnimation(modelAnimation.Value)
		return nil
	}
	return newError("`unload` error: Failed to find gg object to unload, expected any GO_OBJ of rl.* that has an unload function")
}
