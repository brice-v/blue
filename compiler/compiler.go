package compiler

import (
	"blue/ast"
	"blue/code"
	"blue/lexer"
	"blue/object"
	"blue/token"
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

	ErrorTrace []string

	pushedArg bool

	currentPos int
	Tokens     map[int][]token.Token
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

func New() *Compiler {
	symbolTable := NewSymbolTable()
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
		ErrorTrace:  []string{},
		pushedArg:   false,
		currentPos:  0,
		Tokens:      map[int][]token.Token{},
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
	c.removeLastInstruction()
}

func (c *Compiler) removeLastInstruction() {
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

func (c *Compiler) addNodeToErrorTrace(err error, tok token.Token) error {
	c.ErrorTrace = append(c.ErrorTrace, fmt.Sprintf("%s\n", lexer.GetErrorLineMessage(tok)))
	return err
}

func (c *Compiler) PrintStackTrace() {
	prevS := ""
	for _, s := range c.ErrorTrace {
		if s != prevS {
			fmt.Print(s)
		}
		prevS = s
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	if _, ok := node.(*ast.Program); !ok {
		c.Tokens[c.currentPos] = append(c.Tokens[c.currentPos], node.TokenToken())
	}
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			c.currentPos = len(c.currentInstructions())
			err := c.Compile(s)
			if err != nil {
				return err
			}
			clear(c.ErrorTrace)
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		c.emit(code.OpPop)
	case *ast.InfixExpression:
		if node.Operator == "<" || node.Operator == "<=" {
			err := c.Compile(node.Right)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.Token)
			}
			err = c.Compile(node.Left)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.Token)
			}
			c.compileInfixExpression(node.Operator)
			return nil
		}
		err := c.Compile(node.Left)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		err = c.Compile(node.Right)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		err = c.compileInfixExpression(node.Operator)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
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
			return c.addNodeToErrorTrace(err, node.Token)
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
			return fmt.Errorf("failed to create regex literal %q", node.TokenLiteral())
		}
		literal := &object.Regex{Value: r}
		c.emit(code.OpConstant, c.addConstant(literal))
	case *ast.StringLiteral:
		literal := &object.Stringo{Value: node.Value}
		origStrIndex := c.addConstant(literal)
		c.emit(code.OpConstant, origStrIndex)
		if len(node.InterpolationValues) != 0 {
			for i, interp := range node.InterpolationValues {
				err := c.Compile(interp)
				if err != nil {
					return c.addNodeToErrorTrace(err, node.Token)
				}
				s := node.OriginalInterpolationString[i]
				c.emit(code.OpConstant, c.addConstant(&object.Stringo{Value: s}))
			}
			c.emit(code.OpStringInterp, origStrIndex, len(node.InterpolationValues)*2)
		}
	case *ast.ListLiteral:
		for _, exp := range node.Elements {
			err := c.Compile(exp)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.Token)
			}
		}
		c.emit(code.OpList, len(node.Elements))
	case *ast.SetLiteral:
		for _, exp := range node.Elements {
			err := c.Compile(exp)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.Token)
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
			keyNode1 := keyNode
			// Support keys in map without requiring quotes
			ident, ok := keyNode.(*ast.Identifier)
			if ok {
				_, ok1 := c.symbolTable.Resolve(ident.Value)
				if !ok1 {
					keyNode1 = &ast.StringLiteral{Value: ident.Value}
				}
			}
			err := c.Compile(keyNode1)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.Token)
			}
			valueNode := node.Pairs[keyNode]
			err = c.Compile(valueNode)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.Token)
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
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.Token)
			}
		}
	case *ast.VarStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
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
			return c.addNodeToErrorTrace(err, node.Token)
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
	case *ast.AssignmentExpression:
		err := c.compileAssignmentExpression(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("identifier not found %s\n%s", node.Value, lexer.GetErrorLineMessage(node.Token))
		}
		c.loadSymbol(symbol)
	case *ast.IndexExpression:
		// Support uniform function call syntax "".println()
		str, ok := node.Index.(*ast.StringLiteral)
		if ok {
			s, ok1 := c.symbolTable.Resolve(str.Value)
			if ok1 {
				c.pushedArg = true
				c.loadSymbol(s)
			}
		}
		err := c.Compile(node.Left)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		if !c.pushedArg {
			err = c.Compile(node.Index)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.Token)
			}
			c.emit(code.OpIndex)
		}
	case *ast.FunctionLiteral:
		c.enterScope()
		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value, false)
		}
		err := c.Compile(node.Body)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}
		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()
		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}
		compiledFun := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}
		funIndex := c.addConstant(compiledFun)
		c.emit(code.OpClosure, funIndex, len(freeSymbols))
	case *ast.FunctionStatement:
		symbol := c.symbolTable.Define(node.Name.Value, true)
		c.enterScope()
		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value, false)
		}
		err := c.Compile(node.Body)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}
		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()
		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}
		compiledFun := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}
		funIndex := c.addConstant(compiledFun)
		c.emit(code.OpClosure, funIndex, len(freeSymbols))
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobalImm, symbol.Index)
		} else {
			c.emit(code.OpSetLocalImm, symbol.Index)
		}
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.Token)
			}
		}
		// If we updated arg based on ufcs need to increment argument len
		argLen := len(node.Arguments)
		if c.pushedArg {
			argLen++
			c.pushedArg = false
		}
		c.emit(code.OpCall, argLen)
	default:
		log.Fatalf("Failed to compile %T %+#v", node, node)
	}
	return nil
}
