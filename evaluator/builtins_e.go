package evaluator

import (
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"
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
				if strings.Contains(s, "+Inf") {
					return &object.Float{Value: math.Inf(1)}
				} else if strings.Contains(s, "-Inf") {
					return &object.Float{Value: math.Inf(-1)}
				}
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

func getSortedListHelper(e *Evaluator, args ...object.Object) object.Object {
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
}

func createSortBuiltin(e *Evaluator) *object.Builtin {
	if sortBuiltin == nil {
		sortBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				return getSortedListHelper(e, args...)
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

var sortedBuiltin *object.Builtin = nil

func createSortedBuiltin(e *Evaluator) *object.Builtin {
	if sortedBuiltin == nil {
		sortedBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				o := getSortedListHelper(e, args...)
				if isError(o) {
					return o
				}
				l, ok := o.(*object.List)
				if !ok {
					return l
				}
				args[0].(*object.List).Elements = l.Elements
				return object.NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`sorted` sorts the given list, if its ints, floats, or strings no custom key is needed, otherwise a function returning the key to sort should be returned (ie. a str, float, or int). This function Mutates the underlying List",
				signature:   "sorted(l: list[int|float|str|any], reverse: bool=false, key: null|fun(e: list[any])=>int|str|float=null) -> list[int|float|str|any] (sorted)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "sorted(['c','b','a']) => null => side effect ['a','b','c']",
			}.String(),
		}
	}
	return sortedBuiltin
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
						obj = e.applyFunctionFast(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
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
						obj = e.applyFunctionFast(fn, []object.Object{elem}, map[string]object.Object{}, []bool{false})
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

var loadBuiltin *object.Builtin = nil

func createLoadBuiltin(e *Evaluator) *object.Builtin {
	if loadBuiltin == nil {
		loadBuiltin = &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newInvalidArgCountError("load", len(args), 1, "")
				}
				if args[0].Type() != object.BYTES_OBJ {
					return newPositionalTypeError("load", 1, object.BYTES_OBJ, args[0].Type())
				}
				obj, err := object.Decode(args[0].(*object.Bytes).Value)
				if err != nil {
					return newError("`load` error: %s", err.Error())
				}
				switch o := obj.(type) {
				case *object.Boolean:
					return nativeToBooleanObject(o.Value)
				case *object.Null:
					return object.NULL
				case *object.StringFunction:
					l := lexer.New(o.Value, "load")
					p := parser.New(l)
					prog := p.ParseProgram()
					if len(p.Errors()) != 0 {
						return newError("`load` error: failed to decode function %s", o.Value)
					}
					obj := e.Eval(prog)
					if isError(obj) {
						return newError("`load` error: %s", obj.(*object.Error).Message)
					}
					if o, ok := obj.(*object.Function); ok {
						return o
					}
					return newError("`load` error: failed to decode function %s", o.Value)
				case *object.GoObjectGob:
					// Note: This is disabled for now due to the complexity of handling all Go Object Types supported by blue
					// log.Printf("GO OBJECT = %#+v", o)
					// decoder := goObjDecoders[o.T].(func([]byte) (any, error))
					// log.Printf("%T", decoder)
					// a, err := decoder(o.Value)
					// if err != nil {
					// 	return newError("`load` error: %s", err)
					// }
					// log.Printf("t = %T, a = %+#v", a, a)
					// switch o := a.(type) {
					// case object.GoObj[color.RGBA]:
					// 	return &o
					// case *object.GoObj[*os.File]:
					// 	return o
					// default:
					// 	return newError("`load` error: %T is not handled", a)
					// }
					return newError("`load` error: Go Object %T not enabled for decoding", o)
				default:
					return obj
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`load` returns the object decoded from bytes",
				signature:   "load(arg: bytes) -> any",
				errors:      "InvalidArgCount,PositionalTypeError,CustomError",
				example:     "load('82001904d2'.to_bytes(is_hex=true)) => 1234",
			}.String(),
		}
	}
	return loadBuiltin
}
