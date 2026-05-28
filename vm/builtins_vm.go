package vm

import (
	"blue/blueutil"
	"blue/code"
	"blue/consts"
	"blue/object"
	"bytes"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/gookit/color"
)

type helpStrArgs struct {
	explanation string
	signature   string
	errors      string
	example     string
}

func (hsa helpStrArgs) String() string {
	return fmt.Sprintf("%s\n    Signature:  %s\n    Error(s):   %s\n    Example(s): %s\n", hsa.explanation, hsa.signature, hsa.errors, hsa.example)
}

// Core Builtins

var strBuiltinFun func(args ...object.Object) object.Object = nil

func createStrBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	if strBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		strBuiltinFun = func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("str", len(args), 1, "")
			}
			if args[0].Type() == object.BYTES_OBJ {
				return &object.Stringo{Value: string(args[0].(*object.Bytes).Value)}
			}
			return &object.Stringo{Value: vm.CustomInspect(args[0])}
		}
	}
	return strBuiltinFun
}

var printBuiltinFun func(args ...object.Object) object.Object = nil

func printHelper(vm *VM, useLn bool, args ...object.Object) object.Object {
	useColorPrinter := false
	var style color.Style
	for i, arg := range args {
		if i == 0 {
			t, s, ok := object.GetBasicObjectForGoObj[color.Style](arg)
			if ok && t == "color" {
				// Use color printer
				useColorPrinter = true
				style = s
				continue
			} else {
				useColorPrinter = false
			}
		}
		inspectedStr := vm.CustomInspect(arg)
		if useColorPrinter {
			if useLn {
				style.Println(inspectedStr)
			} else {
				style.Print(inspectedStr)
			}
		} else {
			if useLn {
				fmt.Println(inspectedStr)
			} else {
				fmt.Print(inspectedStr)
			}
		}
	}
	return object.NULL
}

func createPrintBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	if printBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		printBuiltinFun = func(args ...object.Object) object.Object {
			return printHelper(vm, false, args...)
		}
	}
	return printBuiltinFun
}

var printLnBuiltinFun func(args ...object.Object) object.Object = nil

func createPrintLnBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	if printLnBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		printLnBuiltinFun = func(args ...object.Object) object.Object {
			return printHelper(vm, true, args...)
		}
	}
	return printLnBuiltinFun
}

var toNumBuiltinFun func(args ...object.Object) object.Object = nil

func createToNumBuiltinFun() func(args ...object.Object) object.Object {
	if toNumBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		toNumBuiltinFun = func(args ...object.Object) object.Object {
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
			obj := vmStr(s)
			if isError(obj) {
				return obj
			}
			if obj.Type() != object.INTEGER_OBJ && obj.Type() != object.UINTEGER_OBJ && obj.Type() != object.FLOAT_OBJ && obj.Type() != object.BIG_FLOAT_OBJ && obj.Type() != object.BIG_INTEGER_OBJ {
				return newError("`to_num` error: failed to get number type from string '%s'. got=%s", s, obj.Type())
			}
			return obj
		}
	}
	return toNumBuiltinFun
}

var sortBuiltinFun func(args ...object.Object) object.Object = nil

func simpleKeyErrorPrint(obj object.Object) {
	err := obj.(*object.Error)
	var buf bytes.Buffer
	buf.WriteString(err.Message)
	buf.WriteByte('\n')
	// for e.ErrorTokens.Len() > 0 {
	// 	tok := e.ErrorTokens.PopBack()
	// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
	// }
	fmt.Printf("%s`sort` key error: %s\n", consts.VM_ERROR_PREFIX, buf.String())
}

