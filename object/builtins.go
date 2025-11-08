package object

import (
	"blue/consts"
	"blue/util"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"runtime/metrics"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gobuffalo/plush"
	"github.com/google/uuid"
	"github.com/gookit/color"
	"github.com/huandu/go-clone"
	"github.com/huandu/xstrings"
	"github.com/shopspring/decimal"
)

type BuiltinType string

type NewBuiltinSliceType []struct {
	Name    string
	Builtin *Builtin
}

const (
	BuiltinBaseType   BuiltinType = "BUILTIN"
	BuiltinHttpType   BuiltinType = "HTTP"
	BuiltinTimeType   BuiltinType = "TIME"
	BuiltinSearchType BuiltinType = "SEARCH"
	BuiltinDbType     BuiltinType = "DB"
	BuiltinMathType   BuiltinType = "MATH"
	BuiltinConfigType BuiltinType = "CONFIG"
	BuiltinCryptoType BuiltinType = "CRYPTO"
	BuiltinNetType    BuiltinType = "NET"
	BuiltinColorType  BuiltinType = "COLOR"
	BuiltinCsvType    BuiltinType = "CSV"
	BuiltinPsutilType BuiltinType = "PSUTIL"
	BuiltinWasmType   BuiltinType = "WASM"
	BuiltinUiType     BuiltinType = "UI"
	BuiltinGgType     BuiltinType = "GG"
)

type BuiltinMapType struct {
	*ConcurrentMap[string, *Builtin]
}

func NewBuiltinObjMap(input map[string]*Builtin) BuiltinMapType {
	return BuiltinMapType{&ConcurrentMap[string, *Builtin]{
		Kv: input,
	}}
}

type BuiltinMapTypeInternal map[string]*Builtin

