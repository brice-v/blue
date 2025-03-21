//go:build !static
// +build !static

package evaluator

import (
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"bytes"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// UI Builtins

var uiButtonBuiltin *object.Builtin = nil

func createUIButtonBuiltin(e *Evaluator) *object.Builtin {
	if uiButtonBuiltin == nil {
		uiButtonBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newInvalidArgCountError("button", len(args), 2, "")
				}
				if args[0].Type() != object.STRING_OBJ {
					return newPositionalTypeError("button", 1, object.STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != object.FUNCTION_OBJ {
					return newPositionalTypeError("button", 2, object.FUNCTION_OBJ, args[1].Type())
				}
				s := args[0].(*object.Stringo).Value
				fn := args[1].(*object.Function)
				button := widget.NewButton(s, func() {
					obj := e.applyFunctionFast(fn, []object.Object{}, make(map[string]object.Object), []bool{})
					if isError(obj) {
						err := obj.(*object.Error)
						var buf bytes.Buffer
						buf.WriteString(err.Message)
						buf.WriteByte('\n')
						for e.ErrorTokens.Len() > 0 {
							tok := e.ErrorTokens.PopBack()
							buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
						}
						fmt.Printf("%s`button` click handler error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
					}
				})
				return NewGoObj[fyne.CanvasObject](button)
			},
			HelpStr: helpStrArgs{
				explanation: "`button` returns a ui button widget object with a string label and an onclick function handler",
				signature:   "button(label: str, fn: fun()) -> GoObj[fyne.CanvasObject](Value: *widget.Button)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "button('Click Me!', || => {println('clicked')}) => GoObj[fyne.CanvasObject](Value: *widget.Button)",
			}.String(),
		}
	}
	return uiButtonBuiltin
}

var uiCheckboxBuiltin *object.Builtin = nil

func createUICheckBoxBuiltin(e *Evaluator) *object.Builtin {
	if uiCheckboxBuiltin == nil {
		uiCheckboxBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newInvalidArgCountError("checkbox", len(args), 2, "")
				}
				if args[0].Type() != object.STRING_OBJ {
					return newPositionalTypeError("checkbox", 1, object.STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != object.FUNCTION_OBJ {
					return newPositionalTypeError("checkbox", 2, object.FUNCTION_OBJ, args[1].Type())
				}
				lbl := args[0].(*object.Stringo).Value
				fn := args[1].(*object.Function)
				if len(fn.Parameters) != 1 {
					return newError("`checkbox` error: handler needs 1 argument. got=%d", len(fn.Parameters))
				}
				checkBox := widget.NewCheck(lbl, func(value bool) {
					obj := e.applyFunctionFast(fn, []object.Object{nativeToBooleanObject(value)}, make(map[string]object.Object), []bool{true})
					if isError(obj) {
						err := obj.(*object.Error)
						var buf bytes.Buffer
						buf.WriteString(err.Message)
						buf.WriteByte('\n')
						for e.ErrorTokens.Len() > 0 {
							tok := e.ErrorTokens.PopBack()
							buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
						}
						fmt.Printf("%s`check_box` handler error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
					}
				})
				return NewGoObj[fyne.CanvasObject](checkBox)
			},
			HelpStr: helpStrArgs{
				explanation: "`check_box` returns a ui check_box widget object with a string label and an onchecked function handler",
				signature:   "check_box(label: str, fn: fun(is_checked: bool)) -> GoObj[fyne.CanvasObject](Value: *widget.Check)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "check_box('Check Me!', |e| => {println('checked? #{e}')}) => GoObj[fyne.CanvasObject](Value: *widget.Check)",
			}.String(),
		}
	}
	return uiCheckboxBuiltin
}

var uiRadioButtonBuiltin *object.Builtin = nil

func createUIRadioBuiltin(e *Evaluator) *object.Builtin {
	if uiRadioButtonBuiltin == nil {
		uiRadioButtonBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newInvalidArgCountError("radio_group", len(args), 2, "")
				}
				if args[0].Type() != object.LIST_OBJ {
					return newPositionalTypeError("radio_group", 1, object.LIST_OBJ, args[0].Type())
				}
				if args[1].Type() != object.FUNCTION_OBJ {
					return newPositionalTypeError("radio_group", 2, object.FUNCTION_OBJ, args[1].Type())
				}
				elems := args[0].(*object.List).Elements
				fn := args[1].(*object.Function)
				options := make([]string, len(elems))
				for i, e := range elems {
					if e.Type() != object.STRING_OBJ {
						return newError("`radio_group` error: all elements in list should be STRING. found=%s", e.Type())
					}
					options[i] = e.(*object.Stringo).Value
				}
				if len(fn.Parameters) != 1 {
					return newError("`radio_group` error: handler needs 1 argument. got=%d", len(fn.Parameters))
				}
				radio := widget.NewRadioGroup(options, func(value string) {
					obj := e.applyFunctionFast(fn, []object.Object{&object.Stringo{Value: value}}, make(map[string]object.Object), []bool{true})
					if isError(obj) {
						err := obj.(*object.Error)
						var buf bytes.Buffer
						buf.WriteString(err.Message)
						buf.WriteByte('\n')
						for e.ErrorTokens.Len() > 0 {
							tok := e.ErrorTokens.PopBack()
							buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
						}
						fmt.Printf("%s`radio_group` handler error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
					}
				})
				return NewGoObj[fyne.CanvasObject](radio)
			},
			HelpStr: helpStrArgs{
				explanation: "`radio_group` returns a ui radio_group widget object with a list of string radio labels and an onchecked function handler",
				signature:   "radio_group(labels: list[str], fn: fun(checked_label: str)) -> GoObj[fyne.CanvasObject](Value: *widget.RadioGroup)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "radio_group(['Check Me 1!', 'Check Me 2!'], |e| => {println('checked #{e}')}) => GoObj[fyne.CanvasObject](Value: *widget.RadioGroup)",
			}.String(),
		}
	}
	return uiRadioButtonBuiltin
}

