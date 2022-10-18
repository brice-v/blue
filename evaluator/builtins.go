package evaluator

import (
	"blue/object"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type BuiltinMapTypeInternal map[string]*object.Builtin

var builtins = NewBuiltinObjMap(BuiltinMapTypeInternal{
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
				newElements := make([]object.Object, length-1, length-1)
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
			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, l.Elements)
			newElements[length] = args[1]
			return &object.List{Elements: newElements}
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
	"read": {
		Fun: func(args ...object.Object) object.Object {
			argLen := len(args)
			if argLen != 1 {
				return newError("wrong number of arguments to `read`. got=%d, want=1", argLen)
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument to `read` must be STRING, got %s", args[0].Type())
			}
			fnameo, ok := args[0].(*object.Stringo)
			if !ok {
				return newError("filename string object did not match expected object in `read`. got=%T", args[0])
			}
			bs, err := os.ReadFile(fnameo.Value)
			if err != nil {
				return newError("`read` error reading file `%s`: %s", fnameo.Value, err.Error())
			}
			return &object.Stringo{Value: string(bs)}
		},
	},
	"write": {
		Fun: func(args ...object.Object) object.Object {
			argLen := len(args)
			if argLen != 2 {
				return newError("wrong number of arguments to `write`. got=%d, want=2", argLen)
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("arguments to `write` must be STRING, got %s", args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newError("arguments to `write` must be STRING, got %s", args[1].Type())
			}
			fname := args[0].(*object.Stringo).Value
			contents := args[1].(*object.Stringo).Value
			err := os.WriteFile(fname, []byte(contents), 0644)
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
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("`is_alive` expects argument to be INTEGER. got=%s", args[0].Type())
			}
			_, isAlive := ProcessMap.Get(args[0].(*object.Integer).Value)
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
	"recv": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`recv` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("argument 1 to `recv` should be INTEGER. got=%s", args[0].Type())
			}
			pid := args[0].(*object.Integer).Value
			process, ok := ProcessMap.Get(pid)
			if !ok {
				return newError("`recv` failed, pid=%d not found", pid)
			}
			val := <-process.Ch

			return val
		},
	},
	"send": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("`send` expects 2 arguments")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("first argument to `send` must be INTEGER got %s", args[0].Type())
			}
			pid := args[0].(*object.Integer).Value
			process, ok := ProcessMap.Get(pid)
			if !ok {
				return newError("`send` failed, pid=%d not found", pid)
			}
			process.Ch <- args[1]
			return NULL
		},
	},
	// TODO: Eventually we need to support files better (and possibly, stdin, stderr, stdout) and then http stuff
})
