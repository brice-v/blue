package evaluator

import (
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"bytes"
	"fmt"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// Core Builtins

var toNumBuiltin *object.Builtin = nil

func createToNumBuiltin(e *Evaluator) *object.Builtin {
	if toNumBuiltin == nil {
		toNumBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_num", len(args), 1, "")
				}
				if args[0].Type() != object.STRING_OBJ {
					return newPositionalTypeError("to_num", 1, object.STRING_OBJ, args[0].Type())
				}
				s := args[0].(*object.Stringo).Value
				ll := lexer.New(s, "")
				pp := parser.New(ll)
				prog := pp.ParseProgram()
				if len(pp.Errors()) != 0 {
					return newError("`to_num` error: failed to parse number from string '%s'", s)
				}
				obj := e.Eval(prog)
				if isError(obj) {
					return obj
				}
				if obj.Type() != object.INTEGER_OBJ && obj.Type() != object.UINTEGER_OBJ && obj.Type() != object.FLOAT_OBJ && obj.Type() != object.BIG_FLOAT_OBJ && obj.Type() != object.BIG_INTEGER_OBJ {
					return newError("`to_num` error: failed to get number type from string '%s'. got=%s", s, obj.Type())
				}
				return obj
			},
			HelpStr: helpStrArgs{
				explanation: "`to_num` returns the NUM value of the given STRING (int, uint, float, bigint, bigfloat)",
				signature:   "to_num(arg: str) -> num",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "to_num('1') => 1",
			}.String(),
		}
	}
	return toNumBuiltin
}

var sortBuiltin *object.Builtin = nil

