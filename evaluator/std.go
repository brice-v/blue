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
var _http_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_url_encode":          object.GetBuiltinByName(object.BuiltinHttpType, "_url_encode"),
	"_url_escape":          object.GetBuiltinByName(object.BuiltinHttpType, "_url_escape"),
	"_url_unescape":        object.GetBuiltinByName(object.BuiltinHttpType, "_url_unescape"),
	"_download":            object.GetBuiltinByName(object.BuiltinHttpType, "_download"),
	"_new_server":          object.GetBuiltinByName(object.BuiltinHttpType, "_new_server"),
	"_serve":               object.GetBuiltinByName(object.BuiltinHttpType, "_serve"),
	"_shutdown_server":     object.GetBuiltinByName(object.BuiltinHttpType, "_shutdown_server"),
	"_static":              object.GetBuiltinByName(object.BuiltinHttpType, "_static"),
	"_ws_send":             object.GetBuiltinByName(object.BuiltinHttpType, "_ws_send"),
	"_ws_recv":             object.GetBuiltinByName(object.BuiltinHttpType, "_ws_recv"),
	"_new_ws":              object.GetBuiltinByName(object.BuiltinHttpType, "_new_ws"),
	"_ws_client_send":      object.GetBuiltinByName(object.BuiltinHttpType, "_ws_client_send"),
	"_ws_client_recv":      object.GetBuiltinByName(object.BuiltinHttpType, "_ws_client_recv"),
	"_handle_monitor":      object.GetBuiltinByName(object.BuiltinHttpType, "_handle_monitor"),
	"_md_to_html":          object.GetBuiltinByName(object.BuiltinHttpType, "_md_to_html"),
	"_sanitize_and_minify": object.GetBuiltinByName(object.BuiltinHttpType, "_sanitize_and_minify"),
	"_inspect":             object.GetBuiltinByName(object.BuiltinHttpType, "_inspect"),
	"_open_browser":        object.GetBuiltinByName(object.BuiltinHttpType, "_open_browser"),
})

var _time_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_sleep":  object.GetBuiltinByName(object.BuiltinTimeType, "_sleep"),
	"_now":    object.GetBuiltinByName(object.BuiltinTimeType, "_now"),
	"_parse":  object.GetBuiltinByName(object.BuiltinTimeType, "_parse"),
	"_to_str": object.GetBuiltinByName(object.BuiltinTimeType, "_to_str"),
})

var _search_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_by_xpath": object.GetBuiltinByName(object.BuiltinSearchType, "_by_xpath"),
	"_by_regex": object.GetBuiltinByName(object.BuiltinSearchType, "_by_regex"),
})

var _db_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_db_open":  object.GetBuiltinByName(object.BuiltinDbType, "_db_open"),
	"_db_ping":  object.GetBuiltinByName(object.BuiltinDbType, "_db_ping"),
	"_db_close": object.GetBuiltinByName(object.BuiltinDbType, "_db_close"),
	"_db_exec":  object.GetBuiltinByName(object.BuiltinDbType, "_db_exec"),
	"_db_query": object.GetBuiltinByName(object.BuiltinDbType, "_db_query"),
})

