package compiler

import (
	"blue/ast"
	"blue/parser"
	"blue/token"
	"fmt"
)

var emptySym = Symbol{}

func (c *Compiler) isIdentOnLeftInIterableOnRight(cond ast.Expression) (bool, Symbol, ast.Expression) {
	infix, ok := cond.(*ast.InfixExpression)
	if !ok {
		return ok, emptySym, nil
	}
	if infix.Operator != "in" {
		return false, emptySym, nil
	}
	ident, ok1 := infix.Left.(*ast.Identifier)
	if !ok1 {
		return false, emptySym, nil
	}
	sym := c.symbolTable.Define(ident.Value, false)
	return true, sym, infix.Right
}

var _equalToken = token.Token{Type: token.EQ, Literal: "="}
var _plusEqualToken = token.Token{Type: token.PLUSEQ, Literal: "+="}
var _zeroAstLit = &ast.IntegerLiteral{Value: 0}
var _oneAstLit = &ast.IntegerLiteral{Value: 1}
var _lenIdent = &ast.Identifier{Value: "len"}
var _getIdent = &ast.Identifier{Value: "_get_"}

func (c *Compiler) compileIdentInIterableFor(sym Symbol, node *ast.ForStatement, right ast.Expression) error {
	// Hoist the iterable expression into its own variable so it is only
	// evaluated once before the loop starts. Compiling it into the loop
	// condition/setter below would rebuild things like ranges or list
	// literals on every single iteration.
	iterableIdent := &ast.Identifier{Value: fmt.Sprintf("__iterable_%s", sym.Name)}
	iterableVs := &ast.VarStatement{
		Names: []*ast.Identifier{iterableIdent},
		Value: right,
	}
	err := c.Compile(iterableVs)
	if err != nil {
		return err
	}
	indexIdent := &ast.Identifier{Value: fmt.Sprintf("__index_%s", sym.Name)}
	vs := &ast.VarStatement{
		Names: []*ast.Identifier{indexIdent},
		Value: _zeroAstLit,
	}
	err = c.Compile(vs)
	if err != nil {
		return err
	}
	node.Condition = &ast.InfixExpression{
		Left:     indexIdent,
		Operator: "<",
		Right: &ast.CallExpression{
			Function:  _lenIdent,
			Arguments: []ast.Expression{iterableIdent},
		},
	}
	node.PostExp = &ast.AssignmentExpression{
		Token: _plusEqualToken,
		Left:  indexIdent,
		Value: _oneAstLit,
	}
	symIdent := &ast.Identifier{Value: sym.Name}
	node.IterableSetters = []ast.Expression{
		&ast.AssignmentExpression{
			Token: _equalToken,
			Left:  symIdent,
			Value: &ast.CallExpression{
				Function:  _getIdent,
				Arguments: []ast.Expression{iterableIdent, indexIdent},
			},
		},
	}
	return nil
}

func (c *Compiler) isListIdentsOnLeftInIterableOnRight(cond ast.Expression) (bool, Symbol, Symbol, ast.Expression) {
	infix, ok := cond.(*ast.InfixExpression)
	if !ok {
		return ok, emptySym, emptySym, nil
	}
	if infix.Operator != "in" {
		return false, emptySym, emptySym, nil
	}
	l, ok := infix.Left.(*ast.ListLiteral)
	if !ok {
		return false, emptySym, emptySym, nil
	}
	if len(l.Elements) != 2 {
		return false, emptySym, emptySym, nil
	}
	ident1, ok := l.Elements[0].(*ast.Identifier)
	if !ok {
		return false, emptySym, emptySym, nil
	}
	ident2, ok := l.Elements[1].(*ast.Identifier)
	if !ok {
		return false, emptySym, emptySym, nil
	}
	sym1 := c.symbolTable.Define(ident1.Value, false)
	sym2 := c.symbolTable.Define(ident2.Value, false)
	return true, sym1, sym2, infix.Right
}

var _trueBool = &ast.Boolean{Value: true}

func (c *Compiler) compileListIdentsInIterableFor(sym1 Symbol, sym2 Symbol, node *ast.ForStatement, right ast.Expression) error {
	// Hoist the iterable expression so it is only evaluated once before the
	// loop starts (see compileIdentInIterableFor)
	iterableIdent := &ast.Identifier{Value: fmt.Sprintf("__iterable_%s%s", sym1.Name, sym2.Name)}
	iterableVs := &ast.VarStatement{
		Names: []*ast.Identifier{iterableIdent},
		Value: right,
	}
	err := c.Compile(iterableVs)
	if err != nil {
		return err
	}
	indexIdent := &ast.Identifier{Value: fmt.Sprintf("__index_%s%s", sym1.Name, sym2.Name)}
	vs := &ast.VarStatement{
		Names: []*ast.Identifier{indexIdent},
		Value: _zeroAstLit,
	}
	err = c.Compile(vs)
	if err != nil {
		return err
	}
	indexedIdent := &ast.Identifier{Value: fmt.Sprintf("__indexed_%s%s", sym1.Name, sym2.Name)}
	vs1 := &ast.VarStatement{
		Names: []*ast.Identifier{indexedIdent},
		Value: parser.AST_NULL,
	}
	err = c.Compile(vs1)
	if err != nil {
		return err
	}
	node.Condition = &ast.InfixExpression{
		Left:     indexIdent,
		Operator: "<",
		Right: &ast.CallExpression{
			Function:  _lenIdent,
			Arguments: []ast.Expression{iterableIdent},
		},
	}
	node.PostExp = &ast.AssignmentExpression{
		Token: _plusEqualToken,
		Left:  indexIdent,
		Value: _oneAstLit,
	}
	sym1Ident := &ast.Identifier{Value: sym1.Name}
	sym2Ident := &ast.Identifier{Value: sym2.Name}
	node.IterableSetters = []ast.Expression{
		&ast.AssignmentExpression{
			Token: _equalToken,
			Left:  indexedIdent,
			Value: &ast.CallExpression{
				Function:  _getIdent,
				Arguments: []ast.Expression{iterableIdent, indexIdent, _trueBool},
			},
		},
		&ast.AssignmentExpression{
			Token: _equalToken,
			Left:  sym1Ident,
			Value: &ast.IndexExpression{
				Left:  indexedIdent,
				Index: _zeroAstLit,
			},
		},
		&ast.AssignmentExpression{
			Token: _equalToken,
			Left:  sym2Ident,
			Value: &ast.IndexExpression{
				Left:  indexedIdent,
				Index: _oneAstLit,
			},
		},
	}
	return nil
}
