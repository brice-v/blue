package evaluator

import (
	"blue/ast"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/token"
	"bytes"
	"container/list"
	"embed"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/shopspring/decimal"
)

// IsEmbed is a global variable to be used to determine whether the code is on the os
// or if it has been embedded
var IsEmbed = false
var Files embed.FS

var (
	// TRUE is the true object which should be the same everywhere
	TRUE = &object.Boolean{Value: true}
	// FALSE is the false object which should be the same everywhere
	FALSE = &object.Boolean{Value: false}
	// NULL is the null object which should be the same everywhere
	NULL = &object.Null{}
	// IGNORE is the object which is used to ignore variables when necessary
	IGNORE = &object.Null{}

	// BREAK is the break object to be used the same everywhere
	BREAK = &object.BreakStatement{}
	// CONTINUE is the continue object to be used the same everywhere
	CONTINUE = &object.ContinueStatement{}
)

type Evaluator struct {
	env *object.Environment

	// PID is the process ID of this evaluator
	PID uint64

	// EvalBasePath is the base directory from which the current file is being run
	EvalBasePath string

	// CurrentFile is the file being executed (or <stdin> if run from the REPL)
	CurrentFile string

	// UFCSArg is the argument to be given to the builtin function
	UFCSArg *Stack[*object.Object]

	// Builtins is the list of builtin elements to look through based on the files imported
	Builtins *list.List

	// ErrorTokens is the set 'stack' of tokens which can get the error with file:line:col
	ErrorTokens *TokenStackSet

	// Used for: indx, elem in for expression
	nestLevel         int
	iterCount         []int
	cleanupTmpVar     map[string]int
	cleanupTmpVarIter map[string]int
	oneElementForIn   bool
	doneWithFor       bool

	isInScopeBlock bool
	scopeNestLevel int
	// scopeVars is the map of scopeNestLevel to the variables that need to be removed
	scopeVars       map[int][]string
	cleanupScopeVar map[string]bool
}

// Note: When creating multiple new evaluators with `spawn` there were race conditions
// this mostly solves that issue but parallel code still needs some work
var NewEvaluatorLock = &sync.Mutex{}

func New() *Evaluator {
	NewEvaluatorLock.Lock()
	defer NewEvaluatorLock.Unlock()
	e := &Evaluator{
		env: object.NewEnvironment(),

		PID: pidCount.Load(),

		EvalBasePath: ".",
		CurrentFile:  "<stdin>",

		UFCSArg: NewStack[*object.Object](),

		Builtins: list.New(),

		ErrorTokens: NewTokenStackSet(),

		nestLevel:         -1,
		iterCount:         []int{},
		cleanupTmpVar:     make(map[string]int),
		cleanupTmpVarIter: make(map[string]int),
		oneElementForIn:   false,
		doneWithFor:       false,

		isInScopeBlock:  false,
		scopeNestLevel:  0,
		scopeVars:       make(map[int][]string),
		cleanupScopeVar: make(map[string]bool),
	}

	builtins.Put("to_num", createToNumBuiltin(e))
	e.Builtins.PushBack(builtins)
	e.Builtins.PushBack(stringbuiltins)
	e.Builtins.PushBack(builtinobjs)
	// builtinobjs["__FILE__"] = &object.BuiltinObj{
	// 	Obj: &object.Stringo{Value: e.CurrentFile},
	// }
	e.AddCoreLibToEnv()
	// Create an empty process so we can recv without spawning
	process := &object.Process{
		Fun: nil,
		Ch:  make(chan object.Object, 1),
	}
	ProcessMap.Put(e.PID, process)

	_http_builtin_map.Put("_handle", createHttpHandleBuiltin(e))
	_http_builtin_map.Put("_handle_ws", createHttpHandleWSBuiltin(e))

	_ui_builtin_map.Put("_button", createUIButtonBuiltin(e))
	_ui_builtin_map.Put("_check_box", createUICheckBoxBuiltin(e))
	_ui_builtin_map.Put("_radio_group", createUIRadioBuiltin(e))
	_ui_builtin_map.Put("_option_select", createUIOptionSelectBuiltin(e))
	_ui_builtin_map.Put("_form", createUIFormBuiltin(e))

	return e
}