var _math_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_rand":          object.GetBuiltinByName(object.BuiltinMathType, "_rand"),
	"_NaN":           object.GetBuiltinByName(object.BuiltinMathType, "_NaN"),
	"_acos":          object.GetBuiltinByName(object.BuiltinMathType, "_acos"),
	"_acosh":         object.GetBuiltinByName(object.BuiltinMathType, "_acosh"),
	"_asin":          object.GetBuiltinByName(object.BuiltinMathType, "_asin"),
	"_asinh":         object.GetBuiltinByName(object.BuiltinMathType, "_asinh"),
	"_atan":          object.GetBuiltinByName(object.BuiltinMathType, "_atan"),
	"_atan2":         object.GetBuiltinByName(object.BuiltinMathType, "_atan2"),
	"_atanh":         object.GetBuiltinByName(object.BuiltinMathType, "_atanh"),
	"_cbrt":          object.GetBuiltinByName(object.BuiltinMathType, "_cbrt"),
	"_ceil":          object.GetBuiltinByName(object.BuiltinMathType, "_ceil"),
	"_copysign":      object.GetBuiltinByName(object.BuiltinMathType, "_copysign"),
	"_cos":           object.GetBuiltinByName(object.BuiltinMathType, "_cos"),
	"_cosh":          object.GetBuiltinByName(object.BuiltinMathType, "_cosh"),
	"_dim":           object.GetBuiltinByName(object.BuiltinMathType, "_dim"),
	"_erf":           object.GetBuiltinByName(object.BuiltinMathType, "_erf"),
	"_erfc":          object.GetBuiltinByName(object.BuiltinMathType, "_erfc"),
	"_erfcinv":       object.GetBuiltinByName(object.BuiltinMathType, "_erfcinv"),
	"_erfinv":        object.GetBuiltinByName(object.BuiltinMathType, "_erfinv"),
	"_fma":           object.GetBuiltinByName(object.BuiltinMathType, "_fma"),
	"_floor":         object.GetBuiltinByName(object.BuiltinMathType, "_floor"),
	"_frexp":         object.GetBuiltinByName(object.BuiltinMathType, "_frexp"),
	"_gamma":         object.GetBuiltinByName(object.BuiltinMathType, "_gamma"),
	"_gcd":           object.GetBuiltinByName(object.BuiltinMathType, "_gcd"),
	"_hypot":         object.GetBuiltinByName(object.BuiltinMathType, "_hypot"),
	"_ilogb":         object.GetBuiltinByName(object.BuiltinMathType, "_ilogb"),
	"_inf":           object.GetBuiltinByName(object.BuiltinMathType, "_inf"),
	"_is_inf":        object.GetBuiltinByName(object.BuiltinMathType, "_is_inf"),
	"_is_NaN":        object.GetBuiltinByName(object.BuiltinMathType, "_is_NaN"),
	"_j0":            object.GetBuiltinByName(object.BuiltinMathType, "_j0"),
	"_j1":            object.GetBuiltinByName(object.BuiltinMathType, "_j1"),
	"_jn":            object.GetBuiltinByName(object.BuiltinMathType, "_jn"),
	"_lcm":           object.GetBuiltinByName(object.BuiltinMathType, "_lcm"),
	"_ldexp":         object.GetBuiltinByName(object.BuiltinMathType, "_ldexp"),
	"_lgamma":        object.GetBuiltinByName(object.BuiltinMathType, "_lgamma"),
	"_log":           object.GetBuiltinByName(object.BuiltinMathType, "_log"),
	"_log10":         object.GetBuiltinByName(object.BuiltinMathType, "_log10"),
	"_log1p":         object.GetBuiltinByName(object.BuiltinMathType, "_log1p"),
	"_log2":          object.GetBuiltinByName(object.BuiltinMathType, "_log2"),
	"_logb":          object.GetBuiltinByName(object.BuiltinMathType, "_logb"),
	"_modf":          object.GetBuiltinByName(object.BuiltinMathType, "_modf"),
	"_next_after":    object.GetBuiltinByName(object.BuiltinMathType, "_next_after"),
	"_remainder":     object.GetBuiltinByName(object.BuiltinMathType, "_remainder"),
	"_round":         object.GetBuiltinByName(object.BuiltinMathType, "_round"),
	"_round_to_even": object.GetBuiltinByName(object.BuiltinMathType, "_round_to_even"),
	"_signbit":       object.GetBuiltinByName(object.BuiltinMathType, "_signbit"),
	"_sin":           object.GetBuiltinByName(object.BuiltinMathType, "_sin"),
	"_sincos":        object.GetBuiltinByName(object.BuiltinMathType, "_sincos"),
	"_sinh":          object.GetBuiltinByName(object.BuiltinMathType, "_sinh"),
	"_tan":           object.GetBuiltinByName(object.BuiltinMathType, "_tan"),
	"_tanh":          object.GetBuiltinByName(object.BuiltinMathType, "_tanh"),
	"_trunc":         object.GetBuiltinByName(object.BuiltinMathType, "_trunc"),
	"_y0":            object.GetBuiltinByName(object.BuiltinMathType, "_y0"),
	"_y1":            object.GetBuiltinByName(object.BuiltinMathType, "_y1"),
	"_yn":            object.GetBuiltinByName(object.BuiltinMathType, "_yn"),
})

var _config_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_load_file":   object.GetBuiltinByName(object.BuiltinConfigType, "_load_file"),
	"_dump_config": object.GetBuiltinByName(object.BuiltinConfigType, "_dump_config"),
})

var _crypto_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_sha":                       object.GetBuiltinByName(object.BuiltinCryptoType, "_sha"),
	"_md5":                       object.GetBuiltinByName(object.BuiltinCryptoType, "_md5"),
	"_generate_from_password":    object.GetBuiltinByName(object.BuiltinCryptoType, "_generate_from_password"),
	"_compare_hash_and_password": object.GetBuiltinByName(object.BuiltinCryptoType, "_compare_hash_and_password"),
	"_encrypt":                   object.GetBuiltinByName(object.BuiltinCryptoType, "_encrypt"),
	"_decrypt":                   object.GetBuiltinByName(object.BuiltinCryptoType, "_decrypt"),
	"_encode_base_64_32":         object.GetBuiltinByName(object.BuiltinCryptoType, "_encode_base_64_32"),
	"_decode_base_64_32":         object.GetBuiltinByName(object.BuiltinCryptoType, "_decode_base_64_32"),
	"_decode_hex":                object.GetBuiltinByName(object.BuiltinCryptoType, "_decode_hex"),
	"_encode_hex":                object.GetBuiltinByName(object.BuiltinCryptoType, "_encode_hex"),
})