func getSortedListHelper(vm *VM, args ...object.Object) object.Object {
	if len(args) != 3 {
		return newInvalidArgCountError("sort", len(args), 3, "")
	}
	if args[0].Type() != object.LIST_OBJ {
		return newPositionalTypeError("sort", 1, object.LIST_OBJ, args[0].Type())
	}
	if args[1].Type() != object.BOOLEAN_OBJ {
		return newPositionalTypeError("sort", 2, object.BOOLEAN_OBJ, args[1].Type())
	}
	if args[2].Type() != object.NULL_OBJ && args[2].Type() != object.CLOSURE && args[2].Type() != object.LIST_OBJ {
		return newPositionalTypeError("sort", 3, object.CLOSURE+" or null", args[2].Type())
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
	var funs []*object.Closure
	if args[2].Type() == object.LIST_OBJ {
		ll := args[2].(*object.List)
		funs = make([]*object.Closure, len(ll.Elements))
		for i, e := range ll.Elements {
			if e.Type() != object.CLOSURE {
				return newError("`sort` key error: all elemments must be function")
			}
			fun := e.(*object.Closure)
			if len(fun.Fun.Parameters) != 1 {
				return newError("`sort` key error: each key function must take 1 arg. got=%d for index %d", len(fun.Fun.Parameters), i)
			}
			funs[i] = fun
		}
	} else {
		fun := args[2].(*object.Closure)
		funs = []*object.Closure{fun}
		if len(fun.Fun.Parameters) != 1 {
			return newError("`sort` key error: key function must take 1 arg. got=%d", len(fun.Fun.Parameters))
		}
	}
	// Using custom comparator
	anys := make([]any, len(l.Elements))
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
			fmt.Printf("%s`sort` key error: %s\n", consts.VM_ERROR_PREFIX, err.Error())
			return false
		}
		ajb, err := goObjectToBlueObject(aj)
		if err != nil {
			fmt.Printf("%s`sort` key error: %s\n", consts.VM_ERROR_PREFIX, err.Error())
			return false
		}
		for k := 0; k < len(funs); k++ {
			biObj := vm.applyFunctionFast(funs[k], aib)
			if isError(biObj) {
				simpleKeyErrorPrint(biObj)
				return false
			}
			if biObj.Type() != object.FLOAT_OBJ && biObj.Type() != object.INTEGER_OBJ && biObj.Type() != object.STRING_OBJ {
				fmt.Printf("%s`sort` ||| key error: key function must return INTEGER, STRING, or FLOAT. got = %T (%s)\n", consts.VM_ERROR_PREFIX, biObj, biObj.Inspect())
				return false
			}
			bjObj := vm.applyFunctionFast(funs[k], ajb)
			if isError(bjObj) {
				simpleKeyErrorPrint(bjObj)
				return false
			}
			if bjObj.Type() != object.FLOAT_OBJ && bjObj.Type() != object.INTEGER_OBJ && bjObj.Type() != object.STRING_OBJ {
				fmt.Printf("%s`sort` key error: key function must return INTEGER, STRING, or FLOAT. got = %T (%s)\n", consts.VM_ERROR_PREFIX, bjObj, bjObj.Inspect())
				return false
			}
			left, err := blueObjectToGoObject(biObj)
			if err != nil {
				fmt.Printf("%s`sort` key error: key function returned error: %s\n", consts.VM_ERROR_PREFIX, err.Error())
				return false
			}
			right, err := blueObjectToGoObject(bjObj)
			if err != nil {
				fmt.Printf("%s`sort` key error: key function returned error: %s\n", consts.VM_ERROR_PREFIX, err.Error())
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
			fmt.Printf("%s`sort` key error: key function returned mismatched types: i = %d (%T), j = %d (%T)\n", consts.VM_ERROR_PREFIX, i, left, j, right)
			return false
		}
		fmt.Printf("%s`sort` key error: reached end of for loop, this should not happen\n", consts.VM_ERROR_PREFIX)
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

func createSortBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	if sortBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		sortBuiltinFun = func(args ...object.Object) object.Object {
			return getSortedListHelper(vm, args...)
		}
	}
	return sortBuiltinFun
}

var sortedBuiltinFun func(args ...object.Object) object.Object = nil

func createSortedBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	if sortedBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		sortedBuiltinFun = func(args ...object.Object) object.Object {
			o := getSortedListHelper(vm, args...)
			if isError(o) {
				return o
			}
			l, ok := o.(*object.List)
			if !ok {
				return l
			}
			args[0].(*object.List).Elements = l.Elements
			return object.NULL
		}
	}
	return sortedBuiltinFun
}

var allBuiltinFun func(args ...object.Object) object.Object = nil

func createAllBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	if allBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		allBuiltinFun = func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("all", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("all", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE && args[1].Type() != object.BUILTIN_OBJ {
				return newPositionalTypeError("all", 2, object.CLOSURE+" or BUILTIN", args[1].Type())
			}
			l := args[0].(*object.List)
			allTrue := true
			if args[1].Type() == object.CLOSURE {
				fn := args[1].(*object.Closure)
				if len(fn.Fun.Parameters) != 1 {
					return newError("`all` error: function must have 1 parameter")
				}
				for _, elem := range l.Elements {
					obj := vm.applyFunctionFast(fn, elem)
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
					obj := vm.applyFunctionFast(fn, elem)
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
		}
	}
	return allBuiltinFun
}

var anyBuiltinFun func(args ...object.Object) object.Object = nil

func createAnyBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	if anyBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		anyBuiltinFun = func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("any", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("any", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE && args[1].Type() != object.BUILTIN_OBJ {
				return newPositionalTypeError("any", 2, object.CLOSURE+" or BUILTIN", args[1].Type())
			}
			l := args[0].(*object.List)
			anyTrue := false
			if args[1].Type() == object.CLOSURE {
				fn := args[1].(*object.Closure)
				if len(fn.Fun.Parameters) != 1 {
					return newError("`any` error: function must have 1 parameter")
				}
				for _, elem := range l.Elements {
					obj := vm.applyFunctionFast(fn, elem)
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
					obj := vm.applyFunctionFast(fn, elem)
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
		}
	}
	return anyBuiltinFun
}

var mapBuiltinFun func(args ...object.Object) object.Object = nil

func createMapBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	if mapBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		mapBuiltinFun = func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("map", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("map", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE && args[1].Type() != object.BUILTIN_OBJ {
				return newPositionalTypeError("map", 2, object.CLOSURE+" or BUILTIN", args[1].Type())
			}
			l := args[0].(*object.List)
			newElements := make([]object.Object, len(l.Elements))
			for i, elem := range l.Elements {
				obj := vm.applyFunctionFast(args[1], elem)
				if isError(obj) {
					errMsg := obj.(*object.Error).Message
					return newError("`map` error: %s", errMsg)
				}
				newElements[i] = obj
			}
			return &object.List{Elements: newElements}
		}
	}
	return mapBuiltinFun
}

var filterBuiltinFun func(args ...object.Object) object.Object = nil

func createFilterBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	if filterBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		filterBuiltinFun = func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("filter", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("filter", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE && args[1].Type() != object.BUILTIN_OBJ {
				return newPositionalTypeError("filter", 2, object.CLOSURE+" or BUILTIN", args[1].Type())
			}
			l := args[0].(*object.List)
			newElements := []object.Object{}
			for _, elem := range l.Elements {
				obj := vm.applyFunctionFast(args[1], elem)
				if isError(obj) {
					errMsg := obj.(*object.Error).Message
					return newError("`filter` error: %s", errMsg)
				}
				if isTruthy(obj) {
					newElements = append(newElements, elem)
				}
			}
			return &object.List{Elements: newElements}
		}
	}
	return filterBuiltinFun
}

