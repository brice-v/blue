package evaluator

import (
	"blue/ast"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/token"
	"blue/util"
	"bytes"
	"embed"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/shopspring/decimal"
)

// IsEmbed is a global variable to be used to determine whether the code is on the os
// or if it has been embedded
var IsEmbed = false
var Files embed.FS

// NoExec is a global to prevent execution of shell commands on the system
var NoExec = false

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
	// NodeName is the name of this node
	NodeName string

	// EvalBasePath is the base directory from which the current file is being run
	EvalBasePath string

	// CurrentFile is the file being executed (or <stdin> if run from the REPL)
	CurrentFile string

	// UFCSArg is the argument to be given to the function
	UFCSArg *util.Stack[*object.Object]
	// UFCSArgIsImmutable determines whether the arg passed in to function is immutable
	UFCSArgIsImmutable *util.Stack[bool]

	// Builtins is the list of builtin elements to look through based on the files imported
	Builtins []BuiltinMapType
	// BuiltinObjs is the list of builtin elements to look through based on the files imported
	BuiltinObjs []BuiltinObjMapType

	// ErrorTokens is the set 'stack' of tokens which can get the error with file:line:col
	ErrorTokens        *TokenStackSet
	maybeNullMapFnCall *util.Stack[string]

	// Used for: indx, elem in for expression
	nestLevel         int
	iterCount         []int
	cleanupTmpVar     map[string]int
	cleanupTmpVarIter map[string]int
	oneElementForIn   bool
	doneWithFor       map[int]struct{}

	isInScopeBlock map[int]struct{}
	scopeNestLevel int
	// scopeVars is the map of scopeNestLevel to the variables that need to be removed
	scopeVars       map[int][]string
	cleanupScopeVar map[string]bool
	evalingNodeCond map[int]struct{}

	// deferFuns is the map of scopeNestLevel function to execute at scope block cleanup
	deferFuns map[int]*util.Stack[*FunAndArgs]

	inComprehensionLiteral bool
}

type FunAndArgs struct {
	Fun  *object.Function
	Args []object.Object
}

const (
	cEvalBasePath = "."
	cCurrentFile  = "<stdin>"
)

// Note: When creating multiple new evaluators with `spawn` there were race conditions
// this now only seems necessary for add std lib to env
var NewEvaluatorLock = &sync.Mutex{}

func New() *Evaluator {
	return NewNode("", "")
}

func NewNode(nodeName, address string) *Evaluator {
	e := &Evaluator{
		env: object.NewEnvironmentWithoutCore(),

		PID:      pidCount.Load(),
		NodeName: nodeName,

		EvalBasePath: cEvalBasePath,
		CurrentFile:  cCurrentFile,

		UFCSArg:            util.NewStack[*object.Object](),
		UFCSArgIsImmutable: util.NewStack[bool](),

		ErrorTokens:        NewTokenStackSet(),
		maybeNullMapFnCall: util.NewStack[string](),

		nestLevel:         -1,
		iterCount:         []int{},
		cleanupTmpVar:     make(map[string]int),
		cleanupTmpVarIter: make(map[string]int),
		oneElementForIn:   false,
		doneWithFor:       make(map[int]struct{}),

		isInScopeBlock:  make(map[int]struct{}),
		scopeNestLevel:  0,
		scopeVars:       make(map[int][]string),
		cleanupScopeVar: make(map[string]bool),
		evalingNodeCond: make(map[int]struct{}),

		deferFuns: make(map[int]*util.Stack[*FunAndArgs]),

		inComprehensionLiteral: false,
	}

	e.Builtins = []BuiltinMapType{
		GetBuiltins(e),
		stringbuiltins,
	}
	e.BuiltinObjs = []BuiltinObjMapType{builtinobjs}
	e.env.SetCore(e.AddCoreLibToEnv())
	// Create an empty process so we can recv without spawning
	process := &object.Process{
		// TODO: Eventually update to non-buffered and update send and recv as needed
		Ch: make(chan object.Object, 1),
		Id: e.PID,

		NodeName: nodeName,
	}
	ProcessMap.LoadOrStore(pk(nodeName, e.PID), process)

	return e
}

