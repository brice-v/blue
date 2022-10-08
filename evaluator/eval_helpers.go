package evaluator

import (
	"blue/ast"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"unicode/utf8"
)

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// for now everything that is not null or false returns true
// TODO: Update this list to include non truthy for empty objects, lists, etc.
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

// newError is the wrapper function to add an error to the evaluator
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
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
	cleanedStrings := make([]string, 0)
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
		return newError("got 0 bytes from exec string output of `%s`.", str)
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
	return object.HashObject(leftList) == object.HashObject(rightList)
}

func nativeToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func (e *Evaluator) getBuiltinForDotCall(key string) (*object.Builtin, bool) {
	for b := e.Builtins.Front(); b != nil; b = b.Next() {
		switch t := b.Value.(type) {
		case BuiltinMapType:
			if builtin, isBuiltin := t.Get(key); isBuiltin {
				return builtin, isBuiltin
			}
		}
	}
	return nil, false
}

func (e *Evaluator) tryCreateValidBuiltinForDotCall(left, indx object.Object, leftNode ast.Expression) object.Object {
	// Try to see if the index being used is a builtin function
	if indx.Type() != object.STRING_OBJ {
		return nil
	}
	builtin, isBuiltin := e.getBuiltinForDotCall(indx.Inspect())
	envVar, isInEnv := e.env.Get(indx.Inspect())
	if !isBuiltin && !isInEnv {
		return nil
	}
	if isInEnv && envVar.Type() != object.FUNCTION_OBJ {
		return nil
	}
	// Allow either a string object or identifier to be passed to the builtin
	_, ok1 := left.(*object.Stringo)
	_, ok2 := leftNode.(*ast.Identifier)
	if !ok1 && !ok2 {
		return nil
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

func (e *Evaluator) applyFunction(fun object.Object, args []object.Object, defaultArgs map[string]object.Object) object.Object {
	argElem := e.UFCSArg.Pop()
	if argElem != nil {
		// prepend the argument to pass in to the front
		args = append([]object.Object{*argElem}, args...)
	}
	switch function := fun.(type) {
	case *object.Function:
		newE := New()
		newE.env = extendFunctionEnv(function, args, defaultArgs)
		evaluated := newE.Eval(function.Body)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return function.Fun(args...)
	default:
		return newError("not a function %s", function.Type())
	}
}

func extendFunctionEnv(fun *object.Function, args []object.Object, defaultArgs map[string]object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fun.Env)

	// If the arguments slice is the same length as the parameter list, then we have them all
	// so set them as normal
	if len(args) == len(fun.Parameters) {
		for paramIndx, param := range fun.Parameters {
			env.Set(param.Value, args[paramIndx])
		}
		setDefaultCallExpressionParameters(defaultArgs, env)
	} else if len(args) < len(fun.Parameters) {
		// loop and while less than the total parameters set environment variables accordingly
		argsIndx := 0
		for paramIndx, param := range fun.Parameters {
			if fun.DefaultParameters[paramIndx] == nil {
				if argsIndx < len(args) {
					env.Set(param.Value, args[argsIndx])
					argsIndx++
					continue
				}
			}
			env.Set(param.Value, fun.DefaultParameters[paramIndx])
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

func (e *Evaluator) EvalString(s string) (object.Object, error) {
	l := lexer.New(s)
	p := parser.New(l)
	prog := p.ParseProgram()
	pErrors := p.Errors()
	if len(pErrors) != 0 {
		for _, err := range pErrors {
			fmt.Printf("ParserError in `eval`: %s\n", err)
		}
		return nil, fmt.Errorf("failed to `eval` string, found '%d' parser errors", len(pErrors))
	}
	result := e.Eval(prog)
	return result, nil
}

func MakeEmptyList() object.Object {
	return &object.List{
		Elements: make([]object.Object, 0),
	}
}

func isValidJsonValueType(t object.Type) bool {
	return t == object.STRING_OBJ || t == object.INTEGER_OBJ || t == object.FLOAT_OBJ || t == object.NULL_OBJ || t == object.BOOLEAN_OBJ || t == object.MAP_OBJ || t == object.LIST_OBJ
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
		// log.Printf("got mp.Key = %#v, mp.Value = %#v", mp.Key, mp.Value)
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
