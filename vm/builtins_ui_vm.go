//go:build !static && !wasm

package vm

import (
	"blue/consts"
	"blue/object"
	"fmt"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// getUIStdBuiltin resolves the vm backed std ui builtins. It lives behind
// the !wasm constraint because the fyne dependency has no js/wasm support.
func getUIStdBuiltin(name string, vm *VM) *object.Builtin {
	switch name {
	case "_button":
		return createUIButtonBuiltin(vm)
	case "_check_box":
		return createUICheckBoxBuiltin(vm)
	case "_radio_group":
		return createUIRadioBuiltin(vm)
	case "_option_select":
		return createUIOptionSelectBuiltin(vm)
	case "_form":
		return createUIFormBuiltin(vm)
	case "_toolbar_action":
		return createUIToolbarAction(vm)
	case "_slider":
		return createUISliderBuiltin(vm)
	case "_check_group":
		return createUICheckGroupBuiltin(vm)
	case "_select_entry":
		return createUISelectEntryBuiltin(vm)
	case "_hyperlink_with_handler":
		return createUIHyperlinkWithHandlerBuiltin(vm)
	case "_calendar":
		return createUICalendarBuiltin(vm)
	case "_list":
		return createUIListBuiltin(vm)
	case "_table":
		return createUITableBuiltin(vm)
	case "_grid_wrap":
		return createUIGridWrapBuiltin(vm)
	case "_tree":
		return createUITreeBuiltin(vm)
	case "_dialog_confirm":
		return createUIDialogConfirmBuiltin(vm)
	case "_dialog_file_open":
		return createUIDialogFileOpenBuiltin(vm)
	case "_menu_item":
		return createUIMenuItemBuiltin(vm)
	default:
		return nil
	}
}

func createUIButtonBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "button",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("button", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("button", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("button", 2, object.CLOSURE, args[1].Type())
			}
			s := args[0].(*object.Stringo).Value
			fn := args[1].(*object.Closure)
			button := widget.NewButton(s, func() {
				obj := vm.applyFunctionFast(fn, nil)
				if isError(obj) {
					err := obj.(*object.Error)
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`button` click handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](button)
		},
	}
}

func createUICheckBoxBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "checkbox",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("checkbox", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("checkbox", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("checkbox", 2, object.CLOSURE, args[1].Type())
			}
			lbl := args[0].(*object.Stringo).Value
			fn := args[1].(*object.Closure)
			if len(fn.Fun.Parameters) != 1 {
				return newError("`checkbox` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			checkBox := widget.NewCheck(lbl, func(value bool) {
				obj := vm.applyFunctionFast(fn, nativeToBooleanObject(value))
				if isError(obj) {
					err := obj.(*object.Error)
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`check_box` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](checkBox)
		},
	}
}

func createUIRadioBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "radio_group",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("radio_group", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("radio_group", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("radio_group", 2, object.CLOSURE, args[1].Type())
			}
			elems := args[0].(*object.List).Elements
			fn := args[1].(*object.Closure)
			options := make([]string, len(elems))
			for i, e := range elems {
				if e.Type() != object.STRING_OBJ {
					return newError("`radio_group` error: all elements in list should be STRING. found=%s", e.Type())
				}
				options[i] = e.(*object.Stringo).Value
			}
			if len(fn.Fun.Parameters) != 1 {
				return newError("`radio_group` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			radio := widget.NewRadioGroup(options, func(value string) {
				obj := vm.applyFunctionFast(fn, &object.Stringo{Value: value})
				if isError(obj) {
					err := obj.(*object.Error)
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`radio_group` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](radio)
		},
	}
}

func createUIOptionSelectBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "option_select",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("option_select", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("option_select", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("option_select", 2, object.CLOSURE, args[1].Type())
			}
			elems := args[0].(*object.List).Elements
			fn := args[1].(*object.Closure)
			options := make([]string, len(elems))
			for i, e := range elems {
				if e.Type() != object.STRING_OBJ {
					return newError("`option_select` error: all elements in list should be STRING. found=%s", e.Type())
				}
				options[i] = e.(*object.Stringo).Value
			}
			if len(fn.Fun.Parameters) != 1 {
				return newError("`option_select` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			option := widget.NewSelect(options, func(value string) {
				obj := vm.applyFunctionFast(fn, &object.Stringo{Value: value})
				if isError(obj) {
					err := obj.(*object.Error)
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`option_select` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](option)
		},
	}
}

func createUIFormBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "form",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("form", len(args), 3, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("form", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.LIST_OBJ {
				return newPositionalTypeError("form", 2, object.LIST_OBJ, args[1].Type())
			}
			if args[2].Type() != object.CLOSURE {
				return newPositionalTypeError("form", 3, object.CLOSURE, args[2].Type())
			}
			var formItems []*widget.FormItem
			labels := args[0].(*object.List).Elements
			widgetIds := args[1].(*object.List).Elements
			if len(labels) != len(widgetIds) {
				return newError("`form` error: labels and widget ids must be the same length. len(labels)=%d, len(widgetIds)=%d", len(labels), len(widgetIds))
			}
			fn := args[2].(*object.Closure)
			for i := range labels {
				if labels[i].Type() != object.STRING_OBJ {
					return newError("`form` error: labels were not all STRINGs. found=%s", labels[i].Type())
				}
				if widgetIds[i].Type() != object.GO_OBJ {
					return newError("`form` error: widgetIds were not all GO_OBJs. found=%s", widgetIds[i].Type())
				}
				w, ok := widgetIds[i].(*object.GoObj[fyne.CanvasObject])
				if !ok {
					return newPositionalTypeErrorForGoObj("form", i+1, "fyne.CanvasObject", w)
				}
				formItem := &widget.FormItem{
					Text: labels[i].(*object.Stringo).Value, Widget: w.Value,
				}

				formItems = append(formItems, formItem)
			}

			form := &widget.Form{
				Items: formItems,
				OnSubmit: func() {
					obj := vm.applyFunctionFast(fn, nil)
					if isError(obj) {
						err := obj.(*object.Error)
						// var buf bytes.Buffer
						// buf.WriteString(err.Message)
						// buf.WriteByte('\n')
						// for e.ErrorTokens.Len() > 0 {
						// 	tok := e.ErrorTokens.PopBack()
						// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
						// }
						fmt.Printf("%s`form` on_submit error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
					}
				},
			}
			return NewGoObj[fyne.CanvasObject](form)
		},
	}
}

func createUIToolbarAction(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "toolbar_action",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("toolbar_action", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("toolbar_action", 1, object.GO_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("toolbar_action", 2, object.CLOSURE, args[1].Type())
			}
			r, ok := args[0].(*object.GoObj[fyne.Resource])
			if !ok {
				return newPositionalTypeErrorForGoObj("toolbar_action", 1, "fyne.Resource", args[0])
			}
			fn := args[1].(*object.Closure)
			return NewGoObj[widget.ToolbarItem](widget.NewToolbarAction(r.Value, func() {
				obj := vm.applyFunctionFast(fn, nil)
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`toolbar_action` click handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			}))
		},
	}
}

func createUISliderBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "slider",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("slider", len(args), 4, "")
			}
			var min, max, val float64
			for i := 0; i < 3; i++ {
				if args[i].Type() == object.FLOAT_OBJ {
					switch i {
					case 0:
						min = args[i].(*object.Float).Value
					case 1:
						max = args[i].(*object.Float).Value
					case 2:
						val = args[i].(*object.Float).Value
					}
				} else if args[i].Type() == object.INTEGER_OBJ {
					switch i {
					case 0:
						min = float64(args[i].(*object.Integer).Value)
					case 1:
						max = float64(args[i].(*object.Integer).Value)
					case 2:
						val = float64(args[i].(*object.Integer).Value)
					}
				} else {
					return newPositionalTypeError("slider", i+1, "FLOAT or INTEGER", args[i].Type())
				}
			}
			if args[3].Type() != object.CLOSURE {
				return newPositionalTypeError("slider", 4, object.CLOSURE, args[3].Type())
			}
			fn := args[3].(*object.Closure)
			if len(fn.Fun.Parameters) != 1 {
				return newError("`slider` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			slider := widget.NewSlider(min, max)
			slider.Value = val
			slider.OnChanged = func(v float64) {
				obj := vm.applyFunctionFast(fn, &object.Float{Value: v})
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`slider` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			}
			return NewGoObj[fyne.CanvasObject](slider)
		},
	}
}

func createUICheckGroupBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "check_group",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("check_group", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("check_group", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("check_group", 2, object.CLOSURE, args[1].Type())
			}
			elems := args[0].(*object.List).Elements
			opts := make([]string, len(elems))
			for i, e := range elems {
				if e.Type() != object.STRING_OBJ {
					return newError("`check_group` error: all elements should be STRING. found=%s", e.Type())
				}
				opts[i] = e.(*object.Stringo).Value
			}
			fn := args[1].(*object.Closure)
			if len(fn.Fun.Parameters) != 1 {
				return newError("`check_group` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			cg := widget.NewCheckGroup(opts, func(selected []string) {
				listElems := make([]object.Object, len(selected))
				for i, s := range selected {
					listElems[i] = &object.Stringo{Value: s}
				}
				obj := vm.applyFunctionFast(fn, &object.List{Elements: listElems})
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`check_group` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](cg)
		},
	}
}

func createUISelectEntryBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "select_entry",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("select_entry", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("select_entry", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("select_entry", 2, object.CLOSURE, args[1].Type())
			}
			elems := args[0].(*object.List).Elements
			opts := make([]string, len(elems))
			for i, e := range elems {
				if e.Type() != object.STRING_OBJ {
					return newError("`select_entry` error: all elements should be STRING. found=%s", e.Type())
				}
				opts[i] = e.(*object.Stringo).Value
			}
			fn := args[1].(*object.Closure)
			if len(fn.Fun.Parameters) != 1 {
				return newError("`select_entry` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			se := widget.NewSelectEntry(opts)
			se.OnChanged = func(value string) {
				obj := vm.applyFunctionFast(fn, &object.Stringo{Value: value})
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`select_entry` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			}
			return NewGoObj[fyne.CanvasObject](se)
		},
	}
}

func createUIHyperlinkWithHandlerBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "hyperlink_with_handler",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("hyperlink_with_handler", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("hyperlink_with_handler", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("hyperlink_with_handler", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.CLOSURE {
				return newPositionalTypeError("hyperlink_with_handler", 3, object.CLOSURE, args[2].Type())
			}
			text := args[0].(*object.Stringo).Value
			rawURL := args[1].(*object.Stringo).Value
			u, err := url.Parse(rawURL)
			if err != nil {
				return newError("`hyperlink_with_handler` error: invalid URL `%s`: %s", rawURL, err.Error())
			}
			fn := args[2].(*object.Closure)
			hl := widget.NewHyperlink(text, u)
			hl.OnTapped = func() {
				obj := vm.applyFunctionFast(fn, nil)
				if isError(obj) {
					e := obj.(*object.Error)
					fmt.Printf("%s`hyperlink_with_handler` handler error: %s\n", consts.VM_ERROR_PREFIX, e.Message)
				}
			}
			return NewGoObj[fyne.CanvasObject](hl)
		},
	}
}

func createUICalendarBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "calendar",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("calendar", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("calendar", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("calendar", 2, object.CLOSURE, args[1].Type())
			}
			initialStr := args[0].(*object.Stringo).Value
			fn := args[1].(*object.Closure)
			if len(fn.Fun.Parameters) != 1 {
				return newError("`calendar` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			var initial time.Time
			if initialStr == "" || initialStr == "now" {
				initial = time.Now()
			} else {
				// try YYYY-MM-DD, then RFC3339, then fallback to now
				t, err := time.Parse("2006-01-02", initialStr)
				if err != nil {
					t2, err2 := time.Parse(time.RFC3339, initialStr)
					if err2 != nil {
						return newError("`calendar` error: invalid date `%s` (expected YYYY-MM-DD or RFC3339)", initialStr)
					}
					initial = t2
				} else {
					initial = t
				}
			}
			cal := widget.NewCalendar(initial, func(t time.Time) {
				ds := t.Format("2006-01-02")
				obj := vm.applyFunctionFast(fn, &object.Stringo{Value: ds})
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`calendar` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](cal)
		},
	}
}

func createUIListBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "list",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("list", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("list", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("list", 2, object.CLOSURE, args[1].Type())
			}
			elems := args[0].(*object.List).Elements
			items := make([]string, len(elems))
			for i, e := range elems {
				if e.Type() != object.STRING_OBJ {
					return newError("`list` error: all items should be STRING. found=%s", e.Type())
				}
				items[i] = e.(*object.Stringo).Value
			}
			fn := args[1].(*object.Closure)
			if len(fn.Fun.Parameters) != 2 {
				return newError("`list` error: handler needs 2 arguments (index, value). got=%d", len(fn.Fun.Parameters))
			}
			lst := widget.NewList(
				func() int { return len(items) },
				func() fyne.CanvasObject { return widget.NewLabel("") },
				func(id widget.ListItemID, co fyne.CanvasObject) {
					lbl, ok := co.(*widget.Label)
					if !ok {
						return
					}
					if id >= 0 && id < len(items) {
						lbl.SetText(items[id])
					}
				},
			)
			lst.OnSelected = func(id widget.ListItemID) {
				var val string
				if id >= 0 && id < len(items) {
					val = items[id]
				}
				obj := vm.applyFunctionFastWithMultipleArgs(fn, []object.Object{&object.Integer{Value: int64(id)}, &object.Stringo{Value: val}})
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`list` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			}
			return NewGoObj[fyne.CanvasObject](lst)
		},
	}
}

func createUITableBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "table",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("table", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("table", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("table", 2, object.CLOSURE, args[1].Type())
			}
			outer := args[0].(*object.List).Elements
			if len(outer) == 0 {
				return newError("`table` error: data must have at least 1 row")
			}
			// validate and convert to [][]string
			data := make([][]string, len(outer))
			cols := -1
			for r, rowObj := range outer {
				if rowObj.Type() != object.LIST_OBJ {
					return newError("`table` error: row %d should be LIST. found=%s", r, rowObj.Type())
				}
				rowElems := rowObj.(*object.List).Elements
				if cols == -1 {
					cols = len(rowElems)
				} else if len(rowElems) != cols {
					return newError("`table` error: all rows must have same length. row 0=%d, row %d=%d", cols, r, len(rowElems))
				}
				data[r] = make([]string, len(rowElems))
				for c, cell := range rowElems {
					if cell.Type() != object.STRING_OBJ {
						return newError("`table` error: cell [%d][%d] should be STRING. found=%s", r, c, cell.Type())
					}
					data[r][c] = cell.(*object.Stringo).Value
				}
			}
			rows := len(data)
			fn := args[1].(*object.Closure)
			if len(fn.Fun.Parameters) != 3 {
				return newError("`table` error: handler needs 3 arguments (row, col, value). got=%d", len(fn.Fun.Parameters))
			}
			tbl := widget.NewTable(
				func() (int, int) { return rows, cols },
				func() fyne.CanvasObject { return widget.NewLabel("") },
				func(id widget.TableCellID, co fyne.CanvasObject) {
					lbl, ok := co.(*widget.Label)
					if !ok {
						return
					}
					if id.Row >= 0 && id.Row < rows && id.Col >= 0 && id.Col < cols {
						lbl.SetText(data[id.Row][id.Col])
					}
				},
			)
			tbl.OnSelected = func(id widget.TableCellID) {
				var val string
				if id.Row >= 0 && id.Row < rows && id.Col >= 0 && id.Col < cols {
					val = data[id.Row][id.Col]
				}
				obj := vm.applyFunctionFastWithMultipleArgs(fn, []object.Object{&object.Integer{Value: int64(id.Row)}, &object.Integer{Value: int64(id.Col)}, &object.Stringo{Value: val}})
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`table` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			}
			return NewGoObj[fyne.CanvasObject](tbl)
		},
	}
}

func createUIGridWrapBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "grid_wrap",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("grid_wrap", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("grid_wrap", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("grid_wrap", 2, object.CLOSURE, args[1].Type())
			}
			elems := args[0].(*object.List).Elements
			items := make([]string, len(elems))
			for i, e := range elems {
				if e.Type() != object.STRING_OBJ {
					return newError("`grid_wrap` error: all items should be STRING. found=%s", e.Type())
				}
				items[i] = e.(*object.Stringo).Value
			}
			fn := args[1].(*object.Closure)
			if len(fn.Fun.Parameters) != 2 {
				return newError("`grid_wrap` error: handler needs 2 arguments (index, value). got=%d", len(fn.Fun.Parameters))
			}
			gw := widget.NewGridWrap(
				func() int { return len(items) },
				func() fyne.CanvasObject { return widget.NewLabel("") },
				func(id widget.GridWrapItemID, co fyne.CanvasObject) {
					lbl, ok := co.(*widget.Label)
					if !ok {
						return
					}
					if id >= 0 && id < len(items) {
						lbl.SetText(items[id])
					}
				},
			)
			gw.OnSelected = func(id widget.GridWrapItemID) {
				var val string
				if id >= 0 && id < len(items) {
					val = items[id]
				}
				obj := vm.applyFunctionFastWithMultipleArgs(fn, []object.Object{&object.Integer{Value: int64(id)}, &object.Stringo{Value: val}})
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`grid_wrap` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			}
			return NewGoObj[fyne.CanvasObject](gw)
		},
	}
}

func createUITreeBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "tree",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("tree", len(args), 3, "")
			}
			if args[0].Type() != object.MAP_OBJ {
				return newPositionalTypeError("tree", 1, object.MAP_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("tree", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.CLOSURE {
				return newPositionalTypeError("tree", 3, object.CLOSURE, args[2].Type())
			}
			m := args[0].(*object.Map)
			root := args[1].(*object.Stringo).Value
			fn := args[2].(*object.Closure)
			if len(fn.Fun.Parameters) != 1 {
				return newError("`tree` error: handler needs 1 argument (uid). got=%d", len(fn.Fun.Parameters))
			}
			// Convert Blue map[str]list[str] to Go map[string][]string
			goMap := make(map[string][]string)
			for _, hk := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(hk)
				keyStr, ok := mp.Key.(*object.Stringo)
				if !ok {
					return newError("`tree` error: all keys must be STRING. got=%s", mp.Key.Type())
				}
				if mp.Value.Type() != object.LIST_OBJ {
					return newError("`tree` error: value for key `%s` must be LIST. got=%s", keyStr.Value, mp.Value.Type())
				}
				listElems := mp.Value.(*object.List).Elements
				arr := make([]string, len(listElems))
				for i, e := range listElems {
					if e.Type() != object.STRING_OBJ {
						return newError("`tree` error: children for `%s` must be STRING. found=%s", keyStr.Value, e.Type())
					}
					arr[i] = e.(*object.Stringo).Value
				}
				goMap[keyStr.Value] = arr
			}
			tr := widget.NewTreeWithStrings(goMap)
			// Override root if specified and not empty
			if root != "" {
				tr.Root = root
			}
			tr.OnSelected = func(uid widget.TreeNodeID) {
				obj := vm.applyFunctionFast(fn, &object.Stringo{Value: uid})
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`tree` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			}
			return NewGoObj[fyne.CanvasObject](tr)
		},
	}
}

func createUIDialogConfirmBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "dialog_confirm",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("dialog_confirm", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("dialog_confirm", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("dialog_confirm", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.CLOSURE {
				return newPositionalTypeError("dialog_confirm", 3, object.CLOSURE, args[2].Type())
			}
			title := args[0].(*object.Stringo).Value
			msg := args[1].(*object.Stringo).Value
			fn := args[2].(*object.Closure)
			if len(fn.Fun.Parameters) != 1 {
				return newError("`dialog_confirm` error: handler needs 1 argument (bool). got=%d", len(fn.Fun.Parameters))
			}
			a := fyne.CurrentApp()
			if a == nil {
				return newError("`dialog_confirm` error: no current app (window not yet created?)")
			}
			fyne.Do(func() {
				w := a.NewWindow(title)
				w.SetContent(container.NewVBox(
					widget.NewLabel(msg),
					container.NewHBox(
						widget.NewButton("Yes", func() {
							w.Close()
							obj := vm.applyFunctionFast(fn, object.TRUE)
							if isError(obj) {
								err := obj.(*object.Error)
								fmt.Printf("%s`dialog_confirm` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
							}
						}),
						widget.NewButton("No", func() {
							w.Close()
							obj := vm.applyFunctionFast(fn, object.FALSE)
							if isError(obj) {
								err := obj.(*object.Error)
								fmt.Printf("%s`dialog_confirm` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
							}
						}),
					),
				))
				w.Resize(fyne.NewSize(380, 140))
				w.Show()
			})
			return object.NULL
		},
	}
}

func createUIDialogFileOpenBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "dialog_file_open",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("dialog_file_open", len(args), 1, "")
			}
			if args[0].Type() != object.CLOSURE {
				return newPositionalTypeError("dialog_file_open", 1, object.CLOSURE, args[0].Type())
			}
			fn := args[0].(*object.Closure)
			if len(fn.Fun.Parameters) != 1 {
				return newError("`dialog_file_open` error: handler needs 1 argument (path). got=%d", len(fn.Fun.Parameters))
			}
			a := fyne.CurrentApp()
			if a == nil {
				return newError("`dialog_file_open` error: no current app")
			}
			fyne.Do(func() {
				w := a.NewWindow("Open File")
				entry := widget.NewEntry()
				entry.SetPlaceHolder("Enter file path")
				w.SetContent(container.NewVBox(
					widget.NewLabel("File path:"),
					entry,
					container.NewHBox(
						widget.NewButton("Open", func() {
							path := entry.Text
							w.Close()
							obj := vm.applyFunctionFast(fn, &object.Stringo{Value: path})
							if isError(obj) {
								err := obj.(*object.Error)
								fmt.Printf("%s`dialog_file_open` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
							}
						}),
						widget.NewButton("Cancel", func() {
							w.Close()
							obj := vm.applyFunctionFast(fn, &object.Stringo{Value: ""})
							if isError(obj) {
								err := obj.(*object.Error)
								fmt.Printf("%s`dialog_file_open` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
							}
						}),
					),
				))
				w.Resize(fyne.NewSize(520, 160))
				w.Show()
			})
			return object.NULL
		},
	}
}

func createUIMenuItemBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "menu_item",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("menu_item", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("menu_item", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("menu_item", 2, object.CLOSURE, args[1].Type())
			}
			label := args[0].(*object.Stringo).Value
			fn := args[1].(*object.Closure)
			mi := fyne.NewMenuItem(label, func() {
				obj := vm.applyFunctionFast(fn, nil)
				if isError(obj) {
					err := obj.(*object.Error)
					fmt.Printf("%s`menu_item` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj(mi)
		},
	}
}
