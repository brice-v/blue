package compiler

import (
	"blue/ast"
	"blue/code"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		c.BlockNestLevel++
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
		c.clearBlockSymbols()
		// Emit an `OpJump` with a bogus value
		jumpPos := c.emit(code.OpJump, 9999)
		allEndingJumpPos = append(allEndingJumpPos, jumpPos)
		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)
	}
	if node.Alternative == nil {
		c.emit(code.OpNull)
	} else {
		c.BlockNestLevel++
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
		c.clearBlockSymbols()
	}
	afterAlternativePos := len(c.currentInstructions())
	for _, jumpPos := range allEndingJumpPos {
		c.changeOperand(jumpPos, afterAlternativePos)
	}
	return nil
}

func (c *Compiler) clearBlockSymbols() {
	if c.BlockNestLevel == -1 {
		return
	}
	if len(c.symbolTable.BlockSymbols) > c.BlockNestLevel {
		for _, sym := range c.symbolTable.BlockSymbols[c.BlockNestLevel] {
			delete(c.symbolTable.store, sym.Name)
		}
		if c.BlockNestLevel > 0 {
			c.symbolTable.BlockSymbols = c.symbolTable.BlockSymbols[:c.BlockNestLevel]
		}
	} else {
		clear(c.symbolTable.BlockSymbols)
	}
	c.BlockNestLevel--
}

