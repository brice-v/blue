package vm

import (
	"blue/consts"
	"blue/object"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	ws "github.com/gorilla/websocket"
	"golang.org/x/net/html"
)

func GetStdBuiltinWithVm(mod, name string, vm *VM) func(args ...object.Object) object.Object {
	switch mod {
	case "http":
		switch name {
		case "_handle":
			return createHttpHandleBuiltin(vm, false).Fun
		case "_handle_use":
			return createHttpHandleBuiltin(vm, true).Fun
		case "_handle_ws":
			// The template snapshot is created once, on the first resolution,
			// while the source vm is still young and safe to deep clone.
			// Later resolutions (inlined `from http import` statements
			// re-execute on every ws message, including on connection vms)
			// must reuse it: cloning here would copy a live vm and retain
			// megabytes per message.
			if vm.wsTemplate == nil {
				vm.wsTemplate = vm.Clone(vm.PID)
			}
			return createHttpHandleWSBuiltin(vm.wsTemplate).Fun
		default:
			panic("GetStdBuiltinWithVm called with incorrect builtin function name '" + name + "' for module: " + mod)
		}
	case "ui":
		switch name {
		case "_button":
			return createUIButtonBuiltin(vm).Fun
		case "_check_box":
			return createUICheckBoxBuiltin(vm).Fun
		case "_radio_group":
			return createUIRadioBuiltin(vm).Fun
		case "_option_select":
			return createUIOptionSelectBuiltin(vm).Fun
		case "_form":
			return createUIFormBuiltin(vm).Fun
		case "_toolbar_action":
			return createUIToolbarAction(vm).Fun
		default:
			panic("GetStdBuiltinWithVm called with incorrect builtin function name '" + name + "' for module: " + mod)
		}
	}
	panic("GetStdBuiltinWithVm called with incorrect module: " + mod)
}

