package evaluator

import (
	"blue/ast"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

var (
	// TRUE is the true object which should be the same everywhere
	TRUE = &object.Boolean{Value: true}
	// FALSE is the false object which should be the same everywhere
	FALSE = &object.Boolean{Value: false}
	// NULL is the null object which should be the same everywhere
	NULL = &object.Null{}
	// IGNORE is the object which is used to ignore variables when necessary
	IGNORE = &object.Null{}
)

type Evaluator struct {
	env *object.Environment

	// EvalBasePath is the base directory from which the current file is being run
	EvalBasePath string

	// CurrentFile is the file being executed (or <stdin> if run from the REPL)
	CurrentFile string

	// ArgToPassToBuiltin is the argument to be given to the builtin function
	ArgToPassToBuiltin object.Object

	// Used for: indx, elem in for expression
	nestLevel     int
	iterCount     []int
	cleanupTmpVar map[string]bool
}

func New() *Evaluator {
	return &Evaluator{
		env: object.NewEnvironment(),

		EvalBasePath: ".",
		CurrentFile:  "<stdin>",

		ArgToPassToBuiltin: nil,

		nestLevel:     -1,
		iterCount:     []int{},
		cleanupTmpVar: make(map[string]bool),
	}
}

// Eval takes an ast node and returns an object
func (e *Evaluator) Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return e.evalProgram(node)
	case *ast.ExpressionStatement:
		return e.Eval(node.Expression)
	case *ast.Identifier:
		return e.evalIdentifier(node)
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
			return right
		}
		return e.evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		// If were in an `in` expression, it needs to be evaluated differently (for `for`)
		if node.Operator == "in" {
			return e.evalInExpression(node)
		}
		left := e.Eval(node.Left)
		if isError(left) {
			return left
		}
		right := e.Eval(node.Right)
		if isError(right) {
			return right
		}
		return e.evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return e.evalBlockStatement(node)
	case *ast.IfExpression:
		return e.evalIfExpression(node)
	case *ast.ReturnStatement:
		val := e.Eval(node.ReturnValue)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.ValStatement:
		val := e.Eval(node.Value)
		if isError(val) {
			return val
		}
		e.env.ImmutableSet(node.Name.Value)
		e.env.Set(node.Name.Value, val)
	case *ast.VarStatement:
		val := e.Eval(node.Value)
		if isError(val) {
			return val
		}
		e.env.Set(node.Name.Value, val)
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
				return obj
			}
			defaultParams = append(defaultParams, obj)
		}
		funObj := &object.Function{Parameters: params, Body: body, DefaultParameters: defaultParams, Env: e.env}
		e.env.Set(node.Name.Value, funObj)
	case *ast.CallExpression:
		function := e.Eval(node.Function)
		if isError(function) {
			return function
		}
		args := e.evalExpressions(node.Arguments)
		defaultArgs := make(map[string]object.Object)
		for k, v := range node.DefaultArguments {
			val := e.Eval(v)
			if isError(val) {
				return val
			}
			defaultArgs[k] = val
		}
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return e.applyFunction(function, args, defaultArgs)
	case *ast.StringLiteral:
		if len(node.InterpolationValues) == 0 {
			return &object.Stringo{Value: node.Value}
		}
		return e.evalStringWithInterpolation(node)
	case *ast.ExecStringLiteral:
		return e.evalExecStringLiteral(node)
	case *ast.ListLiteral:
		elements := e.evalExpressions(node.Elements)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.List{Elements: elements}
	case *ast.IndexExpression:
		left := e.Eval(node.Left)
		if isError(left) {
			return left
		}
		indx := e.Eval(node.Index)
		if isError(indx) {
			return indx
		}
		val := e.tryCreateValidBuiltinForDotCall(left, indx, node.Left)
		if val != nil {
			return val
		}
		return e.evalIndexExpression(left, indx)
	case *ast.MapLiteral:
		return e.evalMapLiteral(node)
	case *ast.AssignmentExpression:
		return e.evalAssignmentExpression(node)
	case *ast.ForExpression:
		return e.evalForExpression(node)
	case *ast.ListCompLiteral:
		return e.evalListCompLiteral(node)
	case *ast.MapCompLiteral:
		return e.evalMapCompLiteral(node)
	case *ast.SetCompLiteral:
		return e.evalSetCompLiteral(node)
	case *ast.MatchExpression:
		return e.evalMatchExpression(node)
	case *ast.Null:
		return NULL
	case *ast.SetLiteral:
		return e.evalSetLiteral(node)
	case *ast.ImportStatement:
		return e.evalImportStatement(node)
	case *ast.TryCatchStatement:
		return e.evalTryCatchStatement(node)
	default:
		if node == nil {
			// Just want to get rid of this in my output
			return nil
		}
		fmt.Printf("Handle this type: %T\n", node)
	}

	// In the event that there are only statements, I think this is where we end up
	// so we return NULL because there is nothing to return otherwise
	return NULL
}