var uiOptionSelectBuiltin *object.Builtin = nil

func createUIOptionSelectBuiltin(e *Evaluator) *object.Builtin {
	if uiOptionSelectBuiltin == nil {
		uiOptionSelectBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newInvalidArgCountError("option_select", len(args), 2, "")
				}
				if args[0].Type() != object.LIST_OBJ {
					return newPositionalTypeError("option_select", 1, object.LIST_OBJ, args[0].Type())
				}
				if args[1].Type() != object.FUNCTION_OBJ {
					return newPositionalTypeError("option_select", 2, object.FUNCTION_OBJ, args[1].Type())
				}
				elems := args[0].(*object.List).Elements
				fn := args[1].(*object.Function)
				options := make([]string, len(elems))
				for i, e := range elems {
					if e.Type() != object.STRING_OBJ {
						return newError("`option_select` error: all elements in list should be STRING. found=%s", e.Type())
					}
					options[i] = e.(*object.Stringo).Value
				}
				if len(fn.Parameters) != 1 {
					return newError("`option_select` error: handler needs 1 argument. got=%d", len(fn.Parameters))
				}
				option := widget.NewSelect(options, func(value string) {
					obj := e.applyFunctionFast(fn, []object.Object{&object.Stringo{Value: value}}, make(map[string]object.Object), []bool{true})
					if isError(obj) {
						err := obj.(*object.Error)
						var buf bytes.Buffer
						buf.WriteString(err.Message)
						buf.WriteByte('\n')
						for e.ErrorTokens.Len() > 0 {
							tok := e.ErrorTokens.PopBack()
							buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
						}
						fmt.Printf("%s`option_select` handler error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
					}
				})
				return NewGoObj[fyne.CanvasObject](option)
			},
			HelpStr: helpStrArgs{
				explanation: "`option_select` returns a ui option_select widget object with a list of string options and an onchecked function handler",
				signature:   "option_select(labels: list[str], fn: fun(checked_option: str)) -> GoObj[fyne.CanvasObject](Value: *widget.Select)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "option_select(['Check Me 1!', 'Check Me 2!'], |e| => {println('checked #{e}')}) => GoObj[fyne.CanvasObject](Value: *widget.Select)",
			}.String(),
		}
	}
	return uiOptionSelectBuiltin
}

