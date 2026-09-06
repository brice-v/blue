//go:build !static

package object

import (
	"image/color"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var UiBuiltins = []*Builtin{
	{
		Name: "_new_app",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_window",
		Fun: func(args ...Object) Object {
			if len(args) != 5 {
				return newInvalidArgCountError("window", len(args), 5, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("window", 1, GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*GoObj[fyne.App])
			if !ok {
				return newPositionalTypeErrorForGoObj("window", 1, "fyne.App", args[0])
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("window", 2, INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != INTEGER_OBJ {
				return newPositionalTypeError("window", 3, INTEGER_OBJ, args[2].Type())
			}
			if args[3].Type() != STRING_OBJ {
				return newPositionalTypeError("window", 4, STRING_OBJ, args[3].Type())
			}
			if args[4].Type() != GO_OBJ {
				return newPositionalTypeError("window", 5, GO_OBJ, args[4].Type())
			}
			content, ok := args[4].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("window", 4, "fyne.CanvasObject", args[4])
			}
			width := args[1].(*Integer).Value
			height := args[2].(*Integer).Value
			title := args[3].(*Stringo).Value
			w := app.Value.NewWindow(title)
			w.SetFixedSize(true)
			w.SetContent(content.Value)
			w.Resize(fyne.Size{Width: float32(width), Height: float32(height)})
			w.CenterOnScreen()
			w.ShowAndRun()
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`window` runs the window function on the given app to display the ui with the given content",
			signature:   "window(app: GoObj[fyne.App], width: int=400, height: int=400, title: str='blue ui window', content: GoObj[fyne.CanvasObject]=null) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "window(app) => null (side effect, shows ui window)",
		}.String(),
	},
	{
		Name: "_label",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("label", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("label", 1, STRING_OBJ, args[0].Type())
			}
			label := args[0].(*Stringo).Value
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
	{
		Name: "_progress_bar",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("progress_bar", len(args), 1, "")
			}
			if args[0].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("progress_bar", 1, BOOLEAN_OBJ, args[0].Type())
			}
			isInfinite := args[0].(*Boolean).Value
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
	{
		Name: "_progress_bar_set_value",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("progress_bar_set_value", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("progress_bar_set_value", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("progress_bar_set_value", 2, FLOAT_OBJ, args[1].Type())
			}
			value := args[1].(*Float).Value
			progressBar, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("progress_bar_set_value", 1, "fyne.CanvasObject", args[0])
			}
			switch x := progressBar.Value.(type) {
			case *widget.ProgressBar:
				x.SetValue(value)
				return NULL
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
	{
		Name: "_toolbar",
		Fun: func(args ...Object) Object {
			if len(args) == 0 {
				return newInvalidArgCountError("toolbar", len(args), 1, "or more")
			}
			tis := []widget.ToolbarItem{}
			for i, arg := range args {
				if arg.Type() != GO_OBJ {
					return newPositionalTypeError("toolbar", i+1, GO_OBJ, arg.Type())
				}
				ti, ok := arg.(*GoObj[widget.ToolbarItem])
				if !ok {
					return newPositionalTypeErrorForGoObj("toolbar", i+1, "widget.ToolbarItem", arg)
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
	{
		Name: "_toolbar_spacer",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_toolbar_separator",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_row",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("row", len(args), 1, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("row", 1, LIST_OBJ, args[0].Type())
			}
			elements := args[0].(*List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != GO_OBJ {
					return newError("`row` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				o, ok := e.(*GoObj[fyne.CanvasObject])
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
	{
		Name: "_col",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("col", len(args), 1, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("col", 1, LIST_OBJ, args[0].Type())
			}
			elements := args[0].(*List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != GO_OBJ {
					return newError("`col` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				o, ok := e.(*GoObj[fyne.CanvasObject])
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
	{
		Name: "_grid",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("grid", len(args), 2, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("grid", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("grid", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != LIST_OBJ {
				return newPositionalTypeError("grid", 3, LIST_OBJ, args[2].Type())
			}
			rowsOrCols := int(args[0].(*Integer).Value)
			gridType := args[1].(*Stringo).Value
			if gridType != "COLS" && gridType != "ROWS" {
				return newError("`grid` error: type must be COLS or ROWS. got=%s", gridType)
			}
			elements := args[2].(*List).Elements
			canvasObjects := make([]fyne.CanvasObject, len(elements))
			for i, e := range elements {
				if e.Type() != GO_OBJ {
					return newError("`grid` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				o, ok := e.(*GoObj[fyne.CanvasObject])
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
	{
		Name: "_entry",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("entry", len(args), 1, "")
			}
			if args[0].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("entry", 1, BOOLEAN_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("entry", 2, STRING_OBJ, args[1].Type())
			}
			isMultiline := args[0].(*Boolean).Value
			placeholderText := args[1].(*Stringo).Value
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
	{
		Name: "_entry_get_text",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("entry_get_text", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("entry_get_text", 1, GO_OBJ, args[0].Type())
			}
			entry, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("entry_get_text", 1, "fyne.CanvasObject", args[0])
			}
			switch x := entry.Value.(type) {
			case *widget.Entry:
				return &Stringo{Value: x.Text}
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
	{
		Name: "_entry_set_text",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("entry_set_text", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("entry_set_text", 1, GO_OBJ, args[0].Type())
			}
			entry, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("entry_set_text", 1, "fyne.CanvasObject", args[0])
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("entry_set_text", 2, STRING_OBJ, args[1].Type())
			}
			value := args[1].(*Stringo).Value
			switch x := entry.Value.(type) {
			case *widget.Entry:
				x.SetText(value)
				return NULL
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
	{
		Name: "_append_form",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("append_form", len(args), 3, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("append_form", 1, GO_OBJ, args[0].Type())
			}
			maybeForm, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("append_form", 1, "fyne.CanvasObject", args[0])
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("append_form", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != GO_OBJ {
				return newPositionalTypeError("append_form", 3, GO_OBJ, args[2].Type())
			}
			w, ok := args[2].(*GoObj[fyne.CanvasObject])
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
			form.Append(args[1].(*Stringo).Value, w.Value)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`append_form` appends a label with the given string and a corresponding widget to the given form",
			signature:   "append_form(f: GoObj[fyne.CanvasObject](Value: *widget.Form), title: str, widget: GoObj[fyne.CanvasObject]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "append_form(f, 'test', w) => null (side effect, refresh ui form with updated label/widget)",
		}.String(),
	},
	{
		Name: "_icon_account",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_cancel",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_check_button_checked",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_check_button",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_color_achromatic",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_color_chromatic",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_color_palette",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_computer",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_confirm",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_content_add",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_content_clear",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_content_copy",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_content_cut",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_content_paste",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_content_redo",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_content_remove",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_content_undo",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_delete",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_document_create",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_document",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_document_print",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_document_save",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_download",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_error",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_file_application",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_file_audio",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_file",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_file_image",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_file_text",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_file_video",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_folder",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_folder_new",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_folder_open",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_grid",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_help",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_history",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_home",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_info",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_list",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_login",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_logout",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_mail_attachment",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_mail_compose",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_mail_forward",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_mail_reply_all",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_mail_reply",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_mail_send",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_fast_forward",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_fast_rewind",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_music",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_pause",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_photo",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_play",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_record",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_replay",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_skip_next",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_skip_previous",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_stop",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_media_video",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_menu_drop_down",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_menu_drop_up",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_menu_expand",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_menu",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_more_horizontal",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_more_vertical",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_move_down",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_move_up",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_navigate_back",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_navigate_next",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_question",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_radio_button_checked",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_radio_button",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_search",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_search_replace",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_settings",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_storage",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_upload",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_view_full_screen",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_view_refresh",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_view_restore",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_visibility",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_visibility_off",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_volume_down",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_volume_mute",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_volume_up",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_warning",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_zoom_fit",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_zoom_in",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_icon_zoom_out",
		Fun: func(args ...Object) Object {
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
	{
		Name: "_button",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`button` returns a ui button widget object with a string label and an onclick function handler",
			signature:   "button(label: str, fn: fun()) -> GoObj[fyne.CanvasObject](Value: *widget.Button)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "button('Click Me!', || => {println('clicked')}) => GoObj[fyne.CanvasObject](Value: *widget.Button)",
		}.String(),
	},
	{
		Name: "_check_box",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`check_box` returns a ui check_box widget object with a string label and an onchecked function handler",
			signature:   "check_box(label: str, fn: fun(is_checked: bool)) -> GoObj[fyne.CanvasObject](Value: *widget.Check)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "check_box('Check Me!', |e| => {println('checked? #{e}')}) => GoObj[fyne.CanvasObject](Value: *widget.Check)",
		}.String(),
	},
	{
		Name: "_radio_group",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`radio_group` returns a ui radio_group widget object with a list of string radio labels and an onchecked function handler",
			signature:   "radio_group(labels: list[str], fn: fun(checked_label: str)) -> GoObj[fyne.CanvasObject](Value: *widget.RadioGroup)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "radio_group(['Check Me 1!', 'Check Me 2!'], |e| => {println('checked #{e}')}) => GoObj[fyne.CanvasObject](Value: *widget.RadioGroup)",
		}.String(),
	},
	{
		Name: "_option_select",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`option_select` returns a ui option_select widget object with a list of string options and an onchecked function handler",
			signature:   "option_select(labels: list[str], fn: fun(checked_option: str)) -> GoObj[fyne.CanvasObject](Value: *widget.Select)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "option_select(['Check Me 1!', 'Check Me 2!'], |e| => {println('checked #{e}')}) => GoObj[fyne.CanvasObject](Value: *widget.Select)",
		}.String(),
	},
	{
		Name: "_form",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`form` returns a ui form widget object with the given list of labels and widgets, and a submit function",
			signature:   "form(elements: list[{'label': str, 'widget': GoObj[fyne.CanvasObject]}]=[], fn: fun()) -> GoObj[fyne.CanvasObject](Value: *widget.Form)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "form(|| => {println('submit')}) => GoObj[fyne.CanvasObject](Value: *widget.Form)",
		}.String(),
	},
	{
		Name: "_toolbar_action",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`toolbar.action()`: `toolbar_action` returns a ui toolbar_action widget object which can be added to a toolbar when given a resource a function to execute on action",
			signature:   "toolbar_action(res: GoObj[fyne.Resource], fn: fun()) -> GoObj[widget.ToolbarItem](Value: *widget.ToolbarAction)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "toolbar_action(icon.computer, || => {println('action!')}) => GoObj[widget.ToolbarItem](Value: *widget.ToolbarAction)",
		}.String(),
	},
	// --- Expanded fyne coverage ---
	{
		Name: "_separator",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("separator", len(args), 0, "")
			}
			return NewGoObj[fyne.CanvasObject](widget.NewSeparator())
		},
		HelpStr: helpStrArgs{
			explanation: "`separator` returns a horizontal/vertical separator widget (adapts to layout orientation)",
			signature:   "separator() -> GoObj[fyne.CanvasObject](Value: *widget.Separator)",
			errors:      "InvalidArgCount",
			example:     "separator() => GoObj[fyne.CanvasObject](Value: *widget.Separator)",
		}.String(),
	},
	{
		Name: "_hyperlink",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("hyperlink", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("hyperlink", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("hyperlink", 2, STRING_OBJ, args[1].Type())
			}
			text := args[0].(*Stringo).Value
			rawURL := args[1].(*Stringo).Value
			u, err := url.Parse(rawURL)
			if err != nil {
				return newError("`hyperlink` error: invalid URL `%s`: %s", rawURL, err.Error())
			}
			return NewGoObj[fyne.CanvasObject](widget.NewHyperlink(text, u))
		},
		HelpStr: helpStrArgs{
			explanation: "`hyperlink` returns a clickable hyperlink widget that opens the given URL (or triggers OnTapped if overridden via hyperlink_with_handler)",
			signature:   "hyperlink(text: str, url: str) -> GoObj[fyne.CanvasObject](Value: *widget.Hyperlink)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "hyperlink('Fyne', 'https://fyne.io') => GoObj[fyne.CanvasObject](Value: *widget.Hyperlink)",
		}.String(),
	},
	{
		Name: "_icon_widget",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("icon_widget", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("icon_widget", 1, GO_OBJ, args[0].Type())
			}
			r, ok := args[0].(*GoObj[fyne.Resource])
			if !ok {
				return newPositionalTypeErrorForGoObj("icon_widget", 1, "fyne.Resource", args[0])
			}
			return NewGoObj[fyne.CanvasObject](widget.NewIcon(r.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`icon_widget` returns an icon widget displaying the given fyne.Resource",
			signature:   "icon_widget(res: GoObj[fyne.Resource]) -> GoObj[fyne.CanvasObject](Value: *widget.Icon)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "icon_widget(icon.home) => GoObj[fyne.CanvasObject](Value: *widget.Icon)",
		}.String(),
	},
	{
		Name: "_card",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("card", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("card", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("card", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != GO_OBJ {
				return newPositionalTypeError("card", 3, GO_OBJ, args[2].Type())
			}
			title := args[0].(*Stringo).Value
			subtitle := args[1].(*Stringo).Value
			c, ok := args[2].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("card", 3, "fyne.CanvasObject", args[2])
			}
			return NewGoObj[fyne.CanvasObject](widget.NewCard(title, subtitle, c.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`card` returns a card widget grouping title, subtitle and content with shadow",
			signature:   "card(title: str, subtitle: str, content: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *widget.Card)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "card('Title', 'Subtitle', label('hi')) => GoObj[fyne.CanvasObject](Value: *widget.Card)",
		}.String(),
	},
	{
		Name: "_accordion_item",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("accordion_item", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("accordion_item", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("accordion_item", 2, GO_OBJ, args[1].Type())
			}
			title := args[0].(*Stringo).Value
			c, ok := args[1].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("accordion_item", 2, "fyne.CanvasObject", args[1])
			}
			item := widget.NewAccordionItem(title, c.Value)
			return NewGoObj(item)
		},
		HelpStr: helpStrArgs{
			explanation: "`accordion_item` creates an accordion item with title and detail content",
			signature:   "accordion_item(title: str, detail: GoObj[fyne.CanvasObject]) -> GoObj[*widget.AccordionItem]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "accordion_item('Section', label('detail')) => GoObj[*widget.AccordionItem]",
		}.String(),
	},
	{
		Name: "_accordion",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("accordion", len(args), 1, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("accordion", 1, LIST_OBJ, args[0].Type())
			}
			elems := args[0].(*List).Elements
			items := make([]*widget.AccordionItem, len(elems))
			for i, e := range elems {
				if e.Type() != GO_OBJ {
					return newPositionalTypeError("accordion", 1, GO_OBJ, e.Type())
				}
				item, ok := e.(*GoObj[*widget.AccordionItem])
				if !ok {
					return newPositionalTypeErrorForGoObj("accordion", 1, "*widget.AccordionItem", e)
				}
				items[i] = item.Value
			}
			return NewGoObj[fyne.CanvasObject](widget.NewAccordion(items...))
		},
		HelpStr: helpStrArgs{
			explanation: "`accordion` returns an accordion widget displaying collapsible items",
			signature:   "accordion(items: list[GoObj[*widget.AccordionItem]]) -> GoObj[fyne.CanvasObject](Value: *widget.Accordion)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "accordion([accordion_item('A', label('hi'))]) => GoObj[fyne.CanvasObject](Value: *widget.Accordion)",
		}.String(),
	},
	{
		Name: "_tab_item",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("tab_item", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("tab_item", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("tab_item", 2, GO_OBJ, args[1].Type())
			}
			text := args[0].(*Stringo).Value
			c, ok := args[1].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("tab_item", 2, "fyne.CanvasObject", args[1])
			}
			return NewGoObj(container.NewTabItem(text, c.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`tab_item` creates a tab item with text and content for use in tabs",
			signature:   "tab_item(text: str, content: GoObj[fyne.CanvasObject]) -> GoObj[*container.TabItem]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "tab_item('Tab1', label('content')) => GoObj[*container.TabItem]",
		}.String(),
	},
	{
		Name: "_tab_item_with_icon",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("tab_item_with_icon", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("tab_item_with_icon", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("tab_item_with_icon", 2, GO_OBJ, args[1].Type())
			}
			if args[2].Type() != GO_OBJ {
				return newPositionalTypeError("tab_item_with_icon", 3, GO_OBJ, args[2].Type())
			}
			text := args[0].(*Stringo).Value
			r, ok := args[1].(*GoObj[fyne.Resource])
			if !ok {
				return newPositionalTypeErrorForGoObj("tab_item_with_icon", 2, "fyne.Resource", args[1])
			}
			c, ok := args[2].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("tab_item_with_icon", 3, "fyne.CanvasObject", args[2])
			}
			return NewGoObj(container.NewTabItemWithIcon(text, r.Value, c.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`tab_item_with_icon` creates a tab item with text, icon and content",
			signature:   "tab_item_with_icon(text: str, icon: GoObj[fyne.Resource], content: GoObj[fyne.CanvasObject]) -> GoObj[*container.TabItem]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "tab_item_with_icon('Home', icon.home, label('hi')) => GoObj[*container.TabItem]",
		}.String(),
	},
	{
		Name: "_tabs",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("tabs", len(args), 1, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("tabs", 1, LIST_OBJ, args[0].Type())
			}
			elems := args[0].(*List).Elements
			items := make([]*container.TabItem, len(elems))
			for i, e := range elems {
				if e.Type() != GO_OBJ {
					return newPositionalTypeError("tabs", 1, GO_OBJ, e.Type())
				}
				ti, ok := e.(*GoObj[*container.TabItem])
				if !ok {
					return newPositionalTypeErrorForGoObj("tabs", 1, "*container.TabItem", e)
				}
				items[i] = ti.Value
			}
			return NewGoObj[fyne.CanvasObject](container.NewAppTabs(items...))
		},
		HelpStr: helpStrArgs{
			explanation: "`tabs` returns an app-tabs container with the given tab items (top bar, overflow menu)",
			signature:   "tabs(items: list[GoObj[*container.TabItem]]) -> GoObj[fyne.CanvasObject](Value: *container.AppTabs)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "tabs([tab_item('A', label('hi'))]) => GoObj[fyne.CanvasObject](Value: *container.AppTabs)",
		}.String(),
	},
	{
		Name: "_doc_tabs",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("doc_tabs", len(args), 1, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("doc_tabs", 1, LIST_OBJ, args[0].Type())
			}
			elems := args[0].(*List).Elements
			items := make([]*container.TabItem, len(elems))
			for i, e := range elems {
				if e.Type() != GO_OBJ {
					return newPositionalTypeError("doc_tabs", 1, GO_OBJ, e.Type())
				}
				ti, ok := e.(*GoObj[*container.TabItem])
				if !ok {
					return newPositionalTypeErrorForGoObj("doc_tabs", 1, "*container.TabItem", e)
				}
				items[i] = ti.Value
			}
			return NewGoObj[fyne.CanvasObject](container.NewDocTabs(items...))
		},
		HelpStr: helpStrArgs{
			explanation: "`doc_tabs` returns a document-tabs container (closable tabs, typically for editors)",
			signature:   "doc_tabs(items: list[GoObj[*container.TabItem]]) -> GoObj[fyne.CanvasObject](Value: *container.DocTabs)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "doc_tabs([tab_item('Doc1', label('hi'))]) => GoObj[fyne.CanvasObject](Value: *container.DocTabs)",
		}.String(),
	},
	{
		Name: "_scroll",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("scroll", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("scroll", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("scroll", 1, "fyne.CanvasObject", args[0])
			}
			return NewGoObj[fyne.CanvasObject](container.NewScroll(c.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`scroll` wraps content in a scroll container (both directions, auto scrollbars)",
			signature:   "scroll(content: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *container.Scroll)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "scroll(label('long')) => GoObj[fyne.CanvasObject](Value: *container.Scroll)",
		}.String(),
	},
	{
		Name: "_hscroll",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("hscroll", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("hscroll", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("hscroll", 1, "fyne.CanvasObject", args[0])
			}
			return NewGoObj[fyne.CanvasObject](container.NewHScroll(c.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`hscroll` wraps content in a horizontal-only scroll container",
			signature:   "hscroll(content: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *container.Scroll)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "hscroll(label('wide')) => GoObj[fyne.CanvasObject](Value: *container.Scroll)",
		}.String(),
	},
	{
		Name: "_vscroll",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("vscroll", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("vscroll", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("vscroll", 1, "fyne.CanvasObject", args[0])
			}
			return NewGoObj[fyne.CanvasObject](container.NewVScroll(c.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`vscroll` wraps content in a vertical-only scroll container",
			signature:   "vscroll(content: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *container.Scroll)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "vscroll(col([...])) => GoObj[fyne.CanvasObject](Value: *container.Scroll)",
		}.String(),
	},
	{
		Name: "_hsplit",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("hsplit", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("hsplit", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("hsplit", 2, GO_OBJ, args[1].Type())
			}
			a, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("hsplit", 1, "fyne.CanvasObject", args[0])
			}
			b, ok := args[1].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("hsplit", 2, "fyne.CanvasObject", args[1])
			}
			return NewGoObj[fyne.CanvasObject](container.NewHSplit(a.Value, b.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`hsplit` returns a horizontally split container with draggable divider (left/right resizable)",
			signature:   "hsplit(left: GoObj[fyne.CanvasObject], right: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *container.Split)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "hsplit(label('left'), label('right')) => GoObj[fyne.CanvasObject](Value: *container.Split)",
		}.String(),
	},
	{
		Name: "_vsplit",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("vsplit", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("vsplit", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("vsplit", 2, GO_OBJ, args[1].Type())
			}
			a, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("vsplit", 1, "fyne.CanvasObject", args[0])
			}
			b, ok := args[1].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("vsplit", 2, "fyne.CanvasObject", args[1])
			}
			return NewGoObj[fyne.CanvasObject](container.NewVSplit(a.Value, b.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`vsplit` returns a vertically split container with draggable divider (top/bottom resizable)",
			signature:   "vsplit(top: GoObj[fyne.CanvasObject], bottom: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *container.Split)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "vsplit(label('top'), label('bottom')) => GoObj[fyne.CanvasObject](Value: *container.Split)",
		}.String(),
	},
	{
		Name: "_border",
		Fun: func(args ...Object) Object {
			if len(args) != 5 {
				return newInvalidArgCountError("border", len(args), 5, "")
			}
			toCanvas := func(o Object, pos int) (fyne.CanvasObject, Object) {
				if o.Type() != GO_OBJ {
					// allow null sentinel for optional border slots
					if o.Type() == NULL_OBJ {
						return nil, nil
					}
					return nil, newPositionalTypeError("border", pos, GO_OBJ+" or null", o.Type())
				}
				co, ok := o.(*GoObj[fyne.CanvasObject])
				if !ok {
					return nil, newPositionalTypeErrorForGoObj("border", pos, "fyne.CanvasObject", o)
				}
				return co.Value, nil
			}
			top, err := toCanvas(args[0], 1)
			if err != nil {
				return err
			}
			bottom, err := toCanvas(args[1], 2)
			if err != nil {
				return err
			}
			left, err := toCanvas(args[2], 3)
			if err != nil {
				return err
			}
			right, err := toCanvas(args[3], 4)
			if err != nil {
				return err
			}
			if args[4].Type() != GO_OBJ {
				return newPositionalTypeError("border", 5, GO_OBJ, args[4].Type())
			}
			center, ok := args[4].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("border", 5, "fyne.CanvasObject", args[4])
			}
			return NewGoObj[fyne.CanvasObject](container.NewBorder(top, bottom, left, right, center.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`border` returns a border layout container (optional top/bottom/left/right plus mandatory center content)",
			signature:   "border(top: GoObj[fyne.CanvasObject]|null, bottom: GoObj[fyne.CanvasObject]|null, left: GoObj[fyne.CanvasObject]|null, right: GoObj[fyne.CanvasObject]|null, center: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *fyne.Container)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "border(label('top'), null, null, null, label('center')) => GoObj[fyne.CanvasObject](Value: *fyne.Container)",
		}.String(),
	},
	{
		Name: "_center",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("center", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("center", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("center", 1, "fyne.CanvasObject", args[0])
			}
			return NewGoObj[fyne.CanvasObject](container.NewCenter(c.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`center` returns a centered container with the content centered both horizontally and vertically",
			signature:   "center(content: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *fyne.Container)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "center(label('hi')) => GoObj[fyne.CanvasObject](Value: *fyne.Container)",
		}.String(),
	},
	{
		Name: "_padded",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("padded", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("padded", 1, GO_OBJ, args[0].Type())
			}
			c, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("padded", 1, "fyne.CanvasObject", args[0])
			}
			return NewGoObj[fyne.CanvasObject](container.NewPadded(c.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`padded` returns a container with theme padding around the content",
			signature:   "padded(content: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *fyne.Container)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "padded(label('hi')) => GoObj[fyne.CanvasObject](Value: *fyne.Container)",
		}.String(),
	},
	{
		Name: "_stack",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("stack", len(args), 1, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("stack", 1, LIST_OBJ, args[0].Type())
			}
			elems := args[0].(*List).Elements
			objs := make([]fyne.CanvasObject, len(elems))
			for i, e := range elems {
				if e.Type() != GO_OBJ {
					return newError("`stack` error: all children should be GO_OBJ[fyne.CanvasObject]. found=%s", e.Type())
				}
				co, ok := e.(*GoObj[fyne.CanvasObject])
				if !ok {
					return newPositionalTypeErrorForGoObj("stack", i+1, "fyne.CanvasObject", e)
				}
				objs[i] = co.Value
			}
			return NewGoObj[fyne.CanvasObject](container.NewStack(objs...))
		},
		HelpStr: helpStrArgs{
			explanation: "`stack` returns a stacked container (children drawn on top of each other)",
			signature:   "stack(children: list[GoObj[fyne.CanvasObject]]) -> GoObj[fyne.CanvasObject](Value: *fyne.Container)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "stack([label('a'), label('b')]) => GoObj[fyne.CanvasObject](Value: *fyne.Container)",
		}.String(),
	},
	{
		Name: "_new_color",
		Fun: func(args ...Object) Object {
			if len(args) != 4 {
				return newInvalidArgCountError("new_color", len(args), 4, "")
			}
			for i, a := range args {
				if a.Type() != INTEGER_OBJ {
					return newPositionalTypeError("new_color", i+1, INTEGER_OBJ, a.Type())
				}
			}
			r := uint8(args[0].(*Integer).Value)
			g := uint8(args[1].(*Integer).Value)
			b := uint8(args[2].(*Integer).Value)
			aVal := uint8(args[3].(*Integer).Value)
			col := color.NRGBA{R: r, G: g, B: b, A: aVal}
			return NewGoObj[color.Color](color.Color(col))
		},
		HelpStr: helpStrArgs{
			explanation: "`new_color` creates a color from RGBA bytes (0-255 each) for use with canvas primitives",
			signature:   "new_color(r: int, g: int, b: int, a: int) -> GoObj[color.Color]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "new_color(255, 0, 0, 255) => GoObj[color.Color](Value: color.NRGBA{R:255 G:0 ...})",
		}.String(),
	},
	{
		Name: "_canvas_rectangle",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("canvas_rectangle", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_rectangle", 1, GO_OBJ, args[0].Type())
			}
			co, ok := args[0].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_rectangle", 1, "color.Color", args[0])
			}
			return NewGoObj[fyne.CanvasObject](canvas.NewRectangle(co.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_rectangle` creates a canvas rectangle primitive filled with the given color",
			signature:   "canvas_rectangle(color: GoObj[color.Color]) -> GoObj[fyne.CanvasObject](Value: *canvas.Rectangle)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "canvas_rectangle(new_color(255,0,0,255)) => GoObj[fyne.CanvasObject](Value: *canvas.Rectangle)",
		}.String(),
	},
	{
		Name: "_canvas_circle",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("canvas_circle", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_circle", 1, GO_OBJ, args[0].Type())
			}
			co, ok := args[0].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_circle", 1, "color.Color", args[0])
			}
			return NewGoObj[fyne.CanvasObject](canvas.NewCircle(co.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_circle` creates a canvas circle primitive filled with the given color",
			signature:   "canvas_circle(color: GoObj[color.Color]) -> GoObj[fyne.CanvasObject](Value: *canvas.Circle)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "canvas_circle(new_color(0,255,0,255)) => GoObj[fyne.CanvasObject](Value: *canvas.Circle)",
		}.String(),
	},
	{
		Name: "_canvas_line",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("canvas_line", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_line", 1, GO_OBJ, args[0].Type())
			}
			co, ok := args[0].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_line", 1, "color.Color", args[0])
			}
			return NewGoObj[fyne.CanvasObject](canvas.NewLine(co.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_line` creates a canvas line primitive stroked with the given color",
			signature:   "canvas_line(color: GoObj[color.Color]) -> GoObj[fyne.CanvasObject](Value: *canvas.Line)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "canvas_line(new_color(0,0,255,255)) => GoObj[fyne.CanvasObject](Value: *canvas.Line)",
		}.String(),
	},
	{
		Name: "_canvas_text",
		Fun: func(args ...Object) Object {
			if len(args) != 2 && len(args) != 3 {
				return newInvalidArgCountError("canvas_text", len(args), 2, "or 3")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("canvas_text", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_text", 2, GO_OBJ, args[1].Type())
			}
			txt := args[0].(*Stringo).Value
			co, ok := args[1].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_text", 2, "color.Color", args[1])
			}
			t := canvas.NewText(txt, co.Value)
			if len(args) == 3 {
				if args[2].Type() != INTEGER_OBJ && args[2].Type() != FLOAT_OBJ {
					return newPositionalTypeError("canvas_text", 3, "INTEGER or FLOAT", args[2].Type())
				}
				var sz float32
				if args[2].Type() == INTEGER_OBJ {
					sz = float32(args[2].(*Integer).Value)
				} else {
					sz = float32(args[2].(*Float).Value)
				}
				t.TextSize = sz
			}
			return NewGoObj[fyne.CanvasObject](t)
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_text` creates a canvas text primitive with optional text size",
			signature:   "canvas_text(text: str, color: GoObj[color.Color], size: int|float=theme.TextSize) -> GoObj[fyne.CanvasObject](Value: *canvas.Text)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "canvas_text('hello', new_color(0,0,0,255), 14) => GoObj[fyne.CanvasObject](Value: *canvas.Text)",
		}.String(),
	},
	{
		Name: "_rich_text",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("rich_text", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("rich_text", 1, STRING_OBJ, args[0].Type())
			}
			txt := args[0].(*Stringo).Value
			rt := widget.NewRichTextFromMarkdown(txt)
			return NewGoObj[fyne.CanvasObject](rt)
		},
		HelpStr: helpStrArgs{
			explanation: "`rich_text` creates a rich text widget rendering the given markdown string (headings, bold, links, etc.)",
			signature:   "rich_text(markdown: str) -> GoObj[fyne.CanvasObject](Value: *widget.RichText)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "rich_text('# Hello') => GoObj[fyne.CanvasObject](Value: *widget.RichText)",
		}.String(),
	},
	{
		Name: "_slider_get_value",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("slider_get_value", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("slider_get_value", 1, GO_OBJ, args[0].Type())
			}
			slider, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("slider_get_value", 1, "fyne.CanvasObject", args[0])
			}
			switch x := slider.Value.(type) {
			case *widget.Slider:
				return &Float{Value: x.Value}
			default:
				return newError("`slider_get_value` error: object is not a slider. got=%T", x)
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`slider_get_value` returns the current float value of a slider widget",
			signature:   "slider_get_value(s: GoObj[fyne.CanvasObject](Value: *widget.Slider)) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "slider_get_value(s) => 0.5",
		}.String(),
	},
	{
		Name: "_slider_set_value",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("slider_set_value", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("slider_set_value", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ && args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("slider_set_value", 2, "FLOAT or INTEGER", args[1].Type())
			}
			slider, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("slider_set_value", 1, "fyne.CanvasObject", args[0])
			}
			var v float64
			if args[1].Type() == FLOAT_OBJ {
				v = args[1].(*Float).Value
			} else {
				v = float64(args[1].(*Integer).Value)
			}
			switch x := slider.Value.(type) {
			case *widget.Slider:
				x.SetValue(v)
				return NULL
			default:
				return newError("`slider_set_value` error: object is not a slider. got=%T", x)
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`slider_set_value` sets the float value of a slider widget",
			signature:   "slider_set_value(s: GoObj[fyne.CanvasObject](Value: *widget.Slider), value: float|int) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "slider_set_value(s, 1.0) => null",
		}.String(),
	},
	{
		Name: "_slider",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`slider` returns a slider widget with min, max, and an on_change handler receiving the float value",
			signature:   "slider(min: float|int, max: float|int, value: float|int, fn: fun(value: float)) -> GoObj[fyne.CanvasObject](Value: *widget.Slider)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "slider(0, 100, 50, |v| => {println(v)}) => GoObj[fyne.CanvasObject](Value: *widget.Slider)",
		}.String(),
	},
	{
		Name: "_check_group",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`check_group` returns a check_group widget (multi-select checkboxes) with options and an on_change handler receiving list[str]",
			signature:   "check_group(options: list[str], fn: fun(selected: list[str])) -> GoObj[fyne.CanvasObject](Value: *widget.CheckGroup)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "check_group(['a','b'], |v| => {println(v)}) => GoObj[fyne.CanvasObject](Value: *widget.CheckGroup)",
		}.String(),
	},
	{
		Name: "_select_entry",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`select_entry` returns an editable select entry (combo box) with options and an on_change handler",
			signature:   "select_entry(options: list[str], fn: fun(value: str)) -> GoObj[fyne.CanvasObject](Value: *widget.SelectEntry)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "select_entry(['a','b'], |v| => {println(v)}) => GoObj[fyne.CanvasObject](Value: *widget.SelectEntry)",
		}.String(),
	},
	{
		Name: "_hyperlink_with_handler",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`hyperlink_with_handler` returns a hyperlink widget with a custom tap handler instead of opening URL",
			signature:   "hyperlink_with_handler(text: str, url: str, fn: fun()) -> GoObj[fyne.CanvasObject](Value: *widget.Hyperlink)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "hyperlink_with_handler('click', 'https://example.com', || => {println('tapped')}) => GoObj[fyne.CanvasObject](Value: *widget.Hyperlink)",
		}.String(),
	},
	{
		Name: "_canvas_image",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("canvas_image", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("canvas_image", 1, STRING_OBJ, args[0].Type())
			}
			path := args[0].(*Stringo).Value
			img := canvas.NewImageFromFile(path)
			img.FillMode = canvas.ImageFillContain
			return NewGoObj[fyne.CanvasObject](img)
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_image` creates an image canvas object from a file path (PNG/JPEG/SVG)",
			signature:   "canvas_image(path: str) -> GoObj[fyne.CanvasObject](Value: *canvas.Image)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "canvas_image('./photo.png') => GoObj[fyne.CanvasObject](Value: *canvas.Image)",
		}.String(),
	},
	{
		Name: "_canvas_arc",
		Fun: func(args ...Object) Object {
			if len(args) != 4 {
				return newInvalidArgCountError("canvas_arc", len(args), 4, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_arc", 1, GO_OBJ, args[0].Type())
			}
			co, ok := args[0].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_arc", 1, "color.Color", args[0])
			}
			parseFloat := func(o Object, pos int) (float32, Object) {
				if o.Type() == FLOAT_OBJ {
					return float32(o.(*Float).Value), nil
				}
				if o.Type() == INTEGER_OBJ {
					return float32(o.(*Integer).Value), nil
				}
				return 0, newPositionalTypeError("canvas_arc", pos, "FLOAT or INTEGER", o.Type())
			}
			start, err := parseFloat(args[1], 2)
			if err != nil {
				return err
			}
			end, err := parseFloat(args[2], 3)
			if err != nil {
				return err
			}
			cutout, err := parseFloat(args[3], 4)
			if err != nil {
				return err
			}
			return NewGoObj[fyne.CanvasObject](canvas.NewArc(start, end, cutout, co.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_arc` creates an arc/circle-sector canvas primitive (degrees, cutout 0..1)",
			signature:   "canvas_arc(color: GoObj[color.Color], startAngle: float|int, endAngle: float|int, cutout: float|int) -> GoObj[fyne.CanvasObject](Value: *canvas.Arc)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "canvas_arc(new_color(255,0,0,255), 0, 270, 0.3) => GoObj[fyne.CanvasObject](Value: *canvas.Arc)",
		}.String(),
	},
	{
		Name: "_canvas_polygon",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("canvas_polygon", len(args), 2, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("canvas_polygon", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_polygon", 2, GO_OBJ, args[1].Type())
			}
			sides := uint(args[0].(*Integer).Value)
			if sides < 3 {
				return newError("`canvas_polygon` error: sides must be >=3. got=%d", sides)
			}
			co, ok := args[1].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_polygon", 2, "color.Color", args[1])
			}
			return NewGoObj[fyne.CanvasObject](canvas.NewPolygon(sides, co.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_polygon` creates a regular polygon canvas primitive (n sides)",
			signature:   "canvas_polygon(sides: int, color: GoObj[color.Color]) -> GoObj[fyne.CanvasObject](Value: *canvas.Polygon)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "canvas_polygon(6, new_color(0,0,255,255)) => GoObj[fyne.CanvasObject](Value: *canvas.Polygon)",
		}.String(),
	},
	{
		Name: "_text_grid",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("text_grid", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("text_grid", 1, STRING_OBJ, args[0].Type())
			}
			txt := args[0].(*Stringo).Value
			return NewGoObj[fyne.CanvasObject](widget.NewTextGridFromString(txt))
		},
		HelpStr: helpStrArgs{
			explanation: "`text_grid` creates a monospace text grid widget (terminal/code view) from a string",
			signature:   "text_grid(content: str) -> GoObj[fyne.CanvasObject](Value: *widget.TextGrid)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "text_grid('hello\\nworld') => GoObj[fyne.CanvasObject](Value: *widget.TextGrid)",
		}.String(),
	},
	{
		Name: "_activity",
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("activity", len(args), 0, "")
			}
			a := widget.NewActivity()
			a.Start()
			return NewGoObj[fyne.CanvasObject](a)
		},
		HelpStr: helpStrArgs{
			explanation: "`activity` creates an activity indicator (spinning dots, auto-started)",
			signature:   "activity() -> GoObj[fyne.CanvasObject](Value: *widget.Activity)",
			errors:      "InvalidArgCount",
			example:     "activity() => GoObj[fyne.CanvasObject](Value: *widget.Activity)",
		}.String(),
	},
	{
		Name: "_activity_start",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("activity_start", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("activity_start", 1, GO_OBJ, args[0].Type())
			}
			aw, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("activity_start", 1, "fyne.CanvasObject", args[0])
			}
			switch x := aw.Value.(type) {
			case *widget.Activity:
				x.Start()
				return NULL
			default:
				return newError("`activity_start` error: not an activity. got=%T", x)
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`activity_start` starts the activity indicator animation",
			signature:   "activity_start(a: GoObj[fyne.CanvasObject](Value: *widget.Activity)) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "activity_start(a) => null",
		}.String(),
	},
	{
		Name: "_activity_stop",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("activity_stop", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("activity_stop", 1, GO_OBJ, args[0].Type())
			}
			aw, ok := args[0].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("activity_stop", 1, "fyne.CanvasObject", args[0])
			}
			switch x := aw.Value.(type) {
			case *widget.Activity:
				x.Stop()
				return NULL
			default:
				return newError("`activity_stop` error: not an activity. got=%T", x)
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`activity_stop` stops the activity indicator animation",
			signature:   "activity_stop(a: GoObj[fyne.CanvasObject](Value: *widget.Activity)) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "activity_stop(a) => null",
		}.String(),
	},
	{
		Name: "_inner_window",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("inner_window", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("inner_window", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("inner_window", 2, GO_OBJ, args[1].Type())
			}
			title := args[0].(*Stringo).Value
			c, ok := args[1].(*GoObj[fyne.CanvasObject])
			if !ok {
				return newPositionalTypeErrorForGoObj("inner_window", 2, "fyne.CanvasObject", args[1])
			}
			return NewGoObj[fyne.CanvasObject](container.NewInnerWindow(title, c.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`inner_window` creates a draggable inner window with title and content",
			signature:   "inner_window(title: str, content: GoObj[fyne.CanvasObject]) -> GoObj[fyne.CanvasObject](Value: *container.InnerWindow)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "inner_window('Title', label('hi')) => GoObj[fyne.CanvasObject](Value: *container.InnerWindow)",
		}.String(),
	},
	{
		Name: "_file_icon",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("file_icon", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("file_icon", 1, STRING_OBJ, args[0].Type())
			}
			uriStr := args[0].(*Stringo).Value
			// try parse as URI first, fallback to file path
			parsed, err := storage.ParseURI(uriStr)
			if err != nil || parsed == nil {
				// treat as plain file path
				parsed, err = storage.ParseURI("file://" + uriStr)
				if err != nil || parsed == nil {
					return newError("`file_icon` error: invalid uri `%s`: %s", uriStr, err.Error())
				}
			}
			return NewGoObj[fyne.CanvasObject](widget.NewFileIcon(parsed))
		},
		HelpStr: helpStrArgs{
			explanation: "`file_icon` creates an icon widget for a file URI (auto picks icon by extension)",
			signature:   "file_icon(uri: str) -> GoObj[fyne.CanvasObject](Value: *widget.FileIcon)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "file_icon('file:///tmp/photo.png') => GoObj[fyne.CanvasObject](Value: *widget.FileIcon)",
		}.String(),
	},
	{
		Name: "_calendar",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`calendar` returns a calendar widget with onChanged handler receiving the selected date string (YYYY-MM-DD)",
			signature:   "calendar(initialDate: str='now', fn: fun(date: str)) -> GoObj[fyne.CanvasObject](Value: *widget.Calendar)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "calendar('2026-01-01', |d| => {println(d)}) => GoObj[fyne.CanvasObject](Value: *widget.Calendar)",
		}.String(),
	},
	{
		Name: "_list",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`list` returns a virtualized scrolling list of strings with onSelected handler receiving (index: int, value: str)",
			signature:   "list(items: list[str], fn: fun(index: int, value: str)) -> GoObj[fyne.CanvasObject](Value: *widget.List)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "list(['a','b'], |i,v| => {println(i, v)}) => GoObj[fyne.CanvasObject](Value: *widget.List)",
		}.String(),
	},
	{
		Name: "_table",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`table` returns a table widget from a 2D string array with onSelected handler receiving (row: int, col: int, value: str)",
			signature:   "table(data: list[list[str]], fn: fun(row: int, col: int, value: str)) -> GoObj[fyne.CanvasObject](Value: *widget.Table)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "table([['a','b'],['c','d']], |r,c,v| => {println(r,c,v)}) => GoObj[fyne.CanvasObject](Value: *widget.Table)",
		}.String(),
	},
	{
		Name: "_canvas_linear_gradient",
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("canvas_linear_gradient", len(args), 3, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_linear_gradient", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_linear_gradient", 2, GO_OBJ, args[1].Type())
			}
			c1, ok := args[0].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_linear_gradient", 1, "color.Color", args[0])
			}
			c2, ok := args[1].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_linear_gradient", 2, "color.Color", args[1])
			}
			var angle float64
			if args[2].Type() == FLOAT_OBJ {
				angle = args[2].(*Float).Value
			} else if args[2].Type() == INTEGER_OBJ {
				angle = float64(args[2].(*Integer).Value)
			} else {
				return newPositionalTypeError("canvas_linear_gradient", 3, "FLOAT or INTEGER", args[2].Type())
			}
			return NewGoObj[fyne.CanvasObject](canvas.NewLinearGradient(c1.Value, c2.Value, angle))
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_linear_gradient` creates a linear gradient canvas object (angle in degrees)",
			signature:   "canvas_linear_gradient(start: GoObj[color.Color], end: GoObj[color.Color], angle: float|int) -> GoObj[fyne.CanvasObject](Value: *canvas.LinearGradient)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "canvas_linear_gradient(new_color(255,0,0,255), new_color(0,0,255,255), 90) => GoObj[fyne.CanvasObject](Value: *canvas.LinearGradient)",
		}.String(),
	},
	{
		Name: "_canvas_radial_gradient",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("canvas_radial_gradient", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_radial_gradient", 1, GO_OBJ, args[0].Type())
			}
			if args[1].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_radial_gradient", 2, GO_OBJ, args[1].Type())
			}
			c1, ok := args[0].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_radial_gradient", 1, "color.Color", args[0])
			}
			c2, ok := args[1].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_radial_gradient", 2, "color.Color", args[1])
			}
			return NewGoObj[fyne.CanvasObject](canvas.NewRadialGradient(c1.Value, c2.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_radial_gradient` creates a radial gradient canvas object (center outward)",
			signature:   "canvas_radial_gradient(start: GoObj[color.Color], end: GoObj[color.Color]) -> GoObj[fyne.CanvasObject](Value: *canvas.RadialGradient)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "canvas_radial_gradient(new_color(255,255,0,255), new_color(255,0,0,0)) => GoObj[fyne.CanvasObject](Value: *canvas.RadialGradient)",
		}.String(),
	},
	{
		Name: "_canvas_square",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("canvas_square", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("canvas_square", 1, GO_OBJ, args[0].Type())
			}
			co, ok := args[0].(*GoObj[color.Color])
			if !ok {
				return newPositionalTypeErrorForGoObj("canvas_square", 1, "color.Color", args[0])
			}
			return NewGoObj[fyne.CanvasObject](canvas.NewRectangle(co.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`canvas_square` creates a square rectangle (1:1 aspect) canvas primitive",
			signature:   "canvas_square(color: GoObj[color.Color]) -> GoObj[fyne.CanvasObject](Value: *canvas.Rectangle)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "canvas_square(new_color(128,0,128,255)) => GoObj[fyne.CanvasObject](Value: *canvas.Rectangle)",
		}.String(),
	},
	{
		Name: "_dialog_info",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("dialog_info", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("dialog_info", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("dialog_info", 2, STRING_OBJ, args[1].Type())
			}
			title := args[0].(*Stringo).Value
			msg := args[1].(*Stringo).Value
			a := fyne.CurrentApp()
			if a == nil {
				return NULL
			}
			fyne.Do(func() {
				w := a.NewWindow(title)
				w.SetContent(container.NewVBox(
					widget.NewLabel(msg),
					widget.NewButton("OK", func() { w.Close() }),
				))
				w.Resize(fyne.NewSize(380, 160))
				w.Show()
			})
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`dialog_info` shows a simple information dialog with OK button (non-blocking)",
			signature:   "dialog_info(title: str, message: str) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "dialog_info('Info', 'Hello') => null (shows window)",
		}.String(),
	},
	{
		Name: "_grid_wrap",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`grid_wrap` returns a wrapping grid of strings with onSelected handler receiving (index: int, value: str)",
			signature:   "grid_wrap(items: list[str], fn: fun(index: int, value: str)) -> GoObj[fyne.CanvasObject](Value: *widget.GridWrap)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "grid_wrap(['a','b','c'], |i,v| => {println(i, v)}) => GoObj[fyne.CanvasObject](Value: *widget.GridWrap)",
		}.String(),
	},
	{
		Name: "_tree",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`tree` returns a hierarchical tree widget from map[str]list[str] (parent->children) with onSelected handler receiving (uid: str)",
			signature:   "tree(data: map[str]list[str], root: str, fn: fun(uid: str)) -> GoObj[fyne.CanvasObject](Value: *widget.Tree)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "tree({ '': ['a','b'], 'a': ['a1'] }, '', |uid| => {println(uid)}) => GoObj[fyne.CanvasObject](Value: *widget.Tree)",
		}.String(),
	},
	{
		Name: "_dialog_confirm",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`dialog_confirm` shows a confirm dialog with Yes/No, calling handler with bool result",
			signature:   "dialog_confirm(title: str, message: str, fn: fun(confirmed: bool)) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "dialog_confirm('Confirm', 'Are you sure?', |v| => {println(v)}) => null",
		}.String(),
	},
	{
		Name: "_dialog_file_open",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`dialog_file_open` shows a file-open picker returning chosen path to handler (empty string if cancelled)",
			signature:   "dialog_file_open(fn: fun(path: str)) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "dialog_file_open(|p| => {println(p)}) => null",
		}.String(),
	},
	{
		Name: "_menu_item",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`menu_item` creates a fyne menu item with label and action handler",
			signature:   "menu_item(label: str, fn: fun()) -> GoObj[*fyne.MenuItem]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "menu_item('Open', || => {println('open')}) => GoObj[*fyne.MenuItem]",
		}.String(),
	},
	{
		Name: "_menu",
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("menu", len(args), 2, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("menu", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != LIST_OBJ {
				return newPositionalTypeError("menu", 2, LIST_OBJ, args[1].Type())
			}
			label := args[0].(*Stringo).Value
			elems := args[1].(*List).Elements
			items := make([]*fyne.MenuItem, len(elems))
			for i, e := range elems {
				if e.Type() != GO_OBJ {
					return newPositionalTypeError("menu", 2, GO_OBJ, e.Type())
				}
				mi, ok := e.(*GoObj[*fyne.MenuItem])
				if !ok {
					return newPositionalTypeErrorForGoObj("menu", 2, "*fyne.MenuItem", e)
				}
				items[i] = mi.Value
			}
			return NewGoObj(fyne.NewMenu(label, items...))
		},
		HelpStr: helpStrArgs{
			explanation: "`menu` creates a fyne menu with label and list of menu items",
			signature:   "menu(label: str, items: list[GoObj[*fyne.MenuItem]]) -> GoObj[*fyne.Menu]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "menu('File', [menu_item('Open', ||=>{})]) => GoObj[*fyne.Menu]",
		}.String(),
	},
	{
		Name: "_set_main_menu",
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_main_menu", len(args), 1, "")
			}
			if args[0].Type() != LIST_OBJ {
				return newPositionalTypeError("set_main_menu", 1, LIST_OBJ, args[0].Type())
			}
			elems := args[0].(*List).Elements
			menus := make([]*fyne.Menu, len(elems))
			for i, e := range elems {
				if e.Type() != GO_OBJ {
					return newPositionalTypeError("set_main_menu", 1, GO_OBJ, e.Type())
				}
				m, ok := e.(*GoObj[*fyne.Menu])
				if !ok {
					return newPositionalTypeErrorForGoObj("set_main_menu", 1, "*fyne.Menu", e)
				}
				menus[i] = m.Value
			}
			mainMenu := fyne.NewMainMenu(menus...)
			a := fyne.CurrentApp()
			if a != nil {
				if w := a.Driver().AllWindows(); len(w) > 0 {
					w[0].SetMainMenu(mainMenu)
				}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`set_main_menu` sets the application main menu from a list of menus (must be called after window creation, ideally before Show)",
			signature:   "set_main_menu(menus: list[GoObj[*fyne.Menu]]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_main_menu([menu('File', [menu_item('Quit', ||=>{})])]) => null",
		}.String(),
	},
}
