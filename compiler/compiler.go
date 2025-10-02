package compiler

import (
	"blue/ast"
	"blue/code"
	"blue/object"
	"fmt"
	"regexp"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.InfixExpression:
		if node.Operator == "<" || node.Operator == "<=" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.compileInfixExpression(node.Operator)
			return nil
		}
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		err = c.compileInfixExpression(node.Operator)
		if err != nil {
			return err
		}
	case *ast.IntegerLiteral:
		literal := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.BigIntegerLiteral:
		literal := &object.BigInteger{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.HexLiteral:
		literal := &object.UInteger{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.OctalLiteral:
		literal := &object.UInteger{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.BinaryLiteral:
		literal := &object.UInteger{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.UIntegerLiteral:
		literal := &object.UInteger{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.FloatLiteral:
		literal := &object.Float{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.BigFloatLiteral:
		literal := &object.BigFloat{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.RegexLiteral:
		r, err := regexp.Compile(node.Token.Literal)
		if err != nil {
			panic("TODO: How are errors returned")
			// return newError("failed to create regex literal %q", node.TokenLiteral())
		}
		literal := &object.Regex{Value: r}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.StringLiteral:
		if len(node.InterpolationValues) == 0 {
			literal := &object.Stringo{Value: node.Value}
			c.emit(code.OpConstant, c.addConstant(literal))
		}
		panic("TODO: Implement string with interpolation")
	// obj := e.evalStringWithInterpolation(node)
	// if isError(obj) {
	// 	e.ErrorTokens.Push(node.Token)
	// }
	// return obj
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.Null:
		c.emit(code.OpNull)
	}
	return nil
}

func nativeToBooleanObject(ok bool) object.Object {
	if ok {
		return object.TRUE
	} else {
		return object.FALSE
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

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