func simpleKeyErrorPrint(e *Evaluator, obj object.Object) {
	err := obj.(*object.Error)
	var buf bytes.Buffer
	buf.WriteString(err.Message)
	buf.WriteByte('\n')
	for e.ErrorTokens.Len() > 0 {
		tok := e.ErrorTokens.PopBack()
		buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
	}
	fmt.Printf("%s`sort` key error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
}

func createSortBuiltin(e *Evaluator) *object.Builtin {
	if sortBuiltin == nil {
		sortBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return newInvalidArgCountError("sort", len(args), 3, "")
				}
				if args[0].Type() != object.LIST_OBJ {
					return newPositionalTypeError("sort", 1, object.LIST_OBJ, args[0].Type())
				}
				if args[1].Type() != object.BOOLEAN_OBJ {
					return newPositionalTypeError("sort", 2, object.BOOLEAN_OBJ, args[1].Type())
				}
				if args[2].Type() != object.NULL_OBJ && args[2].Type() != object.FUNCTION_OBJ && args[2].Type() != object.LIST_OBJ {
					return newPositionalTypeError("sort", 3, object.FUNCTION_OBJ+" or null", args[2].Type())
				}
				l := args[0].(*object.List)
				shouldReverse := args[1].(*object.Boolean).Value
				if args[2].Type() == object.NULL_OBJ {
					allInts := true
					allFloats := true
					allStrings := true
					for _, e := range l.Elements {
						allInts = allInts && e.Type() == object.INTEGER_OBJ
						allFloats = allFloats && e.Type() == object.FLOAT_OBJ
						allStrings = allStrings && e.Type() == object.STRING_OBJ
					}
					if !allStrings && !allFloats && !allInts {
						return newError("`sort` error: all elements in list must be STRING, INTEGER, or FLOAT")
					}
					newElems := make([]object.Object, len(l.Elements))
					if allStrings {
						strs := make([]string, len(l.Elements))
						for i, e := range l.Elements {
							strs[i] = e.(*object.Stringo).Value
						}
						if shouldReverse {
							sort.Stable(sort.Reverse(sort.StringSlice(strs)))
						} else {
							sort.Strings(strs)
						}
						for i, e := range strs {
							newElems[i] = &object.Stringo{Value: e}
						}
					}
					if allInts {
						ints := make([]int, len(l.Elements))
						for i, e := range l.Elements {
							ints[i] = int(e.(*object.Integer).Value)
						}
						if shouldReverse {
							sort.Stable(sort.Reverse(sort.IntSlice(ints)))
						} else {
							sort.Ints(ints)
						}
						for i, e := range ints {
							newElems[i] = &object.Integer{Value: int64(e)}
						}
					}
					if allFloats {
						floats := make([]float64, len(l.Elements))
						for i, e := range l.Elements {
							floats[i] = e.(*object.Float).Value
						}
						if shouldReverse {
							sort.Stable(sort.Reverse(sort.Float64Slice(floats)))
						} else {
							sort.Float64s(floats)
						}
						for i, e := range floats {
							newElems[i] = &object.Float{Value: e}
						}
					}
					return &object.List{Elements: newElems}
				}
				var funs []*object.Function
				if args[2].Type() == object.LIST_OBJ {
					ll := args[2].(*object.List)
					funs = make([]*object.Function, len(ll.Elements))
					for i, e := range ll.Elements {
						if e.Type() != object.FUNCTION_OBJ {
							return newError("`sort` key error: all elemments must be function")
						}
						fun := e.(*object.Function)
						if len(fun.Parameters) != 1 {
							return newError("`sort` key error: each key function must take 1 arg. got=%d for index %d", len(fun.Parameters), i)
						}
						funs[i] = fun
					}
				} else {
					fun := args[2].(*object.Function)
					funs = []*object.Function{fun}
					if len(fun.Parameters) != 1 {
						return newError("`sort` key error: key function must take 1 arg. got=%d", len(fun.Parameters))
					}
				}
				// Using custom comparator
				anys := make([]interface{}, len(l.Elements))
				for i, e := range l.Elements {
					obj, err := blueObjectToGoObject(e)
					if err != nil {
						return newError("`sort` key error: %s", err.Error())
					}
					anys[i] = obj
				}

				sorter := func(i, j int) bool {
					ai := anys[i]
					aj := anys[j]
					aib, err := goObjectToBlueObject(ai)
					if err != nil {
						fmt.Printf("%s`sort` key error: %s\n", consts.EVAL_ERROR_PREFIX, err.Error())
						return false
					}
					ajb, err := goObjectToBlueObject(aj)
					if err != nil {
						fmt.Printf("%s`sort` key error: %s\n", consts.EVAL_ERROR_PREFIX, err.Error())
						return false
					}
					for k := 0; k < len(funs); k++ {
						biObj := e.applyFunctionFast(funs[k], []object.Object{aib}, make(map[string]object.Object), []bool{false})
						if isError(biObj) {
							simpleKeyErrorPrint(e, biObj)
							return false
						}
						if biObj.Type() != object.FLOAT_OBJ && biObj.Type() != object.INTEGER_OBJ && biObj.Type() != object.STRING_OBJ {
							fmt.Printf("%s`sort` key error: key function must return INTEGER, STRING, or FLOAT\n", consts.EVAL_ERROR_PREFIX)
							return false
						}
						bjObj := e.applyFunctionFast(funs[k], []object.Object{ajb}, make(map[string]object.Object), []bool{false})
						if isError(bjObj) {
							simpleKeyErrorPrint(e, bjObj)
							return false
						}
						if bjObj.Type() != object.FLOAT_OBJ && bjObj.Type() != object.INTEGER_OBJ && bjObj.Type() != object.STRING_OBJ {
							fmt.Printf("%s`sort` key error: key function must return INTEGER, STRING, or FLOAT\n", consts.EVAL_ERROR_PREFIX)
							return false
						}
						left, err := blueObjectToGoObject(biObj)
						if err != nil {
							fmt.Printf("%s`sort` key error: key function returned error: %s\n", consts.EVAL_ERROR_PREFIX, err.Error())
							return false
						}
						right, err := blueObjectToGoObject(bjObj)
						if err != nil {
							fmt.Printf("%s`sort` key error: key function returned error: %s\n", consts.EVAL_ERROR_PREFIX, err.Error())
							return false
						}
						if leftO, ok := left.(int64); ok {
							if rightO, ok := right.(int64); ok {
								if shouldReverse {
									if leftO == rightO && k != len(funs)-1 {
										continue
									}
									return leftO > rightO
								}
								if leftO == rightO && k != len(funs)-1 {
									continue
								}
								return leftO < rightO
							}
						}
						if leftO, ok := left.(int); ok {
							if rightO, ok := right.(int); ok {
								if shouldReverse {
									if leftO == rightO && k != len(funs)-1 {
										continue
									}
									return leftO > rightO
								}
								if leftO == rightO && k != len(funs)-1 {
									continue
								}
								return leftO < rightO
							}
						}
						if leftO, ok := left.(float64); ok {
							if rightO, ok := right.(float64); ok {
								if shouldReverse {
									if leftO == rightO && k != len(funs)-1 {
										continue
									}
									return leftO > rightO
								}
								if leftO == rightO && k != len(funs)-1 {
									continue
								}
								return leftO < rightO
							}
						}
						if leftO, ok := left.(string); ok {
							if rightO, ok := right.(string); ok {
								if shouldReverse {
									if leftO == rightO && k != len(funs)-1 {
										continue
									}
									return leftO > rightO
								}
								if leftO == rightO && k != len(funs)-1 {
									continue
								}
								return leftO < rightO
							}
						}
						fmt.Printf("%s`sort` key error: key function returned mismatched types: i = %d (%T), j = %d (%T)\n", consts.EVAL_ERROR_PREFIX, i, left, j, right)
						return false
					}
					fmt.Printf("%s`sort` key error: reached end of for loop, this should not happen\n", consts.EVAL_ERROR_PREFIX)
					return false
				}
				sort.SliceStable(anys, sorter)
				newObjs := make([]object.Object, len(l.Elements))
				for i, e := range anys {
					obj, err := goObjectToBlueObject(e)
					if err != nil {
						return newError("`sort` key error: %s", err.Error())
					}
					newObjs[i] = obj
				}
				return &object.List{Elements: newObjs}
			},
			HelpStr: helpStrArgs{
				explanation: "`sort` sorts the given list, if its ints, floats, or strings no custom key is needed, otherwise a function returning the key to sort should be returned (ie. a str, float, or int)",
				signature:   "sort(l: list[int|float|str|any], reverse: bool=false, key: null|fun(e: list[any])=>int|str|float=null) -> list[int|float|str|any] (sorted)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "sort(['c','b','a']) => ['a','b','c']",
			}.String(),
		}
	}
	return sortBuiltin
}

