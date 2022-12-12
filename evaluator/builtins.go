package evaluator

import (
	"blue/object"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	clone "github.com/huandu/go-clone"
)

type BuiltinMapTypeInternal map[string]*object.Builtin

var builtins = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"new": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`new` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.MAP_OBJ {
				return newError("argument 1 to `new` should be MAP. got=%s", args[0].Type())
			}
			m := args[0].(*object.Map)
			newMap := clone.Clone(m).(*object.Map)

			return newMap
		},
	},
	"keys": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`keys` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.MAP_OBJ {
				return newError("argument 1 to `keys` should be MAP. got=%s", args[0].Type())
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
	},
	"values": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`values` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.MAP_OBJ {
				return newError("argument 1 to `values` should be MAP. got=%s", args[0].Type())
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
	},
	"len": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `len`. got=%d, want=1", len(args))
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
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"first": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `first`. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument to `first` must be LIST, got %s", args[0].Type())
			}
			l := args[0].(*object.List)
			if len(l.Elements) > 0 {
				return l.Elements[0]
			}
			return NULL
		},
	},
	"last": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `last`. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument to `last` must be LIST, got %s", args[0].Type())
			}
			l := args[0].(*object.List)
			last := len(l.Elements)
			if last != 0 {
				return l.Elements[last-1]
			}
			return NULL
		},
	},
	"rest": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `rest`. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument to `rest` must be LIST, got %s", args[0].Type())
			}
			l := args[0].(*object.List)
			length := len(l.Elements)
			if length > 0 {
				// NOTE: This is an efficient way of copying slices/lists so use elsewhere
				newElements := make([]object.Object, length-1)
				copy(newElements, l.Elements[1:length])
				return &object.List{Elements: newElements}
			}
			return NULL
		},
	},
	"append": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments to `append`. got=%d, want=2", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument to `append` must be LIST, got %s", args[0].Type())
			}
			l := args[0].(*object.List)
			length := len(l.Elements)
			// NOTE: This is an efficient way of appending but probably could just append onto the list
			newElements := make([]object.Object, length+1)
			copy(newElements, l.Elements)
			newElements[length] = args[1]
			return &object.List{Elements: newElements}
		},
	},
	"push": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments to `push`. got=%d, want=2", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument to `push` must be LIST, got %s", args[0].Type())
			}
			l := args[0].(*object.List).Elements
			l = append([]object.Object{args[1]}, l...)
			return &object.List{Elements: l}
		},
	},
	"println": {
		Fun: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
	"print": {
		Fun: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			return NULL
		},
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
					return newError("argument to `input` must be STRING, got %s", args[0].Type())
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
			return newError("wrong number of arguments to `input`. got=%d, want=1 (or none)", len(args))
		},
	},
	"_read": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`read` expected 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument to `read` must be STRING, got %s", args[0].Type())
			}
			if args[1].Type() != object.BOOLEAN_OBJ {
				return newError("argument 2 to `read` must be BOOLEAN, got %s", args[0].Type())
			}
			fnameo, ok := args[0].(*object.Stringo)
			if !ok {
				return newError("filename string object did not match expected object in `read`. got=%T", args[0])
			}
			bs, err := os.ReadFile(fnameo.Value)
			if err != nil {
				return newError("`read` error reading file `%s`: %s", fnameo.Value, err.Error())
			}
			if args[1].(*object.Boolean).Value {
				return &object.Bytes{Value: bs}
			}
			return &object.Stringo{Value: string(bs)}
		},
	},
	"_write": {
		Fun: func(args ...object.Object) object.Object {
			argLen := len(args)
			if argLen != 2 {
				return newError("wrong number of arguments to `write`. got=%d, want=2", argLen)
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `write` must be STRING, got %s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.BYTES_OBJ {
				return newError("argument 2 to `write` must be STRING or BYTES, got %s", args[1].Type())
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
	},
	"set": {
		// This is needed so we can return a set from a list of objects
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments to `set`. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("arguments to `set` must be LIST, got %s", args[0].Type())
			}
			elements := args[0].(*object.List).Elements
			setMap := object.NewSetElements()
			for _, e := range elements {
				hashKey := object.HashObject(e)
				setMap.Set(hashKey, object.SetPair{Value: e, Present: true})
			}
			return &object.Set{Elements: setMap}
		},
	},
	"error": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`error` expects 1 argument")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("`error` argument 1 was not STRING. got=%s", args[0].Type())
			}
			msg, ok := args[0].(*object.Stringo)
			if !ok {
				return newError("`error` argument 1 was not STRING. got=%T", args[0])
			}
			return &object.Error{Message: msg.Value}
		},
	},
	"assert": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 && len(args) != 2 {
				return newError("`assert` expects 1 or 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.BOOLEAN_OBJ {
				return newError("`assert` expects first argument to be BOOLEAN. got=%s", args[0].Type())
			}
			b, ok := args[0].(*object.Boolean)
			if !ok {
				return newError("`assert` first argument was not BOOLEAN. got=%T", args[0])
			}
			if len(args) == 1 {
				if b.Value {
					return TRUE
				} else {
					return newError("`assert` failed")
				}
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("`assert` expects second argument to be STRING. got=%s", args[1].Type())
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
	},
	"type": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`type` expects 1 argument. got=%d", len(args))
			}
			return &object.Stringo{Value: string(args[0].Type())}
		},
	},
	"exec": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`exec` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("`exec` expects first argument to be STRING. got=%s", args[0].Type())
			}
			return ExecStringCommand(args[0].(*object.Stringo).Value)
		},
	},
	"is_alive": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`is_alive` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("`is_alive` expects argument to be UINTEGER. got=%s", args[0].Type())
			}
			_, isAlive := ProcessMap.Get(args[0].(*object.UInteger).Value)
			if isAlive {
				return TRUE
			}
			return FALSE
		},
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
				return newError("`exit` expects 1 or no arguments. got=%d", len(args))
			}
		},
	},
	"cwd": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) > 0 {
				return newError("`cwd` expects no arguments. got=%d", len(args))
			}
			dir, err := os.Getwd()
			if err != nil {
				return newError("`cwd` error: %s", err.Error())
			}
			return &object.Stringo{Value: dir}
		},
	},
	"cd": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`cd` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("`cd` argument must be STRING. got=%s", args[0].Type())
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
	},
	"_recv": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`recv` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("argument 1 to `recv` should be UINTEGER. got=%s", args[0].Type())
			}
			pid := args[0].(*object.UInteger).Value
			process, ok := ProcessMap.Get(pid)
			if !ok {
				return newError("`recv` failed, pid=%d not found", pid)
			}
			val := <-process.Ch

			return val
		},
	},
	"_send": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`send` expects 2 arguments. got=%d", len(args))
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newError("first argument to `send` must be UINTEGER got %s", args[0].Type())
			}
			pid := args[0].(*object.UInteger).Value
			process, ok := ProcessMap.Get(pid)
			if !ok {
				return newError("`send` failed, pid=%d not found", pid)
			}
			process.Ch <- args[1]
			return NULL
		},
	},
	"to_bytes": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`to_bytes` expects 1 argument. got=%d", len(args))
			}
			switch x := args[0].(type) {
			case *object.Stringo:
				// TODO: Support hexadecimal string? like what is currently returned in crypto
				return &object.Bytes{Value: []byte(x.Value)}
			// case *object.List:
			// Confirm all elements are ints and then convert?
			// All the ints also have to be less than 255
			default:
				// TODO: Could we maybe take the string representation and convert it to bytes?
				return newError("type '%s' not supported for `to_bytes`", args[0].Type())
			}
		},
	},
	"is_file": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`is_file` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `is_file` should be STRING. got=%s", args[0].Type())
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
	},
	"is_dir": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`is_dir` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `is_dir` should be STRING. got=%s", args[0].Type())
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
	},
	"find_exe": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`find_exe` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `find_exe` should be STRING. got=%s", args[0].Type())
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
	},
	// TODO: Support uint, and big int?
	"to_num": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`to_num` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `to_num` should be STRING. got=%s", args[0].Type())
			}
			i, err := strconv.ParseInt(args[0].(*object.Stringo).Value, 10, 64)
			if err != nil {
				return newError("`to_num` error: %s", err.Error())
			}
			return &object.Integer{Value: i}
		},
	},
	// TODO: Eventually we need to support files better (and possibly, stdin, stderr, stdout) and then http stuff
})