func createHttpHandleBuiltin(vm *VM, isHandleUse bool) *object.Builtin {
	return &object.Builtin{
		Name: "handle",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("handle", len(args), 4, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("handle", 1, object.GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*object.GoObj[*object.Server])
			if !ok {
				return newPositionalTypeErrorForGoObj("handle", 1, "*Server", args[0])
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("handle", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.CLOSURE {
				return newPositionalTypeError("handle", 3, object.CLOSURE, args[2].Type())
			}
			if args[3].Type() != object.STRING_OBJ {
				return newPositionalTypeError("handle", 4, object.STRING_OBJ, args[3].Type())
			}
			method := strings.ToUpper(args[3].(*object.Stringo).Value)
			pattern := args[1].(*object.Stringo).Value
			fun := args[2].(*object.Closure)
			handler := func(c *object.Ctx) {
				_ = processHandlerFn(vm, fun, c, method)
			}
			if isHandleUse {
				if method != "" {
					return newError("`handle_use` error: method should be '', got=%s", method)
				}
				app.Value.Add("", pattern, handler, true)
			} else {
				switch method {
				case "GET":
					app.Value.Add("GET", pattern, handler, false)
				case "POST":
					app.Value.Add("POST", pattern, handler, false)
				case "PATCH":
					app.Value.Add("PATCH", pattern, handler, false)
				case "PUT":
					app.Value.Add("PUT", pattern, handler, false)
				case "DELETE":
					app.Value.Add("DELETE", pattern, handler, false)
				}
			}
			return object.NULL
		},
	}
}

func processHandlerFn(vm *VM, fn *object.Closure, c *object.Ctx, method string) error {
	ok, respObj, errors := prepareAndApplyHttpHandleFn(vm, fn, c, method)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(errors)
	}
	// First check if the respObj is a MAP and if its a valid http handler response action
	if respObj.Type() == object.MAP_OBJ {
		isAction, action, m := tryGetHttpActionAndMap(respObj)
		if isAction {
			switch action {
			case "status":
				maybeCode, ok := m.Get("code")
				if !ok {
					err := "http/status 'code' key not found."
					return c.Status(http.StatusInternalServerError).JSON(err)
				}
				code, ok := maybeCode.(int64)
				if !ok {
					err := fmt.Sprintf("http/status 'code' must be INTEGER. got=%T", maybeCode)
					return c.Status(http.StatusInternalServerError).JSON(err)
				}
				return c.SendStatus(int(code))
			case "redirect":
				maybeLocation, ok := m.Get("location")
				if !ok {
					err := "http/redirect 'location' key not found."
					return c.Status(http.StatusInternalServerError).JSON(err)
				}
				location, ok := maybeLocation.(string)
				if !ok {
					err := fmt.Sprintf("http/redirect 'location' must be STRING. got=%T", maybeLocation)
					return c.Status(http.StatusInternalServerError).JSON(err)
				}
				maybeCode, ok := m.Get("code")
				if !ok {
					err := "http/redirect 'code' key not found."
					return c.Status(http.StatusInternalServerError).JSON(err)
				}
				code, ok := maybeCode.(int64)
				if !ok {
					err := fmt.Sprintf("http/redirect 'code' must be INTEGER. got=%T", maybeCode)
					return c.Status(http.StatusInternalServerError).JSON(err)
				}
				return c.Redirect(location, int(code))
			case "next":
				return c.Next()
			case "send_file":
				maybePath, ok := m.Get("path")
				if !ok {
					err := "http/send_file 'path' key not found."
					return c.Status(http.StatusInternalServerError).JSON(err)
				}
				path, ok := maybePath.(string)
				if !ok {
					err := fmt.Sprintf("http/send_file 'path' must be STRING. got=%T", maybePath)
					return c.Status(http.StatusInternalServerError).JSON(err)
				}
				return c.SendFile(path, false)
			}
		}
	}
	if method != "GET" {
		if respObj.Type() == object.STRING_OBJ {
			return c.SendString(respObj.(*object.Stringo).Value)
		}
		if respObj.Type() == object.NULL_OBJ {
			return c.SendStatus(http.StatusOK)
		} else {
			obj := blueObjToJsonObject(respObj)
			if isError(obj) {
				errors := getErrorTokenTraceAsJsonWithError(vm, obj.(*object.Error).Message).([]string)
				errors = append(errors, fmt.Sprintf("%s Response Type is not STRING, valid JSON, or NULL. got=%s", method, obj.Type()))
				return c.Status(http.StatusInternalServerError).JSON(errors)
			} else {
				if respStr, ok := obj.(*object.Stringo); ok {
					respStrBs := []byte(respStr.Value)
					if json.Valid(respStrBs) {
						c.Set("Content-Type", "application/json")
						return c.Send(respStrBs)
					}
				}
			}
			errors := getErrorTokenTraceAsJson(vm).([]string)
			errors = append(errors, fmt.Sprintf("%s Response Type is not NULL or STRING. got=%s", method, respObj.Type()))
			return c.Status(http.StatusInternalServerError).JSON(errors)
		}
	} else {
		if respObj.Type() == object.STRING_OBJ {
			respStr := respObj.(*object.Stringo).Value
			respStrBs := []byte(respStr)
			if json.Valid(respStrBs) {
				c.Set("Content-Type", "application/json")
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
			if isError(obj) {
				errors := getErrorTokenTraceAsJsonWithError(vm, obj.(*object.Error).Message).([]string)
				errors = append(errors, "error converting object to JSON")
				return c.Status(http.StatusInternalServerError).JSON(errors)
			}
			if respStr, ok := obj.(*object.Stringo); ok {
				respStrBs := []byte(respStr.Value)
				if json.Valid(respStrBs) {
					c.Set("Content-Type", "application/json")
					return c.Send(respStrBs)
				}
			}
			errors := getErrorTokenTraceAsJson(vm).([]string)
			errors = append(errors, "STRING NOT RETURNED FROM JSON CONVERSION")
			return c.Status(http.StatusInternalServerError).JSON(errors)
		}
	}
}

func createHttpHandleWSBuiltin(vm *VM) *object.Builtin {
	var disableHttpServerDebug bool
	disableHttpServerDebugStr := os.Getenv(consts.BLUE_DISABLE_HTTP_SERVER_DEBUG)
	disableHttpServerDebug, err := strconv.ParseBool(disableHttpServerDebugStr)
	if err != nil {
		disableHttpServerDebug = false
	}
	return &object.Builtin{
		Name: "handle_ws",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("handle_ws", len(args), 3, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("handle_ws", 1, object.GO_OBJ, args[0].Type())
			}
			app, ok := args[0].(*object.GoObj[*object.Server])
			if !ok {
				return newPositionalTypeErrorForGoObj("handle_ws", 1, "*Server", args[0])
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("handle_ws", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.CLOSURE {
				return newPositionalTypeError("handle_ws", 3, object.CLOSURE, args[2].Type())
			}
			pattern := args[1].(*object.Stringo).Value
			fn := args[2].(*object.Closure)
			if len(fn.Fun.Parameters) == 0 {
				return newError("function arguments should be at least 1 to store the websocket connection")
			}
			upgrader := ws.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}
			app.Value.Add("GET", pattern, func(c *object.Ctx) {
				if !isWebSocketUpgradeRequest(c.R) {
					http.Error(c.W, "websocket: upgrade required", http.StatusUpgradeRequired)
					return
				}
				conn, err := upgrader.Upgrade(c.W, c.R, nil)
				if err != nil {
					return
				}
				defer func() {
					_ = conn.Close()
				}()
				// Each connection runs blue code for its whole lifetime, so
				// it gets its own vm and its own copy of the handler closure.
				// Sharing one vm (or the closure's special parameter maps)
				// between concurrent connections corrupts vm state. The
				// connection vm shares immutable program data with the
				// registration-time snapshot instead of deep cloning it, and
				// is made from that snapshot, never from the live vm.
				connVm := vm.CloneForConnection(vm.PID)
				connFn := cloneHandlerClosure(fn)
				handleSpecialFunctionArgs(connFn, c.R)
				fnArgs := make([]object.Object, len(connFn.Fun.Parameters))
				for i, v := range connFn.Fun.Parameters {
					if i == 0 {
						fnArgs[i] = object.CreateBasicMapObjectForGoObj("ws", NewGoObj(conn))
					} else {
						fnArgs[i] = &object.Stringo{Value: c.Params(v)}
					}
				}
				returnObj := connVm.applyFunctionFastWithMultipleArgs(connFn, fnArgs)
				if isError(returnObj) {
					// var buf bytes.Buffer
					// buf.WriteString(returnObj.(*object.Error).Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					if !disableHttpServerDebug {
						fmt.Printf("%s`handle_ws` error: %s\n", consts.VM_ERROR_PREFIX, returnObj.(*object.Error).Message)
					}
				} else {
					if returnObj == object.NULL {
						// Dont need to log if its null - probably no error then
						return
					}
					if !disableHttpServerDebug {
						fmt.Printf("%s`handle_ws` returned with %#+v\n", consts.VM_ERROR_PREFIX, returnObj)
					}
				}
			}, false)

			// Always returns NULL here
			return object.NULL
		},
	}
}

// isWebSocketUpgradeRequest reports whether the request asks for a websocket
// upgrade, mirroring the old fiber IsWebSocketUpgrade check.
func isWebSocketUpgradeRequest(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	if !headerContainsToken(r.Header, "Connection", "upgrade") {
		return false
	}
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func headerContainsToken(header http.Header, key, token string) bool {
	for _, line := range header.Values(key) {
		for _, part := range strings.Split(line, ",") {
			if strings.EqualFold(strings.TrimSpace(part), token) {
				return true
			}
		}
	}
	return false
}

func handleSpecialFunctionArgs(fn *object.Closure, r *http.Request) {
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
					objectMap[k] = &object.Stringo{Value: r.URL.Query().Get(k.Name)}
				}
			}
		case "cookies":
			if objectMap, ok := fn.Fun.SpecialFunctionParameters[key]; ok {
				for k := range objectMap {
					c, err := r.Cookie(k.Name)
					value := ""
					if err == nil {
						value = c.Value
					}
					objectMap[k] = &object.Stringo{Value: value}
				}
			}
		}
	}
}