var loadBuiltinFun func(args ...object.Object) object.Object = nil

func createLoadBuiltinFun(_ *VM) func(args ...object.Object) object.Object {
	if loadBuiltinFun == nil || !blueutil.ENABLE_VM_CACHING {
		loadBuiltinFun = func(args ...object.Object) object.Object {
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
				obj := vmStr(o.Value)
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
		}
	}
	return loadBuiltinFun
}

func GetBuiltinFunWithVm(name string, vm *VM) func(args ...object.Object) object.Object {
	switch name {
	case "str":
		return createStrBuiltinFun(vm)
	case "print":
		return createPrintBuiltinFun(vm)
	case "println":
		return createPrintLnBuiltinFun(vm)
	case "to_num":
		return createToNumBuiltinFun()
	case "_sort":
		return createSortBuiltinFun(vm)
	case "_sorted":
		return createSortedBuiltinFun(vm)
	case "all":
		return createAllBuiltinFun(vm)
	case "any":
		return createAnyBuiltinFun(vm)
	case "map":
		return createMapBuiltinFun(vm)
	case "filter":
		return createFilterBuiltinFun(vm)
	case "load":
		return createLoadBuiltinFun(vm)
	default:
		panic(name + " is not supported in GetBuiltinWithVm")
	}
}

func (vm *VM) applyFunctionFastWithMultipleArgs(fun object.Object, args []object.Object) object.Object {
	existingFrames := vm.frames
	existingFrameIndex := vm.framesIndex
	existingStackPointer := vm.sp
	vm.frames = make([]*Frame, MaxFrames)
	vm.frames[0] = NewFrame(&object.Closure{Fun: &object.CompiledFunction{Instructions: code.Instructions{}}}, 0)
	vm.framesIndex = 2
	vm.push(fun)
	argCount := 0
	for _, arg := range args {
		vm.push(arg)
		argCount++
	}
	vm.executeCallFastFrame(argCount)
	err := vm.Run()
	var returnValue object.Object
	if err != nil && err.Error() != consts.NORMAL_EXIT_ON_RETURN {
		returnValue = &object.Error{Message: err.Error()}
	} else {
		returnValue = vm.pop()
	}
	vm.frames = existingFrames
	vm.framesIndex = existingFrameIndex
	vm.sp = existingStackPointer
	return returnValue
}

func (vm *VM) applyFunctionFast(fun, arg object.Object) object.Object {
	var returnValue object.Object
	if _, isClosure := fun.(*object.Closure); isClosure {
		existingFrames := vm.frames
		existingFrameIndex := vm.framesIndex
		existingStackPointer := vm.sp
		vm.frames = make([]*Frame, 3)
		vm.frames[0] = NewFrame(&object.Closure{Fun: &object.CompiledFunction{Instructions: code.Instructions{}}}, 0)
		vm.framesIndex = 2
		vm.push(fun)
		if arg != nil {
			vm.push(arg)
		}
		argCount := 0
		if arg != nil {
			argCount++
		}
		vm.executeCallFastFrame(argCount)
		err := vm.Run()
		if err != nil && err.Error() != consts.NORMAL_EXIT_ON_RETURN {
			returnValue = &object.Error{Message: err.Error()}
		} else {
			returnValue = vm.pop()
		}
		vm.frames = existingFrames
		vm.framesIndex = existingFrameIndex
		vm.sp = existingStackPointer
	} else if _, isBuiltin := fun.(*object.Builtin); isBuiltin {
		vm.push(fun)
		vm.push(arg)
		vm.executeCall(1)
		returnValue = vm.pop()
	} else {
		return newError("%T (%s) is not callable", fun, fun)
	}

	return returnValue
}
