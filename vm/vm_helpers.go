package vm

import (
	"blue/code"
	"blue/consts"
	"blue/object"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/clbanning/mxj/v2"
	"github.com/gofiber/fiber/v2"
)

func executeIntegerRangeOperator(leftVal, rightVal int64) object.Object {
	var i int64

	if leftVal < rightVal {
		size := rightVal - leftVal
		listElems := make([]object.Object, 0, size)
		for i = leftVal; i <= rightVal; i++ {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	} else if rightVal < leftVal {
		size := leftVal - rightVal
		listElems := make([]object.Object, 0, size)
		for i = leftVal; i >= rightVal; i-- {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	}
	// When they are equal just return a value (leftVal in this case)
	return &object.List{Elements: []object.Object{&object.Integer{Value: leftVal}}}
}

func executeIntegerNonInclusiveRangeOperator(leftVal, rightVal int64) object.Object {
	var i int64

	if leftVal < rightVal {
		size := rightVal - leftVal
		listElems := make([]object.Object, 0, size-1)
		for i = leftVal; i < rightVal; i++ {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	} else if rightVal < leftVal {
		size := leftVal - rightVal
		listElems := make([]object.Object, 0, size-1)
		for i = leftVal; i > rightVal; i-- {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	}
	return &object.List{Elements: []object.Object{}}
}

func NewGoObj[T any](obj T) *object.GoObj[T] {
	gob := &object.GoObj[T]{Value: obj, Id: object.GoObjId.Add(1)}
	// Note: This is disabled for now due to the complexity of handling all Go Object Types supported by blue
	// t := fmt.Sprintf("%T", gob)
	// if _, ok := goObjDecoders[t]; !ok {
	// 	goObjDecoders[t] = gob.Decoder
	// }
	return gob
}

func newError(format string, a ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func newPositionalTypeError(funName string, pos int, expectedType object.Type, currentType object.Type) *object.Error {
	return newError("PositionalTypeError: `%s` expects argument %d to be %s. got=%s", funName, pos, expectedType, currentType)
}

func newPositionalTypeErrorForGoObj(funName string, pos int, expectedType object.Type, currentObj any) *object.Error {
	return newError("PositionalTypeError: `%s` expects argument %d to be %s. got=%T", funName, pos, expectedType, currentObj)
}

func newInvalidArgCountError(funName string, got, want int, otherCount string) *object.Error {
	if otherCount == "" {
		return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d", funName, got, want)
	}
	return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d %s", funName, got, want, otherCount)
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func nativeToBooleanObject(b bool) *object.Boolean {
	if b {
		return object.TRUE
	}
	return object.FALSE
}

// for now everything that is not null or false returns true
func isTruthy(obj object.Object) bool {
	switch obj {
	case object.NULL:
		return false
	case object.TRUE:
		return true
	case object.FALSE:
		return false
	default:
		if obj.Type() == object.MAP_OBJ {
			return len(obj.(*object.Map).Pairs.Keys) > 0
		}
		if obj.Type() == object.LIST_OBJ {
			return len(obj.(*object.List).Elements) > 0
		}
		if obj.Type() == object.SET_OBJ {
			return len(obj.(*object.Set).Elements.Keys) > 0
		}
		return true
	}
}

// isError is the helper function to determine if an object is an error
func isError(obj object.Object) bool {
	if obj != nil {
		_, isError := obj.(*object.Error)
		return isError
	}
	return false
}

func (vm *VM) stackString() string {
	var out bytes.Buffer
	for i, o := range vm.stack {
		if o != nil {
			fmt.Fprintf(&out, "%d: %s\n", i, o.Inspect())
		}
	}
	return out.String()
}

func matches(left, right object.Object) bool {
	if left.Type() != right.Type() {
		return false
	}
	// TODO: Support other types that can have an _ to signify ignore
	switch left.Type() {
	case object.MAP_OBJ:
		return matchMap(left.(*object.Map), right.(*object.Map))
	default:
		if isPrimitive(left.Type()) {
			return matchPrimitive(left, right)
		}
		return object.HashObject(left) == object.HashObject(right)
	}
}

func isPrimitive(t object.Type) bool {
	switch t {
	case object.STRING_OBJ, object.INTEGER_OBJ, object.UINTEGER_OBJ, object.FLOAT_OBJ, object.BIG_INTEGER_OBJ, object.BIG_FLOAT_OBJ, object.BOOLEAN_OBJ:
		return true
	default:
		return false
	}
}

func matchMap(left, right *object.Map) bool {
	keys := left.Pairs.Keys
	for _, k := range keys {
		leftPair, _ := left.Pairs.Get(k)
		if leftPair.Key == object.VM_IGNORE {
			continue
		}
		rightPair, ok := right.Pairs.Get(k)
		if !ok {
			return false
		}
		// Both keys matched at this point, check values
		if leftPair.Value == object.VM_IGNORE {
			continue
		}
		if leftPair.Value.Type() != rightPair.Value.Type() {
			return false
		}
		if isPrimitive(leftPair.Value.Type()) && matchPrimitive(leftPair.Value, rightPair.Value) {
			continue
		} else if leftPair.Value.Type() == object.MAP_OBJ && matchMap(leftPair.Value.(*object.Map), rightPair.Value.(*object.Map)) {
			continue
		} else if object.HashObject(leftPair.Value) == object.HashObject(rightPair.Value) {
			continue
		} else {
			return false
		}
	}
	return true
}

func matchPrimitive(left, right object.Object) bool {
	switch left.Type() {
	case object.STRING_OBJ:
		return left.(*object.Stringo).Value == right.(*object.Stringo).Value
	case object.INTEGER_OBJ:
		return left.(*object.Integer).Value == right.(*object.Integer).Value
	case object.UINTEGER_OBJ:
		return left.(*object.UInteger).Value == right.(*object.UInteger).Value
	case object.FLOAT_OBJ:
		return left.(*object.Float).Value == right.(*object.Float).Value
	case object.BIG_INTEGER_OBJ:
		return left.(*object.BigInteger).Value.Cmp(right.(*object.BigInteger).Value) == 0
	case object.BIG_FLOAT_OBJ:
		return left.(*object.BigFloat).Value.Equal(right.(*object.BigFloat).Value)
	case object.BOOLEAN_OBJ:
		return left.(*object.Boolean).Value == right.(*object.Boolean).Value
	default:
		return false
	}
}

func isBooleanOperator(op code.Opcode) bool {
	return op == code.OpEqual || op == code.OpNotEqual || op == code.OpAnd || op == code.OpOr || op == code.OpNot
}

// goObjectToBlueObject will only work for simple go types
func goObjectToBlueObject(goObject any) (object.Object, error) {
	switch obj := goObject.(type) {
	case string:
		return &object.Stringo{Value: obj}, nil
	case int:
		return &object.Integer{Value: int64(obj)}, nil
	case int64:
		return &object.Integer{Value: obj}, nil
	case uint:
		return &object.UInteger{Value: uint64(obj)}, nil
	case uint64:
		return &object.UInteger{Value: obj}, nil
	case float32:
		return &object.Float{Value: float64(obj)}, nil
	case float64:
		return &object.Float{Value: obj}, nil
	case bool:
		x := nativeToBooleanObject(obj)
		return x, nil
	case nil:
		return object.NULL, nil
	case []any:
		l := &object.List{Elements: make([]object.Object, len(obj))}
		for i, e := range obj {
			val, err := goObjectToBlueObject(e)
			if err != nil {
				return nil, err
			}
			l.Elements[i] = val
		}
		return l, nil
	case map[string]any:
		m := &object.Map{Pairs: object.NewPairsMap()}
		for k, v := range obj {
			key := &object.Stringo{Value: k}
			hashKey := object.HashObject(key)
			hk := object.HashKey{
				Type:  object.STRING_OBJ,
				Value: hashKey,
			}
			val, err := goObjectToBlueObject(v)
			if err != nil {
				return nil, err
			}
			m.Pairs.Set(hk, object.MapPair{
				Key:   key,
				Value: val,
			})
		}
		return m, nil
	case *object.OrderedMap2[string, any]:
		m := &object.Map{Pairs: object.NewPairsMap()}
		for _, k := range obj.Keys {
			v, _ := obj.Get(k)
			key := &object.Stringo{Value: k}
			hk := object.HashKey{
				Type:  object.STRING_OBJ,
				Value: object.HashObject(key),
			}
			val, err := goObjectToBlueObject(v)
			if err != nil {
				return nil, err
			}
			m.Pairs.Set(hk, object.MapPair{
				Key:   key,
				Value: val,
			})
		}
		return m, nil
	case *object.OrderedMap2[int64, any]:
		m := &object.Map{Pairs: object.NewPairsMap()}
		for _, k := range obj.Keys {
			v, _ := obj.Get(k)
			key := &object.Integer{Value: k}
			hk := object.HashKey{
				Type:  object.INTEGER_OBJ,
				Value: object.HashObject(key),
			}
			val, err := goObjectToBlueObject(v)
			if err != nil {
				return nil, err
			}
			m.Pairs.Set(hk, object.MapPair{
				Key:   key,
				Value: val,
			})
		}
		return m, nil
	case *object.OrderedMap2[float64, any]:
		m := &object.Map{Pairs: object.NewPairsMap()}
		for _, k := range obj.Keys {
			v, _ := obj.Get(k)
			key := &object.Float{Value: k}
			hk := object.HashKey{
				Type:  object.FLOAT_OBJ,
				Value: object.HashObject(key),
			}
			val, err := goObjectToBlueObject(v)
			if err != nil {
				return nil, err
			}
			m.Pairs.Set(hk, object.MapPair{
				Key:   key,
				Value: val,
			})
		}
		return m, nil
	case *object.OrderedMap2[uint64, object.SetPairGo]:
		set := &object.Set{Elements: object.NewOrderedMap[uint64, object.SetPair]()}
		for _, k := range obj.Keys {
			v, _ := obj.Get(k)
			val, err := goObjectToBlueObject(v.Value)
			if err != nil {
				return nil, err
			}
			set.Elements.Set(k, object.SetPair{Value: val, Present: v.Present})
		}
		return set, nil
	default:
		return nil, fmt.Errorf("goObjectToBlueObject: TODO: Type currently unsupported: (%T)", obj)
	}
}

func blueObjectToGoObject(blueObject object.Object) (any, error) {
	if blueObject == nil {
		return nil, fmt.Errorf("blueObjectToGoObject: blueObject must not be nil")
	}
	switch blueObject.Type() {
	case object.STRING_OBJ:
		return blueObject.(*object.Stringo).Value, nil
	case object.INTEGER_OBJ:
		return blueObject.(*object.Integer).Value, nil
	case object.FLOAT_OBJ:
		return blueObject.(*object.Float).Value, nil
	case object.NULL_OBJ:
		return nil, nil
	case object.BOOLEAN_OBJ:
		return blueObject.(*object.Boolean).Value, nil
	case object.MAP_OBJ:
		m := blueObject.(*object.Map)
		allInts := true
		allFloats := true
		allStrings := true
		for _, k := range m.Pairs.Keys {
			mp, _ := m.Pairs.Get(k)
			allInts = allInts && mp.Key.Type() == object.INTEGER_OBJ
			allFloats = allFloats && mp.Key.Type() == object.FLOAT_OBJ
			allStrings = allStrings && mp.Key.Type() == object.STRING_OBJ
		}
		if !allStrings && !allFloats && !allInts {
			return nil, fmt.Errorf("blueObjectToGoObject: Map must only have STRING, INTEGER, or FLOAT keys")
		}
		if allStrings {
			pairs := object.NewOrderedMap[string, any]()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Value.Type() == object.MAP_OBJ {
					return nil, fmt.Errorf("blueObjectToGoObject: Map must not have map values. got=%s", mp.Value.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return nil, err
				}
				pairs.Set(mp.Key.(*object.Stringo).Value, val)
			}
			return pairs, nil
		} else if allInts {
			pairs := object.NewOrderedMap[int64, any]()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Value.Type() == object.MAP_OBJ {
					return nil, fmt.Errorf("blueObjectToGoObject: Map must not have map values. got=%s", mp.Value.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return nil, err
				}
				pairs.Set(mp.Key.(*object.Integer).Value, val)
			}
			return pairs, nil
		} else {
			// Floats
			pairs := object.NewOrderedMap[float64, any]()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Value.Type() == object.MAP_OBJ {
					return nil, fmt.Errorf("blueObjectToGoObject: Map must not have map values. got=%s", mp.Value.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return nil, err
				}
				pairs.Set(mp.Key.(*object.Float).Value, val)
			}
			return pairs, nil
		}
	case object.LIST_OBJ:
		l := blueObject.(*object.List).Elements
		elements := make([]any, len(l))
		for i, e := range l {
			if e.Type() == object.LIST_OBJ {
				return nil, fmt.Errorf("blueObjectToGoObject: List of lists unsupported")
			}
			val, err := blueObjectToGoObject(e)
			if err != nil {
				return nil, err
			}
			elements[i] = val
		}
		return elements, nil
	case object.SET_OBJ:
		s := blueObject.(*object.Set)
		set := object.NewOrderedMap[uint64, object.SetPairGo]()
		for _, k := range s.Elements.Keys {
			v, _ := s.Elements.Get(k)
			val, err := blueObjectToGoObject(v.Value)
			if err != nil {
				return nil, err
			}
			set.Set(k, object.SetPairGo{Value: val, Present: struct{}{}})
		}
		return set, nil
	default:
		return nil, fmt.Errorf("blueObjectToGoObject: TODO: Type currently unsupported: %s (%T)", blueObject.Type(), blueObject)
	}
}

func (vm *VM) buildSliceFrom(maybeSliceable object.Object, sliceIndexes object.Object) object.Object {
	if maybeSliceable.Type() != object.LIST_OBJ && maybeSliceable.Type() != object.STRING_OBJ && maybeSliceable.Type() != object.SET_OBJ {
		return newError("slice cannot be created with type: %s", maybeSliceable.Type())
	}
	sliceIndexesList := sliceIndexes.(*object.List).Elements
	minSliceIndex := sliceIndexesList[0].(*object.Integer).Value
	maxSliceIndex := sliceIndexesList[len(sliceIndexesList)-1].(*object.Integer).Value + 1
	if maybeSliceable.Type() == object.LIST_OBJ {
		result := make([]object.Object, len(sliceIndexesList))
		copy(result[:], maybeSliceable.(*object.List).Elements[minSliceIndex:maxSliceIndex])
		return &object.List{Elements: result}
	} else if maybeSliceable.Type() == object.SET_OBJ {
		result := object.NewSetElementsWithSize(len(sliceIndexesList))
		s := maybeSliceable.(*object.Set)
		for _, o := range sliceIndexesList {
			index := o.(*object.Integer).Value
			sp := s.Elements.Keys[index]
			obj, _ := s.Elements.Get(sp)
			result.Set(sp, obj)
		}
		return &object.Set{Elements: result}
	} else {
		sliceable := []rune(maybeSliceable.(*object.Stringo).Value)
		result := sliceable[minSliceIndex:maxSliceIndex]
		return &object.Stringo{Value: string(result)}
	}
}

// TODO: Doesnt work properly currently, mostly here as a stub
func getErrorTokenTraceAsJson(vm *VM) any {
	return getErrorTokenTraceAsJsonWithError(vm, "")
}

func getErrorTokenTraceAsJsonWithError(vm *VM, errorMsg string) any {
	var disableHttpServerDebug bool
	disableHttpServerDebugStr := os.Getenv(consts.BLUE_DISABLE_HTTP_SERVER_DEBUG)
	disableHttpServerDebug, err := strconv.ParseBool(disableHttpServerDebugStr)
	if err != nil {
		disableHttpServerDebug = false
	}
	var errors []string
	if errorMsg == "" {
		errors = []string{}
	} else {
		errors = []string{errorMsg}
	}
	if !disableHttpServerDebug {
		// for e.ErrorTokens.Len() > 0 {
		// 	firstPart, carat := lexer.GetErrorLineMessageForJson(e.ErrorTokens.PopBack())
		// 	errors = append(errors, firstPart, carat)
		// }
		fmt.Println("`http handler` error: " + errorMsg)
		for _, err := range errors {
			fmt.Printf("%s\n", err)
		}
	}
	return errors
}

func blueObjToJsonObject(obj object.Object) object.Object {
	rootNodeType := obj.Type()
	if isError(obj) {
		return obj
	}
	// https://www.w3schools.com/Js/js_json_objects.asp
	// Keys must be strings, and values must be a valid JSON data type:
	// string
	// number
	// object
	// array
	// boolean
	// null
	if !isValidJsonValueType(rootNodeType) {
		return newPositionalTypeError("to_json", 1, "MAP, LIST, NUM, NULL, BOOLEAN", rootNodeType)
	}
	switch rootNodeType {
	case object.MAP_OBJ:
		mObj := obj.(*object.Map)
		ok, err := checkMapObjPairsForValidJsonKeysAndValues(mObj.Pairs)
		if !ok {
			return newError("`to_json` error validating MAP object. %s", err.Error())
		}
		var buf bytes.Buffer
		jsonString := generateJsonStringFromValidMapObjPairs(buf, mObj.Pairs)
		return &object.Stringo{Value: jsonString.String()}
	case object.LIST_OBJ:
		lObj := obj.(*object.List)
		ok, err := checkListElementsForValidJsonValues(lObj.Elements)
		if !ok {
			return newError("`to_json` error validating LIST object. %s", err.Error())
		}
		var buf bytes.Buffer
		jsonString := generateJsonStringFromValidListElements(buf, lObj.Elements)
		return &object.Stringo{Value: jsonString.String()}
	default:
		// We can do this as the default because the arg is verified up above
		var buf bytes.Buffer
		jsonString := generateJsonStringFromOtherValidTypes(buf, obj)
		return &object.Stringo{Value: jsonString.String()}
	}
}

func isValidJsonValueType(t object.Type) bool {
	return t == object.STRING_OBJ || t == object.INTEGER_OBJ || t == object.UINTEGER_OBJ || t == object.BIG_FLOAT_OBJ || t == object.BIG_INTEGER_OBJ || t == object.FLOAT_OBJ || t == object.NULL_OBJ || t == object.BOOLEAN_OBJ || t == object.MAP_OBJ || t == object.LIST_OBJ
}

func checkListElementsForValidJsonValues(elements []object.Object) (bool, error) {
	for _, e := range elements {
		elType := e.Type()
		if !isValidJsonValueType(elType) {
			return false, errors.New("all values should be of the types STRING, INTEGER, FLOAT, BOOLEAN, NULL, LIST, or MAP. found " + string(elType))
		}
		if elType == object.LIST_OBJ {
			elements2 := e.(*object.List).Elements
			return checkListElementsForValidJsonValues(elements2)
		}
		if elType == object.MAP_OBJ {
			mapPairs := e.(*object.Map).Pairs
			return checkMapObjPairsForValidJsonKeysAndValues(mapPairs)
		}
	}
	return true, nil
}

func checkMapObjPairsForValidJsonKeysAndValues(pairs object.OrderedMap2[object.HashKey, object.MapPair]) (bool, error) {
	for _, hk := range pairs.Keys {
		mp, _ := pairs.Get(hk)
		keyType := mp.Key.Type()
		valueType := mp.Value.Type()
		if keyType != object.STRING_OBJ {
			return false, errors.New("all keys should be STRING, found invalid key type " + string(keyType))
		}
		if !isValidJsonValueType(valueType) {
			return false, errors.New("all values should be of the types STRING, INTEGER, FLOAT, BOOLEAN, NULL, LIST, or MAP. found " + string(valueType))
		}
		// These types are all okay but if its
		if valueType == object.MAP_OBJ {
			mapPairs := mp.Value.(*object.Map).Pairs
			return checkMapObjPairsForValidJsonKeysAndValues(mapPairs)
		}
		if valueType == object.LIST_OBJ {
			elements := mp.Value.(*object.List).Elements
			return checkListElementsForValidJsonValues(elements)
		}
	}
	// Empty pairs should be okay
	return true, nil
}

func generateJsonStringFromValidListElements(buf bytes.Buffer, elements []object.Object) bytes.Buffer {
	buf.WriteRune('[')
	for i, e := range elements {
		elemType := e.Type()
		switch elemType {
		case object.LIST_OBJ:
			elements1 := e.(*object.List).Elements
			buf = generateJsonStringFromValidListElements(buf, elements1)
		case object.MAP_OBJ:
			pairs := e.(*object.Map).Pairs
			buf = generateJsonStringFromValidMapObjPairs(buf, pairs)
		default:
			buf = generateJsonStringFromOtherValidTypes(buf, e)
		}
		if i < len(elements)-1 {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune(']')
	return buf
}

func generateJsonStringFromValidMapObjPairs(buf bytes.Buffer, pairs object.OrderedMap2[object.HashKey, object.MapPair]) bytes.Buffer {
	buf.WriteRune('{')
	length := len(pairs.Keys)
	i := 0
	for _, hk := range pairs.Keys {
		mp, _ := pairs.Get(hk)
		buf.WriteString(fmt.Sprintf("%q:", mp.Key.Inspect()))
		valueType := mp.Value.Type()
		switch valueType {
		case object.MAP_OBJ:
			pairs1 := mp.Value.(*object.Map).Pairs
			buf = generateJsonStringFromValidMapObjPairs(buf, pairs1)
		case object.LIST_OBJ:
			elements := mp.Value.(*object.List).Elements
			buf = generateJsonStringFromValidListElements(buf, elements)
		default:
			buf = generateJsonStringFromOtherValidTypes(buf, mp.Value)
		}
		if i < length-1 {
			buf.WriteRune(',')
		}
		i++
	}
	buf.WriteRune('}')
	return buf
}

func generateJsonStringFromOtherValidTypes(buf bytes.Buffer, element object.Object) bytes.Buffer {
	switch t := element.(type) {
	case *object.Integer:
		fmt.Fprintf(&buf, "%d", t.Value)
	case *object.UInteger:
		fmt.Fprintf(&buf, "%d", t.Value)
	case *object.BigInteger:
		buf.WriteString(t.Value.String())
	case *object.BigFloat:
		buf.WriteString(t.Value.String())
	case *object.Stringo:
		fmt.Fprintf(&buf, "%q", t.Value)
	case *object.Null:
		buf.WriteString("null")
	case *object.Float:
		fmt.Fprintf(&buf, "%f", t.Value)
	case *object.Boolean:
		fmt.Fprintf(&buf, "%t", t.Value)
	}
	return buf
}

func prepareAndApplyHttpHandleFn(vm *VM, fn *object.Closure, c *fiber.Ctx, method string) (bool, object.Object, []string) {
	isGet := method == "GET"
	isDelete := method == "DELETE"
	methodLower := strings.ToLower(method)
	if !isGet && !isDelete {
		handleSpecialFunctionArgs2(fn, methodLower+"_values", c)
	}
	handleSpecialFunctionArgs2(fn, "query_params", c)
	fnArgs := getAndSetHttpParams(fn, c)
	return true, vm.applyFunctionFastWithMultipleArgs(fn, fnArgs), []string{}
}

func handleSpecialFunctionArgs2(fn *object.Closure, varName string, c *fiber.Ctx) {
	if fn.Fun.SpecialFunctionParameters == nil {
		return
	}
	for i, p := range fn.Fun.Parameters {
		if !fn.Fun.ParameterHasDefault[i] {
			continue
		}
		key := object.NameIndexKey{Name: p, Index: i}
		switch p {
		case "query_params":
			if objectMap, ok := fn.Fun.SpecialFunctionParameters[key]; ok {
				for k := range objectMap {
					objectMap[k] = &object.Stringo{Value: c.Query(k.Name)}
				}
			}
		case "cookies":
			if objectMap, ok := fn.Fun.SpecialFunctionParameters[key]; ok {
				for k := range objectMap {
					objectMap[k] = &object.Stringo{Value: c.Cookies(k.Name)}
				}
			}
		case varName:
			contentType := c.Get("Content-Type")
			body := strings.NewReader(string(c.Body()))
			returnMap, err := decodeBodyToMap(contentType, body)
			if objectMap, ok := fn.Fun.SpecialFunctionParameters[key]; ok {
				for k := range objectMap {
					if err != nil {
						objectMap[k] = &object.Error{Message: err.Error()}
						continue
					}
					s := k.Name
					if v, ok := returnMap[s]; ok {
						objectMap[k] = v
					} else {
						objectMap[k] = &object.Stringo{Value: c.FormValue(s)}
					}
				}
			}
		}
	}
}

func decodeBodyToMap(contentType string, body io.Reader) (map[string]object.Object, error) {
	returnMap := make(map[string]object.Object)
	var v map[string]any
	if strings.Contains(contentType, "xml") {
		xmld := xml.NewDecoder(body)
		err := xmld.Decode(&v)
		if err != nil {
			err = nil
			for {
				mv, err := mxj.NewMapXmlReader(body)
				if err != nil {
					break
				}
				if mv == nil {
					break
				}
				if v == nil {
					v = make(map[string]any)
				}
				maps.Copy(v, mv)
			}
		}
	} else if strings.Contains(contentType, "json") {
		jsond := json.NewDecoder(body)
		err := jsond.Decode(&v)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, nil
	}
	for key, value := range v {
		returnMap[key] = decodeInterfaceToObject(value)
	}
	return returnMap, nil
}

func decodeInterfaceToObject(value any) object.Object {
	switch x := value.(type) {
	case int64:
		return &object.Integer{Value: x}
	case float64:
		return &object.Float{Value: x}
	case string:
		return &object.Stringo{Value: x}
	case bool:
		return nativeToBooleanObject(x)
	case []any:
		list := &object.List{Elements: make([]object.Object, len(x))}
		for i, e := range x {
			list.Elements[i] = decodeInterfaceToObject(e)
		}
		return list
	case map[string]any:
		mapObj := object.NewOrderedMap[string, object.Object]()
		for k, v := range x {
			mapObj.Set(k, decodeInterfaceToObject(v))
		}
		return object.CreateMapObjectForGoMap(*mapObj)
	case *object.OrderedMap2[string, any]:
		mapObj := object.NewOrderedMap[string, object.Object]()
		for _, k := range x.Keys {
			v, _ := x.Get(k)
			mapObj.Set(k, decodeInterfaceToObject(v))
		}
		return object.CreateMapObjectForGoMap(*mapObj)
	default:
		consts.ErrorPrinter("decodeInterfaceToObject: HANDLE TYPE = %T\n", x)
		os.Exit(1)
	}
	return object.NULL
}

func getAndSetHttpParams(fn *object.Closure, c *fiber.Ctx) []object.Object {
	fnArgs := make([]object.Object, len(fn.Fun.Parameters))
	for i, v := range fn.Fun.Parameters {
		switch v {
		case "headers":
			// Handle headers
			fnArgs[i] = getReqHeaderMapObj(c)
		case "request":
			req := c.Request()
			mapObj := object.NewOrderedMap[string, object.Object]()
			mapObj.Set("method", &object.Stringo{Value: c.Method()})
			mapObj.Set("proto", &object.Stringo{Value: c.Protocol()})
			mapObj.Set("uri", &object.Stringo{Value: string(req.URI().FullURI())})
			mapObj.Set("scheme", &object.Stringo{Value: string(req.URI().Scheme())})
			mapObj.Set("host", &object.Stringo{Value: string(req.URI().Host())})
			mapObj.Set("request_uri", &object.Stringo{Value: string(req.URI().RequestURI())})
			mapObj.Set("hash", &object.Stringo{Value: string(req.URI().Hash())})
			headersMapObj := getReqHeaderMapObj(c)
			mapObj.Set("headers", headersMapObj)
			mapObj.Set("ip", &object.Stringo{Value: c.IP()})
			mapObj.Set("is_from_local", nativeToBooleanObject(c.IsFromLocal()))
			mapObj.Set("is_secure", nativeToBooleanObject(c.Secure()))
			fnArgs[i] = object.CreateMapObjectForGoMap(*mapObj)
		case "ctx", "context":
			fnArgs[i] = getCtxFunctionMapObj(c)
		default:
			fnArgs[i] = &object.Stringo{Value: c.Params(v)}
		}
	}
	return fnArgs
}

func getReqHeaderMapObj(c *fiber.Ctx) object.Object {
	headers := c.GetReqHeaders()
	mapObj := object.NewOrderedMap[string, object.Object]()
	headerKeys := make([]string, len(headers))
	i := 0
	for k := range headers {
		headerKeys[i] = k
		i++
	}
	// Sort by key to always have the headers in order
	sort.Strings(headerKeys)
	for i := 0; i < len(headers); i++ {
		mapObj.Set(headerKeys[i], &object.Stringo{Value: headers[headerKeys[i]]})
	}
	return object.CreateMapObjectForGoMap(*mapObj)
}

func getCtxFunctionMapObj(c *fiber.Ctx) object.Object {
	mapObj := object.NewOrderedMap[string, object.Object]()
	mapObj.Set("clear_cookie", &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			cookieArgs := []string{}
			for i, arg := range args {
				if args[i].Type() != object.STRING_OBJ {
					return newPositionalTypeError("clear_cookie", i+1, object.STRING_OBJ, args[i].Type())
				}
				cookie := arg.(*object.Stringo).Value
				cookieArgs = append(cookieArgs, cookie)
			}
			c.ClearCookie(cookieArgs...)
			return object.NULL
		},
	})
	mapObj.Set("set_cookie", &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			// Arg len should be 1
			// Arg should be map
			// Map requires name - all the rest could be empty
			if len(args) != 1 {
				return newInvalidArgCountError("set_cookie", len(args), 1, "")
			}
			if args[0].Type() != object.MAP_OBJ {
				return newPositionalTypeError("set_cookie", 1, object.MAP_OBJ, args[0].Type())
			}
			jsonO := blueObjToJsonObject(args[0])
			if isError(jsonO) {
				return newError("`set_cookie` error: %s", jsonO.(*object.Error).Message)
			}
			if jj, ok := jsonO.(*object.Stringo); ok {
				cookie := new(fiber.Cookie)
				err := json.Unmarshal([]byte(jj.Value), cookie)
				if err != nil {
					return newError("`set_cookie` error: %s", err.Error())
				}
				if cookie.Domain == "" {
					cookie.Domain = strings.Split(c.Hostname(), ":")[0]
				}
				c.Cookie(cookie)
			}
			return object.NULL
		},
	})
	mapObj.Set("get_cookie", &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("get_cookie", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("get_cookie", 1, object.STRING_OBJ, args[0].Type())
			}
			return &object.Stringo{Value: c.Cookies(args[0].(*object.Stringo).Value)}
		},
	})
	mapObj.Set("set_local", &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("set_local", len(args), 2, "")
			}
			if isError(args[0]) {
				return args[0]
			}
			if isError(args[1]) {
				return args[1]
			}
			a, err := blueObjectToGoObject(args[0])
			if err != nil {
				return newError("`set_local` error: %s", err.Error())
			}
			b, err := blueObjectToGoObject(args[1])
			if err != nil {
				return newError("`set_local` error: %s", err.Error())
			}
			c.Locals(a, b)
			return object.NULL
		},
	})
	mapObj.Set("get_local", &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("get_local", len(args), 1, "")
			}
			if isError(args[0]) {
				return args[0]
			}
			a, err := blueObjectToGoObject(args[0])
			if err != nil {
				return newError("`get_local` error: %s", err.Error())
			}
			localObj := c.Locals(a)
			obj, err := goObjectToBlueObject(localObj)
			if err != nil {
				return newError("`get_local` error: Locals variable was not an object. got=%s", err.Error())
			}
			return obj
		},
	})
	return object.CreateMapObjectForGoMap(*mapObj)
}

