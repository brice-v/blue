//go:build !static
// +build !static

package object

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var UiBuiltins = NewBuiltinSliceType{
	{
		Name: "_new_app",
		Builtin: &Builtin{
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
	},
	{
		Name: "_window",
		Builtin: &Builtin{
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
				w.Resize(fyne.Size{Width: float32(width), Height: float32(height)})
				w.SetContent(content.Value)
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
	},
	{
		Name: "_label",
		Builtin: &Builtin{
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
	},
	{
		Name: "_progress_bar",
		Builtin: &Builtin{
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
	},
	{
		Name: "_progress_bar_set_value",
		Builtin: &Builtin{
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
	},
	{
		Name: "_toolbar",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) == 0 {
					return newInvalidArgCountError("toolbar", len(args), 1, "or more")
				}
				tis := []widget.ToolbarItem{}
				for i, arg := range args {
					if arg.Type() != GO_OBJ {
						return newPositionalTypeError("toolbar", i+1, GO_OBJ, arg.Type())
					}
					ti, ok := args[0].(*GoObj[widget.ToolbarItem])
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
	},
	{
		Name: "_toolbar_spacer",
		Builtin: &Builtin{
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
	},
	{
		Name: "_toolbar_separator",
		Builtin: &Builtin{
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
	},
	{
		Name: "_row",
		Builtin: &Builtin{
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
	},
	{
		Name: "_col",
		Builtin: &Builtin{
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
	},
	{
		Name: "_grid",
		Builtin: &Builtin{
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
	},
	{
		Name: "_entry",
		Builtin: &Builtin{
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
	},
	{
		Name: "_entry_get_text",
		Builtin: &Builtin{
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
	},
	{
		Name: "_entry_set_text",
		Builtin: &Builtin{
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
	},
	{
		Name: "_append_form",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_account",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_cancel",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_check_button_checked",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_check_button",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_color_achromatic",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_color_chromatic",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_color_palette",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_computer",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_confirm",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_content_add",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_content_clear",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_content_copy",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_content_cut",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_content_paste",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_content_redo",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_content_remove",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_content_undo",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_delete",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_document_create",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_document",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_document_print",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_document_save",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_download",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_error",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_file_application",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_file_audio",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_file",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_file_image",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_file_text",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_file_video",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_folder",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_folder_new",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_folder_open",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_grid",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_help",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_history",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_home",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_info",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_list",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_login",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_logout",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_mail_attachment",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_mail_compose",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_mail_forward",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_mail_reply_all",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_mail_reply",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_mail_send",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_fast_forward",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_fast_rewind",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_music",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_pause",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_photo",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_play",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_record",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_replay",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_skip_next",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_skip_previous",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_stop",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_media_video",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_menu_drop_down",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_menu_drop_up",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_menu_expand",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_menu",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_more_horizontal",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_more_vertical",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_move_down",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_move_up",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_navigate_back",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_navigate_next",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_question",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_radio_button_checked",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_radio_button",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_search",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_search_replace",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_settings",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_storage",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_upload",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_view_full_screen",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_view_refresh",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_view_restore",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_visibility",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_visibility_off",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_volume_down",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_volume_mute",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_volume_up",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_warning",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_zoom_fit",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_zoom_in",
		Builtin: &Builtin{
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
	},
	{
		Name: "_icon_zoom_out",
		Builtin: &Builtin{
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
	},
	// Functions that must be setup with evaluator/vm reference
	{
		Name: "_button",
		Builtin: &Builtin{
			Fun: nil,
			HelpStr: helpStrArgs{
				explanation: "`button` returns a ui button widget object with a string label and an onclick function handler",
				signature:   "button(label: str, fn: fun()) -> GoObj[fyne.CanvasObject](Value: *widget.Button)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "button('Click Me!', || => {println('clicked')}) => GoObj[fyne.CanvasObject](Value: *widget.Button)",
			}.String(),
		},
	},
	{
		Name: "_check_box",
		Builtin: &Builtin{
			Fun: nil,
			HelpStr: helpStrArgs{
				explanation: "`check_box` returns a ui check_box widget object with a string label and an onchecked function handler",
				signature:   "check_box(label: str, fn: fun(is_checked: bool)) -> GoObj[fyne.CanvasObject](Value: *widget.Check)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "check_box('Check Me!', |e| => {println('checked? #{e}')}) => GoObj[fyne.CanvasObject](Value: *widget.Check)",
			}.String(),
		},
	},
	{
		Name: "_radio_group",
		Builtin: &Builtin{
			Fun: nil,
			HelpStr: helpStrArgs{
				explanation: "`radio_group` returns a ui radio_group widget object with a list of string radio labels and an onchecked function handler",
				signature:   "radio_group(labels: list[str], fn: fun(checked_label: str)) -> GoObj[fyne.CanvasObject](Value: *widget.RadioGroup)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "radio_group(['Check Me 1!', 'Check Me 2!'], |e| => {println('checked #{e}')}) => GoObj[fyne.CanvasObject](Value: *widget.RadioGroup)",
			}.String(),
		},
	},
	{
		Name: "_option_select",
		Builtin: &Builtin{
			Fun: nil,
			HelpStr: helpStrArgs{
				explanation: "`option_select` returns a ui option_select widget object with a list of string options and an onchecked function handler",
				signature:   "option_select(labels: list[str], fn: fun(checked_option: str)) -> GoObj[fyne.CanvasObject](Value: *widget.Select)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "option_select(['Check Me 1!', 'Check Me 2!'], |e| => {println('checked #{e}')}) => GoObj[fyne.CanvasObject](Value: *widget.Select)",
			}.String(),
		},
	},
	{
		Name: "_form",
		Builtin: &Builtin{
			Fun: nil,
			HelpStr: helpStrArgs{
				explanation: "`form` returns a ui form widget object with the given list of labels and widgets, and a submit function",
				signature:   "form(elements: list[{'label': str, 'widget': GoObj[fyne.CanvasObject]}]=[], fn: fun()) -> GoObj[fyne.CanvasObject](Value: *widget.Form)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "form(|| => {println('submit')}) => GoObj[fyne.CanvasObject](Value: *widget.Form)",
			}.String(),
		},
	},
	{
		Name: "_toolbar_action",
		Builtin: &Builtin{
			Fun: nil,
			HelpStr: helpStrArgs{
				explanation: "`toolbar.action()`: `toolbar_action` returns a ui toolbar_action widget object which can be added to a toolbar when given a resource a function to execute on action",
				signature:   "toolbar_action(res: GoObj[fyne.Resource], fn: fun()) -> GoObj[widget.ToolbarItem](Value: *widget.ToolbarAction)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "toolbar_action(icon.computer, || => {println('action!')}) => GoObj[widget.ToolbarItem](Value: *widget.ToolbarAction)",
			}.String(),
		},
	},
}
