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
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

// EvalBasePath is the base directory from which the current file is being run
var EvalBasePath = "."

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

// Eval takes an ast node and returns an object
func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.Identifier:
		return evalIdentifier(node, env)
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
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		// If were in an `in` expression, it needs to be evaluated differently (for `for`)
		if node.Operator == "in" {
			return evalInExpression(node, env)
		}
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.ValStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.ImmutableSet(node.Name.Value)
		env.Set(node.Name.Value, val)
	case *ast.VarStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		defaultParams := []object.Object{}
		for _, val := range node.ParameterExpressions {
			if val == nil {
				defaultParams = append(defaultParams, nil)
				continue
			}
			obj := Eval(val, env)
			if isError(obj) {
				return obj
			}
			defaultParams = append(defaultParams, obj)
		}
		return &object.Function{Parameters: params, Body: body, DefaultParameters: defaultParams, Env: env}
	case *ast.FunctionStatement:
		params := node.Parameters
		body := node.Body
		defaultParams := []object.Object{}
		for _, val := range node.ParameterExpressions {
			if val == nil {
				defaultParams = append(defaultParams, nil)
				continue
			}
			obj := Eval(val, env)
			if isError(obj) {
				return obj
			}
			defaultParams = append(defaultParams, obj)
		}
		funObj := &object.Function{Parameters: params, Body: body, DefaultParameters: defaultParams, Env: env}
		env.Set(node.Name.Value, funObj)
	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		defaultArgs := make(map[string]object.Object)
		for k, v := range node.DefaultArguments {
			val := Eval(v, env)
			if isError(val) {
				return val
			}
			defaultArgs[k] = val
		}
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args, defaultArgs)
	case *ast.StringLiteral:
		if len(node.InterpolationValues) == 0 {
			return &object.Stringo{Value: node.Value}
		}
		return evalStringWithInterpolation(node, env)
	case *ast.ExecStringLiteral:
		return evalExecStringLiteral(node, env)
	case *ast.ListLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.List{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		indx := Eval(node.Index, env)
		if isError(indx) {
			return indx
		}
		val := tryCreateValidBuiltinForDotCall(left, indx, node.Left)
		if val != nil {
			return val
		}
		return evalIndexExpression(left, indx, env)
	case *ast.MapLiteral:
		return evalMapLiteral(node, env)
	case *ast.AssignmentExpression:
		return evalAssignmentExpression(node, env)
	case *ast.ForExpression:
		return evalForExpression(node, env)
	case *ast.ListCompLiteral:
		return evalListCompLiteral(node, env)
	case *ast.MapCompLiteral:
		return evalMapCompLiteral(node, env)
	case *ast.SetCompLiteral:
		return evalSetCompLiteral(node, env)
	case *ast.MatchExpression:
		return evalMatchExpression(node, env)
	case *ast.Null:
		return NULL
	case *ast.SetLiteral:
		return evalSetLiteral(node, env)
	case *ast.ImportStatement:
		return evalImportStatement(node, env)
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

func createFilePathFromImportPath(importPath string) string {
	var fpath bytes.Buffer
	if EvalBasePath != "." {
		fpath.WriteString(EvalBasePath)
		fpath.WriteString(string(os.PathSeparator))
	}
	importPath = strings.ReplaceAll(importPath, ".", string(os.PathSeparator))
	fpath.WriteString(importPath)
	fpath.WriteString(".b")
	return fpath.String()
}

func evalImportStatement(node *ast.ImportStatement, env *object.Environment) object.Object {
	name := node.Path.Value
	fpath := createFilePathFromImportPath(name)
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
	newEnv := object.NewEnvironment()
	val := Eval(program, newEnv)
	if isError(val) {
		return val
	}
	mod := &object.Module{Name: modName, Env: newEnv}
	env.Set(modName, mod)
	return NULL
}

func evalSetLiteral(node *ast.SetLiteral, env *object.Environment) object.Object {
	elements := evalExpressions(node.Elements, env)
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

func evalMatchExpression(node *ast.MatchExpression, env *object.Environment) object.Object {
	conditionLen := len(node.Condition)
	consequenceLen := len(node.Consequence)
	if conditionLen != consequenceLen {
		return newError("conditons length is not equal to consequences length in match expression")
	}
	var optVal object.Object
	if node.OptionalValue != nil {
		optVal = Eval(node.OptionalValue, env)
		if isError(optVal) {
			return optVal
		}
	}
	env.Set("_", NULL)
	for i := 0; i < conditionLen; i++ {
		// Run through each condtion and if it evaluates to "true" then return the evaluated consequence
		condVal := Eval(node.Condition[i], env)
		// This is our very basic form of pattern matching
		if condVal.Type() == object.MAP_OBJ && optVal != nil && optVal.Type() == object.MAP_OBJ {
			// Do our shape matching on it
			if doCondAndMatchExpEqual(condVal, optVal) {
				return Eval(node.Consequence[i], env)
			}
		}
		if optVal == nil {
			evald := Eval(node.Condition[i], env)
			if isError(evald) {
				return evald
			}
			if evald == TRUE {
				return Eval(node.Consequence[i], env)
			}
		}
		if object.HashObject(condVal) == object.HashObject(optVal) {
			return Eval(node.Consequence[i], env)
		}
		if condVal == IGNORE {
			return Eval(node.Consequence[i], env)
		}
	}
	// Shouldnt reach here ideally
	return nil
}

func doCondAndMatchExpEqual(condVal, matchVal object.Object) bool {
	condValPairs := condVal.(*object.Map).Pairs
	matchValPairs := matchVal.(*object.Map).Pairs
	condValLen := len(condValPairs)
	matchValLen := len(matchValPairs)
	if condValLen != matchValLen {
		return false
	}
	for condKey, condValue := range condValPairs {
		_, ok := matchValPairs[condKey]
		if !ok {
			return false
		}
		if condValue.Value == IGNORE {
			continue
		}
		val, ok := matchValPairs[condKey]
		if !ok {
			return false
		}
		if object.HashObject(val.Value) != object.HashObject(condValue.Value) {
			return false
		}
	}

	return true
}

func evalListCompLiteral(node *ast.ListCompLiteral, env *object.Environment) object.Object {
	l := lexer.New(node.NonEvaluatedProgram)
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return nil
	}
	_ = Eval(rootNode, env)
	someVal, ok := env.Get("__internal__")
	if !ok {
		return nil
	}
	return &object.ListCompLiteral{Elements: someVal.(*object.List).Elements}
}

func evalMapCompLiteral(node *ast.MapCompLiteral, env *object.Environment) object.Object {
	l := lexer.New(node.NonEvaluatedProgram)
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return nil
	}
	_ = Eval(rootNode, env)
	someVal, ok := env.Get("__internal__")
	if !ok {
		return nil
	}
	return &object.MapCompLiteral{Pairs: someVal.(*object.Map).Pairs}
}

