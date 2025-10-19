package compiler

import (
	"blue/ast"
	"blue/code"
	"fmt"
)

func (c *Compiler) compileInfixExpression(operator string) error {
	switch operator {
	case "+":
		c.emit(code.OpAdd)
	case "-":
		c.emit(code.OpMinus)
	case "*":
		c.emit(code.OpStar)
	case "**":
		c.emit(code.OpPow)
	case "/":
		c.emit(code.OpDiv)
	case "//":
		c.emit(code.OpFlDiv)
	case "%":
		c.emit(code.OpPercent)
	case "^":
		c.emit(code.OpCarat)
	case "&":
		c.emit(code.OpAmpersand)
	case "|":
		c.emit(code.OpPipe)
	case "in":
		c.emit(code.OpIn)
	case "notin":
		c.emit(code.OpNotin)
	case "..":
		c.emit(code.OpRange)
	case "..<":
		c.emit(code.OpNonIncRange)
	case ">>":
		c.emit(code.OpRshift)
	case "<<":
		c.emit(code.OpLshift)
	case "==":
		c.emit(code.OpEqual)
	case "!=":
		c.emit(code.OpNotEqual)
	case "||", "or":
		c.emit(code.OpOr)
	case "&&", "and":
		c.emit(code.OpAnd)
	case ">=", "<=":
		c.emit(code.OpGreaterThanOrEqual)
	case ">", "<":
		c.emit(code.OpGreaterThan)
	default:
		return fmt.Errorf("unsupported operator: %s", operator)
	}
	return nil
}

func (c *Compiler) compileIfExpression(node *ast.IfExpression) error {
	allEndingJumpPos := []int{}
	for i := range node.Conditions {
		err := c.Compile(node.Conditions[i])
		if err != nil {
			return err
		}
		// Emit an `OpJumpNotTruthy` with a bogus value
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)
		err = c.Compile(node.Consequences[i])
		if err != nil {
			return err
		}
		if c.lastInstructionIsSet() {
			c.emit(code.OpNull)
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}
		// Emit an `OpJump` with a bogus value
		jumpPos := c.emit(code.OpJump, 9999)
		allEndingJumpPos = append(allEndingJumpPos, jumpPos)
		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)
	}
	if node.Alternative == nil {
		c.emit(code.OpNull)
	} else {
		err := c.Compile(node.Alternative)
		if err != nil {
			return err
		}
		if c.lastInstructionIsSet() {
			c.emit(code.OpNull)
		}
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}
	}
	afterAlternativePos := len(c.currentInstructions())
	for _, jumpPos := range allEndingJumpPos {
		c.changeOperand(jumpPos, afterAlternativePos)
	}
	return nil
}

func (c *Compiler) loadSymbol(s Symbol) {
	if s.Immutable {
		switch s.Scope {
		case GlobalScope:
			c.emit(code.OpGetGlobalImm, s.Index)
		case LocalScope:
			c.emit(code.OpGetLocalImm, s.Index)
		case BuiltinScope:
			c.emit(code.OpGetBuiltin, s.Index)
		case FreeScope:
			c.emit(code.OpGetFreeImm, s.Index)
		}
	} else {
		switch s.Scope {
		case GlobalScope:
			c.emit(code.OpGetGlobal, s.Index)
		case LocalScope:
			c.emit(code.OpGetLocal, s.Index)
		case BuiltinScope:
			c.emit(code.OpGetBuiltin, s.Index)
		case FreeScope:
			c.emit(code.OpGetFree, s.Index)
		}
	}
}

func (c *Compiler) compileAssignmentExpression(node *ast.AssignmentExpression) error {
	if ident, ok := node.Left.(*ast.Identifier); ok {
		return c.compileAssignmentWithIdent(ident, node.Token.Literal, node.Value)
	} else if indexExp, ok := node.Left.(*ast.IndexExpression); ok {
		return c.compileAssignmentWithIndex(indexExp, node.Token.Literal, node.Value)
	}
	return fmt.Errorf("left side type not supported for assignment expression: %T", node.Left)
}

