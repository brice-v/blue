package object

import (
	"blue/ast"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os/exec"
	"runtime"
	"runtime/metrics"
	"sort"
	"strings"
)

// CreateBasicMapObject creates an object that looks like {'t': objType, 'v': objValue}
// This is currently being used for `spawn` and `db.open()` so that a unique return value
// is created (and it allows the functions defined for them to work)
func CreateBasicMapObject(objType string, objValue uint64) *Map {
	m := &Map{
		Pairs: NewPairsMapWithSize(1),
	}
	typeKeyStr := &Stringo{Value: "t"}
	typeValueStr := &Stringo{Value: objType}

	typeKeyHashedKey := HashObject(typeKeyStr)
	typeKeyHashKey := HashKey{Type: STRING_OBJ, Value: typeKeyHashedKey}
	m.Pairs.Set(typeKeyHashKey, MapPair{
		Key:   typeKeyStr,
		Value: typeValueStr,
	})

	valueKeyStr := &Stringo{Value: "v"}
	valueValueStr := &UInteger{Value: objValue}

	valueKeyHashedKey := HashObject(valueKeyStr)
	valueKeyHashKey := HashKey{Type: STRING_OBJ, Value: valueKeyHashedKey}
	m.Pairs.Set(valueKeyHashKey, MapPair{
		Key:   valueKeyStr,
		Value: valueValueStr,
	})

	return m
}

func CreateBasicMapObjectForGoObj[T any](objType string, goObj *GoObj[T]) *Map {
	m := &Map{
		Pairs: NewPairsMapWithSize(1),
	}
	typeKeyStr := &Stringo{Value: "t"}
	typeValueStr := &Stringo{Value: objType}

	typeKeyHashedKey := HashObject(typeKeyStr)
	typeKeyHashKey := HashKey{Type: STRING_OBJ, Value: typeKeyHashedKey}
	m.Pairs.Set(typeKeyHashKey, MapPair{
		Key:   typeKeyStr,
		Value: typeValueStr,
	})

	valueKeyStr := &Stringo{Value: "v"}

	valueKeyHashedKey := HashObject(valueKeyStr)
	valueKeyHashKey := HashKey{Type: STRING_OBJ, Value: valueKeyHashedKey}
	m.Pairs.Set(valueKeyHashKey, MapPair{
		Key:   valueKeyStr,
		Value: goObj,
	})

	return m
}

func CreateMapObjectForGoMap(input OrderedMap2[string, Object]) *Map {
	m := &Map{
		Pairs: NewPairsMapWithSize(len(input.Keys)),
	}

	for _, k := range input.Keys {
		v, _ := input.Get(k)
		kStr := &Stringo{Value: k}

		hkStr := HashObject(kStr)
		hk := HashKey{Type: STRING_OBJ, Value: hkStr}
		m.Pairs.Set(hk, MapPair{
			Key:   kStr,
			Value: v,
		})
	}

	return m
}

func CreateObjectFromDbInterface(input any) Object {
	switch x := input.(type) {
	case int64:
		return &Integer{Value: x}
	case string:
		return &Stringo{Value: x}
	case float64:
		return &Float{Value: x}
	case []byte:
		return &Bytes{Value: x}
	case bool:
		return &Boolean{Value: x}
	default:
		return nil
	}
}

func createHelpStringForObject(name, desc string, obj Object) string {
	return fmt.Sprintf("Help:    `%s` %s\nType:    '%s'\nInspect: %s", name, desc, obj.Type(), obj.Inspect())
}

func IsCollectionType(t Type) bool {
	return t == LIST_OBJ || t == SET_OBJ || t == MAP_OBJ
}