var Builtins = NewBuiltinSliceType{
	{
		Name: "_get_",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 || len(args) > 3 {
					return newInvalidArgCountError("_get_", len(args), 2, "or 3")
				}
				arg0Type := args[0].Type()
				if arg0Type != STRING_OBJ && arg0Type != LIST_OBJ && arg0Type != SET_OBJ && arg0Type != MAP_OBJ {
					return newPositionalTypeError("_get_", 1, "STR or MAP or LIST or SET", arg0Type)
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("_get_", 2, INTEGER_OBJ, args[1].Type())
				}
				iterableIndex := args[1].(*Integer).Value
				withIndex := false
				if len(args) == 3 {
					if args[2].Type() != BOOLEAN_OBJ {
						return newPositionalTypeError("_get_", 3, BOOLEAN_OBJ, args[2].Type())
					}
					withIndex = args[2].(*Boolean).Value
				}
				switch arg0Type {
				case STRING_OBJ:
					s := args[0].(*Stringo).Value
					if iterableIndex > int64(utf8.RuneCountInString(s)) || iterableIndex < 0 {
						return newError("`_get_` index %d out of bounds", iterableIndex)
					}
					rs := []rune(s)
					indexed := &Stringo{Value: string(rs[iterableIndex])}
					if withIndex {
						return &List{Elements: []Object{args[1], indexed}}
					} else {
						return indexed
					}
				case LIST_OBJ:
					l := args[0].(*List).Elements
					if iterableIndex > int64(len(l)) || iterableIndex < 0 {
						return newError("`_get_` index %d out of bounds", iterableIndex)
					}
					indexed := l[iterableIndex]
					if withIndex {
						return &List{Elements: []Object{args[1], indexed}}
					} else {
						return indexed
					}
				case SET_OBJ:
					s := args[0].(*Set).Elements
					if iterableIndex > int64(s.Len()) || iterableIndex < 0 {
						return newError("`_get_` index %d out of bounds", iterableIndex)
					}
					indexedKey := s.Keys[iterableIndex]
					indexed, _ := s.Get(indexedKey)
					if withIndex {
						return &List{Elements: []Object{args[1], indexed.Value}}
					} else {
						return indexed.Value
					}
				case MAP_OBJ:
					m := args[0].(*Map).Pairs
					if iterableIndex > int64(m.Len()) || iterableIndex < 0 {
						return newError("`_get_` index %d out of bounds", iterableIndex)
					}
					indexedKey := m.Keys[iterableIndex]
					indexed, _ := m.Get(indexedKey)
					if withIndex {
						return &List{Elements: []Object{indexed.Key, indexed.Value}}
					} else {
						return indexed.Value
					}
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`_get_` returns the item at the index specified for the given iterable",
				signature:   "_get_(arg: list|str|map|set, index: int, with_index: bool=false) -> any",
				errors:      "None",
				example:     "_get([1,2,3], 2) => 3",
			}.String(),
		},
	},
	{
		Name: "println",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				useColorPrinter := false
				var style color.Style
				for i, arg := range args {
					if i == 0 {
						t, s, ok := getBasicObjectForGoObj[color.Style](arg)
						if ok && t == "color" {
							// Use color printer
							useColorPrinter = true
							style = s
							continue
						} else {
							useColorPrinter = false
						}
					}
					if useColorPrinter {
						style.Println(arg.Inspect())
					} else {
						fmt.Println(arg.Inspect())
					}
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`println` returns NULL and prints each of args on a new line",
				signature:   "println(args... : any...) -> null",
				errors:      "None",
				example:     "println(1,2,3) => null\n    STDOUT: 1\\n2\\n3\\n",
			}.String(),
		},
	},
	{
		Name: "print",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				useColorPrinter := false
				var style color.Style
				for i, arg := range args {
					if i == 0 {
						t, s, ok := getBasicObjectForGoObj[color.Style](arg)
						if ok && t == "color" {
							// Use color printer
							useColorPrinter = true
							style = s
							continue
						} else {
							useColorPrinter = false
						}
					}
					if useColorPrinter {
						style.Print(arg.Inspect())
					} else {
						fmt.Print(arg.Inspect())
					}
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`print` returns NULL and prints each of args",
				signature:   "print(args... : any...) -> null",
				errors:      "None",
				example:     "print(1,2,3) => null\n    STDOUT: 123",
			}.String(),
		},
	},
	{
		Name: "help",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("help", len(args), 1, "")
				}
				return &Stringo{Value: args[0].Help()}
			},
			HelpStr: helpStrArgs{
				explanation: "`help` returns the help STRING for a given OBJECT",
				signature:   "help(arg: any) -> str",
				errors:      "None",
				example:     "help(print) => `prints help string for print builtin function`",
			}.String(),
		},
	},
	{
		Name: "new",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("new", len(args), 1, "")
				}
				if !IsCollectionType(args[0].Type()) {
					return newPositionalTypeError("new", 1, "MAP or LIST or SET", args[0].Type())
				}
				if args[0].Type() == MAP_OBJ {
					m := args[0].(*Map)
					return clone.Clone(m).(*Map)
				} else if args[0].Type() == LIST_OBJ {
					l := args[0].(*List)
					return clone.Clone(l).(*List)
				} else {
					s := args[0].(*Set)
					return clone.Clone(s).(*Set)
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`new` returns a cloned MAP object from the given arg",
				signature:   "new(arg: map|list|set) -> map|list|set",
				errors:      "InvalidArgCount,PositionalType",
				example:     "new({'x': 1}) => {'x': 1}",
			}.String(),
		},
	},
	{
		Name: "keys",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("keys", len(args), 1, "")
				}
				if args[0].Type() != MAP_OBJ {
					return newPositionalTypeError("keys", 1, MAP_OBJ, args[0].Type())
				}
				returnList := &List{
					Elements: []Object{},
				}
				m := args[0].(*Map)
				for _, k := range m.Pairs.Keys {
					mp, _ := m.Pairs.Get(k)
					returnList.Elements = append(returnList.Elements, mp.Key)
				}
				return returnList
			},
			HelpStr: helpStrArgs{
				explanation: "`keys` returns a LIST of key objects from given arg",
				signature:   "keys(arg: map) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "keys({'a': 1, 'B': 2}) => ['a', 'B']",
			}.String(),
		},
	},
	{
		Name: "values",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("values", len(args), 1, "")
				}
				if args[0].Type() != MAP_OBJ {
					return newPositionalTypeError("values", 1, MAP_OBJ, args[0].Type())
				}
				returnList := &List{
					Elements: []Object{},
				}
				m := args[0].(*Map)
				for _, k := range m.Pairs.Keys {
					mp, _ := m.Pairs.Get(k)
					returnList.Elements = append(returnList.Elements, mp.Value)
				}
				return returnList
			},
			HelpStr: helpStrArgs{
				explanation: "`values` returns a LIST of value OBJECTs from given arg",
				signature:   "value(arg: map) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "keys({'a': 1, 'B': 2}) => [1, 2]",
			}.String(),
		},
	},
	{
		Name: "del",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("del", len(args), 2, "")
				}
				if !IsCollectionType(args[0].Type()) {
					return newPositionalTypeError("del", 1, "MAP or LIST or SET", args[0].Type())
				}
				if args[0].Type() == MAP_OBJ {
					m := args[0].(*Map)
					hk := HashKey{
						Type:  args[1].Type(),
						Value: HashObject(args[1]),
					}
					m.Pairs.Delete(hk)
				} else if args[1].Type() == LIST_OBJ {
					if args[1].Type() != INTEGER_OBJ {
						return newPositionalTypeError("del", 2, LIST_OBJ, args[1].Type())
					}
					l := args[0].(*List)
					index := args[1].(*Integer).Value
					l.Elements = append(l.Elements[:index], l.Elements[index+1:]...)
				} else {
					hk := HashKey{
						Type:  args[1].Type(),
						Value: HashObject(args[1]),
					}
					s := args[0].(*Set)
					s.Elements.Delete(hk.Value)
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`del` deletes a key from a MAP or index from a LIST",
				signature:   "del(m: map|list, key: any|int) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "del({'a': 1, 'B': 2}, 'a') => null - side effect: {'B': 2}",
			}.String(),
		},
	},
	{
		Name: "len",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("len", len(args), 1, "")
				}

				switch arg := args[0].(type) {
				case *Stringo:
					return &Integer{Value: int64(utf8.RuneCountInString(arg.Value))}
				case *List:
					return &Integer{Value: int64(len(arg.Elements))}
				case *Map:
					return &Integer{Value: int64(arg.Pairs.Len())}
				case *Set:
					return &Integer{Value: int64(arg.Elements.Len())}
				case *Bytes:
					return &Integer{Value: int64(len(arg.Value))}
				default:
					return newPositionalTypeError("len", 1, "STRING, LIST, MAP, SET, or BYTES", args[0].Type())
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`len` returns the INTEGER length of the given STRING, LIST, MAP, or SET",
				signature:   "len(arg: str|list|map|set) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "len([1,2,3]) => 3",
			}.String(),
		},
	},
	{
		Name: "append",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("append", len(args), 2, " or more")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("append", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				length := len(l.Elements)
				args = args[1:]
				argsLength := len(args)
				newElements := make([]Object, length+argsLength)
				copy(newElements, l.Elements)
				copy(newElements[length:length+argsLength], args)
				return &List{Elements: newElements}
			},
			HelpStr: helpStrArgs{
				explanation: "`append` returns the LIST of elements with given arg OBJECT at the end",
				signature:   "append(arg0: list, args...: any) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "append([1,2,3], 1) => [1,2,3,4]",
			}.String(),
		},
	},
	{
		Name: "prepend",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("prepend", len(args), 2, "")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("prepend", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				length := len(l.Elements)
				args = args[1:]
				argsLength := len(args)
				newElements := make([]Object, length+argsLength)
				copy(newElements, args)
				copy(newElements[argsLength:argsLength+length], l.Elements)
				return &List{Elements: newElements}
			},
			HelpStr: helpStrArgs{
				explanation: "`prepend` returns the LIST of elements with given arg OBJECT at the front",
				signature:   "prepend(arg0: list, args...: any) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "prepend([1,2,3], 4) => [4,1,2,3]",
			}.String(),
		},
	},
	{
		Name: "push",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("push", len(args), 2, " or more")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("push", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				args = args[1:]
				l.Elements = append(l.Elements, args...)
				return &Integer{Value: int64(len(l.Elements))}
			},
			HelpStr: helpStrArgs{
				explanation: "`push` puts the given args at the end of the LIST and mutates it. The value returned is the length after pushing",
				signature:   "push(arg0: list[any], args...: any) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "push([1,2,3], 1) => 4",
			}.String(),
			Mutates: true,
		},
	},
	{
		Name: "pop",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("pop", len(args), 1, "")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("pop", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				if len(l.Elements) == 0 {
					return NULL
				}
				elem := l.Elements[len(l.Elements)-1]
				l.Elements = l.Elements[:len(l.Elements)-1]
				return elem
			},
			HelpStr: helpStrArgs{
				explanation: "`pop` returns the last element of the LIST and mutates it",
				signature:   "pop(arg0: list[any]) -> any",
				errors:      "InvalidArgCount,PositionalType",
				example:     "pop([1,2,3]) => 3",
			}.String(),
			Mutates: true,
		},
	},
	{
		Name: "unshift",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("unshift", len(args), 2, " or more")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("unshift", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				length := len(l.Elements)
				args = args[1:]
				argsLength := len(args)
				elems := make([]Object, argsLength+length)
				copy(elems, args)
				copy(elems[argsLength:argsLength+length], l.Elements)
				l.Elements = elems
				return &Integer{Value: int64(len(l.Elements))}
			},
			HelpStr: helpStrArgs{
				explanation: "`unshift` prepends the LIST with the given arguments and mutates it. The new length is returned",
				signature:   "unshift(arg0: list[any], args...: any) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "unshift([1,2,3], 1) => 4",
			}.String(),
			Mutates: true,
		},
	},
	{
		Name: "shift",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("shift", len(args), 1, "")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("shift", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				if len(l.Elements) == 0 {
					return NULL
				}
				elem := l.Elements[0]
				if len(l.Elements) == 1 {
					l.Elements = []Object{}
				} else {
					l.Elements = l.Elements[1:]
				}
				return elem
			},
			HelpStr: helpStrArgs{
				explanation: "`shift` returns the first element of the LIST and mutates it",
				signature:   "shift(arg0: list[any]) -> any",
				errors:      "InvalidArgCount,PositionalType",
				example:     "shift([1,2,3]) => 1",
			}.String(),
			Mutates: true,
		},
	},
	{
		Name: "concat",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("concat", len(args), 2, " or more")
				}
				allElems := []Object{}
				for i, e := range args {
					if e.Type() != LIST_OBJ {
						return newPositionalTypeError("concat", i+1, LIST_OBJ, e.Type())
					}
					l := e.(*List)
					allElems = append(allElems, l.Elements...)
				}
				return &List{Elements: allElems}
			},
			HelpStr: helpStrArgs{
				explanation: "`concat` merges 2 or more LISTs together and returns the result",
				signature:   "concat(arg0: list[any], args...: list[any]) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "concat([1,2,3], [1]) => [1,2,3,1]",
			}.String(),
		},
	},
	{
		Name: "reverse",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("reverse", len(args), 1, "")
				}
				if args[0].Type() != LIST_OBJ && args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("reverse", 1, LIST_OBJ+" or STRING", args[0].Type())
				}
				if args[0].Type() == STRING_OBJ {
					s := args[0].(*Stringo).Value
					return &Stringo{Value: xstrings.Reverse(s)}
				}
				l := args[0].(*List)
				length := len(l.Elements)
				newl := make([]Object, length)
				for i := 0; i < len(l.Elements); i++ {
					newl[length-i-1] = l.Elements[i]
				}
				return &List{Elements: newl}
			},
			HelpStr: helpStrArgs{
				explanation: "`reverse` reverse a string or list",
				signature:   "reverse(arg: list[any]|str) -> list[any]|str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "reverse([1,2,3]) => [3,2,1]",
			}.String(),
		},
	},
	{
		Name: "input",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				argLen := len(args)
				switch argLen {
				case 0:
					// read input with no prompt
					scanner := bufio.NewScanner(os.Stdin)
					if ok := scanner.Scan(); ok {
						return &Stringo{Value: scanner.Text()}
					}
					if err := scanner.Err(); err != nil {
						return newError("`input` error reading standard input: %s", err.Error())
					}
				case 1:
					// read input with prompt
					if args[0].Type() != STRING_OBJ {
						return newPositionalTypeError("input", 1, STRING_OBJ, args[0].Type())
					}
					scanner := bufio.NewScanner(os.Stdin)
					fmt.Print(args[0].(*Stringo).Value)
					if ok := scanner.Scan(); ok {
						return &Stringo{Value: scanner.Text()}
					}
					if err := scanner.Err(); err != nil {
						return newError("`input` error reading stdin: %s", err.Error())
					}
				}
				return newInvalidArgCountError("input", len(args), 0, "or 1")
			},
			HelpStr: helpStrArgs{
				explanation: "`input` returns STRING input from STDIN and takes an optional prompt",
				signature:   "input(prompt='') -> str",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "input('enter a value') => STDOUT: enter a value STDIN: 1 => 1",
			}.String(),
		}},
	// Default headers seem to be host, user-agent, accept-encoding (not case sensitive for these check pictures)
	// deno also used accept: */* (not sure what that is)
	{
		Name: "_fetch",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 5 {
					return newInvalidArgCountError("fetch", len(args), 5, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("fetch", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != STRING_OBJ {
					return newPositionalTypeError("fetch", 2, STRING_OBJ, args[1].Type())
				}
				if args[2].Type() != MAP_OBJ {
					return newPositionalTypeError("fetch", 3, MAP_OBJ, args[2].Type())
				}
				if args[3].Type() != NULL_OBJ && args[3].Type() != STRING_OBJ {
					return newPositionalTypeError("fetch", 4, "NULL or STRING", args[3].Type())
				}
				if args[4].Type() != BOOLEAN_OBJ {
					return newPositionalTypeError("fetch", 5, BOOLEAN_OBJ, args[4].Type())
				}
				resource := args[0].(*Stringo).Value
				method := args[1].(*Stringo).Value
				headersMap := args[2].(*Map).Pairs
				isFullResp := args[4].(*Boolean).Value
				var body io.Reader
				if args[3].Type() == NULL_OBJ {
					body = nil
				} else {
					body = strings.NewReader(args[3].(*Stringo).Value)
				}
				request, err := http.NewRequest(method, resource, body)
				if err != nil {
					return newError("`fetch` error: %s", err.Error())
				}
				// Add User Agent always and then it can be overwritten
				request.Header.Add("user-agent", "blue/v"+consts.VERSION)
				for _, k := range headersMap.Keys {
					mp, _ := headersMap.Get(k)
					if key, ok := mp.Key.(*Stringo); ok {
						if val, ok := mp.Value.(*Stringo); ok {
							request.Header.Add(key.Value, val.Value)
						}
					}
				}
				resp, err := http.DefaultClient.Do(request)
				if err != nil {
					return newError("`fetch` error: %s", err.Error())
				}
				if isFullResp {
					return getBlueObjectFromResp(resp)
				}
				defer resp.Body.Close()
				respBody, err := io.ReadAll(resp.Body)
				if err != nil {
					return newError("`fetch` error: %s", err.Error())
				}
				return &Stringo{Value: string(respBody)}
			},
			HelpStr: helpStrArgs{
				explanation: "`fetch` returns the body or full response of a network request",
				signature:   "fetch(resource: str, method: str('POST'|'PUT'|'PATCH'|'GET'|'HEAD'|'DELETE')='GET', headers: map[str]str, body: null|str|bytes, full_resp: bool)",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "fetch('https://danluu.com',full_resp=false) => <html>...</html>",
			}.String(),
		},
	},
	{
		Name: "_read",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("read", len(args), 2, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("read", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != BOOLEAN_OBJ {
					return newPositionalTypeError("read", 2, BOOLEAN_OBJ, args[1].Type())
				}
				fnameo := args[0].(*Stringo)
				var bs []byte
				if IsEmbed {
					s := fnameo.Value
					if strings.HasPrefix(s, "./") {
						s = strings.TrimLeft(s, "./")
					}
					fileData, err := Files.ReadFile(consts.EMBED_FILES_PREFIX + s)
					if err != nil {
						// Fallback option for reading when in embedded context
						fileData, err := os.ReadFile(fnameo.Value)
						if err != nil {
							return newError("`read` error reading file `%s`: %s", fnameo.Value, err.Error())
						}
						bs = fileData
					} else {
						bs = fileData
					}
				} else {
					fileData, err := os.ReadFile(fnameo.Value)
					if err != nil {
						return newError("`read` error reading file `%s`: %s", fnameo.Value, err.Error())
					}
					bs = fileData
				}
				if args[1].(*Boolean).Value {
					return &Bytes{Value: bs}
				}
				return &Stringo{Value: string(bs)}
			},
			HelpStr: helpStrArgs{
				explanation: "`_read` returns a STRING if given a FILE and false, or BYTES if given a file and true",
				signature:   "_read(arg: str, to_bytes=false) -> (to_bytes) ? bytes : str",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "_read('/test.txt') => example file data",
			}.String(),
		},
	},
	{
		Name: "_write",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				argLen := len(args)
				if argLen != 2 {
					return newInvalidArgCountError("write", len(args), 2, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("write", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != STRING_OBJ && args[1].Type() != BYTES_OBJ {
					return newPositionalTypeError("write", 2, "STRING or BYTES", args[1].Type())
				}
				fname := args[0].(*Stringo).Value
				var contents []byte
				if args[1].Type() == STRING_OBJ {
					contents = []byte(args[1].(*Stringo).Value)
				} else {
					contents = args[1].(*Bytes).Value
				}
				err := os.WriteFile(fname, contents, 0644)
				if err != nil {
					return newError("error writing file `%s`: %s", fname, err.Error())
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`_write` writes a STRING or BYTES to a given FILE",
				signature:   "_write(filename: str, contents: str|bytes) -> null",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "_write('/test.txt', 'example file data') => null",
			}.String(),
		},
	},
	{
		Name: "set",
		Builtin: &Builtin{
			// This is needed so we can return a set from a list of objects
			Fun: func(args ...Object) Object {
				if len(args) != 0 && len(args) != 1 {
					return newInvalidArgCountError("set", len(args), 0, "or 1")
				}
				setMap := NewSetElements()
				resultSet := &Set{Elements: setMap}
				if len(args) == 0 {
					return resultSet
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("set", 1, LIST_OBJ, args[0].Type())
				}
				elements := args[0].(*List).Elements
				for _, e := range elements {
					hashKey := HashObject(e)
					resultSet.Elements.Set(hashKey, SetPair{Value: e, Present: struct{}{}})
				}
				return resultSet
			},
			HelpStr: helpStrArgs{
				explanation: "`set` returns the SET version of a LIST, or an empty set with no args",
				signature:   "set(arg: list[any]|none) -> set[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "set([1,2,2,3]) => {1,2,3}",
			}.String(),
		},
	},
	// This function is lossy
	{
		Name: "int",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("int", len(args), 1, "")
				}
				switch args[0].Type() {
				case FLOAT_OBJ:
					return &Integer{Value: int64(args[0].(*Float).Value)}
				case UINTEGER_OBJ:
					return &Integer{Value: int64(args[0].(*UInteger).Value)}
				case BIG_INTEGER_OBJ:
					return &Integer{Value: args[0].(*BigInteger).Value.Int64()}
				case BIG_FLOAT_OBJ:
					return &Integer{Value: args[0].(*BigFloat).Value.IntPart()}
				case INTEGER_OBJ:
					return args[0]
				case STRING_OBJ:
					s := args[0].(*Stringo).Value
					i, err := strconv.ParseInt(s, 10, 64)
					if err != nil {
						return newError("`int` error: %s", err.Error())
					}
					return &Integer{Value: i}
				default:
					return newError("`int` error: unsupported type '%s'", args[0].Type())
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`int` returns the INTEGER version of a number or STRING, it will error out on unsupported types or a parse error",
				signature:   "int(arg: float|uint|bint|bfloat|str|int) -> int",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "int('123') => 123",
			}.String(),
		},
	},
	// This function is lossy
	{
		Name: "float",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("float", len(args), 1, "")
				}
				switch args[0].Type() {
				case INTEGER_OBJ:
					return &Float{Value: float64(args[0].(*Integer).Value)}
				case UINTEGER_OBJ:
					return &Float{Value: float64(args[0].(*UInteger).Value)}
				case BIG_INTEGER_OBJ:
					return &Float{Value: float64(args[0].(*BigInteger).Value.Int64())}
				case BIG_FLOAT_OBJ:
					return &Float{Value: args[0].(*BigFloat).Value.InexactFloat64()}
				case FLOAT_OBJ:
					return args[0]
				case STRING_OBJ:
					s := args[0].(*Stringo).Value
					f, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return newError("`float` error: %s", err.Error())
					}
					return &Float{Value: f}
				default:
					return newError("`float` error: unsupported type '%s'", args[0].Type())
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`float` returns the FLOAT version of the number or string, with potential loss, and it will error out on unsupported types or a parse error",
				signature:   "float(arg: int|uint|bint|bfloat|str|float) -> float",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "float('123') => 123.0",
			}.String(),
		},
	},
	// This function is lossy
	{
		Name: "bigint",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("bigint", len(args), 1, "")
				}
				switch args[0].Type() {
				case INTEGER_OBJ:
					return &BigInteger{Value: new(big.Int).SetInt64(args[0].(*Integer).Value)}
				case FLOAT_OBJ:
					return &BigInteger{Value: new(big.Int).SetInt64(int64(args[0].(*Float).Value))}
				case UINTEGER_OBJ:
					return &BigInteger{Value: new(big.Int).SetUint64(args[0].(*UInteger).Value)}
				case BIG_FLOAT_OBJ:
					return &BigInteger{Value: args[0].(*BigFloat).Value.BigInt()}
				case BIG_INTEGER_OBJ:
					return args[0]
				case STRING_OBJ:
					s := args[0].(*Stringo).Value
					b, ok := new(big.Int).SetString(s, 10)
					if !ok || b == nil {
						return newError("`bigint` error: %s is not a valid bigint", s)
					}
					return &BigInteger{Value: b}
				default:
					return newError("`bigint` error: unsupported type '%s'", args[0].Type())
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`bigint` returns the BIG INTEGER version of a number or STRING, it will error out on unsupported types or a parse error",
				signature:   "bigint(arg: int|float|uint|bfloat|str|bint) -> bint",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "bigint('123') => 123n",
			}.String(),
		},
	},
	// This function is lossy
	{
		Name: "bigfloat",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("bigfloat", len(args), 1, "")
				}
				switch args[0].Type() {
				case INTEGER_OBJ:
					return &BigFloat{Value: decimal.NewFromInt(args[0].(*Integer).Value)}
				case FLOAT_OBJ:
					return &BigFloat{Value: decimal.NewFromFloat(args[0].(*Float).Value)}
				case UINTEGER_OBJ:
					return &BigFloat{Value: decimal.NewFromBigInt(new(big.Int).SetUint64(args[0].(*UInteger).Value), 0)}
				case BIG_INTEGER_OBJ:
					return &BigFloat{Value: decimal.NewFromBigInt(args[0].(*BigInteger).Value, 0)}
				case BIG_FLOAT_OBJ:
					return args[0]
				case STRING_OBJ:
					s := args[0].(*Stringo).Value
					bf, err := decimal.NewFromString(s)
					if err != nil {
						return newError("`bigfloat` error: %s", err.Error())
					}
					return &BigFloat{Value: bf}
				default:
					return newError("`bigfloat` error: unsupported type '%s'", args[0].Type())
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`bigfloat` returns the BIG FLOAT version of a number or STRING, it will error out on unsupported types or a parse error",
				signature:   "bigfloat(arg: int|float|uint|bint|str|bfloat) -> bfloat",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "bigfloat('123') => 123.0n",
			}.String(),
		},
	},
	// This function is lossy
	{
		Name: "uint",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("uint", len(args), 1, "")
				}
				switch args[0].Type() {
				case INTEGER_OBJ:
					return &UInteger{Value: uint64(args[0].(*Integer).Value)}
				case FLOAT_OBJ:
					return &UInteger{Value: uint64(args[0].(*Float).Value)}
				case BIG_INTEGER_OBJ:
					return &UInteger{Value: args[0].(*BigInteger).Value.Uint64()}
				case BIG_FLOAT_OBJ:
					return &UInteger{Value: args[0].(*BigFloat).Value.BigInt().Uint64()}
				case UINTEGER_OBJ:
					return args[0]
				case STRING_OBJ:
					s := args[0].(*Stringo).Value
					u, err := strconv.ParseUint(s, 10, 64)
					if err != nil {
						return newError("`uint` error: %s", err.Error())
					}
					return &UInteger{Value: u}
				default:
					return newError("`uint` error: unsupported type '%s'", args[0].Type())
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`uint` returns the UINTEGER version of a number or STRING, it will error out on unsupported types or a parse error",
				signature:   "uint(arg: int|float|bint|bfloat|str|uint) -> uint",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "uint('123') => 0u123",
			}.String(),
		},
	},
	{
		Name: "eval_template",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("eval_template", len(args), 2, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("eval_template", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != MAP_OBJ {
					return newPositionalTypeError("eval_template", 2, MAP_OBJ, args[1].Type())
				}
				m := args[1].(*Map)
				ctx := plush.NewContext()
				for _, k := range m.Pairs.Keys {
					mp, _ := m.Pairs.Get(k)
					if mp.Key.Type() != STRING_OBJ {
						return newError("`eval_template` error: found key in MAP that was not STRING. got=%s", mp.Key.Type())
					}
					val, err := blueObjectToGoObject(mp.Value)
					if err != nil {
						return newError("`eval_template` error: %s", err.Error())
					}
					ctx.Set(mp.Key.(*Stringo).Value, val)
				}
				inputStr := args[0].(*Stringo).Value
				s, err := plush.Render(inputStr, ctx)
				if err != nil {
					return newError("`eval_template` error: %s", err.Error())
				}
				return &Stringo{Value: s}
			},
			HelpStr: helpStrArgs{
				explanation: "`eval_template` returns the STRING version of a template parsed with plush (https://github.com/gobuffalo/plush)",
				signature:   "eval_template(tmplStr: str, tmplMap: map) -> str",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "eval_template('<%= arg %>', {'arg': 123}) => '123'",
			}.String(),
		},
	},
	{
		Name: "error",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("error", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("error", 1, STRING_OBJ, args[0].Type())
				}
				msg, ok := args[0].(*Stringo)
				if !ok {
					return newError("`error` argument 1 was not STRING. got=%T", args[0])
				}
				return &Error{Message: msg.Value}
			},
			HelpStr: helpStrArgs{
				explanation: "`error` returns an EvaluatorError for the given STRING",
				signature:   "error(arg: str) -> EvaluatorError: #{arg}",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "error('fail') => ERROR| EvaluatorError: fail",
			}.String(),
		},
	},
	{
		Name: "assert",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 && len(args) != 2 {
					return newInvalidArgCountError("assert", len(args), 1, "or 2")
				}
				if args[0].Type() != BOOLEAN_OBJ {
					return newPositionalTypeError("assert", 1, BOOLEAN_OBJ, args[0].Type())
				}
				b, ok := args[0].(*Boolean)
				if !ok {
					return newError("`assert` argument 1 was not BOOLEAN. got=%T", args[0])
				}
				if len(args) == 1 {
					if b.Value {
						return TRUE
					} else {
						return newError("`assert` failed")
					}
				}
				if args[1].Type() != STRING_OBJ {
					return newPositionalTypeError("assert", 2, STRING_OBJ, args[1].Type())
				}
				msg, ok := args[1].(*Stringo)
				if !ok {
					return newError("`assert` second argument was not STRING. got=%T", args[1])
				}
				if b.Value {
					return TRUE
				}
				return newError("`assert` failed: %s", msg.Value)
			},
			HelpStr: helpStrArgs{
				explanation: "`assert` returns an EvaluatorError if the given Condition is false, and it will print the optional STRING provided as an error message",
				signature:   "assert(cond: bool, message='') -> true|EvaluatorError",
				errors:      "InvalidArgCount,PositionalType,AssertError,CustomAssertError",
				example:     "assert(true) => true\nassert(false, 'Message') => ERROR| EvaluatorError: `asser` failed: Message",
			}.String(),
		},
	},
	{
		Name: "type",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("type", len(args), 1, "")
				}
				return &Stringo{Value: string(args[0].Type())}
			},
			HelpStr: helpStrArgs{
				explanation: "`type` returns the STRING type representation of the given arg",
				signature:   "type(arg: any) -> str",
				errors:      "InvalidArgCount",
				example:     "type('Hello') => 'STRING'",
			}.String(),
		},
	},
	{
		Name: "exec",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("exec", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("exec", 1, STRING_OBJ, args[0].Type())
				}
				return ExecStringCommand(args[0].(*Stringo).Value)
			},
			HelpStr: helpStrArgs{
				explanation: "`exec` returns a STRING from the executed command",
				signature:   "exec(command: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "exec('echo hello') => hello\\n",
			}.String(),
		},
	},
	{
		Name: "is_alive",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("is_alive", len(args), 1, "")
				}
				if args[0].Type() != PROCESS_OBJ {
					return newPositionalTypeError("is_alive", 1, PROCESS_OBJ, args[0].Type())
				}
				p := args[0].(*Process)
				_, isAlive := ProcessMap.Load(Pk(p.NodeName, p.Id))
				if isAlive {
					return TRUE
				}
				return FALSE
			},
			HelpStr: helpStrArgs{
				explanation: "`is_alive` returns a BOOLEAN if the given UINTEGER pid is alive",
				signature:   "is_alive(pid: uint) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "is_alive(0x0) => true",
			}.String(),
		},
	},
	{
		Name: "exit",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) == 0 {
					os.Exit(0)
					// Unreachable
					return NULL
				} else if len(args) == 1 {
					if args[0].Type() != INTEGER_OBJ {
						return newError("argument passed to `exit` must be INTEGER. got=%s", args[0].Type())
					}
					os.Exit(int(args[0].(*Integer).Value))
					// Unreachable
					return NULL
				} else {
					return newInvalidArgCountError("exit", len(args), 0, "or 1")
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`exit` returns nothing as it will exit the program's execution",
				signature:   "exit(exit_code: int) -> None",
				errors:      "InvalidArgCount",
				example:     "exit(0) => PROGRAM EXECUTION HALTS",
			}.String(),
		},
	},
	{
		Name: "cwd",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) > 0 {
					return newInvalidArgCountError("cwd", len(args), 0, "")
				}
				dir, err := os.Getwd()
				if err != nil {
					return newError("`cwd` error: %s", err.Error())
				}
				return &Stringo{Value: dir}
			},
			HelpStr: helpStrArgs{
				explanation: "`cwd` returns the STRING path of the current working directory",
				signature:   "cwd() -> str",
				errors:      "InvalidArgCount,CustomError",
				example:     "cwd() => '/home/user/...'",
			}.String(),
		},
	},
	{
		Name: "cd",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("cd", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("cd", 1, STRING_OBJ, args[0].Type())
				}
				path := args[0].(*Stringo).Value
				if strings.HasPrefix(path, "~") {
					home, err := os.UserHomeDir()
					if err != nil {
						return newError("`cd` error: %s", err.Error())
					}
					path = strings.ReplaceAll(path, "~", home)
				}
				p, err := filepath.Abs(path)
				if err != nil {
					return newError("`cd` error: %s", err.Error())
				}
				err = os.Chdir(p)
				if err != nil {
					return newError("`cd` error: %s", err.Error())
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`cd` returns NULL and changes the current working directory to the given path",
				signature:   "cd(path: str) -> null",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "cd('/home/user') => null",
			}.String(),
		},
	},
	{
		Name: "_to_bytes",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_bytes", len(args), 3, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("to_bytes", 1, STRING_OBJ, args[0].Type())
				}
				return &Bytes{
					Value: []byte(args[0].(*Stringo).Value),
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`to_bytes` returns the BYTE representation of the given STRING",
				signature:   "to_bytes(arg: str) -> bytes",
				errors:      "InvalidArgCount,TypeNotSupported",
				example:     "to_bytes('hello') => []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}",
			}.String(),
		},
	},
	{
		Name: "str",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("str", len(args), 1, "")
				}
				if args[0].Type() == BYTES_OBJ {
					return &Stringo{Value: string(args[0].(*Bytes).Value)}
				}
				return &Stringo{Value: args[0].Inspect()}
			},
			HelpStr: helpStrArgs{
				explanation: "`str` returns the STRING representation of the given BYTES or the inspected object",
				signature:   "str(arg: any) -> str",
				errors:      "InvalidArgCount",
				example:     "str([]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}) => 'hello'",
			}.String(),
		},
	},
	{
		Name: "is_file",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("is_file", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("is_file", 1, STRING_OBJ, args[0].Type())
				}
				fpath := args[0].(*Stringo).Value
				info, err := os.Stat(fpath)
				if os.IsNotExist(err) {
					return FALSE
				}
				if info.IsDir() {
					return FALSE
				}
				return TRUE
			},
			HelpStr: helpStrArgs{
				explanation: "`is_file` returns TRUE if the given STRING path is a file",
				signature:   "is_file(path: str) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "is_file('/test') => false",
			}.String(),
		},
	},
	{
		Name: "is_dir",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("is_dir", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("is_dir", 1, STRING_OBJ, args[0].Type())
				}
				fpath := args[0].(*Stringo).Value
				info, err := os.Stat(fpath)
				if err != nil {
					return FALSE
				}
				if info.IsDir() {
					return TRUE
				}
				return FALSE
			},
			HelpStr: helpStrArgs{
				explanation: "`is_dir` returns TRUE if the given STRING path is a directory",
				signature:   "is_dir(path: str) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "is_dir('/test') => false",
			}.String(),
		},
	},
	{
		Name: "find_exe",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("find_exe", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("find_exe", 1, STRING_OBJ, args[0].Type())
				}
				exePath := args[0].(*Stringo).Value
				fname, err := exec.LookPath(exePath)
				if err == nil {
					fname, err = filepath.Abs(fname)
				}
				if err != nil {
					return newError("`find_exe` error: %s", err.Error())
				}
				return &Stringo{Value: fname}
			},
			HelpStr: helpStrArgs{
				explanation: "`find_exe` returns the STRING path of the given STRING executable name",
				signature:   "find_exe(exe_name: str) -> str",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "find_exe('blue') => /home/user/.blue/bin/blue",
			}.String(),
		},
	},
	// TODO: Do we want to do that thing where we shell expand home dir? or other things like that?
	{
		Name: "rm",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("rm", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("rm", 1, STRING_OBJ, args[0].Type())
				}
				s := args[0].(*Stringo).Value
				finfo, err := os.Stat(s)
				if err != nil {
					return newError("`rm` error: %s", err.Error())
				}
				if finfo.IsDir() {
					err = os.RemoveAll(s)
					if err != nil {
						return newError("`rm` error: %s", err.Error())
					}
					return NULL
				}
				err = os.Remove(s)
				if err != nil {
					return newError("`rm` error: %s", err.Error())
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`rm` removes the given STRING file or directory path",
				signature:   "rm(path: str) -> null",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "rm('/test') => null",
			}.String(),
		},
	},
	// TODO: Do we want to do that thing where we shell expand home dir? or other things like that?
	{
		Name: "ls",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) > 1 {
					return newInvalidArgCountError("ls", len(args), 0, "or 1")
				}
				var cwd string
				if len(args) == 0 {
					// use cwd to ls
					dir, err := os.Getwd()
					if err != nil {
						return newError("`ls` error: %s", err.Error())
					}
					cwd = dir
				} else {
					// use argument passed in
					if args[0].Type() != STRING_OBJ {
						return newPositionalTypeError("ls", 1, STRING_OBJ, args[0].Type())
					}
					cwd = args[0].(*Stringo).Value
				}
				fileOrDirs, err := os.ReadDir(cwd)
				if err != nil {
					return newError("`ls` error: %s", err.Error())
				}
				result := &List{Elements: make([]Object, len(fileOrDirs))}
				for i := 0; i < len(fileOrDirs); i++ {
					result.Elements[i] = &Stringo{Value: fileOrDirs[i].Name()}
				}
				return result
			},
			HelpStr: helpStrArgs{
				explanation: "`ls` returns a LIST of STRINGs of all the files and directories in the given path",
				signature:   "ls(path: str) -> list[str]",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "ls('/test') => []",
			}.String(),
		},
	},
	// TODO: Eventually we need to support files better (and possibly, stdin, stderr, stdout) and then http stuff
	{
		Name: "is_valid_json",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("is_valid_json", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("is_valid_json", 1, STRING_OBJ, args[0].Type())
				}
				s := args[0].(*Stringo).Value
				return nativeToBooleanObject(json.Valid([]byte(s)))
			},
			HelpStr: helpStrArgs{
				explanation: "`is_valid_json` returns a BOOLEAN if the given STRING is valid json",
				signature:   "is_valid_json(json: str) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "is_valid_json('{}') => true",
			}.String(),
		},
	},
	{
		Name: "wait",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				// This function will avoid returning errors
				// but that means random inputs will technically be allowed
				processesToWaitFor := []*Process{}
				for _, arg := range args {
					if processes, ok := getListOfProcesses(arg); ok {
						processesToWaitFor = append(processesToWaitFor, processes...)
						continue
					}
					if arg.Type() == PROCESS_OBJ {
						v := arg.(*Process)
						processesToWaitFor = append(processesToWaitFor, v)
					}
				}
				for {
					allDone := false
					for _, p := range processesToWaitFor {
						_, ok := ProcessMap.Load(Pk(p.NodeName, p.Id))
						allDone = allDone || ok
					}
					// They should all be false for us to exit
					if !allDone {
						return NULL
					}
					// 10us sleep between round of checks?
					time.Sleep(10 * time.Microsecond)
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`wait` will wait for the given PIDs to finish execution",
				signature:   "wait(pids: list[{t: 'pid', v: _}]|uint...) -> null",
				errors:      "None",
				example:     "wait({t: 'pid', v: 1}) => null",
			}.String(),
		},
	},
	{
		Name: "_publish",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				// publish('TOPIC', MSG) -> non-blocking send
				if len(args) != 2 {
					return newInvalidArgCountError("publish", len(args), 2, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("publish", 1, STRING_OBJ, args[0].Type())
				}
				topic := args[0].(*Stringo).Value
				PubSubBroker.Publish(topic, args[1])
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`publish` will publish an OBJECT on a topic STRING to all subscribers of the topic",
				signature:   "publish(topic: str, value: any) -> null",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "publish('blue', 123) => null",
			}.String(),
		},
	},
	{
		Name: "_broadcast",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				// broadcast(MSG) -> non-blocking send
				// broadcast(MSG, ['some', 'specifc', 'channels']) -> non-blocking send
				if len(args) != 1 && len(args) != 2 {
					return newInvalidArgCountError("broadcast", len(args), 1, "or 2")
				}
				if len(args) == 2 && args[1].Type() != LIST_OBJ {
					return newPositionalTypeError("broadcast", 2, LIST_OBJ, args[1].Type())
				}
				if len(args) == 1 {
					PubSubBroker.BroadcastToAllTopics(args[0])
					return NULL
				}
				l := args[1].(*List).Elements
				topics := make([]string, len(l))
				for i, e := range l {
					if e.Type() != STRING_OBJ {
						return newError("`broadcast` error: all elements in list should be STRING. found=%s", e.Type())
					}
					topics[i] = e.(*Stringo).Value
				}
				PubSubBroker.Broadcast(args[0], topics)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`broadcast` will broadcast an OBJECT to all subscribers of the pubsub broker",
				signature:   "broadcast(value: any) -> null",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "broadcast(123) => null",
			}.String(),
		},
	},
	// Functions for subscribers in pubsub
	{
		Name: "add_topic",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("add_topic", len(args), 2, "")
				}
				if args[0].Type() != MAP_OBJ {
					return newPositionalTypeError("add_topic", 1, MAP_OBJ, args[0].Type())
				}
				t, sub, ok := getBasicObjectForGoObj[*Subscriber](args[0])
				if t != "sub" {
					return newError("`add_topic` error: argument 1 should be in the format {t: 'sub', v: uint}")
				}
				if !ok {
					return newError("`add_topic` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*Subscriber]}. got=%s", args[0].Inspect())
				}
				if args[1].Type() != STRING_OBJ {
					return newPositionalTypeError("add_topic", 2, STRING_OBJ, args[1].Type())
				}
				topic := args[1].(*Stringo).Value
				sub.AddTopic(topic)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`add_topic` will add a topic STRING to a subscriber object",
				signature:   "add_topic(sub: {t: 'sub', v: _}, topic: str) -> null",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "add_topic({t: 'sub', v: 1}, 'blue') => null",
			}.String(),
		},
	},
	// TODO: add_topics, remove_topics?
	{
		Name: "remove_topic",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("remove_topic", len(args), 2, "")
				}
				if args[0].Type() != MAP_OBJ {
					return newPositionalTypeError("remove_topic", 1, MAP_OBJ, args[0].Type())
				}
				t, sub, ok := getBasicObjectForGoObj[*Subscriber](args[0])
				if t != "sub" {
					return newError("`remove_topic` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*Subscriber]}")
				}
				if !ok {
					return newError("`remove_topic` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*Subscriber]}. got=%s", args[0].Inspect())
				}
				if args[1].Type() != STRING_OBJ {
					return newPositionalTypeError("remove_topic", 2, STRING_OBJ, args[1].Type())
				}
				topic := args[1].(*Stringo).Value
				sub.RemoveTopic(topic)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`remove_topic` will remove a topic STRING from a subscriber object",
				signature:   "remove_topic(sub: {t: 'sub', v: _}, topic: str) -> null",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "remove_topic({t: 'sub', v: _}, 'blue') => null",
			}.String(),
		},
	},
	{
		Name: "_subscribe",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				// subscribe('TOPIC') -> {t: 'sub', v: _} -> _.recv()
				if len(args) != 1 {
					return newInvalidArgCountError("subscribe", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("subscribe", 1, STRING_OBJ, args[0].Type())
				}
				topic := args[0].(*Stringo).Value
				subId := SubscriberCount.Add(1)
				sub := PubSubBroker.AddSubscriber(subId)
				PubSubBroker.Subscribe(sub, topic)
				return CreateBasicMapObjectForGoObj("sub", NewGoObj(sub))
			},
			HelpStr: helpStrArgs{
				explanation: "`subscribe` will add a subscriber to the pubsub broker for a topic",
				signature:   "subscribe(sub: str) -> {t: 'sub', v: _}",
				errors:      "InvalidArgCount,PositionalType",
				example:     "subscribe('TOPIC') => {t: 'sub', v: _} -> _.recv()",
			}.String(),
		},
	},
	{
		Name: "unsubscribe",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("unsubscribe", len(args), 2, "")
				}
				if args[0].Type() != MAP_OBJ {
					return newPositionalTypeError("unsubscribe", 1, MAP_OBJ, args[0].Type())
				}
				t, sub, ok := getBasicObjectForGoObj[*Subscriber](args[0])
				if t != "sub" {
					return newError("`unsubscribe` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*Subscriber]}")
				}
				if !ok {
					return newError("`unsubscribe` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*Subscriber]}. got=%s", args[0].Inspect())
				}
				// Remove should also destruct the subscriber
				PubSubBroker.RemoveSubscriber(sub)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`unsubscribe` will remove a subscriber object from the pubsub broker",
				signature:   "unsubscribe(sub: {t: 'sub', v: _}) -> null",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "unsubscribe({t: 'sub', v: _}) => null",
			}.String(),
		},
	},
	{
		Name: "_pubsub_sub_listen",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("pubsub_sub_listen", len(args), 1, "")
				}
				if args[0].Type() != GO_OBJ {
					return newPositionalTypeError("pubsub_sub_listen", 1, GO_OBJ, args[0].Type())
				}
				sub, ok := args[0].(*GoObj[*Subscriber])
				if !ok {
					return newPositionalTypeErrorForGoObj("pubsub_sub_listen", 1, "*Subscriber", args[0])
				}
				return sub.Value.PollMessage()
			},
			HelpStr: helpStrArgs{
				explanation: "`pubsub_sub_listen` is used when receiving on a subscribed topic for the kv",
				signature:   "pubsub_sub_listen(arg: {t: 'sub', v: _}) -> any",
				errors:      "InvalidArgCount,PositionalType",
				example:     "pubsub_sub_listen({t: 'sub', v: _}) => any",
			}.String(),
		},
	},
	{
		Name: "_get_subscriber_count",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 && len(args) != 0 {
					return newInvalidArgCountError("get_subscriber_count", len(args), 0, "or 1")
				}
				if len(args) == 0 {
					// Get total count of subscribers
					return &Integer{Value: int64(PubSubBroker.GetTotalSubscribers())}
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("get_subscriber_count", 1, STRING_OBJ, args[0].Type())
				}
				topic := args[0].(*Stringo).Value
				return &Integer{Value: int64(PubSubBroker.GetNumSubscribersForTopic(topic))}
			},
			HelpStr: helpStrArgs{
				explanation: "`get_subscriber_count` returns the number of subscribers for a topic, if there is no topic passed in the total subscribers are returned",
				signature:   "get_subscriber_count(arg: str|none) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "get_subscriber_count('TOPIC') => 1",
			}.String(),
		},
	},
	{
		Name: "_kv_put",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 3 {
					return newInvalidArgCountError("kv_put", len(args), 3, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("kv_put", 1, STRING_OBJ, args[0].Type())
				}
				topic := args[0].(*Stringo).Value
				var m *Map
				m, ok := KVMap.Load(topic)
				if !ok {
					m = &Map{
						Pairs: NewPairsMap(),
					}
				}
				hashedKey := HashObject(args[1])
				hk := HashKey{Type: args[1].Type(), Value: hashedKey}
				m.Pairs.Set(hk, MapPair{
					Key:   args[1],
					Value: args[2],
				})
				KVMap.Store(topic, m)
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`kv_put` puts a value into the kv for a specific topic and key",
				signature:   "kv_put(topic: str, key: any, val: any) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "kv_put('TOPIC', 1, 3) => null",
			}.String(),
		},
	},
	{
		Name: "_kv_get",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("kv_get", len(args), 2, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("kv_get", 1, STRING_OBJ, args[0].Type())
				}
				topic := args[0].(*Stringo).Value
				var m *Map
				m, ok := KVMap.Load(topic)
				if !ok {
					// Return NULL if the topic doesn't have a map that exists
					return NULL
				}
				hashedKey := HashObject(args[1])
				hk := HashKey{Type: args[1].Type(), Value: hashedKey}
				val, ok := m.Pairs.Get(hk)
				if !ok {
					// Return NULL if the key doesnt exist on the map at the topic
					return NULL
				}
				return val.Value
			},
			HelpStr: helpStrArgs{
				explanation: "`kv_get` gets a value from the kv for a specific topic and key",
				signature:   "kv_get(topic: str, key: any) -> any",
				errors:      "InvalidArgCount,PositionalType",
				example:     "kv_get('TOPIC', 1) => 3",
			}.String(),
		},
	},
	{
		Name: "_kv_delete",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 && len(args) != 2 {
					return newInvalidArgCountError("kv_delete", len(args), 1, "or 2")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("kv_delete", 1, STRING_OBJ, args[0].Type())
				}
				topic := args[0].(*Stringo).Value
				if len(args) == 1 {
					// If its 1 we want to delete a topic, and the associated map
					KVMap.Delete(topic)
					return NULL
				} else {
					// If its 2 we want to delete a key from a map on a topic
					var m *Map
					m, ok := KVMap.Load(topic)
					if !ok {
						// Return NULL if the topic doesn't have a map that exists
						// theres nothing to delete in this case
						return NULL
					}
					hashedKey := HashObject(args[1])
					hk := HashKey{Type: args[1].Type(), Value: hashedKey}
					m.Pairs.Delete(hk)
					KVMap.Store(topic, m)
					return NULL
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`kv_delete` deletes a value from the kv for a specific topic and key, if there is no key the topic is removed",
				signature:   "kv_delete(topic: str, key: any|none) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "kv_delete('TOPIC', 1) => null",
			}.String(),
		},
	},
	{
		Name: "_new_uuid",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("new_uuid", len(args), 0, "")
				}
				return &Stringo{Value: uuid.NewString()}
			},
			HelpStr: helpStrArgs{
				explanation: "`new_uuid` returns a new random UUID STRING",
				signature:   "new_uuid() -> str",
				errors:      "InvalidArgCount",
				example:     "new_uuid() => 'a38dc5fa-7f18-4e1c-8a70-f8d343109708'",
			}.String(),
		},
	},
	// This is straight out of golang's example for runtime/metrics https://pkg.go.dev/runtime/metrics
	{
		Name: "_go_metrics",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("go_metrics", len(args), 0, "")
				}
				// TODO: Update this to return a map?
				var out bytes.Buffer
				// Get descriptions for all supported metrics.
				descs := metrics.All()

				// Create a sample for each metric.
				samples := make([]metrics.Sample, len(descs))
				for i := range samples {
					samples[i].Name = descs[i].Name
				}

				// Sample the metrics. Re-use the samples slice if you can!
				metrics.Read(samples)

				// Iterate over all results.
				for _, sample := range samples {
					// Pull out the name and value.
					name, value := sample.Name, sample.Value

					// Handle each sample.
					switch value.Kind() {
					case metrics.KindUint64:
						out.WriteString(fmt.Sprintf("%s: %d\n", name, value.Uint64()))
					case metrics.KindFloat64:
						out.WriteString(fmt.Sprintf("%s: %f\n", name, value.Float64()))
					case metrics.KindFloat64Histogram:
						// The histogram may be quite large, so let's just pull out
						// a crude estimate for the median for the sake of this example.
						out.WriteString(fmt.Sprintf("%s: %f\n", name, medianBucket(value.Float64Histogram())))
					case metrics.KindBad:
						// This should never happen because all metrics are supported
						// by construction.
						panic("bug in runtime/metrics package!")
					default:
						// This may happen as new metrics get added.
						//
						// The safest thing to do here is to simply log it somewhere
						// as something to look into, but ignore it for now.
						// In the worst case, you might temporarily miss out on a new metric.
						out.WriteString(fmt.Sprintf("%s: unexpected metric Kind: %v\n", name, value.Kind()))
					}
				}
				return &Stringo{Value: out.String()}
			},
			HelpStr: helpStrArgs{
				explanation: "`go_metrics` returns the STRING version of the golang runtime metrics",
				signature:   "go_metrics() -> str",
				errors:      "InvalidArgCount",
				example:     "go_metrics() => str",
			}.String(),
		},
	},
	{
		Name: "get_os",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("get_os", len(args), 0, "")
				}
				return &Stringo{Value: runtime.GOOS}
			},
			HelpStr: helpStrArgs{
				explanation: "`get_os` returns the STRING GOOS of the runtime",
				signature:   "get_os() -> str",
				errors:      "InvalidArgCount",
				example:     "get_os() => windows",
			}.String(),
		},
	},
	{
		Name: "get_arch",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("get_arch", len(args), 0, "")
				}
				return &Stringo{Value: runtime.GOARCH}
			},
			HelpStr: helpStrArgs{
				explanation: "`get_arch` returns the STRING GOARCH of the runtime",
				signature:   "get_arch() -> str",
				errors:      "InvalidArgCount",
				example:     "get_arch() => amd64",
			}.String(),
		},
	},
	{
		Name: "_gc",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("gc", len(args), 0, "")
				}
				runtime.GC()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`gc` calls the golang garbage collector",
				signature:   "gc() -> null",
				errors:      "InvalidArgCount",
				example:     "gc() => null",
			}.String(),
		},
	},
	{
		Name: "_version",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("version", len(args), 0, "")
				}
				return &Stringo{Value: fmt.Sprintf("%s-%s", runtime.Version(), consts.VERSION)}
			},
			HelpStr: helpStrArgs{
				explanation: "`version` returns the golang version and blue version hyphenated",
				signature:   "version() -> str",
				errors:      "InvalidArgCount",
				example:     "version() => go1.21.5-0.1.16-684f398-windows/amd64",
			}.String(),
		},
	},
	{
		Name: "_num_cpu",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("num_cpu", len(args), 0, "")
				}
				return &Integer{Value: int64(runtime.NumCPU())}
			},
			HelpStr: helpStrArgs{
				explanation: "`num_cpu` returns the number of cpus available to the blue process",
				signature:   "num_cpu() -> int",
				errors:      "InvalidArgCount",
				example:     "num_cpu() => 12",
			}.String(),
		},
	},
	{
		Name: "_num_process",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("num_process", len(args), 0, "")
				}
				return &Integer{Value: int64(runtime.NumGoroutine())}
			},
			HelpStr: helpStrArgs{
				explanation: "`num_process` returns the number of processes used by the runtime",
				signature:   "num_process() -> int",
				errors:      "InvalidArgCount",
				example:     "num_process() => 6",
			}.String(),
		},
	},
	{
		Name: "_num_max_cpu",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("num_max_cpu", len(args), 0, "")
				}
				return &Integer{Value: int64(runtime.GOMAXPROCS(-1))}
			},
			HelpStr: helpStrArgs{
				explanation: "`num_max_cpu` returns the max number of cpus available to the runtime",
				signature:   "num_max_cpu() -> int",
				errors:      "InvalidArgCount",
				example:     "num_max_cpu() => 12",
			}.String(),
		},
	},
	{
		Name: "_num_os_thread",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("num_os_thread", len(args), 0, "")
				}
				return &Integer{Value: int64(pprof.Lookup("threadcreate").Count())}
			},
			HelpStr: helpStrArgs{
				explanation: "`num_os_thread` returns the number of os threads being used by the runtime",
				signature:   "num_os_thread() -> int",
				errors:      "InvalidArgCount",
				example:     "num_os_thread() => 15",
			}.String(),
		},
	},
	{
		Name: "_set_max_cpu",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("set_max_cpu", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("set_max_cpu", 1, INTEGER_OBJ, args[0].Type())
				}
				i := int(args[0].(*Integer).Value)
				return &Integer{Value: int64(runtime.GOMAXPROCS(i))}
			},
			HelpStr: helpStrArgs{
				explanation: "`set_max_cpu` sets the max number of cpus for the runtime and returns the previous setting. if arg < 1 => defaults to current number of cpus",
				signature:   "set_max_cpu(arg: int) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "set_max_cpu(3) => 6",
			}.String(),
		},
	},
	{
		Name: "_set_gc_percent",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("set_gc_percent", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("set_gc_percent", 1, INTEGER_OBJ, args[0].Type())
				}
				i := int(args[0].(*Integer).Value)
				return &Integer{Value: int64(debug.SetGCPercent(i))}
			},
			HelpStr: helpStrArgs{
				explanation: "`set_gc_percent` sets the gc target percentage and returns the previous setting. a lower setting essentially limits the memory, 100 is default, and negative numbers turn gc off",
				signature:   "set_gc_percent(arg: int) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "set_gc_percent(30) => 100",
			}.String(),
		},
	},
	{
		Name: "_get_mem_stats",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("get_mem_stats", len(args), 0, "")
				}
				formatBytes := func(val uint64) string {
					units := []string{" bytes", "KB", "MB", "GB", "TB", "PB"}
					var i int
					var target uint64
					for i = range units {
						target = 1 << uint(10*(i+1))
						if val < target {
							break
						}
					}
					if i > 0 {
						return fmt.Sprintf("%0.2f%s (%d bytes)", float64(val)/(float64(target)/1024), units[i], val)
					}
					return fmt.Sprintf("%d bytes", val)
				}
				var s runtime.MemStats
				runtime.ReadMemStats(&s)
				mapObj := NewOrderedMap[string, Object]()
				mapObj.Set("alloc", &Stringo{Value: formatBytes(s.Alloc)})
				mapObj.Set("total-alloc", &Stringo{Value: formatBytes(s.TotalAlloc)})
				mapObj.Set("sys", &Stringo{Value: formatBytes(s.Sys)})
				mapObj.Set("lookups", &UInteger{Value: s.Lookups})
				mapObj.Set("mallocs", &UInteger{Value: s.Mallocs})
				mapObj.Set("frees", &UInteger{Value: s.Frees})
				mapObj.Set("heap-alloc", &Stringo{Value: formatBytes(s.HeapAlloc)})
				mapObj.Set("heap-sys", &Stringo{Value: formatBytes(s.HeapSys)})
				mapObj.Set("heap-idle", &Stringo{Value: formatBytes(s.HeapIdle)})
				mapObj.Set("heap-in-use", &Stringo{Value: formatBytes(s.HeapInuse)})
				mapObj.Set("heap-released", &Stringo{Value: formatBytes(s.HeapReleased)})
				mapObj.Set("heap-objects", &UInteger{Value: s.HeapObjects})
				mapObj.Set("stack-in-use", &Stringo{Value: formatBytes(s.StackInuse)})
				mapObj.Set("stack-sys", &Stringo{Value: formatBytes(s.StackSys)})
				mapObj.Set("stack-mspan-inuse", &Stringo{Value: formatBytes(s.MSpanInuse)})
				mapObj.Set("stack-mspan-sys", &Stringo{Value: formatBytes(s.MSpanSys)})
				mapObj.Set("stack-mcache-inuse", &Stringo{Value: formatBytes(s.MCacheInuse)})
				mapObj.Set("stack-mcache-sys", &Stringo{Value: formatBytes(s.MCacheSys)})
				mapObj.Set("other-sys", &Stringo{Value: formatBytes(s.OtherSys)})
				mapObj.Set("gc-sys", &Stringo{Value: formatBytes(s.GCSys)})
				mapObj.Set("next-gc: when heap-alloc >=", &Stringo{Value: formatBytes(s.NextGC)})
				lastGC := "-"
				if s.LastGC != 0 {
					lastGC = fmt.Sprint(time.Unix(0, int64(s.LastGC)))
				}
				mapObj.Set("last-gc", &Stringo{Value: lastGC})
				mapObj.Set("gc-pause-total", &Stringo{Value: time.Duration(s.PauseTotalNs).String()})
				mapObj.Set("gc-pause", &UInteger{Value: s.PauseNs[(s.NumGC+255)%256]})
				mapObj.Set("gc-pause-end", &UInteger{Value: s.PauseEnd[(s.NumGC+255)%256]})
				mapObj.Set("num-gc", &UInteger{Value: uint64(s.NumGC)})
				mapObj.Set("num-forced-gc", &UInteger{Value: uint64(s.NumForcedGC)})
				mapObj.Set("gc-cpu-fraction", &Float{Value: s.GCCPUFraction})
				mapObj.Set("enable-gc", &Boolean{Value: s.EnableGC})
				mapObj.Set("debug-gc", &Boolean{Value: s.DebugGC})
				return CreateMapObjectForGoMap(*mapObj)
			},
			HelpStr: helpStrArgs{
				explanation: "`get_mem_stats` returns the runtime memory stats",
				signature:   "get_mem_stats() -> map[str]any",
				errors:      "InvalidArgCount",
				example:     "get_mem_stats() => object (all mem stats)",
			}.String(),
		},
	},
	{
		Name: "_get_stack_trace",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("get_stack_trace", len(args), 0, "")
				}
				return &Stringo{Value: string(debug.Stack())}
			},
			HelpStr: helpStrArgs{
				explanation: "`get_stack_trace` returns the runtime current stack trace",
				signature:   "get_stack_trace() -> str",
				errors:      "InvalidArgCount",
				example:     "get_stack_trace() => str",
			}.String(),
		},
	},
	{
		Name: "_free_os_mem",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("free_os_mem", len(args), 0, "")
				}
				debug.FreeOSMemory()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`free_os_mem` runs gc and returns memory to the os (this happens in background task regardless)",
				signature:   "free_os_mem() -> null",
				errors:      "InvalidArgCount",
				example:     "free_os_mem() => null",
			}.String(),
		},
	},
	{
		Name: "_print_stack_trace",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 0 {
					return newInvalidArgCountError("print_stack_trace", len(args), 0, "")
				}
				debug.PrintStack()
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`print_stack_trace` prints the current stack trace to stderr",
				signature:   "print_stack_trace() -> null",
				errors:      "InvalidArgCount",
				example:     "print_stack_trace() => null",
			}.String(),
		},
	},
	{
		Name: "_set_max_stack",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("set_max_stack", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("set_max_stack", 1, INTEGER_OBJ, args[0].Type())
				}
				i := int(args[0].(*Integer).Value)
				return &Integer{Value: int64(debug.SetMaxStack(i))}
			},
			HelpStr: helpStrArgs{
				explanation: "`set_max_stack` sets the max amount of memory that can be used by blue process, only limiting future stack sizes and returning previous setting",
				signature:   "set_max_stack(arg: int) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "set_max_stack(12*1024*1024) => 1073741824",
			}.String(),
		},
	},
	{
		Name: "_set_max_threads",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("set_max_threads", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("set_max_threads", 1, INTEGER_OBJ, args[0].Type())
				}
				i := int(args[0].(*Integer).Value)
				return &Integer{Value: int64(debug.SetMaxThreads(i))}
			},
			HelpStr: helpStrArgs{
				explanation: "`set_max_threads` sets the max number of os threads the program can use and returns the previous setting",
				signature:   "set_max_threads(arg: int) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "set_max_threads(20) => 1000",
			}.String(),
		},
	},
	{
		Name: "_set_mem_limit",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("set_mem_limit", len(args), 1, "")
				}
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("set_mem_limit", 1, INTEGER_OBJ, args[0].Type())
				}
				i := args[0].(*Integer).Value
				return &Integer{Value: debug.SetMemoryLimit(i)}
			},
			HelpStr: helpStrArgs{
				explanation: "`set_mem_limit` sets the soft max memory limit of the program, returning the previous setting, a negative limit retuns the current setting",
				signature:   "set_mem_limit(arg: int) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "set_mem_limit(12*1024*1024*1024*1024) => 2**64-1",
			}.String(),
		},
	},
	{
		Name: "re",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("re", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("re", 1, STRING_OBJ, args[0].Type())
				}
				s := args[0].(*Stringo).Value
				re, err := regexp.Compile(s)
				if err != nil {
					return newError("`re` error: %s", err.Error())
				}
				return &Regex{Value: re}
			},
			HelpStr: helpStrArgs{
				explanation: "`re` returns a regex object for the given string",
				signature:   "re(arg: str) -> regex",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "re('abc') => r/abc/",
			}.String(),
		},
	},
	{
		Name: "to_list",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_list", len(args), 1, "")
				}
				if args[0].Type() != SET_OBJ {
					return newPositionalTypeError("to_list", 1, SET_OBJ, args[0].Type())
				}
				s := args[0].(*Set).Elements
				newElems := []Object{}
				for _, k := range s.Keys {
					if obj, ok := s.Get(k); ok {
						newElems = append(newElems, obj.Value)
					}
				}
				return &List{Elements: newElems}
			},
			HelpStr: helpStrArgs{
				explanation: "`to_list` returns a list from the given set",
				signature:   "to_list(arg: set[any]) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "to_list({1,2,3}) => [1,2,3]",
			}.String(),
		},
	},
	{
		Name: "abs_path",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("abs_path", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("abs_path", 1, STRING_OBJ, args[0].Type())
				}
				// TODO: Fix so this works with embedded files?
				fpath := args[0].(*Stringo).Value
				path, err := filepath.Abs(fpath)
				if err != nil {
					return newError("`abs_path` error: %s", err.Error())
				}
				return &Stringo{Value: path}
			},
			HelpStr: helpStrArgs{
				explanation: "`abs_path` returns the absolute path of the given filepath",
				signature:   "abs_path(arg: str) -> str",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "abs_path('some_file.txt') => '/the/path/to/some_file.txt'",
			}.String(),
		},
	},
	{
		Name: "fmt",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("fmt", len(args), 2, "")
				}
				val, err := anyBlueObjectToGoObject(args[0])
				if err != nil {
					// Escape to just using value as string
					val = args[0].Inspect()
				}
				if args[1].Type() != STRING_OBJ {
					return newPositionalTypeError("fmt", 2, STRING_OBJ, args[1].Type())
				}
				fmtString := args[1].(*Stringo).Value
				return &Stringo{Value: fmt.Sprintf(fmtString, val)}
			},
			HelpStr: helpStrArgs{
				explanation: "`fmt` returns the formatted version of the given INTEGER",
				signature:   "fmt(arg: int, fmtStr: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "fmt(3, '%04b') => '0011'",
			}.String(),
		},
	},
	{
		Name: "save",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("save", len(args), 1, "")
				}
				bs, err := args[0].Encode()
				if err != nil {
					return newError("`save` error: %s", err.Error())
				}
				return &Bytes{Value: bs}
			},
			HelpStr: helpStrArgs{
				explanation: "`save` returns the bytes of the encoded object",
				signature:   "save(arg: any) -> bytes",
				errors:      "InvalidArgCount,CustomError",
				example:     "save(1234) => '82001904d2'",
			}.String(),
		},
	},
	{
		Name: "__hash",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("__hash", len(args), 1, "")
				}
				return &UInteger{Value: HashObject(args[0])}
			},
			HelpStr: "__hash returns the internal hash of an object",
		},
	},
	{
		Name: "startswith",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("startswith", len(args), 2, "")
				}
				arg0, ok := args[0].(*Stringo)
				if !ok {
					return newPositionalTypeError("startswith", 1, STRING_OBJ, args[0].Type())
				}
				arg1, ok := args[1].(*Stringo)
				if !ok {
					return newPositionalTypeError("startswith", 2, STRING_OBJ, args[1].Type())
				}
				return nativeToBooleanObject(strings.HasPrefix(arg0.Value, arg1.Value))
			},
			HelpStr: helpStrArgs{
				explanation: "`startswith` returns a BOOLEAN if the given STRING starts with the prefix STRING",
				signature:   "startswith(arg: str, prefix: str) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "startswith('Hello', 'Hell') => true",
			}.String(),
		},
	},
	{
		Name: "endswith",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("endswith", len(args), 2, "")
				}
				arg0, ok := args[0].(*Stringo)
				if !ok {
					return newPositionalTypeError("endswith", 1, STRING_OBJ, args[0].Type())
				}
				arg1, ok := args[1].(*Stringo)
				if !ok {
					return newPositionalTypeError("endswith", 2, STRING_OBJ, args[1].Type())
				}
				return nativeToBooleanObject(strings.HasSuffix(arg0.Value, arg1.Value))
			},
			HelpStr: helpStrArgs{
				explanation: "`endswith` returns a BOOLEAN if the given STRING ends with the suffix STRING",
				signature:   "endswith(arg: str, suffix: str) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "endswith('Hello', 'o') => true",
			}.String(),
		},
	},
	{
		Name: "split",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 && len(args) != 2 {
					return newInvalidArgCountError("split", len(args), 1, "or 2")
				}
				if len(args) == 1 {
					arg0, ok := args[0].(*Stringo)
					if !ok {
						return newPositionalTypeError("split", 1, STRING_OBJ, args[0].Type())
					}
					strList := strings.Split(arg0.Value, " ")
					return &List{Elements: createStringList(strList)}
				}
				if len(args) == 2 {
					arg0, ok := args[0].(*Stringo)
					if !ok {
						return newPositionalTypeError("split", 1, STRING_OBJ, args[0].Type())
					}
					if args[1].Type() != STRING_OBJ && args[1].Type() != REGEX_OBJ {
						return newPositionalTypeError("split", 2, "STRING or REGEX", args[1].Type())
					}
					var strList []string
					arg1, isStr := args[1].(*Stringo)
					if isStr {
						strList = strings.Split(arg0.Value, arg1.Value)
					} else {
						re := args[1].(*Regex).Value
						strList = re.Split(arg0.Value, -1)

					}
					return &List{Elements: createStringList(strList)}
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`split` returns a LIST of STRINGs based on a STRING separator",
				signature:   "split(arg: str, sep: str|regex) -> list[str]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "split('Hello', '') => ['H','e','l','l','o']",
			}.String(),
		},
	},
	{
		Name: "_replace",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 3 {
					return newInvalidArgCountError("replace", len(args), 3, "")
				}
				arg0, ok := args[0].(*Stringo)
				if !ok {
					return newPositionalTypeError("replace", 1, STRING_OBJ, args[0].Type())
				}
				arg1, ok := args[1].(*Stringo)
				if !ok {
					return newPositionalTypeError("replace", 2, STRING_OBJ, args[1].Type())
				}
				arg2, ok := args[2].(*Stringo)
				if !ok {
					return newPositionalTypeError("replace", 3, STRING_OBJ, args[2].Type())
				}
				replacedString := strings.ReplaceAll(arg0.Value, arg1.Value, arg2.Value)
				return &Stringo{Value: replacedString}
			},
			HelpStr: helpStrArgs{
				explanation: "`replace` will return a STRING with all occurrences of the given replacer STRING replaced by the next given STRING",
				signature:   "replace(arg: str, replacer: str, replaced: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "replace('Hello', 'l', 'X') => 'HeXXo'",
			}.String(),
		},
	},
	{
		Name: "_replace_regex",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 3 {
					return newInvalidArgCountError("replace_regex", len(args), 3, "")
				}
				arg0, ok := args[0].(*Stringo)
				if !ok {
					return newPositionalTypeError("replace_regex", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != STRING_OBJ && args[1].Type() != REGEX_OBJ {
					return newPositionalTypeError("replace_regex", 2, STRING_OBJ+" or REGEX", args[1].Type())
				}
				var re *regexp.Regexp
				if args[1].Type() == STRING_OBJ {
					re1, err := regexp.Compile(args[1].(*Stringo).Value)
					if err != nil {
						return newError("`replace_regex` error: %s", err.Error())
					}
					re = re1
				} else {
					re = args[1].(*Regex).Value
				}
				arg2, ok := args[2].(*Stringo)
				if !ok {
					return newPositionalTypeError("replace_regex", 3, STRING_OBJ, args[2].Type())
				}
				return &Stringo{Value: string(re.ReplaceAll([]byte(arg0.Value), []byte(arg2.Value)))}
			},
			HelpStr: helpStrArgs{
				explanation: "`replace_regex` will return a STRING with all occurrences of the given replacer REGEX STRING replaced by the next given STRING",
				signature:   "replace_regex(arg: str, replacer: str, replaced: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "replace_regex('Hello', 'l', 'X') => 'HeXXo'",
			}.String(),
		},
	},
	{
		Name: "strip",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 && len(args) != 2 {
					return newInvalidArgCountError("strip", len(args), 1, "or 2")
				}
				if len(args) == 1 {
					arg0, ok := args[0].(*Stringo)
					if !ok {
						return newPositionalTypeError("strip", 1, STRING_OBJ, args[0].Type())
					}
					str := strings.TrimSpace(arg0.Value)
					return &Stringo{Value: str}
				}
				if len(args) == 2 {
					arg0, ok := args[0].(*Stringo)
					if !ok {
						return newPositionalTypeError("strip", 1, STRING_OBJ, args[0].Type())
					}
					arg1, ok := args[1].(*Stringo)
					if !ok {
						return newPositionalTypeError("strip", 2, STRING_OBJ, args[1].Type())
					}
					str := strings.Trim(arg0.Value, arg1.Value)
					return &Stringo{Value: str}
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`strip` will return a STRING with the given cutset STRING removed",
				signature:   "strip(arg: str, cutset: str='') -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "strip(' Hello ') => 'Hello'",
			}.String(),
		},
	},
	{
		Name: "lstrip",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 && len(args) != 2 {
					return newInvalidArgCountError("lstrip", len(args), 1, "or 2")
				}
				if len(args) == 1 {
					arg0, ok := args[0].(*Stringo)
					if !ok {
						return newPositionalTypeError("lstrip", 1, STRING_OBJ, args[0].Type())
					}
					str := strings.TrimLeft(arg0.Value, " ")
					return &Stringo{Value: str}
				}
				if len(args) == 2 {
					arg0, ok := args[0].(*Stringo)
					if !ok {
						return newPositionalTypeError("lstrip", 1, STRING_OBJ, args[0].Type())
					}
					arg1, ok := args[1].(*Stringo)
					if !ok {
						return newPositionalTypeError("lstrip", 2, STRING_OBJ, args[1].Type())
					}
					str := strings.TrimLeft(arg0.Value, arg1.Value)
					return &Stringo{Value: str}
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`lstrip` returns a STRING with the given character stripped from the left side",
				signature:   "lstrip(s: str, chr: str=' ') -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "lstrip(' Hello') => 'Hello'",
			}.String(),
		},
	},
	{
		Name: "rstrip",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 && len(args) != 2 {
					return newInvalidArgCountError("rstrip", len(args), 1, "or 2")
				}
				if len(args) == 1 {
					arg0, ok := args[0].(*Stringo)
					if !ok {
						return newPositionalTypeError("rstrip", 1, STRING_OBJ, args[0].Type())
					}
					str := strings.TrimRight(arg0.Value, " ")
					return &Stringo{Value: str}
				}
				if len(args) == 2 {
					arg0, ok := args[0].(*Stringo)
					if !ok {
						return newPositionalTypeError("rstrip", 1, STRING_OBJ, args[0].Type())
					}
					arg1, ok := args[1].(*Stringo)
					if !ok {
						return newPositionalTypeError("rstrip", 2, STRING_OBJ, args[1].Type())
					}
					str := strings.TrimRight(arg0.Value, arg1.Value)
					return &Stringo{Value: str}
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`rstrip` returns a STRING with the given character stripped from the right side",
				signature:   "rstrip(s: str, chr: str=' ') -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "rstrip('Hello ') => 'Hello'",
			}.String(),
		},
	},
	{
		Name: "to_json",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_json", len(args), 1, "")
				}
				if isError(args[0]) {
					return args[0]
				}
				return blueObjToJsonObject(args[0])
			},
			HelpStr: helpStrArgs{
				explanation: "`to_json` will return the JSON STRING of the given MAP",
				signature:   "to_json(arg: map[str:any]) -> str",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "to_json({'x': 123}) => '{\"x\":123}'",
			}.String(),
		},
	},
	{
		Name: "to_upper",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_upper", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("to_upper", 1, STRING_OBJ, args[0].Type())
				}
				s := args[0].(*Stringo).Value
				return &Stringo{Value: strings.ToUpper(s)}
			},
			HelpStr: helpStrArgs{
				explanation: "`to_upper` returns the uppercase version of the given STRING",
				signature:   "to_upper(arg: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "to_upper('Hello') => 'HELLO'",
			}.String(),
		},
	},
	{
		Name: "to_lower",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_lower", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("to_lower", 1, STRING_OBJ, args[0].Type())
				}
				s := args[0].(*Stringo).Value
				return &Stringo{Value: strings.ToLower(s)}
			},
			HelpStr: helpStrArgs{
				explanation: "`to_lower` returns the lowercase version of the given STRING",
				signature:   "to_lower(arg: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "to_lower('Hello') => 'hello'",
			}.String(),
		},
	},
	{
		Name: "join",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("join", len(args), 2, "")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("join", 1, LIST_OBJ, args[0].Type())
				}
				if args[1].Type() != STRING_OBJ {
					return newPositionalTypeError("join", 2, STRING_OBJ, args[1].Type())
				}
				joiner := args[1].(*Stringo).Value
				elements := args[0].(*List).Elements
				strsToJoin := make([]string, len(elements))
				for i, e := range elements {
					if e.Type() != STRING_OBJ {
						return newError("found a type that was not a STRING in `join`. found=%s", e.Type())
					}
					strsToJoin[i] = e.(*Stringo).Value
				}
				return &Stringo{Value: strings.Join(strsToJoin, joiner)}
			},
			HelpStr: helpStrArgs{
				explanation: "`join` returns a STRING joined by the given joiner STRING for a list of STRINGs",
				signature:   "join(arg: list[str], joiner: str) -> str",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "join(['H','e','l','l','o'], ' ') => 'H e l l o'",
			}.String(),
		},
	},
	{
		Name: "_substr",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 3 {
					return newInvalidArgCountError("substr", len(args), 3, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("substr", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("substr", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != INTEGER_OBJ {
					return newPositionalTypeError("substr", 3, INTEGER_OBJ, args[2].Type())
				}
				s := args[0].(*Stringo).Value
				start := args[1].(*Integer).Value
				end := args[2].(*Integer).Value
				if end == -1 {
					end = int64(len(s))
				}
				return &Stringo{Value: s[start:end]}
			},
			HelpStr: helpStrArgs{
				explanation: "`substr` returns the STRING from start INTEGER to end INTEGER",
				signature:   "substr(arg: str, start: int=0, end: int=-1) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "substr('Hello', 1, 3) => 'el'",
			}.String(),
		},
	},
	{
		Name: "index_of",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("index_of", len(args), 2, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("index_of", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != STRING_OBJ {
					return newPositionalTypeError("index_of", 2, STRING_OBJ, args[1].Type())
				}
				s := args[0].(*Stringo).Value
				subs := args[1].(*Stringo).Value
				return &Integer{Value: int64(strings.Index(s, subs))}
			},
			HelpStr: helpStrArgs{
				explanation: "`index_of` returns the INTEGER index of the given STRING substring",
				signature:   "index_of(arg: str, substr: str) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "index_of('Hello', 'ell') => 1",
			}.String(),
		},
	},
	{
		Name: "_center",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 3 {
					return newInvalidArgCountError("center", len(args), 3, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("center", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("center", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != STRING_OBJ {
					return newPositionalTypeError("center", 3, STRING_OBJ, args[2].Type())
				}
				s := args[0].(*Stringo).Value
				length := int(args[1].(*Integer).Value)
				pad := args[2].(*Stringo).Value
				return &Stringo{Value: xstrings.Center(s, length, pad)}
			},
			HelpStr: helpStrArgs{
				explanation: "`center` returns a STRING centered given the length and pad character",
				signature:   "center(s: str, length: int, pad: str=' ') -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "center('Hello', 11) => '   Hello   '",
			}.String(),
		},
	},
	{
		Name: "_ljust",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 3 {
					return newInvalidArgCountError("ljust", len(args), 3, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("ljust", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("ljust", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != STRING_OBJ {
					return newPositionalTypeError("ljust", 3, STRING_OBJ, args[2].Type())
				}
				s := args[0].(*Stringo).Value
				length := int(args[1].(*Integer).Value)
				pad := args[2].(*Stringo).Value
				return &Stringo{Value: xstrings.LeftJustify(s, length, pad)}
			},
			HelpStr: helpStrArgs{
				explanation: "`ljust` returns a STRING left justified given the length and pad character",
				signature:   "ljust(s: str, length: int, pad: str=' ') -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "ljust('Hello', 10) => 'Hello     '",
			}.String(),
		},
	},
	{
		Name: "_rjust",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 3 {
					return newInvalidArgCountError("rjust", len(args), 3, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("rjust", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("rjust", 2, INTEGER_OBJ, args[1].Type())
				}
				if args[2].Type() != STRING_OBJ {
					return newPositionalTypeError("rjust", 3, STRING_OBJ, args[2].Type())
				}
				s := args[0].(*Stringo).Value
				length := int(args[1].(*Integer).Value)
				pad := args[2].(*Stringo).Value
				return &Stringo{Value: xstrings.RightJustify(s, length, pad)}
			},
			HelpStr: helpStrArgs{
				explanation: "`rjust` returns a STRING right justified given the length and pad character",
				signature:   "rjust(s: str, length: int, pad: str=' ') -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "rjust('Hello', 10) => '     Hello'",
			}.String(),
		},
	},
	{
		Name: "to_title",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_title", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("to_title", 1, STRING_OBJ, args[0].Type())
				}
				s := args[0].(*Stringo).Value
				titleS := util.ToTitleCase(s)
				return &Stringo{Value: titleS}
			},
			HelpStr: helpStrArgs{
				explanation: "`to_title` returns a STRING title cased",
				signature:   "to_title(s: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "to_title('hello world') => 'Hello World'",
			}.String(),
		},
	},
	{
		Name: "to_kebab",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_kebab", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("to_kebab", 1, STRING_OBJ, args[0].Type())
				}
				s := args[0].(*Stringo).Value
				return &Stringo{Value: xstrings.ToKebabCase(s)}
			},
			HelpStr: helpStrArgs{
				explanation: "`to_kebab` returns a STRING kebab cased",
				signature:   "to_kebab(s: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "to_kebab('hello world') => 'hello-world'",
			}.String(),
		},
	},
	{
		Name: "to_camel",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_camel", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("to_camel", 1, STRING_OBJ, args[0].Type())
				}
				s := args[0].(*Stringo).Value
				return &Stringo{Value: xstrings.ToCamelCase(s)}
			},
			HelpStr: helpStrArgs{
				explanation: "`to_camel` returns a STRING camel cased",
				signature:   "to_camel(s: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "to_camel('hello world') => 'HelloWorld'",
			}.String(),
		},
	},
	{
		Name: "to_snake",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("to_snake", len(args), 1, "")
				}
				if args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("to_snake", 1, STRING_OBJ, args[0].Type())
				}
				s := args[0].(*Stringo).Value
				return &Stringo{Value: xstrings.ToSnakeCase(s)}
			},
			HelpStr: helpStrArgs{
				explanation: "`to_snake` returns a STRING snake cased",
				signature:   "to_snake(s: str) -> str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "to_snake('hello world') => 'hello_world'",
			}.String(),
		},
	},
	{
		Name: "matches",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("matches", len(args), 2, "")
				}
				if args[0].Type() != STRING_OBJ && args[0].Type() != REGEX_OBJ {
					return newPositionalTypeError("matches", 1, STRING_OBJ+" or REGEX", args[0].Type())
				}
				if args[0].Type() == REGEX_OBJ {
					if args[1].Type() != STRING_OBJ {
						return newPositionalTypeError("matches", 2, STRING_OBJ, args[1].Type())
					}
					re := args[0].(*Regex).Value
					str := args[1].(*Stringo).Value
					return nativeToBooleanObject(re.MatchString(str))
				}
				// TODO: Support inverted arg as well? Like regex on left and string on right
				arg0, ok := args[0].(*Stringo)
				if !ok {
					return newPositionalTypeError("matches", 1, STRING_OBJ, args[0].Type())
				}
				if args[1].Type() == STRING_OBJ {
					arg1, ok := args[1].(*Stringo)
					if !ok {
						return newPositionalTypeError("matches", 2, STRING_OBJ, args[1].Type())
					}
					re, err := regexp.Compile(arg1.Value)
					if err != nil {
						return newError("`matches` error: %s", err.Error())
					}
					return nativeToBooleanObject(re.MatchString(arg0.Value))
				}
				if args[1].Type() != REGEX_OBJ {
					return newPositionalTypeError("matches", 2, REGEX_OBJ, args[1].Type())
				}
				re := args[1].(*Regex).Value
				return nativeToBooleanObject(re.MatchString(arg0.Value))
			},
			HelpStr: helpStrArgs{
				explanation: "`matches` returns true if the regex matches the string (on the left or right)",
				signature:   "matches(arg0: str|regex, arg1: str|regex) -> bool",
				errors:      "InvalidArgCount,PositionalType",
				example:     "matches('hello', re('hello')) => true  ||  matches(/hello/, 'hello') => true",
			}.String(),
		},
	},
}

