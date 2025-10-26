package object

import (
	"blue/consts"
	"blue/evaluator/wazm"
	"io"
	"os"
	"strings"
	"time"

	"github.com/tetratelabs/wazero/api"
)

var WazmBuiltins = NewBuiltinSliceType{
	{Name: "_wasm_init", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			//_wasm_init(wasm_code_path, args, mounts, stdout, stderr, stdin, envs, enable_rand, enable_time_and_sleep_precision, host_logging, listens, timeout)
			//(wasm_code_path, args=ARGV, mounts={'.':'/'}, stdout=FSTDOUT, stderr=FSTDERR, stdin=FSTDIN,
			//envs=ENV, enable_rand=true, enable_time_and_sleep_precision=true, host_logging='', listens=[], timeout=0) {
			if len(args) != 12 {
				return newInvalidArgCountError("wasm_init", len(args), 12, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("wasm_init", 1, STRING_OBJ, args[0].Type())
			}
			wasmCodePath := args[0].(*Stringo).Value
			wasmArgs := []string{}
			if args[1].Type() != LIST_OBJ && args[1].Type() != NULL_OBJ {
				return newPositionalTypeError("wasm_init", 2, "list[str] or null", args[1].Type())
			}
			if args[1].Type() == LIST_OBJ {
				l := args[1].(*List).Elements
				wasmArgs = make([]string, len(l))
				for i, e := range l {
					if e.Type() != STRING_OBJ {
						return newError("`wasm_init` error: found non-string element in 'args' list")
					}
					wasmArgs[i] = e.(*Stringo).Value
				}
			}
			mounts := make(map[string]string)
			if args[2].Type() != MAP_OBJ && args[2].Type() != NULL_OBJ {
				return newPositionalTypeError("wasm_init", 3, "map[str]str or null", args[2].Type())
			}
			if args[2].Type() == MAP_OBJ {
				m := args[2].(*Map).Pairs
				for _, k := range m.Keys {
					mp, _ := m.Get(k)
					if mp.Key.Type() != STRING_OBJ {
						return newError("`wasm_init` error: found non-string key in 'mounts' map")
					}
					if mp.Value.Type() != STRING_OBJ {
						return newError("`wasm_init` error: found non-string key in 'mounts' map")
					}
					mounts[mp.Key.(*Stringo).Value] = mp.Value.(*Stringo).Value
				}
			}
			if args[3].Type() != GO_OBJ && args[3].Type() != NULL_OBJ {
				return newPositionalTypeError("wasm_init", 4, "GO_OBJ[*os.File] or null", args[3].Type())
			}
			var stdout io.Writer = nil
			var stdin io.Reader = nil
			var stderr *os.File
			if args[3].Type() == GO_OBJ {
				sout, ok := args[3].(*GoObj[*os.File])
				if !ok {
					return newPositionalTypeErrorForGoObj("wasm_init", 4, "*os.File", args[3])
				}
				stdout = sout.Value
			} else {
				stdout = nil
			}
			if args[4].Type() != GO_OBJ && args[4].Type() != NULL_OBJ {
				return newPositionalTypeError("wasm_init", 5, "GO_OBJ[*os.File] or null", args[4].Type())
			}
			if args[4].Type() == GO_OBJ {
				serr, ok := args[4].(*GoObj[*os.File])
				if !ok {
					return newPositionalTypeErrorForGoObj("wasm_init", 5, "*os.File", args[4])
				}
				stderr = serr.Value
			} else {
				stderr = nil
			}
			if args[5].Type() != GO_OBJ && args[5].Type() != NULL_OBJ {
				return newPositionalTypeError("wasm_init", 6, "GO_OBJ[*os.File] or null", args[5].Type())
			}
			if args[5].Type() == GO_OBJ {
				sin, ok := args[5].(*GoObj[*os.File])
				if !ok {
					return newPositionalTypeErrorForGoObj("wasm_init", 6, "*os.File", args[5])
				}
				stdin = sin.Value
			} else {
				stdin = nil
			}
			envs := make(map[string]string)
			if args[6].Type() != MAP_OBJ && args[6].Type() != NULL_OBJ {
				return newPositionalTypeError("wasm_init", 7, "map[str]str or null", args[6].Type())
			}
			if args[6].Type() == MAP_OBJ {
				m := args[6].(*Map).Pairs
				for _, k := range m.Keys {
					mp, _ := m.Get(k)
					if mp.Key.Type() != STRING_OBJ {
						return newError("`wasm_init` error: found non-string key in 'envs' map")
					}
					if mp.Value.Type() != STRING_OBJ {
						return newError("`wasm_init` error: found non-string value in 'envs' map")
					}
					envs[mp.Key.(*Stringo).Value] = mp.Value.(*Stringo).Value
				}
			}
			if args[7].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("wasm_init", 8, BOOLEAN_OBJ, args[7].Type())
			}
			enableRand := args[7].(*Boolean).Value
			if args[8].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("wasm_init", 9, BOOLEAN_OBJ, args[8].Type())
			}
			enableTimeAndSleepPrecision := args[8].(*Boolean).Value
			if args[9].Type() != STRING_OBJ {
				return newPositionalTypeError("wasm_init", 10, STRING_OBJ, args[9].Type())
			}
			hostLogging := args[9].(*Stringo).Value
			listens := []string{}
			if args[10].Type() != LIST_OBJ && args[10].Type() != NULL_OBJ {
				return newPositionalTypeError("wasm_init", 11, "list[str] or null", args[10].Type())
			}
			if args[10].Type() == LIST_OBJ {
				l := args[10].(*List).Elements
				listens = make([]string, len(l))
				for i, e := range l {
					if e.Type() != STRING_OBJ {
						return newError("`wasm_init` error: found non-string element in 'listens' list")
					}
					listens[i] = e.(*Stringo).Value
				}
			}
			if args[11].Type() != INTEGER_OBJ {
				return newPositionalTypeError("wasm_init", 12, INTEGER_OBJ, args[11].Type())
			}
			timeoutDuration := time.Duration(args[11].(*Integer).Value)

			var bs []byte
			if IsEmbed {
				s := wasmCodePath
				if strings.HasPrefix(s, "./") {
					s = strings.TrimLeft(s, "./")
				}
				fileData, err := Files.ReadFile(consts.EMBED_FILES_PREFIX + s)
				if err != nil {
					// Fallback option for reading when in embedded context
					fileData, err := os.ReadFile(wasmCodePath)
					if err != nil {
						return newError("`wasm_init` error reading wasm_code_path `%s`: %s", wasmCodePath, err.Error())
					}
					bs = fileData
				} else {
					bs = fileData
				}
			} else {
				fileData, err := os.ReadFile(wasmCodePath)
				if err != nil {
					return newError("`wasm_init` error reading wasm_code_path `%s`: %s", wasmCodePath, err.Error())
				}
				bs = fileData
			}
			wc := wazm.Config{
				WasmExe: bs,
				StdIn:   stdin,
				StdOut:  stdout,
				StdErr:  stderr,
				Args:    wasmArgs,
				Envs:    envs,
				Mounts:  mounts,
				Listens: listens,

				EnableRandSource:            enableRand,
				EnableTimeAndSleepPrecision: enableTimeAndSleepPrecision,

				HostLogging: hostLogging,
				Timeout:     timeoutDuration,
			}
			wm, err := wazm.WazmInit(wc)
			if err != nil {
				return newError("`wasm_init` error: failed initalizing %s", err.Error())
			}
			return NewGoObj(wm)
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_init` initalizes a wasm module with all the necessary parameters to interact with it. Note: the module should be built with wasi_preview1 ie. GOOS=wasip1 GOARCH=wasm go build -o cat.wasm",
			signature: `wasm_init(wasm_code_path: str, args: list[str], mounts: map[str:str], stdout: GoObj[*os.File],
		stderr: GoObj[*os.File], stdin: GoObj[*os.File], envs: map[str:str], enable_rand: bool=true
		enable_time_and_sleep_precision: bool=true, host_logging: str='', listens: list[str]|null=[], timeout: int=0) -> GoObj[*wazm.Module]`,
			errors:  "InvalidArgCount,PositionalType,CustomError",
			example: "wasm_init('wasm_test_files/cat.wasm', args=['wasm_test_files/cat.go.tmp']) => GoObj[*wazm.Module]",
		}.String(),
	}},
	{Name: "_wasm_get_functions", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("wasm_get_functions", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("wasm_get_functions", 1, GO_OBJ, "")
			}
			wm, ok := args[0].(*GoObj[*wazm.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("wasm_get_functions", 1, "*wazm.Module", args[0])
			}
			funs := wazm.GetFunctions(wm.Value)
			l := &List{
				Elements: make([]Object, len(funs)),
			}
			for i, fun := range funs {
				l.Elements[i] = &Stringo{Value: fun}
			}
			return l
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_get_functions` returns the available functions on the wasm module and works closely with wasm_get_exported_functions",
			signature:   "wasm_get_functions(mod: GoObj[*wazm.Module])",
			errors:      "InvalidArgCount,PositionalType",
			example:     "wasm_get_functions(add_mod) => ['realloc', '_start', 'add', 'asyncify_start_unwind', 'asyncify_stop_unwind', 'asyncify_start_rewind', 'free', 'calloc', 'asyncify_stop_rewind', 'malloc', 'asyncify_get_state']",
		}.String(),
	}},
	{Name: "_wasm_get_exported_function", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("wasm_get_exported_function", len(args), 2, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("wasm_get_exported_function", 1, GO_OBJ, args[0].Type())
			}
			wm, ok := args[0].(*GoObj[*wazm.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("wasm_get_exported_function", 1, "*wazm.Module", args[0])
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("wasm_get_exported_function", 2, STRING_OBJ, args[1].Type())
			}
			fnName := args[1].(*Stringo).Value
			if _, ok := wm.Value.Module.ExportedFunctions()[fnName]; !ok {
				return newError("`wasm_get_exported_function` error: function '%s' not found", fnName)
			}
			// Return a builtin function to call the wasm function
			return &Builtin{
				Fun: func(args ...Object) Object {
					argsForCall := make([]uint64, len(args))
					for i, arg := range args {
						if arg.Type() != UINTEGER_OBJ {
							return newPositionalTypeError("wasm_call", i+1, UINTEGER_OBJ, arg.Type())
						}
						argsForCall[i] = arg.(*UInteger).Value
					}
					var mod api.Module
					// TODO: Figure out timeout stuff
					// if wm.Value.CancelFun != nil {
					// 	defer wm.Value.CancelFun()
					// }
					if !wm.Value.IsInstantiated {
						module, _, err := wazm.WazmRun(wm.Value)
						if err != nil {
							return newError("`wasm_call` error: instantiating failed %s", err.Error())
						}
						wm.Value = module
						mod = wm.Value.ApiMod
					} else {
						mod = wm.Value.ApiMod
					}
					fn := mod.ExportedFunction(fnName)
					var err error
					var retVal []uint64
					if len(argsForCall) == 0 {
						retVal, err = fn.Call(wm.Value.Ctx)
					} else {
						retVal, err = fn.Call(wm.Value.Ctx, argsForCall...)
					}
					if err != nil {
						return newError("`wasm_call` error: calling '%s' failed with params %v. %s", fnName, argsForCall, err.Error())
					}
					returnValue := &List{
						Elements: make([]Object, len(retVal)),
					}
					for i, e := range retVal {
						returnValue.Elements[i] = &UInteger{Value: e}
					}
					return returnValue
				},
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_get_exported_functions` returns the available function on the wasm module to be callable (via a BUILTIN) and works closely with wasm_get_functions",
			signature:   "wasm_get_exported_functions(mod: GoObj[*wazm.Module], func: str) -> (fn(any...) -> any)",
			errors:      "InvalidArgCount,PositionalType",
			example:     "wasm_get_exported_functions(add_mod, 'add')(0x3, 0x7) => 0u10",
		}.String(),
	}},
	{Name: "_wasm_run", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("wasm_run", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("wasm_run", 1, GO_OBJ, args[0].Type())
			}
			wm, ok := args[0].(*GoObj[*wazm.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("wasm_run", 1, "*wazm.Module", args[0])
			}
			if wm.Value.CancelFun != nil {
				defer wm.Value.CancelFun()
			}
			defer wm.Value.Runtime.Close(wm.Value.Ctx)
			module, rc, err := wazm.WazmRun(wm.Value)
			if err != nil {
				return newError("`wasm_run` error: %s", err.Error())
			}
			wm.Value = module
			return &Integer{Value: int64(rc)}
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_run` runs the main or _start of the wasm module and returns its return code as an integer",
			signature:   "wasm_run(mod: GoObj[*wazm.Module]) -> int",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "wasm_run(cat_mod) => 0 (side-effects may happen such as writing to stdout)",
		}.String(),
	}},
	{Name: "_wasm_close", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("wasm_close", len(args), 1, "")
			}
			if args[0].Type() != GO_OBJ {
				return newPositionalTypeError("wasm_close", 1, GO_OBJ, args[0].Type())
			}
			wm, ok := args[0].(*GoObj[*wazm.Module])
			if !ok {
				return newPositionalTypeErrorForGoObj("wasm_close", 1, "*wazm.Module", args[0])
			}
			err := wm.Value.Runtime.Close(wm.Value.Ctx)
			if err != nil {
				return newError("`wasm_close` error: %s", err.Error())
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`wasm_close` closes the wasm module and disposes of the resource, currently if an error occurs a string is returned with the error",
			signature:   "wasm_close(mod: GoObj[*wazm.Module]) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "wasm_close(cat_mod) => null",
		}.String(),
	}},
}