func (c *Compiler) loadSymbol(s Symbol) {
	if s.Immutable {
		switch s.Scope {
		case GlobalScope:
			c.emit(code.OpGetGlobalImm, s.Index)
		case LocalScope:
			c.emit(code.OpGetLocalImm, s.Index)
		case BuiltinScope:
			c.emit(code.OpGetBuiltin, s.BuiltinModuleIndex, s.Index)
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
			c.emit(code.OpGetBuiltin, s.BuiltinModuleIndex, s.Index)
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
	// TODO: Look into why this is necessary (happens with for ident in iterable in imported file)
	var sym Symbol
	var ok bool
	sym, ok = c.symbolTable.Resolve(c.getName(ident.Value))
	if !ok {
		// Try without qualifier to resolve local variables from imported files
		sym, ok = c.symbolTable.Resolve(ident.Value)
		if !ok {
			return fmt.Errorf("identifier not found: %s", ident.Value)
		}
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
	sym, ok := c.symbolTable.Resolve(c.getName(rootIdent.Value))
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

func (c *Compiler) compileForStatement(node *ast.ForStatement) error {
	c.BlockNestLevel++
	c.forIndex++
	if node.UsesVar {
		err := c.Compile(node.Initializer)
		if err != nil {
			return err
		}
	} else {
		ok, sym, right := c.isIdentOnLeftInIterableOnRight(node.Condition)
		if ok {
			err := c.compileIdentInIterableFor(sym, node, right)
			if err != nil {
				return err
			}
		} else {
			ok, sym1, sym2, right := c.isListIdentsOnLeftInIterableOnRight(node.Condition)
			if ok {
				err := c.compileListIdentsInIterableFor(sym1, sym2, node, right)
				if err != nil {
					return err
				}
			}
		}
	}
	condPos := len(c.currentInstructions())
	err := c.Compile(node.Condition)
	if err != nil {
		return err
	}
	// Emit an `OpJumpNotTruthy` with a bogus value
	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)
	for _, setter := range node.IterableSetters {
		err := c.Compile(setter)
		if err != nil {
			return err
		}
	}
	err = c.Compile(node.Consequence)
	if err != nil {
		return err
	}
	postExpPos := -1
	if node.PostExp != nil {
		postExpPos = len(c.currentInstructions())
		err = c.Compile(node.PostExp)
		if err != nil {
			return err
		}
	}
	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
	}
	c.emit(code.OpJump, condPos)
	afterConsequencePos := len(c.currentInstructions())
	c.changeOperand(jumpNotTruthyPos, afterConsequencePos)
	if breakPoss, ok := c.breakPos[c.forIndex]; ok {
		for _, pos := range breakPoss {
			c.changeOperand(pos, afterConsequencePos)
		}
	}
	if contPoss, ok := c.contPos[c.forIndex]; ok {
		for _, pos := range contPoss {
			if node.PostExp != nil {
				c.changeOperand(pos, postExpPos)
			} else {
				c.changeOperand(pos, condPos)
			}
		}
	}
	delete(c.breakPos, c.forIndex)
	delete(c.contPos, c.forIndex)
	c.forIndex--
	c.clearBlockSymbols()
	return nil
}

func (c *Compiler) getName(name string) string {
	if c.importNestLevel == -1 {
		return name
	}
	return fmt.Sprintf("%s.%s", c.modName[c.importNestLevel], name)
}

func (c *Compiler) createFilePathFromImportPath(importPath string) string {
	var fpath bytes.Buffer
	if c.CompilerBasePath != "." {
		fpath.WriteString(c.CompilerBasePath)
		fpath.WriteString(string(os.PathSeparator))
	}
	importPath = strings.ReplaceAll(importPath, ".", string(os.PathSeparator))
	fpath.WriteString(importPath)
	fpath.WriteString(".b")
	return fpath.String()
}

func (c *Compiler) compileImportStatement(node *ast.ImportStatement) error {
	name := node.Path.Value
	if c.IsStd(name) {
		if node.Alias != nil {
			return fmt.Errorf("alias for std module not supported")
		}
		return c.CompileStdModule(name, node.IdentsToImport, node.ImportAll)
	}
	fpath := c.createFilePathFromImportPath(name)
	modName := strings.ReplaceAll(filepath.Base(fpath), ".b", "")
	var inputStr string
	if !object.IsEmbed {
		file, err := filepath.Abs(fpath)
		if err != nil {
			return fmt.Errorf("failed to import '%s'. Could not get absolute filepath", name)
		}
		ofile, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("failed to import '%s'. Could not open file '%s' for reading", name, file)
		}
		defer ofile.Close()
		fileData, err := io.ReadAll(ofile)
		if err != nil {
			return fmt.Errorf("failed to import '%s'. Could not read the file", name)
		}
		inputStr = string(fileData)
	} else {
		fileData, err := object.Files.ReadFile(consts.EMBED_FILES_PREFIX + fpath)
		if err != nil {
			return fmt.Errorf("failed to import '%s'. Could not read the file at path '%s'", name, fpath)
		}
		inputStr = string(fileData)
	}

	l := lexer.New(inputStr, fpath)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			splitMsg := strings.Split(msg, "\n")
			firstPart := fmt.Sprintf("%s%s\n", consts.PARSER_ERROR_PREFIX, splitMsg[0])
			consts.ErrorPrinter(firstPart)
			for i, s := range splitMsg {
				if i == 0 {
					continue
				}
				fmt.Println(s)
			}
		}
		return fmt.Errorf("%sFile '%s' contains Parser Errors", consts.PARSER_ERROR_PREFIX, name)
	}
	if node.ImportAll {
		// Import All acts as if everything is in the current file
		return c.Compile(program)
	}
	if len(node.IdentsToImport) >= 1 {
		for _, ident := range node.IdentsToImport {
			if strings.HasPrefix(ident.Value, "_") {
				return fmt.Errorf("imports must be public to import them. failed to import %s from %s", ident.Value, modName)
			}
		}
		// TODO: Handle only importing these? Or only making them accessible during compiling
		// TODO: Add test case trying to call method such as abc._hello() => this should ideally fail to compile
		// when called from the file importing abc
	}
	if node.Alias != nil {
		modName = node.Alias.Value
	}
	c.importNestLevel++
	c.modName = append(c.modName, modName)
	err := c.Compile(program)
	if err != nil {
		return err
	}
	c.modName = c.modName[:c.importNestLevel]
	c.importNestLevel--
	c.ValidModuleNames = append(c.ValidModuleNames, modName)
	// So the problem now is that index operator, needs to work based off available modules
	// while compiling, if we encounter a identifier that is a module, we must pull it in
	// TODO: Figure out help string later
	literal := &object.Module{Name: modName, Env: nil, HelpStr: ""}
	c.emit(code.OpConstant, c.addConstant(literal))
	symbol := c.symbolTable.Define(modName, true, c.BlockNestLevel)
	switch symbol.Scope {
	case GlobalScope:
		c.emit(code.OpSetGlobalImm, symbol.Index)
	case LocalScope:
		c.emit(code.OpSetLocalImm, symbol.Index)
	}
	return nil
}