var AllBuiltins = []struct {
	Name     string
	Builtins NewBuiltinSliceType
}{
	{Name: "core", Builtins: Builtins},
	{Name: "http", Builtins: HttpBuiltins},
	{Name: "time", Builtins: TimeBuiltins},
	{Name: "search", Builtins: SearchBuiltins},
	{Name: "db", Builtins: DbBuiltins},
	{Name: "math", Builtins: MathBuiltins},
	{Name: "config", Builtins: ConfigBuiltins},
	{Name: "crypto", Builtins: CryptoBuiltins},
	{Name: "net", Builtins: NetBuiltins},
	{Name: "color", Builtins: ColorBuiltins},
	{Name: "csv", Builtins: CsvBuiltins},
	{Name: "psutil", Builtins: PsutilBuiltins},
	{Name: "wasm", Builtins: WazmBuiltins},
	{Name: "crypto", Builtins: CryptoBuiltins},
	{Name: "ui", Builtins: UiBuiltins},
	{Name: "gg", Builtins: GgBuiltins},
}

func GetIndexAndBuiltinsOf(name string) (int, NewBuiltinSliceType) {
	for i, builtins := range AllBuiltins {
		if builtins.Name == name {
			return i, builtins.Builtins
		}
	}
	return -1, nil
}