// newError is the wrapper function to add an error to the evaluator
func newError(format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func newPositionalTypeError(funName string, pos int, expectedType Type, currentType Type) *Error {
	return newError("PositionalTypeError: `%s` expects argument %d to be %s. got=%s", funName, pos, expectedType, currentType)
}

func newPositionalTypeErrorForGoObj(funName string, pos int, expectedType Type, currentObj any) *Error {
	return newError("PositionalTypeError: `%s` expects argument %d to be %s. got=%T", funName, pos, expectedType, currentObj)
}

func newInvalidArgCountError(funName string, got, want int, otherCount string) *Error {
	if otherCount == "" {
		return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d", funName, got, want)
	}
	return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d %s", funName, got, want, otherCount)
}

func anyBlueObjectToGoObject(blueObject Object) (any, error) {
	if blueObject == nil {
		return nil, fmt.Errorf("blueObjectToGoObject: blueObject must not be nil")
	}
	switch blueObject.Type() {
	case STRING_OBJ, INTEGER_OBJ, FLOAT_OBJ, NULL_OBJ, BOOLEAN_OBJ, MAP_OBJ, LIST_OBJ, SET_OBJ:
		return blueObjectToGoObject(blueObject)
	case UINTEGER_OBJ:
		return blueObject.(*UInteger).Value, nil
	default:
		return nil, fmt.Errorf("blueObjectToGoObject: TODO: Type currently unsupported: %s (%T)", blueObject.Type(), blueObject)
	}
}

func blueObjectToGoObject(blueObject Object) (any, error) {
	if blueObject == nil {
		return nil, fmt.Errorf("blueObjectToGoObject: blueObject must not be nil")
	}
	switch blueObject.Type() {
	case STRING_OBJ:
		return blueObject.(*Stringo).Value, nil
	case INTEGER_OBJ:
		return blueObject.(*Integer).Value, nil
	case FLOAT_OBJ:
		return blueObject.(*Float).Value, nil
	case NULL_OBJ:
		return nil, nil
	case BOOLEAN_OBJ:
		return blueObject.(*Boolean).Value, nil
	case MAP_OBJ:
		m := blueObject.(*Map)
		allInts := true
		allFloats := true
		allStrings := true
		for _, k := range m.Pairs.Keys {
			mp, _ := m.Pairs.Get(k)
			allInts = allInts && mp.Key.Type() == INTEGER_OBJ
			allFloats = allFloats && mp.Key.Type() == FLOAT_OBJ
			allStrings = allStrings && mp.Key.Type() == STRING_OBJ
		}
		if !allStrings && !allFloats && !allInts {
			return nil, fmt.Errorf("blueObjectToGoObject: Map must only have STRING, INTEGER, or FLOAT keys")
		}
		if allStrings {
			pairs := NewOrderedMap[string, any]()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Value.Type() == MAP_OBJ {
					return nil, fmt.Errorf("blueObjectToGoObject: Map must not have map values. got=%s", mp.Value.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return nil, err
				}
				pairs.Set(mp.Key.(*Stringo).Value, val)
			}
			return pairs, nil
		} else if allInts {
			pairs := NewOrderedMap[int64, any]()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Value.Type() == MAP_OBJ {
					return nil, fmt.Errorf("blueObjectToGoObject: Map must not have map values. got=%s", mp.Value.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return nil, err
				}
				pairs.Set(mp.Key.(*Integer).Value, val)
			}
			return pairs, nil
		} else {
			// Floats
			pairs := NewOrderedMap[float64, any]()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Value.Type() == MAP_OBJ {
					return nil, fmt.Errorf("blueObjectToGoObject: Map must not have map values. got=%s", mp.Value.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return nil, err
				}
				pairs.Set(mp.Key.(*Float).Value, val)
			}
			return pairs, nil
		}
	case LIST_OBJ:
		l := blueObject.(*List).Elements
		elements := make([]any, len(l))
		for i, e := range l {
			if e.Type() == LIST_OBJ {
				return nil, fmt.Errorf("blueObjectToGoObject: List of lists unsupported")
			}
			val, err := blueObjectToGoObject(e)
			if err != nil {
				return nil, err
			}
			elements[i] = val
		}
		return elements, nil
	case SET_OBJ:
		s := blueObject.(*Set)
		set := NewOrderedMap[uint64, SetPairGo]()
		for _, k := range s.Elements.Keys {
			v, _ := s.Elements.Get(k)
			val, err := blueObjectToGoObject(v.Value)
			if err != nil {
				return nil, err
			}
			set.Set(k, SetPairGo{Value: val, Present: struct{}{}})
		}
		return set, nil
	default:
		return nil, fmt.Errorf("blueObjectToGoObject: TODO: Type currently unsupported: %s (%T)", blueObject.Type(), blueObject)
	}
}

func getBlueObjectFromResp(resp *http.Response) Object {
	_body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()
	body := &Stringo{Value: string(_body)}
	contentLength := &Integer{Value: resp.ContentLength}
	headersToMapObj := func(header http.Header) Object {
		mapObj := NewOrderedMap[string, Object]()
		for k, v := range header {
			list := &List{Elements: make([]Object, len(v))}
			for i, e := range v {
				list.Elements[i] = &Stringo{Value: e}
			}
			mapObj.Set(k, list)
		}
		return CreateMapObjectForGoMap(*mapObj)
	}
	headers := headersToMapObj(resp.Header)
	proto := &Stringo{Value: resp.Proto}
	requestToMapObj := func(req *http.Request) Object {
		mapObj := NewOrderedMap[string, Object]()
		mapObj.Set("method", &Stringo{Value: req.Method})
		mapObj.Set("proto", &Stringo{Value: req.Proto})
		mapObj.Set("url", &Stringo{Value: req.URL.String()})
		return CreateMapObjectForGoMap(*mapObj)
	}
	request := requestToMapObj(resp.Request)
	rawStatus := &Stringo{Value: resp.Status}
	status := &Integer{Value: int64(resp.StatusCode)}

	trailer := headersToMapObj(resp.Trailer)
	transferEncoding := &List{Elements: make([]Object, len(resp.TransferEncoding))}
	for i, v := range resp.TransferEncoding {
		transferEncoding.Elements[i] = &Stringo{Value: v}
	}
	uncompressed := nativeToBooleanObject(resp.Uncompressed)
	_cookies := resp.Cookies()
	cookieToMapObj := func(c *http.Cookie) Object {
		mapObj := NewOrderedMap[string, Object]()
		mapObj.Set("name", &Stringo{Value: c.Name})
		mapObj.Set("value", &Stringo{Value: c.Value})
		mapObj.Set("path", &Stringo{Value: c.Path})
		mapObj.Set("domain", &Stringo{Value: c.Domain})
		mapObj.Set("expires", &Integer{Value: c.Expires.Unix()})
		mapObj.Set("http_only", nativeToBooleanObject(c.HttpOnly))
		mapObj.Set("secure", nativeToBooleanObject(c.Secure))
		mapObj.Set("samesite", &Integer{Value: int64(c.SameSite)})
		mapObj.Set("raw", &Stringo{Value: c.String()})
		return CreateMapObjectForGoMap(*mapObj)
	}
	cookies := &List{Elements: make([]Object, len(_cookies))}
	for i, c := range _cookies {
		cookies.Elements[i] = cookieToMapObj(c)
	}

	returnMap := NewOrderedMap[string, Object]()
	returnMap.Set("status", status)
	returnMap.Set("body", body)
	returnMap.Set("content_len", contentLength)
	returnMap.Set("headers", headers)
	returnMap.Set("proto", proto)
	returnMap.Set("request", request)
	returnMap.Set("raw_status", rawStatus)
	returnMap.Set("trailer", trailer)
	returnMap.Set("transfer_encoding", transferEncoding)
	returnMap.Set("uncompressed", uncompressed)
	returnMap.Set("cookies", cookies)

	return CreateMapObjectForGoMap(*returnMap)
}

func nativeToBooleanObject(input bool) *Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func NewGoObj[T any](obj T) *GoObj[T] {
	gob := &GoObj[T]{Value: obj, Id: GoObjId.Add(1)}
	// Note: This is disabled for now due to the complexity of handling all Go Object Types supported by blue
	// t := fmt.Sprintf("%T", gob)
	// if _, ok := goObjDecoders[t]; !ok {
	// 	goObjDecoders[t] = gob.Decoder
	// }
	return gob
}

func ExecStringCommand(str string) Object {
	if NoExec {
		return newError("cannot execute string command `%s`. NoExec set to true.", str)
	}
	splitStr := strings.Split(str, " ")
	if len(splitStr) == 0 {
		return newError("unable to exec the string `%s`", str)
	}
	if len(splitStr) == 1 {
		output, err := execCommand(splitStr[0]).Output()
		if err != nil {
			return newError("unable to exec the string `%s`. Error: %s", str, err)
		}
		return &Stringo{Value: string(output[:])}
	}
	cleanedStrings := []string{}
	for _, v := range splitStr {
		if v != "" {
			cleanedStrings = append(cleanedStrings, v)
			continue
		}
	}
	first := cleanedStrings[0]
	rest := cleanedStrings[1:]

	output, err := execCommand(first, rest...).CombinedOutput()
	if err != nil {
		return newError("unable to exec the string `%s`. Error: %s", str, err)
	}
	if len(output) == 0 {
		return NULL
	}
	return &Stringo{Value: string(output[:])}
}

func execCommand(arg0 string, args ...string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		cmdArgs := append([]string{"cmd", "/c", arg0}, args...)
		return exec.Command(cmdArgs[0], cmdArgs...)
	} else {
		return exec.Command(arg0, args...)
	}
}

