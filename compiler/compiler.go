package compiler

import (
	"blue/ast"
	"blue/code"
	"blue/lexer"
	"blue/object"
	"blue/token"
	"bytes"
	"fmt"
	"log"
	"regexp"
	"sort"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

const cCompilerBasePath = "."

type Compiler struct {
	constants     []object.Object
	constantFolds map[uint64]int

	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int

	ErrorTrace []string

	currentPos int
	Tokens     map[int][]token.Token

	BlockNestLevel int

	forIndex int
	breakPos map[int][]int
	contPos  map[int][]int

	importNestLevel  int
	modName          []string
	CompilerBasePath string
	ValidModuleNames []string // TODO: Maybe eventually use map[string]struct{}

	listSetMapCompLiteralIndex int

	coreCompiled bool

	inMatch bool
}

func (c *Compiler) DebugString() string {
	var out bytes.Buffer
	out.WriteString("Compiler{\n")
	fmt.Fprintf(&out, "\tconstanstLen: %d\n", len(c.constants))
	fmt.Fprintf(&out, "\tconstantFoldsLen: %d\n", len(c.constantFolds))
	fmt.Fprintf(&out, "\tsymbolTable: \n|\n%s\n|\n", c.symbolTable.String())
	fmt.Fprintf(&out, "\tscopes: %#+v\n", c.scopes)
	fmt.Fprintf(&out, "\tscopeIndex: %d\n", c.scopeIndex)
	fmt.Fprintf(&out, "\tErrorTrace: %#+v\n", c.ErrorTrace)
	fmt.Fprintf(&out, "\tcurrentPos: %d\n", c.currentPos)
	fmt.Fprintf(&out, "\tTokensLen: %d\n", len(c.Tokens))
	fmt.Fprintf(&out, "\tBlockNestLevel: %d\n", c.BlockNestLevel)
	fmt.Fprintf(&out, "\tforIndex: %d\n", c.forIndex)
	fmt.Fprintf(&out, "\tbreakPos: %#+v\n", c.breakPos)
	fmt.Fprintf(&out, "\tcontPos: %#+v\n", c.contPos)
	fmt.Fprintf(&out, "\timportNestLevel: %d\n", c.importNestLevel)
	fmt.Fprintf(&out, "\tmodName: %#+v\n", c.modName)
	fmt.Fprintf(&out, "\tCompilerBasePath: %q\n", c.CompilerBasePath)
	fmt.Fprintf(&out, "\tValidModulesNames: %#+v\n", c.ValidModuleNames)
	fmt.Fprintf(&out, "\tlistSetMapCompLiteralIndex: %d\n", c.listSetMapCompLiteralIndex)
	fmt.Fprintf(&out, "\tcoreCompiled: %t\n", c.coreCompiled)
	fmt.Fprintf(&out, "\tinMatch: %t\n", c.inMatch)
	out.WriteString("}")
	return out.String()
}

type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	pushedArg           bool
}

func New() *Compiler {
	symbolTable := NewSymbolTable()
	for i, v := range object.AllBuiltins[0].Builtins {
		symbolTable.DefineBuiltin(i, v.Name, 0)
	}
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		pushedArg:           false,
	}
	return &Compiler{
		constants:     object.OBJECT_CONSTANTS,
		constantFolds: map[uint64]int{},
		symbolTable:   symbolTable,
		scopes:        []CompilationScope{mainScope},
		scopeIndex:    0,
		ErrorTrace:    []string{},

		currentPos: 0,
		Tokens:     map[int][]token.Token{},

		BlockNestLevel: -1,
		forIndex:       0,
		breakPos:       map[int][]int{},
		contPos:        map[int][]int{},

		importNestLevel:  -1,
		modName:          []string{},
		CompilerBasePath: cCompilerBasePath,
	}
}

func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

func NewWithStateAndCore(s *SymbolTable, constants []object.Object) *Compiler {
	c := NewWithState(s, constants)
	c.compileCore()
	return c
}

func NewFromCore() *Compiler {
	return newFromCore()
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
	if index := object.IsConstantObject(obj); index != -1 {
		// return reserved index for constant object
		return index
	}
	if index := c.isConstantFolded(obj); index != -1 {
		return index
	}
	c.constants = append(c.constants, obj)
	index := len(c.constants) - 1
	c.addToConstantFolds(obj, index)
	return index
}

func (c *Compiler) addToConstantFolds(obj object.Object, index int) {
	ho := object.HashObject(obj)
	c.constantFolds[ho] = index
}