func getBasicObjectForGoObj[T any](arg Object) (string, T, bool) {
	var zero T
	if arg == nil {
		return "", zero, false
	}
	if arg.Type() != MAP_OBJ {
		return "", zero, false
	}
	objPairs := arg.(*Map).Pairs
	if objPairs.Len() != 2 {
		return "", zero, false
	}
	// Get the 't' value
	hk1 := objPairs.Keys[0]
	mp1, ok := objPairs.Get(hk1)
	if !ok {
		return "", zero, false
	}
	if mp1.Key.Type() != STRING_OBJ {
		return "", zero, false
	}
	if mp1.Value.Type() != STRING_OBJ {
		return "", zero, false
	}
	if mp1.Key.(*Stringo).Value != "t" {
		return "", zero, false
	}
	t := mp1.Value.(*Stringo).Value
	// Get the 'v' value
	hk2 := objPairs.Keys[1]
	mp2, ok := objPairs.Get(hk2)
	if !ok {
		return "", zero, false
	}
	if mp2.Key.Type() != STRING_OBJ {
		return "", zero, false
	}
	if mp2.Value.Type() != GO_OBJ {
		return "", zero, false
	}
	if mp2.Key.(*Stringo).Value != "v" {
		return "", zero, false
	}
	v := mp2.Value.(*GoObj[T]).Value
	return t, v, true
}