func (e *Evaluator) evalImportStatement(node *ast.ImportStatement) object.Object {
	name := node.Path.Value
	fpath := e.createFilePathFromImportPath(name)
	modName := strings.ReplaceAll(filepath.Base(fpath), ".b", "")
	file, err := filepath.Abs(fpath)
	if err != nil {
		return newError("Failed to import '%s'. Could not get absolute filepath.", name)
	}
	ofile, err := os.Open(file)
	if err != nil {
		return newError("Failed to import '%s'. Could not open file '%s' for reading.", name, file)
	}
	defer ofile.Close()
	fileData, err := ioutil.ReadAll(ofile)
	if err != nil {
		return newError("Failed to import '%s'. Could not read the file.", name)
	}
	inputStr := string(fileData)
	l := lexer.New(inputStr)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			io.WriteString(os.Stdout, "ParserError: "+msg+"\n")
		}
		return newError("ParserError: File '%s' contains Parser Errors.", name)
	}
	newE := New()
	val := newE.Eval(program)
	if isError(val) {
		return val
	}
	mod := &object.Module{Name: modName, Env: newE.env}
	e.env.Set(modName, mod)
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
			// TODO: Need to figure out the order of returns in case of errors in catch or finally block
			e.Eval(node.FinallyBlock)
		}
		return evaldCatch
	}
	if node.FinallyBlock != nil {
		// TODO: Need to figure out the order of returns in case of errors in catch or finally block
		e.Eval(node.FinallyBlock)
	}
	return NULL
}

func (e *Evaluator) evalSetLiteral(node *ast.SetLiteral) object.Object {
	elements := e.evalExpressions(node.Elements)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}

	setMap := make(map[uint64]object.SetPair, len(elements))
	for _, e := range elements {
		hashKey := object.HashObject(e)
		setMap[hashKey] = object.SetPair{Value: e, Present: true}
	}
	return &object.Set{Elements: setMap}
}

func (e *Evaluator) evalMatchExpression(node *ast.MatchExpression) object.Object {
	conditionLen := len(node.Condition)
	consequenceLen := len(node.Consequence)
	if conditionLen != consequenceLen {
		return newError("conditons length is not equal to consequences length in match expression")
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
		// Run through each condtion and if it evaluates to "true" then return the evaluated consequence
		condVal := e.Eval(node.Condition[i])
		// This is our very basic form of pattern matching
		if condVal.Type() == object.MAP_OBJ && optVal != nil && optVal.Type() == object.MAP_OBJ {
			// Do our shape matching on it
			if doCondAndMatchExpEqual(condVal, optVal) {
				return e.Eval(node.Consequence[i])
			}
		}
		if optVal == nil {
			evald := e.Eval(node.Condition[i])
			if isError(evald) {
				return evald
			}
			if evald == TRUE {
				return e.Eval(node.Consequence[i])
			}
		}
		if object.HashObject(condVal) == object.HashObject(optVal) {
			return e.Eval(node.Consequence[i])
		}
		if condVal == IGNORE {
			return e.Eval(node.Consequence[i])
		}
	}
	// Shouldnt reach here ideally
	return nil
}

func (e *Evaluator) evalListCompLiteral(node *ast.ListCompLiteral) object.Object {
	l := lexer.New(node.NonEvaluatedProgram)
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return nil
	}
	_ = e.Eval(rootNode)
	someVal, ok := e.env.Get("__internal__")
	if !ok {
		return nil
	}
	return &object.ListCompLiteral{Elements: someVal.(*object.List).Elements}
}

