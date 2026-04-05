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
					rl.CheckCollisionPointRec(point.Value, rec.Value)
				} else if rec, ok := args[0].(*GoObj[rl.Rectangle]); ok {
					rec1, ok := args[1].(*GoObj[rl.Rectangle])
					if !ok {
						return newPositionalTypeErrorForGoObj("check_collision", 2, "rl.Rectangle", args[1])
					}
					rl.CheckCollisionRecs(rec.Value, rec1.Value)
				} else {
					return newPositionalTypeErrorForGoObj("check_collision", 1, "rl.Vector2 or rl.Rectangle", args[0])
				}
			case 3:
				pointOrCenter, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("check_collision", 1, "rl.Vector2", args[0])
				}
				if args[1].Type() == FLOAT_OBJ {
					rec, ok := args[2].(*GoObj[rl.Rectangle])
					if !ok {
						return newPositionalTypeErrorForGoObj("check_collision", 3, "rl.Rectangle", args[2])
					}
					rl.CheckCollisionCircleRec(pointOrCenter.Value, float32(args[1].(*Float).Value), rec.Value)
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
					rl.CheckCollisionPointPoly(pointOrCenter.Value, points, int32(args[2].(*Integer).Value))
				} else if args[1].Type() == GO_OBJ {
					center, ok := args[1].(*GoObj[rl.Vector2])
					if !ok {
						return newPositionalTypeErrorForGoObj("check_collision", 2, "rl.Vector2", args[1])
					}
					if args[2].Type() != FLOAT_OBJ {
						return newPositionalTypeError("check_collision", 3, FLOAT_OBJ, args[2].Type())
					}
					rl.CheckCollisionPointCircle(pointOrCenter.Value, center.Value, float32(args[2].(*Float).Value))
				} else {
					return newPositionalTypeErrorForGoObj("check_collision", 2, "float or rl.Vector2 or list[rl.Vector2]", args[1])
				}
			case 4:
				pointOrCenter, ok := args[0].(*GoObj[rl.Vector2])
				if !ok {
					return newPositionalTypeErrorForGoObj("check_collision", 1, "rl.Vector2", args[0])
				}
				if args[1].Type() == FLOAT_OBJ {
					center, ok := args[2].(*GoObj[rl.Vector2])
					if !ok {
						return newPositionalTypeErrorForGoObj("check_collision", 3, "rl.Vector2", args[2])
					}
					if args[3].Type() != FLOAT_OBJ {
						return newPositionalTypeError("check_collision", 4, FLOAT_OBJ, args[3].Type())
					}
					rl.CheckCollisionCircles(pointOrCenter.Value, float32(args[1].(*Float).Value), center.Value, float32(args[3].(*Float).Value))
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
						rl.CheckCollisionPointTriangle(pointOrCenter.Value, point2.Value, point3.Value, point4.Value)
					} else if args[3].Type() != INTEGER_OBJ {
						return newPositionalTypeError("check_collision", 4, INTEGER_OBJ, args[3].Type())
					}
					rl.CheckCollisionPointLine(pointOrCenter.Value, point2.Value, point3.Value, int32(args[3].(*Integer).Value))
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
				"bool CheckCollisionPointPoly(Vector2 point, const Vector2 *points, int pointCount);",
			errors:  "InvalidArgCount,PositionalType",
			example: "check_collision() => (see signature for examples)=>true",
		}.String(),
	},
	{
		Name: "_get_collision_rec",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("get_collision_rect", len(args), 2, "")
			}
			rec1, ok := args[0].(*GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("get_collision_rec", 1, "rl.Rectangle", args[0])
			}
			rec2, ok := args[1].(*GoObj[rl.Rectangle])
			if !ok {
				return newPositionalTypeErrorForGoObj("get_collision_rec", 2, "rl.Rectangle", args[1])
			}
			return NewGoObj(rl.GetCollisionRec(rec1.Value, rec2.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`get_collision_rec` returns the collision rectangle",
			signature: "get_collision_rec(rec1: GoObj[rl.Rectangle], rec2: GoObj[rl.Rectangle]) -> GoObj[rl.Rectangle]\n" +
				"// Get collision rectangle for two rectangles collision\n" +
				"Rectangle GetCollisionRec(Rectangle rec1, Rectangle rec2);",
			errors:  "InvalidArgCount,PositionalType",
			example: "get_collision_rec() => (see signature for examples)=>rl.Rectangle",
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
		Name: "_unload",
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