type helpStrArgs struct {
	explanation string
	signature   string
	errors      string
	example     string
}

func (hsa helpStrArgs) String() string {
	return fmt.Sprintf("%s\n    Signature:  %s\n    Error(s):   %s\n    Example(s): %s\n", hsa.explanation, hsa.signature, hsa.errors, hsa.example)
}

func getBuiltinMap(bt BuiltinType) NewBuiltinSliceType {
	switch bt {
	case BuiltinBaseType:
		return Builtins
	case BuiltinHttpType:
		return HttpBuiltins
	case BuiltinTimeType:
		return TimeBuiltins
	case BuiltinSearchType:
		return SearchBuiltins
	case BuiltinDbType:
		return DbBuiltins
	case BuiltinMathType:
		return MathBuiltins
	case BuiltinConfigType:
		return ConfigBuiltins
	case BuiltinCryptoType:
		return CryptoBuiltins
	case BuiltinNetType:
		return NetBuiltins
	case BuiltinColorType:
		return ColorBuiltins
	case BuiltinCsvType:
		return CsvBuiltins
	case BuiltinPsutilType:
		return PsutilBuiltins
	case BuiltinWasmType:
		return WazmBuiltins
	case BuiltinUiType:
		return UiBuiltins
	case BuiltinGgType:
		return GgBuiltins
	}
	log.Fatalf("Unsupported Builtin Type: %s", bt)
	return nil
}

func GetBuiltinByName(bt BuiltinType, name string) *Builtin {
	for _, def := range getBuiltinMap(bt) {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}
