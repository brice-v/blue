package evaluator

import (
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"golang.org/x/net/html"
)

func createHttpHandleBuiltin(e *Evaluator) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("handle", len(args), 4, "")
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newPositionalTypeError("handle", 1, object.UINTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("handle", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.FUNCTION_OBJ {
				return newPositionalTypeError("handle", 3, object.FUNCTION_OBJ, args[2].Type())
			}
			if args[3].Type() != object.STRING_OBJ {
				return newPositionalTypeError("handle", 4, object.STRING_OBJ, args[3].Type())
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
					return processHandlerFn(e, fn, c, method)
				})
			case "POST":
				app.Post(pattern, func(c *fiber.Ctx) error {
					return processHandlerFn(e, fn, c, method)
				})
			case "PATCH":
				app.Patch(pattern, func(c *fiber.Ctx) error {
					return processHandlerFn(e, fn, c, method)
				})
			case "PUT":
				app.Put(pattern, func(c *fiber.Ctx) error {
					return processHandlerFn(e, fn, c, method)
				})
			case "DELETE":
				app.Delete(pattern, func(c *fiber.Ctx) error {
					return processHandlerFn(e, fn, c, method)
				})
			}
			return NULL
		},
	}
}

func tryGetHttpActionAndMap(respObj object.Object) (isAction bool, action string, m map[string]interface{}) {
	isAction, action, m = false, "", nil
	mObj, err := blueObjectToGoObject(respObj)
	if err == nil {
		if mm, ok := mObj.(map[string]interface{}); ok {
			if kt, ok := mm["t"]; ok {
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

func processHandlerFn(e *Evaluator, fn *object.Function, c *fiber.Ctx, method string) error {
	ok, respObj, errors := prepareAndApplyHttpHandleFn(e, fn, c, method)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(errors)
	}
	// First check if the respObj is a MAP and if its a valid http handler response action
	if respObj.Type() == object.MAP_OBJ {
		isAction, action, m := tryGetHttpActionAndMap(respObj)
		if isAction {
			switch action {
			case "status":
				code, ok := m["code"].(int64)
				if !ok {
					err := fmt.Sprintf("http/status 'code' must be INTEGER. got=%T", m["code"])
					return c.Status(fiber.StatusInternalServerError).JSON(err)
				}
				return c.SendStatus(int(code))
			case "redirect":
				location, ok := m["location"].(string)
				if !ok {
					err := fmt.Sprintf("http/redirect 'location' must be STRING. got=%T", m["location"])
					return c.Status(fiber.StatusInternalServerError).JSON(err)
				}
				code, ok := m["code"].(int64)
				if !ok {
					err := fmt.Sprintf("http/redirect code 'must' be INTEGER. got=%T", m["code"])
					return c.Status(fiber.StatusInternalServerError).JSON(err)
				}
				return c.Redirect(location, int(code))
			case "next":
				return c.Next()
			}
		}
	}
	if method != "GET" {
		if respObj.Type() == object.STRING_OBJ {
			return c.SendString(respObj.(*object.Stringo).Value)
		}
		if respObj.Type() == object.NULL_OBJ {
			return c.SendStatus(fiber.StatusOK)
		} else {
			obj := blueObjToJsonObject(respObj)
			if obj.Type() == object.ERROR_OBJ {
				errors := getErrorTokenTraceAsJson(e).([]string)
				errors = append(errors, fmt.Sprintf("%s Response Type is not STRING, valid JSON, or NULL. got=%s", method, obj.Type()))
				return c.Status(fiber.StatusInternalServerError).JSON(errors)
			} else {
				if respStr, ok := obj.(*object.Stringo); ok {
					respStrBs := []byte(respStr.Value)
					if json.Valid(respStrBs) {
						c.Set("Content-Type", "application/json")
						return c.Send(respStrBs)
					}
				}
			}
			errors := getErrorTokenTraceAsJson(e).([]string)
			errors = append(errors, fmt.Sprintf("%s Response Type is not NULL or STRING. got=%s", method, respObj.Type()))
			return c.Status(fiber.StatusInternalServerError).JSON(errors)
		}
	} else {
		if respObj.Type() == object.STRING_OBJ {
			respStr := respObj.(*object.Stringo).Value
			log.Printf("respStr = %s", respStr)
			respStrBs := []byte(respStr)
			if json.Valid(respStrBs) {
				c.Set("Content-Type", "application/json")
				log.Printf("sending as json?")
				return c.Send(respStrBs)
			}
			// If this is a <html></html> snippet being returned then we will manually set
			// the content type so that other things could be included in the <head>
			if strings.HasPrefix(strings.TrimLeft(respStr, "\n\r \t"), "<html") {
				if strings.HasSuffix(strings.TrimRight(respStr, "\n\r \t"), "</html>") {
					_, err := html.Parse(strings.NewReader(respStr))
					if err == nil {
						// This will allow things like <head> to be properly populated
						c.Set("Content-Type", "text/html")
						return c.Send(respStrBs)
					}
				}
			}
			return c.Format(respStr)
		} else {
			// If the value returned here would be a valid JSON root node then we will return it
			// assuming it all works (ie. if a list - all the values are valid JSON)
			obj := blueObjToJsonObject(respObj)
			if obj.Type() == object.ERROR_OBJ {
				errors := getErrorTokenTraceAsJson(e).([]string)
				errors = append(errors, "error converting object to JSON")
				return c.Status(fiber.StatusInternalServerError).JSON(errors)
			}
			if respStr, ok := obj.(*object.Stringo); ok {
				respStrBs := []byte(respStr.Value)
				if json.Valid(respStrBs) {
					c.Set("Content-Type", "application/json")
					return c.Send(respStrBs)
				}
			}
			errors := getErrorTokenTraceAsJson(e).([]string)
			errors = append(errors, "STRING NOT RETURNED FROM JSON CONVERSION")
			return c.Status(fiber.StatusInternalServerError).JSON(errors)
		}
	}
}

func prepareAndApplyHttpHandleFn(e *Evaluator, fn *object.Function, c *fiber.Ctx, method string) (bool, object.Object, []string) {
	isGet := method == "GET"
	isDelete := method == "DELETE"
	methodLower := strings.ToLower(method)
	if !isGet && !isDelete {
		ok, errors := getAndSetDefaultHttpParams(e, methodLower+"_values", fn, c)
		if !ok {
			return false, nil, errors
		}
	}
	ok, errors := getAndSetDefaultHttpParams(e, "query_params", fn, c)
	if !ok {
		return false, nil, errors
	}
	fnArgs, immutableArgs := getAndSetHttpParams(e, fn, c)
	// TODO: Allow different things to be returned
	// TODO: Need to figure this out, it should be allowed to return anything Im pretty sure
	return true, e.applyFunction(fn, fnArgs, make(map[string]object.Object), immutableArgs), []string{}
}

func getAndSetDefaultHttpParams(e *Evaluator, varName string, fn *object.Function, c *fiber.Ctx) (bool, []string) {
	for k, v := range fn.DefaultParameters {
		isQueryParams := v != nil && fn.Parameters[k].Value == "query_params"
		isCookies := v != nil && fn.Parameters[k].Value == "cookies"
		if v != nil {
			if isQueryParams {
				// Handle query_params
				if v.Type() != object.LIST_OBJ {
					errors := getErrorTokenTraceAsJson(e).([]string)
					errors = append(errors, fmt.Sprintf("query_params must be LIST. got=%s", v.Type()))
					return false, errors
				}
				l := v.(*object.List).Elements
				for _, elem := range l {
					if elem.Type() != object.STRING_OBJ {
						errors := getErrorTokenTraceAsJson(e).([]string)
						errors = append(errors, fmt.Sprintf("query_params must be LIST of STRINGs. found=%s", elem.Type()))
						return false, errors
					}
					// Now we know its a list of strings so we can set the variables accordingly for the fn
					s := elem.(*object.Stringo).Value
					fn.Env.Set(s, &object.Stringo{Value: c.Query(s)})
				}
			} else if isCookies {
				// Handle cookies
				if v.Type() != object.LIST_OBJ {
					errors := getErrorTokenTraceAsJson(e).([]string)
					errors = append(errors, fmt.Sprintf("cookies must be LIST. got=%s", v.Type()))
					return false, errors
				}
				l := v.(*object.List).Elements
				for _, elem := range l {
					if elem.Type() != object.STRING_OBJ {
						errors := getErrorTokenTraceAsJson(e).([]string)
						errors = append(errors, fmt.Sprintf("cookies must be LIST of STRINGs. found=%s", elem.Type()))
						return false, errors
					}
					// Now we know its a list of strings so we can set the variables accordingly for the fn
					s := elem.(*object.Stringo).Value
					fn.Env.Set(s, &object.Stringo{Value: c.Cookies(s)})
				}
			} else if fn.Parameters[k].Value == varName {
				// Handle post_values, put_values, patch_values (in body)
				if v.Type() != object.LIST_OBJ {
					errors := getErrorTokenTraceAsJson(e).([]string)
					errors = append(errors, fmt.Sprintf("%s must be LIST. got=%s", varName, v.Type()))
					return false, errors
				}
				l := v.(*object.List).Elements

				contentType := c.Get("Content-Type")
				body := strings.NewReader(string(c.Body()))

				returnMap, err := decodeBodyToMap(contentType, body)
				if err != nil {
					errors := getErrorTokenTraceAsJson(e).([]string)
					errors = append(errors, fmt.Sprintf("received input that could not be decoded in `%s`", string(c.Body())))
					return false, errors
				}
				for _, elem := range l {
					if elem.Type() != object.STRING_OBJ {
						errors := getErrorTokenTraceAsJson(e).([]string)
						errors = append(errors, fmt.Sprintf("%s must be LIST of STRINGs. found=%s", varName, elem.Type()))
						return false, errors
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
		}
	}
	return true, []string{}
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

func getAndSetHttpParams(e *Evaluator, fn *object.Function, c *fiber.Ctx) ([]object.Object, []bool) {
	fnArgs := make([]object.Object, len(fn.Parameters))
	immutableArgs := make([]bool, len(fnArgs))
	for i, v := range fn.Parameters {
		if v != nil {
			if v.Value == "headers" {
				// Handle headers
				fnArgs[i] = getReqHeaderMapObj(c)
			} else if v.Value == "request" {
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
				mapObj.Set("is_from_local", &object.Boolean{Value: c.IsFromLocal()})
				mapObj.Set("is_secure", &object.Boolean{Value: c.Secure()})
				fnArgs[i] = object.CreateMapObjectForGoMap(*mapObj)
			} else {
				fnArgs[i] = &object.Stringo{Value: c.Params(v.Value)}
			}
			immutableArgs[i] = true
		}
	}
	return fnArgs, immutableArgs
}

func createHttpHandleWSBuiltin(e *Evaluator) *object.Builtin {
	return &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("handle_ws", len(args), 3, "")
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newPositionalTypeError("handle_ws", 1, object.UINTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("handle_ws", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.FUNCTION_OBJ {
				return newPositionalTypeError("handle_ws", 3, object.FUNCTION_OBJ, args[2].Type())
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
			app.Use(pattern, func(c *fiber.Ctx) error {
				if websocket.IsWebSocketUpgrade(c) {
					return c.Next()
				}
				return fiber.ErrUpgradeRequired
			})

			var returnObj object.Object = NULL
			wsHandler := websocket.New(func(c *websocket.Conn) {
				connCount := wsConnCount.Add(1)
				WSConnMap.Put(connCount, c)
				for k, v := range fn.DefaultParameters {
					isQueryParams := v != nil && fn.Parameters[k].Value == "query_params"
					isCookies := v != nil && fn.Parameters[k].Value == "cookies"
					if v != nil {
						if isQueryParams {
							// Handle query_params
							if v.Type() != object.LIST_OBJ {
								errors := getErrorTokenTraceAsJson(e).([]string)
								log.Printf("%s\nquery_params must be LIST. got=%s", strings.Join(errors, "\n"), v.Type())
								return
							}
							l := v.(*object.List).Elements
							for _, elem := range l {
								if elem.Type() != object.STRING_OBJ {
									errors := getErrorTokenTraceAsJson(e).([]string)
									log.Printf("%s\nquery_params must be LIST of STRINGs. found=%s", strings.Join(errors, "\n"), elem.Type())
									return
								}
								// Now we know its a list of strings so we can set the variables accordingly for the fn
								s := elem.(*object.Stringo).Value
								fn.Env.Set(s, &object.Stringo{Value: c.Query(s)})
							}
						} else if isCookies {
							// Handle cookies
							if v.Type() != object.LIST_OBJ {
								errors := getErrorTokenTraceAsJson(e).([]string)
								log.Printf("%s\ncookies must be LIST. got=%s", strings.Join(errors, "\n"), v.Type())
								return
							}
							l := v.(*object.List).Elements
							for _, elem := range l {
								if elem.Type() != object.STRING_OBJ {
									errors := getErrorTokenTraceAsJson(e).([]string)
									log.Printf("%s\ncookies must be LIST of STRINGs. found=%s", strings.Join(errors, "\n"), elem.Type())
									return
								}
								// Now we know its a list of strings so we can set the variables accordingly for the fn
								s := elem.(*object.Stringo).Value
								fn.Env.Set(s, &object.Stringo{Value: c.Cookies(s)})
							}
						}
					}
				}
				fnArgs := make([]object.Object, len(fn.Parameters))
				immutableArgs := make([]bool, len(fnArgs))
				for i, v := range fn.Parameters {
					if i == 0 {
						fnArgs[i] = object.CreateBasicMapObject("ws", connCount)
					} else {
						fnArgs[i] = &object.Stringo{Value: c.Params(v.Value)}
					}
					immutableArgs[i] = true
				}
				returnObj = e.applyFunction(fn, fnArgs, make(map[string]object.Object), immutableArgs)
				if isError(returnObj) {
					var buf bytes.Buffer
					buf.WriteString(returnObj.(*object.Error).Message)
					buf.WriteByte('\n')
					for e.ErrorTokens.Len() > 0 {
						tok := e.ErrorTokens.PopBack()
						buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					}
					fmt.Printf("%s`handle_ws` return error: %s\n", consts.EVAL_ERROR_PREFIX, buf.String())
				} else {
					if returnObj == NULL {
						// Dont need to log if its null - probably no error then
						return
					}
					log.Printf("`handle_ws` returned with %#v", returnObj)
				}
			})
			app.Get(pattern, wsHandler)

			// Always returns NULL here
			return returnObj
		},
	}
}
