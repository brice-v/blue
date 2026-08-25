package vm

import (
	"blue/consts"
	"blue/object"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

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

func createStrBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
		if len(args) != 1 {
			return newInvalidArgCountError("str", len(args), 1, "")
		}
		if args[0].Type() == object.BYTES_OBJ {
			return object.NewString(string(args[0].(*object.Bytes).Value))
		}
		if it, ok := args[0].(*object.Integer); ok {
			// str(int) repeats heavily in practice; format through a stack
			// buffer and intern so repeated values do not allocate.
			return object.InternInt(it.Value)
		}
		return object.InternString(vm.CustomInspect(args[0]))
	}
}

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
	return func(args ...object.Object) object.Object {
		return printHelper(vm, false, args...)
	}
}

func createPrintLnBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
		return printHelper(vm, true, args...)
	}
}

func createToNumBuiltinFun() func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
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
				newElems[i] = object.NewInteger(int64(e))
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
	// Using custom comparator.
	//
	// Decorate-sort-undecorate: evaluate every key function exactly ONCE
	// per element up front, then sort by the precomputed keys. Key
	// functions receive the original elements (matching Python's sorted
	// key semantics) and the result list reuses those same objects, so no
	// blue <-> Go conversion round-trips are needed. The single-key case
	// (the common one) stores its key inline to avoid a per-element slice.
	type keyedElem struct {
		obj  object.Object
		key  object.Object
		keys []object.Object
	}
	singleKey := len(funs) == 1
	elems := make([]keyedElem, len(l.Elements))
	for i, e := range l.Elements {
		if singleKey {
			keyObj := vm.applyFunctionFast(funs[0], e)
			if isError(keyObj) {
				errMsg := keyObj.(*object.Error).Message
				return newError("`sort` key error: %s", errMsg)
			}
			if keyObj.Type() != object.FLOAT_OBJ && keyObj.Type() != object.INTEGER_OBJ && keyObj.Type() != object.STRING_OBJ {
				return newError("`sort` key error: key function must return INTEGER, STRING, or FLOAT. got = %T (%s)", keyObj, keyObj.Inspect())
			}
			elems[i] = keyedElem{obj: e, key: keyObj}
			continue
		}
		keys := make([]object.Object, len(funs))
		for k := range funs {
			keyObj := vm.applyFunctionFast(funs[k], e)
			if isError(keyObj) {
				errMsg := keyObj.(*object.Error).Message
				return newError("`sort` key error: %s", errMsg)
			}
			if keyObj.Type() != object.FLOAT_OBJ && keyObj.Type() != object.INTEGER_OBJ && keyObj.Type() != object.STRING_OBJ {
				return newError("`sort` key error: key function must return INTEGER, STRING, or FLOAT. got = %T (%s)", keyObj, keyObj.Inspect())
			}
			keys[k] = keyObj
		}
		elems[i] = keyedElem{obj: e, keys: keys}
	}

	sort.SliceStable(elems, func(i, j int) bool {
		a := elems[i].key
		b := elems[j].key
		var ak, bk []object.Object
		if a == nil {
			ak = elems[i].keys
			bk = elems[j].keys
			for k := range ak {
				c := compareSortKeys(ak[k], bk[k])
				if c == 0 && k != len(ak)-1 {
					continue
				}
				if shouldReverse {
					return c > 0
				}
				return c < 0
			}
			return false
		}
		c := compareSortKeys(a, b)
		if shouldReverse {
			return c > 0
		}
		return c < 0
	})

	newObjs := make([]object.Object, len(l.Elements))
	for i := range elems {
		newObjs[i] = elems[i].obj
	}
	return &object.List{Elements: newObjs}
}

// compareSortKeys compares two precomputed sort keys that have already been
// validated as INTEGER, FLOAT, or STRING. Keys of different kinds (including
// int vs float) are treated as equal, mirroring the old comparator's
// mismatch fall-through which returned false for such pairs.
func compareSortKeys(a, b object.Object) int {
	switch left := a.(type) {
	case *object.Integer:
		if right, ok := b.(*object.Integer); ok {
			switch {
			case left.Value < right.Value:
				return -1
			case left.Value > right.Value:
				return 1
			}
		}
	case *object.Float:
		if right, ok := b.(*object.Float); ok {
			switch {
			case left.Value < right.Value:
				return -1
			case left.Value > right.Value:
				return 1
			}
		}
	case *object.Stringo:
		if right, ok := b.(*object.Stringo); ok {
			switch {
			case left.Value < right.Value:
				return -1
			case left.Value > right.Value:
				return 1
			}
		}
	}
	return 0
}

func createSortBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
		return getSortedListHelper(vm, args...)
	}
}

func createSortedBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
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

func createAllBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
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

func createAnyBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
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

func createMapBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
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

func createFilterBuiltinFun(vm *VM) func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
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

