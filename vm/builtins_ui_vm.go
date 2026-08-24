//go:build !static && !wasm

package vm

import (
	"blue/consts"
	"blue/object"
	"fmt"

	"fyne.io/fyne/v2"
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
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`toolbar_action` click handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			}))
		},
	}
}