func (e *Evaluator) ReplEnvAdd(varName string, o object.Object) {
	e.env.Set(varName, o)
	e.env.ImmutableSet(varName)
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
	case *ast.UIntegerLiteral:
		return &object.UInteger{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.BigFloatLiteral:
		return &object.BigFloat{Value: node.Value}
	case *ast.StructLiteral:
		values := e.evalExpressions(node.Values)
		if len(values) == 1 && isError(values[0]) {
			e.ErrorTokens.Push(node.Token)
			return values[0]
		}
		sl, err := object.NewBlueStruct(node.Fields, values)
		if err != nil {
			e.ErrorTokens.Push(node.Token)
			return newError("%s", err.Error())
		}
		return sl
	case *ast.Boolean:
		return nativeToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := e.Eval(node.Right)
		if isError(right) {
			e.ErrorTokens.Push(node.Token)
			return right
		}
		obj := e.evalPrefixExpression(node.Operator, right, node.Right)
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
		// Implement Shortcuts for Truthiness in Infix Expressions
		if ((node.Operator == "&&" || node.Operator == "and") && left == FALSE) ||
			((node.Operator == "||" || node.Operator == "or") && left == TRUE) {
			return left
		}
		right := e.Eval(node.Right)
		if isError(right) {
			e.ErrorTokens.Push(node.Token)
			return right
		}
		if node.Operator == ">>" || node.Operator == "<<" {
			operator := node.Operator
			switch {
			// Special cases for shift operators
			case left.Type() == object.LIST_OBJ && operator == "<<":
				if ident, ok := node.Left.(*ast.Identifier); ok {
					if e.env.IsImmutable(ident.Value) {
						return newError("'%s' is immutable", ident.Value)
					}
				}
				l := left.(*object.List)
				l.Elements = append(l.Elements, right)
				return NULL
			case right.Type() == object.LIST_OBJ && operator == ">>":
				if ident, ok := node.Right.(*ast.Identifier); ok {
					if e.env.IsImmutable(ident.Value) {
						return newError("'%s' is immutable", ident.Value)
					}
				}
				l := right.(*object.List)
				l.Elements = append([]object.Object{left}, l.Elements...)
				return NULL
			case left.Type() == object.SET_OBJ && operator == "<<":
				if ident, ok := node.Left.(*ast.Identifier); ok {
					if e.env.IsImmutable(ident.Value) {
						return newError("'%s' is immutable", ident.Value)
					}
				}
				s := left.(*object.Set)
				key := object.HashObject(right)
				if _, ok := s.Elements.Get(key); ok {
					// If obj exists do nothing
					return NULL
				}
				s.Elements.Set(key, object.SetPair{Value: right, Present: struct{}{}})
				return NULL
			case right.Type() == object.SET_OBJ && operator == ">>":
				if ident, ok := node.Right.(*ast.Identifier); ok {
					if e.env.IsImmutable(ident.Value) {
						return newError("'%s' is immutable", ident.Value)
					}
				}
				s := right.(*object.Set)
				key := object.HashObject(left)
				if _, ok := s.Elements.Get(key); ok {
					// If obj exists do nothing
					return NULL
				}
				s.Elements.Set(key, object.SetPair{Value: left, Present: struct{}{}})
				return NULL
			}
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
	case *ast.ForStatement:
		obj := e.evalForStatement(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		return obj
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
		// Note: Clone is really slow
		return &object.Function{Parameters: params, Body: body, DefaultParameters: defaultParams, Env: e.env.Clone()}
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
		// Note: Clone is really slow
		funObj := &object.Function{Parameters: params, DefaultParameters: defaultParams, Body: body, Env: e.env}
		funObj.HelpStr = createHelpStringFromBodyTokens(node.Name.Value, funObj, body.HelpStrTokens)
		e.env.SetFunStatementAndHelp(node.Name.Value, funObj)
	case *ast.CallExpression:
		e.UFCSArg.Push(nil)
		e.UFCSArgIsImmutable.Push(false)
		function := e.Eval(node.Function)
		if isError(function) {
			e.ErrorTokens.Push(node.Token)
			return function
		}
		args := e.evalExpressions(node.Arguments)
		immutableArgs := make([]bool, len(node.Arguments))
		for i, arg := range node.Arguments {
			if ident, ok := arg.(*ast.Identifier); ok {
				immutableArgs[i] = e.env.IsImmutable(ident.Value)
			} else {
				immutableArgs[i] = false
			}
		}
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
		argElem := e.UFCSArg.Peek()
		if argElem != nil {
			isImmutable := e.UFCSArgIsImmutable.Peek()
			immutableArgs = append([]bool{isImmutable}, immutableArgs...)
		}
		val := e.applyFunction(function, args, defaultArgs, immutableArgs)
		if isError(val) {
			e.ErrorTokens.Push(node.Token)
		}
		return val
	case *ast.RegexLiteral:
		r, err := regexp.Compile(node.Token.Literal)
		if err != nil {
			return newError("failed to create regex literal %q", node.TokenLiteral())
		}
		return &object.Regex{Value: r}
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
		// Skip this call to optimize when calling with struct or process
		if left.Type() != object.BLUE_STRUCT_OBJ && left.Type() != object.PROCESS_OBJ {
			val := e.tryCreateValidDotCall(left, indx, node.Left)
			if val != nil {
				return val
			}
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
	case *ast.ListCompLiteral:
		e.inComprehensionLiteral = true
		obj := e.evalListCompLiteral(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		e.inComprehensionLiteral = false
		return obj
	case *ast.MapCompLiteral:
		e.inComprehensionLiteral = true
		obj := e.evalMapCompLiteral(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		e.inComprehensionLiteral = false
		return obj
	case *ast.SetCompLiteral:
		e.inComprehensionLiteral = true
		obj := e.evalSetCompLiteral(node)
		if isError(obj) {
			e.ErrorTokens.Push(node.Token)
		}
		e.inComprehensionLiteral = false
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
	case *ast.DeferExpression:
		obj := e.evalDeferExpression(node)
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
			return newError("'%s' is already defined as immutable, cannot reassign", name.Value)
		}
		if _, ok := e.env.Get(name.Value); ok {
			e.ErrorTokens.Push(tok)
			return newError("'%s' is already defined", name.Value)
		}
		if _, ok := e.isInScopeBlock[e.scopeNestLevel]; ok {
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
			return newError("'%s' is already defined as immutable, cannot reassign", name.Value)
		}
		if _, ok := e.env.Get(name.Value); ok {
			e.ErrorTokens.Push(tok)
			return newError("'%s' is already defined", name.Value)
		}
		if _, ok := e.isInScopeBlock[e.scopeNestLevel]; ok {
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
		return e.AddStdLibToEnv(name, node.IdentsToImport, node.ImportAll)
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
		newE.env.SetAllPublicOnEnv(e.env)
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
	evald := e.evalBlockStatement(node.TryBlock)
	if isError(evald) {
		e.env.Set(node.CatchIdentifier.Value, &object.Stringo{Value: evald.Inspect()})
		evaldCatch := e.evalBlockStatement(node.CatchBlock)
		// Need to remove the catch identifier after evaluating the catch block
		e.env.RemoveIdentifier(node.CatchIdentifier.Value)
		if node.FinallyBlock != nil {
			obj := e.Eval(node.FinallyBlock)
			if isError(obj) {
				e.ErrorTokens.Push(node.Token)
				return obj
			}
		}
		// TODO: Make try blocks point to the right place on errors
		// e.ErrorTokens.RemoveAllEntries()
		// Removing this does show where it happens in try and catch block but I'd like to put a ^ on catch ident as well
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
		setMap.Set(hashKey, object.SetPair{Value: e, Present: struct{}{}})
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
		return newError("%s", err.Error())
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
	fun := arg0.(*object.Function)
	pid := pidCount.Add(1)
	process := &object.Process{
		Id: pid,
		// TODO: Eventually update to non-buffered and update send and recv as needed
		Ch: make(chan object.Object, 1),

		NodeName: e.NodeName,
	}
	ProcessMap.Store(pk(e.NodeName, pid), process)
	go spawnFunction(pid, e.NodeName, fun, arg1)
	return process
}

func spawnFunction(pid uint64, nodeName string, fun *object.Function, arg1 object.Object) {
	newE := New()
	newE.PID = pid
	elems := arg1.(*object.List).Elements
	newObj := newE.applyFunctionFast(fun, elems, make(map[string]object.Object), make([]bool, len(elems)))
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
	if process, ok := ProcessMap.LoadAndDelete(pk(nodeName, pid)); ok {
		close(process.Ch)
	}
}

func (e *Evaluator) evalDeferExpression(node *ast.DeferExpression) object.Object {
	argLen := len(node.Arguments)
	if argLen > 2 || argLen == 0 {
		return newInvalidArgCountError("defer", argLen, 1, "or 2")
	}
	arg0 := e.Eval(node.Arguments[0])
	if isError(arg0) {
		return arg0
	}
	if arg0.Type() != object.FUNCTION_OBJ {
		return newPositionalTypeError("defer", 1, object.FUNCTION_OBJ, arg0.Type())
	}
	arg1 := MakeEmptyList()
	if argLen == 2 {
		arg1 = e.Eval(node.Arguments[1])
		if isError(arg1) {
			return arg1
		}
		if arg1.Type() != object.LIST_OBJ {
			return newPositionalTypeError("defer", 2, object.LIST_OBJ, arg1.Type())
		}
	}
	fun := arg0.(*object.Function)
	if _, ok := e.deferFuns[e.scopeNestLevel]; !ok {
		// Initialize map if its not there yet
		e.deferFuns[e.scopeNestLevel] = util.NewStack[*FunAndArgs]()
	}
	e.deferFuns[e.scopeNestLevel].Push(&FunAndArgs{Fun: fun, Args: arg1.(*object.List).Elements})
	return NULL
}

func (e *Evaluator) evalSelfExpression(_ *ast.SelfExpression) object.Object {
	if p, ok := ProcessMap.Load(pk(e.NodeName, e.PID)); ok {
		return p
	}
	return newError("`self` error: process not found")
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
			return e.evalBlockStatement(node.Consequences[i])
		}
		// Run through each condtion and if it evaluates to "true" then return the evaluated consequence
		condVal := e.Eval(node.Conditions[i])
		// This is our very basic form of pattern matching
		if condVal.Type() == object.MAP_OBJ && optVal != nil && optVal.Type() == object.MAP_OBJ {
			// Do our shape matching on it
			if doCondAndMatchExpEqual(condVal, optVal) {
				return e.evalBlockStatement(node.Consequences[i])
			}
		}
		if optVal == nil {
			evald := e.Eval(node.Conditions[i])
			if isError(evald) {
				return evald
			}
			if evald == TRUE {
				return e.evalBlockStatement(node.Consequences[i])
			}
			continue
		}
		if object.HashObject(condVal) == object.HashObject(optVal) {
			return e.evalBlockStatement(node.Consequences[i])
		}
		if condVal == IGNORE {
			return e.evalBlockStatement(node.Consequences[i])
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
	defer e.env.RemoveIdentifier("__internal__")
	if isError(result) {
		return newError("ListCompLiteral error: %s", result.(*object.Error).Message)
	}
	someVal, ok := e.env.Get("__internal__")
	if !ok {
		return nil
	}
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
	defer e.env.RemoveIdentifier("__internal__")
	if isError(result) {
		return newError("MapCompLiteral error: %s", result.(*object.Error).Message)
	}
	someVal, ok := e.env.Get("__internal__")
	if !ok {
		return nil
	}
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
	defer e.env.RemoveIdentifier("__internal__")
	if isError(result) {
		return newError("SetCompLiteral error: %s", result.(*object.Error).Message)
	}
	someVal, ok := e.env.Get("__internal__")
	if !ok {
		return nil
	}
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
	identValue, identExists := e.env.Get(ident.Value)
	if _, ok := e.cleanupScopeVar[ident.Value]; !ok {
		e.cleanupScopeVar[ident.Value] = identExists
	}
	if _, ok := e.evalingNodeCond[e.nestLevel]; !ok && identExists {
		return e.evalDefaultInfixExpression("in", identValue, evaluatedRight)
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
			return nativeToBooleanObject(len(list) != e.iterCount[e.nestLevel])
		}
		e.env.Set(ident.Value, list[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		return nativeToBooleanObject(len(list) != e.iterCount[e.nestLevel])
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
			return nativeToBooleanObject(len(pairObjs) != e.iterCount[e.nestLevel])
		}
		e.env.Set(ident.Value, pairObjs[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		return nativeToBooleanObject(mapPairs.Len() != e.iterCount[e.nestLevel])
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
			return nativeToBooleanObject(len(chars) != e.iterCount[e.nestLevel])
		}
		e.env.Set(ident.Value, stringObjs[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		return nativeToBooleanObject(len(stringObjs) != e.iterCount[e.nestLevel])
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
					if sp, ok := set.Get(k); ok {
						val = sp.Value
					}
				}
			}
			e.env.Set(ident.Value, val)
			e.iterCount[e.nestLevel]++
			return nativeToBooleanObject(set.Len() != e.iterCount[e.nestLevel])
		}
		var val object.Object
		for i, k := range set.Keys {
			if i == e.iterCount[e.nestLevel] {
				if sp, ok := set.Get(k); ok {
					val = sp.Value
				}
			}
		}
		e.env.Set(ident.Value, val)
		e.iterCount[e.nestLevel]++
		return nativeToBooleanObject(set.Len() != e.iterCount[e.nestLevel])
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
			return nativeToBooleanObject(len(list) != e.iterCount[e.nestLevel])
		}
		e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
		e.env.Set(identRight.Value, list[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		return nativeToBooleanObject(len(list) != e.iterCount[e.nestLevel])
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
			return nativeToBooleanObject(len(pairObjs) != e.iterCount[e.nestLevel])
		}
		e.env.Set(identLeft.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[0])
		e.env.Set(identRight.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[1])
		e.iterCount[e.nestLevel]++
		return nativeToBooleanObject(mapPairs.Len() != e.iterCount[e.nestLevel])
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
			return nativeToBooleanObject(len(chars) != e.iterCount[e.nestLevel])
		}
		e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
		e.env.Set(identRight.Value, stringObjs[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		return nativeToBooleanObject(len(stringObjs) != e.iterCount[e.nestLevel])
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
					if sp, ok := set.Get(k); ok {
						val = sp.Value
					}
				}
			}
			e.env.Set(identRight.Value, val)
			e.iterCount[e.nestLevel]++
			return nativeToBooleanObject(set.Len() != e.iterCount[e.nestLevel])
		}
		e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
		var val object.Object
		for i, k := range set.Keys {
			if i == e.iterCount[e.nestLevel] {
				if sp, ok := set.Get(k); ok {
					val = sp.Value
				}
			}
		}
		e.env.Set(identRight.Value, val)
		e.iterCount[e.nestLevel]++
		return nativeToBooleanObject(set.Len() != e.iterCount[e.nestLevel])
	}
	return newError("Expected List, Map, Set, or String on right hand side. got=%s", evaluatedRight.Type())
}

func (e *Evaluator) evalForStatement(node *ast.ForStatement) object.Object {
	var evalBlock object.Object
	defer func() {
		// Cleanup any temporary for variables
		doCleanup := false
		for k, v := range e.cleanupTmpVar {
			if maxIter, ok := e.cleanupTmpVarIter[k]; ok {
				_, maybeDoneWithFor := e.doneWithFor[e.scopeNestLevel]
				if v == e.nestLevel && maxIter >= e.iterCount[e.nestLevel] && maybeDoneWithFor {
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
			delete(e.doneWithFor, e.scopeNestLevel)
		}
	}()
	firstRun := true
	initalizerRan := false
	scopeToClear := -1
	defer func() {
		if node.UsesVar && initalizerRan {
			for _, k := range e.scopeVars[scopeToClear] {
				e.env.RemoveIdentifier(k)
			}
			delete(e.scopeVars, scopeToClear)
			delete(e.isInScopeBlock, scopeToClear)
			e.scopeNestLevel--
		}
	}()
	for {
		if node.UsesVar && !initalizerRan {
			val := e.Eval(node.Initializer.Value)
			if isError(val) {
				e.ErrorTokens.Push(node.Token)
				return val
			}
			e.scopeNestLevel++
			scopeToClear = e.scopeNestLevel
			e.isInScopeBlock[e.scopeNestLevel] = struct{}{}
			initalizerStmt := e.evalVariableStatement(false, node.Initializer.IsMapDestructor,
				node.Initializer.IsListDestructor, val, node.Initializer.Names, node.Initializer.KeyValueNames, node.Initializer.Token)
			if isError(initalizerStmt) {
				return initalizerStmt
			}
			initalizerRan = true
		}
		if !node.UsesVar {
			e.evalingNodeCond[e.nestLevel] = struct{}{}
		}
		evalCond := e.Eval(node.Condition)
		if !node.UsesVar {
			delete(e.evalingNodeCond, e.nestLevel)
		}
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
			// Note: There was a bug where we were setting the scopeNestLevel to be cleaned up even when it wasnt ready
			// see test_for_scope_still_broken.b from d3p2 of aoc2023, specifically for (for x in ...) when nested heavily
			// the for var x = 0; ... loops seemed to avoid this issue due to a different way of setting e.doneWithFor
			e.doneWithFor[e.scopeNestLevel+1] = struct{}{}
			return NULL
		}
		firstRun = false
		e.oneElementForIn = false
		evalBlock = e.evalBlockStatement(node.Consequence)
		if evalBlock == nil {
			e.doneWithFor[e.scopeNestLevel] = struct{}{}
			return NULL
		}
		if isError(evalBlock) {
			return evalBlock
		}
		rv, isReturn := evalBlock.(*object.ReturnValue)
		if isReturn {
			e.doneWithFor[e.scopeNestLevel] = struct{}{}
			return rv
		}
		if evalBlock == BREAK {
			evalBlock = NULL
			break
		}
		if node.UsesVar {
			canFastCalc, orig, increment, condResult := canFastCalcNodeCondAndPostExp(e, node.Condition, node.PostExp)
			if !canFastCalc {
				result := e.Eval(node.PostExp)
				if isError(result) {
					return result
				}
			} else {
				orig.Value += increment
			}
			var ok bool
			if canFastCalc && condResult != -1 {
				if condResult == 0 {
					ok = true
				} else if condResult == 1 {
					ok = false
				}
			} else {
				// Check the eval cond at the end of our post exp in case we need to exit
				evalCond := e.Eval(node.Condition)
				if isError(evalCond) {
					return evalCond
				}
				if evalCond.Type() != object.BOOLEAN_OBJ {
					return newError("for expression condition expects BOOLEAN. got=%s", evalCond.Type())
				}
				ok = evalCond.(*object.Boolean).Value
			}
			if !ok {
				e.doneWithFor[e.scopeNestLevel] = struct{}{}
				return NULL
			}
		}
		if evalBlock == CONTINUE && ok {
			evalBlock = NULL
			continue
		} else if evalBlock == CONTINUE && !ok {
			evalBlock = NULL
			break
		}
		// Still evaluate on the last run then break if its false
		if !ok {
			break
		}
	}
	e.doneWithFor[e.scopeNestLevel] = struct{}{}
	return evalBlock
}

func canFastCalcNodeCondAndPostExp(e *Evaluator, cond, postExp ast.Expression) (bool, *object.Integer, int64, int) {
	if e.inComprehensionLiteral {
		return false, nil, 0, -1
	}
	aExp, isAExp := postExp.(*ast.AssignmentExpression)
	if !isAExp {
		return false, nil, 0, -1
	}
	li, isLi := aExp.Left.(*ast.Identifier)
	if !isLi {
		return false, nil, 0, -1
	}
	il, isIl := aExp.Value.(*ast.IntegerLiteral)
	if !isIl {
		return false, nil, 0, -1
	}
	op := aExp.Token.Literal
	if op != "+=" {
		// Support other ones later on
		return false, nil, 0, -1
	}
	infix, isInfix := cond.(*ast.InfixExpression)
	if !isInfix {
		return false, nil, 0, -1
	}
	condInfixOp := infix.Operator
	if condInfixOp != "<" && condInfixOp != ">" && condInfixOp != ">=" && condInfixOp != "<=" {
		return false, nil, 0, -1
	}

	existingObj := e.evalIdentifier(li)
	if isError(existingObj) {
		return false, nil, 0, -1
	}
	existingI, isInteger := existingObj.(*object.Integer)
	if !isInteger {
		return false, nil, 0, -1
	}

	condResult := -1 // -1 will mean ignore, 0 will be true, 1 will be false
	// if leftIdent, leftIsIdent := infix.Left.(*ast.Identifier); leftIsIdent && leftIdent.Value == li.Value {
	// 	if rightIl, rightIsIl := infix.Right.(*ast.IntegerLiteral); rightIsIl {
	// 		switch condInfixOp {
	// 		case "<":
	// 			x := int(existingI.Value) < int(rightIl.Value)
	// 			if x {
	// 				condResult = 0
	// 			} else {
	// 				condResult = 1
	// 			}
	// 		case ">":
	// 			x := int(existingI.Value) > int(rightIl.Value)
	// 			if x {
	// 				condResult = 0
	// 			} else {
	// 				condResult = 1
	// 			}
	// 		case "<=":
	// 			x := int(existingI.Value) <= int(rightIl.Value)
	// 			if x {
	// 				condResult = 0
	// 			} else {
	// 				condResult = 1
	// 			}
	// 		case ">=":
	// 			x := int(existingI.Value) >= int(rightIl.Value)
	// 			if x {
	// 				condResult = 0
	// 			} else {
	// 				condResult = 1
	// 			}
	// 		}
	// 	}
	// }
	// if rightIdent, rightIsIdent := infix.Right.(*ast.Identifier); rightIsIdent && rightIdent.Value == li.Value {
	// 	if leftIl, leftIsIl := infix.Left.(*ast.IntegerLiteral); leftIsIl {
	// 		switch condInfixOp {
	// 		case "<":
	// 			x := int(leftIl.Value) < int(existingI.Value)
	// 			if x {
	// 				condResult = 0
	// 			} else {
	// 				condResult = 1
	// 			}
	// 		case ">":
	// 			x := int(leftIl.Value) > int(existingI.Value)
	// 			if x {
	// 				condResult = 0
	// 			} else {
	// 				condResult = 1
	// 			}
	// 		case "<=":
	// 			x := int(leftIl.Value) <= int(existingI.Value)
	// 			if x {
	// 				condResult = 0
	// 			} else {
	// 				condResult = 1
	// 			}
	// 		case ">=":
	// 			x := int(leftIl.Value) >= int(existingI.Value)
	// 			if x {
	// 				condResult = 0
	// 			} else {
	// 				condResult = 1
	// 			}
	// 		}
	// 	}
	// }

	return true, existingI, il.Value, condResult

	// 	2025/03/25 22:13:04 cond (*ast.InfixExpression) = (a < 5)
	// 2025/03/25 22:13:04 postExp (*ast.AssignmentExpression) = a += 1
	// log.Printf("cond (%T) = %s", cond, cond.String())
	// log.Printf("postExp (%T) = %s", postExp, postExp.String())
	// return false
}

// isRootIndexObjectImmutable checks if the left side contains an index expression
// within a loop and if the root of it is an ident, return whether that ident
// is immutable and the ident
func (e *Evaluator) isRootIndexObjectImmutable(ie *ast.IndexExpression) (*ast.Identifier, bool) {
	var left ast.Expression = ie.Left
	for {
		if leftIndex, isIndexExp := left.(*ast.IndexExpression); isIndexExp {
			left = leftIndex.Left
			continue
		}
		if ident, isIdent := left.(*ast.Identifier); isIdent {
			return ident, e.env.IsImmutable(ident.Value)
		}
		break
	}
	return nil, false
}

func (e *Evaluator) evalAssignmentExpression(node *ast.AssignmentExpression) object.Object {
	// If its a simple identifier allow reassigning like so
	if ident, ok := node.Left.(*ast.Identifier); ok {
		orig := e.evalIdentifier(ident)
		if isError(orig) {
			return orig
		}
		if e.env.IsImmutable(ident.Value) {
			return newError("'%s' is immutable", ident.Value)
		}
		operator := node.Token.Literal
		if operator == "=" {
			value := e.Eval(node.Value)
			if isError(value) {
				return value
			}
			e.env.Set(ident.Value, value)
			return NULL
		} else if operator == "+=" {
			// TODO: If I want to use this, I likely will need to optimize directly in evalForStatement

			// This is a fast pass optimization - can be likely be updated to support others as well
			// When we have a literal on the right hand side and original is something we can deal with
			// This avoids call to set which is _super_ helpful
			// if oi, ok := orig.(*object.Integer); ok {
			// 	if il, ok1 := node.Value.(*ast.IntegerLiteral); ok1 {
			// 		oi.Value = oi.Value + il.Value
			// 		return NULL
			// 	}
			// }
		}
		value := e.Eval(node.Value)
		if isError(value) {
			return value
		}
		evaluated := e.evalMultiCharAssignmentInfixExpression(operator, "IDENT", orig, value)
		if isError(evaluated) {
			return evaluated
		}
		e.env.Set(ident.Value, evaluated)
	} else if ie, ok := node.Left.(*ast.IndexExpression); ok {
		if rootIdent, ok := e.isRootIndexObjectImmutable(ie); ok {
			return newError("'%s' is immutable", rootIdent)
		}
		value := e.Eval(node.Value)
		if isError(value) {
			return value
		}
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
			idx, ok := index.(*object.Integer)
			if !ok {
				return newError("cannot index list with %#v", index)
			}
			operator := node.Token.Literal
			indexInt := int(idx.Value)
			listLen := len(list.Elements)
			if indexInt > listLen || indexInt < 0 {
				return newError("index out of bounds: %d", idx.Value)
			}
			if indexInt == listLen {
				list.Elements = append(list.Elements, NULL)
			}
			if operator == "=" {
				list.Elements[idx.Value] = value
				return NULL
			}
			orig := list.Elements[idx.Value]
			evaluated := e.evalMultiCharAssignmentInfixExpression(operator, object.LIST_OBJ, orig, value)
			if isError(evaluated) {
				return evaluated
			}
			list.Elements[idx.Value] = evaluated
		} else if mapObj, ok := leftObj.(*object.Map); ok {
			key := e.Eval(ie.Index)
			if isError(key) {
				return key
			}

			if ok := object.IsHashable(key); ok {
				hk := object.HashObject(key)
				hashed := object.HashKey{Type: key.Type(), Value: hk}
				operator := node.Token.Literal
				if operator == "=" {
					mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: value})
					return NULL
				}
				origMapPair, ok := mapObj.Pairs.Get(hashed)
				if !ok {
					return newError("map key `%s` does not exist", key.Inspect())
				}
				evaluated := e.evalMultiCharAssignmentInfixExpression(operator, object.MAP_OBJ, origMapPair.Value, value)
				if isError(evaluated) {
					return evaluated
				}
				mapObj.Pairs.Set(hashed, object.MapPair{Key: key, Value: evaluated})
			} else {
				return newError("cannot index map with %T", key)
			}
		} else if str, ok := leftObj.(*object.Stringo); ok && value.Type() == object.STRING_OBJ {
			index := e.Eval(ie.Index)
			if isError(index) {
				return index
			}
			s := str.Value
			c := value.(*object.Stringo).Value
			if runeLen(c) != 1 {
				return newError("string index assignment value must be 1 character long. got=%d", runeLen(c))
			}
			if idx, ok := index.(*object.Integer); ok {
				switch node.Token.Literal {
				case "=":
					sb := strings.Builder{}
					for i, ch := range s {
						if i == int(idx.Value) {
							sb.WriteString(c)
						} else {
							sb.WriteRune(ch)
						}
					}
					str.Value = sb.String()
				default:
					return newError("unknown assignment operator: STRING INDEX %s", node.Token.Literal)
				}
			} else {
				return newError("cannot index string with %#v", index)
			}
		} else if bs, ok := leftObj.(*object.BlueStruct); ok {
			indexField, ok := ie.Index.(*ast.StringLiteral)
			if !ok {
				return newError("index operator not supported: BLUE_STRUCT.%s", ie.Index.String())
			}
			fieldName := indexField.Value
			operator := node.Token.Literal
			orig, origIndex := bs.Get(fieldName)
			if orig == nil {
				return newError("field name `%s` not found on blue struct: %s", fieldName, bs.Inspect())
			}
			if operator == "=" {
				err := bs.Set(origIndex, value)
				if err != nil {
					return newError("%s", err.Error())
				}
				return NULL
			}
			evaluated := e.evalMultiCharAssignmentInfixExpression(operator, object.BLUE_STRUCT_OBJ, orig, value)
			if isError(evaluated) {
				return evaluated
			}
			err := bs.Set(origIndex, evaluated)
			if err != nil {
				return newError("%s", err.Error())
			}
			return NULL
		} else {
			return newError("object type %T does not support item assignment", leftObj)
		}
	} else {
		return newError("expected identifier or index expression got=%T", node.Left)
	}

	return NULL
}

func (e *Evaluator) evalMultiCharAssignmentInfixExpression(operator string, t object.Type, left, right object.Object) object.Object {
	var evaluated object.Object
	switch operator {
	case "+=":
		evaluated = e.evalInfixExpression("+", left, right)
	case "-=":
		evaluated = e.evalInfixExpression("-", left, right)
	case "*=":
		evaluated = e.evalInfixExpression("*", left, right)
	case "/=":
		evaluated = e.evalInfixExpression("/", left, right)
	case "//=":
		evaluated = e.evalInfixExpression("//", left, right)
	case "**=":
		evaluated = e.evalInfixExpression("**", left, right)
	case "&=":
		evaluated = e.evalInfixExpression("&", left, right)
	case "|=":
		evaluated = e.evalInfixExpression("|", left, right)
	case "~=":
		evaluated = e.evalInfixExpression("~", left, right)
	case "<<=":
		evaluated = e.evalInfixExpression("<<", left, right)
	case ">>=":
		evaluated = e.evalInfixExpression(">>", left, right)
	case "%=":
		evaluated = e.evalInfixExpression("%", left, right)
	case "^=":
		evaluated = e.evalInfixExpression("^", left, right)
	case "&&=":
		evaluated = e.evalInfixExpression("&&", left, right)
	case "||=":
		evaluated = e.evalInfixExpression("||", left, right)
	default:
		return newError("assignment operator unsupported: %s %s", t, operator)
	}
	return evaluated
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
				return newError("failed to unset ENV key '%s'", key)
			}
		} else {
			// set the env var
			v := value.(*object.Stringo).Value
			err := os.Setenv(key, v)
			if err != nil {
				return newError("failed to set ENV key='%s', value='%s'", key, v)
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

	return newError("unhandled builtin obj assignment on '%s'", ident.Value)
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

		ok := object.IsHashable(key)
		if !ok {
			return newError("unusable as a map key: %s", key.Type())
		}
		hk := object.HashObject(key)
		hashed := object.HashKey{Type: key.Type(), Value: hk}

		value := e.Eval(valueNode)
		if isError(value) {
			return value
		}

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
	case left.Type() == object.PROCESS_OBJ && indx.Type() == object.STRING_OBJ:
		return e.evalProcessIndexExpression(left, indx)
	case left.Type() == object.BLUE_STRUCT_OBJ && indx.Type() == object.STRING_OBJ:
		return e.evalBlueStructIndexExpression(left, indx)
	default:
		return newError("index operator not supported: %s.%s", left.Type(), indx.Type())
	}
}

func (e *Evaluator) evalBlueStructIndexExpression(left, indx object.Object) object.Object {
	bs := left.(*object.BlueStruct)
	fieldName := indx.(*object.Stringo).Value
	obj, _ := bs.Get(fieldName)
	if obj == nil {
		return newError("field name `%s` does not exist on blue struct", fieldName)
	}
	return obj
}

func (e *Evaluator) evalProcessIndexExpression(left, indx object.Object) object.Object {
	p := left.(*object.Process)
	s := indx.(*object.Stringo).Value
	switch s {
	case "id":
		return &object.UInteger{Value: p.Id}
	case "name":
		return &object.Stringo{Value: p.NodeName}
	case "send":
		proc := p
		return &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newInvalidArgCountError("send", len(args), 1, "")
				}
				proc.Ch <- args[0]
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`send` will take the given value and send it to the process",
				signature:   "send(pid: PROCESS, val: any) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "send(#{name: '', id: 1}, 'hello') => null",
			}.String(),
		}
	case "recv":
		proc := p
		return &object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 0 {
					return newInvalidArgCountError("recv", len(args), 0, "")
				}
				val := <-proc.Ch
				if val == nil {
					return newError("`recv` error: process channel was closed")
				}
				return val
			},
			HelpStr: helpStrArgs{
				explanation: "`recv` waits for a value on the given process and returns it",
				signature:   "recv(pid: PROCESS) -> any",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "recv(#{name: '', id: 1}) => 'something'",
			}.String(),
		}
	default:
		return newError("%s.%s %q not supported", left.Type(), indx.Type(), s)
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

	ok := object.IsHashable(indx)
	if !ok {
		return newError("unusable as a map key: %s", indx.Type()), false
	}
	hashed := object.HashObject(indx)
	key := object.HashKey{Type: indx.Type(), Value: hashed}

	pair, ok := mapObj.Pairs.Get(key)
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
	// TODO: Return if its immutable here as well
	if val, ok := e.env.Get(node.Value); ok {
		return val
	}

	for _, b := range e.Builtins {
		if builtin, ok := b.Get(node.Value); ok {
			return builtin
		}
	}
	for _, b := range e.BuiltinObjs {
		if builtin, ok := b[node.Value]; ok {
			return builtin.Obj
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
			return e.evalBlockStatement(ie.Consequences[i])
		}
	}
	if ie.Alternative != nil {
		return e.evalBlockStatement(ie.Alternative)
	} else {
		return NULL
	}
}