func createLoadBuiltinFun(_ *VM) func(args ...object.Object) object.Object {
	return func(args ...object.Object) object.Object {
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

func GetBuiltinFunWithVm(name string, vm *VM) func(args ...object.Object) object.Object {
	if vm.builtinFuns != nil {
		if builtinFun, ok := vm.builtinFuns[name]; ok {
			return builtinFun
		}
	} else {
		vm.builtinFuns = make(map[string]func(args ...object.Object) object.Object)
	}
	var builtinFun func(args ...object.Object) object.Object
	switch name {
	case "str":
		builtinFun = createStrBuiltinFun(vm)
	case "print":
		builtinFun = createPrintBuiltinFun(vm)
	case "println":
		builtinFun = createPrintLnBuiltinFun(vm)
	case "to_num":
		builtinFun = createToNumBuiltinFun()
	case "_sort":
		builtinFun = createSortBuiltinFun(vm)
	case "_sorted":
		builtinFun = createSortedBuiltinFun(vm)
	case "all":
		builtinFun = createAllBuiltinFun(vm)
	case "any":
		builtinFun = createAnyBuiltinFun(vm)
	case "map":
		builtinFun = createMapBuiltinFun(vm)
	case "filter":
		builtinFun = createFilterBuiltinFun(vm)
	case "load":
		builtinFun = createLoadBuiltinFun(vm)
	default:
		panic(name + " is not supported in GetBuiltinWithVm")
	}
	vm.builtinFuns[name] = builtinFun
	return builtinFun
}

// isolatedFramesPool lends full-depth frame arrays to applyFunctionFast and
// applyFunctionFastWithMultipleArgs so repeated callback invocations (map /
// filter / sort keys / dunder calls / ws handlers / spawned processes) do not
// allocate a fresh array per call. The arrays are MaxFrames long because the
// invoked function runs on this array like on a real vm: it can call other
// blue functions, which push more frames (a small fixed-size array here used
// to overflow and panic for callbacks that make nested calls). Slots are
// never read before being written during a run, so reuse across vms is safe.
// The pool holds *[]Frame because putting a slice value would box a new
// interface allocation on every Put.
var isolatedFramesPool = sync.Pool{
	New: func() any {
		frames := make([]Frame, MaxFrames)
		return &frames
	},
}

func (vm *VM) applyFunctionFastWithMultipleArgs(fun object.Object, args []object.Object) object.Object {
	existingFrames := vm.frames
	existingFrameIndex := vm.framesIndex
	existingStackPointer := vm.sp
	framesPtr := isolatedFramesPool.Get().(*[]Frame)
	frames := *framesPtr
	defer isolatedFramesPool.Put(framesPtr)
	vm.frames = frames
	vm.frames[0] = *NewFrame(emptyMainClosure, 0)
	vm.framesIndex = 2
	err := vm.push(fun)
	if err != nil {
		vm.frames = existingFrames
		vm.framesIndex = existingFrameIndex
		vm.sp = existingStackPointer
		return newError("error: %s", err.Error())
	}
	argCount := 0
	for _, arg := range args {
		err = vm.push(arg)
		if err != nil {
			vm.frames = existingFrames
			vm.framesIndex = existingFrameIndex
			vm.sp = existingStackPointer
			return newError("error: %s", err.Error())
		}
		argCount++
	}
	err = vm.executeCallFastFrame(argCount)
	if err != nil {
		vm.frames = existingFrames
		vm.framesIndex = existingFrameIndex
		vm.sp = existingStackPointer
		return newError("error: %s", err.Error())
	}
	err = vm.Run()
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
		framesPtr := isolatedFramesPool.Get().(*[]Frame)
		frames := *framesPtr
		defer isolatedFramesPool.Put(framesPtr)
		vm.frames = frames
		vm.frames[0] = *NewFrame(emptyMainClosure, 0)
		vm.framesIndex = 2
		err := vm.push(fun)
		if err != nil {
			vm.frames = existingFrames
			vm.framesIndex = existingFrameIndex
			vm.sp = existingStackPointer
			return newError("error: %s", err.Error())
		}
		if arg != nil {
			err = vm.push(arg)
			if err != nil {
				vm.frames = existingFrames
				vm.framesIndex = existingFrameIndex
				vm.sp = existingStackPointer
				return newError("error: %s", err.Error())
			}
		}
		argCount := 0
		if arg != nil {
			argCount++
		}
		err = vm.executeCallFastFrame(argCount)
		if err != nil {
			vm.frames = existingFrames
			vm.framesIndex = existingFrameIndex
			vm.sp = existingStackPointer
			return newError("error: %s", err.Error())
		}
		err = vm.Run()
		if err != nil && err.Error() != consts.NORMAL_EXIT_ON_RETURN {
			returnValue = &object.Error{Message: err.Error()}
		} else {
			returnValue = vm.pop()
		}
		vm.frames = existingFrames
		vm.framesIndex = existingFrameIndex
		vm.sp = existingStackPointer
	} else if _, isBuiltin := fun.(*object.Builtin); isBuiltin {
		err := vm.push(fun)
		if err != nil {
			return newError("error: %s", err.Error())
		}
		err = vm.push(arg)
		if err != nil {
			return newError("error: %s", err.Error())
		}
		err = vm.executeCall(1)
		if err != nil {
			return newError("error: %s", err.Error())
		}
		returnValue = vm.pop()
	} else {
		return newError("%T (%s) is not callable", fun, fun)
	}

	return returnValue
}
