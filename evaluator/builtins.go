package evaluator

import (
	"blue/object"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gobuffalo/plush"
	clone "github.com/huandu/go-clone"
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
			if args[0].Type() != object.MAP_OBJ {
				return newPositionalTypeError("new", 1, object.MAP_OBJ, args[0].Type())
			}
			m := args[0].(*object.Map)
			newMap := clone.Clone(m).(*object.Map)

			return newMap
		},
		HelpStr: helpStrArgs{
			explanation: "`new` returns a cloned MAP object from the given arg",
			signature:   "new(arg: map) -> map",
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
			default:
				return newPositionalTypeError("len", 1, "STRING, LIST, MAP, or SET", args[0].Type())
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
			if len(args) != 2 {
				return newInvalidArgCountError("append", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("append", 1, object.LIST_OBJ, args[0].Type())
			}
			l := args[0].(*object.List)
			length := len(l.Elements)
			// NOTE: This is an efficient way of appending but probably could just append onto the list
			newElements := make([]object.Object, length+1)
			copy(newElements, l.Elements)
			newElements[length] = args[1]
			return &object.List{Elements: newElements}
		},
		HelpStr: helpStrArgs{
			explanation: "`append` returns the LIST of elements with given arg OBJECT at the end",
			signature:   "append(arg0: list, arg1: any) -> list[any]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "append([1,2,3], 1) => [1,2,3,4]",
		}.String(),
	},
	"push": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("push", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("push", 1, object.LIST_OBJ, args[0].Type())
			}
			l := args[0].(*object.List).Elements
			l = append([]object.Object{args[1]}, l...)
			return &object.List{Elements: l}
		},
		HelpStr: helpStrArgs{
			explanation: "`push` returns the LIST of elements with given arg OBJECT at the front",
			signature:   "push(arg0: list[any], arg1: any) -> list[any]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "push([1,2,3], 1) => [1,1,2,3]",
		}.String(),
	},
	"println": {
		Fun: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
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
			for _, arg := range args {
				fmt.Print(arg.Inspect())
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
			bs, err := os.ReadFile(fnameo.Value)
			if err != nil {
				return newError("`read` error reading file `%s`: %s", fnameo.Value, err.Error())
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
			if len(args) != 1 {
				return newInvalidArgCountError("set", len(args), 1, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("set", 1, object.LIST_OBJ, args[0].Type())
			}
			elements := args[0].(*object.List).Elements
			setMap := object.NewSetElements()
			for _, e := range elements {
				hashKey := object.HashObject(e)
				setMap.Set(hashKey, object.SetPair{Value: e, Present: true})
			}
			return &object.Set{Elements: setMap}
		},
		HelpStr: helpStrArgs{
			explanation: "`set` returns the SET version of a LIST",
			signature:   "set(arg: list[any]) -> set[any]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "set([1,2,2,3]) => {1,2,3}",
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
				ctx.Set(mp.Key.(*object.Stringo).Value, blueObjectToGoObject(mp.Value))
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
			if args[0].Type() != object.UINTEGER_OBJ {
				return newPositionalTypeError("is_alive", 1, object.UINTEGER_OBJ, args[0].Type())
			}
			_, isAlive := ProcessMap.Get(args[0].(*object.UInteger).Value)
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
	"_recv": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("recv", len(args), 1, "")
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newPositionalTypeError("recv", 1, object.UINTEGER_OBJ, args[0].Type())
			}
			pid := args[0].(*object.UInteger).Value
			process, ok := ProcessMap.Get(pid)
			if !ok {
				return newError("`recv` failed, pid=%d not found", pid)
			}
			val := <-process.Ch

			return val
		},
		HelpStr: helpStrArgs{
			explanation: "`_recv` waits for a value on the given UINTEGER (process) and returns it",
			signature:   "_recv(pid: uint) -> any",
			errors:      "InvalidArgCount,PositionalType,PidNotFound",
			example:     "_recv(0x0) => 'something'",
		}.String(),
	},
	"_send": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("send", len(args), 2, "")
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newPositionalTypeError("send", 1, object.UINTEGER_OBJ, args[0].Type())
			}
			pid := args[0].(*object.UInteger).Value
			process, ok := ProcessMap.Get(pid)
			if !ok {
				return newError("`send` failed, pid=%d not found", pid)
			}
			process.Ch <- args[1]
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`_send` will take the given value and send it to the UINTEGER (process)",
			signature:   "_send(pid: uint, val: any) -> null",
			errors:      "InvalidArgCount,PositionalType,PidNotFound",
			example:     "_send(0x0, 'hello') => null",
		}.String(),
	},
	"to_bytes": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_bytes", len(args), 1, "")
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
		HelpStr: helpStrArgs{
			explanation: "`to_bytes` returns the BYTE representation of the given STRING",
			signature:   "to_bytes(arg: str) -> bytes",
			errors:      "InvalidArgCount,TypeNotSupported",
			example:     "to_bytes('hello') => []byte{0x68, 0x65, 0x6c, 0x6c, 0x6f}",
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
	// TODO: Support uint, and big int?
	"to_num": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_num", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("to_num", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			// This is an overkill way of making sure the string is cleaned for parsing
			// - at least with regards to whitespace
			re := regexp.MustCompile("[\r\n\t ]")
			cleanS := string(re.ReplaceAll([]byte(s), []byte("")))
			i, err := strconv.ParseInt(cleanS, 10, 64)
			if err != nil {
				return newError("`to_num` error: %s", err.Error())
			}
			return &object.Integer{Value: i}
		},
		HelpStr: helpStrArgs{
			explanation: "`to_num` returns the INTEGER value of the given STRING",
			signature:   "to_num(arg: str) -> int",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "to_num('1') => 1",
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
})
