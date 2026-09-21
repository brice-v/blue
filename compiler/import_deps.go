package compiler

import (
	"blue/ast"
	"blue/lexer"
	"blue/parser"
	"fmt"
	"reflect"
)

// importDepScan collects the identifiers a piece of a module references. A
// selective import keeps the declaration bound to every requested name plus
// whatever those declarations reach transitively, so the scan is deliberately
// an over approximation: a local variable that happens to share a name with a
// top level declaration only ever causes an extra declaration to be kept.
type importDepScan struct {
	names    map[string]struct{}
	usesEval bool
}

func newImportDepScan() *importDepScan {
	return &importDepScan{names: map[string]struct{}{}}
}

func (s *importDepScan) walk(node ast.Node) {
	// Optional AST fields are frequently a typed nil pointer (for example an
	// if expression's alternative), which a plain nil check does not catch.
	if node == nil {
		return
	}
	if value := reflect.ValueOf(node); value.Kind() == reflect.Pointer && value.IsNil() {
		return
	}
	switch n := node.(type) {
	case *ast.Program:
		for _, statement := range n.Statements {
			s.walk(statement)
		}
	case *ast.BlockStatement:
		for _, statement := range n.Statements {
			s.walk(statement)
		}
	case *ast.ExpressionStatement:
		s.walk(n.Expression)
	case *ast.ImportStatement:
		// Imports are always kept, so nothing to collect here.
	case *ast.BreakStatement, *ast.ContinueStatement:
	case *ast.VarStatement:
		s.walkNames(n.Names, n.KeyValueNames)
		s.walk(n.Value)
	case *ast.ValStatement:
		s.walkNames(n.Names, n.KeyValueNames)
		s.walk(n.Value)
	case *ast.FunctionStatement:
		for _, e := range n.ParameterExpressions {
			s.walk(e)
		}
		s.walk(n.Body)
	case *ast.ReturnStatement:
		s.walk(n.ReturnValue)
	case *ast.TryCatchStatement:
		s.walk(n.TryBlock)
		s.walk(n.CatchBlock)
		s.walk(n.FinallyBlock)
	case *ast.ForStatement:
		s.walk(n.Condition)
		s.walk(n.Consequence)
		s.walk(n.Initializer)
		s.walk(n.PostExp)
		for _, setter := range n.IterableSetters {
			s.walk(setter)
		}
	case *ast.Identifier:
		s.names[n.Value] = struct{}{}
	case *ast.PrefixExpression:
		s.walk(n.Right)
	case *ast.PostfixExpression:
		s.walk(n.Left)
	case *ast.InfixExpression:
		s.walk(n.Left)
		s.walk(n.Right)
	case *ast.IfExpression:
		for _, condition := range n.Conditions {
			s.walk(condition)
		}
		for _, consequence := range n.Consequences {
			s.walk(consequence)
		}
		s.walk(n.Alternative)
	case *ast.MatchExpression:
		s.walk(n.OptionalValue)
		for _, conditions := range n.Conditions {
			for _, condition := range conditions {
				s.walk(condition)
			}
		}
		for _, consequence := range n.Consequences {
			s.walk(consequence)
		}
	case *ast.CallExpression:
		s.walk(n.Function)
		for _, arg := range n.Arguments {
			s.walk(arg)
		}
		for _, arg := range n.DefaultArguments {
			s.walk(arg)
		}
	case *ast.IndexExpression:
		s.walk(n.Left)
		s.walk(n.Index)
	case *ast.AssignmentExpression:
		s.walk(n.Left)
		s.walk(n.Value)
	case *ast.EvalExpression:
		s.usesEval = true
		s.walk(n.StrToEval)
	case *ast.SpawnExpression:
		for _, arg := range n.Arguments {
			s.walk(arg)
		}
	case *ast.DeferExpression:
		for _, arg := range n.Arguments {
			s.walk(arg)
		}
	case *ast.FunctionLiteral:
		for _, e := range n.ParameterExpressions {
			s.walk(e)
		}
		s.walk(n.Body)
	case *ast.StringLiteral:
		for _, e := range n.InterpolationValues {
			s.walk(e)
		}
	case *ast.ListLiteral:
		for _, e := range n.Elements {
			s.walk(e)
		}
	case *ast.SetLiteral:
		for _, e := range n.Elements {
			s.walk(e)
		}
	case *ast.MapLiteral:
		for key, value := range n.Pairs {
			s.walk(key)
			s.walk(value)
		}
	case *ast.StructLiteral:
		for _, value := range n.Values {
			s.walk(value)
		}
	case *ast.ListCompLiteral:
		s.walkCompProgram(n.NonEvaluatedProgram)
	case *ast.SetCompLiteral:
		s.walkCompProgram(n.NonEvaluatedProgram)
	case *ast.MapCompLiteral:
		s.walkCompProgram(n.NonEvaluatedProgram)
	default:
		// Literals, null, booleans, self, regex and exec strings reference
		// nothing that could be a top level declaration.
	}
}