func getListOfProcesses(arg Object) ([]*Process, bool) {
	if arg.Type() != LIST_OBJ {
		return nil, false
	}
	elems := arg.(*List).Elements
	processes := make([]*Process, 0, len(elems))
	for _, e := range elems {
		if e.Type() != PROCESS_OBJ {
			return nil, false
		}
		v := e.(*Process)
		processes = append(processes, v)
	}
	return processes, true
}

func medianBucket(h *metrics.Float64Histogram) float64 {
	total := uint64(0)
	for _, count := range h.Counts {
		total += count
	}
	thresh := total / 2
	total = 0
	for i, count := range h.Counts {
		total += count
		if total >= thresh {
			return h.Buckets[i]
		}
	}
	panic("medianBucket: should not happen")
}

func createStringList(input []string) []Object {
	list := make([]Object, len(input))
	for i, v := range input {
		list[i] = &Stringo{Value: v}
	}
	return list
}

// isError is the helper function to determine if an object is an error
func isError(obj Object) bool {
	if obj != nil {
		_, isError := obj.(*Error)
		return isError
	}
	return false
}

func blueObjToJsonObject(obj Object) Object {
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
	case MAP_OBJ:
		mObj := obj.(*Map)
		ok, err := checkMapObjPairsForValidJsonKeysAndValues(mObj.Pairs)
		if !ok {
			return newError("`to_json` error validating MAP object. %s", err.Error())
		}
		var buf bytes.Buffer
		jsonString := generateJsonStringFromValidMapObjPairs(buf, mObj.Pairs)
		return &Stringo{Value: jsonString.String()}
	case LIST_OBJ:
		lObj := obj.(*List)
		ok, err := checkListElementsForValidJsonValues(lObj.Elements)
		if !ok {
			return newError("`to_json` error validating LIST object. %s", err.Error())
		}
		var buf bytes.Buffer
		jsonString := generateJsonStringFromValidListElements(buf, lObj.Elements)
		return &Stringo{Value: jsonString.String()}
	default:
		// We can do this as the default because the arg is verified up above
		var buf bytes.Buffer
		jsonString := generateJsonStringFromOtherValidTypes(buf, obj)
		return &Stringo{Value: jsonString.String()}
	}
}