// cloneHandlerClosure returns a copy of fn whose compiled function carries a
// private copy of SpecialFunctionParameters so concurrent handlers never write
// to the same default parameter maps. Bytecode is immutable at runtime and
// shared safely; free variables are shared as well.
func cloneHandlerClosure(fn *object.Closure) *object.Closure {
	var sfp map[object.NameIndexKey]map[object.NameIndexKey]object.Object
	if fn.Fun.SpecialFunctionParameters != nil {
		sfp = make(map[object.NameIndexKey]map[object.NameIndexKey]object.Object, len(fn.Fun.SpecialFunctionParameters))
		for k, v := range fn.Fun.SpecialFunctionParameters {
			inner := make(map[object.NameIndexKey]object.Object, len(v))
			for kk, vv := range v {
				inner[kk] = vv
			}
			sfp[k] = inner
		}
	}
	fun := &object.CompiledFunction{
		Instructions:              fn.Fun.Instructions,
		NumLocals:                 fn.Fun.NumLocals,
		NumParameters:             fn.Fun.NumParameters,
		Parameters:                fn.Fun.Parameters,
		ParameterHasDefault:       fn.Fun.ParameterHasDefault,
		NumDefaultParams:          fn.Fun.NumDefaultParams,
		DisplayString:             fn.Fun.DisplayString,
		SpecialFunctionParameters: sfp,
		HelpStr:                   fn.Fun.HelpStr,
	}
	return &object.Closure{Fun: fun, Free: fn.Free}
}

func createUIButtonBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "button",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("button", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("button", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("button", 2, object.CLOSURE, args[1].Type())
			}
			s := args[0].(*object.Stringo).Value
			fn := args[1].(*object.Closure)
			button := widget.NewButton(s, func() {
				obj := vm.applyFunctionFast(fn, nil)
				if isError(obj) {
					err := obj.(*object.Error)
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`button` click handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](button)
		},
	}
}

func createUICheckBoxBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "checkbox",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("checkbox", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("checkbox", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("checkbox", 2, object.CLOSURE, args[1].Type())
			}
			lbl := args[0].(*object.Stringo).Value
			fn := args[1].(*object.Closure)
			if len(fn.Fun.Parameters) != 1 {
				return newError("`checkbox` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			checkBox := widget.NewCheck(lbl, func(value bool) {
				obj := vm.applyFunctionFast(fn, nativeToBooleanObject(value))
				if isError(obj) {
					err := obj.(*object.Error)
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`check_box` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](checkBox)
		},
	}
}

func createUIRadioBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "radio_group",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("radio_group", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("radio_group", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("radio_group", 2, object.CLOSURE, args[1].Type())
			}
			elems := args[0].(*object.List).Elements
			fn := args[1].(*object.Closure)
			options := make([]string, len(elems))
			for i, e := range elems {
				if e.Type() != object.STRING_OBJ {
					return newError("`radio_group` error: all elements in list should be STRING. found=%s", e.Type())
				}
				options[i] = e.(*object.Stringo).Value
			}
			if len(fn.Fun.Parameters) != 1 {
				return newError("`radio_group` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			radio := widget.NewRadioGroup(options, func(value string) {
				obj := vm.applyFunctionFast(fn, &object.Stringo{Value: value})
				if isError(obj) {
					err := obj.(*object.Error)
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`radio_group` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](radio)
		},
	}
}

func createUIOptionSelectBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "option_select",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("option_select", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("option_select", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("option_select", 2, object.CLOSURE, args[1].Type())
			}
			elems := args[0].(*object.List).Elements
			fn := args[1].(*object.Closure)
			options := make([]string, len(elems))
			for i, e := range elems {
				if e.Type() != object.STRING_OBJ {
					return newError("`option_select` error: all elements in list should be STRING. found=%s", e.Type())
				}
				options[i] = e.(*object.Stringo).Value
			}
			if len(fn.Fun.Parameters) != 1 {
				return newError("`option_select` error: handler needs 1 argument. got=%d", len(fn.Fun.Parameters))
			}
			option := widget.NewSelect(options, func(value string) {
				obj := vm.applyFunctionFast(fn, &object.Stringo{Value: value})
				if isError(obj) {
					err := obj.(*object.Error)
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`option_select` handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			})
			return NewGoObj[fyne.CanvasObject](option)
		},
	}
}

func createUIFormBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "form",
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
			if args[2].Type() != object.CLOSURE {
				return newPositionalTypeError("form", 3, object.CLOSURE, args[2].Type())
			}
			var formItems []*widget.FormItem
			labels := args[0].(*object.List).Elements
			widgetIds := args[1].(*object.List).Elements
			if len(labels) != len(widgetIds) {
				return newError("`form` error: labels and widget ids must be the same length. len(labels)=%d, len(widgetIds)=%d", len(labels), len(widgetIds))
			}
			fn := args[2].(*object.Closure)
			for i := range labels {
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
					obj := vm.applyFunctionFast(fn, nil)
					if isError(obj) {
						err := obj.(*object.Error)
						// var buf bytes.Buffer
						// buf.WriteString(err.Message)
						// buf.WriteByte('\n')
						// for e.ErrorTokens.Len() > 0 {
						// 	tok := e.ErrorTokens.PopBack()
						// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
						// }
						fmt.Printf("%s`form` on_submit error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
					}
				},
			}
			return NewGoObj[fyne.CanvasObject](form)
		},
	}
}

func createUIToolbarAction(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "toolbar_action",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("toolbar_action", len(args), 2, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("toolbar_action", 1, object.GO_OBJ, args[0].Type())
			}
			if args[1].Type() != object.CLOSURE {
				return newPositionalTypeError("toolbar_action", 2, object.CLOSURE, args[1].Type())
			}
			r, ok := args[0].(*object.GoObj[fyne.Resource])
			if !ok {
				return newPositionalTypeErrorForGoObj("toolbar_action", 1, "fyne.Resource", args[0])
			}
			fn := args[1].(*object.Closure)
			return NewGoObj[widget.ToolbarItem](widget.NewToolbarAction(r.Value, func() {
				obj := vm.applyFunctionFast(fn, nil)
				if isError(obj) {
					err := obj.(*object.Error)
					// var buf bytes.Buffer
					// buf.WriteString(err.Message)
					// buf.WriteByte('\n')
					// for e.ErrorTokens.Len() > 0 {
					// 	tok := e.ErrorTokens.PopBack()
					// 	buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
					// }
					fmt.Printf("%s`toolbar_action` click handler error: %s\n", consts.VM_ERROR_PREFIX, err.Message)
				}
			}))
		},
	}
}
