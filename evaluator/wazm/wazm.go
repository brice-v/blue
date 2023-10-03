package wazm

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/experimental/gojs"
	"github.com/tetratelabs/wazero/experimental/logging"
	"github.com/tetratelabs/wazero/experimental/sock"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

type Config struct {
	WasmExe []byte
	StdIn   io.Reader
	StdOut  io.Writer
	StdErr  *os.File

	EnableRandSource            bool
	EnableTimeAndSleepPrecision bool

	Args        []string
	Envs        map[string]string
	Mounts      map[string]string
	Listens     []string
	Timeout     time.Duration
	HostLogging string
}

type Module struct {
	Module    wazero.CompiledModule
	Ctx       context.Context
	Runtime   wazero.Runtime
	Config    wazero.ModuleConfig
	CancelFun context.CancelFunc

	IsInstantiated bool
	ApiMod         api.Module
}

// args is a slice of strings to be passed to the wasm binary
// envs is map of string to string (key to value of environment variables)
// mounts is map of string (host dir) to string (wasm dir) (:ro at the end of the wasm dir for readonly mode)
// listens is a slice of strings in the format <host:port> where host is not required and if 0 is used then a random port is chosen
// timeout is a duration if 0 its disabled
// hostlogging is a comma separated list of any of these: all,clock,filesystem,memory,proc,poll,random,sock
func WazmInit(wc Config) (*Module, error) {
	rtc := wazero.NewRuntimeConfig()

	l, err := ParseHostLogging(wc.HostLogging)
	if err != nil {
		return nil, fmt.Errorf("error reading hostlogging: %v", err)
	}
	ctx := maybeHostLogging(context.Background(), logging.LogScopes(l), wc.StdErr)

	var cancelFun context.CancelFunc = nil
	if wc.Timeout > 0 {
		newCtx, cancel := context.WithTimeout(ctx, wc.Timeout)
		ctx = newCtx
		cancelFun = cancel
		rtc = rtc.WithCloseOnContextDone(true)
	} else if wc.Timeout < 0 {
		return nil, fmt.Errorf("timeout duration may not be negative, %v given", wc.Timeout)
	}

	if sockCfg, err := validateListens(wc.Listens); err != nil {
		return nil, fmt.Errorf("failed to validate listens: %v", err)
	} else {
		ctx = sock.WithConfig(ctx, sockCfg)
	}

	rt := wazero.NewRuntimeWithConfig(ctx, rtc)

	fs := wazero.NewFSConfig()
	for hostDir, wasmDir := range wc.Mounts {
		readOnly := false
		if strings.Contains(wasmDir, ":ro") {
			wasmDir = strings.Trim(wasmDir, ":ro")
			readOnly = true
		}
		hostDir, err := filepath.Abs(hostDir)
		if err != nil {
			return nil, fmt.Errorf("invalid mount: path %q invalid: %v", hostDir, err)
		}

		stat, err := os.Stat(hostDir)
		if err != nil {
			return nil, fmt.Errorf("invalid mount: path %q error: %v", hostDir, err)
		}
		if !stat.IsDir() {
			return nil, fmt.Errorf("invalid mount: path %q is not a directory", hostDir)
		}

		if readOnly {
			fs = fs.WithReadOnlyDirMount(hostDir, wasmDir)
		} else {
			fs = fs.WithDirMount(hostDir, wasmDir)
		}
	}

	conf := wazero.NewModuleConfig().
		WithFSConfig(fs)
	if wc.StdOut != nil {
		conf = conf.WithStdout(wc.StdOut)
	}
	if wc.StdErr != nil {
		conf = conf.WithStderr(wc.StdErr)
	}
	if wc.StdIn != nil {
		conf = conf.WithStdin(wc.StdIn)
	}
	if wc.EnableRandSource {
		conf = conf.WithRandSource(rand.Reader)
	}
	if wc.EnableTimeAndSleepPrecision {
		conf = conf.WithSysNanosleep().
			WithSysNanotime().
			WithSysWalltime()
	}
	conf = conf.WithArgs(wc.Args...)
	for k, v := range wc.Envs {
		if k == "" {
			continue
		}
		conf = conf.WithEnv(k, v)
	}

	guest, err := rt.CompileModule(ctx, wc.WasmExe)
	if err != nil {
		return nil, fmt.Errorf("error compiling wasm binary: %v", err)
	}

	wm := &Module{
		Module:    guest,
		Ctx:       ctx,
		Runtime:   rt,
		Config:    conf,
		CancelFun: cancelFun,
	}
	return wm, nil
}

// validateListens returns a non-nil net.Config, if there were any listen flags.
func validateListens(listens []string) (config sock.Config, err error) {
	err = nil
	for _, listen := range listens {
		idx := strings.LastIndexByte(listen, ':')
		if idx < 0 {
			return config, fmt.Errorf("invalid listen")
		}
		port, err := strconv.Atoi(listen[idx+1:])
		if err != nil {
			return config, fmt.Errorf("invalid listen port: %v", err)
		}
		if config == nil {
			config = sock.NewConfig()
		}
		config = config.WithTCPListener(listen[:idx], port)
	}
	return
}