func (e *Evaluator) evalMapCompLiteral(node *ast.MapCompLiteral) object.Object {
	l := lexer.New(node.NonEvaluatedProgram)
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return nil
	}
	_ = e.Eval(rootNode)
	someVal, ok := e.env.Get("__internal__")
	if !ok {
		return nil
	}
	return &object.MapCompLiteral{Pairs: someVal.(*object.Map).Pairs}
}

func (e *Evaluator) evalSetCompLiteral(node *ast.SetCompLiteral) object.Object {
	l := lexer.New(node.NonEvaluatedProgram)
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return nil
	}
	_ = e.Eval(rootNode)
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
	if evaluatedRight.Type() == object.LIST_OBJ {
		// This is where we handle if its a list
		list := evaluatedRight.(*object.List).Elements
		_, ok := e.env.Get(ident.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(ident.Value, list[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(list) == e.iterCount[e.nestLevel] {
				// Reset iteration for other items
				e.iterCount[e.nestLevel] = 0
				e.nestLevel--
				e.cleanupTmpVar[ident.Value] = true
				return FALSE
			}
			return TRUE
		}
		e.env.Set(ident.Value, list[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if len(list) == e.iterCount[e.nestLevel] {
			// Reset iteration for other items
			e.iterCount[e.nestLevel] = 0
			e.nestLevel--
			e.cleanupTmpVar[ident.Value] = true
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.MAP_OBJ {
		// This is where we handle if its a Map
		// TODO: We need to get the key as a string/number/boolean instead of hashkey, maybe their could be some lookup method

		// Right now we are actually using a list of the pair when the left side is an ident
		// but we can probably allow the user to use a list, that destructures to 2 idents
		// or we will need a new ast expression for multiple ident assignments
		// TODO: This is where we can modify if we want to only use keys
		mapPairs := evaluatedRight.(*object.Map).Pairs
		pairObjs := make([]*object.List, 0, len(mapPairs))
		keys := []object.HashKey{}
		for k := range mapPairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(p, q int) bool {
			return keys[p].Value < keys[q].Value
		})
		for i := 0; i < len(mapPairs); i++ {
			listObj := []object.Object{mapPairs[keys[i]].Key, mapPairs[keys[i]].Value}
			pairObjs = append(pairObjs, &object.List{Elements: listObj})
		}
		_, ok := e.env.Get(ident.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(ident.Value, pairObjs[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(pairObjs) == e.iterCount[e.nestLevel] {
				// Reset iteration for other items
				e.iterCount[e.nestLevel] = 0
				e.nestLevel--
				e.cleanupTmpVar[ident.Value] = true
				return FALSE
			}
			return TRUE
		}
		e.env.Set(ident.Value, pairObjs[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if len(mapPairs) == e.iterCount[e.nestLevel] {
			// Reset iteration for other items
			e.iterCount[e.nestLevel] = 0
			e.nestLevel--
			e.cleanupTmpVar[ident.Value] = true
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.STRING_OBJ {
		// This is where we handle if its a string
		strVal := evaluatedRight.(*object.Stringo).Value
		chars := []byte(strVal)
		stringObjs := make([]*object.Stringo, 0, len(chars))
		for _, ch := range chars {
			stringObjs = append(stringObjs, &object.Stringo{Value: string(ch)})
		}
		_, ok := e.env.Get(ident.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(ident.Value, stringObjs[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(chars) == e.iterCount[e.nestLevel] {
				// Reset iteration for other items
				e.iterCount[e.nestLevel] = 0
				e.nestLevel--
				e.cleanupTmpVar[ident.Value] = true
				return FALSE
			}
			return TRUE
		}
		e.env.Set(ident.Value, stringObjs[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if len(stringObjs) == e.iterCount[e.nestLevel] {
			// Reset iteration for other items
			e.iterCount[e.nestLevel] = 0
			e.nestLevel--
			e.cleanupTmpVar[ident.Value] = true
			return FALSE
		}
		return TRUE
	}
	return newError("Expected List, Map, or String on right hand side. got %T", evaluatedRight.Type())
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
		_, ok := e.env.Get(identRight.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
			e.env.Set(identRight.Value, list[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(list) == e.iterCount[e.nestLevel] {
				// Reset iteration for other items
				e.iterCount[e.nestLevel] = 0
				e.nestLevel--
				e.cleanupTmpVar[identLeft.Value] = true
				e.cleanupTmpVar[identRight.Value] = true
				return FALSE
			}
			return TRUE
		}
		e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
		e.env.Set(identRight.Value, list[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if len(list) == e.iterCount[e.nestLevel] {
			// Reset iteration for other items
			e.iterCount[e.nestLevel] = 0
			e.nestLevel--
			e.cleanupTmpVar[identLeft.Value] = true
			e.cleanupTmpVar[identRight.Value] = true
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.MAP_OBJ {
		mapPairs := evaluatedRight.(*object.Map).Pairs
		pairObjs := make([]*object.List, 0, len(mapPairs))
		keys := []object.HashKey{}
		for k := range mapPairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(p, q int) bool {
			return keys[p].Value < keys[q].Value
		})
		for i := 0; i < len(mapPairs); i++ {
			listObj := []object.Object{mapPairs[keys[i]].Key, mapPairs[keys[i]].Value}
			pairObjs = append(pairObjs, &object.List{Elements: listObj})
		}
		_, ok := e.env.Get(identRight.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(identLeft.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[0])
			e.env.Set(identRight.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[1])
			e.iterCount[e.nestLevel]++
			if len(pairObjs) == e.iterCount[e.nestLevel] {
				// Reset iteration for other items
				e.iterCount[e.nestLevel] = 0
				e.nestLevel--
				e.cleanupTmpVar[identLeft.Value] = true
				e.cleanupTmpVar[identRight.Value] = true
				return FALSE
			}
			return TRUE
		}
		e.env.Set(identLeft.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[0])
		e.env.Set(identRight.Value, pairObjs[e.iterCount[e.nestLevel]].Elements[1])
		e.iterCount[e.nestLevel]++
		if len(mapPairs) == e.iterCount[e.nestLevel] {
			// Reset iteration for other items
			e.iterCount[e.nestLevel] = 0
			e.nestLevel--
			e.cleanupTmpVar[identLeft.Value] = true
			e.cleanupTmpVar[identRight.Value] = true
			return FALSE
		}
		return TRUE
	} else if evaluatedRight.Type() == object.STRING_OBJ {
		// This is where we handle if its a string
		strVal := evaluatedRight.(*object.Stringo).Value
		chars := []byte(strVal)
		stringObjs := make([]*object.Stringo, 0, len(chars))
		for _, ch := range chars {
			stringObjs = append(stringObjs, &object.Stringo{Value: string(ch)})
		}
		_, ok := e.env.Get(identRight.Value)
		if !ok {
			e.iterCount = append(e.iterCount, 0)
			e.nestLevel++
			e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
			e.env.Set(identRight.Value, stringObjs[e.iterCount[e.nestLevel]])
			e.iterCount[e.nestLevel]++
			if len(chars) == e.iterCount[e.nestLevel] {
				// Reset iteration for other items
				e.iterCount[e.nestLevel] = 0
				e.nestLevel--
				e.cleanupTmpVar[identLeft.Value] = true
				e.cleanupTmpVar[identRight.Value] = true
				return FALSE
			}
			return TRUE
		}
		e.env.Set(identLeft.Value, &object.Integer{Value: int64(e.iterCount[e.nestLevel])})
		e.env.Set(identRight.Value, stringObjs[e.iterCount[e.nestLevel]])
		e.iterCount[e.nestLevel]++
		if len(stringObjs) == e.iterCount[e.nestLevel] {
			// Reset iteration for other items
			e.iterCount[e.nestLevel] = 0
			e.nestLevel--
			e.cleanupTmpVar[identRight.Value] = true
			e.cleanupTmpVar[identRight.Value] = true
			return FALSE
		}
		return TRUE
	}
	return newError("Expected List, Map, or String on right hand side. got %T", evaluatedRight.Type())
}

func (e *Evaluator) evalForExpression(node *ast.ForExpression) object.Object {
	var evalBlock object.Object
	for {
		evalCond := e.Eval(node.Condition)
		if isError(evalCond) {
			return evalCond
		}
		ok := evalCond.(*object.Boolean).Value
		evalBlock = e.Eval(node.Consequence)
		if isError(evalBlock) {
			return evalBlock
		}
		// Cleanup any temporary for variables
		tmpMapCopy := e.cleanupTmpVar
		for k, v := range tmpMapCopy {
			if v {
				e.env.RemoveIdentifier(k)
				delete(e.cleanupTmpVar, k)
			}
		}
		// Still evaluate on the last run then break if its false
		if !ok {
			break
		}
	}
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
		rootObjIdent := strings.Split(removeLeftParens, "[")[0]
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
		// TODO: Handle all assignment tokens that apply
		obj := e.Eval(ie.Left)
		if isError(obj) {
			return obj
		}

		if list, ok := obj.(*object.List); ok {
			index := e.Eval(ie.Index)
			if isError(index) {
				return index
			}

			if idx, ok := index.(*object.Integer); ok {
				list.Elements[idx.Value] = value
			} else {
				return newError("cannot index list with %#v", index)
			}
		} else if mapObj, ok := obj.(*object.Map); ok {
			key := e.Eval(ie.Index)
			if isError(key) {
				return key
			}

			if hashKey, ok := key.(object.Hashable); ok {
				hashed := hashKey.HashKey()
				mapObj.Pairs[hashed] = object.MapPair{Key: key, Value: value}
			} else {
				return newError("cannot index map with %T", key)
			}
		} else {
			return newError("object type %T does not support item assignment", obj)
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
	pairs := make(map[object.HashKey]object.MapPair)

	for keyNode, valueNode := range node.Pairs {
		ident, _ := keyNode.(*ast.Identifier)
		key := e.Eval(keyNode)
		if isError(key) && ident != nil {
			key = &object.Stringo{Value: ident.String()}
		} else if isError(key) {
			return key
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
		pairs[hashed] = object.MapPair{Key: key, Value: value}
	}

	return &object.Map{Pairs: pairs}
}

func (e *Evaluator) evalIndexExpression(left, indx object.Object) object.Object {
	switch {
	case left.Type() == object.LIST_OBJ:
		return e.evalListIndexExpression(left, indx)
	case left.Type() == object.MAP_OBJ:
		return e.evalMapIndexExpression(left, indx)
	case left.Type() == object.MODULE_OBJ:
		return e.evalModuleIndexExpression(left, indx)
	case left.Type() == object.STRING_OBJ:
		return e.evalStringIndexExpression(left, indx)
	default:
		// TODO: Support all other index expressions, such as member lookup and hash literals, sets, etc.
		return newError("index operator not supported: %s", left.Type())
	}
}

func (e *Evaluator) evalModuleIndexExpression(module, indx object.Object) object.Object {
	mod := module.(*object.Module)
	name := indx.(*object.Stringo).Value
	val, ok := mod.Env.Get(name)
	if !ok {
		return newError("failed to find '%s' in imported file '%s'", name, mod.Name)
	}
	return val
}

func (e *Evaluator) evalMapIndexExpression(mapObject, indx object.Object) object.Object {
	mapObj := mapObject.(*object.Map)

	key, ok := indx.(object.Hashable)
	if !ok {
		return newError("unusable as map key: %s", indx.Type())
	}

	pair, ok := mapObj.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
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
		// TODO: possibly support -1 to get last element and negative numbers to go in reverse lookup of the list
		// This would make the code below a bit more complex but still fairly easy to implement
		return NULL
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
	default:
		return NULL
	}
	max := int64(runeLen(strObj.Value) - 1)
	if idx < 0 || idx > max {
		// TODO: possibly support -1 to get last element and negative numbers to go in reverse lookup of the list
		// This would make the code below a bit more complex but still fairly easy to implement
		return NULL
	}
	return &object.Stringo{Value: string([]rune(strObj.Value)[idx])}
}

func (e *Evaluator) evalExecStringLiteral(execStringNode *ast.ExecStringLiteral) object.Object {
	str := execStringNode.Value

	splitStr := strings.Split(str, " ")
	if len(splitStr) == 0 {
		return newError("unable to exec the string `%s`", str)
	}
	if len(splitStr) == 1 {
		output, err := execCommand(splitStr[0]).Output()
		if err != nil {
			return newError("unable to exec the string `%s`. Error: %s", str, err)
		}
		return &object.Stringo{Value: string(output[:])}
	}
	cleanedStrings := make([]string, 0)
	for _, v := range splitStr {
		if v != "" {
			cleanedStrings = append(cleanedStrings, v)
			continue
		}
	}
	first := cleanedStrings[0]
	rest := cleanedStrings[1:]

	output, err := execCommand(first, rest...).CombinedOutput()
	if err != nil {
		return newError("unable to exec the string `%s`. Error: %s", str, err)
	}
	if len(output) == 0 {
		return newError("got 0 bytes from exec string output of `%s`.", str)
	}
	return &object.Stringo{Value: string(output[:])}
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
		}
		newString = strings.Replace(newString, stringNode.OriginalInterpolationString[i], "", 1)
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

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	if builtin, ok := stringbuiltins[node.Value]; ok {
		return builtin
	}

	if builtin, ok := builtinobjs[node.Value]; ok {
		return builtin.Obj
	}

	if node.Value == "FILE" {
		return &object.Stringo{Value: e.CurrentFile}
	}

	return newError("identifier not found: " + node.Value)
}

func (e *Evaluator) evalIfExpression(ie *ast.IfExpression) object.Object {
	condition := e.Eval(ie.Condition)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return e.Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return e.Eval(ie.Alternative)
	} else {
		return NULL
	}
}

func (e *Evaluator) evalInfixExpression(operator string, left, right object.Object) object.Object {
	// TODO: implement these similar to how list is set up with one type checked on one side
	// and then check all the other sides in the next eval function
	switch {
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
			// TODO: Consider using copy in cases like this, if its more efficient, need to measure
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
	case right.Type() == object.INTEGER_OBJ && left.Type() == object.FLOAT_OBJ:
		return e.evalIntegerFloatInfixExpression(operator, right, left)
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
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.LIST_OBJ:
		leftVal := left.(*object.Integer)
		righListElems := right.(*object.List).Elements
		if operator == "in" {
			for _, e := range righListElems {
				if object.HashObject(leftVal) == object.HashObject(e) {
					return TRUE
				}
			}
			return FALSE
		} else if operator == "notin" {
			for _, e := range righListElems {
				if object.HashObject(leftVal) == object.HashObject(e) {
					return FALSE
				}
			}
			return TRUE
		}
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.LIST_OBJ:
		leftVal := left.(*object.Float)
		righListElems := right.(*object.List).Elements
		if operator == "in" {
			for _, e := range righListElems {
				if object.HashObject(leftVal) == object.HashObject(e) {
					return TRUE
				}
			}
			return FALSE
		} else if operator == "notin" {
			for _, e := range righListElems {
				if object.HashObject(leftVal) == object.HashObject(e) {
					return FALSE
				}
			}
			return TRUE
		}
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.UINTEGER_OBJ && right.Type() == object.LIST_OBJ:
		leftVal := left.(*object.UInteger)
		righListElems := right.(*object.List).Elements
		if operator == "in" {
			for _, e := range righListElems {
				if object.HashObject(leftVal) == object.HashObject(e) {
					return TRUE
				}
			}
			return FALSE
		} else if operator == "notin" {
			for _, e := range righListElems {
				if object.HashObject(leftVal) == object.HashObject(e) {
					return FALSE
				}
			}
			return TRUE
		}
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.MAP_OBJ && right.Type() == object.LIST_OBJ:
		leftVal := left.(*object.Map)
		righListElems := right.(*object.List).Elements
		if operator == "in" {
			for _, e := range righListElems {
				if object.HashObject(leftVal) == object.HashObject(e) {
					return TRUE
				}
			}
			return FALSE
		} else if operator == "notin" {
			for _, e := range righListElems {
				if object.HashObject(leftVal) == object.HashObject(e) {
					return FALSE
				}
			}
			return TRUE
		}
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.STRING_OBJ && right.Type() == object.MAP_OBJ:
		leftStr := left.(*object.Stringo)
		rightMap := right.(*object.Map).Pairs
		if operator == "in" {
			for k := range rightMap {
				if k.Value == leftStr.HashKey().Value {
					return TRUE
				}
			}
			return FALSE
		} else if operator == "notin" {
			for k := range rightMap {
				if k.Value == leftStr.HashKey().Value {
					return FALSE
				}
			}
			return TRUE
		}
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.LIST_OBJ:
		return e.evalListInfixExpression(operator, left, right)
	case right.Type() == object.SET_OBJ:
		return e.evalRightSideSetInfixExpression(operator, left, right)
	// NOTE: THESE OPERATORS MUST STAY BELOW THE TYPE CHECKING OTHERWISE IT COULD BREAK THINGS!!
	case operator == "==":
		return nativeToBooleanObject(left == right)
	case operator == "!=":
		return nativeToBooleanObject(left != right)
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
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalRightSideSetInfixExpression(operator string, left, right object.Object) object.Object {
	setElems := right.(*object.Set).Elements
	switch left.Type() {
	case object.INTEGER_OBJ:
		intVal := object.HashObject(left.(*object.Integer))
		switch operator {
		case "in":
			if _, ok := setElems[intVal]; ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems[intVal]; ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case object.UINTEGER_OBJ:
		uintVal := left.(*object.UInteger).HashKey().Value
		switch operator {
		case "in":
			if _, ok := setElems[uintVal]; ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems[uintVal]; ok {
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
			if _, ok := setElems[funHash]; ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems[funHash]; ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case object.MAP_OBJ:
		mapHash := object.HashObject(left.(*object.Map))
		switch operator {
		case "in":
			if _, ok := setElems[mapHash]; ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems[mapHash]; ok {
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
			if _, ok := setElems[boolHash]; ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems[boolHash]; ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case object.STRING_OBJ:
		strHash := left.(*object.Stringo).HashKey().Value
		switch operator {
		case "in":
			if _, ok := setElems[strHash]; ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems[strHash]; ok {
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
			if _, ok := setElems[nullHash]; ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems[nullHash]; ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	case object.LIST_OBJ:
		listHash := object.HashObject(left.(*object.List))
		switch operator {
		case "in":
			if _, ok := setElems[listHash]; ok {
				return TRUE
			}
			return FALSE
		case "notin":
			if _, ok := setElems[listHash]; ok {
				return FALSE
			}
			return TRUE
		default:
			return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
		}
	// TODO: Do we want to support, BigFloat, BigInt, in sets
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

//TODO: Handle `in` and `notin` for set operations
func (e *Evaluator) evalSetInfixExpression(operator string, left, right object.Object) object.Object {
	leftE := left.(*object.Set).Elements
	rightE := right.(*object.Set).Elements
	newSet := &object.Set{Elements: make(map[uint64]object.SetPair)}
	var leftElems, rightElems map[uint64]object.SetPair
	if len(leftE) >= len(rightE) {
		leftElems = leftE
		rightElems = rightE
	} else {
		leftElems = rightE
		rightElems = leftE
	}
	switch operator {
	case "|":
		// union
		for k, v := range leftElems {
			newSet.Elements[k] = v
		}
		for k, v := range rightElems {
			newSet.Elements[k] = v
		}
		return newSet
	case "&":
		// intersect
		for k, v := range leftElems {
			if rightElems[k].Present {
				newSet.Elements[k] = v
			}
		}
		return newSet
	case "^":
		// symmetric difference
		for k, v := range leftElems {
			if !rightElems[k].Present {
				newSet.Elements[k] = v
			}
		}
		for k, v := range rightElems {
			if !leftElems[k].Present {
				newSet.Elements[k] = v
			}
		}
		return newSet
	case ">=":
		// left is superset of right
		for k := range rightE {
			if _, ok := leftE[k]; !ok {
				return FALSE
			}
		}
		return TRUE
	case "<=":
		// right is a superset of left
		for k := range leftE {
			if _, ok := rightE[k]; !ok {
				return FALSE
			}
		}
		return TRUE
	case "-":
		// difference
		for k, v := range leftElems {
			if !rightElems[k].Present {
				newSet.Elements[k] = v
			}
		}
		return newSet
	case "==":
		for k := range leftElems {
			if !rightElems[k].Present {
				return FALSE
			}
		}
		return TRUE
	case "!=":
		for k := range leftElems {
			if !rightElems[k].Present {
				return TRUE
			}
		}
		return FALSE
	// TODO: Should the set support `in` and `notin`
	// case "in":
	// case "notin":
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
	leftVal := right.(*object.BigInteger).Value
	rightVal := left.(*object.Float).Value
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
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
	listElems := make([]object.Object, 0, 1)
	listElems = append(listElems, &object.Integer{Value: leftVal})
	return &object.List{Elements: listElems}
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
	return newError("Can not use non inclusive range when both values equal")
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
	return NULL
}

func (e *Evaluator) evalBitwiseNotOperatorExpression(right object.Object) object.Object {
	value := right.(*object.UInteger).Value
	return &object.UInteger{Value: 0xFFFFFFFFFFFFFFFF ^ value}
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

	for _, stmt := range block.Statements {
		result = e.Eval(stmt)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}
