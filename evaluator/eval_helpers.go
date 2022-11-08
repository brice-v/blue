package evaluator

import (
	"blue/ast"
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
	"os"
	"os/exec"
	"runtime"
	"strings"
	"unicode/utf8"

	"fyne.io/fyne/v2/widget"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
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
	l := lexer.New(s, "<internal: string>")
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

func decodeInterfaceToObject(value interface{}) object.Object {
	switch x := value.(type) {
	case int64:
		return &object.Integer{Value: x}
	case float64:
		return &object.Float{Value: x}
	case string:
		return &object.Stringo{Value: x}
	case bool:
		if x {
			return TRUE
		} else {
			return FALSE
		}
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
	default:
		log.Fatalf("HANDLE TYPE = %T", x)
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
			return nil, err
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
	log.Printf("v = %#v", v)
	for key, value := range v {
		returnMap[key] = decodeInterfaceToObject(value)
	}
	return returnMap, nil
}

func createHttpHandleBuiltinWithEvaluator(e *Evaluator) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newError("`handle` expects 4 aguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `handle` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `handle` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.FUNCTION_OBJ {
				return newError("argument 3 to `handle` should be FUNCTION. got=%s", args[2].Type())
			}
			if args[3].Type() != object.STRING_OBJ {
				return newError("argument 4 to `handle` should be STRING. got=%s", args[3].Type())
			}
			serverId := args[0].(*object.UInteger).Value
			app, ok := ServerMap.Get(serverId)
			if !ok {
				return newError("`handle` could not find Server Object")
			}
			method := strings.ToUpper(args[3].(*object.Stringo).Value)
			pattern := args[1].(*object.Stringo).Value
			fn := args[2].(*object.Function)
			switch method {
			case "GET":
				app.Get(pattern, func(c *fiber.Ctx) error {
					for k, v := range fn.DefaultParameters {
						if v != nil && fn.Parameters[k].Value == "query_params" {
							// Handle query_params
							if v.Type() != object.LIST_OBJ {
								return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("query_params must be LIST. got=%s", v.Type()))
							}
							l := v.(*object.List).Elements
							for _, elem := range l {
								if elem.Type() != object.STRING_OBJ {
									return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("query_params must be LIST of STRINGs. found=%s", elem.Type()))
								}
								// Now we know its a list of strings so we can set the variables accordingly for the fn
								s := elem.(*object.Stringo).Value
								fn.Env.Set(s, &object.Stringo{Value: c.Query(s)})
							}
						}
						// TODO: Otherwise chcek that it is null?
					}
					fnArgs := make([]object.Object, len(fn.Parameters))
					for i, v := range fn.Parameters {
						fnArgs[i] = &object.Stringo{Value: c.Params(v.Value)}
					}
					respObj := e.applyFunction(fn, fnArgs, make(map[string]object.Object))
					if respObj.Type() != object.STRING_OBJ {
						return c.Status(fiber.StatusInternalServerError).SendString("STRING NOT RETURNED FROM FUNCTION")
					}
					respStr := respObj.(*object.Stringo).Value
					if json.Valid([]byte(respStr)) {
						c.Set("Content-Type", "application/json")
						return c.Send([]byte(respStr))
					}
					return c.Format(respStr)
				})
			case "POST":
				app.Post(pattern, func(c *fiber.Ctx) error {
					for k, v := range fn.DefaultParameters {
						if v != nil && fn.Parameters[k].Value == "post_values" {
							// Handle post_values
							if v.Type() != object.LIST_OBJ {
								return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("post_values must be LIST. got=%s", v.Type()))
							}
							l := v.(*object.List).Elements

							contentType := c.Get("Content-Type")
							log.Printf("content-type = %s", contentType)
							body := strings.NewReader(string(c.Body()))
							log.Printf("body = %s", string(c.Body()))

							returnMap, err := decodeBodyToMap(contentType, body)
							if err != nil {
								return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("received input that could not be decoded in `%s`", string(c.Body())))
							}
							for _, elem := range l {
								if elem.Type() != object.STRING_OBJ {
									return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("post_values must be LIST of STRINGs. found=%s", elem.Type()))
								}
								s := elem.(*object.Stringo).Value
								if v, ok := returnMap[s]; ok {
									fn.Env.Set(s, v)
								} else {
									fn.Env.Set(s, &object.Stringo{Value: c.FormValue(s)})
								}
								// Now we know its a list of strings so we can set the variables accordingly for the fn
							}
						}
						// TODO: Otherwise chcek that it is null?
					}
					fnArgs := make([]object.Object, len(fn.Parameters))
					for i, v := range fn.Parameters {
						fnArgs[i] = &object.Stringo{Value: c.Params(v.Value)}
					}
					// TODO: Allow different things to be returned
					// TODO: Need to figure this out, it should be allowed to return anything Im pretty sure
					respObj := e.applyFunction(fn, fnArgs, make(map[string]object.Object))
					if respObj.Type() == object.STRING_OBJ {
						return c.SendString(respObj.(*object.Stringo).Value)
					}
					if respObj.Type() == object.NULL_OBJ {
						return c.SendStatus(fiber.StatusOK)
					} else {
						return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("POST Response Type is not NULL or STRING. got=%s", respObj.Type()))
					}
				})
				// case "PATCH":
				// case "POST":
				// case "DELETE":
			}
			return NULL
		},
	}
}