func (e *Evaluator) evalInfixExpression(operator string, left, right object.Object) object.Object {
	// Special Case for adding to set
	if operator == "+" && (left.Type() == object.SET_OBJ || right.Type() == object.SET_OBJ) {
		return e.evalSetInfixExpression(operator, left, right)
	}
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.BIG_INTEGER_OBJ ||
		left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.INTEGER_OBJ ||
		left.Type() == object.INTEGER_OBJ && right.Type() == object.BIG_INTEGER_OBJ ||
		left.Type() == object.UINTEGER_OBJ && right.Type() == object.BIG_INTEGER_OBJ ||
		left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.UINTEGER_OBJ:
		return e.evalBigIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ ||
		left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ ||
		left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.BIG_FLOAT_OBJ ||
		left.Type() == object.FLOAT_OBJ && right.Type() == object.BIG_INTEGER_OBJ ||
		left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.FLOAT_OBJ ||
		left.Type() == object.FLOAT_OBJ && right.Type() == object.BIG_FLOAT_OBJ ||
		left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.FLOAT_OBJ ||
		left.Type() == object.INTEGER_OBJ && right.Type() == object.BIG_FLOAT_OBJ ||
		left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.INTEGER_OBJ ||
		left.Type() == object.UINTEGER_OBJ && right.Type() == object.BIG_FLOAT_OBJ ||
		left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.UINTEGER_OBJ ||
		left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.BIG_INTEGER_OBJ ||
		left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.BIG_FLOAT_OBJ:
		return e.evalBigFloatInfixExpression(operator, left, right)
	case left.Type() == object.UINTEGER_OBJ && right.Type() == object.UINTEGER_OBJ ||
		left.Type() == object.INTEGER_OBJ && right.Type() == object.UINTEGER_OBJ ||
		left.Type() == object.UINTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalUintegerInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return e.evalStringInfixExpression(operator, left, right)
	case left.Type() == object.LIST_OBJ && right.Type() == object.LIST_OBJ:
		return e.evalDefaultListInfixExpression(operator, left, right)
	case left.Type() == object.SET_OBJ && right.Type() == object.SET_OBJ:
		return e.evalSetInfixExpression(operator, left, right)
	case left.Type() == object.BYTES_OBJ && right.Type() == object.BYTES_OBJ:
		return e.evalBytesInfixExpression(operator, left, right)
	// These are the cases where they differ
	case left.Type() == object.STRING_OBJ && right.Type() == object.INTEGER_OBJ ||
		left.Type() == object.INTEGER_OBJ && right.Type() == object.STRING_OBJ ||
		left.Type() == object.STRING_OBJ && right.Type() == object.UINTEGER_OBJ ||
		left.Type() == object.UINTEGER_OBJ && right.Type() == object.STRING_OBJ:
		return e.evalStringUIntegerInfixExpression(operator, left, right)
	case left.Type() == object.LIST_OBJ && !isBooleanOperator(operator):
		return e.evalListInfixExpression(operator, left, right)
	case right.Type() == object.SET_OBJ && !isBooleanOperator(operator):
		return e.evalRightSideSetInfixExpression(operator, left, right)
	default:
		return e.evalDefaultInfixExpression(operator, left, right)
	}
}