func isValidJsonValueType(t Type) bool {
	return t == STRING_OBJ || t == INTEGER_OBJ || t == UINTEGER_OBJ || t == BIG_FLOAT_OBJ || t == BIG_INTEGER_OBJ || t == FLOAT_OBJ || t == NULL_OBJ || t == BOOLEAN_OBJ || t == MAP_OBJ || t == LIST_OBJ
}

func checkListElementsForValidJsonValues(elements []Object) (bool, error) {
	for _, e := range elements {
		elType := e.Type()
		if !isValidJsonValueType(elType) {
			return false, errors.New("all values should be of the types STRING, INTEGER, FLOAT, BOOLEAN, NULL, LIST, or MAP. found " + string(elType))
		}
		if elType == LIST_OBJ {
			elements2 := e.(*List).Elements
			return checkListElementsForValidJsonValues(elements2)
		}
		if elType == MAP_OBJ {
			mapPairs := e.(*Map).Pairs
			return checkMapObjPairsForValidJsonKeysAndValues(mapPairs)
		}
	}
	return true, nil
}

func checkMapObjPairsForValidJsonKeysAndValues(pairs OrderedMap2[HashKey, MapPair]) (bool, error) {
	for _, hk := range pairs.Keys {
		mp, _ := pairs.Get(hk)
		keyType := mp.Key.Type()
		valueType := mp.Value.Type()
		if keyType != STRING_OBJ {
			return false, errors.New("all keys should be STRING, found invalid key type " + string(keyType))
		}
		if !isValidJsonValueType(valueType) {
			return false, errors.New("all values should be of the types STRING, INTEGER, FLOAT, BOOLEAN, NULL, LIST, or MAP. found " + string(valueType))
		}
		// These types are all okay but if its
		if valueType == MAP_OBJ {
			mapPairs := mp.Value.(*Map).Pairs
			return checkMapObjPairsForValidJsonKeysAndValues(mapPairs)
		}
		if valueType == LIST_OBJ {
			elements := mp.Value.(*List).Elements
			return checkListElementsForValidJsonValues(elements)
		}
	}
	// Empty pairs should be okay
	return true, nil
}