func GetFunctions(wm *Module) []string {
	result := make([]string, len(wm.Module.ExportedFunctions()))
	i := 0
	for k := range wm.Module.ExportedFunctions() {
		result[i] = k
		i++
	}
	sort.Strings(result)
	return result
}

func wazmInstantiate(wm *Module) (apiMod api.Module, err error) {
	switch detectImports(wm.Module.ImportedFunctions()) {
	case modeWasi:
		wasi_snapshot_preview1.MustInstantiate(wm.Ctx, wm.Runtime)
		apiMod, err = wm.Runtime.InstantiateModule(wm.Ctx, wm.Module, wm.Config)
	case modeWasiUnstable:
		// Instantiate the current WASI functions under the wasi_unstable
		// instead of wasi_snapshot_preview1.
		wasiBuilder := wm.Runtime.NewHostModuleBuilder("wasi_unstable")
		wasi_snapshot_preview1.NewFunctionExporter().ExportFunctions(wasiBuilder)
		apiMod, err = wasiBuilder.Instantiate(wm.Ctx)
		if err == nil {
			// Instantiate our binary, but using the old import names.
			apiMod, err = wm.Runtime.InstantiateModule(wm.Ctx, wm.Module, wm.Config)
		}
	case modeGo:
		gojs.MustInstantiate(wm.Ctx, wm.Runtime, wm.Module)

		config := gojs.NewConfig(wm.Config)

		err = gojs.Run(wm.Ctx, wm.Runtime, wm.Module, config)
	case modeDefault:
		apiMod, err = wm.Runtime.InstantiateModule(wm.Ctx, wm.Module, wm.Config)
	}
	return
}

func WazmRun(wm *Module) (*Module, int, error) {
	wm.IsInstantiated = true
	apiMod, err := wazmInstantiate(wm)
	wm.ApiMod = apiMod
	if err != nil {
		if exitErr, ok := err.(*sys.ExitError); ok {
			exitCode := exitErr.ExitCode()
			if exitCode == sys.ExitCodeDeadlineExceeded {
				return wm, int(exitCode), fmt.Errorf("error: %v (timeout)", exitErr)
			}
			return wm, int(exitCode), nil
		}
		return wm, 1, fmt.Errorf("error instantiating wasm binary: %v", err)
	}
	return wm, 0, nil
}

const (
	modeDefault importMode = iota
	modeWasi
	modeWasiUnstable
	modeGo
)

type importMode uint

func detectImports(imports []api.FunctionDefinition) importMode {
	for _, f := range imports {
		moduleName, _, _ := f.Import()
		switch moduleName {
		case wasi_snapshot_preview1.ModuleName:
			return modeWasi
		case "wasi_unstable":
			return modeWasiUnstable
		case "go", "gojs":
			return modeGo
		}
	}
	return modeDefault
}

func maybeHostLogging(ctx context.Context, scopes logging.LogScopes, stdErr *os.File) context.Context {
	if scopes != 0 && stdErr != nil {
		return context.WithValue(ctx, experimental.FunctionListenerFactoryKey{}, logging.NewHostLoggingListenerFactory(stdErr, scopes))
	}
	return ctx
}

type logScopesFlag logging.LogScopes

func (f *logScopesFlag) String() string {
	return logging.LogScopes(*f).String()
}

func (f *logScopesFlag) Set(input string) error {
	for _, s := range strings.Split(input, ",") {
		switch s {
		case "":
			continue
		case "all":
			*f |= logScopesFlag(logging.LogScopeAll)
		case "clock":
			*f |= logScopesFlag(logging.LogScopeClock)
		case "filesystem":
			*f |= logScopesFlag(logging.LogScopeFilesystem)
		case "memory":
			*f |= logScopesFlag(logging.LogScopeMemory)
		case "proc":
			*f |= logScopesFlag(logging.LogScopeProc)
		case "poll":
			*f |= logScopesFlag(logging.LogScopePoll)
		case "random":
			*f |= logScopesFlag(logging.LogScopeRandom)
		case "sock":
			*f |= logScopesFlag(logging.LogScopeSock)
		default:
			return errors.New("not a log scope")
		}
	}
	return nil
}

func ParseHostLogging(input string) (f logScopesFlag, err error) {
	for _, s := range strings.Split(input, ",") {
		switch s {
		case "":
			continue
		case "all":
			f |= logScopesFlag(logging.LogScopeAll)
		case "clock":
			f |= logScopesFlag(logging.LogScopeClock)
		case "filesystem":
			f |= logScopesFlag(logging.LogScopeFilesystem)
		case "memory":
			f |= logScopesFlag(logging.LogScopeMemory)
		case "proc":
			f |= logScopesFlag(logging.LogScopeProc)
		case "poll":
			f |= logScopesFlag(logging.LogScopePoll)
		case "random":
			f |= logScopesFlag(logging.LogScopeRandom)
		case "sock":
			f |= logScopesFlag(logging.LogScopeSock)
		default:
			err = errors.New("not a log scope")
		}
	}
	return
}