func (c *Compiler) isConstantFolded(obj object.Object) int {
	ho := object.HashObject(obj)
	if index, ok := c.constantFolds[ho]; ok {
		return index
	}
	return -1
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

func (c *Compiler) lastInstructionIsSet() bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	currentOp := c.scopes[c.scopeIndex].lastInstruction.Opcode
	return currentOp == code.OpSetGlobal || currentOp == code.OpSetGlobalImm || currentOp == code.OpSetLocal || currentOp == code.OpSetLocalImm
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
		pushedArg:           false,
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

const ignoreStr = `Filepath: "", LineNumber: 0, PositionInLine: 0
`

func (c *Compiler) PrintStackTrace() {
	prevS := ""
	for _, s := range c.ErrorTrace {
		if s != prevS && s != ignoreStr {
			fmt.Print(s)
		}
		prevS = s
	}
}

func existsInTokens(t token.Token, toks []token.Token) bool {
	for _, tok := range toks {
		if tok.LineNumber == t.LineNumber && tok.PositionInLine == t.PositionInLine {
			return true
		}
	}
	return false
}

func (c *Compiler) Compile(node ast.Node) error {
	if _, ok := node.(*ast.Program); !ok {
		t := node.TokenToken()
		if t.LineNumber != 0 && t.PositionInLine != 0 && t.Filepath != "" && !existsInTokens(t, c.Tokens[c.currentPos]) {
			c.Tokens[c.currentPos] = append(c.Tokens[c.currentPos], t)
		}
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
		var literal *object.Stringo
		if node.Value == object.USE_PARAM_STR {
			literal = object.USE_PARAM_STR_OBJ
		} else {
			literal = &object.Stringo{Value: node.Value}
		}
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
		// Note: this is needed for list comp literals to work properly
		// a similar thing is done in evaluator
		if !c.lastInstructionIs(code.OpListCompLiteral) {
			c.emit(code.OpList, len(node.Elements))
		}
	case *ast.SetLiteral:
		if !c.lastInstructionIs(code.OpSetCompLiteral) {
			for _, exp := range node.Elements {
				err := c.Compile(exp)
				if err != nil {
					return c.addNodeToErrorTrace(err, node.Token)
				}
			}
			c.emit(code.OpSet, len(node.Elements))
		}
	case *ast.MapLiteral:
		if !c.lastInstructionIs(code.OpMapCompLiteral) {
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
					_, ok1 := c.symbolTable.Resolve(c.getName(ident.Value))
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
		}
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
		// c.listCompilationContext.VariableName = c.getName(node.Names[0].Value)
		err := c.Compile(node.Value)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		// c.listCompilationContext = ListCompilationContext{
		// 	VariableName: "",
		// 	NestingLevel: -1,
		// 	Indices:      []int{},
		// }
		if node.IsListDestructor || node.IsMapDestructor {
			return fmt.Errorf("List/Map Destructor not yet supported, failed to compile %#+v", node)
		}
		if len(node.Names) > 1 {
			return fmt.Errorf("multiple identifiers to define, not supported yet %#+v", node.Names)
		}
		var symbol Symbol
		if fun, isFun := node.Value.(*ast.FunctionLiteral); isFun {
			symbol = c.symbolTable.DefineFun(c.getName(node.Names[0].Value), false, c.BlockNestLevel, fun.Parameters, fun.ParameterExpressions)
		} else {
			symbol = c.symbolTable.Define(c.getName(node.Names[0].Value), false, c.BlockNestLevel)
		}
		switch symbol.Scope {
		case GlobalScope:
			c.emit(code.OpSetGlobal, symbol.Index)
		case LocalScope:
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
		var symbol Symbol
		if fun, isFun := node.Value.(*ast.FunctionLiteral); isFun {
			symbol = c.symbolTable.DefineFun(c.getName(node.Names[0].Value), true, c.BlockNestLevel, fun.Parameters, fun.ParameterExpressions)
		} else {
			symbol = c.symbolTable.Define(c.getName(node.Names[0].Value), true, c.BlockNestLevel)
		}
		switch symbol.Scope {
		case GlobalScope:
			c.emit(code.OpSetGlobalImm, symbol.Index)
		case LocalScope:
			c.emit(code.OpSetLocalImm, symbol.Index)
		}
	case *ast.AssignmentExpression:
		err := c.compileAssignmentExpression(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.Identifier:
		if c.inMatch && node.Value == "_" {
			c.emit(code.OpMatchAny)
		} else {
			symbol, ok := c.symbolTable.Resolve(c.getName(node.Value))
			if !ok {
				// Due to the way compiling works, if its a builtin we need to try again
				symbol, ok = c.symbolTable.Resolve(node.Value)
				if !ok {
					return fmt.Errorf("identifier not found %s\n%s", node.Value, lexer.GetErrorLineMessage(node.Token))
				}
			}
			c.loadSymbol(symbol)
		}
	case *ast.IndexExpression:
		err := c.compileIndexExpression(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.FunctionLiteral:
		c.enterScope()
		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value, false, c.BlockNestLevel)
		}
		compiledFun := c.setupFunction(node.Parameters, node.ParameterExpressions, node.Body)
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
		compiledFun.Instructions = instructions
		compiledFun.NumLocals = numLocals
		funIndex := c.addConstant(compiledFun)
		c.emit(code.OpClosure, funIndex, len(freeSymbols))
	case *ast.FunctionStatement:
		symbol := c.symbolTable.Define(c.getName(node.Name.Value), true, c.BlockNestLevel)
		c.enterScope()
		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value, false, c.BlockNestLevel)
		}
		compiledFun := c.setupFunction(node.Parameters, node.ParameterExpressions, node.Body)
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
		compiledFun.Instructions = instructions
		compiledFun.NumLocals = numLocals
		funIndex := c.addConstant(compiledFun)
		c.emit(code.OpClosure, funIndex, len(freeSymbols))
		switch symbol.Scope {
		case GlobalScope:
			c.emit(code.OpSetGlobalImm, symbol.Index)
		case LocalScope:
			c.emit(code.OpSetLocalImm, symbol.Index)
		}
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.compileCallExpression(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.ForStatement:
		err := c.compileForStatement(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.BreakStatement:
		pos := c.emit(code.OpJump, 9999)
		if c.breakPos[c.forIndex] == nil {
			c.breakPos[c.forIndex] = []int{}
		}
		c.breakPos[c.forIndex] = append(c.breakPos[c.forIndex], pos)
	case *ast.ContinueStatement:
		pos := c.emit(code.OpJump, 9999)
		if c.contPos[c.forIndex] == nil {
			c.contPos[c.forIndex] = []int{}
		}
		c.contPos[c.forIndex] = append(c.contPos[c.forIndex], pos)
	case *ast.TryCatchStatement:
		c.currentPos = len(c.currentInstructions())
		c.BlockNestLevel++
		c.emit(code.OpTry)
		err := c.Compile(node.TryBlock)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.TryBlock.Token)
		}
		c.clearBlockSymbols()
		if node.CatchBlock != nil {
			c.currentPos = len(c.currentInstructions())
			c.BlockNestLevel++
			c.emit(code.OpCatch)
			symbol := c.symbolTable.Define(node.CatchIdentifier.Value, true, c.BlockNestLevel)
			switch symbol.Scope {
			case GlobalScope:
				c.emit(code.OpSetGlobalImm, symbol.Index)
			case LocalScope:
				c.emit(code.OpSetLocalImm, symbol.Index)
			}
			err := c.Compile(node.CatchIdentifier)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.CatchBlock.Token)
			}
			err = c.Compile(node.CatchBlock)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.CatchBlock.Token)
			}
			c.emit(code.OpCatchEnd)
			c.clearBlockSymbols()
		}
		if node.FinallyBlock != nil {
			c.currentPos = len(c.currentInstructions())
			c.BlockNestLevel++
			c.emit(code.OpFinally)
			err := c.Compile(node.FinallyBlock)
			if err != nil {
				return c.addNodeToErrorTrace(err, node.FinallyBlock.Token)
			}
			c.clearBlockSymbols()
			c.emit(code.OpFinallyEnd)
		}
	case *ast.ImportStatement:
		err := c.compileImportStatement(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.ListCompLiteral:
		err := c.compileListCompLiteral(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.SetCompLiteral:
		err := c.compileSetCompLiteral(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.MapCompLiteral:
		err := c.compileMapCompLiteral(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.MatchExpression:
		err := c.compileMatchExpression(node)
		if err != nil {
			return c.addNodeToErrorTrace(err, node.Token)
		}
	case *ast.EvalExpression:
		literal := &object.Stringo{Value: node.StrToEval.String()}
		c.emit(code.OpConstant, c.addConstant(literal))
		c.emit(code.OpEval)
	case *ast.ExecStringLiteral:
		if node.Value == "" {
			err := fmt.Errorf("exec string must not be empty")
			return c.addNodeToErrorTrace(err, node.Token)
		}
		literal := &object.ExecString{Value: node.Value}
		c.emit(code.OpExecString, c.addConstant(literal))
	default:
		log.Fatalf("Failed to compile %T %+#v", node, node)
	}
	return nil
}

func (c *Compiler) setIsPushedArg(a bool) {
	c.scopes[c.scopeIndex].pushedArg = a
}

func (c *Compiler) isPushedArg() bool {
	return c.scopes[c.scopeIndex].pushedArg
}