func generateJsonStringFromOtherValidTypes(buf bytes.Buffer, element Object) bytes.Buffer {
	switch t := element.(type) {
	case *Integer:
		buf.WriteString(fmt.Sprintf("%d", t.Value))
	case *UInteger:
		buf.WriteString(fmt.Sprintf("%d", t.Value))
	case *BigInteger:
		buf.WriteString(t.Value.String())
	case *BigFloat:
		buf.WriteString(t.Value.String())
	case *Stringo:
		buf.WriteString(fmt.Sprintf("%q", t.Value))
	case *Null:
		buf.WriteString("null")
	case *Float:
		buf.WriteString(fmt.Sprintf("%f", t.Value))
	case *Boolean:
		buf.WriteString(fmt.Sprintf("%t", t.Value))
	}
	return buf
}

func generateJsonStringFromValidListElements(buf bytes.Buffer, elements []Object) bytes.Buffer {
	buf.WriteRune('[')
	for i, e := range elements {
		elemType := e.Type()
		switch elemType {
		case LIST_OBJ:
			elements1 := e.(*List).Elements
			buf = generateJsonStringFromValidListElements(buf, elements1)
		case MAP_OBJ:
			pairs := e.(*Map).Pairs
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

func generateJsonStringFromValidMapObjPairs(buf bytes.Buffer, pairs OrderedMap2[HashKey, MapPair]) bytes.Buffer {
	buf.WriteRune('{')
	length := len(pairs.Keys)
	i := 0
	for _, hk := range pairs.Keys {
		mp, _ := pairs.Get(hk)
		buf.WriteString(fmt.Sprintf("%q:", mp.Key.Inspect()))
		valueType := mp.Value.Type()
		switch valueType {
		case MAP_OBJ:
			pairs1 := mp.Value.(*Map).Pairs
			buf = generateJsonStringFromValidMapObjPairs(buf, pairs1)
		case LIST_OBJ:
			elements := mp.Value.(*List).Elements
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

func parseMapLiteral(node *ast.MapLiteral) Object {
	pairs := NewPairsMapWithSize(len(node.Pairs))

	indices := []int{}
	for k := range node.PairsIndex {
		indices = append(indices, k)
	}
	sort.Ints(indices)
	for _, i := range indices {
		keyNode := node.PairsIndex[i]
		valueNode := node.Pairs[keyNode]
		// Should always be an *ast.StringLiteral
		key := ParseJson(keyNode)
		if isError(key) {
			return key
		}
		// Should always be true
		ok := IsHashable(key)
		if !ok {
			return newError("unusable as a map key: %s", key.Type())
		}
		hk := HashObject(key)
		hashed := HashKey{Type: key.Type(), Value: hk}

		value := ParseJson(valueNode)
		if isError(value) {
			return value
		}

		pairs.Set(hashed, MapPair{Key: key, Value: value})
	}

	return &Map{Pairs: pairs}
}

func parseListLiteral(node *ast.ListLiteral) Object {
	result := make([]Object, len(node.Elements))
	for i, e := range node.Elements {
		result[i] = ParseJson(e)
	}
	return &List{Elements: result}
}

func ParseJson(expr ast.Expression) Object {
	switch t := expr.(type) {
	case *ast.IntegerLiteral:
		return &Integer{Value: t.Value}
	case *ast.FloatLiteral:
		return &Float{Value: t.Value}
	case *ast.BigIntegerLiteral:
		return &BigInteger{Value: t.Value}
	case *ast.BigFloatLiteral:
		return &BigFloat{Value: t.Value}
	case *ast.Boolean:
		return nativeToBooleanObject(t.Value)
	case *ast.Null:
		return NULL
	case *ast.StringLiteral:
		return &Stringo{Value: t.Value}
	case *ast.MapLiteral:
		return parseMapLiteral(t)
	case *ast.ListLiteral:
		return parseListLiteral(t)
	case *ast.PrefixExpression:
		if t.TokenLiteral() != "-" {
			panic("Unexpected Prefix Expression Token " + t.TokenLiteral())
		}
		right := ParseJson(t.Right)
		switch rt := right.(type) {
		case *Integer:
			rt.Value = -rt.Value
		case *Float:
			rt.Value = -rt.Value
		case *BigInteger:
			bi := new(big.Int)
			rt.Value = bi.Neg(rt.Value)
		case *BigFloat:
			rt.Value = rt.Value.Neg()
		default:
			panic("Unexpected Type for Prefix Expression " + right.Type())
		}
		return right
	default:
		log.Fatalf("ParseJson: UNHANDLED t = %#+v (%T)", t, t)
	}
	panic("UNREACHABLE")
}