var uiFormBuiltin *object.Builtin = nil

func createUIFormBuiltin(e *Evaluator) *object.Builtin {
	if uiFormBuiltin == nil {
		uiFormBuiltin = &object.Builtin{
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
				if args[2].Type() != object.FUNCTION_OBJ {
					return newPositionalTypeError("form", 3, object.FUNCTION_OBJ, args[2].Type())
				}
				var formItems []*widget.FormItem
				labels := args[0].(*object.List).Elements
				widgetIds := args[1].(*object.List).Elements
				if len(labels) != len(widgetIds) {
					return newError("`form` error: labels and widget ids must be the same length. len(labels)=%d, len(widgetIds)=%d", len(labels), len(widgetIds))
				}
				fn := args[2].(*object.Function)
				for i := 0; i < len(labels); i++ {
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
						obj := e.applyFunctionFast(fn, []object.Object{}, make(map[string]object.Object), []bool{})
						if isError(obj) {
							err := obj.(*object.Error)
							var buf bytes.Buffer
							buf.WriteString(err.Message)
							buf.WriteByte('\n')
							for e.ErrorTokens.Len() > 0 {
								tok := e.ErrorTokens.PopBack()
								buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
							}
							fmt.Printf("%s`form` on_submit error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
						}
					},
				}
				return NewGoObj[fyne.CanvasObject](form)
			},
			HelpStr: helpStrArgs{
				explanation: "`form` returns a ui form widget object with the given list of labels and widgets, and a submit function",
				signature:   "form(elements: list[{'label': str, 'widget': GoObj[fyne.CanvasObject]}]=[], fn: fun()) -> GoObj[fyne.CanvasObject](Value: *widget.Form)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "form(|| => {println('submit')}) => GoObj[fyne.CanvasObject](Value: *widget.Form)",
			}.String(),
		}
	}
	return uiFormBuiltin
}

var uiToolbarAction *object.Builtin = nil

func createUIToolbarAction(e *Evaluator) *object.Builtin {
	if uiToolbarAction == nil {
		uiToolbarAction = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newInvalidArgCountError("toolbar_action", len(args), 2, "")
				}
				if args[0].Type() != object.GO_OBJ {
					return newPositionalTypeError("toolbar_action", 1, object.GO_OBJ, args[0].Type())
				}
				if args[1].Type() != object.FUNCTION_OBJ {
					return newPositionalTypeError("toolbar_action", 2, object.FUNCTION_OBJ, args[1].Type())
				}
				r, ok := args[0].(*object.GoObj[fyne.Resource])
				if !ok {
					return newPositionalTypeErrorForGoObj("toolbar_action", 1, "fyne.Resource", args[0])
				}
				fn := args[1].(*object.Function)
				return NewGoObj[widget.ToolbarItem](widget.NewToolbarAction(r.Value, func() {
					obj := e.applyFunctionFast(fn, []object.Object{}, make(map[string]object.Object), []bool{})
					if isError(obj) {
						err := obj.(*object.Error)
						var buf bytes.Buffer
						buf.WriteString(err.Message)
						buf.WriteByte('\n')
						for e.ErrorTokens.Len() > 0 {
							tok := e.ErrorTokens.PopBack()
							buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
						}
						fmt.Printf("%s`toolbar_action` click handler error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
					}
				}))
			},
			HelpStr: helpStrArgs{
				explanation: "`toolbar.action()`: `toolbar_action` returns a ui toolbar_action widget object which can be added to a toolbar when given a resource a function to execute on action",
				signature:   "toolbar_action(res: GoObj[fyne.Resource], fn: fun()) -> GoObj[widget.ToolbarItem](Value: *widget.ToolbarAction)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "toolbar_action(icon.computer, || => {println('action!')}) => GoObj[widget.ToolbarItem](Value: *widget.ToolbarAction)",
			}.String(),
		}
	}
	return uiToolbarAction
}
