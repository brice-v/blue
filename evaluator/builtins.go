package evaluator

import (
	"blue/object"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/metrics"
	"strings"
	"time"

	"github.com/gobuffalo/plush"
	"github.com/gookit/color"
	clone "github.com/huandu/go-clone"
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
			useColorPrinter := false
			var style color.Style
			for i, arg := range args {
				if i == 0 {
					t, v, ok := getBasicObject(arg)
					if ok && t == "color" {
						// Use color printer
						useColorPrinter = true
						s, ok := ColorStyleMap.Get(v)
						if !ok {
							useColorPrinter = false
						}
						style = s
						continue
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
					t, v, ok := getBasicObject(arg)
					if ok && t == "color" {
						// Use color printer
						useColorPrinter = true
						s, ok := ColorStyleMap.Get(v)
						if !ok {
							useColorPrinter = false
						}
						style = s
						continue
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
			default:
				return newError("`int` error: unsupported type '%s'", args[0].Type())
			}
		},
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
			default:
				return newError("`float` error: unsupported type '%s'", args[0].Type())
			}
		},
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
			default:
				return newError("`bigint` error: unsupported type '%s'", args[0].Type())
			}
		},
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
				return &object.BigFloat{Value: decimal.NewFromBigInt(new(big.Int).SetUint64(args[0].(*object.UInteger).Value), 1)}
			case object.BIG_INTEGER_OBJ:
				return &object.BigFloat{Value: decimal.NewFromBigInt(args[0].(*object.BigInteger).Value, 1)}
			case object.BIG_FLOAT_OBJ:
				return args[0]
			default:
				return newError("`bigfloat` error: unsupported type '%s'", args[0].Type())
			}
		},
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
			default:
				return newError("`uint` error: unsupported type '%s'", args[0].Type())
			}
		},
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
	},
	"get_os": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_os", len(args), 0, "")
			}
			return &object.Stringo{Value: runtime.GOOS}
		},
	},
	"get_arch": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newInvalidArgCountError("get_arch", len(args), 0, "")
			}
			return &object.Stringo{Value: runtime.GOARCH}
		},
	},
	"is_valid_json": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_valid_json", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("is_valid_json", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return &object.Boolean{Value: json.Valid([]byte(s))}
		},
	},
	"wait": {
		Fun: func(args ...object.Object) object.Object {
			// This function will avoid returning errors
			// but that means random inputs will technically be allowed
			pidsToWaitFor := []uint64{}
			for _, arg := range args {
				if pids, ok := getListOfPids(arg); ok {
					pidsToWaitFor = append(pidsToWaitFor, pids...)
					continue
				}
				if arg.Type() == object.MAP_OBJ {
					t, v, ok := getBasicObject(arg)
					if !ok || t != "pid" {
						continue
					}
					pidsToWaitFor = append(pidsToWaitFor, v)
				}
			}
			for {
				allDone := false
				for _, pid := range pidsToWaitFor {
					_, ok := ProcessMap.Get(pid)
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
			t, v, ok := getBasicObject(args[0])
			if !ok {
				return newError("`add_topic` error: argument 1 should be in the format {t: 'sub', v: uint}. got=%s", args[0].Inspect())
			}
			if t != "sub" {
				return newError("`add_topic` error: argument 1 should be in the format {t: 'sub', v: uint}")
			}
			sub, ok := SubscriberMap.Get(v)
			if !ok {
				return newError("`add_topic` error: subscriber with id %d, did not exist", v)
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("add_topic", 2, object.STRING_OBJ, args[1].Type())
			}
			topic := args[1].(*object.Stringo).Value
			sub.AddTopic(topic)
			return NULL
		},
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
			t, v, ok := getBasicObject(args[0])
			if !ok {
				return newError("`remove_topic` error: argument 1 should be in the format {t: 'sub', v: uint}. got=%s", args[0].Inspect())
			}
			if t != "sub" {
				return newError("`remove_topic` error: argument 1 should be in the format {t: 'sub', v: uint}")
			}
			sub, ok := SubscriberMap.Get(v)
			if !ok {
				return newError("`remove_topic` error: subscriber with id %d, did not exist", v)
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("remove_topic", 2, object.STRING_OBJ, args[1].Type())
			}
			topic := args[1].(*object.Stringo).Value
			sub.RemoveTopic(topic)
			return NULL
		},
	},
	"unsubscribe": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("unsubscribe", len(args), 2, "")
			}
			if args[0].Type() != object.MAP_OBJ {
				return newPositionalTypeError("unsubscribe", 1, object.MAP_OBJ, args[0].Type())
			}
			t, v, ok := getBasicObject(args[0])
			if !ok {
				return newError("`unsubscribe` error: argument 1 should be in the format {t: 'sub', v: uint}. got=%s", args[0].Inspect())
			}
			if t != "sub" {
				return newError("`unsubscribe` error: argument 1 should be in the format {t: 'sub', v: uint}")
			}
			sub, ok := SubscriberMap.Get(v)
			if !ok {
				return newError("`unsubscribe` error: subscriber with id %d, did not exist", v)
			}
			// Remove should also destruct the subscriber
			PubSubBroker.RemoveSubscriber(sub)
			SubscriberMap.Remove(v)
			return NULL
		},
	},
	"_pubsub_sub_listen": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("pubsub_sub_listen", len(args), 1, "")
			}
			if args[0].Type() != object.UINTEGER_OBJ {
				return newPositionalTypeError("pubsub_sub_listen", 1, object.UINTEGER_OBJ, args[0].Type())
			}
			v := args[0].(*object.UInteger).Value
			sub, ok := SubscriberMap.Get(v)
			if !ok {
				return newError("`pubsub_sub_listen` error: subscriber with id %d, did not exist", v)
			}
			return sub.PollMessage()
		},
	},
})

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

func getListOfPids(arg object.Object) ([]uint64, bool) {
	if arg.Type() != object.LIST_OBJ {
		return nil, false
	}
	pids := []uint64{}
	for _, e := range arg.(*object.List).Elements {
		if e.Type() != object.MAP_OBJ {
			return nil, false
		}
		t, v, ok := getBasicObject(e)
		if !ok || t != "pid" {
			return nil, false
		}
		pids = append(pids, v)
	}
	return pids, true
}