// Eval takes an ast node and returns an object
func (e *Evaluator) Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return e.evalProgram(node)
	case *ast.BreakStatement:
		return BREAK
	case *ast.ContinueStatement:
		return CONTINUE
	case *ast.ExpressionStatement:
		obj := e.Eval(node.Expression)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.Identifier:
		obj := e.evalIdentifier(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.BigIntegerLiteral:
		return &object.BigInteger{Value: node.Value}
	case *ast.HexLiteral:
		return &object.UInteger{Value: node.Value}
	case *ast.OctalLiteral:
		return &object.UInteger{Value: node.Value}
	case *ast.BinaryLiteral:
		return &object.UInteger{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.BigFloatLiteral:
		return &object.BigFloat{Value: node.Value}
	case *ast.Boolean:
		return nativeToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := e.Eval(node.Right)
		if isError(right) {
			e.ErrorTokens.Push(node.Token)
			return right
		}
		obj := e.evalPrefixExpression(node.Operator, right)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.InfixExpression:
		// If were in an `in` expression, it needs to be evaluated differently (for `for`)
		if node.Operator == "in" {
			obj := e.evalInExpression(node)
			if isError(obj) {
				e.ErrorTokens.Push(node.Token)
			}
			return obj
		}
		left := e.Eval(node.Left)
		if isError(left) {
			e.ErrorTokens.Push(node.Token)
			return left
		}
		right := e.Eval(node.Right)
		if isError(right) {
			e.ErrorTokens.Push(node.Token)
			return right
		}
		obj := e.evalInfixExpression(node.Operator, left, right)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.PostfixExpression:
		left := e.Eval(node.Left)
		if isError(left) {
			e.ErrorTokens.Push(node.Token)
			return left
		}
		obj := e.evalPostfixExpression(node.Operator, left)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.BlockStatement:
		obj := e.evalBlockStatement(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.IfExpression:
		obj := e.evalIfExpression(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.ReturnStatement:
		val := e.Eval(node.ReturnValue)
		if isError(val) {
			e.ErrorTokens.Push(node.Token)
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.ValStatement:
		val := e.Eval(node.Value)
		if isError(val) {
			e.ErrorTokens.Push(node.Token)
			return val
		}
		return e.evalVariableStatement(true, node.IsMapDestructor, node.IsListDestructor, val, node.Names, node.KeyValueNames, node.Token)
	case *ast.VarStatement:
		val := e.Eval(node.Value)
		if isError(val) {
			e.ErrorTokens.Push(node.Token)
			return val
		}
		return e.evalVariableStatement(false, node.IsMapDestructor, node.IsListDestructor, val, node.Names, node.KeyValueNames, node.Token)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		defaultParams := []object.Object{}
		for _, val := range node.ParameterExpressions {
			if val == nil {
				defaultParams = append(defaultParams, nil)
				continue
			}
			obj := e.Eval(val)
			if isError(obj) {
				e.ErrorTokens.Push(node.Token)
				return obj
			}
			defaultParams = append(defaultParams, obj)
		}
		return &object.Function{Parameters: params, Body: body, DefaultParameters: defaultParams, Env: e.env}
	case *ast.FunctionStatement:
		params := node.Parameters
		body := node.Body
		defaultParams := []object.Object{}
		for _, val := range node.ParameterExpressions {
			if val == nil {
				defaultParams = append(defaultParams, nil)
				continue
			}
			obj := e.Eval(val)
			if isError(obj) {
				e.ErrorTokens.Push(node.Token)
				return obj
			}
			defaultParams = append(defaultParams, obj)
		}
		funObj := &object.Function{Parameters: params, DefaultParameters: defaultParams, Body: body, Env: e.env}
		funObj.HelpStr = createHelpStringFromBodyTokens(node.Name.Value, funObj, body.HelpStrTokens)
		e.env.Set(node.Name.Value, funObj)
	case *ast.CallExpression:
		e.UFCSArg.Push(nil)
		function := e.Eval(node.Function)
		if isError(function) {
			e.ErrorTokens.Push(node.Token)
			return function
		}
		args := e.evalExpressions(node.Arguments)
		defaultArgs := make(map[string]object.Object)
		for k, v := range node.DefaultArguments {
			val := e.Eval(v)
			if isError(val) {
				e.ErrorTokens.Push(node.Token)
				return val
			}
			defaultArgs[k] = val
		}
		if len(args) == 1 && isError(args[0]) {
			e.ErrorTokens.Push(node.Token)
			return args[0]
		}
		val := e.applyFunction(function, args, defaultArgs)
		if isError(val) {
			e.ErrorTokens.Push(node.Token)
		}
		return val
	case *ast.StringLiteral:
		if len(node.InterpolationValues) == 0 {
			return &object.Stringo{Value: node.Value}
		}
		obj := e.evalStringWithInterpolation(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.ExecStringLiteral:
		obj := e.evalExecStringLiteral(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.ListLiteral:
		elements := e.evalExpressions(node.Elements)
		if len(elements) == 1 && isError(elements[0]) {
			e.ErrorTokens.Push(node.Token)
			return elements[0]
		}
		return &object.List{Elements: elements}
	case *ast.IndexExpression:
		left := e.Eval(node.Left)
		if isError(left) {
			e.ErrorTokens.Push(node.Token)
			return left
		}
		indx := e.Eval(node.Index)
		if isError(indx) {
			e.ErrorTokens.Push(node.Token)
			return indx
		}
		if left.Type() == object.MAP_OBJ {
			// If its a map we want to check if it actually returns NULL with the key
			obj, ok := e.evalMapIndexExpression(left.(*object.Map), indx)
			// ok means we returned NULL
			// IF we set a variable this way (as in x.println = 'something')
			// we cant call like so 'x.println()'
			if !ok {
				// If it didnt return null - we are probably using a builtin function
				// name as a key
				return obj
			}
		}
		val := e.tryCreateValidBuiltinForDotCall(left, indx, node.Left)
		if val != nil {
			return val
		}
		obj := e.evalIndexExpression(left, indx)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.MapLiteral:
		obj := e.evalMapLiteral(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.AssignmentExpression:
		obj := e.evalAssignmentExpression(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.ForExpression:
		obj := e.evalForExpression(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.ListCompLiteral:
		obj := e.evalListCompLiteral(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.MapCompLiteral:
		obj := e.evalMapCompLiteral(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.SetCompLiteral:
		obj := e.evalSetCompLiteral(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.MatchExpression:
		obj := e.evalMatchExpression(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.Null:
		return NULL
	case *ast.SetLiteral:
		obj := e.evalSetLiteral(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.ImportStatement:
		obj := e.evalImportStatement(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.TryCatchStatement:
		obj := e.evalTryCatchStatement(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.EvalExpression:
		obj := e.evalEvalExpression(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.SpawnExpression:
		obj := e.evalSpawnExpression(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	case *ast.SelfExpression:
		obj := e.evalSelfExpression(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
	default:
		if node == nil {
			// Just want to get rid of this in my output
			return nil
		}
		fmt.Printf("Handle this type: %T\n", node)
	}
	// Remove any 'errors' if we reach this because no error must've been encountered
	e.ErrorTokens.RemoveAllEntries()
	// In the event that there are only statements, I think this is where we end up
	// so we return NULL because there is nothing to return otherwise
	return NULL
}

func (e *Evaluator) evalVariableStatement(isVal, isMapDestructor, isListDestructor bool,
	val object.Object,
	names []*ast.Identifier,
	keyValueNames map[ast.Expression]*ast.Identifier,
	tok token.Token) object.Object {
	if isListDestructor {
		if val.Type() != object.LIST_OBJ {
			return newError("List Destructor must be used with list value. got=%s", val.Type())
		}
	} else if isMapDestructor {
		if val.Type() != object.MAP_OBJ {
			return newError("Map Destructor must be used with map value. got=%s", val.Type())
		}
	}
	ifNameInMapSetEnv := func(m object.OrderedMap2[object.HashKey, object.MapPair], name string) bool {
		for _, k := range m.Keys {
			mp, _ := m.Get(k)
			if mp.Key.Type() == object.STRING_OBJ {
				s := mp.Key.(*object.Stringo).Value
				if name == s {
					e.env.Set(name, mp.Value)
					return true
				}
			}
		}
		return false
	}
	for i, name := range names {
		if ok := e.env.IsImmutable(name.Value); ok {
			e.ErrorTokens.Push(tok)
			return newError("'" + name.Value + "' is already defined as immutable, cannot reassign")
		}
		if _, ok := e.env.Get(name.Value); ok {
			e.ErrorTokens.Push(tok)
			return newError("'" + name.Value + "' is already defined")
		}
		if e.isInScopeBlock {
			e.scopeVars[e.scopeNestLevel] = append(e.scopeVars[e.scopeNestLevel], name.Value)
		}
		if isVal {
			e.env.ImmutableSet(name.Value)
		}
		if isListDestructor {
			l := val.(*object.List).Elements
			if i > len(l) {
				return newError("List destructor has too many identifiers for list. len=%d", len(l))
			}
			e.env.Set(name.Value, l[i])
		} else if isMapDestructor {
			m := val.(*object.Map)
			if !ifNameInMapSetEnv(m.Pairs, name.Value) {
				return newError("Map destructor key name '%s' was not found in map", name.Value)
			}
		} else {
			e.env.Set(name.Value, val)
		}
	}

	for keyExp, name := range keyValueNames {
		ident, isIdent := keyExp.(*ast.Identifier)
		str, isStr := keyExp.(*ast.StringLiteral)
		if !isIdent && !isStr {
			return newError("Key Expression in Destructor must be STRING or IDENTIFIER. found=%T", keyExp)
		}
		var kVal object.Object
		if isIdent {
			kVal = &object.Stringo{Value: ident.Value}
		} else {
			kVal = &object.Stringo{Value: str.Value}
		}

		hk := object.HashKey{
			Type:  kVal.Type(),
			Value: object.HashObject(kVal),
		}
		if ok := e.env.IsImmutable(name.Value); ok {
			e.ErrorTokens.Push(tok)
			return newError("'" + name.Value + "' is already defined as immutable, cannot reassign")
		}
		if _, ok := e.env.Get(name.Value); ok {
			e.ErrorTokens.Push(tok)
			return newError("'" + name.Value + "' is already defined")
		}
		if e.isInScopeBlock {
			e.scopeVars[e.scopeNestLevel] = append(e.scopeVars[e.scopeNestLevel], name.Value)
		}
		if isVal {
			e.env.ImmutableSet(name.Value)
		}
		if isListDestructor {
			return newError("List Destructor should not be reached when in the Map Destructor KeyValueNames")
		} else if isMapDestructor {
			m := val.(*object.Map)
			mp, ok := m.Pairs.Get(hk)
			if !ok {
				return newError("Key Expression `%s` Not Found in Map", kVal.Inspect())
			}
			e.env.Set(name.Value, mp.Value)
		} else {
			e.env.Set(name.Value, val)
		}
	}

	e.ErrorTokens.RemoveAllEntries()
	return NULL
}

func (e *Evaluator) evalImportStatement(node *ast.ImportStatement) object.Object {
	name := node.Path.Value
	if e.IsStd(name) {
		e.AddStdLibToEnv(name)
		return NULL
	}
	fpath := e.createFilePathFromImportPath(name)
	modName := strings.ReplaceAll(filepath.Base(fpath), ".b", "")
	var inputStr string
	if !IsEmbed {
		file, err := filepath.Abs(fpath)
		if err != nil {
			return newError("Failed to import '%s'. Could not get absolute filepath.", name)
		}
		ofile, err := os.Open(file)
		if err != nil {
			return newError("Failed to import '%s'. Could not open file '%s' for reading.", name, file)
		}
		defer ofile.Close()
		fileData, err := io.ReadAll(ofile)
		if err != nil {
			return newError("Failed to import '%s'. Could not read the file.", name)
		}
		inputStr = string(fileData)
	} else {
		fileData, err := Files.ReadFile(consts.EMBED_FILES_PREFIX + fpath)
		if err != nil {
			return newError("Failed to import '%s'. Could not read the file at path '%s'.", name, fpath)
		}
		inputStr = string(fileData)
	}

	l := lexer.New(inputStr, fpath)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			splitMsg := strings.Split(msg, "\n")
			firstPart := fmt.Sprintf("%s%s\n", consts.PARSER_ERROR_PREFIX, splitMsg[0])
			consts.ErrorPrinter(firstPart)
			for i, s := range splitMsg {
				if i == 0 {
					continue
				}
				fmt.Println(s)
			}
		}
		return newError("%sFile '%s' contains Parser Errors.", consts.PARSER_ERROR_PREFIX, name)
	}
	newE := New()
	val := newE.Eval(program)
	if isError(val) {
		return val
	}

	if len(node.IdentsToImport) >= 1 {
		for _, ident := range node.IdentsToImport {
			if strings.HasPrefix(ident.Value, "_") {
				return newError("ImportError: imports must be public to import them. failed to import %s from %s", ident.Value, modName)
			}
			o, ok := newE.env.Get(ident.Value)
			if !ok {
				return newError("ImportError: failed to import %s from %s", ident.Value, modName)
			}
			e.env.Set(ident.Value, o)
		}
		// return early if we specifically import some objects
		return NULL
	} else if node.ImportAll {
		// Here we want to import everything from the module
		for k, v := range newE.env.GetAll() {
			if !strings.HasPrefix(k, "_") {
				e.env.Set(k, v)
			}
		}
		return NULL
	}
	// Set HelpStr from program HelpStrToks
	pubFunHelpStr := newE.env.GetPublicFunctionHelpString()
	mod := &object.Module{
		Name:    modName,
		Env:     newE.env,
		HelpStr: CreateHelpStringFromProgramTokens(modName, program.HelpStrTokens, pubFunHelpStr),
	}
	if node.Alias != nil {
		e.env.Set(node.Alias.Value, mod)
	} else {
		e.env.Set(modName, mod)
	}
	return NULL
}

func (e *Evaluator) evalTryCatchStatement(node *ast.TryCatchStatement) object.Object {
	evald := e.Eval(node.TryBlock)
	if isError(evald) {
		e.env.Set(node.CatchIdentifier.Value, &object.Stringo{Value: evald.Inspect()})
		evaldCatch := e.Eval(node.CatchBlock)
		// Need to remove the catch identifier after evaluating the catch block
		e.env.RemoveIdentifier(node.CatchIdentifier.Value)
		if node.FinallyBlock != nil {
			obj := e.Eval(node.FinallyBlock)
			if isError(obj) {
				e.ErrorTokens.Push(node.Token)
				return obj
			}
		}
		e.ErrorTokens.RemoveAllEntries()
		if evaldCatch == nil {
			// Set to Null so we continue in for loop if its empty
			evaldCatch = NULL
		}
		return evaldCatch
	}
	if node.FinallyBlock != nil {
		obj := e.Eval(node.FinallyBlock)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
			return obj
		}
	}
	e.ErrorTokens.RemoveAllEntries()
	return evald
}

func (e *Evaluator) evalSetLiteral(node *ast.SetLiteral) object.Object {
	elements := e.evalExpressions(node.Elements)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}

	setMap := object.NewSetElements()
	for _, e := range elements {
		hashKey := object.HashObject(e)
		setMap.Set(hashKey, object.SetPair{Value: e, Present: true})
	}
	return &object.Set{Elements: setMap}
}

func (e *Evaluator) evalEvalExpression(node *ast.EvalExpression) object.Object {
	evalStr := e.Eval(node.StrToEval)
	if evalStr.Type() != object.STRING_OBJ {
		return newError("value after `eval` must be STRING. got %s", evalStr.Type())
	}
	s := evalStr.(*object.Stringo).Value
	obj, err := e.EvalString(s)
	if err != nil {
		return newError(err.Error())
	}
	return obj
}

func (e *Evaluator) evalSpawnExpression(node *ast.SpawnExpression) object.Object {
	argLen := len(node.Arguments)
	if argLen > 2 || argLen == 0 {
		return newInvalidArgCountError("spawn", argLen, 1, "or 2")
	}
	arg0 := e.Eval(node.Arguments[0])
	if isError(arg0) {
		return arg0
	}
	if arg0.Type() != object.FUNCTION_OBJ {
		return newPositionalTypeError("spawn", 1, object.FUNCTION_OBJ, arg0.Type())
	}
	arg1 := MakeEmptyList()
	if argLen == 2 {
		arg1 = e.Eval(node.Arguments[1])
		if isError(arg1) {
			return arg1
		}
		if arg1.Type() != object.LIST_OBJ {
			return newPositionalTypeError("spawn", 2, object.LIST_OBJ, arg1.Type())
		}
	}
	fun, _ := arg0.(*object.Function)
	process := &object.Process{
		Fun: fun,
		Ch:  make(chan object.Object, 1),
	}
	pid := pidCount.Add(1)
	ProcessMap.Put(pid, process)
	go spawnFunction(pid, fun, arg1)
	return object.CreateBasicMapObject("pid", pid)
}

func spawnFunction(pid uint64, fun *object.Function, arg1 object.Object) {
	newE := New()
	newE.PID = pid
	newObj := newE.applyFunction(fun, arg1.(*object.List).Elements, make(map[string]object.Object))
	if isError(newObj) {
		err := newObj.(*object.Error)
		var buf bytes.Buffer
		buf.WriteString(err.Message)
		buf.WriteByte('\n')
		for newE.ErrorTokens.Len() > 0 {
			tok := newE.ErrorTokens.PopBack()
			buf.WriteString(fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
		}
		fmt.Printf("%s%s\n", consts.PROCESS_ERROR_PREFIX, buf.String())
	}
	// Delete from concurrent map and close channel (not 100% sure its necessary)
	if process, ok := ProcessMap.Get(pid); ok {
		close(process.Ch)
	}
	ProcessMap.Remove(pid)
}

func (e *Evaluator) evalSelfExpression(node *ast.SelfExpression) object.Object {
	return object.CreateBasicMapObject("pid", e.PID)
}

func (e *Evaluator) evalMatchExpression(node *ast.MatchExpression) object.Object {
	conditionLen := len(node.Conditions)
	consequenceLen := len(node.Consequences)
	if conditionLen != consequenceLen {
		return newError("conditions length is not equal to consequences length in match expression")
	}
	var optVal object.Object
	if node.OptionalValue != nil {
		optVal = e.Eval(node.OptionalValue)
		if isError(optVal) {
			return optVal
		}
	}
	e.env.Set("_", NULL)
	for i := 0; i < conditionLen; i++ {
		if node.Conditions[i].String() == "_" {
			// Default case should always run (even if its before others)
			return e.Eval(node.Consequences[i])
		}
		// Run through each condtion and if it evaluates to "true" then return the evaluated consequence
		condVal := e.Eval(node.Conditions[i])
		// This is our very basic form of pattern matching
		if condVal.Type() == object.MAP_OBJ && optVal != nil && optVal.Type() == object.MAP_OBJ {
			// Do our shape matching on it
			if doCondAndMatchExpEqual(condVal, optVal) {
				return e.Eval(node.Consequences[i])
			}
		}
		if optVal == nil {
			evald := e.Eval(node.Conditions[i])
			if isError(evald) {
				return evald
			}
			if evald == TRUE {
				return e.Eval(node.Consequences[i])
			}
			continue
		}
		if object.HashObject(condVal) == object.HashObject(optVal) {
			return e.Eval(node.Consequences[i])
		}
		if condVal == IGNORE {
			return e.Eval(node.Consequences[i])
		}
	}
	// Shouldnt reach here ideally
	return NULL
}

func (e *Evaluator) evalListCompLiteral(node *ast.ListCompLiteral) object.Object {
	l := lexer.New(node.NonEvaluatedProgram, "<internal: ListCompLiteral>")
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return nil
	}
	if len(p.Errors()) > 0 {
		return newError("ListCompLiteral error: %s", strings.Join(p.Errors(), " | "))
	}
	result := e.Eval(rootNode)
	if isError(result) {
		return newError("ListCompLiteral error: %s", result.(*object.Error).Message)
	}
	someVal, ok := e.env.Get("__internal__")
	if !ok {
		return nil
	}
	e.env.RemoveIdentifier("__internal__")
	return &object.ListCompLiteral{Elements: someVal.(*object.List).Elements}
}

func (e *Evaluator) evalMapCompLiteral(node *ast.MapCompLiteral) object.Object {
	l := lexer.New(node.NonEvaluatedProgram, "<internal: MapCompLiteral>")
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return nil
	}
	if len(p.Errors()) > 0 {
		return newError("MapCompLiteral error: %s", strings.Join(p.Errors(), " | "))
	}
	result := e.Eval(rootNode)
	if isError(result) {
		return newError("MapCompLiteral error: %s", result.(*object.Error).Message)
	}
	someVal, ok := e.env.Get("__internal__")
	if !ok {
		return nil
	}
	e.env.RemoveIdentifier("__internal__")
	return &object.Map{Pairs: someVal.(*object.Map).Pairs}
}

func (e *Evaluator) evalSetCompLiteral(node *ast.SetCompLiteral) object.Object {
	l := lexer.New(node.NonEvaluatedProgram, "<internal: SetCompLiteral>")
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return nil
	}
	if len(p.Errors()) > 0 {
		return newError("SetCompLiteral error: %s", strings.Join(p.Errors(), " | "))
	}
	result := e.Eval(rootNode)
	if isError(result) {
		return newError("SetCompLiteral error: %s", result.(*object.Error).Message)
	}
	someVal, ok := e.env.Get("__internal__")
	if !ok {
		return nil
	}
	e.env.RemoveIdentifier("__internal__")
	return &object.Set{Elements: someVal.(*object.Set).Elements}
}

// evalInExpression evaluates `in` statements when they refer to a loop context
func (e *Evaluator) evalInExpression(node *ast.InfixExpression) object.Object {
	ident, ok := node.Left.(*ast.Identifier)
	if ok {
		return e.evalInExpressionWithIdentOnLeft(node.Right, ident)
	}
	listWithIdents, ok := node.Left.(*ast.ListLiteral)
	if ok {
		allAreIdents := true
		for _, e := range listWithIdents.Elements {
			_, isI := e.(*ast.Identifier)
			allAreIdents = allAreIdents && isI
		}
		if allAreIdents && len(listWithIdents.Elements) == 2 {
			return e.evalInExpressionWithListOnLeft(node.Right, listWithIdents)
		}
	}
	leftEval := e.Eval(node.Left)
	if isError(leftEval) {
		return leftEval
	}
	rightEval := e.Eval(node.Right)
	if isError(rightEval) {
		return rightEval
	}
	return e.evalInfixExpression(node.Operator, leftEval, rightEval)
}

func (e *Evaluator) evalInExpressionWithIdentOnLeft(right ast.Expression, ident *ast.Identifier) object.Object {
	// So if it is an identifier than we need to find out what we are trying to
	// unpack/bind our value to
	evaluatedRight := e.Eval(right)
	if isError(evaluatedRight) {
		return evaluatedRight
	}
	_, identExists := e.env.Get(ident.Value)
	if _, ok := e.cleanupScopeVar[ident.Value]; !ok {
		e.cleanupScopeVar[ident.Value] = identExists
	}
	// IF the item existed originally then we fallback to regular evalInfixExpression code
	if existedOriginally, ok := e.cleanupScopeVar[ident.Value]; ok && existedOriginally {
		delete(e.cleanupScopeVar, ident.Value)
		evaluatedLeft := e.Eval(ident)
		if isError(evaluatedLeft) {
			return evaluatedLeft
		}
		return e.evalInfixExpression("in", evaluatedLeft, evaluatedRight)
	}
	if evaluatedRight.Type() == object.LIST_OBJ {
		// This is where we handle if its a list
		list := evaluatedRight.(*object.List).Elements
		if len(list) == 0 {
			return FALSE
		}
		if len(list) == 1 {
			e.oneElementForIn = true
		}
		defer func() {
			e.cleanupTmpVar[ident.Value] = e.nestLevel
			e.cleanupTmpVarIter[ident.Value] = len(list)
		}()
		_, ok := e.env.Get(ident.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(ident.Value, list[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(list) == e.iterCount[e.nestLevel] {
				return FALSE
			}
			return TRUE
		}
		e.env.Set(ident.Value, list[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if len(list) == e.iterCount[e.nestLevel] {
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.MAP_OBJ {
		// This is where we handle if its a Map
		mapPairs := evaluatedRight.(*object.Map).Pairs
		if mapPairs.Len() == 0 {
			return FALSE
		}
		if mapPairs.Len() == 1 {
			e.oneElementForIn = true
		}
		defer func() {
			e.cleanupTmpVar[ident.Value] = e.nestLevel
			e.cleanupTmpVarIter[ident.Value] = mapPairs.Len()
		}()
		pairObjs := make([]*object.List, mapPairs.Len())
		for i, k := range mapPairs.Keys {
			pair, _ := mapPairs.Get(k)
			listObj := []object.Object{pair.Key, pair.Value}
			pairObjs[i] = &object.List{Elements: listObj}
		}
		_, ok := e.env.Get(ident.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(ident.Value, pairObjs[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(pairObjs) == e.iterCount[e.nestLevel] {
				return FALSE
			}
			return TRUE
		}
		e.env.Set(ident.Value, pairObjs[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if mapPairs.Len() == e.iterCount[e.nestLevel] {
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.STRING_OBJ {
		// This is where we handle if its a string
		strVal := evaluatedRight.(*object.Stringo).Value
		chars := []rune(strVal)
		if len(chars) == 0 {
			return FALSE
		}
		if len(chars) == 1 {
			e.oneElementForIn = true
		}
		defer func() {
			e.cleanupTmpVar[ident.Value] = e.nestLevel
			e.cleanupTmpVarIter[ident.Value] = len(chars)
		}()
		stringObjs := make([]*object.Stringo, len(chars))
		for i, ch := range chars {
			stringObjs[i] = &object.Stringo{Value: string(ch)}
		}
		_, ok := e.env.Get(ident.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(ident.Value, stringObjs[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(chars) == e.iterCount[e.nestLevel] {
				return FALSE
			}
			return TRUE
		}
		e.env.Set(ident.Value, stringObjs[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if len(stringObjs) == e.iterCount[e.nestLevel] {
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.SET_OBJ {
		// This is where we handle if its a set
		set := evaluatedRight.(*object.Set).Elements
		if set.Len() == 0 {
			return FALSE
		}
		if set.Len() == 1 {
			e.oneElementForIn = true
		}
		defer func() {
			// Getting marked for deletion here but we only want to delete it if its a new var
			e.cleanupTmpVar[ident.Value] = e.nestLevel
			e.cleanupTmpVarIter[ident.Value] = set.Len()
		}()
		_, ok := e.env.Get(ident.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			var val object.Object
			for i, k := range set.Keys {
				if i == e.iterCount[e.nestLevel] {
					sp, _ := set.Get(k)
					if sp.Present {
						val = sp.Value
					}
				}
			}
			e.env.Set(ident.Value, val)
			e.iterCount[e.nestLevel]++
			if set.Len() == e.iterCount[e.nestLevel] {
				return FALSE
			}
			return TRUE
		}
		var val object.Object
		for i, k := range set.Keys {
			if i == e.iterCount[e.nestLevel] {
				sp, _ := set.Get(k)
				if sp.Present {
					val = sp.Value
				}
			}
		}
		e.env.Set(ident.Value, val)
		e.iterCount[e.nestLevel]++
		if set.Len() == e.iterCount[e.nestLevel] {
			return FALSE
		}
		return TRUE
	}
	return newError("Expected List, Map, Set, or String on right hand side. got=%s", evaluatedRight.Type())
}

func (e *Evaluator) evalInExpressionWithListOnLeft(right ast.Expression, listWithIdents *ast.ListLiteral) object.Object {
	// Note: We validate above that there are only 2 'ident' elements in the list
	identLeft, ok := listWithIdents.Elements[0].(*ast.Identifier)
	if !ok {
		return newError("List of Identifiers left element was not Identifier. got=%T", listWithIdents.Elements[0])
	}
	identRight, ok := listWithIdents.Elements[1].(*ast.Identifier)
	if !ok {
		return newError("List of Identifiers right element was not Identifier. got=%T", listWithIdents.Elements[1])
	}

	evaluatedRight := e.Eval(right)
	if isError(evaluatedRight) {
		return evaluatedRight
	}
	if evaluatedRight.Type() == object.LIST_OBJ {
		// This is where we handle if its a list
		list := evaluatedRight.(*object.List).Elements
		if len(list) == 0 {
			return FALSE
		}
		if len(list) == 1 {
			e.oneElementForIn = true
		}
		defer func() {
			e.cleanupTmpVar[identLeft.Value] = e.nestLevel
			e.cleanupTmpVarIter[identLeft.Value] = len(list)
			e.cleanupTmpVar[identRight.Value] = e.nestLevel
			e.cleanupTmpVarIter[identRight.Value] = len(list)
		}()
		_, ok := e.env.Get(identRight.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
			e.env.Set(identRight.Value, list[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(list) == e.iterCount[e.nestLevel] {
				return FALSE
			}
			return TRUE
		}
		e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
		e.env.Set(identRight.Value, list[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if len(list) == e.iterCount[e.nestLevel] {
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.MAP_OBJ {
		mapPairs := evaluatedRight.(*object.Map).Pairs
		if mapPairs.Len() == 0 {
			return FALSE
		}
		if mapPairs.Len() == 1 {
			e.oneElementForIn = true
		}
		defer func() {
			e.cleanupTmpVar[identLeft.Value] = e.nestLevel
			e.cleanupTmpVarIter[identLeft.Value] = mapPairs.Len()
			e.cleanupTmpVar[identRight.Value] = e.nestLevel
			e.cleanupTmpVarIter[identRight.Value] = mapPairs.Len()
		}()
		pairObjs := make([]*object.List, mapPairs.Len())
		for i, k := range mapPairs.Keys {
			pair, _ := mapPairs.Get(k)
			listObj := []object.Object{pair.Key, pair.Value}
			pairObjs[i] = &object.List{Elements: listObj}
		}
		_, ok := e.env.Get(identRight.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(identLeft.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[0])
			e.env.Set(identRight.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[1])
			e.iterCount[e.nestLevel]++
			if len(pairObjs) == e.iterCount[e.nestLevel] {
				return FALSE
			}
			return TRUE
		}
		e.env.Set(identLeft.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[0])
		e.env.Set(identRight.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[1])
		e.iterCount[e.nestLevel]++
		if mapPairs.Len() == e.iterCount[e.nestLevel] {
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.STRING_OBJ {
		// This is where we handle if its a string
		strVal := evaluatedRight.(*object.Stringo).Value
		chars := []byte(strVal)
		if len(chars) == 0 {
			return FALSE
		}
		if len(chars) == 1 {
			e.oneElementForIn = true
		}
		defer func() {
			e.cleanupTmpVar[identLeft.Value] = e.nestLevel
			e.cleanupTmpVarIter[identLeft.Value] = len(chars)
			e.cleanupTmpVar[identRight.Value] = e.nestLevel
			e.cleanupTmpVarIter[identRight.Value] = len(chars)
		}()
		stringObjs := make([]*object.Stringo, len(chars))
		for i, ch := range chars {
			stringObjs[i] = &object.Stringo{Value: string(ch)}
		}
		_, ok := e.env.Get(identRight.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
			e.env.Set(identRight.Value, stringObjs[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(chars) == e.iterCount[e.nestLevel] {
				return FALSE
			}
			return TRUE
		}
		e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
		e.env.Set(identRight.Value, stringObjs[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if len(stringObjs) == e.iterCount[e.nestLevel] {
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.SET_OBJ {
		// This is where we handle if its a set
		set := evaluatedRight.(*object.Set).Elements
		if set.Len() == 0 {
			return FALSE
		}
		if set.Len() == 1 {
			e.oneElementForIn = true
		}
		defer func() {
			e.cleanupTmpVar[identLeft.Value] = e.nestLevel
			e.cleanupTmpVarIter[identLeft.Value] = set.Len()
			e.cleanupTmpVar[identRight.Value] = e.nestLevel
			e.cleanupTmpVarIter[identRight.Value] = set.Len()
		}()
		_, ok := e.env.Get(identRight.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
			var val object.Object
			for i, k := range set.Keys {
				if i == e.iterCount[e.nestLevel] {
					sp, _ := set.Get(k)
					if sp.Present {
						val = sp.Value
					}
				}
			}
			e.env.Set(identRight.Value, val)
			e.iterCount[e.nestLevel]++
			if set.Len() == e.iterCount[e.nestLevel] {
				return FALSE
			}
			return TRUE
		}
		e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
		var val object.Object
		for i, k := range set.Keys {
			if i == e.iterCount[e.nestLevel] {
				sp, _ := set.Get(k)
				if sp.Present {
					val = sp.Value
				}
			}
		}
		e.env.Set(identRight.Value, val)
		e.iterCount[e.nestLevel]++
		if set.Len() == e.iterCount[e.nestLevel] {
			return FALSE
		}
		return TRUE
	}
	return newError("Expected List, Map, Set, or String on right hand side. got=%s", evaluatedRight.Type())
}

func (e *Evaluator) evalForExpression(node *ast.ForExpression) object.Object {
	var evalBlock object.Object
	defer func() {
		// Cleanup any temporary for variables
		doCleanup := false
		for k, v := range e.cleanupTmpVar {
			if maxIter, ok := e.cleanupTmpVarIter[k]; ok {
				if v == e.nestLevel && maxIter >= e.iterCount[e.nestLevel] && e.doneWithFor {
					e.env.RemoveIdentifier(k)
					delete(e.cleanupTmpVar, k)
					delete(e.cleanupScopeVar, k)
					delete(e.cleanupTmpVarIter, k)
					doCleanup = true
				}
			} else {
				if v > e.nestLevel {
					e.env.RemoveIdentifier(k)
					delete(e.cleanupTmpVar, k)
					delete(e.cleanupScopeVar, k)
				}
			}
		}
		if doCleanup {
			e.iterCount[e.nestLevel] = 0
			if e.nestLevel != 0 {
				if len(e.iterCount) > 1 {
					e.iterCount = e.iterCount[:len(e.iterCount)-1]
				}
				e.nestLevel--
			}
			e.doneWithFor = false
		}
	}()
	firstRun := true
	for {
		evalCond := e.Eval(node.Condition)
		if isError(evalCond) {
			return evalCond
		}
		if evalCond.Type() != object.BOOLEAN_OBJ {
			return newError("for expression condition expects BOOLEAN. got=%s", evalCond.Type())
		}
		ok := evalCond.(*object.Boolean).Value
		// If theres one element on the right hand side of a for in list expression then we dont want to return early
		if !e.oneElementForIn && !ok && firstRun {
			// If the condition is FALSE to begin with we need to return early
			// The evaluated block may not be valid in that case (ie. a list could be empty)
			e.doneWithFor = true
			return NULL
		}
		firstRun = false
		e.oneElementForIn = false
		evalBlock = e.Eval(node.Consequence)
		if evalBlock == nil {
			e.doneWithFor = true
			return NULL
		}
		if isError(evalBlock) {
			return evalBlock
		}
		if evalBlock == BREAK {
			evalBlock = NULL
			break
		}
		if evalBlock == CONTINUE && ok {
			evalBlock = NULL
			continue
		} else if evalBlock == CONTINUE && !ok {
			evalBlock = NULL
			break
		}
		rv, isReturn := evalBlock.(*object.ReturnValue)
		if isReturn {
			e.doneWithFor = true
			return rv
		}
		// Still evaluate on the last run then break if its false
		if !ok {
			break
		}
	}
	e.doneWithFor = true
	return evalBlock
}

func (e *Evaluator) evalAssignmentExpression(node *ast.AssignmentExpression) object.Object {
	left := e.Eval(node.Left)
	if isError(left) {
		return left
	}

	// If the left side contains an index expression where the identifier is immutable
	// then return an error saying so
	_, ok := node.Left.(*ast.IndexExpression)
	if ok {
		// Check the left most item in the index expression to see if it contains
		// an identifier that is immutable
		removeLeftParens := strings.ReplaceAll(node.Left.String(), "(", "")
		var rootObjIdent string
		if strings.Contains("[", removeLeftParens) {
			rootObjIdent = strings.Split(removeLeftParens, "[")[0]
		} else {
			rootObjIdent = strings.Split(removeLeftParens, ".")[0]
		}
		if ok := e.env.IsImmutable(rootObjIdent); ok {
			return newError("'" + rootObjIdent + "' is immutable")
		}
	}

	value := e.Eval(node.Value)
	if isError(value) {
		return value
	}

	// If its a simple identifier allow reassigning like so
	if ident, ok := node.Left.(*ast.Identifier); ok {
		if e.env.IsImmutable(ident.Value) {
			return newError("'" + ident.Value + "' is immutable")
		}
		switch node.Token.Literal {
		case "=":
			e.env.Set(ident.Value, value)
		case "+=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("+", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "-=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("-", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "*=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("*", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "/=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("/", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "//=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("//", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "**=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("**", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "&=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("&", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "|=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("|", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "~=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("~", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "<<=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("<<", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case ">>=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression(">>", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "%=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("%", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		case "^=":
			orig, ok := e.env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := e.evalInfixExpression("^", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			e.env.Set(ident.Value, evaluated)
		default:
			return newError("assignment operator not supported `" + node.Token.Literal + "`")
		}
	} else if ie, ok := node.Left.(*ast.IndexExpression); ok {
		// Handle Assignment to Builtin Obj
		if v, ok := ie.Left.(*ast.Identifier); ok {
			if _, ok = builtinobjs[v.Value]; ok {
				return e.evalAssignToBuiltinObj(ie, value)
			}
		}
		leftObj := e.Eval(ie.Left)
		if isError(leftObj) {
			return leftObj
		}
		if list, ok := leftObj.(*object.List); ok {
			index := e.Eval(ie.Index)
			if isError(index) {
				return index
			}

			if idx, ok := index.(*object.Integer); ok {
				switch node.Token.Literal {
				case "=":
					list.Elements[idx.Value] = value
				case "+=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("+", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "-=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("-", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "*=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("*", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "/=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("/", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "//=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("//", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "**=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("**", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "&=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("&", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "|=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("|", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "~=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("~", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "<<=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("<<", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case ">>=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression(">>", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "%=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("&", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				case "^=":
					if int(idx.Value) > len(list.Elements) || idx.Value < 0 {
						return newError("index out of bounds: %d", idx.Value)
					}
					orig := list.Elements[idx.Value]
					evaluated := e.evalInfixExpression("^", orig, value)
					if isError(evaluated) {
						return evaluated
					}
					list.Elements[idx.Value] = evaluated
				default:
					return newError("unknown assignment operator: MAP INDEX %s", node.Token.Literal)
				}
			} else {
				return newError("cannot index list with %#v", index)
			}
		} else if mapObj, ok := leftObj.(*object.Map); ok {
			key := e.Eval(ie.Index)
			if isError(key) {
				return key
			}

			if hashKey, ok := key.(object.Hashable); ok {
				hashed := hashKey.HashKey()
				switch node.Token.Literal {
				case "=":
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: value})
				case "+=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("+", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "-=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("-", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "*=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("*", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "/=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("/", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "//=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("//", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "**=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("**", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "&=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("&", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "|=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("|", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "~=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("~", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "<<=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("<<", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case ">>=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression(">>", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "%=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("&", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				case "^=":
					orig, ok := mapObj.Pairs.Get(hashed)
					if !ok {
						return newError("map key `%s` does not exist", key.Inspect())
					}
					evaluated := e.evalInfixExpression("^", orig.Value, value)
					if isError(evaluated) {
						return evaluated
					}
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
				default:
					return newError("unknown assignment operator: MAP INDEX %s", node.Token.Literal)
				}
			} else {
				return newError("cannot index map with %T", key)
			}
		} else {
			return newError("object type %T does not support item assignment", leftObj)
		}
	} else {
		return newError("expected identifier or index expression got=%T", left)
	}

	return NULL
}

func (e *Evaluator) evalAssignToBuiltinObj(ie *ast.IndexExpression, value object.Object) object.Object {
	ident, ok := ie.Left.(*ast.Identifier)
	if !ok {
		return newError("Builtin Obj was not Identifier")
	}
	var key string
	var i int64
	if ident.Value == "ENV" {
		indexStr, ok := ie.Index.(*ast.StringLiteral)
		if !ok {
			return newError("Builtin Obj Index needs to be a String. got=%T", ie.Index)
		}
		key = indexStr.Value
		if value.Type() != object.STRING_OBJ && value != NULL {
			return newError("Builtin Obj Assignment value need to be string or null. got=%s", value.Type())
		}
	} else {
		integer, ok := ie.Index.(*ast.IntegerLiteral)
		if !ok {
			return newError("Builtin Obj Index needs to be an Integer. got=%T", ie.Index)
		}
		i = integer.Value
	}
	switch ident.Value {
	case "ENV":
		if value == NULL {
			// unset the var
			err := os.Unsetenv(key)
			if err != nil {
				return newError("failed to unset ENV key '" + key + "'")
			}
		} else {
			// set the env var
			v := value.(*object.Stringo).Value
			err := os.Setenv(key, v)
			if err != nil {
				return newError("failed to set ENV key='" + key + "', value='" + v + "'")
			}
		}
		builtinobjs["ENV"].Obj = populateENVObj()
		return NULL
	case "ARGV":
		v, ok := value.(*object.Stringo)
		if !ok {
			return newError("ARGV value must be string. got=%T", value)
		}
		x := e.Eval(ie.Left)
		list, ok := x.(*object.List)
		if !ok {
			return newError("ARGV is not list. got=%T", x)
		}
		if int(i) > len(list.Elements) {
			return newError("index %d is greater than len(ARGV) of %d", i, len(list.Elements))
		}
		list.Elements = append(list.Elements[:i+1], list.Elements[i:]...)
		list.Elements[i] = v
		return list.Elements[i]
	}

	return newError("unhandled builtin obj assignment on '" + ident.Value + "'")
}

func (e *Evaluator) evalMapLiteral(node *ast.MapLiteral) object.Object {
	pairs := object.NewPairsMap()

	indices := []int{}
	for k := range node.PairsIndex {
		indices = append(indices, k)
	}
	sort.Ints(indices)
	for _, i := range indices {
		keyNode := node.PairsIndex[i]
		valueNode := node.Pairs[keyNode]
		ident, _ := keyNode.(*ast.Identifier)
		key := e.Eval(keyNode)
		if isError(key) && ident != nil {
			key = &object.Stringo{Value: ident.String()}
		} else if isError(key) {
			return key
		} else if key.Type() == object.BUILTIN_OBJ {
			key = &object.Stringo{Value: keyNode.String()}
		}

		mapKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as a map key: %s", key.Type())
		}

		value := e.Eval(valueNode)
		if isError(value) {
			return value
		}

		hashed := mapKey.HashKey()
		pairs.Set(hashed, object.MapPair{Key: key, Value: value})
	}

	return &object.Map{Pairs: pairs}
}

func (e *Evaluator) evalIndexExpression(left, indx object.Object) object.Object {
	switch {
	case left.Type() == object.LIST_OBJ:
		return e.evalListIndexExpression(left, indx)
	case left.Type() == object.SET_OBJ:
		return e.evalSetIndexExpression(left, indx)
	case left.Type() == object.MAP_OBJ:
		obj, _ := e.evalMapIndexExpression(left, indx)
		return obj
	case left.Type() == object.MODULE_OBJ:
		return e.evalModuleIndexExpression(left, indx)
	case left.Type() == object.STRING_OBJ:
		return e.evalStringIndexExpression(left, indx)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func (e *Evaluator) evalModuleIndexExpression(module, indx object.Object) object.Object {
	mod := module.(*object.Module)
	name := indx.(*object.Stringo).Value
	if strings.HasPrefix(name, "_") {
		return newError("cannot use private object '%s' from imported file '%s'", name, mod.Name)
	}
	val, ok := mod.Env.Get(name)
	if !ok {
		return newError("failed to find '%s' in imported file '%s'", name, mod.Name)
	}
	return val
}

// If this returns bool as true it means the index key was not found in the map
func (e *Evaluator) evalMapIndexExpression(mapObject, indx object.Object) (object.Object, bool) {
	mapObj := mapObject.(*object.Map)

	key, ok := indx.(object.Hashable)
	if !ok {
		return newError("unusable as a map key: %s", indx.Type()), false
	}

	pair, ok := mapObj.Pairs.Get(key.HashKey())
	if !ok {
		return NULL, true
	}

	return pair.Value, false
}

func (e *Evaluator) evalSetIndexExpression(set, indx object.Object) object.Object {
	setObj := set.(*object.Set)
	var idx int64
	switch indx.Type() {
	case object.INTEGER_OBJ:
		idx = indx.(*object.Integer).Value
	case object.STRING_OBJ:
		stringVal := indx.(*object.Stringo).Value
		envVal, ok := e.env.Get(stringVal)
		if !ok {
			return NULL
		}
		intVal, ok := envVal.(*object.Integer)
		if !ok {
			return NULL
		}
		idx = intVal.Value
	case object.LIST_OBJ:
		// Handle range expressions (1..3) or (1..<3) => they come back as a list
		indxList := indx.(*object.List).Elements
		indexes := make([]int64, len(indxList))
		for i, e := range indxList {
			if e.Type() != object.INTEGER_OBJ {
				return newError("index range needs to be INTEGER. got=%s", e.Type())
			}
			indexes[i] = e.(*object.Integer).Value
		}
		newSet := &object.Set{Elements: object.NewSetElements()}
		for _, index := range indexes {
			var j int64
			for _, k := range setObj.Elements.Keys {
				if v, ok := setObj.Elements.Get(k); ok {
					if j == index {
						newSet.Elements.Set(k, v)
					}
				}
				j++
			}
		}
		return newSet
	default:
		return newError("evalSetIndexExpression:expected index to be INT, STRING, or LIST. got=%s", indx.Type())
	}
	var i int64
	for _, k := range setObj.Elements.Keys {
		if v, ok := setObj.Elements.Get(k); ok {
			if i == idx {
				return v.Value
			}
		}
		i++
	}
	return NULL
}

func (e *Evaluator) evalListIndexExpression(list, indx object.Object) object.Object {
	listObj := list.(*object.List)
	var idx int64
	switch indx.Type() {
	case object.INTEGER_OBJ:
		idx = indx.(*object.Integer).Value
	case object.STRING_OBJ:
		stringVal := indx.(*object.Stringo).Value
		envVal, ok := e.env.Get(stringVal)
		if !ok {
			return NULL
		}
		intVal, ok := envVal.(*object.Integer)
		if !ok {
			return NULL
		}
		idx = intVal.Value
	case object.LIST_OBJ:
		// Handle range expressions (1..3) or (1..<3) => they come back as a list
		indxList := indx.(*object.List).Elements
		indexes := make([]int64, len(indxList))
		for i, e := range indxList {
			if e.Type() != object.INTEGER_OBJ {
				return newError("index range needs to be INTEGER. got=%s", e.Type())
			}
			indexes[i] = e.(*object.Integer).Value
		}
		// Support setting arbitrary index with value for list
		if listObj.Elements == nil {
			listObj.Elements = []object.Object{}
		}
		for _, index := range indexes {
			for index > int64(len(listObj.Elements)-1) {
				listObj.Elements = append(listObj.Elements, NULL)
			}
		}
		max := int64(len(listObj.Elements) - 1)
		for _, index := range indexes {
			if index < 0 || index > max {
				return newError("index out of bounds: length=%d, index=%d", len(listObj.Elements), index)
			}
		}
		newList := &object.List{Elements: make([]object.Object, len(indexes))}
		for i, index := range indexes {
			newList.Elements[i] = listObj.Elements[index]
		}
		return newList
	default:
		return NULL
	}
	// Support setting arbitrary index with value for list
	if listObj.Elements == nil {
		listObj.Elements = []object.Object{}
	}
	for idx > int64(len(listObj.Elements)-1) {
		listObj.Elements = append(listObj.Elements, NULL)
	}
	max := int64(len(listObj.Elements) - 1)
	if idx < 0 || idx > max {
		return newError("index out of bounds: length=%d, index=%d", len(listObj.Elements), idx)
	}
	return listObj.Elements[idx]
}

func (e *Evaluator) evalStringIndexExpression(str, indx object.Object) object.Object {
	strObj := str.(*object.Stringo)
	var idx int64
	switch indx.Type() {
	case object.INTEGER_OBJ:
		idx = indx.(*object.Integer).Value
	case object.STRING_OBJ:
		stringVal := indx.(*object.Stringo).Value
		envVal, ok := e.env.Get(stringVal)
		if !ok {
			return NULL
		}
		intVal, ok := envVal.(*object.Integer)
		if !ok {
			return NULL
		}
		idx = intVal.Value
	case object.LIST_OBJ:
		// Handle range expressions (1..3) or (1..<3) => they come back as a list
		indxList := indx.(*object.List).Elements
		indexes := make([]int64, len(indxList))
		for i, e := range indxList {
			if e.Type() != object.INTEGER_OBJ {
				return newError("index range needs to be INTEGER. got=%s", e.Type())
			}
			indexes[i] = e.(*object.Integer).Value
		}
		max := int64(runeLen(strObj.Value) - 1)
		for _, index := range indexes {
			if index < 0 || index > max {
				return newError("index out of bounds: length=%d, index=%d", runeLen(strObj.Value), index)
			}
		}
		newStr := make([]rune, len(indexes))
		runeStr := []rune(strObj.Value)
		for i, index := range indexes {
			newStr[i] = runeStr[index]
		}
		return &object.Stringo{Value: string(newStr)}
	default:
		return NULL
	}
	max := int64(runeLen(strObj.Value) - 1)
	if idx < 0 || idx > max {
		return newError("index out of bounds: length=%d, index=%d", runeLen(strObj.Value), idx)
	}
	return &object.Stringo{Value: string([]rune(strObj.Value)[idx])}
}

func (e *Evaluator) evalExecStringLiteral(execStringNode *ast.ExecStringLiteral) object.Object {
	str := execStringNode.Value
	return ExecStringCommand(str)
}

func (e *Evaluator) evalStringWithInterpolation(stringNode *ast.StringLiteral) object.Object {
	someObjs := e.evalExpressions(stringNode.InterpolationValues)
	if len(someObjs) == 1 && isError(someObjs[0]) {
		return someObjs[0]
	}
	newString := stringNode.Value
	for i, obj := range someObjs {
		if obj != nil {
			newString = strings.Replace(newString, stringNode.OriginalInterpolationString[i], obj.Inspect(), 1)
		} else {
			newString = strings.Replace(newString, stringNode.OriginalInterpolationString[i], "", 1)
		}
	}
	return &object.Stringo{Value: newString}
}

func (e *Evaluator) evalExpressions(exps []ast.Expression) []object.Object {
	var result []object.Object

	for _, elem := range exps {
		evaluated := e.Eval(elem)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		// Cheaty way of making list comprehensions to work
		if evaluated != nil && evaluated.Type() == object.LIST_COMP_OBJ {
			result = append(result, evaluated.(*object.ListCompLiteral).Elements...)
			return result
		}
		result = append(result, evaluated)
	}
	return result
}

func (e *Evaluator) evalIdentifier(node *ast.Identifier) object.Object {
	if val, ok := e.env.Get(node.Value); ok {
		return val
	}

	for b := e.Builtins.Front(); b != nil; b = b.Next() {
		switch t := b.Value.(type) {
		case BuiltinMapType:
			if builtin, ok := t.Get(node.Value); ok {
				return builtin
			}
		case BuiltinObjMapType:
			if builtin, ok := t[node.Value]; ok {
				return builtin.Obj
			}
		}
	}

	if node.Value == "FILE" {
		return &object.Stringo{Value: e.CurrentFile}
	}

	return newError("identifier not found: %s", node.Value)
}

func (e *Evaluator) evalIfExpression(ie *ast.IfExpression) object.Object {
	for i := 0; i < len(ie.Conditions); i++ {
		condition := e.Eval(ie.Conditions[i])
		if isError(condition) {
			return condition
		}
		if isTruthy(condition) {
			return e.Eval(ie.Consequences[i])
		}
	}
	if ie.Alternative != nil {
		return e.Eval(ie.Alternative)
	} else {
		return NULL
	}
}

func (e *Evaluator) evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	// Special cases for shift operators
	case left.Type() == object.LIST_OBJ && operator == "<<":
		l := left.(*object.List)
		l.Elements = append(l.Elements, right)
		return NULL
	case right.Type() == object.LIST_OBJ && operator == ">>":
		l := right.(*object.List)
		l.Elements = append([]object.Object{left}, l.Elements...)
		return NULL
	// These are the cases where they are the same type
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.BIG_INTEGER_OBJ:
		return e.evalBigIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return e.evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalFloatIntInfixExpression(operator, left, right)
	case left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.BIG_FLOAT_OBJ:
		return e.evalBigFloatInfixExpression(operator, left, right)
	case left.Type() == object.UINTEGER_OBJ && right.Type() == object.UINTEGER_OBJ:
		return e.evalUintInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		leftStr := left.(*object.Stringo).Value
		rightStr := right.(*object.Stringo).Value
		switch operator {
		case "+":
			return &object.Stringo{Value: leftStr + rightStr}
		case "==":
			return nativeToBooleanObject(leftStr == rightStr)
		case "!=":
			return nativeToBooleanObject(leftStr != rightStr)
		case "in":
			if strings.Contains(rightStr, leftStr) {
				return TRUE
			}
			return FALSE
		case "notin":
			if strings.Contains(rightStr, leftStr) {
				return FALSE
			}
			return TRUE
		case "..":
			if runeLen(leftStr) != 1 {
				return newError("operator .. expects left string to be 1 rune")
			}
			if runeLen(rightStr) != 1 {
				return newError("operator .. expects right string to be 1 rune")
			}
			lr := []rune(leftStr)[0]
			rr := []rune(rightStr)[0]
			if lr == rr {
				// If they are the same just return a list with the single element
				// because this is the inclusive operator
				return &object.List{Elements: []object.Object{left}}
			}
			elements := []object.Object{}
			if lr > rr {
				// Left rune is > so we are descending
				for i := lr; i >= rr; i-- {
					s := string(i)
					elements = append(elements, &object.Stringo{Value: s})
				}
				return &object.List{Elements: elements}
			} else {
				// Right rune is > so we are ascending
				for i := lr; i <= rr; i++ {
					s := string(i)
					elements = append(elements, &object.Stringo{Value: s})
				}
				return &object.List{Elements: elements}
			}
		case "..<":
			if runeLen(leftStr) != 1 {
				return newError("operator ..< expects left string to be 1 rune")
			}
			if runeLen(rightStr) != 1 {
				return newError("operator ..< expects right string to be 1 rune")
			}
			lr := []rune(leftStr)[0]
			rr := []rune(rightStr)[0]
			if lr == rr {
				// If they are the same just return an empty list because this is non-inclusive
				return &object.List{Elements: []object.Object{}}
			}
			elements := []object.Object{}
			if lr > rr {
				// Left rune is > so we are descending
				for i := lr; i > rr; i-- {
					s := string(i)
					elements = append(elements, &object.Stringo{Value: s})
				}
				return &object.List{Elements: elements}
			} else {
				// Right rune is > so we are ascending
				for i := lr; i < rr; i++ {
					s := string(i)
					elements = append(elements, &object.Stringo{Value: s})
				}
				return &object.List{Elements: elements}
			}
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case left.Type() == object.LIST_OBJ && right.Type() == object.LIST_OBJ:
		leftListObj := left.(*object.List)
		rightListObj := right.(*object.List)
		leftElements := leftListObj.Elements
		rightElements := rightListObj.Elements
		leftSize := len(leftElements)
		rightSize := len(rightElements)
		switch operator {
		case "+":
			newList := make([]object.Object, 0, leftSize+rightSize)
			newList = append(newList, leftElements...)
			newList = append(newList, rightElements...)
			return &object.List{Elements: newList}
		case "==":
			if object.HashObject(leftListObj) == object.HashObject(rightListObj) {
				return TRUE
			}
			return FALSE
		case "!=":
			if object.HashObject(leftListObj) == object.HashObject(rightListObj) {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case left.Type() == object.SET_OBJ && right.Type() == object.SET_OBJ:
		return e.evalSetInfixExpression(operator, left, right)
	case left.Type() == object.BYTES_OBJ && right.Type() == object.BYTES_OBJ:
		return e.evalBytesInfixExpression(operator, left, right)
	// These are the cases where they differ
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.BIG_INTEGER_OBJ:
		return e.evalFloatBigIntegerInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.BIG_INTEGER_OBJ:
		return e.evalIntegerBigIntegerInfixExpression(operator, left, right)
	case left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return e.evalBigIntegerFloatInfixExpression(operator, left, right)
	case left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalBigIntegerIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.BIG_FLOAT_OBJ:
		return e.evalFloatBigFloatInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.BIG_FLOAT_OBJ:
		return e.evalIntegerBigFloatInfixExpression(operator, left, right)
	case left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return e.evalBigFloatFloatInfixExpression(operator, left, right)
	case left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalBigFloatIntegerInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return e.evalIntegerFloatInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.UINTEGER_OBJ:
		return e.evalIntegerUintegerInfixExpression(operator, left, right)
	case right.Type() == object.INTEGER_OBJ && left.Type() == object.UINTEGER_OBJ:
		return e.evalIntegerUintegerInfixExpression(operator, right, left)
	case left.Type() == object.STRING_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalStringIntegerInfixExpression(operator, left, right)
	case right.Type() == object.STRING_OBJ && left.Type() == object.INTEGER_OBJ:
		return e.evalStringIntegerInfixExpression(operator, right, left)
	case left.Type() == object.STRING_OBJ && right.Type() == object.UINTEGER_OBJ:
		return e.evalStringUintegerInfixExpression(operator, left, right)
	case right.Type() == object.STRING_OBJ && left.Type() == object.UINTEGER_OBJ:
		return e.evalStringUintegerInfixExpression(operator, right, left)
	case left.Type() == object.LIST_OBJ && !isBooleanOperator(operator):
		return e.evalListInfixExpression(operator, left, right)
	case right.Type() == object.SET_OBJ && !isBooleanOperator(operator):
		return e.evalRightSideSetInfixExpression(operator, left, right)
	default:
		return e.evalDefaultInfixExpression(operator, left, right)
	}
}

func (e *Evaluator) evalDefaultInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case operator == "==":
		return nativeToBooleanObject(object.HashObject(left) == object.HashObject(right))
	case operator == "!=":
		return nativeToBooleanObject(object.HashObject(left) != object.HashObject(right))
	case operator == "and":
		leftBool, ok := left.(*object.Boolean)
		if !ok {
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
		rightBool, ok := right.(*object.Boolean)
		if !ok {
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
		if leftBool.Value && rightBool.Value {
			return TRUE
		}
		return FALSE
	case operator == "or":
		leftBool, ok := left.(*object.Boolean)
		if !ok {
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
		rightBool, ok := right.(*object.Boolean)
		if !ok {
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
		if leftBool.Value || rightBool.Value {
			return TRUE
		}
		return FALSE
	case (operator == "in" || operator == "notin") && (right.Type() == object.LIST_OBJ || right.Type() == object.SET_OBJ || right.Type() == object.MAP_OBJ):
		return e.evalInOrNotinInfixExpression(operator, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalInOrNotinInfixExpression(operator string, left, right object.Object) object.Object {
	leftHash := object.HashObject(left)
	switch rt := right.(type) {
	case *object.List:
		if operator == "in" {
			for _, e := range rt.Elements {
				if leftHash == object.HashObject(e) {
					return TRUE
				}
			}
			return FALSE
		} else if operator == "notin" {
			for _, e := range rt.Elements {
				if leftHash == object.HashObject(e) {
					return FALSE
				}
			}
			return TRUE
		}
	case *object.Set:
		if operator == "in" {
			for _, k := range rt.Elements.Keys {
				if leftHash == k {
					return TRUE
				}
			}
			return FALSE
		} else if operator == "notin" {
			for _, k := range rt.Elements.Keys {
				if leftHash == k {
					return FALSE
				}
			}
			return TRUE
		}
	case *object.Map:
		if operator == "in" {
			for _, k := range rt.Pairs.Keys {
				if leftHash == k.Value {
					return TRUE
				}
			}
			return FALSE
		} else if operator == "notin" {
			for _, k := range rt.Pairs.Keys {
				if leftHash == k.Value {
					return FALSE
				}
			}
			return TRUE
		}
	}
	return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
}

func (e *Evaluator) evalRightSideSetInfixExpression(operator string, left, right object.Object) object.Object {
	setElems := right.(*object.Set).Elements
	switch left.Type() {
	case object.INTEGER_OBJ:
		intVal := object.HashObject(left.(*object.Integer))
		switch operator {
		case "in":
			if _, ok := setElems.Get(intVal); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(intVal); ok {
				return FALSE
			}
			return TRUE
		default:
			return e.evalDefaultInfixExpression(operator, left, right)
		}
	case object.UINTEGER_OBJ:
		uintVal := left.(*object.UInteger).HashKey().Value
		switch operator {
		case "in":
			if _, ok := setElems.Get(uintVal); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(uintVal); ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case object.FUNCTION_OBJ:
		funHash := object.HashObject(left.(*object.Function))
		switch operator {
		case "in":
			if _, ok := setElems.Get(funHash); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(funHash); ok {
				return FALSE
			}
			return TRUE
		default:
			return e.evalDefaultInfixExpression(operator, left, right)
		}
	case object.MAP_OBJ:
		mapHash := object.HashObject(left.(*object.Map))
		switch operator {
		case "in":
			if _, ok := setElems.Get(mapHash); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(mapHash); ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case object.BOOLEAN_OBJ:
		boolHash := left.(*object.Boolean).HashKey().Value
		switch operator {
		case "in":
			if _, ok := setElems.Get(boolHash); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(boolHash); ok {
				return FALSE
			}
			return TRUE
		default:
			return e.evalDefaultInfixExpression(operator, left, right)
		}
	case object.STRING_OBJ:
		strHash := left.(*object.Stringo).HashKey().Value
		switch operator {
		case "in":
			if _, ok := setElems.Get(strHash); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(strHash); ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case object.NULL_OBJ:
		nullHash := object.HashObject(left.(*object.Null))
		switch operator {
		case "in":
			if _, ok := setElems.Get(nullHash); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(nullHash); ok {
				return FALSE
			}
			return TRUE
		default:
			return e.evalDefaultInfixExpression(operator, left, right)
		}
	case object.LIST_OBJ:
		listHash := object.HashObject(left.(*object.List))
		switch operator {
		case "in":
			if _, ok := setElems.Get(listHash); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(listHash); ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case object.BIG_FLOAT_OBJ:
		bigFloat := object.HashObject(left.(*object.BigFloat))
		switch operator {
		case "in":
			if _, ok := setElems.Get(bigFloat); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(bigFloat); ok {
				return FALSE
			}
			return TRUE
		default:
			return e.evalDefaultInfixExpression(operator, left, right)
		}
	case object.BIG_INTEGER_OBJ:
		bigInt := object.HashObject(left.(*object.BigInteger))
		switch operator {
		case "in":
			if _, ok := setElems.Get(bigInt); ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems.Get(bigInt); ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	default:
		return e.evalDefaultInfixExpression(operator, left, right)
	}
}

func (e *Evaluator) evalBytesInfixExpression(operator string, left, right object.Object) object.Object {
	leftBs := left.(*object.Bytes).Value
	rightBs := right.(*object.Bytes).Value
	switch operator {
	case "&":
		if len(leftBs) != len(rightBs) {
			return newError("length of left and right bytes must match to perform bitwise AND operation. got: len(l)=%d, len(r)=%d", len(leftBs), len(rightBs))
		}
		buf := make([]byte, len(leftBs))
		for i := 0; i < len(leftBs); i++ {
			buf[i] = leftBs[i] & rightBs[i]
		}
		return &object.Bytes{Value: buf}
	case "|":
		if len(leftBs) != len(rightBs) {
			return newError("length of left and right bytes must match to perform bitwise OR operation. got: len(l)=%d, len(r)=%d", len(leftBs), len(rightBs))
		}
		buf := make([]byte, len(leftBs))
		for i := 0; i < len(leftBs); i++ {
			buf[i] = leftBs[i] | rightBs[i]
		}
		return &object.Bytes{Value: buf}
	case "^":
		if len(leftBs) != len(rightBs) {
			return newError("length of left and right bytes must match to perform bitwise XOR operation. got: len(l)=%d, len(r)=%d", len(leftBs), len(rightBs))
		}
		buf := make([]byte, len(leftBs))
		for i := 0; i < len(leftBs); i++ {
			buf[i] = leftBs[i] ^ rightBs[i]
		}
		return &object.Bytes{Value: buf}
	default:
		return e.evalDefaultInfixExpression(operator, left, right)
	}
}

func (e *Evaluator) evalSetInfixExpression(operator string, left, right object.Object) object.Object {
	leftE := left.(*object.Set).Elements
	rightE := right.(*object.Set).Elements
	newSet := &object.Set{Elements: object.NewSetElements()}
	leftElems := object.NewSetElements()
	rightElems := object.NewSetElements()
	if leftE.Len() >= rightE.Len() {
		leftElems = leftE
		rightElems = rightE
	} else {
		leftElems = rightE
		rightElems = leftE
	}
	switch operator {
	case "|":
		// union
		for _, k := range leftElems.Keys {
			v, ok := leftElems.Get(k)
			if !ok {
				continue
			}
			newSet.Elements.Set(k, v)
		}
		for _, k := range rightElems.Keys {
			v, ok := rightElems.Get(k)
			if !ok {
				continue
			}
			newSet.Elements.Set(k, v)
		}
		return newSet
	case "&":
		// intersect
		for _, k := range leftElems.Keys {
			v, ok := leftElems.Get(k)
			if !ok {
				continue
			}
			v1, ok := rightElems.Get(k)
			if !ok {
				continue
			}
			if v1.Present {
				newSet.Elements.Set(k, v)
			}
		}
		return newSet
	case "^":
		// symmetric difference
		for _, k := range leftElems.Keys {
			v, ok := leftElems.Get(k)
			if !ok {
				continue
			}
			v1, ok := rightElems.Get(k)
			if !ok {
				newSet.Elements.Set(k, v)
			}
			if !v1.Present {
				newSet.Elements.Set(k, v)
			}
		}
		for _, k := range rightElems.Keys {
			v, ok := rightElems.Get(k)
			if !ok {
				continue
			}
			v1, ok := leftElems.Get(k)
			if !ok {
				newSet.Elements.Set(k, v)
			}
			if !v1.Present {
				newSet.Elements.Set(k, v)
			}
		}
		return newSet
	case ">=":
		// left is superset of right
		for _, k := range rightE.Keys {
			if _, ok := leftE.Get(k); !ok {
				return FALSE
			}
		}
		return TRUE
	case "<=":
		// right is a superset of left
		for _, k := range leftE.Keys {
			if _, ok := rightE.Get(k); !ok {
				return FALSE
			}
		}
		return TRUE
	case "-":
		// difference
		for _, k := range leftElems.Keys {
			v, ok := leftElems.Get(k)
			if !ok {
				continue
			}
			v1, ok := rightElems.Get(k)
			if !ok {
				newSet.Elements.Set(k, v)
			}
			if !v1.Present {
				newSet.Elements.Set(k, v)
			}
		}
		return newSet
	case "==":
		for _, k := range leftElems.Keys {
			v1, ok := rightElems.Get(k)
			if !ok {
				return FALSE
			}
			if !v1.Present {
				return FALSE
			}
		}
		return TRUE
	case "!=":
		for _, k := range leftElems.Keys {
			v1, ok := rightElems.Get(k)
			if !ok {
				return TRUE
			}
			if !v1.Present {
				return TRUE
			}
		}
		return FALSE
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalBigFloatIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	rightVal := right.(*object.Integer).Value
	rightBigFloat := decimal.NewFromInt(rightVal)
	return e.evalBigFloatInfixExpression(operator, left, &object.BigFloat{Value: rightBigFloat})
}

func (e *Evaluator) evalBigFloatFloatInfixExpression(operator string, left, right object.Object) object.Object {
	rightVal := right.(*object.Float).Value
	rightBigFloat := decimal.NewFromFloat(rightVal)
	return e.evalBigFloatInfixExpression(operator, left, &object.BigFloat{Value: rightBigFloat})
}

func (e *Evaluator) evalIntegerBigFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	leftBigFloat := decimal.NewFromInt(leftVal)
	return e.evalBigFloatInfixExpression(operator, &object.BigFloat{Value: leftBigFloat}, right)
}

func (e *Evaluator) evalFloatBigFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	leftBigFloat := decimal.NewFromFloat(leftVal)
	return e.evalBigFloatInfixExpression(operator, &object.BigFloat{Value: leftBigFloat}, right)
}

func (e *Evaluator) evalBigIntegerIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	rightVal := right.(*object.Integer).Value
	rightBigInt := new(big.Int).SetInt64(rightVal)
	return e.evalBigIntegerInfixExpression(operator, left, &object.BigInteger{Value: rightBigInt})
}

func (e *Evaluator) evalBigIntegerFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.BigInteger).Value
	rightVal := right.(*object.Float).Value
	leftBigFloat := decimal.NewFromBigInt(leftVal, 1)
	rightBigFloat := decimal.NewFromFloat(rightVal)
	return e.evalBigFloatInfixExpression(operator, &object.BigFloat{Value: leftBigFloat}, &object.BigFloat{Value: rightBigFloat})
}

func (e *Evaluator) evalIntegerBigIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	leftBigInt := new(big.Int).SetInt64(leftVal)
	return e.evalBigIntegerInfixExpression(operator, &object.BigInteger{Value: leftBigInt}, right)
}

func (e *Evaluator) evalFloatBigIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.BigInteger).Value
	leftBigFloat := decimal.NewFromFloat(leftVal)
	rightBigFloat := decimal.NewFromBigInt(rightVal, 1)
	return e.evalBigFloatInfixExpression(operator, &object.BigFloat{Value: leftBigFloat}, &object.BigFloat{Value: rightBigFloat})
}

func (e *Evaluator) evalBigFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.BigFloat).Value
	rightVal := right.(*object.BigFloat).Value
	switch operator {
	case "+":
		return &object.BigFloat{Value: leftVal.Add(rightVal)}
	case "-":
		return &object.BigFloat{Value: leftVal.Sub(rightVal)}
	case "/":
		return &object.BigFloat{Value: leftVal.Div(rightVal)}
	case "*":
		return &object.BigFloat{Value: leftVal.Mul(rightVal)}
	case "**":
		return &object.BigFloat{Value: leftVal.Pow(rightVal)}
	case "//":
		return &object.BigFloat{Value: leftVal.Div(rightVal).Floor()}
	case "%":
		return &object.BigFloat{Value: leftVal.Mod(rightVal)}
	case "<":
		compared := leftVal.Cmp(rightVal)
		if compared == -1 {
			return TRUE
		}
		return FALSE
	case ">":
		compared := leftVal.Cmp(rightVal)
		if compared == 1 {
			return TRUE
		}
		return FALSE
	case "<=":
		compared := leftVal.Cmp(rightVal)
		if compared == -1 || compared == 0 {
			return TRUE
		}
		return FALSE
	case ">=":
		compared := leftVal.Cmp(rightVal)
		if compared == 1 || compared == 0 {
			return TRUE
		}
		return FALSE
	case "==":
		compared := leftVal.Cmp(rightVal)
		if compared == 0 {
			return TRUE
		}
		return FALSE
	case "!=":
		compared := leftVal.Cmp(rightVal)
		if compared != 0 {
			return TRUE
		}
		return FALSE
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalBigIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.BigInteger).Value
	rightVal := right.(*object.BigInteger).Value
	result := big.NewInt(0)
	switch operator {
	case "+":
		return &object.BigInteger{Value: result.Add(leftVal, rightVal)}
	case "-":
		return &object.BigInteger{Value: result.Sub(leftVal, rightVal)}
	case "/":
		return &object.BigInteger{Value: result.Div(leftVal, rightVal)}
	case "*":
		return &object.BigInteger{Value: result.Mul(leftVal, rightVal)}
	case "**":
		return &object.BigInteger{Value: result.Exp(leftVal, rightVal, nil)}
	case "//":
		maybeWanted := new(big.Int)
		floored, modulus := result.DivMod(leftVal, rightVal, maybeWanted)
		fmt.Printf("TODO: FIGURE OUT WHAT WE WANT TO DO WITH THIS %v", modulus)
		return &object.BigInteger{Value: floored}
	case "%":
		return &object.BigInteger{Value: result.Mod(leftVal, rightVal)}
	case "<":
		compared := leftVal.Cmp(rightVal)
		if compared == -1 {
			return TRUE
		}
		return FALSE
	case ">":
		compared := leftVal.Cmp(rightVal)
		if compared == 1 {
			return TRUE
		}
		return FALSE
	case "<=":
		compared := leftVal.Cmp(rightVal)
		if compared == -1 || compared == 0 {
			return TRUE
		}
		return FALSE
	case ">=":
		compared := leftVal.Cmp(rightVal)
		if compared == 1 || compared == 0 {
			return TRUE
		}
		return FALSE
	case "==":
		compared := leftVal.Cmp(rightVal)
		if compared == 0 {
			return TRUE
		}
		return FALSE
	case "!=":
		compared := leftVal.Cmp(rightVal)
		if compared != 0 {
			return TRUE
		}
		return FALSE
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalStringIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Stringo).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "*":
		var out bytes.Buffer
		var i int64
		for i = 0; i < rightVal; i++ {
			out.WriteString(leftVal)
		}
		return &object.Stringo{Value: out.String()}
	default:
		return e.evalDefaultInfixExpression(operator, left, right)
	}
}

func (e *Evaluator) evalStringUintegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Stringo).Value
	rightVal := right.(*object.UInteger).Value
	switch operator {
	case "*":
		var out bytes.Buffer
		var i uint64
		for i = 0; i < rightVal; i++ {
			out.WriteString(leftVal)
		}
		return &object.Stringo{Value: out.String()}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalIntegerUintegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftIntVal := left.(*object.Integer).Value
	if leftIntVal < 0 {
		return newError("Negative Integers are not allowed for Unsigned Integer operations")
	}
	leftVal := uint64(left.(*object.Integer).Value)
	rightVal := right.(*object.UInteger).Value
	switch operator {
	case "+":
		// Note: i think overflow is okay when dealing with unsigned
		return &object.UInteger{Value: leftVal + rightVal}
	case "-":
		// Note i think ill allow underflow when dealing with unsigned
		return &object.UInteger{Value: leftVal - rightVal}
	case "/":
		return &object.UInteger{Value: leftVal / rightVal}
	case "*":
		return &object.UInteger{Value: leftVal * rightVal}
	case "**":
		return &object.UInteger{Value: uint64(math.Pow(float64(leftVal), float64(rightVal)))}
	case "//":
		return &object.UInteger{Value: uint64(math.Floor(float64(leftVal) / float64(rightVal)))}
	case "%":
		return &object.UInteger{Value: uint64(math.Mod(float64(leftVal), float64(rightVal)))}
	case "<":
		return nativeToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalIntegerFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := float64(left.(*object.Integer).Value)
	rightVal := right.(*object.Float).Value
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "//":
		return &object.Float{Value: math.Floor(leftVal / rightVal)}
	case "%":
		return &object.Float{Value: math.Mod(leftVal, rightVal)}
	case "<":
		return nativeToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalListInfixExpression(operator string, left, right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER_OBJ:
		return e.evalListIntegerInfixExpression(operator, left, right)
	case object.LIST_OBJ:
		return e.evalListListInfixExpression(operator, left, right)
	case object.SET_OBJ:
		return e.evalRightSideSetInfixExpression(operator, left, right)
	default:
		return newError("unhandled type for list infix expressions: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalListListInfixExpression(operator string, left, right object.Object) object.Object {
	listLeftObj := left.(*object.List)
	listRightObj := right.(*object.List)
	listLeftElems := listLeftObj.Elements
	listRightElems := listRightObj.Elements
	listLenLeft := int64(len(listLeftElems))
	listLenRight := int64(len(listRightElems))
	switch operator {
	case "+":
		newSize := listLenLeft + listLenRight
		// this creates a new list with capacity of newSize but a length of 0 (we dont want it filling with nil)
		newList := make([]object.Object, 0, newSize)
		newList = append(newList, listLeftElems...)
		newList = append(newList, listRightElems...)
		return &object.List{Elements: newList}
	case "!=":
		return nativeToBooleanObject(!twoListsEqual(listLeftObj, listRightObj))
	case "==":
		return nativeToBooleanObject(twoListsEqual(listLeftObj, listRightObj))
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalListIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	listObj := left.(*object.List).Elements
	intObj := right.(*object.Integer).Value
	switch operator {
	case "*":
		listLen := int64(len(listObj))
		newSize := listLen * intObj
		// this creates a new list with capacity of newSize but a length of 0 (we dont want it filling with nil)
		newList := make([]object.Object, 0, newSize)
		for i := 0; int64(i) < intObj; i++ {
			newList = append(newList, listObj...)
		}
		return &object.List{Elements: newList}
	default:
		return e.evalDefaultInfixExpression(operator, left, right)
	}
}

func (e *Evaluator) evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "not":
		return e.evalNotOperatorExpression(right)
	case "-":
		return e.evalMinusPrefixOperatorExpression(right)
	case "~":
		return e.evalBitwiseNotOperatorExpression(right)
	case "<<":
		return e.evalLshiftPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func (e *Evaluator) evalPostfixExpression(operator string, left object.Object) object.Object {
	switch operator {
	case ">>":
		return e.evalRshiftPostfixExpression(left)
	default:
		return newError("unknown operator: %s%s", left.Type(), operator)
	}
}

func (e *Evaluator) evalRshiftPostfixExpression(left object.Object) object.Object {
	switch left.Type() {
	case object.LIST_OBJ:
		l := left.(*object.List)
		listLen := len(l.Elements)
		if listLen == 0 {
			return NULL
		}
		e := l.Elements[listLen-1]
		if listLen == 1 {
			l.Elements = []object.Object{}
		} else {
			l.Elements = l.Elements[0 : listLen-1]
		}
		return e
	default:
		return newError("unknown operator: %s >>", left.Type())
	}
}

func (e *Evaluator) evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	checkOverflow := func(leftVal, rightVal int64) bool {
		result := leftVal + rightVal
		return result-leftVal != rightVal
	}
	checkUnderflow := func(leftVal, rightVal int64) bool {
		result := leftVal - rightVal
		return result+rightVal != leftVal
	}

	checkOverflowMul := func(leftVal, rightVal int64) bool {
		if leftVal == 0 || rightVal == 0 || leftVal == 1 || rightVal == 1 {
			return false
		}
		if leftVal == math.MinInt64 || rightVal == math.MinInt64 {
			return true
		}
		result := leftVal * rightVal
		return result/rightVal != leftVal
	}

	checkOverflowPow := func(leftVal, rightVal int64) bool {
		if leftVal == 0 || rightVal == 0 || leftVal == 1 || rightVal == 1 {
			return false
		}
		if leftVal == math.MinInt64 || rightVal == math.MinInt64 {
			return true
		}
		if rightVal > 63 && leftVal > 1 {
			return true
		}
		return false
	}

	switch operator {
	case "+":
		overflowed := checkOverflow(leftVal, rightVal)
		if overflowed {
			left := new(big.Int).SetInt64(leftVal)
			right := new(big.Int).SetInt64(rightVal)
			result := big.NewInt(0)
			return &object.BigInteger{Value: result.Add(left, right)}
		}
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		underflowed := checkUnderflow(leftVal, rightVal)
		if underflowed {
			left := new(big.Int).SetInt64(leftVal)
			right := new(big.Int).SetInt64(rightVal)
			result := big.NewInt(0)
			return &object.BigInteger{Value: result.Sub(left, right)}
		}
		return &object.Integer{Value: leftVal - rightVal}
	case "/":
		if rightVal == 0 {
			return newError("Division by zero is not allowed")
		}
		if rightVal > leftVal {
			return &object.Integer{Value: 0}
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "*":
		overflowed := checkOverflowMul(leftVal, rightVal)
		if overflowed {
			left := new(big.Int).SetInt64(leftVal)
			right := new(big.Int).SetInt64(rightVal)
			result := big.NewInt(0)
			return &object.BigInteger{Value: result.Mul(left, right)}
		}
		return &object.Integer{Value: leftVal * rightVal}
	case "**":
		overflowed := checkOverflowPow(leftVal, rightVal)
		if overflowed {
			left := new(big.Int).SetInt64(leftVal)
			right := new(big.Int).SetInt64(rightVal)
			result := big.NewInt(0)
			return &object.BigInteger{Value: result.Exp(left, right, nil)}
		}
		return &object.Integer{Value: int64(math.Pow(float64(leftVal), float64(rightVal)))}
	case "//":
		if rightVal == 0 {
			return newError("Floor Division by zero is not allowed")
		}
		if rightVal > leftVal {
			return &object.Integer{Value: 0}
		}
		return &object.Integer{Value: int64(leftVal / rightVal)}
	case "%":
		if rightVal == 0 {
			return newError("Modulus by zero is not allowed")
		}
		if leftVal < 0 || rightVal < 0 {
			left := new(big.Int).SetInt64(leftVal)
			right := new(big.Int).SetInt64(rightVal)
			result := big.NewInt(0)
			return &object.BigInteger{Value: result.Mod(left, right)}
		}
		return &object.Integer{Value: int64(math.Mod(float64(leftVal), float64(rightVal)))}
	case "<":
		return nativeToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeToBooleanObject(leftVal != rightVal)
	case "..":
		return e.evalIntegerRange(leftVal, rightVal)
	case "..<":
		return e.evalIntegerNonIncRange(leftVal, rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalIntegerRange(leftVal, rightVal int64) object.Object {
	var i int64

	if leftVal < rightVal {
		size := rightVal - leftVal
		listElems := make([]object.Object, 0, size)
		for i = leftVal; i <= rightVal; i++ {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	} else if rightVal < leftVal {
		size := leftVal - rightVal
		listElems := make([]object.Object, 0, size)
		for i = leftVal; i >= rightVal; i-- {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	}
	// When they are equal just return a value (leftVal in this case)
	return &object.List{Elements: []object.Object{&object.Integer{Value: leftVal}}}
}

func (e *Evaluator) evalIntegerNonIncRange(leftVal, rightVal int64) object.Object {
	var i int64

	if leftVal < rightVal {
		size := rightVal - leftVal
		listElems := make([]object.Object, 0, size-1)
		for i = leftVal; i < rightVal; i++ {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	} else if rightVal < leftVal {
		size := leftVal - rightVal
		listElems := make([]object.Object, 0, size-1)
		for i = leftVal; i > rightVal; i-- {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	}
	return &object.List{Elements: []object.Object{}}
}

func (e *Evaluator) evalFloatIntInfixExpression(operator string, left, right object.Object) object.Object {
	// Note: this may cause errors to print incorrectly but this is a quick way to solve
	// for this use case
	return e.evalIntegerFloatInfixExpression(operator, right, left)
}

func (e *Evaluator) evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "//":
		return &object.Float{Value: float64(int64(leftVal) / int64(rightVal))}
	case "%":
		return &object.Float{Value: math.Mod(leftVal, rightVal)}
	case "<":
		return nativeToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalUintInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.UInteger).Value
	rightVal := right.(*object.UInteger).Value

	switch operator {
	case "+":
		return &object.UInteger{Value: leftVal + rightVal}
	case "-":
		return &object.UInteger{Value: leftVal - rightVal}
	case "/":
		return &object.UInteger{Value: leftVal / rightVal}
	case "*":
		return &object.UInteger{Value: leftVal * rightVal}
	case "**":
		return &object.UInteger{Value: uint64(math.Pow(float64(leftVal), float64(rightVal)))}
	case "&":
		return &object.UInteger{Value: leftVal & rightVal}
	case "|":
		return &object.UInteger{Value: leftVal | rightVal}
	case "^":
		return &object.UInteger{Value: leftVal ^ rightVal}
	case ">>":
		return &object.UInteger{Value: leftVal >> rightVal}
	case "<<":
		return &object.UInteger{Value: leftVal << rightVal}
	case "<":
		return nativeToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeToBooleanObject(leftVal > rightVal)
	case "<=":
		return nativeToBooleanObject(leftVal <= rightVal)
	case ">=":
		return nativeToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() == object.INTEGER_OBJ {
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	}
	if right.Type() == object.FLOAT_OBJ {
		value := right.(*object.Float).Value
		return &object.Float{Value: -value}
	}
	if right.Type() != object.FLOAT_OBJ || right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}
	return newError("unknown operator: -%s", right.Type())
}

func (e *Evaluator) evalBitwiseNotOperatorExpression(right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER_OBJ:
		value := right.(*object.Integer).Value
		return &object.Integer{Value: ^value}
	case object.UINTEGER_OBJ:
		value := right.(*object.UInteger).Value
		return &object.UInteger{Value: ^value}
	case object.BYTES_OBJ:
		value := right.(*object.Bytes).Value
		buf := make([]byte, len(value))
		for i, b := range value {
			buf[i] = ^b
		}
		return &object.Bytes{Value: buf}
	default:
		return newError("unknown operator: ~%s", right.Type())
	}
}

func (e *Evaluator) evalNotOperatorExpression(right object.Object) object.Object {
	// here we are defining what happend on an object when the not operator is used on it
	// to check if a list is empty we would need to put something to check it here
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func (e *Evaluator) evalLshiftPrefixOperatorExpression(right object.Object) object.Object {
	switch right.Type() {
	case object.LIST_OBJ:
		l := right.(*object.List)
		listLen := len(l.Elements)
		if listLen == 0 {
			return NULL
		}
		e := l.Elements[0]
		if listLen == 1 {
			l.Elements = []object.Object{}
		} else {
			l.Elements = l.Elements[1:listLen]
		}
		return e
	default:
		return newError("unknown operator: << %s", right.Type())
	}
}

func (e *Evaluator) evalProgram(program *ast.Program) object.Object {
	var result object.Object

	for _, stmt := range program.Statements {
		result = e.Eval(stmt)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func (e *Evaluator) evalBlockStatement(block *ast.BlockStatement) object.Object {
	var result object.Object

	e.isInScopeBlock = true
	e.scopeNestLevel++

	defer func() {
		e.isInScopeBlock = false
		if vars, ok := e.scopeVars[e.scopeNestLevel]; ok {
			// Cleanup any temporary for variables
			for _, v := range vars {
				if existedOriginally, ok := e.cleanupScopeVar[v]; ok && !existedOriginally {
					e.env.RemoveIdentifier(v)
					delete(e.cleanupTmpVar, v)
					delete(e.cleanupScopeVar, v)
				} else {
					e.env.RemoveIdentifier(v)
				}
			}
			delete(e.scopeVars, e.scopeNestLevel)
		}
		e.scopeNestLevel--
	}()

	for _, stmt := range block.Statements {
		result = e.Eval(stmt)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
			if result == BREAK {
				return BREAK
			}
			if result == CONTINUE {
				return CONTINUE
			}
		}
	}
	return result
}