func (c *Compiler) compileIndexExpression(node *ast.IndexExpression) error {
	leftIdent, leftIsIdent := node.Left.(*ast.Identifier)
	rightStr, rightIsStr := node.Index.(*ast.StringLiteral)
	if leftIsIdent && rightIsStr {
		// Check if left is a module and if together this can be resolved
		sym, ok := c.symbolTable.Resolve(fmt.Sprintf("%s.%s", leftIdent.Value, rightStr.Value))
		if ok {
			c.loadSymbol(sym)
			return nil
		}
	}
	// Support uniform function call syntax "".println()
	str, ok := node.Index.(*ast.StringLiteral)
	if ok {
		s, ok1 := c.symbolTable.Resolve(c.getName(str.Value))
		if ok1 {
			c.pushedArg = true
			c.loadSymbol(s)
		}
	}
	err := c.Compile(node.Left)
	if err != nil {
		return err
	}
	if !c.pushedArg {
		err = c.Compile(node.Index)
		if err != nil {
			return err
		}
		c.emit(code.OpIndex)
	}
	return nil
}

func (c *Compiler) compileCompLiteral(t, nonEvaluatedProgram string) error {
	symName := fmt.Sprintf("__internal__%d", c.listSetMapCompLiteralIndex)
	s := strings.ReplaceAll(nonEvaluatedProgram, "__internal__", symName)
	l := lexer.New(s, fmt.Sprintf("<internal: %s>", t))
	p := parser.New(l)
	rootNode := p.ParseProgram()
	if len(rootNode.Statements) < 1 {
		return fmt.Errorf("%s error:, not enough statements", t)
	}
	if len(p.Errors()) > 0 {
		return fmt.Errorf("%s error: %s", t, strings.Join(p.Errors(), " | "))
	}
	err := c.Compile(rootNode)
	if err != nil {
		return err
	}
	sym, ok := c.symbolTable.Resolve(symName)
	if !ok {
		return fmt.Errorf("this should never occur, failed to resolve: %s", symName)
	}
	c.loadSymbol(sym)
	return nil
}

func (c *Compiler) compileListCompLiteral(node *ast.ListCompLiteral) error {
	err := c.compileCompLiteral("ListCompLiteral", node.NonEvaluatedProgram)
	if err != nil {
		return err
	}
	c.emit(code.OpListCompLiteral)
	return nil
}

func (c *Compiler) compileSetCompLiteral(node *ast.SetCompLiteral) error {
	err := c.compileCompLiteral("SetCompLiteral", node.NonEvaluatedProgram)
	if err != nil {
		return err
	}
	c.emit(code.OpSetCompLiteral)
	return nil
}

func (c *Compiler) compileMapCompLiteral(node *ast.MapCompLiteral) error {
	err := c.compileCompLiteral("MapCompLiteral", node.NonEvaluatedProgram)
	if err != nil {
		return err
	}
	c.emit(code.OpMapCompLiteral)
	return nil
}

func (c *Compiler) compileMatchExpression(node *ast.MatchExpression) error {
	conditionLen := len(node.Conditions)
	consequenceLen := len(node.Consequences)
	if conditionLen != consequenceLen {
		return fmt.Errorf("conditions length is not equal to consequences length in match expression")
	}
	if node.OptionalValue != nil {
		return fmt.Errorf("handle compile of optional value: %s", node.OptionalValue.String())
	}
	allEndingJumpPos := []int{}
	for i := range node.Conditions {
		var err error
		var jumpNotTruthyPos int
		condIsDefault := node.Conditions[i].String() == "_"
		if !condIsDefault {
			err = c.Compile(node.Conditions[i])
			if err != nil {
				return err
			}
			// Emit an `OpJumpNotTruthy` with a bogus value
			jumpNotTruthyPos = c.emit(code.OpJumpNotTruthy, 9999)
		}
		c.BlockNestLevel++
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
		c.clearBlockSymbols()
		// Emit an `OpJump` with a bogus value
		jumpPos := c.emit(code.OpJump, 9999)
		allEndingJumpPos = append(allEndingJumpPos, jumpPos)
		afterConsequencePos := len(c.currentInstructions())
		if !condIsDefault {
			c.changeOperand(jumpNotTruthyPos, afterConsequencePos)
		}
	}
	afterAlternativePos := len(c.currentInstructions())
	for _, jumpPos := range allEndingJumpPos {
		c.changeOperand(jumpPos, afterAlternativePos)
	}
	return nil
}