func createHttpHandleWSBuiltinWithEvaluator(e *Evaluator) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("`handle_ws` expects 3 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `handle_ws` should be UINTEGER. got=%s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("argument 2 to `handle_ws` should be STRING. got=%s", args[1].Type())
			}
			if args[2].Type() != object.FUNCTION_OBJ {
				return newError("argument 3 to `handle_ws` should be FUNCTION. got=%s", args[2].Type())
			}
			pattern := args[1].(*object.Stringo).Value
			fn := args[2].(*object.Function)
			if len(fn.Parameters) == 0 {
				return newError("function arguments should be at least 1 to store the websocket connection")
			}
			serverId := args[0].(*object.UInteger).Value
			app, ok := ServerMap.Get(serverId)
			if !ok {
				return newError("`handle_ws` could not find Server Object")
			}
			getRootPath := func(s string) string {
				x := strings.SplitAfterN(s, "/", 2)[1]
				x1 := strings.SplitAfterN(x, "/", 2)[0]
				var x2 string
				if strings.HasSuffix(x1, "/") {
					x2 = x1[:len(x1)-1]
				} else {
					x2 = x1
				}
				return "/" + x2
			}
			rootPath := getRootPath(pattern)
			app.Use(rootPath, func(c *fiber.Ctx) error {
				if websocket.IsWebSocketUpgrade(c) {
					return c.Next()
				}
				return fiber.ErrUpgradeRequired
			})

			var returnObj object.Object = NULL
			app.Get(pattern, websocket.New(func(c *websocket.Conn) {
				connCount := wsConnCount.Add(1)
				WSConnMap.Put(connCount, c)
				for k, v := range fn.DefaultParameters {
					if v != nil && fn.Parameters[k].Value == "query_params" {
						// Handle query_params
						if v.Type() != object.LIST_OBJ {
							log.Printf("query_params must be LIST. got=%s", v.Type())
							return
						}
						l := v.(*object.List).Elements
						for _, elem := range l {
							if elem.Type() != object.STRING_OBJ {
								log.Printf("query_params must be LIST of STRINGs. found=%s", elem.Type())
								return
							}
							// Now we know its a list of strings so we can set the variables accordingly for the fn
							s := elem.(*object.Stringo).Value
							fn.Env.Set(s, &object.Stringo{Value: c.Query(s)})
						}
					}
					// TODO: Otherwise chcek that it is null?
				}
				fnArgs := make([]object.Object, len(fn.Parameters))
				for i, v := range fn.Parameters {
					if i == 0 {
						fnArgs[i] = object.CreateBasicMapObject("ws", connCount)
					} else {
						fnArgs[i] = &object.Stringo{Value: c.Params(v.Value)}
					}
				}
				returnObj = e.applyFunction(fn, fnArgs, make(map[string]object.Object))
				log.Printf("`handle_ws` returned with %#v", returnObj)
				// TODO: On error what can we do?
			}))
			// TODO: Does the return obj above get returned here?
			return returnObj
		},
	}
}

func createUIButtonBuiltinWithEvaluator(e *Evaluator) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`button` expects 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `button` should be STRING. got=%s", args[0].Type())
			}
			if args[1].Type() != object.FUNCTION_OBJ {
				return newError("argument 2 to `button` should be FUNCTION. got=%s", args[1].Type())
			}
			s := args[0].(*object.Stringo).Value
			fn := args[1].(*object.Function)
			buttonId := uiCanvasObjectCount.Add(1)
			button := widget.NewButton(s, func() {
				obj := e.applyFunction(fn, []object.Object{}, make(map[string]object.Object))
				if isError(obj) {
					err := obj.(*object.Error)
					var buf bytes.Buffer
					buf.WriteString(err.Message)
					buf.WriteByte('\n')
					for e.ErrorTokens.Len() > 0 {
						buf.WriteString(fmt.Sprintf("%#v\n", e.ErrorTokens.PopBack()))
					}
					fmt.Printf("EvaluatorError: `button` click handler error: %s\n", err)
				}
			})
			UICanvasObjectMap.Put(buttonId, button)
			return object.CreateBasicMapObject("ui", buttonId)
		},
	}
}
