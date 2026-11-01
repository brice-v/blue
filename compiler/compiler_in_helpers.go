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
	if infix, ok := right.(*ast.InfixExpression); ok && (infix.Operator == ".." || infix.Operator == "..<") {
		if _, ok := infix.Left.(*ast.IntegerLiteral); ok {
			return c.compileRangeIterableFor(sym, node, infix)
		}
	}
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

func (c *Compiler) compileRangeIterableFor(sym Symbol, node *ast.ForStatement, right *ast.InfixExpression) error {
	prefix := fmt.Sprintf("__range_%s_%d", sym.Name, c.forIndex)
	startIdent := &ast.Identifier{Value: prefix + "_start"}
	endIdent := &ast.Identifier{Value: prefix + "_end"}
	stepIdent := &ast.Identifier{Value: prefix + "_step"}
	countIdent := &ast.Identifier{Value: prefix + "_count"}
	idxIdent := &ast.Identifier{Value: prefix + "_index"}
	symIdent := &ast.Identifier{Value: sym.Name}

	compileVar := func(name *ast.Identifier, value ast.Expression) error {
		return c.Compile(&ast.VarStatement{
			Token: token.Token{Type: token.VAR, Literal: "var"},
			Names: []*ast.Identifier{name},
			Value: value,
		})
	}
	if err := compileVar(startIdent, right.Left); err != nil {
		return err
	}
	if err := compileVar(endIdent, right.Right); err != nil {
		return err
	}
	if err := compileVar(stepIdent, _oneAstLit); err != nil {
		return err
	}

	extra := ast.Expression(_oneAstLit)
	if right.Operator == "..<" {
		extra = _zeroAstLit
	}
	countFor := func(a, b *ast.Identifier) ast.Expression {
		return &ast.InfixExpression{
			Left:     &ast.InfixExpression{Left: b, Operator: "-", Right: a},
			Operator: "+",
			Right:    extra,
		}
	}
	if err := compileVar(countIdent, countFor(startIdent, endIdent)); err != nil {
		return err
	}

	stepDown := &ast.ExpressionStatement{Expression: &ast.AssignmentExpression{
		Token: _equalToken,
		Left:  stepIdent,
		Value: &ast.IntegerLiteral{Value: -1},
	}}
	countDown := &ast.ExpressionStatement{Expression: &ast.AssignmentExpression{
		Token: _equalToken,
		Left:  countIdent,
		Value: countFor(endIdent, startIdent),
	}}
	stepIf := &ast.ExpressionStatement{
		Expression: &ast.IfExpression{
			Conditions: []ast.Expression{
				&ast.InfixExpression{Left: startIdent, Operator: ">", Right: endIdent},
			},
			Consequences: []*ast.BlockStatement{
				{Statements: []ast.Statement{stepDown, countDown}},
			},
		},
	}
	if err := c.Compile(stepIf); err != nil {
		return err
	}
	if err := compileVar(idxIdent, _zeroAstLit); err != nil {
		return err
	}

	node.Condition = &ast.InfixExpression{Left: idxIdent, Operator: "<", Right: countIdent}
	node.IterableSetters = []ast.Expression{
		&ast.AssignmentExpression{
			Token: _equalToken,
			Left:  symIdent,
			Value: &ast.InfixExpression{
				Left:     startIdent,
				Operator: "+",
				Right: &ast.InfixExpression{
					Left:     idxIdent,
					Operator: "*",
					Right:    stepIdent,
				},
			},
		},
	}
	node.PostExp = &ast.AssignmentExpression{
		Token: _plusEqualToken,
		Left:  idxIdent,
		Value: _oneAstLit,
	}
	return nil
}