func (s *importDepScan) walkNames(names []*ast.Identifier, keyValueNames map[ast.Expression]*ast.Identifier) {
	for _, name := range names {
		s.walk(name)
	}
	for _, name := range keyValueNames {
		s.walk(name)
	}
}

// walkCompProgram scans the deferred source of a list/set/map comprehension.
// The compiler parses that string at compile time, so a declaration referenced
// only from inside a comprehension still has to be kept. Parse failures are
// ignored here and reported later when the comprehension is actually compiled.
func (s *importDepScan) walkCompProgram(nonEvaluatedProgram string) {
	if nonEvaluatedProgram == "" {
		return
	}
	l := lexer.New(nonEvaluatedProgram, "<import-dep-scan>")
	p := parser.New(l)
	program := p.ParseProgram()
	if p.HasErrors() {
		return
	}
	s.walk(program)
}

// compileModuleForImportRoots compiles only the part of a module that a
// `from mod import {a, b}` needs. It keeps the declarations bound to the
// requested names, every top level declaration those reference transitively, and
// the module's top level imports and other non declaration statements so that
// module initialisation still runs.
//
// Modules that use eval are compiled in full: a dynamically built program can
// reach any name, which static analysis cannot see.
func (c *Compiler) compileModuleForImportRoots(program *ast.Program, roots []*ast.Identifier, modName string) error {
	wholeProgramScan := newImportDepScan()
	wholeProgramScan.walk(program)
	if wholeProgramScan.usesEval {
		return c.Compile(program)
	}

	declIndex := map[string]int{}
	for i, statement := range program.Statements {
		switch node := statement.(type) {
		case *ast.FunctionStatement:
			if node.Name != nil {
				declIndex[node.Name.Value] = i
			}
		case ast.VarValStatement:
			for _, name := range node.VVNames() {
				declIndex[name.Value] = i
			}
			for _, name := range node.VVKeyValueNames() {
				declIndex[name.Value] = i
			}
		}
	}

	needed := map[int]struct{}{}
	queue := []int{}
	for _, ident := range roots {
		index, ok := declIndex[ident.Value]
		if !ok {
			return fmt.Errorf("failed to import '%s' from '%s': no such top level declaration", ident.Value, modName)
		}
		queue = append(queue, index)
	}
	// Keep every statement that is not a declaration so the module body (imports
	// and initialisation code) still runs, and scan it like any other retained
	// statement so the declarations it uses come along too.
	for i, statement := range program.Statements {
		switch statement.(type) {
		case *ast.FunctionStatement, ast.VarValStatement:
		default:
			queue = append(queue, i)
		}
	}
	for len(queue) > 0 {
		index := queue[0]
		queue = queue[1:]
		if _, ok := needed[index]; ok {
			continue
		}
		needed[index] = struct{}{}
		deps := newImportDepScan()
		deps.walk(program.Statements[index])
		for name := range deps.names {
			if depIndex, ok := declIndex[name]; ok {
				if _, has := needed[depIndex]; !has {
					queue = append(queue, depIndex)
				}
			}
		}
	}

	filtered := make([]ast.Statement, 0, len(needed))
	for i, statement := range program.Statements {
		if _, ok := needed[i]; ok {
			filtered = append(filtered, statement)
		}
	}
	return c.Compile(&ast.Program{Statements: filtered, HelpStrTokens: program.HelpStrTokens})
}