func (e *Evaluator) evalDefaultListInfixExpression(operator string, left, right object.Object) object.Object {
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
	case "==", "!=":
		return e.evalDefaultInfixExpression(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalStringInfixExpression(operator string, left, right object.Object) object.Object {
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
		return nativeToBooleanObject(strings.Contains(rightStr, leftStr))
	case "notin":
		return nativeToBooleanObject(!strings.Contains(rightStr, leftStr))
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
}

func (e *Evaluator) evalDefaultInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case operator == "==":
		return nativeToBooleanObject(object.HashObject(left) == object.HashObject(right))
	case operator == "!=":
		return nativeToBooleanObject(object.HashObject(left) != object.HashObject(right))
	case operator == "and" || operator == "&&":
		leftBool, ok := left.(*object.Boolean)
		if !ok {
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
		rightBool, ok := right.(*object.Boolean)
		if !ok {
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
		return nativeToBooleanObject(leftBool.Value && rightBool.Value)
	case operator == "or" || operator == "||":
		if left == NULL {
			// Null coalescing operator returns right side if left is null
			return right
		}
		leftBool, ok := left.(*object.Boolean)
		if !ok {
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
		rightBool, ok := right.(*object.Boolean)
		if !ok {
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
		return nativeToBooleanObject(leftBool.Value || rightBool.Value)
	case (operator == "in" || operator == "notin") && (right.Type() == object.LIST_OBJ || right.Type() == object.SET_OBJ || right.Type() == object.MAP_OBJ):
		return e.evalInOrNotinInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return e.evalStringInfixExpression(operator, left, right)
	case left.Type() == object.LIST_OBJ && right.Type() == object.LIST_OBJ:
		return e.evalDefaultListInfixExpression(operator, left, right)
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
	if operator == "in" || operator == "notin" {
		hashed := object.HashObject(left)
		_, ok := setElems.Get(hashed)
		if operator == "in" {
			return nativeToBooleanObject(ok)
		} else {
			return nativeToBooleanObject(!ok)
		}
	}
	return e.evalDefaultInfixExpression(operator, left, right)
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
	newSet := &object.Set{Elements: object.NewSetElements()}
	if operator == "+" {
		var s *object.Set
		var key uint64
		var obj object.Object
		if left.Type() == object.SET_OBJ {
			// return set with right obj added
			s = left.(*object.Set)
			key = object.HashObject(right)
			obj = right
		} else {
			// return set with left obj added
			s = right.(*object.Set)
			key = object.HashObject(left)
			obj = left
		}
		for _, k := range s.Elements.Keys {
			v, ok := s.Elements.Get(k)
			if ok {
				newSet.Elements.Set(k, v)
			}
		}
		if _, ok := s.Elements.Get(key); !ok {
			// Key does not exist, add new elem
			newSet.Elements.Set(key, object.SetPair{Value: obj, Present: struct{}{}})
		}
		return newSet
	}
	leftE := left.(*object.Set).Elements
	rightE := right.(*object.Set).Elements
	var leftElems *object.OrderedMap2[uint64, object.SetPair]
	var rightElems *object.OrderedMap2[uint64, object.SetPair]
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
			_, ok = rightElems.Get(k)
			if !ok {
				continue
			}
			newSet.Elements.Set(k, v)
		}
		return newSet
	case "^":
		// symmetric difference
		for _, k := range leftElems.Keys {
			v, ok := leftElems.Get(k)
			if !ok {
				continue
			}
			_, ok = rightElems.Get(k)
			if !ok {
				newSet.Elements.Set(k, v)
			}
		}
		for _, k := range rightElems.Keys {
			v, ok := rightElems.Get(k)
			if !ok {
				continue
			}
			_, ok = leftElems.Get(k)
			if !ok {
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
			_, ok = rightElems.Get(k)
			if !ok {
				newSet.Elements.Set(k, v)
			}
		}
		return newSet
	case "==":
		for _, k := range leftElems.Keys {
			_, ok := rightElems.Get(k)
			if !ok {
				return FALSE
			}
		}
		return TRUE
	case "!=":
		for _, k := range leftElems.Keys {
			_, ok := rightElems.Get(k)
			if !ok {
				return TRUE
			}
		}
		return FALSE
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalBigFloatInfixExpression returns the infix expression object from the left and right BigInteger/UInteger/Integer/Float/BigFloat
// it converts both left and right to their *big.Int values
func (e *Evaluator) evalBigFloatInfixExpression(operator string, left, right object.Object) object.Object {
	var leftVal, rightVal decimal.Decimal
	if lBF, ok := left.(*object.BigFloat); ok {
		leftVal = lBF.Value
	} else if lF, ok := left.(*object.Float); ok {
		leftVal = decimal.NewFromFloat(lF.Value)
	} else if lI, ok := left.(*object.Integer); ok {
		leftVal = decimal.NewFromInt(lI.Value)
	} else if lBI, ok := left.(*object.BigInteger); ok {
		leftVal = decimal.NewFromBigInt(lBI.Value, 0)
	}
	if rBF, ok := right.(*object.BigFloat); ok {
		rightVal = rBF.Value
	} else if rF, ok := right.(*object.Float); ok {
		rightVal = decimal.NewFromFloat(rF.Value)
	} else if rI, ok := right.(*object.Integer); ok {
		rightVal = decimal.NewFromInt(rI.Value)
	} else if rBI, ok := right.(*object.BigInteger); ok {
		rightVal = decimal.NewFromBigInt(rBI.Value, 0)
	}
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
		return nativeToBooleanObject(compared == -1)
	case ">":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared == 1)
	case "<=":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared == -1 || compared == 0)
	case ">=":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared == 1 || compared == 0)
	case "==":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared == 0)
	case "!=":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared != 0)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalBigIntegerInfixExpression returns the infix expression object from the left and right BigInteger/UInteger/Integer
// it converts both left and right to their *big.Int values
func (e *Evaluator) evalBigIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	// Only Integers, Uintegers or BigIntegers should be passed in
	var leftVal, rightVal *big.Int
	if lBI, ok := left.(*object.BigInteger); ok {
		leftVal = lBI.Value
	} else if lI, ok := left.(*object.Integer); ok {
		leftVal = new(big.Int).SetInt64(lI.Value)
	} else {
		leftVal = new(big.Int).SetUint64(left.(*object.UInteger).Value)
	}
	if rBI, ok := right.(*object.BigInteger); ok {
		rightVal = rBI.Value
	} else if rI, ok := right.(*object.Integer); ok {
		rightVal = new(big.Int).SetInt64(rI.Value)
	} else {
		rightVal = new(big.Int).SetUint64(right.(*object.UInteger).Value)
	}
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
		floored, _ := result.DivMod(leftVal, rightVal, maybeWanted)
		// Note: Ignoring the modulus here
		return &object.BigInteger{Value: floored}
	case "%":
		return &object.BigInteger{Value: result.Mod(leftVal, rightVal)}
	case "<":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared == -1)
	case ">":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared == 1)
	case "<=":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared == -1 || compared == 0)
	case ">=":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared == 1 || compared == 0)
	case "==":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared == 0)
	case "!=":
		compared := leftVal.Cmp(rightVal)
		return nativeToBooleanObject(compared != 0)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalStringUIntegerInfixExpression returns the infix expression object from the left or right string and left/right UInteger/Integer