var allBuiltin *object.Builtin = nil

func createAllBuiltin(e *Evaluator) *object.Builtin {
	if allBuiltin == nil {
		allBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newInvalidArgCountError("all", len(args), 2, "")
				}
				if args[0].Type() != object.LIST_OBJ {
					return newPositionalTypeError("all", 1, object.LIST_OBJ, args[0].Type())
				}
				if args[1].Type() != object.FUNCTION_OBJ && args[1].Type() != object.BUILTIN_OBJ {
					return newPositionalTypeError("all", 2, object.FUNCTION_OBJ+" or BUILTIN", args[1].Type())
				}
				l := args[0].(*object.List)
				allTrue := true
				if args[1].Type() == object.FUNCTION_OBJ {
					fn := args[1].(*object.Function)
					if len(fn.Parameters) != 1 {
						return newError("`all` error: function must have 1 parameter")
					}
					for _, elem := range l.Elements {
						obj := e.applyFunctionFast(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
						if isError(obj) {
							errMsg := obj.(*object.Error).Message
							return newError("`all` error: %s", errMsg)
						}
						if obj.Type() != object.BOOLEAN_OBJ {
							return newError("`all` error: function must return boolean")
						}
						allTrue = allTrue && obj.(*object.Boolean).Value
					}
				} else {
					fn := args[1].(*object.Builtin)
					for _, elem := range l.Elements {
						obj := e.applyFunction(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
						if isError(obj) {
							errMsg := obj.(*object.Error).Message
							return newError("`all` error: %s", errMsg)
						}
						if obj.Type() != object.BOOLEAN_OBJ {
							return newError("`all` error: function must return boolean")
						}
						allTrue = allTrue && obj.(*object.Boolean).Value
					}
				}
				return nativeToBooleanObject(allTrue)
			},
			HelpStr: helpStrArgs{
				explanation: "`all` returns the true if all the elements in the list return true for the given function",
				signature:   "all(arg: list[any], f: fun(e: any)=>bool) -> bool",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "all([1,2,3], |e| => e > 0) => true",
			}.String(),
		}
	}
	return allBuiltin
}

var anyBuiltin *object.Builtin = nil

