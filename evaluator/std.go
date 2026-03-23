package evaluator

import (
	"blue/ast"
	"blue/consts"
	"blue/lexer"
	"blue/lib"
	"blue/object"
	"blue/parser"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// StdModFileAndBuiltins keeps the file and builtins together for each std lib module
type StdModFileAndBuiltins struct {
	File     string              // File is the actual code used for the module
	Builtins BuiltinMapType      // Builtins is the map of functions to be used by the module
	Env      *object.Environment // Env is the environment to pull the lib functions/variables from
	HelpStr  string              // HelpStr is the help string for the std lib program
}

var _std_mods = map[string]StdModFileAndBuiltins{
	"http":   {File: lib.ReadStdFileToString("http.b"), Builtins: _http_builtin_map},
	"time":   {File: lib.ReadStdFileToString("time.b"), Builtins: _time_builtin_map},
	"search": {File: lib.ReadStdFileToString("search.b"), Builtins: _search_builtin_map},
	"db":     {File: lib.ReadStdFileToString("db.b"), Builtins: _db_builtin_map},
	"math":   {File: lib.ReadStdFileToString("math.b"), Builtins: _math_builtin_map},
	"config": {File: lib.ReadStdFileToString("config.b"), Builtins: _config_builtin_map},
	"crypto": {File: lib.ReadStdFileToString("crypto.b"), Builtins: _crypto_builtin_map},
	"net":    {File: lib.ReadStdFileToString("net.b"), Builtins: _net_builtin_map},
	"color":  {File: lib.ReadStdFileToString("color.b"), Builtins: _color_builtin_map},
	"csv":    {File: lib.ReadStdFileToString("csv.b"), Builtins: _csv_builtin_map},
	"psutil": {File: lib.ReadStdFileToString("psutil.b"), Builtins: _psutil_builtin_map},
	"wasm":   {File: lib.ReadStdFileToString("wasm.b"), Builtins: _wasm_builtin_map},
	"ui":     {File: lib.ReadStdFileToString("ui-static.b"), Builtins: NewBuiltinObjMap(BuiltinMapTypeInternal{})},
	"gg":     {File: lib.ReadStdFileToString("gg-static.b"), Builtins: NewBuiltinObjMap(BuiltinMapTypeInternal{})},
}

func (e *Evaluator) IsStd(name string) bool {
	_, ok := _std_mods[name]
	return ok
}

func (e *Evaluator) AddStdLibToEnv(name string, nodeIdentsToImport []*ast.Identifier, shouldImportAll bool) object.Object {
	if !e.IsStd(name) {
		fmt.Printf("AddStdLibToEnv: '%s' is not in std lib map\n", name)
		os.Exit(1)
	}
	fb := _std_mods[name]
	if fb.Env == nil {
		l := lexer.New(fb.File, "<std/"+name+".b>")
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			for _, msg := range p.Errors() {
				splitMsg := strings.Split(msg, "\n")
				firstPart := fmt.Sprintf("%smodule `%s`: %s\n", consts.PARSER_ERROR_PREFIX, name, splitMsg[0])
				consts.ErrorPrinter(firstPart)
				for i, s := range splitMsg {
					if i == 0 {
						continue
					}
					fmt.Println(s)
				}
			}
			os.Exit(1)
		}
		newE := New()
		newE.Builtins = append(newE.Builtins, fb.Builtins)
		setupBuiltinsWithEvaluator(name, newE)
		val := newE.Eval(program)
		if isError(val) {
			errorObj := val.(*object.Error)
			var buf bytes.Buffer
			buf.WriteString(errorObj.Message)
			buf.WriteByte('\n')
			for newE.ErrorTokens.Len() > 0 {
				buf.WriteString(lexer.GetErrorLineMessage(newE.ErrorTokens.PopBack()))
				buf.WriteByte('\n')
			}
			msg := fmt.Sprintf("%smodule `%s`: %s", consts.EVAL_ERROR_PREFIX, name, buf.String())
			splitMsg := strings.Split(msg, "\n")
			for i, s := range splitMsg {
				if i == 0 {
					consts.ErrorPrinter(s + "\n")
					continue
				}
				delimeter := ""
				if i != len(splitMsg)-1 {
					delimeter = "\n"
				}
				fmt.Printf("%s%s", s, delimeter)
			}
			os.Exit(1)
		}
		NewEvaluatorLock.Lock()
		fb.Env = newE.env.Clone()
		// TODO: See if we can cache this somehow
		pubFunHelpStr := fb.Env.GetOrderedPublicFunctionHelpString()
		fb.HelpStr = CreateHelpStringFromProgramTokens(name, program.HelpStrTokens, pubFunHelpStr)
		NewEvaluatorLock.Unlock()
	}

	if len(nodeIdentsToImport) >= 1 {
		for _, ident := range nodeIdentsToImport {
			if strings.HasPrefix(ident.Value, "_") {
				return newError("ImportError: imports must be public to import them. failed to import %s from %s", ident.Value, name)
			}
			o, ok := fb.Env.Get(ident.Value)
			if !ok {
				return newError("ImportError: failed to import %s from %s", ident.Value, name)
			}
			e.env.Set(ident.Value, o)
		}
		// return early if we specifically import some objects
		return object.NULL
	} else if shouldImportAll {
		// Here we want to import everything from the module
		fb.Env.SetAllPublicOnEnv(e.env)
		return object.NULL
	}

	mod := &object.Module{Name: name, Env: fb.Env, HelpStr: fb.HelpStr}
	e.env.Set(name, mod)
	return nil
}

// var goObjDecoders = map[string]any{}

func NewGoObj[T any](obj T) *object.GoObj[T] {
	gob := &object.GoObj[T]{Value: obj, Id: object.GoObjId.Add(1)}
	// Note: This is disabled for now due to the complexity of handling all Go Object Types supported by blue
	// t := fmt.Sprintf("%T", gob)
	// if _, ok := goObjDecoders[t]; !ok {
	// 	goObjDecoders[t] = gob.Decoder
	// }
	return gob
}

// Note: Look at how we import the get function in http.b
var _http_builtin_map = NewBuiltinMap(object.HttpBuiltins)

var _time_builtin_map = NewBuiltinMap(object.TimeBuiltins)

var _search_builtin_map = NewBuiltinMap(object.SearchBuiltins)

var _db_builtin_map = NewBuiltinMap(object.DbBuiltins)

var _math_builtin_map = NewBuiltinMap(object.MathBuiltins)

var _config_builtin_map = NewBuiltinMap(object.ConfigBuiltins)

var _crypto_builtin_map = NewBuiltinMap(object.CryptoBuiltins)

var _net_builtin_map = NewBuiltinMap(object.NetBuiltins)

var _color_builtin_map = NewBuiltinMap(object.ColorBuiltins)

var _csv_builtin_map = NewBuiltinMap(object.CsvBuiltins)

var _psutil_builtin_map = NewBuiltinMap(object.PsutilBuiltins)

var _wasm_builtin_map = NewBuiltinMap(object.WazmBuiltins)