// it converts the integer to a uinteger and a string will be returned
func (e *Evaluator) evalStringUIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	var strToBuild string
	var amount uint64
	if s, ok := left.(*object.Stringo); ok {
		strToBuild = s.Value
	} else if lI, ok := left.(*object.Integer); ok {
		amount = uint64(lI.Value)
	} else {
		amount = left.(*object.UInteger).Value
	}
	if s, ok := right.(*object.Stringo); ok {
		strToBuild = s.Value
	} else if rI, ok := right.(*object.Integer); ok {
		amount = uint64(rI.Value)
	} else {
		amount = right.(*object.UInteger).Value
	}
	switch operator {
	case "*":
		var out bytes.Buffer
		var i uint64
		for i = 0; i < amount; i++ {
			out.WriteString(strToBuild)
		}
		return &object.Stringo{Value: out.String()}
	default:
		return e.evalDefaultInfixExpression(operator, left, right)
	}
}

// evalUintegerInfixExpression returns the infix expression object from the left and right Uinteger/Integer
// it converts both left and right to their uint64 values
func (e *Evaluator) evalUintegerInfixExpression(operator string, left, right object.Object) object.Object {
	// Only UIntegers and Integers should be passed into this
	var leftVal, rightVal uint64
	if lUI, ok := left.(*object.UInteger); ok {
		leftVal = lUI.Value
	} else {
		leftIntVal := left.(*object.Integer).Value
		if leftIntVal < 0 {
			return newError("Left Integer was negative, and is not allowed for Unsigned Integer operations. %s %s %s", left.Inspect(), operator, right.Inspect())
		}
		leftVal = uint64(leftIntVal)
	}
	if rUI, ok := right.(*object.UInteger); ok {
		rightVal = rUI.Value
	} else {
		rightIntVal := right.(*object.Integer).Value
		if rightIntVal < 0 {
			return newError("Right Integer was negative, and is not allowed for Unsigned Integer operations. %s %s %s", left.Inspect(), operator, right.Inspect())
		}
	}
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

// evalFloatInfixExpression returns the infix expression object from the left and right Float/Integer
// it converts both left and right to their float64 values
func (e *Evaluator) evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	// Only Integers and Floats should be passed into this
	var leftVal, rightVal float64
	if lF, ok := left.(*object.Float); ok {
		leftVal = lF.Value
	} else {
		leftVal = float64(left.(*object.Integer).Value)
	}
	if rF, ok := right.(*object.Float); ok {
		rightVal = rF.Value
	} else {
		rightVal = float64(right.(*object.Integer).Value)
	}
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

func (e *Evaluator) evalPrefixExpression(operator string, right object.Object, rightNode ast.Expression) object.Object {
	switch operator {
	case "not", "!":
		return e.evalNotOperatorExpression(right)
	case "-":
		return e.evalMinusPrefixOperatorExpression(right)
	case "~":
		return e.evalBitwiseNotOperatorExpression(right)
	case "<<":
		// Because this mutates the list we will check here for immutability
		if ident, ok := rightNode.(*ast.Identifier); ok {
			if e.env.IsImmutable(ident.Value) {
				return newError("'%s' is immutable", ident.Value)
			}
		}
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

	defer func() {
		if funAndArgs, ok := e.deferFuns[e.scopeNestLevel]; ok {
			for funAndArgs.Len() > 0 {
				funAndArg := funAndArgs.Pop()
				// Note: Return values are ignored for defer functions
				e.applyFunctionFast(funAndArg.Fun, funAndArg.Args, make(map[string]object.Object), []bool{})
			}
			delete(e.deferFuns, e.scopeNestLevel)
		}
		for i := 0; i < e.maybeNullMapFnCall.Len(); i++ {
			e.maybeNullMapFnCall.Pop()
		}
	}()

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
	var result object.Object = NULL

	e.scopeNestLevel++
	e.isInScopeBlock[e.scopeNestLevel] = struct{}{}

	defer func() {
		delete(e.isInScopeBlock, e.scopeNestLevel)
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
		if funAndArgs, ok := e.deferFuns[e.scopeNestLevel]; ok {
			for funAndArgs.Len() > 0 {
				funAndArg := funAndArgs.Pop()
				// Note: Return values are ignored for defer functions
				e.applyFunctionFast(funAndArg.Fun, funAndArg.Args, make(map[string]object.Object), []bool{})
			}
			delete(e.deferFuns, e.scopeNestLevel)
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
