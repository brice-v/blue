package evaluator

import (
	"blue/lexer"
	"blue/object"
	"blue/parser"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// StdModFileAndBuiltins keeps the file and builtins together for each std lib module
type StdModFileAndBuiltins struct {
	File     string         // File is the actual code used for the module
	Builtins BuiltinMapType // Builtins is the map of functions to be used by the module
}

//go:embed std/http.b
var stdHttpFile string

//go:embed std/time.b
var stdTimeFile string

// TODO: Could use an embed.FS and read the files that way rather then set each one individually
// but it works well enough for now (if we used embed.FS we probably just need a helper)
var _std_mods = map[string]StdModFileAndBuiltins{
	"http": {File: stdHttpFile, Builtins: _http_builtin_map},
	"time": {File: stdTimeFile, Builtins: _time_builtin_map},
}

func (e *Evaluator) IsStd(name string) bool {
	_, ok := _std_mods[name]
	return ok
}

func (e *Evaluator) AddStdLibToEnv(name string) {
	if !e.IsStd(name) {
		fmt.Printf("AddStdLibToEnv: '%s' is not in std lib map\n", name)
		os.Exit(1)
	}
	fb := _std_mods[name]
	l := lexer.New(fb.File)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("ParserError in `%s` module: %s\n", name, msg)
		}
		os.Exit(1)
	}
	newE := New()
	newE.Builtins.PushBack(fb.Builtins)
	val := newE.Eval(program)
	if isError(val) {
		fmt.Printf("EvaluatorError in `%s` module: %s\n", name, val.(*object.Error).Message)
		os.Exit(1)
	}
	mod := &object.Module{Name: name, Env: newE.env}
	e.env.Set(name, mod)
}

// Note: Look at how we import the get function in http.b
var _http_builtin_map = BuiltinMapType{
	"_get": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`get` expects 1 argument")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument to `get` must be STRING. got %s", args[0].Type())
			}
			resp, err := http.Get(args[0].(*object.Stringo).Value)
			if err != nil {
				return newError("`get` failed: %s", err.Error())
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return newError("`get` failed: %s", err.Error())
			}
			return &object.Stringo{Value: string(body)}
		},
	},
}

var _time_builtin_map = BuiltinMapType{
	"_sleep": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`sleep` expects 1 argument")
			}
			if args[0].Type() != object.INTEGER_OBJ {
				return newError("argument to `sleep` must be INTEGER, got %s", args[0].Type())
			}
			i := args[0].(*object.Integer).Value
			if i < 0 {
				return newError("INTEGER argument to `sleep` must be > 0, got %d", i)
			}
			time.Sleep(time.Duration(i) * time.Millisecond)
			return NULL
		},
	},
	"_now": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 0 {
				return newError("`now` expects 0 arguments")
			}
			return &object.Integer{Value: time.Now().Unix()}
		},
	},
}