func tryGetHttpActionAndMap(respObj object.Object) (isAction bool, action string, m *object.OrderedMap2[string, any]) {
	isAction, action, m = false, "", nil
	mObj, err := blueObjectToGoObject(respObj)
	if err == nil {
		if mm, ok := mObj.(*object.OrderedMap2[string, any]); ok {
			if kt, ok := mm.Get("t"); ok {
				if kts, ok := kt.(string); ok {
					if strings.Contains(kts, "http/") {
						// Now we know this is good to use
						isAction = true
						action = strings.Split(kts, "/")[1]
						m = mm
						return
					}
				}
			}
		}
	}
	return
}

func (vm *VM) pushSpecialFunctionParameter(parameterIndex, listIndex int) error {
	for k, v := range vm.currentFrame().cl.Fun.SpecialFunctionParameters {
		if k.Index != parameterIndex {
			continue
		}
		for kk, vv := range v {
			if kk.Index != listIndex {
				continue
			}
			return vm.push(vv)
		}
	}
	return fmt.Errorf("failed to GetSpecialFunctionParameter with Parameter Index: %d and List Index: %d", parameterIndex, listIndex)
}

func (vm *VM) pushSpecialFunctionParameter2(listIndex int) error {
	for _, v := range vm.currentFrame().cl.Fun.SpecialFunctionParameters {
		for kk, vv := range v {
			if vv == nil || kk.Index != listIndex {
				continue
			}
			return vm.push(vv)
		}
	}
	return fmt.Errorf("failed to GetSpecialFunctionParameter2 with List Index: %d", listIndex)
}