func createAnyBuiltin(e *Evaluator) *object.Builtin {
	if anyBuiltin == nil {
		anyBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newInvalidArgCountError("any", len(args), 2, "")
				}
				if args[0].Type() != object.LIST_OBJ {
					return newPositionalTypeError("any", 1, object.LIST_OBJ, args[0].Type())
				}
				if args[1].Type() != object.FUNCTION_OBJ && args[1].Type() != object.BUILTIN_OBJ {
					return newPositionalTypeError("any", 2, object.FUNCTION_OBJ+" or BUILTIN", args[1].Type())
				}
				l := args[0].(*object.List)
				anyTrue := false
				if args[1].Type() == object.FUNCTION_OBJ {
					fn := args[1].(*object.Function)
					if len(fn.Parameters) != 1 {
						return newError("`any` error: function must have 1 parameter")
					}
					for _, elem := range l.Elements {
						obj := e.applyFunctionFast(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
						if isError(obj) {
							errMsg := obj.(*object.Error).Message
							return newError("`any` error: %s", errMsg)
						}
						if obj.Type() != object.BOOLEAN_OBJ {
							return newError("`any` error: function must return boolean")
						}
						anyTrue = anyTrue || obj.(*object.Boolean).Value
					}
				} else {
					fn := args[1].(*object.Builtin)
					for _, elem := range l.Elements {
						obj := e.applyFunction(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
						if isError(obj) {
							errMsg := obj.(*object.Error).Message
							return newError("`any` error: %s", errMsg)
						}
						if obj.Type() != object.BOOLEAN_OBJ {
							return newError("`any` error: function must return boolean")
						}
						anyTrue = anyTrue || obj.(*object.Boolean).Value
					}
				}
				return nativeToBooleanObject(anyTrue)
			},
			HelpStr: helpStrArgs{
				explanation: "`any` returns the true if any of the elements in the list return true for the given function",
				signature:   "any(arg: list[any], f: fun(e: any)=>bool) -> bool",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "any([1,2,3], |e| => e > 0) => true",
			}.String(),
		}
	}
	return anyBuiltin
}

var mapBuiltin *object.Builtin = nil

func createMapBuiltin(e *Evaluator) *object.Builtin {
	if mapBuiltin == nil {
		mapBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newInvalidArgCountError("map", len(args), 2, "")
				}
				if args[0].Type() != object.LIST_OBJ {
					return newPositionalTypeError("map", 1, object.LIST_OBJ, args[0].Type())
				}
				if args[1].Type() != object.FUNCTION_OBJ && args[1].Type() != object.BUILTIN_OBJ {
					return newPositionalTypeError("map", 2, object.FUNCTION_OBJ+" or BUILTIN", args[1].Type())
				}
				isBuiltin := args[1].Type() == object.BUILTIN_OBJ
				l := args[0].(*object.List)
				newElements := make([]object.Object, len(l.Elements))
				for i, elem := range l.Elements {
					var obj object.Object
					if !isBuiltin {
						fn := args[1].(*object.Function)
						obj = e.applyFunctionFast(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
					} else {
						fn := args[1].(*object.Builtin)
						obj = e.applyFunction(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
					}
					if isError(obj) {
						errMsg := obj.(*object.Error).Message
						return newError("`map` error: %s", errMsg)
					}
					newElements[i] = obj
				}
				return &object.List{Elements: newElements}
			},
			HelpStr: helpStrArgs{
				explanation: "`map` returns the a new list with the given function mapped to each element",
				signature:   "map(arg: list[any], f: fun(e: any)=>any) -> list[any]",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "map([1,2,3], |e| => e + 1) => [2,3,4]",
			}.String(),
		}
	}
	return mapBuiltin
}

var filterBuiltin *object.Builtin = nil

func createFilterBuiltin(e *Evaluator) *object.Builtin {
	if filterBuiltin == nil {
		filterBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newInvalidArgCountError("filter", len(args), 2, "")
				}
				if args[0].Type() != object.LIST_OBJ {
					return newPositionalTypeError("filter", 1, object.LIST_OBJ, args[0].Type())
				}
				if args[1].Type() != object.FUNCTION_OBJ && args[1].Type() != object.BUILTIN_OBJ {
					return newPositionalTypeError("filter", 2, object.FUNCTION_OBJ+" or BUILTIN", args[1].Type())
				}
				isBuiltin := args[1].Type() == object.BUILTIN_OBJ
				l := args[0].(*object.List)
				newElements := []object.Object{}
				for _, elem := range l.Elements {
					var obj object.Object
					if !isBuiltin {
						fn := args[1].(*object.Function)
						obj = e.applyFunctionFast(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
					} else {
						fn := args[1].(*object.Builtin)
						obj = e.applyFunction(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
					}
					if isError(obj) {
						errMsg := obj.(*object.Error).Message
						return newError("`filter` error: %s", errMsg)
					}
					if isTruthy(obj) {
						newElements = append(newElements, elem)
					}
				}
				return &object.List{Elements: newElements}
			},
			HelpStr: helpStrArgs{
				explanation: "`filter` returns the a new list with the elements that return true on the given function",
				signature:   "filter(arg: list[any], f: fun(e: any)=>bool|any) -> list[any]",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "filter([1,2,3], |e| => e > 1) => [2,3]",
			}.String(),
		}
	}
	return filterBuiltin
}

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
