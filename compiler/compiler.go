package compiler

import (
	"blue/ast"
	"blue/code"
	"blue/object"
	"fmt"
	"log"
	"regexp"
	"sort"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	constants []object.Object

	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	return &Compiler{
		constants:   []object.Object{},
		symbolTable: NewSymbolTable(),
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = updatedInstructions
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}
	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction
	old := c.currentInstructions()
	new := old[:last.Position]
	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	instructions := c.currentInstructions()
	for i := range newInstruction {
		instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer
	return instructions
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
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "!", "not":
			c.emit(code.OpNot)
		case "-":
			c.emit(code.OpNeg)
		case "~":
			c.emit(code.OpTilde)
		case "<<":
			c.emit(code.OpLshiftPre)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.PostfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		if node.Operator == ">>" {
			c.emit(code.OpRshiftPost)
		} else {
			return fmt.Errorf("unknown operator %s", node.Operator)
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
		} else {
			panic("TODO: Implement string with interpolation")
		}
	// obj := e.evalStringWithInterpolation(node)
	// if isError(obj) {
	// 	e.ErrorTokens.Push(node.Token)
	// }
	// return obj
	case *ast.ListLiteral:
		for _, exp := range node.Elements {
			err := c.Compile(exp)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpList, len(node.Elements))
	case *ast.SetLiteral:
		for _, exp := range node.Elements {
			err := c.Compile(exp)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpSet, len(node.Elements))
	case *ast.MapLiteral:
		indices := make([]int, 0, len(node.PairsIndex))
		for k := range node.PairsIndex {
			indices = append(indices, k)
		}
		sort.Ints(indices)
		for _, i := range indices {
			keyNode := node.PairsIndex[i]
			err := c.Compile(keyNode)
			if err != nil {
				return err
			}
			valueNode := node.Pairs[keyNode]
			err = c.Compile(valueNode)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpMap, len(node.Pairs)*2)
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.Null:
		c.emit(code.OpNull)
	case *ast.IfExpression:
		err := c.compileIfExpression(node)
		if err != nil {
			return err
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.VarStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		if node.IsListDestructor || node.IsMapDestructor {
			return fmt.Errorf("List/Map Destructor not yet supported, failed to compile %#+v", node)
		}
		if len(node.Names) > 1 {
			return fmt.Errorf("multiple identifiers to define, not supported yet %#+v", node.Names)
		}
		symbol := c.symbolTable.Define(node.Names[0].Value, false)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
	case *ast.ValStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		if node.IsListDestructor || node.IsMapDestructor {
			return fmt.Errorf("List/Map Destructor not yet supported, failed to compile %#+v", node)
		}
		if len(node.Names) > 1 {
			return fmt.Errorf("multiple identifiers to define, not supported yet %#+v", node.Names)
		}
		symbol := c.symbolTable.Define(node.Names[0].Value, true)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobalImm, symbol.Index)
		} else {
			c.emit(code.OpSetLocalImm, symbol.Index)
		}
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		if symbol.Scope == GlobalScope {
			if symbol.Immutable {
				c.emit(code.OpGetGlobalImm, symbol.Index)
			} else {
				c.emit(code.OpGetGlobal, symbol.Index)
			}
		} else {
			if symbol.Immutable {
				c.emit(code.OpGetLocalImm, symbol.Index)
			} else {
				c.emit(code.OpGetLocal, symbol.Index)
			}
		}
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Index)
		if err != nil {
			return err
		}
		c.emit(code.OpIndex)
	case *ast.FunctionLiteral:
		c.enterScope()
		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value, false)
		}
		err := c.Compile(node.Body)
		if err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()
		compiledFun := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}
		c.emit(code.OpConstant, c.addConstant(compiledFun))
	case *ast.FunctionStatement:
		c.enterScope()
		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value, false)
		}
		err := c.Compile(node.Body)
		if err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()
		compiledFun := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}
		c.emit(code.OpConstant, c.addConstant(compiledFun))
		symbol := c.symbolTable.Define(node.Name.Value, true)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobalImm, symbol.Index)
		} else {
			c.emit(code.OpSetLocalImm, symbol.Index)
		}
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}
		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpCall, len(node.Arguments))
	default:
		log.Fatalf("Failed to compile %T %+#v", node, node)
	}
	return nil
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
