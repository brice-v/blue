package evaluator

import (
	"blue/consts"
	"blue/evaluator/pubsub"
	"blue/object"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	"github.com/gobuffalo/plush"
	"github.com/google/uuid"
	"github.com/gookit/color"
	clone "github.com/huandu/go-clone"
	"github.com/huandu/xstrings"
	"github.com/shopspring/decimal"
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

type BuiltinMapTypeInternal map[string]*object.Builtin

var builtins = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"help": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("help", len(args), 1, "")
			}
			return &object.Stringo{Value: args[0].Help()}
		},
		HelpStr: helpStrArgs{
			explanation: "`help` returns the help STRING for a given OBJECT",
			signature:   "help(arg: any) -> str",
			errors:      "None",
			example:     "help(print) => `prints help string for print builtin function`",
		}.String(),
	},
	"new": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("new", len(args), 1, "")
			}
			if !object.IsCollectionType(args[0].Type()) {
				return newPositionalTypeError("new", 1, "MAP or LIST or SET", args[0].Type())
			}
			if args[0].Type() == object.MAP_OBJ {
				m := args[0].(*object.Map)
				return clone.Clone(m).(*object.Map)
			} else if args[0].Type() == object.LIST_OBJ {
				l := args[0].(*object.List)
				return clone.Clone(l).(*object.List)
			} else {
				s := args[0].(*object.Set)
				return clone.Clone(s).(*object.Set)
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`new` returns a cloned MAP object from the given arg",
			signature:   "new(arg: map|list|set) -> map|list|set",
			errors:      "InvalidArgCount,PositionalType",
			example:     "new({'x': 1}) => {'x': 1}",
		}.String(),
	},
	"keys": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("keys", len(args), 1, "")
			}
			if args[0].Type() != object.MAP_OBJ {
				return newPositionalTypeError("keys", 1, object.MAP_OBJ, args[0].Type())
			}
			returnList := &object.List{
				Elements: []object.Object{},
			}
			m := args[0].(*object.Map)
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
	"values": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("values", len(args), 1, "")
			}
			if args[0].Type() != object.MAP_OBJ {
				return newPositionalTypeError("values", 1, object.MAP_OBJ, args[0].Type())
			}
			returnList := &object.List{
				Elements: []object.Object{},
			}
			m := args[0].(*object.Map)
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
	"del": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("del", len(args), 2, "")
			}
			if !object.IsCollectionType(args[0].Type()) {
				return newPositionalTypeError("del", 1, "MAP or LIST or SET", args[0].Type())
			}
			if args[0].Type() == object.MAP_OBJ {
				m := args[0].(*object.Map)
				hk := object.HashKey{
					Type:  args[1].Type(),
					Value: object.HashObject(args[1]),
				}
				m.Pairs.Delete(hk)
			} else if args[1].Type() == object.LIST_OBJ {
				if args[1].Type() != object.INTEGER_OBJ {
					return newPositionalTypeError("del", 2, object.LIST_OBJ, args[1].Type())
				}
				l := args[0].(*object.List)
				index := args[1].(*object.Integer).Value
				l.Elements = append(l.Elements[:index], l.Elements[index+1:]...)
			} else {
				hk := object.HashKey{
					Type:  args[1].Type(),
					Value: object.HashObject(args[1]),
				}
				s := args[0].(*object.Set)
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
	"len": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("len", len(args), 1, "")
			}

			switch arg := args[0].(type) {
			case *object.Stringo:
				return &object.Integer{Value: int64(runeLen(arg.Value))}
			case *object.List:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.Map:
				return &object.Integer{Value: int64(arg.Pairs.Len())}
			case *object.Set:
				return &object.Integer{Value: int64(arg.Elements.Len())}
			case *object.Bytes:
				return &object.Integer{Value: int64(len(arg.Value))}
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
	"append": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newInvalidArgCountError("append", len(args), 2, " or more")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("append", 1, object.LIST_OBJ, args[0].Type())
			}
			l := args[0].(*object.List)
			length := len(l.Elements)
			args = args[1:]
			argsLength := len(args)
			newElements := make([]object.Object, length+argsLength)
			copy(newElements, l.Elements)
			copy(newElements[length:length+argsLength], args)
			return &object.List{Elements: newElements}
		},
		HelpStr: helpStrArgs{
			explanation: "`append` returns the LIST of elements with given arg OBJECT at the end",
			signature:   "append(arg0: list, args...: any) -> list[any]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "append([1,2,3], 1) => [1,2,3,4]",
		}.String(),
	},
	"prepend": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newInvalidArgCountError("prepend", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("prepend", 1, object.LIST_OBJ, args[0].Type())
			}
			l := args[0].(*object.List)
			length := len(l.Elements)
			args = args[1:]
			argsLength := len(args)
			newElements := make([]object.Object, length+argsLength)
			copy(newElements, args)
			copy(newElements[argsLength:argsLength+length], l.Elements)
			return &object.List{Elements: newElements}
		},
		HelpStr: helpStrArgs{
			explanation: "`prepend` returns the LIST of elements with given arg OBJECT at the front",
			signature:   "prepend(arg0: list, args...: any) -> list[any]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "prepend([1,2,3], 4) => [4,1,2,3]",
		}.String(),
	},
	"push": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newInvalidArgCountError("push", len(args), 2, " or more")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("push", 1, object.LIST_OBJ, args[0].Type())
			}
			l := args[0].(*object.List)
			args = args[1:]
			l.Elements = append(l.Elements, args...)
			return &object.Integer{Value: int64(len(l.Elements))}
		},
		HelpStr: helpStrArgs{
			explanation: "`push` puts the given args at the end of the LIST and mutates it. The value returned is the length after pushing",
			signature:   "push(arg0: list[any], args...: any) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "push([1,2,3], 1) => 4",
		}.String(),
		Mutates: true,
	},
	"pop": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("pop", len(args), 1, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("pop", 1, object.LIST_OBJ, args[0].Type())
			}
			l := args[0].(*object.List)
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
	"unshift": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newInvalidArgCountError("unshift", len(args), 2, " or more")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("unshift", 1, object.LIST_OBJ, args[0].Type())
			}
			l := args[0].(*object.List)
			length := len(l.Elements)
			args = args[1:]
			argsLength := len(args)
			elems := make([]object.Object, argsLength+length)
			copy(elems, args)
			copy(elems[argsLength:argsLength+length], l.Elements)
			l.Elements = elems
			return &object.Integer{Value: int64(len(l.Elements))}
		},
		HelpStr: helpStrArgs{
			explanation: "`unshift` prepends the LIST with the given arguments and mutates it. The new length is returned",
			signature:   "unshift(arg0: list[any], args...: any) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "unshift([1,2,3], 1) => 4",
		}.String(),
		Mutates: true,
	},
	"shift": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("shift", len(args), 1, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("shift", 1, object.LIST_OBJ, args[0].Type())
			}
			l := args[0].(*object.List)
			if len(l.Elements) == 0 {
				return NULL
			}
			elem := l.Elements[0]
			if len(l.Elements) == 1 {
				l.Elements = []object.Object{}
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
	"concat": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) < 2 {
				return newInvalidArgCountError("concat", len(args), 2, " or more")
			}
			allElems := []object.Object{}
			for i, e := range args {
				if e.Type() != object.LIST_OBJ {
					return newPositionalTypeError("concat", i+1, object.LIST_OBJ, e.Type())
				}
				l := e.(*object.List)
				allElems = append(allElems, l.Elements...)
			}
			return &object.List{Elements: allElems}
		},
		HelpStr: helpStrArgs{
			explanation: "`concat` merges 2 or more LISTs together and returns the result",
			signature:   "concat(arg0: list[any], args...: list[any]) -> list[any]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "concat([1,2,3], [1]) => [1,2,3,1]",
		}.String(),
	},
	"reverse": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("reverse", len(args), 1, "")
			}
			if args[0].Type() != object.LIST_OBJ && args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("reverse", 1, object.LIST_OBJ+" or STRING", args[0].Type())
			}
			if args[0].Type() == object.STRING_OBJ {
				s := args[0].(*object.Stringo).Value
				return &object.Stringo{Value: xstrings.Reverse(s)}
			}
			l := args[0].(*object.List)
			length := len(l.Elements)
			newl := make([]object.Object, length)
			for i := 0; i < len(l.Elements); i++ {
				newl[length-i-1] = l.Elements[i]
			}
			return &object.List{Elements: newl}
		},
		HelpStr: helpStrArgs{
			explanation: "`reverse` reverse a string or list",
			signature:   "reverse(arg: list[any]|str) -> list[any]|str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "reverse([1,2,3]) => [3,2,1]",
		}.String(),
	},
	"println": {
		Fun: func(args ...object.Object) object.Object {
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
	"print": {
		Fun: func(args ...object.Object) object.Object {
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
	"input": {
		Fun: func(args ...object.Object) object.Object {
			argLen := len(args)
			if argLen == 0 {
				// read input with no prompt
				scanner := bufio.NewScanner(os.Stdin)
				if ok := scanner.Scan(); ok {
					return &object.Stringo{Value: scanner.Text()}
				}
				if err := scanner.Err(); err != nil {
					return newError("`input` error reading standard input: %s", err.Error())
				}
			} else if argLen == 1 {
				// read input with prompt
				if args[0].Type() != object.STRING_OBJ {
					return newPositionalTypeError("input", 1, object.STRING_OBJ, args[0].Type())
				}
				scanner := bufio.NewScanner(os.Stdin)
				fmt.Print(args[0].(*object.Stringo).Value)
				if ok := scanner.Scan(); ok {
					return &object.Stringo{Value: scanner.Text()}
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
	},
	// Default headers seem to be host, user-agent, accept-encoding (not case sensitive for these check pictures)
	// deno also used accept: */* (not sure what that is)
	"_fetch": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 5 {
				return newInvalidArgCountError("fetch", len(args), 5, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("fetch", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("fetch", 2, object.STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != object.MAP_OBJ {
				return newPositionalTypeError("fetch", 3, object.MAP_OBJ, args[2].Type())
			}
			if args[3].Type() != object.NULL_OBJ && args[3].Type() != object.STRING_OBJ {
				return newPositionalTypeError("fetch", 4, "NULL or STRING", args[3].Type())
			}
			if args[4].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("fetch", 5, object.BOOLEAN_OBJ, args[4].Type())
			}
			resource := args[0].(*object.Stringo).Value
			method := args[1].(*object.Stringo).Value
			headersMap := args[2].(*object.Map).Pairs
			isFullResp := args[4].(*object.Boolean).Value
			var body io.Reader
			if args[3].Type() == object.NULL_OBJ {
				body = nil
			} else {
				body = strings.NewReader(args[3].(*object.Stringo).Value)
			}
			request, err := http.NewRequest(method, resource, body)
			if err != nil {
				return newError("`fetch` error: %s", err.Error())
			}
			// Add User Agent always and then it can be overwritten
			request.Header.Add("user-agent", "blue/v"+consts.VERSION)
			for _, k := range headersMap.Keys {
				mp, _ := headersMap.Get(k)
				if key, ok := mp.Key.(*object.Stringo); ok {
					if val, ok := mp.Value.(*object.Stringo); ok {
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
			return &object.Stringo{Value: string(respBody)}
		},
		HelpStr: helpStrArgs{
			explanation: "`fetch` returns the body or full response of a network request",
			signature:   "fetch(resource: str, method: str('POST'|'PUT'|'PATCH'|'GET'|'HEAD'|'DELETE')='GET', headers: map[str]str, body: null|str|bytes, full_resp: bool)",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "fetch('https://danluu.com',full_resp=false) => <html>...</html>",
		}.String(),
	},
	"_read": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("read", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("read", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("read", 2, object.BOOLEAN_OBJ, args[1].Type())
			}
			fnameo := args[0].(*object.Stringo)
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
			if args[1].(*object.Boolean).Value {
				return &object.Bytes{Value: bs}
			}
			return &object.Stringo{Value: string(bs)}
		},
		HelpStr: helpStrArgs{
			explanation: "`_read` returns a STRING if given a FILE and false, or BYTES if given a file and true",
			signature:   "_read(arg: str, to_bytes=false) -> (to_bytes) ? bytes : str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "_read('/test.txt') => example file data",
		}.String(),
	},
	"_write": {
		Fun: func(args ...object.Object) object.Object {
			argLen := len(args)
			if argLen != 2 {
				return newInvalidArgCountError("write", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("write", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.BYTES_OBJ {
				return newPositionalTypeError("write", 2, "STRING or BYTES", args[1].Type())
			}
			fname := args[0].(*object.Stringo).Value
			var contents []byte
			if args[1].Type() == object.STRING_OBJ {
				contents = []byte(args[1].(*object.Stringo).Value)
			} else {
				contents = args[1].(*object.Bytes).Value
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
	"set": {
		// This is needed so we can return a set from a list of objects
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 && len(args) != 1 {
				return newInvalidArgCountError("set", len(args), 0, "or 1")
			}
			setMap := object.NewSetElements()
			resultSet := &object.Set{Elements: setMap}
			if len(args) == 0 {
				return resultSet
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("set", 1, object.LIST_OBJ, args[0].Type())
			}
			elements := args[0].(*object.List).Elements
			for _, e := range elements {
				hashKey := object.HashObject(e)
				resultSet.Elements.Set(hashKey, object.SetPair{Value: e, Present: struct{}{}})
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
	// This function is lossy
	"int": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("int", len(args), 1, "")
			}
			switch args[0].Type() {
			case object.FLOAT_OBJ:
				return &object.Integer{Value: int64(args[0].(*object.Float).Value)}
			case object.UINTEGER_OBJ:
				return &object.Integer{Value: int64(args[0].(*object.UInteger).Value)}
			case object.BIG_INTEGER_OBJ:
				return &object.Integer{Value: args[0].(*object.BigInteger).Value.Int64()}
			case object.BIG_FLOAT_OBJ:
				return &object.Integer{Value: args[0].(*object.BigFloat).Value.IntPart()}
			case object.INTEGER_OBJ:
				return args[0]
			case object.STRING_OBJ:
				s := args[0].(*object.Stringo).Value
				i, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					return newError("`int` error: %s", err.Error())
				}
				return &object.Integer{Value: i}
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
	// This function is lossy
	"float": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("float", len(args), 1, "")
			}
			switch args[0].Type() {
			case object.INTEGER_OBJ:
				return &object.Float{Value: float64(args[0].(*object.Integer).Value)}
			case object.UINTEGER_OBJ:
				return &object.Float{Value: float64(args[0].(*object.UInteger).Value)}
			case object.BIG_INTEGER_OBJ:
				return &object.Float{Value: float64(args[0].(*object.BigInteger).Value.Int64())}
			case object.BIG_FLOAT_OBJ:
				return &object.Float{Value: args[0].(*object.BigFloat).Value.InexactFloat64()}
			case object.FLOAT_OBJ:
				return args[0]
			case object.STRING_OBJ:
				s := args[0].(*object.Stringo).Value
				f, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return newError("`float` error: %s", err.Error())
				}
				return &object.Float{Value: f}
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
	// This function is lossy
	"bigint": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("bigint", len(args), 1, "")
			}
			switch args[0].Type() {
			case object.INTEGER_OBJ:
				return &object.BigInteger{Value: new(big.Int).SetInt64(args[0].(*object.Integer).Value)}
			case object.FLOAT_OBJ:
				return &object.BigInteger{Value: new(big.Int).SetInt64(int64(args[0].(*object.Float).Value))}
			case object.UINTEGER_OBJ:
				return &object.BigInteger{Value: new(big.Int).SetUint64(args[0].(*object.UInteger).Value)}
			case object.BIG_FLOAT_OBJ:
				return &object.BigInteger{Value: args[0].(*object.BigFloat).Value.BigInt()}
			case object.BIG_INTEGER_OBJ:
				return args[0]
			case object.STRING_OBJ:
				s := args[0].(*object.Stringo).Value
				b, ok := new(big.Int).SetString(s, 10)
				if !ok || b == nil {
					return newError("`bigint` error: %s is not a valid bigint", s)
				}
				return &object.BigInteger{Value: b}
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
	// This function is lossy
	"bigfloat": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("bigfloat", len(args), 1, "")
			}
			switch args[0].Type() {
			case object.INTEGER_OBJ:
				return &object.BigFloat{Value: decimal.NewFromInt(args[0].(*object.Integer).Value)}
			case object.FLOAT_OBJ:
				return &object.BigFloat{Value: decimal.NewFromFloat(args[0].(*object.Float).Value)}
			case object.UINTEGER_OBJ:
				return &object.BigFloat{Value: decimal.NewFromBigInt(new(big.Int).SetUint64(args[0].(*object.UInteger).Value), 0)}
			case object.BIG_INTEGER_OBJ:
				return &object.BigFloat{Value: decimal.NewFromBigInt(args[0].(*object.BigInteger).Value, 0)}
			case object.BIG_FLOAT_OBJ:
				return args[0]
			case object.STRING_OBJ:
				s := args[0].(*object.Stringo).Value
				bf, err := decimal.NewFromString(s)
				if err != nil {
					return newError("`bigfloat` error: %s", err.Error())
				}
				return &object.BigFloat{Value: bf}
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
	// This function is lossy
	"uint": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("uint", len(args), 1, "")
			}
			switch args[0].Type() {
			case object.INTEGER_OBJ:
				return &object.UInteger{Value: uint64(args[0].(*object.Integer).Value)}
			case object.FLOAT_OBJ:
				return &object.UInteger{Value: uint64(args[0].(*object.Float).Value)}
			case object.BIG_INTEGER_OBJ:
				return &object.UInteger{Value: args[0].(*object.BigInteger).Value.Uint64()}
			case object.BIG_FLOAT_OBJ:
				return &object.UInteger{Value: args[0].(*object.BigFloat).Value.BigInt().Uint64()}
			case object.UINTEGER_OBJ:
				return args[0]
			case object.STRING_OBJ:
				s := args[0].(*object.Stringo).Value
				u, err := strconv.ParseUint(s, 10, 64)
				if err != nil {
					return newError("`uint` error: %s", err.Error())
				}
				return &object.UInteger{Value: u}
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
	"eval_template": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("eval_template", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("eval_template", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.MAP_OBJ {
				return newPositionalTypeError("eval_template", 2, object.MAP_OBJ, args[1].Type())
			}
			m := args[1].(*object.Map)
			ctx := plush.NewContext()
			for _, k := range m.Pairs.Keys {
				mp, _ := m.Pairs.Get(k)
				if mp.Key.Type() != object.STRING_OBJ {
					return newError("`eval_template` error: found key in MAP that was not STRING. got=%s", mp.Key.Type())
				}
				val, err := blueObjectToGoObject(mp.Value)
				if err != nil {
					return newError("`eval_template` error: %s", err.Error())
				}
				ctx.Set(mp.Key.(*object.Stringo).Value, val)
			}
			inputStr := args[0].(*object.Stringo).Value
			s, err := plush.Render(inputStr, ctx)
			if err != nil {
				return newError("`eval_template` error: %s", err.Error())
			}
			return &object.Stringo{Value: s}
		},
		HelpStr: helpStrArgs{
			explanation: "`eval_template` returns the STRING version of a template parsed with plush (https://github.com/gobuffalo/plush)",
			signature:   "eval_template(tmplStr: str, tmplMap: map) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "eval_template('<%= arg %>', {'arg': 123}) => '123'",
		}.String(),
	},
	"error": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("error", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("error", 1, object.STRING_OBJ, args[0].Type())
			}
			msg, ok := args[0].(*object.Stringo)
			if !ok {
				return newError("`error` argument 1 was not STRING. got=%T", args[0])
			}
			return &object.Error{Message: msg.Value}
		},
		HelpStr: helpStrArgs{
			explanation: "`error` returns an EvaluatorError for the given STRING",
			signature:   "error(arg: str) -> EvaluatorError: #{arg}",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "error('fail') => ERROR| EvaluatorError: fail",
		}.String(),
	},
	"assert": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("assert", len(args), 1, "or 2")
			}
			if args[0].Type() != object.BOOLEAN_OBJ {
				return newPositionalTypeError("assert", 1, object.BOOLEAN_OBJ, args[0].Type())
			}
			b, ok := args[0].(*object.Boolean)
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
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("assert", 2, object.STRING_OBJ, args[1].Type())
			}
			msg, ok := args[1].(*object.Stringo)
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
	"type": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("type", len(args), 1, "")
			}
			return &object.Stringo{Value: string(args[0].Type())}
		},
		HelpStr: helpStrArgs{
			explanation: "`type` returns the STRING type representation of the given arg",
			signature:   "type(arg: any) -> str",
			errors:      "InvalidArgCount",
			example:     "type('Hello') => 'STRING'",
		}.String(),
	},
	"exec": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("exec", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("exec", 1, object.STRING_OBJ, args[0].Type())
			}
			return ExecStringCommand(args[0].(*object.Stringo).Value)
		},
		HelpStr: helpStrArgs{
			explanation: "`exec` returns a STRING from the executed command",
			signature:   "exec(command: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "exec('echo hello') => hello\\n",
		}.String(),
	},
	"is_alive": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_alive", len(args), 1, "")
			}
			if args[0].Type() != object.PROCESS_OBJ {
				return newPositionalTypeError("is_alive", 1, object.PROCESS_OBJ, args[0].Type())
			}
			p := args[0].(*object.Process)
			_, isAlive := ProcessMap.Load(pk(p.NodeName, p.Id))
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
	"exit": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) == 0 {
				os.Exit(0)
				// Unreachable
				return NULL
			} else if len(args) == 1 {
				if args[0].Type() != object.INTEGER_OBJ {
					return newError("argument passed to `exit` must be INTEGER. got=%s", args[0].Type())
				}
				os.Exit(int(args[0].(*object.Integer).Value))
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
	"cwd": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) > 0 {
				return newInvalidArgCountError("cwd", len(args), 0, "")
			}
			dir, err := os.Getwd()
			if err != nil {
				return newError("`cwd` error: %s", err.Error())
			}
			return &object.Stringo{Value: dir}
		},
		HelpStr: helpStrArgs{
			explanation: "`cwd` returns the STRING path of the current working directory",
			signature:   "cwd() -> str",
			errors:      "InvalidArgCount,CustomError",
			example:     "cwd() => '/home/user/...'",
		}.String(),
	},
	"cd": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("cd", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("cd", 1, object.STRING_OBJ, args[0].Type())
			}
			path := args[0].(*object.Stringo).Value
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
	"_to_bytes": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_bytes", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("to_bytes", 1, object.STRING_OBJ, args[0].Type())
			}
			return &object.Bytes{
				Value: []byte(args[0].(*object.Stringo).Value),
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`to_bytes` returns the BYTE representation of the given STRING",
			signature:   "to_bytes(arg: str) -> bytes",
			errors:      "InvalidArgCount,TypeNotSupported",
			example:     "to_bytes('hello') => []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}",
		}.String(),
	},
	"str": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("str", len(args), 1, "")
			}
			if args[0].Type() == object.BYTES_OBJ {
				return &object.Stringo{Value: string(args[0].(*object.Bytes).Value)}
			}
			return &object.Stringo{Value: args[0].Inspect()}
		},
		HelpStr: helpStrArgs{
			explanation: "`str` returns the STRING representation of the given BYTES or the inspected object",
			signature:   "str(arg: any) -> str",
			errors:      "InvalidArgCount",
			example:     "str([]byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}) => 'hello'",
		}.String(),
	},
	"is_file": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_file", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("is_file", 1, object.STRING_OBJ, args[0].Type())
			}
			fpath := args[0].(*object.Stringo).Value
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
	"is_dir": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_dir", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("is_dir", 1, object.STRING_OBJ, args[0].Type())
			}
			fpath := args[0].(*object.Stringo).Value
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
	"find_exe": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("find_exe", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("find_exe", 1, object.STRING_OBJ, args[0].Type())
			}
			exePath := args[0].(*object.Stringo).Value
			fname, err := exec.LookPath(exePath)
			if err == nil {
				fname, err = filepath.Abs(fname)
			}
			if err != nil {
				return newError("`find_exe` error: %s", err.Error())
			}
			return &object.Stringo{Value: fname}
		},
		HelpStr: helpStrArgs{
			explanation: "`find_exe` returns the STRING path of the given STRING executable name",
			signature:   "find_exe(exe_name: str) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "find_exe('blue') => /home/user/.blue/bin/blue",
		}.String(),
	},
	// TODO: Do we want to do that thing where we shell expand home dir? or other things like that?
	"rm": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("rm", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("rm", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
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
	// TODO: Do we want to do that thing where we shell expand home dir? or other things like that?
	"ls": {
		Fun: func(args ...object.Object) object.Object {
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
				if args[0].Type() != object.STRING_OBJ {
					return newPositionalTypeError("ls", 1, object.STRING_OBJ, args[0].Type())
				}
				cwd = args[0].(*object.Stringo).Value
			}
			fileOrDirs, err := os.ReadDir(cwd)
			if err != nil {
				return newError("`ls` error: %s", err.Error())
			}
			result := &object.List{Elements: make([]object.Object, len(fileOrDirs))}
			for i := 0; i < len(fileOrDirs); i++ {
				result.Elements[i] = &object.Stringo{Value: fileOrDirs[i].Name()}
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
	// TODO: Eventually we need to support files better (and possibly, stdin, stderr, stdout) and then http stuff
	"is_valid_json": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_valid_json", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("is_valid_json", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return nativeToBooleanObject(json.Valid([]byte(s)))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_valid_json` returns a BOOLEAN if the given STRING is valid json",
			signature:   "is_valid_json(json: str) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_valid_json('{}') => true",
		}.String(),
	},
	"wait": {
		Fun: func(args ...object.Object) object.Object {
			// This function will avoid returning errors
			// but that means random inputs will technically be allowed
			processesToWaitFor := []*object.Process{}
			for _, arg := range args {
				if processes, ok := getListOfProcesses(arg); ok {
					processesToWaitFor = append(processesToWaitFor, processes...)
					continue
				}
				if arg.Type() == object.PROCESS_OBJ {
					v := arg.(*object.Process)
					processesToWaitFor = append(processesToWaitFor, v)
				}
			}
			for {
				allDone := false
				for _, p := range processesToWaitFor {
					_, ok := ProcessMap.Load(pk(p.NodeName, p.Id))
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
	"_publish": {
		Fun: func(args ...object.Object) object.Object {
			// pubsub.publish('TOPIC', MSG) -> non-blocking send
			if len(args) != 2 {
				return newInvalidArgCountError("publish", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("publish", 1, object.STRING_OBJ, args[0].Type())
			}
			topic := args[0].(*object.Stringo).Value
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
	"_broadcast": {
		Fun: func(args ...object.Object) object.Object {
			// pubsub.broadcast(MSG) -> non-blocking send
			// pubsub.broadcast(MSG, ['some', 'specifc', 'channels']) -> non-blocking send
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("broadcast", len(args), 1, "or 2")
			}
			if len(args) == 2 && args[1].Type() != object.LIST_OBJ {
				return newPositionalTypeError("broadcast", 2, object.LIST_OBJ, args[1].Type())
			}
			if len(args) == 1 {
				PubSubBroker.BroadcastToAllTopics(args[0])
				return NULL
			}
			l := args[1].(*object.List).Elements
			topics := make([]string, len(l))
			for i, e := range l {
				if e.Type() != object.STRING_OBJ {
					return newError("`broadcast` error: all elements in list should be STRING. found=%s", e.Type())
				}
				topics[i] = e.(*object.Stringo).Value
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
	// Functions for subscribers in pubsub
	"add_topic": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("add_topic", len(args), 2, "")
			}
			if args[0].Type() != object.MAP_OBJ {
				return newPositionalTypeError("add_topic", 1, object.MAP_OBJ, args[0].Type())
			}
			t, sub, ok := getBasicObjectForGoObj[*pubsub.Subscriber](args[0])
			if t != "sub" {
				return newError("`add_topic` error: argument 1 should be in the format {t: 'sub', v: uint}")
			}
			if !ok {
				return newError("`add_topic` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*pubsub.Subscriber]}. got=%s", args[0].Inspect())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("add_topic", 2, object.STRING_OBJ, args[1].Type())
			}
			topic := args[1].(*object.Stringo).Value
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
	// TODO: add_topics, remove_topics?
	"remove_topic": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("remove_topic", len(args), 2, "")
			}
			if args[0].Type() != object.MAP_OBJ {
				return newPositionalTypeError("remove_topic", 1, object.MAP_OBJ, args[0].Type())
			}
			t, sub, ok := getBasicObjectForGoObj[*pubsub.Subscriber](args[0])
			if t != "sub" {
				return newError("`remove_topic` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*pubsub.Subscriber]}")
			}
			if !ok {
				return newError("`remove_topic` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*pubsub.Subscriber]}. got=%s", args[0].Inspect())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("remove_topic", 2, object.STRING_OBJ, args[1].Type())
			}
			topic := args[1].(*object.Stringo).Value
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
	"_subscribe": {
		Fun: func(args ...object.Object) object.Object {
			// pubsub.subscribe('TOPIC') -> {t: 'sub', v: _} -> _.recv()
			if len(args) != 1 {
				return newInvalidArgCountError("subscribe", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("subscribe", 1, object.STRING_OBJ, args[0].Type())
			}
			topic := args[0].(*object.Stringo).Value
			subId := subscriberCount.Add(1)
			sub := PubSubBroker.AddSubscriber(subId)
			PubSubBroker.Subscribe(sub, topic)
			return object.CreateBasicMapObjectForGoObj("sub", NewGoObj(sub))
		},
		HelpStr: helpStrArgs{
			explanation: "`subscribe` will add a subscriber to the pubsub broker for a topic",
			signature:   "subscribe(sub: str) -> {t: 'sub', v: _}",
			errors:      "InvalidArgCount,PositionalType",
			example:     "subscribe('TOPIC') => {t: 'sub', v: _} -> _.recv()",
		}.String(),
	},
	"unsubscribe": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("unsubscribe", len(args), 2, "")
			}
			if args[0].Type() != object.MAP_OBJ {
				return newPositionalTypeError("unsubscribe", 1, object.MAP_OBJ, args[0].Type())
			}
			t, sub, ok := getBasicObjectForGoObj[*pubsub.Subscriber](args[0])
			if t != "sub" {
				return newError("`unsubscribe` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*pubsub.Subscriber]}")
			}
			if !ok {
				return newError("`unsubscribe` error: argument 1 should be in the format {t: 'sub', v: GO_OBJ[*pubsub.Subscriber]}. got=%s", args[0].Inspect())
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
	"_pubsub_sub_listen": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("pubsub_sub_listen", len(args), 1, "")
			}
			if args[0].Type() != object.GO_OBJ {
				return newPositionalTypeError("pubsub_sub_listen", 1, object.GO_OBJ, args[0].Type())
			}
			sub, ok := args[0].(*object.GoObj[*pubsub.Subscriber])
			if !ok {
				return newPositionalTypeErrorForGoObj("pubsub_sub_listen", 1, "*pubsub.Subscriber", args[0])
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
	"_get_subscriber_count": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 && len(args) != 0 {
				return newInvalidArgCountError("get_subscriber_count", len(args), 0, "or 1")
			}
			if len(args) == 0 {
				// Get total count of subscribers
				return &object.Integer{Value: int64(PubSubBroker.GetTotalSubscribers())}
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("get_subscriber_count", 1, object.STRING_OBJ, args[0].Type())
			}
			topic := args[0].(*object.Stringo).Value
			return &object.Integer{Value: int64(PubSubBroker.GetNumSubscribersForTopic(topic))}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_subscriber_count` returns the number of subscribers for a topic, if there is no topic passed in the total subscribers are returned",
			signature:   "get_subscriber_count(arg: str|none) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "get_subscriber_count('TOPIC') => 1",
		}.String(),
	},
	"_kv_put": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("kv_put", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("kv_put", 1, object.STRING_OBJ, args[0].Type())
			}
			topic := args[0].(*object.Stringo).Value
			var m *object.Map
			m, ok := KVMap.Load(topic)
			if !ok {
				m = &object.Map{
					Pairs: object.NewPairsMap(),
				}
			}
			hashedKey := object.HashObject(args[1])
			hk := object.HashKey{Type: args[1].Type(), Value: hashedKey}
			m.Pairs.Set(hk, object.MapPair{
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
	"_kv_get": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("kv_get", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("kv_get", 1, object.STRING_OBJ, args[0].Type())
			}
			topic := args[0].(*object.Stringo).Value
			var m *object.Map
			m, ok := KVMap.Load(topic)
			if !ok {
				// Return NULL if the topic doesn't have a map that exists
				return NULL
			}
			hashedKey := object.HashObject(args[1])
			hk := object.HashKey{Type: args[1].Type(), Value: hashedKey}
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
	"_kv_delete": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("kv_delete", len(args), 1, "or 2")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("kv_delete", 1, object.STRING_OBJ, args[0].Type())
			}
			topic := args[0].(*object.Stringo).Value
			if len(args) == 1 {
				// If its 1 we want to delete a topic, and the associated map
				KVMap.Delete(topic)
				return NULL
			} else {
				// If its 2 we want to delete a key from a map on a topic
				var m *object.Map
				m, ok := KVMap.Load(topic)
				if !ok {
					// Return NULL if the topic doesn't have a map that exists
					// theres nothing to delete in this case
					return NULL
				}
				hashedKey := object.HashObject(args[1])
				hk := object.HashKey{Type: args[1].Type(), Value: hashedKey}
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
	"_new_uuid": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("new_uuid", len(args), 0, "")
			}
			return &object.Stringo{Value: uuid.NewString()}
		},
		HelpStr: helpStrArgs{
			explanation: "`new_uuid` returns a new random UUID STRING",
			signature:   "new_uuid() -> str",
			errors:      "InvalidArgCount",
			example:     "new_uuid() => 'a38dc5fa-7f18-4e1c-8a70-f8d343109708'",
		}.String(),
	},
	// This is straight out of golang's example for runtime/metrics https://pkg.go.dev/runtime/metrics
	"_go_metrics": {
		Fun: func(args ...object.Object) object.Object {
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
			return &object.Stringo{Value: out.String()}
		},
		HelpStr: helpStrArgs{
			explanation: "`go_metrics` returns the STRING version of the golang runtime metrics",
			signature:   "go_metrics() -> str",
			errors:      "InvalidArgCount",
			example:     "go_metrics() => str",
		}.String(),
	},
	"get_os": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_os", len(args), 0, "")
			}
			return &object.Stringo{Value: runtime.GOOS}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_os` returns the STRING GOOS of the runtime",
			signature:   "get_os() -> str",
			errors:      "InvalidArgCount",
			example:     "get_os() => windows",
		}.String(),
	},
	"get_arch": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_arch", len(args), 0, "")
			}
			return &object.Stringo{Value: runtime.GOARCH}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_arch` returns the STRING GOARCH of the runtime",
			signature:   "get_arch() -> str",
			errors:      "InvalidArgCount",
			example:     "get_arch() => amd64",
		}.String(),
	},
	"_gc": {
		Fun: func(args ...object.Object) object.Object {
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
	"_version": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("version", len(args), 0, "")
			}
			return &object.Stringo{Value: fmt.Sprintf("%s-%s", runtime.Version(), consts.VERSION)}
		},
		HelpStr: helpStrArgs{
			explanation: "`version` returns the golang version and blue version hyphenated",
			signature:   "version() -> str",
			errors:      "InvalidArgCount",
			example:     "version() => go1.21.5-0.1.16-684f398-windows/amd64",
		}.String(),
	},
	"_num_cpu": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("num_cpu", len(args), 0, "")
			}
			return &object.Integer{Value: int64(runtime.NumCPU())}
		},
		HelpStr: helpStrArgs{
			explanation: "`num_cpu` returns the number of cpus available to the blue process",
			signature:   "num_cpu() -> int",
			errors:      "InvalidArgCount",
			example:     "num_cpu() => 12",
		}.String(),
	},
	"_num_process": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("num_process", len(args), 0, "")
			}
			return &object.Integer{Value: int64(runtime.NumGoroutine())}
		},
		HelpStr: helpStrArgs{
			explanation: "`num_process` returns the number of processes used by the runtime",
			signature:   "num_process() -> int",
			errors:      "InvalidArgCount",
			example:     "num_process() => 6",
		}.String(),
	},
	"_num_max_cpu": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("num_max_cpu", len(args), 0, "")
			}
			return &object.Integer{Value: int64(runtime.GOMAXPROCS(-1))}
		},
		HelpStr: helpStrArgs{
			explanation: "`num_max_cpu` returns the max number of cpus available to the runtime",
			signature:   "num_max_cpu() -> int",
			errors:      "InvalidArgCount",
			example:     "num_max_cpu() => 12",
		}.String(),
	},
	"_num_os_thread": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("num_os_thread", len(args), 0, "")
			}
			return &object.Integer{Value: int64(pprof.Lookup("threadcreate").Count())}
		},
		HelpStr: helpStrArgs{
			explanation: "`num_os_thread` returns the number of os threads being used by the runtime",
			signature:   "num_os_thread() -> int",
			errors:      "InvalidArgCount",
			example:     "num_os_thread() => 15",
		}.String(),
	},
	"_set_max_cpu": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_max_cpu", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("set_max_cpu", 1, object.INTEGER_OBJ, args[0].Type())
			}
			i := int(args[0].(*object.Integer).Value)
			return &object.Integer{Value: int64(runtime.GOMAXPROCS(i))}
		},
		HelpStr: helpStrArgs{
			explanation: "`set_max_cpu` sets the max number of cpus for the runtime and returns the previous setting. if arg < 1 => defaults to current number of cpus",
			signature:   "set_max_cpu(arg: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_max_cpu(3) => 6",
		}.String(),
	},
	"_set_gc_percent": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_gc_percent", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("set_gc_percent", 1, object.INTEGER_OBJ, args[0].Type())
			}
			i := int(args[0].(*object.Integer).Value)
			return &object.Integer{Value: int64(debug.SetGCPercent(i))}
		},
		HelpStr: helpStrArgs{
			explanation: "`set_gc_percent` sets the gc target percentage and returns the previous setting. a lower setting essentially limits the memory, 100 is default, and negative numbers turn gc off",
			signature:   "set_gc_percent(arg: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_gc_percent(30) => 100",
		}.String(),
	},
	"_get_mem_stats": {
		Fun: func(args ...object.Object) object.Object {
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
			mapObj := object.NewOrderedMap[string, object.Object]()
			mapObj.Set("alloc", &object.Stringo{Value: formatBytes(s.Alloc)})
			mapObj.Set("total-alloc", &object.Stringo{Value: formatBytes(s.TotalAlloc)})
			mapObj.Set("sys", &object.Stringo{Value: formatBytes(s.Sys)})
			mapObj.Set("lookups", &object.UInteger{Value: s.Lookups})
			mapObj.Set("mallocs", &object.UInteger{Value: s.Mallocs})
			mapObj.Set("frees", &object.UInteger{Value: s.Frees})
			mapObj.Set("heap-alloc", &object.Stringo{Value: formatBytes(s.HeapAlloc)})
			mapObj.Set("heap-sys", &object.Stringo{Value: formatBytes(s.HeapSys)})
			mapObj.Set("heap-idle", &object.Stringo{Value: formatBytes(s.HeapIdle)})
			mapObj.Set("heap-in-use", &object.Stringo{Value: formatBytes(s.HeapInuse)})
			mapObj.Set("heap-released", &object.Stringo{Value: formatBytes(s.HeapReleased)})
			mapObj.Set("heap-objects", &object.UInteger{Value: s.HeapObjects})
			mapObj.Set("stack-in-use", &object.Stringo{Value: formatBytes(s.StackInuse)})
			mapObj.Set("stack-sys", &object.Stringo{Value: formatBytes(s.StackSys)})
			mapObj.Set("stack-mspan-inuse", &object.Stringo{Value: formatBytes(s.MSpanInuse)})
			mapObj.Set("stack-mspan-sys", &object.Stringo{Value: formatBytes(s.MSpanSys)})
			mapObj.Set("stack-mcache-inuse", &object.Stringo{Value: formatBytes(s.MCacheInuse)})
			mapObj.Set("stack-mcache-sys", &object.Stringo{Value: formatBytes(s.MCacheSys)})
			mapObj.Set("other-sys", &object.Stringo{Value: formatBytes(s.OtherSys)})
			mapObj.Set("gc-sys", &object.Stringo{Value: formatBytes(s.GCSys)})
			mapObj.Set("next-gc: when heap-alloc >=", &object.Stringo{Value: formatBytes(s.NextGC)})
			lastGC := "-"
			if s.LastGC != 0 {
				lastGC = fmt.Sprint(time.Unix(0, int64(s.LastGC)))
			}
			mapObj.Set("last-gc", &object.Stringo{Value: lastGC})
			mapObj.Set("gc-pause-total", &object.Stringo{Value: time.Duration(s.PauseTotalNs).String()})
			mapObj.Set("gc-pause", &object.UInteger{Value: s.PauseNs[(s.NumGC+255)%256]})
			mapObj.Set("gc-pause-end", &object.UInteger{Value: s.PauseEnd[(s.NumGC+255)%256]})
			mapObj.Set("num-gc", &object.UInteger{Value: uint64(s.NumGC)})
			mapObj.Set("num-forced-gc", &object.UInteger{Value: uint64(s.NumForcedGC)})
			mapObj.Set("gc-cpu-fraction", &object.Float{Value: s.GCCPUFraction})
			mapObj.Set("enable-gc", &object.Boolean{Value: s.EnableGC})
			mapObj.Set("debug-gc", &object.Boolean{Value: s.DebugGC})
			return object.CreateMapObjectForGoMap(*mapObj)
		},
		HelpStr: helpStrArgs{
			explanation: "`get_mem_stats` returns the runtime memory stats",
			signature:   "get_mem_stats() -> map[str]any",
			errors:      "InvalidArgCount",
			example:     "get_mem_stats() => object (all mem stats)",
		}.String(),
	},
	"_get_stack_trace": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_stack_trace", len(args), 0, "")
			}
			return &object.Stringo{Value: string(debug.Stack())}
		},
		HelpStr: helpStrArgs{
			explanation: "`get_stack_trace` returns the runtime current stack trace",
			signature:   "get_stack_trace() -> str",
			errors:      "InvalidArgCount",
			example:     "get_stack_trace() => str",
		}.String(),
	},
	"_free_os_mem": {
		Fun: func(args ...object.Object) object.Object {
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
	"_print_stack_trace": {
		Fun: func(args ...object.Object) object.Object {
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
	"_set_max_stack": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_max_stack", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("set_max_stack", 1, object.INTEGER_OBJ, args[0].Type())
			}
			i := int(args[0].(*object.Integer).Value)
			return &object.Integer{Value: int64(debug.SetMaxStack(i))}
		},
		HelpStr: helpStrArgs{
			explanation: "`set_max_stack` sets the max amount of memory that can be used by blue process, only limiting future stack sizes and returning previous setting",
			signature:   "set_max_stack(arg: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_max_stack(12*1024*1024) => 1073741824",
		}.String(),
	},
	"_set_max_threads": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_max_threads", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("set_max_threads", 1, object.INTEGER_OBJ, args[0].Type())
			}
			i := int(args[0].(*object.Integer).Value)
			return &object.Integer{Value: int64(debug.SetMaxThreads(i))}
		},
		HelpStr: helpStrArgs{
			explanation: "`set_max_threads` sets the max number of os threads the program can use and returns the previous setting",
			signature:   "set_max_threads(arg: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_max_threads(20) => 1000",
		}.String(),
	},
	"_set_mem_limit": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("set_mem_limit", len(args), 1, "")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("set_mem_limit", 1, object.INTEGER_OBJ, args[0].Type())
			}
			i := args[0].(*object.Integer).Value
			return &object.Integer{Value: debug.SetMemoryLimit(i)}
		},
		HelpStr: helpStrArgs{
			explanation: "`set_mem_limit` sets the soft max memory limit of the program, returning the previous setting, a negative limit retuns the current setting",
			signature:   "set_mem_limit(arg: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set_mem_limit(12*1024*1024*1024*1024) => 2**64-1",
		}.String(),
	},
	"re": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("re", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("re", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			re, err := regexp.Compile(s)
			if err != nil {
				return newError("`re` error: %s", err.Error())
			}
			return &object.Regex{Value: re}
		},
		HelpStr: helpStrArgs{
			explanation: "`re` returns a regex object for the given string",
			signature:   "re(arg: str) -> regex",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "re('abc') => r/abc/",
		}.String(),
	},
	"to_list": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_list", len(args), 1, "")
			}
			if args[0].Type() != object.SET_OBJ {
				return newPositionalTypeError("to_list", 1, object.SET_OBJ, args[0].Type())
			}
			s := args[0].(*object.Set).Elements
			newElems := []object.Object{}
			for _, k := range s.Keys {
				if obj, ok := s.Get(k); ok {
					newElems = append(newElems, obj.Value)
				}
			}
			return &object.List{Elements: newElems}
		},
		HelpStr: helpStrArgs{
			explanation: "`to_list` returns a list from the given set",
			signature:   "to_list(arg: set[any]) -> list[any]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "to_list({1,2,3}) => [1,2,3]",
		}.String(),
	},
	"abs_path": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("abs_path", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("abs_path", 1, object.STRING_OBJ, args[0].Type())
			}
			// TODO: Fix so this works with embedded files?
			fpath := args[0].(*object.Stringo).Value
			path, err := filepath.Abs(fpath)
			if err != nil {
				return newError("`abs_path` error: %s", err.Error())
			}
			return &object.Stringo{Value: path}
		},
		HelpStr: helpStrArgs{
			explanation: "`abs_path` returns the absolute path of the given filepath",
			signature:   "abs_path(arg: str) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "abs_path('some_file.txt') => '/the/path/to/some_file.txt'",
		}.String(),
	},
	"fmt": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("fmt", len(args), 2, "")
			}
			val, err := anyBlueObjectToGoObject(args[0])
			if err != nil {
				// Escape to just using value as string
				val = args[0].Inspect()
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("fmt", 2, object.STRING_OBJ, args[1].Type())
			}
			fmtString := args[1].(*object.Stringo).Value
			return &object.Stringo{Value: fmt.Sprintf(fmtString, val)}
		},
		HelpStr: helpStrArgs{
			explanation: "`fmt` returns the formatted version of the given INTEGER",
			signature:   "fmt(arg: int, fmtStr: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "fmt(3, '%04b') => '0011'",
		}.String(),
	},
	"save": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("save", len(args), 1, "")
			}
			bs, err := args[0].Encode()
			if err != nil {
				return newError("`save` error: %s", err.Error())
			}
			return &object.Bytes{Value: bs}
		},
		HelpStr: helpStrArgs{
			explanation: "`save` returns the bytes of the encoded object",
			signature:   "save(arg: any) -> bytes",
			errors:      "InvalidArgCount,CustomError",
			example:     "save(1234) => '82001904d2'",
		}.String(),
	},
	"__hash": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("__hash", len(args), 1, "")
			}
			return &object.UInteger{Value: object.HashObject(args[0])}
		},
		HelpStr: "__hash returns the internal hash of an object",
	},
})

func GetBuiltins(e *Evaluator) BuiltinMapType {
	b := builtins
	b.Put("to_num", createToNumBuiltin(e))
	b.Put("_sort", createSortBuiltin(e))
	b.Put("_sorted", createSortedBuiltin(e))
	b.Put("all", createAllBuiltin(e))
	b.Put("any", createAnyBuiltin(e))
	b.Put("map", createMapBuiltin(e))
	b.Put("filter", createFilterBuiltin(e))
	b.Put("load", createLoadBuiltin(e))
	return b
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

func getBasicObject(arg object.Object) (string, uint64, bool) {
	if arg == nil {
		return "", 0, false
	}
	if arg.Type() != object.MAP_OBJ {
		return "", 0, false
	}
	objPairs := arg.(*object.Map).Pairs
	if objPairs.Len() != 2 {
		return "", 0, false
	}
	// Get the 't' value
	hk1 := objPairs.Keys[0]
	mp1, ok := objPairs.Get(hk1)
	if !ok {
		return "", 0, false
	}
	if mp1.Key.Type() != object.STRING_OBJ {
		return "", 0, false
	}
	if mp1.Value.Type() != object.STRING_OBJ {
		return "", 0, false
	}
	if mp1.Key.(*object.Stringo).Value != "t" {
		return "", 0, false
	}
	t := mp1.Value.(*object.Stringo).Value
	// Get the 'v' value
	hk2 := objPairs.Keys[1]
	mp2, ok := objPairs.Get(hk2)
	if !ok {
		return "", 0, false
	}
	if mp2.Key.Type() != object.STRING_OBJ {
		return "", 0, false
	}
	if mp2.Value.Type() != object.UINTEGER_OBJ {
		return "", 0, false
	}
	if mp2.Key.(*object.Stringo).Value != "v" {
		return "", 0, false
	}
	v := mp2.Value.(*object.UInteger).Value
	return t, v, true
}

func getBasicObjectForGoObj[T any](arg object.Object) (string, T, bool) {
	var zero T
	if arg == nil {
		return "", zero, false
	}
	if arg.Type() != object.MAP_OBJ {
		return "", zero, false
	}
	objPairs := arg.(*object.Map).Pairs
	if objPairs.Len() != 2 {
		return "", zero, false
	}
	// Get the 't' value
	hk1 := objPairs.Keys[0]
	mp1, ok := objPairs.Get(hk1)
	if !ok {
		return "", zero, false
	}
	if mp1.Key.Type() != object.STRING_OBJ {
		return "", zero, false
	}
	if mp1.Value.Type() != object.STRING_OBJ {
		return "", zero, false
	}
	if mp1.Key.(*object.Stringo).Value != "t" {
		return "", zero, false
	}
	t := mp1.Value.(*object.Stringo).Value
	// Get the 'v' value
	hk2 := objPairs.Keys[1]
	mp2, ok := objPairs.Get(hk2)
	if !ok {
		return "", zero, false
	}
	if mp2.Key.Type() != object.STRING_OBJ {
		return "", zero, false
	}
	if mp2.Value.Type() != object.GO_OBJ {
		return "", zero, false
	}
	if mp2.Key.(*object.Stringo).Value != "v" {
		return "", zero, false
	}
	v := mp2.Value.(*object.GoObj[T]).Value
	return t, v, true
}

func getListOfProcesses(arg object.Object) ([]*object.Process, bool) {
	if arg.Type() != object.LIST_OBJ {
		return nil, false
	}
	elems := arg.(*object.List).Elements
	processes := make([]*object.Process, 0, len(elems))
	for _, e := range elems {
		if e.Type() != object.PROCESS_OBJ {
			return nil, false
		}
		v := e.(*object.Process)
		processes = append(processes, v)
	}
	return processes, true
}