var _net_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_connect":   object.GetBuiltinByName(object.BuiltinNetType, "_connect"),
	"_listen":    object.GetBuiltinByName(object.BuiltinNetType, "_listen"),
	"_accept":    object.GetBuiltinByName(object.BuiltinNetType, "_accept"),
	"_net_close": object.GetBuiltinByName(object.BuiltinNetType, "_net_close"),
	"_net_read":  object.GetBuiltinByName(object.BuiltinNetType, "_net_read"),
	"_net_write": object.GetBuiltinByName(object.BuiltinNetType, "_net_write"),
	"_inspect":   object.GetBuiltinByName(object.BuiltinNetType, "_inspect"),
})

var _color_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_style":      object.GetBuiltinByName(object.BuiltinColorType, "_style"),
	"_normal":     object.GetBuiltinByName(object.BuiltinColorType, "_normal"),
	"_red":        object.GetBuiltinByName(object.BuiltinColorType, "_red"),
	"_cyan":       object.GetBuiltinByName(object.BuiltinColorType, "_cyan"),
	"_gray":       object.GetBuiltinByName(object.BuiltinColorType, "_gray"),
	"_blue":       object.GetBuiltinByName(object.BuiltinColorType, "_blue"),
	"_black":      object.GetBuiltinByName(object.BuiltinColorType, "_black"),
	"_green":      object.GetBuiltinByName(object.BuiltinColorType, "_green"),
	"_white":      object.GetBuiltinByName(object.BuiltinColorType, "_white"),
	"_yellow":     object.GetBuiltinByName(object.BuiltinColorType, "_yellow"),
	"_magenta":    object.GetBuiltinByName(object.BuiltinColorType, "_magenta"),
	"_bold":       object.GetBuiltinByName(object.BuiltinColorType, "_bold"),
	"_italic":     object.GetBuiltinByName(object.BuiltinColorType, "_italic"),
	"_underlined": object.GetBuiltinByName(object.BuiltinColorType, "_underlined"),
})

var _csv_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_parse": object.GetBuiltinByName(object.BuiltinCsvType, "_parse"),
	"_dump":  object.GetBuiltinByName(object.BuiltinCsvType, "_dump"),
})

var _psutil_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_cpu_usage_percent": object.GetBuiltinByName(object.BuiltinPsutilType, "_cpu_usage_percent"),
	"_cpu_info":          object.GetBuiltinByName(object.BuiltinPsutilType, "_cpu_info"),
	"_cpu_time_info":     object.GetBuiltinByName(object.BuiltinPsutilType, "_cpu_time_info"),
	"_cpu_count":         object.GetBuiltinByName(object.BuiltinPsutilType, "_cpu_count"),
	"_mem_virt_info":     object.GetBuiltinByName(object.BuiltinPsutilType, "_mem_virt_info"),
	"_mem_swap_info":     object.GetBuiltinByName(object.BuiltinPsutilType, "_mem_swap_info"),
	"_host_info":         object.GetBuiltinByName(object.BuiltinPsutilType, "_host_info"),
	"_host_temps_info":   object.GetBuiltinByName(object.BuiltinPsutilType, "_host_temps_info"),
	"_net_connections":   object.GetBuiltinByName(object.BuiltinPsutilType, "_net_connections"),
	"_net_io_info":       object.GetBuiltinByName(object.BuiltinPsutilType, "_net_io_info"),
	"_disk_partitions":   object.GetBuiltinByName(object.BuiltinPsutilType, "_disk_partitions"),
	"_disk_io_info":      object.GetBuiltinByName(object.BuiltinPsutilType, "_disk_io_info"),
	"_disk_usage":        object.GetBuiltinByName(object.BuiltinPsutilType, "_disk_usage"),
})

var _wasm_builtin_map = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"_wasm_init":                  object.GetBuiltinByName(object.BuiltinWasmType, "_wasm_init"),
	"_wasm_get_functions":         object.GetBuiltinByName(object.BuiltinWasmType, "_wasm_get_functions"),
	"_wasm_get_exported_function": object.GetBuiltinByName(object.BuiltinWasmType, "_wasm_get_exported_function"),
	"_wasm_run":                   object.GetBuiltinByName(object.BuiltinWasmType, "_wasm_run"),
	"_wasm_close":                 object.GetBuiltinByName(object.BuiltinWasmType, "_wasm_close"),
})
