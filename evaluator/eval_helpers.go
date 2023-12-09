package evaluator

import (
	"blue/ast"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/clbanning/mxj/v2"
	"github.com/gookit/color"
)

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// for now everything that is not null or false returns true
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
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

// newError is the wrapper function to add an error to the evaluator
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func newPositionalTypeError(funName string, pos int, expectedType string, currentType object.Type) *object.Error {
	return newError("PositionalTypeError: `%s` expects argument %d to be %s. got=%s", funName, pos, expectedType, currentType)
}

func newPositionalTypeErrorForGoObj(funName string, pos int, expectedType string, currentObj any) *object.Error {
	return newError("PositionalTypeError: `%s` expects argument %d to be %s. got=%T", funName, pos, expectedType, currentObj)
}

func newInvalidArgCountError(funName string, got, want int, otherCount string) *object.Error {
	if otherCount == "" {
		return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d", funName, got, want)
	}
	return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d %s", funName, got, want, otherCount)
}

// isError is the helper function to determine if an object is an error
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func ExecStringCommand(str string) object.Object {
	splitStr := strings.Split(str, " ")
	if len(splitStr) == 0 {
		return newError("unable to exec the string `%s`", str)
	}
	if len(splitStr) == 1 {
		output, err := execCommand(splitStr[0]).Output()
		if err != nil {
			return newError("unable to exec the string `%s`. Error: %s", str, err)
		}
		return &object.Stringo{Value: string(output[:])}
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
	return &object.Stringo{Value: string(output[:])}
}

func execCommand(arg0 string, args ...string) *exec.Cmd {
	if args == nil {
		if runtime.GOOS == "windows" {
			winArgs := []string{"/c"}
			winArgs = append(winArgs, arg0)
			return exec.Command("cmd", winArgs...)
		}
		return exec.Command(arg0)
	}
	if runtime.GOOS == "windows" {
		winArgs := []string{"/c"}
		winArgs = append(winArgs, arg0)
		winArgs = append(winArgs, args...)
		return exec.Command("cmd", winArgs...)
	}
	return exec.Command(arg0, args...)
}

func twoListsEqual(leftList, rightList *object.List) bool {
	// This is a deep equality expensive function
	if len(leftList.Elements) != len(rightList.Elements) {
		return false
	}
	return object.HashObject(leftList) == object.HashObject(rightList)
}

func nativeToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func (e *Evaluator) getBuiltinForDotCall(key string) (*object.Builtin, bool) {
	for _, b := range e.Builtins {
		if builtin, isBuiltin := b.Get(key); isBuiltin {
			return builtin, isBuiltin
		}

	}
	return nil, false
}

func (e *Evaluator) tryCreateValidDotCall(left, indx object.Object, leftNode ast.Expression) object.Object {
	// Try to see if the index being used is a builtin function
	if indx.Type() != object.STRING_OBJ {
		return nil
	}
	builtin, isBuiltin := e.getBuiltinForDotCall(indx.Inspect())
	envVar, isInEnv := e.env.Get(indx.Inspect())
	if !isBuiltin && !isInEnv {
		msg := fmt.Sprintf("index `%s` is not in environment", indx.Inspect())
		e.maybeNullMapFnCall.Push(msg)
		return nil
	}
	if isInEnv && envVar.Type() != object.FUNCTION_OBJ {
		return nil
	}
	// Allow either a string, or collection type to be passed to the builtin
	ident, isIdent := leftNode.(*ast.Identifier)
	isValidType := false
	switch x := left.(type) {
	case *object.List, *object.Set, *object.Stringo:
		isValidType = true
	case *object.Map:
		isValidType = true
		hashKey := object.HashObject(indx)
		hk := object.HashKey{
			Type:  object.STRING_OBJ,
			Value: hashKey,
		}
		// If the Pairs has the index as a key then this is not valid for UFCS
		if _, ok := x.Pairs.Get(hk); ok {
			return nil
		}
	default:
		isValidType = false
	}
	if !isValidType && !isIdent {
		return nil
	}
	// If its immutable and the function can mutate than return an error
	if isIdent && e.env.IsImmutable(ident.Value) {
		if isBuiltin && builtin.Mutates {
			return newError("'%s' is immutable", ident.Value)
		}
		e.UFCSArgIsImmutable.Push(true)
	} else {
		e.UFCSArgIsImmutable.Push(false)
	}
	e.UFCSArg.Push(&left)
	// Return the builtin function object so that it can be used in the call
	// expression
	if isBuiltin {
		return builtin
	} else {
		return envVar.(*object.Function)
	}
}

func checkFunctionArgsAreValid(fun *object.Function, args []object.Object, defaultArgs map[string]object.Object, isFirstArgUFCSMod bool) object.Object {
	defaultParamCount := 0
	for _, v := range fun.DefaultParameters {
		if v != nil {
			defaultParamCount++
		}
	}
	defaultArgCount := 0
	for _, v := range defaultArgs {
		if v != nil {
			defaultArgCount++
		}
	}
	argLen := len(args)
	paramLen := len(fun.Parameters)
	countToCheck := argLen + defaultArgCount
	if isFirstArgUFCSMod {
		countToCheck -= 1
	}
	if countToCheck > paramLen {
		return newError("function called with too many arguments")
	}

	if argLen+defaultParamCount+defaultArgCount < paramLen {
		return newError("function called without enough arguments")
	}
	defaultArgToDefaultParamMap := make(map[string]struct{})
	for i, k := range fun.Parameters {
		if len(fun.DefaultParameters) == paramLen {
			value := fun.DefaultParameters[i]
			if value != nil {
				defaultArgToDefaultParamMap[k.Value] = struct{}{}
			}
		}
	}
	for k := range defaultArgs {
		if _, ok := defaultArgToDefaultParamMap[k]; !ok {
			return newError("function called with default argument that is not in default function parameters")
		}
	}

	return nil
}

func (e *Evaluator) applyFunction(fun object.Object, args []object.Object, defaultArgs map[string]object.Object, immutableArgs []bool) object.Object {
	argElem := e.UFCSArg.Pop()
	// Note: This is just to keep the UFCS stack size consistent for both
	_ = e.UFCSArgIsImmutable.Pop()
	firstArgUFCSIsMod := false
	if argElem != nil {
		argElemType := (*argElem).Type()
		firstArgUFCSIsMod = argElemType == object.MODULE_OBJ
		// prepend the argument to pass in to the front
		args = append([]object.Object{*argElem}, args...)
	}
	switch function := fun.(type) {
	case *object.Function:
		err := checkFunctionArgsAreValid(function, args, defaultArgs, firstArgUFCSIsMod)
		if err != nil {
			return err
		}
		newE := New()
		// Always use our current evaluator PID for the function
		// this allows the self() function to return properly inside evaluated
		// functions because spawnExpression will set the PID initially
		newE.PID = e.PID
		newE.env = extendFunctionEnv(function, args, defaultArgs, immutableArgs)
		evaluated := newE.Eval(function.Body)
		for newE.ErrorTokens.Len() != 0 {
			e.ErrorTokens.s.PushBack(newE.ErrorTokens.Pop())
		}
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return function.Fun(args...)
	default:
		msg := fmt.Sprintf("not a function %s", function.Type())
		if function.Type() == NULL.Type() && e.maybeNullMapFnCall.Peek() != "" {
			extraMsg := e.maybeNullMapFnCall.Pop()
			msg = fmt.Sprintf("%s, %s", msg, extraMsg)
		}
		return newError(msg)
	}
}

func extendFunctionEnv(fun *object.Function, args []object.Object, defaultArgs map[string]object.Object, immutableArgs []bool) *object.Environment {
	env := object.NewEnclosedEnvironment(fun.Env)

	// If the arguments slice is the same length as the parameter list, then we have them all
	// so set them as normal
	if len(args) == len(fun.Parameters) {
		for paramIndx, param := range fun.Parameters {
			env.Set(param.Value, args[paramIndx])
			if immutableArgs[paramIndx] {
				env.ImmutableSet(param.Value)
			}
		}
		setDefaultCallExpressionParameters(defaultArgs, env)
	} else if len(args) < len(fun.Parameters) {
		// loop and while less than the total parameters set environment variables accordingly
		argsIndx := 0
		countDefaultParams := 0
		for _, param := range fun.DefaultParameters {
			if param != nil {
				countDefaultParams++
			}
		}
		for paramIndx, param := range fun.Parameters {
			_, defaultArgGiven := defaultArgs[param.Value]
			if fun.DefaultParameters[paramIndx] == nil {
				if argsIndx < len(args) {
					env.Set(param.Value, args[argsIndx])
					if immutableArgs[argsIndx] {
						env.ImmutableSet(param.Value)
					}
					argsIndx++
					continue
				}
			} else if countDefaultParams+len(args) == len(fun.Parameters) && !defaultArgGiven {
				// If the Count of the defaultParams+args given equals the length of the fun.Parameters
				// then we want to pass in the arg where the fun.DefaultParameter is nil for that indx
				env.Set(param.Value, fun.DefaultParameters[paramIndx])
				continue
			} else {
				// If there is a default param for every arg then we add in
				// regular args as they are given
				// defaultArgs also needs to be non-empty and the number of default params
				// should be greater than the number of args passed in (if we are going
				// to populate it)
				if !defaultArgGiven && countDefaultParams >= len(args) {
					// It needs to be not present as a default arg - otherwise
					// that value will be used
					if argsIndx < len(args) {
						env.Set(param.Value, args[argsIndx])
						if immutableArgs[argsIndx] {
							env.ImmutableSet(param.Value)
						}
						argsIndx++
						continue
					}
				} else if !defaultArgGiven && countDefaultParams < len(args) {
					if argsIndx < len(args) {
						// This is so if we have an extra arg to set we set it over the default param which happens right below
						env.Set(param.Value, args[argsIndx])
						if immutableArgs[argsIndx] {
							env.ImmutableSet(param.Value)
						}
						argsIndx++
						continue
					}
				}
			}
			// Saw that this set [] as the ident to a value but I think it was using an incorrect arg - regardless this should work
			identStr := param.String()
			env.Set(identStr, fun.DefaultParameters[paramIndx])
		}
		setDefaultCallExpressionParameters(defaultArgs, env)
	}
	return env
}

func setDefaultCallExpressionParameters(defaultArgs map[string]object.Object, env *object.Environment) {
	for k, v := range defaultArgs {
		_, ok := env.Get(k)
		if ok {
			env.Set(k, v)
		}
	}
}

func createHelpStringFromBodyTokens(functionName string, funObj *object.Function, helpStrTokens []string) string {
	explanation := ""
	if len(helpStrTokens) == 1 {
		explanation = helpStrTokens[0]
	} else if len(helpStrTokens) == 0 {
		explanation = ""
	} else {
		explanation = strings.Join(helpStrTokens, "\n")
	}
	return fmt.Sprintf("%s\n\ntype(%s) = '%s'\ninspect(%s) = '%s'", explanation, functionName, funObj.Type(), functionName, funObj.Inspect())
}

func CreateHelpStringFromProgramTokens(modName string, helpStrTokens []string, pubFunHelpStr string) string {
	explanation := ""
	if len(helpStrTokens) == 1 {
		explanation = helpStrTokens[0]
	} else if len(helpStrTokens) == 0 {
		explanation = ""
	} else {
		explanation = strings.Join(helpStrTokens, "\n")
	}
	consts.DisableColorIfNoColorEnvVarSet()
	green := color.FgGreen.Render
	bold := color.Bold.Render
	blue := color.FgCyan.Render
	firstPart := fmt.Sprintf("%s`%s`: %s", blue(bold("MODULE ")), blue(bold(modName)), blue(bold(explanation)))
	return fmt.Sprintf("%s\n\ntype(%s) = '%s'\n\n%s:%s", firstPart, modName, object.MODULE_OBJ, bold(green("PUBLIC FUNCTIONS")), pubFunHelpStr)
}

func (e *Evaluator) createFilePathFromImportPath(importPath string) string {
	var fpath bytes.Buffer
	if e.EvalBasePath != "." {
		fpath.WriteString(e.EvalBasePath)
		fpath.WriteString(string(os.PathSeparator))
	}
	importPath = strings.ReplaceAll(importPath, ".", string(os.PathSeparator))
	fpath.WriteString(importPath)
	fpath.WriteString(".b")
	return fpath.String()
}

func doCondAndMatchExpEqual(condVal, matchVal object.Object) bool {
	condValPairs := condVal.(*object.Map).Pairs
	matchValPairs := matchVal.(*object.Map).Pairs
	condValLen := condValPairs.Len()
	matchValLen := matchValPairs.Len()
	if condValLen != matchValLen {
		return false
	}
	for _, condKey := range condValPairs.Keys {
		condValue, _ := condValPairs.Get(condKey)
		_, ok := matchValPairs.Get(condKey)
		if !ok {
			return false
		}
		if condValue.Value == IGNORE {
			continue
		}
		val, ok := matchValPairs.Get(condKey)
		if !ok {
			return false
		}
		if object.HashObject(val.Value) != object.HashObject(condValue.Value) {
			return false
		}
	}

	return true
}

func runeLen(str string) int {
	return utf8.RuneCountInString(str)
}

// isBooleanOperator checks if the given operator is considered a 'boolean' operator
// this currently includes ==, !=, and, or, not
// Note: not is a prefix operator and the rest are infix (notin and in technically
// are boolean ops but we dont include them here)
func isBooleanOperator(operator string) bool {
	return operator == "==" || operator == "!=" || operator == "and" || operator == "or" || operator == "not"
}

func (e *Evaluator) EvalString(s string) (object.Object, error) {
	l := lexer.New(s, "<internal: string>")
	p := parser.New(l)
	prog := p.ParseProgram()
	pErrors := p.Errors()
	if len(pErrors) != 0 {
		for _, err := range pErrors {
			consts.ErrorPrinter("ParserError in `eval`: %s\n", err)
		}
		return nil, fmt.Errorf("failed to `eval` string, found '%d' parser errors", len(pErrors))
	}
	result := e.Eval(prog)
	return result, nil
}

func MakeEmptyList() object.Object {
	return &object.List{
		Elements: []object.Object{},
	}
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

func generateJsonStringFromOtherValidTypes(buf bytes.Buffer, element object.Object) bytes.Buffer {
	switch t := element.(type) {
	case *object.Integer:
		buf.WriteString(fmt.Sprintf("%d", t.Value))
	case *object.UInteger:
		buf.WriteString(fmt.Sprintf("%d", t.Value))
	case *object.BigInteger:
		buf.WriteString(t.Value.String())
	case *object.BigFloat:
		buf.WriteString(t.Value.String())
	case *object.Stringo:
		buf.WriteString(fmt.Sprintf("%q", t.Value))
	case *object.Null:
		buf.WriteString("null")
	case *object.Float:
		buf.WriteString(fmt.Sprintf("%f", t.Value))
	case *object.Boolean:
		buf.WriteString(fmt.Sprintf("%t", t.Value))
	}
	return buf
}

func generateJsonStringFromValidListElements(buf bytes.Buffer, elements []object.Object) bytes.Buffer {
	buf.WriteRune('[')
	for i, e := range elements {
		elemType := e.Type()
		if elemType == object.LIST_OBJ {
			elements1 := e.(*object.List).Elements
			buf = generateJsonStringFromValidListElements(buf, elements1)
		} else if elemType == object.MAP_OBJ {
			pairs := e.(*object.Map).Pairs
			buf = generateJsonStringFromValidMapObjPairs(buf, pairs)
		} else {
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
		if valueType == object.MAP_OBJ {
			pairs1 := mp.Value.(*object.Map).Pairs
			buf = generateJsonStringFromValidMapObjPairs(buf, pairs1)
		} else if valueType == object.LIST_OBJ {
			elements := mp.Value.(*object.List).Elements
			buf = generateJsonStringFromValidListElements(buf, elements)
		} else {
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

func decodeInterfaceToObject(value interface{}) object.Object {
	switch x := value.(type) {
	case int64:
		return &object.Integer{Value: x}
	case float64:
		return &object.Float{Value: x}
	case string:
		return &object.Stringo{Value: x}
	case bool:
		return nativeToBooleanObject(x)
	case []interface{}:
		list := &object.List{Elements: make([]object.Object, len(x))}
		for i, e := range x {
			list.Elements[i] = decodeInterfaceToObject(e)
		}
		return list
	case map[string]interface{}:
		mapObj := object.NewOrderedMap[string, object.Object]()
		for k, v := range x {
			mapObj.Set(k, decodeInterfaceToObject(v))
		}
		return object.CreateMapObjectForGoMap(*mapObj)
	case *object.OrderedMap2[string, interface{}]:
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
	return NULL
}

func decodeBodyToMap(contentType string, body io.Reader) (map[string]object.Object, error) {
	returnMap := make(map[string]object.Object)
	var v map[string]interface{}
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
					v = make(map[string]interface{})
				}
				for k1, v1 := range mv {
					v[k1] = v1
				}
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

func blueObjectToGoObject(blueObject object.Object) (interface{}, error) {
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
		pairs := object.NewOrderedMap[string, interface{}]()
		for _, k := range m.Pairs.Keys {
			mp, _ := m.Pairs.Get(k)
			if mp.Key.Type() != object.STRING_OBJ {
				return nil, fmt.Errorf("blueObjectToGoObject: Map must only have string keys. got=%s", mp.Key.Type())
			}
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
	case object.LIST_OBJ:
		l := blueObject.(*object.List).Elements
		elements := make([]interface{}, len(l))
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
	default:
		return nil, fmt.Errorf("blueObjectToGoObject: TODO: Type currently unsupported: %s (%T)", blueObject.Type(), blueObject)
	}
}

// goObjectToBlueObject will only work for simple go types
func goObjectToBlueObject(goObject interface{}) (object.Object, error) {
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
		return NULL, nil
	case []interface{}:
		l := &object.List{Elements: make([]object.Object, len(obj))}
		for i, e := range obj {
			val, err := goObjectToBlueObject(e)
			if err != nil {
				return nil, err
			}
			l.Elements[i] = val
		}
		return l, nil
	case map[string]interface{}:
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
	case *object.OrderedMap2[string, interface{}]:
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
	default:
		return nil, fmt.Errorf("goObjectToBlueObject: TODO: Type currently unsupported: (%T)", obj)
	}
}

func getBlueObjectFromResp(resp *http.Response) object.Object {
	_body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer resp.Body.Close()
	body := &object.Stringo{Value: string(_body)}
	contentLength := &object.Integer{Value: resp.ContentLength}
	headersToMapObj := func(header http.Header) object.Object {
		mapObj := object.NewOrderedMap[string, object.Object]()
		for k, v := range header {
			list := &object.List{Elements: make([]object.Object, len(v))}
			for i, e := range v {
				list.Elements[i] = &object.Stringo{Value: e}
			}
			mapObj.Set(k, list)
		}
		return object.CreateMapObjectForGoMap(*mapObj)
	}
	headers := headersToMapObj(resp.Header)
	proto := &object.Stringo{Value: resp.Proto}
	requestToMapObj := func(req *http.Request) object.Object {
		mapObj := object.NewOrderedMap[string, object.Object]()
		mapObj.Set("method", &object.Stringo{Value: req.Method})
		mapObj.Set("proto", &object.Stringo{Value: req.Proto})
		mapObj.Set("url", &object.Stringo{Value: req.URL.String()})
		return object.CreateMapObjectForGoMap(*mapObj)
	}
	request := requestToMapObj(resp.Request)
	rawStatus := &object.Stringo{Value: resp.Status}
	status := &object.Integer{Value: int64(resp.StatusCode)}

	trailer := headersToMapObj(resp.Trailer)
	transferEncoding := &object.List{Elements: make([]object.Object, len(resp.TransferEncoding))}
	for i, v := range resp.TransferEncoding {
		transferEncoding.Elements[i] = &object.Stringo{Value: v}
	}
	uncompressed := nativeToBooleanObject(resp.Uncompressed)
	_cookies := resp.Cookies()
	cookieToMapObj := func(c *http.Cookie) object.Object {
		mapObj := object.NewOrderedMap[string, object.Object]()
		mapObj.Set("name", &object.Stringo{Value: c.Name})
		mapObj.Set("value", &object.Stringo{Value: c.Value})
		mapObj.Set("path", &object.Stringo{Value: c.Path})
		mapObj.Set("domain", &object.Stringo{Value: c.Domain})
		mapObj.Set("expires", &object.Integer{Value: c.Expires.Unix()})
		mapObj.Set("http_only", nativeToBooleanObject(c.HttpOnly))
		mapObj.Set("secure", nativeToBooleanObject(c.Secure))
		mapObj.Set("samesite", &object.Integer{Value: int64(c.SameSite)})
		mapObj.Set("raw", &object.Stringo{Value: c.String()})
		return object.CreateMapObjectForGoMap(*mapObj)
	}
	cookies := &object.List{Elements: make([]object.Object, len(_cookies))}
	for i, c := range _cookies {
		cookies.Elements[i] = cookieToMapObj(c)
	}

	returnMap := object.NewOrderedMap[string, object.Object]()
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

	return object.CreateMapObjectForGoMap(*returnMap)
}

// For Builtins

func getErrorTokenTraceAsJson(e *Evaluator) interface{} {
	return getErrorTokenTraceAsJsonWithError(e, "")
}

func getErrorTokenTraceAsJsonWithError(e *Evaluator, errorMsg string) interface{} {
	var disableHttpServerDebug bool
	disableHttpServerDebugStr := os.Getenv(consts.DISABLE_HTTP_SERVER_DEBUG)
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
		for e.ErrorTokens.Len() > 0 {
			firstPart, carat := lexer.GetErrorLineMessageForJson(e.ErrorTokens.PopBack())
			errors = append(errors, firstPart, carat)
		}
		fmt.Println("`http handler` error: " + errorMsg)
		for _, err := range errors {
			fmt.Printf("%s\n", err)
		}
	}
	return errors
}

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
				if args[2].Type() != object.NULL_OBJ && args[2].Type() != object.FUNCTION_OBJ {
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
				// Using custom comparator
				anys := make([]interface{}, len(l.Elements))
				for i, e := range l.Elements {
					obj, err := blueObjectToGoObject(e)
					if err != nil {
						return newError("`sort` key error: %s", err.Error())
					}
					anys[i] = obj
				}
				fun := args[2].(*object.Function)
				if len(fun.Parameters) != 1 {
					return newError("`sort` key error: key function must take 1 arg. got=%d", len(fun.Parameters))
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
					biObj := e.applyFunction(fun, []object.Object{aib}, make(map[string]object.Object), []bool{false})
					if isError(biObj) {
						err := biObj.(*object.Error)
						var buf bytes.Buffer
						buf.WriteString(err.Message)
						buf.WriteByte('\n')
						for e.ErrorTokens.Len() > 0 {
							tok := e.ErrorTokens.PopBack()
							buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
						}
						fmt.Printf("%s`sort` key error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
						return false
					}
					if biObj.Type() != object.FLOAT_OBJ && biObj.Type() != object.INTEGER_OBJ && biObj.Type() != object.STRING_OBJ {
						fmt.Printf("%s`sort` key error: key function must return INTEGER, STRING, or FLOAT\n", consts.EVAL_ERROR_PREFIX)
						return false
					}
					bjObj := e.applyFunction(fun, []object.Object{ajb}, make(map[string]object.Object), []bool{false})
					if isError(bjObj) {
						err := bjObj.(*object.Error)
						var buf bytes.Buffer
						buf.WriteString(err.Message)
						buf.WriteByte('\n')
						for e.ErrorTokens.Len() > 0 {
							tok := e.ErrorTokens.PopBack()
							buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
						}
						fmt.Printf("%s`sort` key error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
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
								return leftO > rightO
							}
							return leftO < rightO
						}
					}
					if leftO, ok := left.(int); ok {
						if rightO, ok := right.(int); ok {
							if shouldReverse {
								return leftO > rightO
							}
							return leftO < rightO
						}
					}
					if leftO, ok := left.(float64); ok {
						if rightO, ok := right.(float64); ok {
							if shouldReverse {
								return leftO > rightO
							}
							return leftO < rightO
						}
					}
					if leftO, ok := left.(string); ok {
						if rightO, ok := right.(string); ok {
							if shouldReverse {
								return leftO > rightO
							}
							return leftO < rightO
						}
					}
					fmt.Printf("%s`sort` key error: key function returned mismatched types: i = %d (%T), j = %d (%T)\n", consts.EVAL_ERROR_PREFIX, i, left, j, right)
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
		}
	}
	return sortBuiltin
}

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
					obj := e.applyFunction(fn, []object.Object{}, make(map[string]object.Object), []bool{})
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
					obj := e.applyFunction(fn, []object.Object{nativeToBooleanObject(value)}, make(map[string]object.Object), []bool{true})
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
					obj := e.applyFunction(fn, []object.Object{&object.Stringo{Value: value}}, make(map[string]object.Object), []bool{true})
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
					obj := e.applyFunction(fn, []object.Object{&object.Stringo{Value: value}}, make(map[string]object.Object), []bool{true})
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
						obj := e.applyFunction(fn, []object.Object{}, make(map[string]object.Object), []bool{})
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
					obj := e.applyFunction(fn, []object.Object{}, make(map[string]object.Object), []bool{})
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
		}
	}
	return uiToolbarAction
}

// Helper for `doc` command
func (e *Evaluator) GetAllStdPublicFunctionHelpStrings() string {
	mods := make([]string, len(_std_mods))
	i := 0
	for mod := range _std_mods {
		mods[i] = mod
		i++
	}
	// Sort by key to always have the docs in order
	sort.Strings(mods)
	var out bytes.Buffer
	for i, mod := range mods {
		// Calling the function like this prevents importing specific pub vars from the module
		_ = e.AddStdLibToEnv(mod, nil, false)
		modObj, ok := e.env.Get(mod)
		if !ok {
			panic("should not fail - mod '" + mod + "' should already be added to env")
		}
		out.WriteString(modObj.Help())
		out.WriteByte('\n')
		if i != len(mods) {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func (e *Evaluator) GetStdModPublicFunctionHelpString(modName string) string {
	if !e.IsStd(modName) {
		panic("should not fail - mod '" + modName + "' should already be verified by caller")
	}
	// Calling the function like this prevents importing specific pub vars from the module
	_ = e.AddStdLibToEnv(modName, nil, false)
	modObj, ok := e.env.Get(modName)
	if !ok {
		panic("should not fail - mod '" + modName + "' should already be added to env")
	}
	return modObj.Help() + "\n"
}

func (e *Evaluator) GetPublicFunctionHelpString() string {
	// passthrough for use by `doc` command
	return e.env.GetPublicFunctionHelpString()
}
