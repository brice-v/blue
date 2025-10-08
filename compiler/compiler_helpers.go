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