func (c *Compiler) compileAssignmentWithIdent(ident *ast.Identifier, operator string, v ast.Expression) error {
	sym, ok := c.symbolTable.Resolve(ident.Value)
	if !ok {
		return fmt.Errorf("identifier not found: %s", ident.Value)
	}
	if sym.Immutable {
		return fmt.Errorf("'%s' is immutable", ident.Value)
	}
	// Always compile right hand side value first
	err := c.Compile(v)
	if err != nil {
		return err
	}
	// If its not assignment then compile as if this is an infix expression
	if operator != "=" {
		// Compile "get" for the variable being assigned to
		c.loadSymbol(sym)
		op, ok := assignmentToInfixOperator[operator]
		if !ok {
			return fmt.Errorf("invalid assignment operator: %s", operator)
		}
		err := c.compileInfixExpression(op)
		if err != nil {
			return err
		}
	}
	if sym.Scope == GlobalScope {
		c.emit(code.OpSetGlobal, sym.Index)
	} else {
		c.emit(code.OpSetLocal, sym.Index)
	}
	c.emit(code.OpNull)
	return nil
}

var assignmentToInfixOperator = map[string]string{
	"+=":  "+",
	"-=":  "-",
	"*=":  "*",
	"/=":  "/",
	"//=": "//",
	"**=": "**",
	"&=":  "&",
	"|=":  "|",
	"~=":  "~",
	"<<=": "<<",
	">>=": ">>",
	"%=":  "%",
	"^=":  "^",
	"&&=": "&&",
	"||=": "||",
}

func (c *Compiler) compileAssignmentWithIndex(index *ast.IndexExpression, operator string, v ast.Expression) error {
	rootIdent, ok := getRootIdent(index)
	if !ok {
		return fmt.Errorf("could not find identifier for assignmenet")
	}
	sym, ok := c.symbolTable.Resolve(rootIdent.Value)
	if !ok {
		return fmt.Errorf("identifier not found: %s", rootIdent.Value)
	}
	if sym.Immutable {
		return fmt.Errorf("'%s' is immutable", rootIdent.Value)
	}
	err := c.Compile(v)
	if err != nil {
		return err
	}
	if operator != "=" {
		err = c.Compile(index)
		if err != nil {
			return err
		}
		op, ok := assignmentToInfixOperator[operator]
		if !ok {
			return fmt.Errorf("invalid assignment operator: %s", operator)
		}
		c.compileInfixExpression(op)
	}
	err = c.Compile(index)
	if err != nil {
		return err
	}
	if c.lastInstructionIs(code.OpIndex) {
		c.removeLastInstruction()
	}
	c.emit(code.OpIndexSet)
	return nil
}

func getRootIdent(node *ast.IndexExpression) (*ast.Identifier, bool) {
	left := node.Left
	for {
		if ident, ok := left.(*ast.Identifier); ok {
			left = ident
			break
		} else if indx, ok := left.(*ast.IndexExpression); ok {
			left = indx.Left
		} else {
			return nil, false
		}
	}
	ident, ok := left.(*ast.Identifier)
	if !ok {
		return nil, false
	}
	return ident, ok
}

// func (c *Compiler) compileForStatement(node *ast.ForStatement) error {
// 	err := c.Compile(node.Condition)
// 	if err != nil {
// 		return err
// 	}
// 	// Emit an `OpJumpNotTruthy` with a bogus value
// 	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)
// 	err = c.Compile(node.Consequence)
// 	if err != nil {
// 		return err
// 	}
// 	if c.lastInstructionIs(code.OpPop) {
// 		c.removeLastPop()
// 	}
// 	afterConsequencePos := len(c.currentInstructions())
// 	c.changeOperand(jumpNotTruthyPos, afterConsequencePos)
// }