func evalSetCompLiteral(node *ast.SetCompLiteral, env *object.Environment) object.Object {
	l := lexer.New(node.NonEvaluatedProgram)
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return nil
	}
	_ = Eval(rootNode, env)
	someVal, ok := env.Get("__internal__")
	if !ok {
		return nil
	}
	return &object.Set{Elements: someVal.(*object.Set).Elements}
}

var nestLevel = -1
var iterCount = []int{}
var cleanupTmpVar = make(map[string]bool)

// evalInExpression evaluates `in` statements when they refer to a loop context
func evalInExpression(node *ast.InfixExpression, env *object.Environment) object.Object {
	ident, ok := node.Left.(*ast.Identifier)
	if !ok {
		leftEval := Eval(node.Left, env)
		if isError(leftEval) {
			return leftEval
		}
		rightEval := Eval(node.Right, env)
		if isError(rightEval) {
			return rightEval
		}
		return evalInfixExpression(node.Operator, leftEval, rightEval)
	}
	// So if it is an identifier than we need to find out what we are trying to
	// unpack/bind our value to
	evaluatedRight := Eval(node.Right, env)
	if isError(evaluatedRight) {
		return evaluatedRight
	}
	if evaluatedRight.Type() == object.LIST_OBJ {
		// This is where we handle if its a list
		list := evaluatedRight.(*object.List).Elements
		_, ok = env.Get(ident.Value)
		if !ok {
			iterCount = append(iterCount, 0)
			nestLevel++
			env.Set(ident.Value, list[iterCount[nestLevel]])
			iterCount[nestLevel]++
			if len(list) == iterCount[nestLevel] {
				// Reset iteration for other items
				iterCount[nestLevel] = 0
				nestLevel--
				cleanupTmpVar[ident.Value] = true
				return FALSE
			}
			return TRUE
		}
		env.Set(ident.Value, list[iterCount[nestLevel]])
		iterCount[nestLevel]++
		if len(list) == iterCount[nestLevel] {
			// Reset iteration for other items
			iterCount[nestLevel] = 0
			nestLevel--
			cleanupTmpVar[ident.Value] = true
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
		_, ok = env.Get(ident.Value)
		if !ok {
			iterCount = append(iterCount, 0)
			nestLevel++
			env.Set(ident.Value, pairObjs[iterCount[nestLevel]])
			iterCount[nestLevel]++
			if len(pairObjs) == iterCount[nestLevel] {
				// Reset iteration for other items
				iterCount[nestLevel] = 0
				nestLevel--
				cleanupTmpVar[ident.Value] = true
				return FALSE
			}
			return TRUE
		}
		env.Set(ident.Value, pairObjs[iterCount[nestLevel]])
		iterCount[nestLevel]++
		if len(mapPairs) == iterCount[nestLevel] {
			// Reset iteration for other items
			iterCount[nestLevel] = 0
			nestLevel--
			cleanupTmpVar[ident.Value] = true
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
		_, ok = env.Get(ident.Value)
		if !ok {
			iterCount = append(iterCount, 0)
			nestLevel++
			env.Set(ident.Value, stringObjs[iterCount[nestLevel]])
			iterCount[nestLevel]++
			if len(chars) == iterCount[nestLevel] {
				// Reset iteration for other items
				iterCount[nestLevel] = 0
				nestLevel--
				cleanupTmpVar[ident.Value] = true
				return FALSE
			}
			return TRUE
		}
		env.Set(ident.Value, stringObjs[iterCount[nestLevel]])
		iterCount[nestLevel]++
		if len(stringObjs) == iterCount[nestLevel] {
			// Reset iteration for other items
			iterCount[nestLevel] = 0
			nestLevel--
			cleanupTmpVar[ident.Value] = true
			return FALSE
		}
		return TRUE
	}
	return newError("Expected List, Map, or String on right hand side. got %T", evaluatedRight.Type())
}

func evalForExpression(node *ast.ForExpression, env *object.Environment) object.Object {
	var evalBlock object.Object
	for {
		evalCond := Eval(node.Condition, env)
		if isError(evalCond) {
			return evalCond
		}
		ok := evalCond.(*object.Boolean).Value
		evalBlock = Eval(node.Consequence, env)
		if isError(evalBlock) {
			return evalBlock
		}
		// Cleanup any temporary for variables
		tmpMapCopy := cleanupTmpVar
		for k, v := range tmpMapCopy {
			if v {
				env.RemoveIdentifier(k)
				delete(cleanupTmpVar, k)
			}
		}
		// Still evaluate on the last run then break if its false
		if !ok {
			break
		}
	}
	return evalBlock
}

func evalAssignmentExpression(node *ast.AssignmentExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
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
		if ok := env.IsImmutable(rootObjIdent); ok {
			return newError("'" + rootObjIdent + "' is immutable")
		}
	}

	value := Eval(node.Value, env)
	if isError(value) {
		return value
	}

	// If its a simple identifier allow reassigning like so
	if ident, ok := node.Left.(*ast.Identifier); ok {
		if env.IsImmutable(ident.Value) {
			return newError("'" + ident.Value + "' is immutable")
		}
		switch node.Token.Literal {
		case "=":
			env.Set(ident.Value, value)
		case "+=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("+", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "-=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("-", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "*=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("*", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "/=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("/", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "//=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("//", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "**=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("**", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "&=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("&", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "|=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("|", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "~=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("~", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "<<=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("<<", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case ">>=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression(">>", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "%=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("%", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		case "^=":
			orig, ok := env.Get(ident.Value)
			if !ok {
				return newError("identifier '" + ident.String() + "' does not exist")
			}
			evaluated := evalInfixExpression("^", orig, value)
			if isError(evaluated) {
				return evaluated
			}
			env.Set(ident.Value, evaluated)
		default:
			return newError("assignment operator not supported `" + node.Token.Literal + "`")
		}
		// TODO: Figure out mutability properly, maybe need 2 environments
		// otherwise if its an index expression
	} else if ie, ok := node.Left.(*ast.IndexExpression); ok {
		// TODO: Handle all assignment tokens that apply
		obj := Eval(ie.Left, env)
		if isError(obj) {
			return obj
		}

		if list, ok := obj.(*object.List); ok {
			index := Eval(ie.Index, env)
			if isError(index) {
				return index
			}

			if idx, ok := index.(*object.Integer); ok {
				list.Elements[idx.Value] = value
			} else {
				return newError("cannot index list with %#v", index)
			}
		} else if mapObj, ok := obj.(*object.Map); ok {
			key := Eval(ie.Index, env)
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

func evalMapLiteral(node *ast.MapLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.MapPair)

	for keyNode, valueNode := range node.Pairs {
		ident, _ := keyNode.(*ast.Identifier)
		key := Eval(keyNode, env)
		if isError(key) && ident != nil {
			key = &object.Stringo{Value: ident.String()}
		} else if isError(key) {
			return key
		}

		mapKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as a map key: %s", key.Type())
		}

		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := mapKey.HashKey()
		pairs[hashed] = object.MapPair{Key: key, Value: value}
	}

	return &object.Map{Pairs: pairs}
}

// ArgToPassToBuiltin is the argument to be given to the builtin function
var ArgToPassToBuiltin object.Object = nil

func tryCreateValidBuiltinForDotCall(left, indx object.Object, leftNode ast.Expression) object.Object {
	// Try to see if the index being used is a builtin function
	if indx.Type() != object.STRING_OBJ {
		return nil
	}
	_, isBuiltin := builtins[indx.Inspect()]
	_, isStringBuiltin := stringbuiltins[indx.Inspect()]
	if !isBuiltin && !isStringBuiltin {
		return nil
	}
	// Allow either a string object or identifier to be passed to the builtin
	_, ok1 := left.(*object.Stringo)
	_, ok2 := leftNode.(*ast.Identifier)
	if !ok1 && !ok2 {
		return nil
	}

	ArgToPassToBuiltin = left
	// Return the builtin function object so that it can be used in the call
	// expression
	if isBuiltin {
		return &object.Builtin{
			Fun: builtins[indx.Inspect()].Fun,
		}
	}
	return &object.Builtin{
		Fun: stringbuiltins[indx.Inspect()].Fun,
	}
}

func evalIndexExpression(left, indx object.Object, env *object.Environment) object.Object {
	switch {
	case left.Type() == object.LIST_OBJ:
		return evalListIndexExpression(left, indx, env)
	case left.Type() == object.MAP_OBJ:
		return evalMapIndexExpression(left, indx, env)
	case left.Type() == object.MODULE_OBJ:
		return evalModuleIndexExpression(left, indx, env)
	default:
		// TODO: Support all other index expressions, such as member lookup and hash literals, sets, etc.
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalModuleIndexExpression(module, indx object.Object, env *object.Environment) object.Object {
	mod := module.(*object.Module)
	name := indx.(*object.Stringo).Value
	val, ok := mod.Env.Get(name)
	if !ok {
		return newError("failed to find '%s' in imported file '%s'", name, mod.Name)
	}
	return val
}

func evalMapIndexExpression(mapObject, indx object.Object, env *object.Environment) object.Object {
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

func evalListIndexExpression(list, indx object.Object, env *object.Environment) object.Object {
	listObj := list.(*object.List)
	var idx int64
	switch indx.Type() {
	case object.INTEGER_OBJ:
		idx = indx.(*object.Integer).Value
	case object.STRING_OBJ:
		stringVal := indx.(*object.Stringo).Value
		envVal, ok := env.Get(stringVal)
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
	// idx := indx.(*object.Integer).Value
	max := int64(len(listObj.Elements) - 1)
	if idx < 0 || idx > max {
		// TODO: possibly support -1 to get last element and negative numbers to go in reverse lookup of the list
		// This would make the code below a bit more complex but still fairly easy to implement
		return NULL
	}

	return listObj.Elements[idx]
}

func execCommand(arg0 string, args ...string) *exec.Cmd {
	if args == nil {
		if runtime.GOOS == "windows" {
			winArgs := []string{"/c"}
			winArgs = append(winArgs, arg0)
			return exec.Command("cmd", winArgs...)
		}
		return exec.Command(arg0)
	}
	if runtime.GOOS == "windows" {
		winArgs := []string{"/c"}
		winArgs = append(winArgs, arg0)
		winArgs = append(winArgs, args...)
		return exec.Command("cmd", winArgs...)
	}
	return exec.Command(arg0, args...)
}

func evalExecStringLiteral(execStringNode *ast.ExecStringLiteral, env *object.Environment) object.Object {
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

func evalStringWithInterpolation(stringNode *ast.StringLiteral, env *object.Environment) object.Object {
	someObjs := evalExpressions(stringNode.InterpolationValues, env)
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

func applyFunction(fun object.Object, args []object.Object, defaultArgs map[string]object.Object) object.Object {
	switch function := fun.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(function, args, defaultArgs)
		evaluated := Eval(function.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		if ArgToPassToBuiltin != nil {
			// prepend the argument to pass in to the front
			args = append([]object.Object{ArgToPassToBuiltin}, args...)
			// Unset the argument to pass in so itll be free next time we come to it
			ArgToPassToBuiltin = nil
		}
		return function.Fun(args...)
	default:
		return newError("not a function %s", function.Type())
	}
}

func extendFunctionEnv(fun *object.Function, args []object.Object, defaultArgs map[string]object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fun.Env)

	// If the arguments slice is the same length as the parameter list, then we have them all
	// so set them as normal
	if len(args) == len(fun.Parameters) {
		for paramIndx, param := range fun.Parameters {
			env.Set(param.Value, args[paramIndx])
		}
		setDefaultCallExpressionParameters(defaultArgs, env)
	} else if len(args) < len(fun.Parameters) {
		// loop and while less than the total parameters set environment variables accordingly
		argsIndx := 0
		for paramIndx, param := range fun.Parameters {
			if fun.DefaultParameters[paramIndx] == nil {
				if argsIndx < len(args) {
					env.Set(param.Value, args[argsIndx])
					argsIndx++
					continue
				}
			}
			env.Set(param.Value, fun.DefaultParameters[paramIndx])
		}
		setDefaultCallExpressionParameters(defaultArgs, env)
	}
	return env
}

func setDefaultCallExpressionParameters(defaultArgs map[string]object.Object, env *object.Environment) {
	for k, v := range defaultArgs {
		_, ok := env.Get(k)
		if ok {
			env.Set(k, v)
		}
	}
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
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

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	if builtin, ok := stringbuiltins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

// for now everything that is not null or false returns true
// TODO: Update this list to include non truthy for empty objects, lists, etc.
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	// TODO: implement these similar to how list is set up with one type checked on one side
	// and then check all the other sides in the next eval function
	switch {
	// These are the cases where they are the same type
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.BIG_INTEGER_OBJ:
		return evalBigIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalFloatIntInfixExpression(operator, left, right)
	case left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.BIG_FLOAT_OBJ:
		return evalBigFloatInfixExpression(operator, left, right)
	case left.Type() == object.UINTEGER_OBJ && right.Type() == object.UINTEGER_OBJ:
		return evalUintInfixExpression(operator, left, right)
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
		return evalSetInfixExpression(operator, left, right)
	// These are the cases where they differ
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.BIG_INTEGER_OBJ:
		return evalFloatBigIntegerInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.BIG_INTEGER_OBJ:
		return evalIntegerBigIntegerInfixExpression(operator, left, right)
	case left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalBigIntegerFloatInfixExpression(operator, left, right)
	case left.Type() == object.BIG_INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalBigIntegerIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.BIG_FLOAT_OBJ:
		return evalFloatBigFloatInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.BIG_FLOAT_OBJ:
		return evalIntegerBigFloatInfixExpression(operator, left, right)
	case left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalBigFloatFloatInfixExpression(operator, left, right)
	case left.Type() == object.BIG_FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalBigFloatIntegerInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalIntegerFloatInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.UINTEGER_OBJ:
		return evalIntegerUintegerInfixExpression(operator, left, right)
	case right.Type() == object.INTEGER_OBJ && left.Type() == object.FLOAT_OBJ:
		return evalIntegerFloatInfixExpression(operator, right, left)
	case right.Type() == object.INTEGER_OBJ && left.Type() == object.UINTEGER_OBJ:
		return evalIntegerUintegerInfixExpression(operator, right, left)
	case left.Type() == object.STRING_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalStringIntegerInfixExpression(operator, left, right)
	case right.Type() == object.STRING_OBJ && left.Type() == object.INTEGER_OBJ:
		return evalStringIntegerInfixExpression(operator, right, left)
	case left.Type() == object.STRING_OBJ && right.Type() == object.UINTEGER_OBJ:
		return evalStringUintegerInfixExpression(operator, left, right)
	case right.Type() == object.STRING_OBJ && left.Type() == object.UINTEGER_OBJ:
		return evalStringUintegerInfixExpression(operator, right, left)
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
		return evalListInfixExpression(operator, left, right)
	case right.Type() == object.SET_OBJ:
		return evalRightSideSetInfixExpression(operator, left, right)
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

func evalRightSideSetInfixExpression(operator string, left, right object.Object) object.Object {
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
func evalSetInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalBigFloatIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	rightVal := right.(*object.Integer).Value
	rightBigFloat := decimal.NewFromInt(rightVal)
	return evalBigFloatInfixExpression(operator, left, &object.BigFloat{Value: rightBigFloat})
}

func evalBigFloatFloatInfixExpression(operator string, left, right object.Object) object.Object {
	rightVal := right.(*object.Float).Value
	rightBigFloat := decimal.NewFromFloat(rightVal)
	return evalBigFloatInfixExpression(operator, left, &object.BigFloat{Value: rightBigFloat})
}

func evalIntegerBigFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	leftBigFloat := decimal.NewFromInt(leftVal)
	return evalBigFloatInfixExpression(operator, &object.BigFloat{Value: leftBigFloat}, right)
}

func evalFloatBigFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	leftBigFloat := decimal.NewFromFloat(leftVal)
	return evalBigFloatInfixExpression(operator, &object.BigFloat{Value: leftBigFloat}, right)
}

func evalBigIntegerIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	rightVal := right.(*object.Integer).Value
	rightBigInt := new(big.Int).SetInt64(rightVal)
	return evalBigIntegerInfixExpression(operator, left, &object.BigInteger{Value: rightBigInt})
}

func evalBigIntegerFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := right.(*object.BigInteger).Value
	rightVal := left.(*object.Float).Value
	leftBigFloat := decimal.NewFromBigInt(leftVal, 1)
	rightBigFloat := decimal.NewFromFloat(rightVal)

	return evalBigFloatInfixExpression(operator, &object.BigFloat{Value: leftBigFloat}, &object.BigFloat{Value: rightBigFloat})
}

func evalIntegerBigIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	leftBigInt := new(big.Int).SetInt64(leftVal)
	return evalBigIntegerInfixExpression(operator, &object.BigInteger{Value: leftBigInt}, right)
}

func evalFloatBigIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.BigInteger).Value
	leftBigFloat := decimal.NewFromFloat(leftVal)
	rightBigFloat := decimal.NewFromBigInt(rightVal, 1)
	return evalBigFloatInfixExpression(operator, &object.BigFloat{Value: leftBigFloat}, &object.BigFloat{Value: rightBigFloat})
}

func evalBigFloatInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalBigIntegerInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalStringIntegerInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalStringUintegerInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalIntegerUintegerInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalIntegerFloatInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalListInfixExpression(operator string, left, right object.Object) object.Object {
	switch right.Type() {
	case object.INTEGER_OBJ:
		return evalListIntegerInfixExpression(operator, left, right)
	case object.LIST_OBJ:
		return evalListListInfixExpression(operator, left, right)
	case object.SET_OBJ:
		return evalRightSideSetInfixExpression(operator, left, right)
	default:
		return newError("unhandled type for list infix expressions: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalListListInfixExpression(operator string, left, right object.Object) object.Object {
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

func twoListsEqual(leftList, rightList *object.List) bool {
	// This is a deep equality expensive function
	if object.HashObject(leftList) == object.HashObject(rightList) {
		return true
	}
	return false
}

func evalListIntegerInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "not":
		return evalNotOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	case "~":
		return evalBitwiseNotOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	checkOverflow := func(leftVal, rightVal int64) bool {
		result := leftVal + rightVal
		if result-leftVal != rightVal {
			return true
		}
		return false
	}
	checkUnderflow := func(leftVal, rightVal int64) bool {
		result := leftVal - rightVal
		if result+rightVal != leftVal {
			return true
		}
		return false
	}

	checkOverflowMul := func(leftVal, rightVal int64) bool {
		if leftVal == 0 || rightVal == 0 || leftVal == 1 || rightVal == 1 {
			return false
		}
		if leftVal == math.MinInt64 || rightVal == math.MinInt64 {
			return true
		}
		result := leftVal * rightVal
		if result/rightVal != leftVal {
			return true
		}
		return false
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
		return evalIntegerRange(leftVal, rightVal)
	case "..<":
		return evalIntegerNonIncRange(leftVal, rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerRange(leftVal, rightVal int64) object.Object {
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

func evalIntegerNonIncRange(leftVal, rightVal int64) object.Object {
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

func evalFloatIntInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Integer).Value
	fmt.Println(leftVal, rightVal)
	panic("TODO: Handle float on left and int on right")
	// switch operator {
	// case "+":
	// 	return &object.Float{Value: leftVal + rightVal}
	// case "-":
	// 	return &object.Float{Value: leftVal - rightVal}
	// case "/":
	// 	return &object.Float{Value: leftVal / rightVal}
	// case "*":
	// 	return &object.Float{Value: leftVal * rightVal}
	// case "**":
	// 	return &object.Float{Value: math.Pow(leftVal, rightVal)}
	// case "//":
	// 	return &object.Float{Value: float64(int64(leftVal) / int64(rightVal))}
	// case "%":
	// 	return &object.Float{Value: math.Mod(leftVal, rightVal)}
	// case "<":
	// 	return nativeToBooleanObject(leftVal < rightVal)
	// case ">":
	// 	return nativeToBooleanObject(leftVal > rightVal)
	// case "<=":
	// 	return nativeToBooleanObject(leftVal <= rightVal)
	// case ">=":
	// 	return nativeToBooleanObject(leftVal >= rightVal)
	// case "==":
	// 	return nativeToBooleanObject(leftVal == rightVal)
	// case "!=":
	// 	return nativeToBooleanObject(leftVal != rightVal)
	// default:
	// 	return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	// }
}

func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalUintInfixExpression(operator string, left, right object.Object) object.Object {
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

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
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

func evalBitwiseNotOperatorExpression(right object.Object) object.Object {
	value := right.(*object.UInteger).Value
	return &object.UInteger{Value: 0xFFFFFFFFFFFFFFFF ^ value}
}

func evalNotOperatorExpression(right object.Object) object.Object {
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

func nativeToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalProgramStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range stmts {
		result = Eval(stmt, env)
	}

	return result
}

func evalStatements(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range stmts {
		result = Eval(stmt, env)
		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
	}
	return result
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range program.Statements {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range block.Statements {
		result = Eval(stmt, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

// newError is the wrapper function to add an error to the evaluator
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// isError is the helper function to determine if an object is an error
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
