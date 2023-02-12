package parser

import (
	"blue/ast"
	"blue/lexer"
	"blue/token"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	_ int = iota
	// LOWEST precedence
	LOWEST
	// COMPOUND_ASSIGNMENT is the precedence for all compund assignment expressions
	COMPOUND_ASSIGNMENT
	// OR_P is the logical or precedence
	OR_P
	// AND_P is the logical and precedence
	AND_P
	// BITWISE_OR is the bitwise or precedence
	BITWISE_OR
	// BITWISE_XOR is the bitwise xor precedence
	BITWISE_XOR
	// BITWISE_ADD is the bitwise and precedence
	BITWISE_ADD
	// EQUALS is the == and != precedence
	EQUALS
	// LESSGREATER < > <= >=
	LESSGREATER
	// BITWISE_SHIFTS is the << and >> precedence
	BITWISE_SHIFTS
	// SUM is the plus and minus precedence
	SUM
	// PRODUCT is the precedence of * / // %
	PRODUCT
	// POW_P is the exponentiation precedence
	POW_P
	// IN_P is the `in` keyword precedence
	IN_P
	// RANGE_P is the precedece for ranges
	RANGE_P
	// PREFIX is the precedence of prefix expressions ie. -x  not x
	PREFIX
	// CALL myFun(x)
	CALL
	// INDEX is for . and [ member access of objects
	INDEX
)

var precedences = map[token.Type]int{
	token.EQ:          EQUALS,
	token.NEQ:         EQUALS,
	token.LT:          LESSGREATER,
	token.GT:          LESSGREATER,
	token.LTEQ:        LESSGREATER,
	token.GTEQ:        LESSGREATER,
	token.AND:         AND_P,
	token.OR:          OR_P,
	token.PLUS:        SUM,
	token.MINUS:       SUM,
	token.FSLASH:      PRODUCT,
	token.STAR:        PRODUCT,
	token.FDIV:        PRODUCT,
	token.HAT:         BITWISE_XOR,
	token.AMPERSAND:   BITWISE_ADD,
	token.PIPE:        BITWISE_OR,
	token.PERCENT:     PRODUCT,
	token.TILDE:       PREFIX,
	token.LSHIFT:      BITWISE_SHIFTS,
	token.RSHIFT:      BITWISE_SHIFTS,
	token.POW:         POW_P,
	token.ASSIGN:      COMPOUND_ASSIGNMENT,
	token.PLUSEQ:      COMPOUND_ASSIGNMENT,
	token.MINUSEQ:     COMPOUND_ASSIGNMENT,
	token.DIVEQ:       COMPOUND_ASSIGNMENT,
	token.MULEQ:       COMPOUND_ASSIGNMENT,
	token.POWEQ:       COMPOUND_ASSIGNMENT,
	token.FDIVEQ:      COMPOUND_ASSIGNMENT,
	token.ANDEQ:       COMPOUND_ASSIGNMENT,
	token.OREQ:        COMPOUND_ASSIGNMENT,
	token.BINNOTEQ:    COMPOUND_ASSIGNMENT,
	token.LSHIFTEQ:    COMPOUND_ASSIGNMENT,
	token.RSHIFTEQ:    COMPOUND_ASSIGNMENT,
	token.PERCENTEQ:   COMPOUND_ASSIGNMENT,
	token.XOREQ:       COMPOUND_ASSIGNMENT,
	token.RANGE:       RANGE_P,
	token.NONINCRANGE: RANGE_P,
	token.IN:          IN_P,
	token.NOTIN:       IN_P,
	token.LPAREN:      CALL,
	token.LBRACKET:    INDEX,
	token.DOT:         INDEX,
}

// Parser is the struct containing information relevant to parsing
type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []string

	prefixParseFuns map[token.Type]prefixParseFun
	infixParseFuns  map[token.Type]infixParseFun
	// postfixParseFuns map[token.Type]postfixParseFun
}

// helper functions at bottom
type (
	prefixParseFun func() ast.Expression
	infixParseFun  func(ast.Expression) ast.Expression
)

// New takes a lexer and returns a Parser object
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	p.prefixParseFuns = make(map[token.Type]prefixParseFun)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.HEX, p.parseHexLiteral)
	p.registerPrefix(token.OCTAL, p.parseOctalLiteral)
	p.registerPrefix(token.BINARY, p.parseBinaryLiteral)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TILDE, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseParenGroupExpresion)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.PIPE, p.parseLambdaLiteral)
	p.registerPrefix(token.STRING_DOUBLE_QUOTE, p.parseStringLiteral)
	p.registerPrefix(token.STRING_SINGLE_QUOTE, p.parseStringLiteral)
	p.registerPrefix(token.RAW_STRING, p.parseRawStringLiteral)
	p.registerPrefix(token.BACKTICK, p.parseExecStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseListLiteral)
	p.registerPrefix(token.LBRACE, p.parseMapOrSetLiteral)
	p.registerPrefix(token.FOR, p.parseForExpression)
	p.registerPrefix(token.MATCH, p.parseMatchExpression)
	p.registerPrefix(token.NULL_KW, p.parseNullKeyword)
	p.registerPrefix(token.EVAL, p.parseEvalExpression)
	p.registerPrefix(token.SPAWN, p.parseSpawnExpression)
	p.registerPrefix(token.SELF, p.parseSelfExpression)
	p.registerPrefix(token.LSHIFT, p.parsePrefixExpression)
	p.infixParseFuns = make(map[token.Type]infixParseFun)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.FSLASH, p.parseInfixExpression)
	p.registerInfix(token.FDIV, p.parseInfixExpression)
	p.registerInfix(token.STAR, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTEQ, p.parseInfixExpression)
	p.registerInfix(token.GTEQ, p.parseInfixExpression)
	p.registerInfix(token.HAT, p.parseInfixExpression)
	p.registerInfix(token.AMPERSAND, p.parseInfixExpression)
	p.registerInfix(token.PIPE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.PERCENT, p.parseInfixExpression)
	p.registerInfix(token.RANGE, p.parseInfixExpression)
	p.registerInfix(token.NONINCRANGE, p.parseInfixExpression)
	p.registerInfix(token.IN, p.parseInfixExpression)
	p.registerInfix(token.NOTIN, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMemberAccessExpression)
	p.registerInfix(token.POW, p.parseInfixExpression)
	p.registerInfix(token.LSHIFT, p.parseInfixExpression)
	p.registerInfix(token.RSHIFT, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.PLUSEQ, p.parseAssignmentExpression)
	p.registerInfix(token.MINUSEQ, p.parseAssignmentExpression)
	p.registerInfix(token.MULEQ, p.parseAssignmentExpression)
	p.registerInfix(token.DIVEQ, p.parseAssignmentExpression)
	p.registerInfix(token.FDIVEQ, p.parseAssignmentExpression)
	p.registerInfix(token.POWEQ, p.parseAssignmentExpression)
	p.registerInfix(token.ANDEQ, p.parseAssignmentExpression)
	p.registerInfix(token.OREQ, p.parseAssignmentExpression)
	p.registerInfix(token.BINNOTEQ, p.parseAssignmentExpression)
	p.registerInfix(token.PERCENTEQ, p.parseAssignmentExpression)
	p.registerInfix(token.LSHIFTEQ, p.parseAssignmentExpression)
	p.registerInfix(token.RSHIFTEQ, p.parseAssignmentExpression)
	p.registerInfix(token.XOREQ, p.parseAssignmentExpression)

	// Read two tokens to give values to curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns a list of all the parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError is a peekToken error and will append the error
// to the list of parser errors
func (p *Parser) peekError(t token.Type) {
	errorLine := lexer.GetErrorLineMessage(p.peekToken)
	msg := fmt.Sprintf("expected next token to be %s, got %s instead\n%s", t, p.peekToken.Type, errorLine)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFunError will append an error if no prefix parse function is found
func (p *Parser) noPrefixParseFunError(t token.Type) {
	errorLine := lexer.GetErrorLineMessage(p.curToken)
	msg := fmt.Sprintf("no prefix parse function for %s found\n%s", t, errorLine)
	p.errors = append(p.errors, msg)
}

// nextToken is a helper function to advance the tokens
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	if p.curToken.Type == token.HASH || p.curToken.Type == token.MULTLINE_COMMENT {
		for p.curToken.Type == token.HASH || p.curToken.Type == token.MULTLINE_COMMENT {
			p.curToken = p.l.NextToken()
		}
	}
	p.peekToken = p.l.NextToken()
	if p.peekToken.Type == token.HASH || p.peekToken.Type == token.MULTLINE_COMMENT {
		for p.peekToken.Type == token.HASH || p.peekToken.Type == token.MULTLINE_COMMENT {
			p.peekToken = p.l.NextToken()
		}
	}
}

// ParseProgram parses the program and returns a program as an ast
func (p *Parser) ParseProgram() *ast.Program {
	// initializing empty program with empty statement structs
	// aka constructing the root node
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	program.HelpStrTokens = []string{}

	// First validate that there are no illegal tokens
	// We make a copy as to not disrupt the state
	lcpy := *p.l
	var tok token.Token
	for {
		tok = lcpy.NextToken()
		if tok.Type == token.ILLEGAL {
			errorLine := lexer.GetErrorLineMessage(tok)
			msg := fmt.Sprintf("%s token encountered. got=%q\n%s", tok.Type, tok.Literal, errorLine)
			p.errors = append(p.errors, msg)
			return nil
		}
		if tok.Type == token.EOF {
			break
		}
	}

	for !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.HASH) || p.curTokenIs(token.MULTLINE_COMMENT) {
			p.nextToken()
			if p.curTokenIs(token.EOF) {
				break
			}
		} else if p.curTokenIs(token.DOCSTRING_COMMENT) {
			program.HelpStrTokens = append(program.HelpStrTokens, strings.TrimLeft(p.curToken.Literal, " \t"))
			p.nextToken()
		} else {
			stmt := p.parseStatement()
			if stmt != nil {
				program.Statements = append(program.Statements, stmt)
			}
			p.nextToken()
		}
	}

	return program
}

// parseStatement will parse any potential statement nodes and
// return a statement node otherwise nil
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.VAR:
		return p.parseVarStatement()
	case token.VAL:
		return p.parseValStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.IMPORT:
		return p.parseImportStatement()
	case token.TRY:
		return p.parseTryCatchBlock()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	default:
		// This is how im handling a function statement becuase otherwise all function literals
		// will get confused and not be able to parse (due to the "fun" prefixed token)
		if p.curToken.Type == token.FUNCTION && p.peekTokenIs(token.IDENT) {
			return p.parseFunctionLiteralStatement()
		}
		return p.parseExpressionStatement()
	}
}

// parseVarStatement will try to parse a var statement and if
// successful will return the constructed ast node otherwise nil
func (p *Parser) parseVarStatement() *ast.VarStatement {
	// initialize the var statement with the current var token
	stmt := &ast.VarStatement{Token: p.curToken}

	// return nil if the next token is not an identifier token
	if !p.expectPeekIs(token.IDENT) {
		return nil
	}

	// set the statement name to the identifier with the current token being
	// token.IDENT and the value being the actual string of the identifier
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.peekTokenIsAssignmentToken() {
		p.peekError(token.ASSIGN)
		return nil
	}
	if p.peekTokenIsAssignmentToken() {
		// nextToken used to be in the peekTokenIsAssignmentToken method which meant it would assign this to nothing
		p.nextToken()
		stmt.AssignmentToken = p.curToken
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseValStatement will try to parse a val statement and if
// successful will return the constructed ast node otherwise nil
func (p *Parser) parseValStatement() *ast.ValStatement {
	// initialize the val statement with the current val token
	stmt := &ast.ValStatement{Token: p.curToken}

	// return nil if the next token is not an identifier token
	if !p.expectPeekIs(token.IDENT) {
		return nil
	}

	// set the statement name to the identifier with the current token being
	// token.IDENT and the value being the actual string of the identifier
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeekIs(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	// initialize a return statement node with the current token.RETURN
	stmt := &ast.ReturnStatement{Token: p.curToken}

	// skip over the token.RETURN
	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFunctionLiteralStatement() *ast.FunctionStatement {
	lit := &ast.FunctionStatement{Token: p.curToken}

	if !p.expectPeekIs(token.IDENT) {
		return nil
	}

	lit.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeekIs(token.LPAREN) {
		return nil
	}

	lit.Parameters, lit.ParameterExpressions = p.parseFunctionParameters()

	if !p.expectPeekIs(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// Expressions

// parseExpression will see if their is an associated parsing function
// with a token and return the parsed ast.Expression
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFuns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFunError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	// TODO: I think if we want mandatory semicolons this is where wed put it
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFuns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

// parseExpressionStatement will create an ast.ExpressionStatement
// and return it with the properly fully parsed Expression
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseIdentifier will return the identifier expression node
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral will return the integer literal ast node
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}
	tokenLiteral := strings.Replace(p.curToken.Literal, "_", "", -1)
	value, err := strconv.ParseInt(tokenLiteral, 0, 64)
	if err != nil {
		bigInt := new(big.Int)
		bigValue, ok := bigInt.SetString(tokenLiteral, 10)
		if ok {
			bigLit := &ast.BigIntegerLiteral{Token: p.curToken}
			bigLit.Value = bigValue
			return bigLit
		}
		errorLine := lexer.GetErrorLineMessage(p.curToken)
		msg := fmt.Sprintf("could not parse %q as an integer\n%s", p.curToken.Literal, errorLine)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// parseFloatLiteral will return the float literal ast node
func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}
	tokenLiteral := strings.Replace(p.curToken.Literal, "_", "", -1)
	value, err := strconv.ParseFloat(tokenLiteral, 64)
	if err != nil || len(tokenLiteral) > len(fmt.Sprintf("%f", value)) {
		bigValue, err := decimal.NewFromString(tokenLiteral)
		if err == nil {
			bigLit := &ast.BigFloatLiteral{Token: p.curToken}
			bigLit.Value = bigValue
			return bigLit
		}
		errorLine := lexer.GetErrorLineMessage(p.curToken)
		msg := fmt.Sprintf("could not parse %q as a float\n%s", p.curToken.Literal, errorLine)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// parseHexLiteral will return the Hex literal ast node
func (p *Parser) parseHexLiteral() ast.Expression {
	lit := &ast.HexLiteral{Token: p.curToken}
	tokenLiteral := strings.Replace(p.curToken.Literal, "_", "", -1)
	tokenLiteral = strings.Replace(tokenLiteral, "0x", "", -1)
	value, err := strconv.ParseUint(tokenLiteral, 16, 64)
	if err != nil {
		errorLine := lexer.GetErrorLineMessage(p.curToken)
		msg := fmt.Sprintf("could not parse %q as an unsigned integer\n%s", p.curToken.Literal, errorLine)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// parseOctalLiteral will return the Octal literal ast node
func (p *Parser) parseOctalLiteral() ast.Expression {
	lit := &ast.OctalLiteral{Token: p.curToken}
	tokenLiteral := strings.Replace(p.curToken.Literal, "_", "", -1)
	tokenLiteral = strings.Replace(tokenLiteral, "0o", "", -1)
	value, err := strconv.ParseUint(tokenLiteral, 8, 64)
	if err != nil {
		errorLine := lexer.GetErrorLineMessage(p.curToken)
		msg := fmt.Sprintf("could not parse %q as an unsigned integer\n%s", p.curToken.Literal, errorLine)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// parseBinaryLiteral will return the Binary literal ast node
func (p *Parser) parseBinaryLiteral() ast.Expression {
	lit := &ast.BinaryLiteral{Token: p.curToken}
	tokenLiteral := strings.Replace(p.curToken.Literal, "_", "", -1)
	tokenLiteral = strings.Replace(tokenLiteral, "0b", "", -1)
	value, err := strconv.ParseUint(tokenLiteral, 2, 64)
	if err != nil {
		errorLine := lexer.GetErrorLineMessage(p.curToken)
		msg := fmt.Sprintf("could not parse %q as an unsigned integer\n%s", p.curToken.Literal, errorLine)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// parsePrefixExpression parses the prefix expression and returns the ast node
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	// once we assign the prefix token we skip over it
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

// parseInfixExpression parses the infix expression and returns the ast node
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	if p.curTokenIs(token.RSHIFT) && p.peekTokenIs(token.SEMICOLON) {
		expression := &ast.PostfixExpression{
			Operator: p.curToken.Literal,
			Token:    p.curToken,
			Left:     left,
		}
		if !p.expectPeekIs(token.SEMICOLON) {
			return nil
		}
		return expression
	}
	p.nextToken()

	expression.Right = p.parseExpression(precedence)
	return expression
}

// parseBoolean returns a boolean ast node (which is an expression) with the
// value coming from testing if it is true
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNullKeyword() ast.Expression {
	return &ast.Null{}
}

func (p *Parser) parseEvalExpression() ast.Expression {
	ee := &ast.EvalExpression{
		Token: p.curToken,
	}
	if !p.expectPeekIs(token.LPAREN) {
		return nil
	}
	strToEvalExpression := p.parseExpression(LOWEST)
	ee.StrToEval = strToEvalExpression
	if !p.curTokenIs(token.RPAREN) {
		errorLine := lexer.GetErrorLineMessage(p.curToken)
		msg := fmt.Sprintf("token after EvalExpression is not ), got %s instead\n%s", p.curToken.Literal, errorLine)
		p.errors = append(p.errors, msg)
		return nil
	}
	if !p.expectPeekIs(token.SEMICOLON) {
		return nil
	}
	return ee
}

func (p *Parser) parseSpawnExpression() ast.Expression {
	se := &ast.SpawnExpression{
		Token: p.curToken,
	}
	p.nextToken()
	se.Arguments, _ = p.parseExpressionList(token.RPAREN)
	return se
}

func (p *Parser) parseSelfExpression() ast.Expression {
	se := &ast.SelfExpression{
		Token: p.curToken,
	}
	if !p.expectPeekIs(token.LPAREN) {
		return nil
	}
	if !p.expectPeekIs(token.RPAREN) {
		return nil
	}
	return se
}

func (p *Parser) parseImportStatement() ast.Statement {
	stmt := &ast.ImportStatement{
		Token: p.curToken,
	}

	// return nil if the next token is not an import token
	if !p.expectPeekIs(token.IMPORT_PATH) {
		return nil
	}

	// set the statement name to the identifier with the current token being
	// token.IDENT and the value being the actual string of the identifier
	stmt.Path = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	// p.nextToken()
	return stmt
}

func (p *Parser) parseBreakStatement() ast.Statement {
	bks := &ast.BreakStatement{
		Token: p.curToken,
	}
	if !p.expectPeekIs(token.SEMICOLON) {
		return nil
	}
	return bks
}

func (p *Parser) parseContinueStatement() ast.Statement {
	cs := &ast.ContinueStatement{
		Token: p.curToken,
	}
	if !p.expectPeekIs(token.SEMICOLON) {
		return nil
	}
	return cs
}

// parseParenGroupExpression parses a parenthesis grouped expression
func (p *Parser) parseParenGroupExpresion() ast.Expression {
	// skip the paren
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeekIs(token.RPAREN) {
		return nil
	}

	return exp
}

// parseIfExpression parses an if expression
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken, Conditions: []ast.Expression{}, Consequences: []*ast.BlockStatement{}}
	for {
		if !p.expectPeekIs(token.LPAREN) {
			return nil
		}
		// parse the (group) expression as the condition
		expression.Conditions = append(expression.Conditions, p.parseExpression(LOWEST))
		if !p.expectPeekIs(token.LBRACE) {
			return nil
		}
		expression.Consequences = append(expression.Consequences, p.parseBlockStatement())
		if p.peekTokenIs(token.ELSE) {
			// if token == ELSE skip over it and parse the other block
			p.nextToken()

			if p.peekTokenIs(token.IF) {
				p.nextToken()
				continue
			}
			if !p.expectPeekIs(token.LBRACE) {
				return nil
			}
			expression.Alternative = p.parseBlockStatement()
			break
		} else {
			break
		}
	}

	return expression
}

// parseBlockStatement parses a block statement and returns a block statement ast node
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	block.HelpStrTokens = []string{}

	// skip over the LBRACE
	p.nextToken()

	for p.curTokenIs(token.DOCSTRING_COMMENT) {
		block.HelpStrTokens = append(block.HelpStrTokens, strings.TrimLeft(p.curToken.Literal, " \t"))
		p.nextToken()
	}

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		// consume tokens and parse statements as necessary
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeekIs(token.LPAREN) {
		return nil
	}

	lit.Parameters, lit.ParameterExpressions = p.parseFunctionParameters()

	if !p.expectPeekIs(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionParameters parses function parameters
func (p *Parser) parseFunctionParameters() ([]*ast.Identifier, []ast.Expression) {
	identifiers := []*ast.Identifier{}
	defaultParameters := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers, defaultParameters
	}

	p.nextToken()
	val := p.parseExpression(LOWEST)
	switch val.(type) {
	case *ast.AssignmentExpression:
		assignedExpression := val.(*ast.AssignmentExpression)
		ident := assignedExpression.Left.(*ast.Identifier)
		identifiers = append(identifiers, ident)
		defaultParameters = append(defaultParameters, assignedExpression.Value)
	case *ast.Identifier:
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
		defaultParameters = append(defaultParameters, nil)
	default:
		errorLine := lexer.GetErrorLineMessage(p.curToken)
		msg := fmt.Sprintf("expected assignment expression or identifier. got=%T\n%s", val, errorLine)
		p.errors = append(p.errors, msg)
		return nil, nil
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		val := p.parseExpression(LOWEST)
		switch val.(type) {
		case *ast.AssignmentExpression:
			assignedExpression := val.(*ast.AssignmentExpression)
			ident := assignedExpression.Left.(*ast.Identifier)
			identifiers = append(identifiers, ident)
			defaultParameters = append(defaultParameters, assignedExpression.Value)
		case *ast.Identifier:
			ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			identifiers = append(identifiers, ident)
			defaultParameters = append(defaultParameters, nil)
		default:
			errorLine := lexer.GetErrorLineMessage(p.curToken)
			msg := fmt.Sprintf("expected assignment expression or identifier. got=%T\n%s", val, errorLine)
			p.errors = append(p.errors, msg)
			return nil, nil
		}
	}

	if !p.expectPeekIs(token.RPAREN) {
		return nil, nil
	}

	return identifiers, defaultParameters
}

// parseLambdaLiteral will parse a lambda expression and return the ast node
func (p *Parser) parseLambdaLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	lit.Parameters = p.parseLambdaParameters()

	if !p.expectPeekIs(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// parseLambdaParameters will parse the parameters of a lambda expression and return the identitiers
func (p *Parser) parseLambdaParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.PIPE) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeekIs(token.PIPE) {
		return nil
	}
	if !p.expectPeekIs(token.RARROW) {
		return nil
	}

	return identifiers
}

// parseCallExpression will parse the call expression and return the ast node
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments, exp.DefaultArguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExecStringLiteral() ast.Expression {
	return &ast.ExecStringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// parseRawStringLiteral is just like parse string however it doesnt allow
// for any string interpolation or escape sequences
func (p *Parser) parseRawStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// parseStringLiteral will parse the string and return its ast node
func (p *Parser) parseStringLiteral() ast.Expression {
	exp := &ast.StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	exp.InterpolationValues, exp.OriginalInterpolationString = p.parseStringInterpolationValues(p.curToken.Literal)
	return exp
}

// parseListLiteral parses a list literal and returns the ast node
func (p *Parser) parseListLiteral() ast.Expression {
	elems, _ := p.parseExpressionList(token.RBRACKET)
	exp := &ast.ListLiteral{
		Token:    p.curToken,
		Elements: elems,
	}
	return exp
}

// parseSetLiteral tries to parse and return a Set Literal ast node
func (p *Parser) parseSetLiteral(firstTok token.Token, firstExp ast.Expression) ast.Expression {
	exp := &ast.SetLiteral{Token: firstTok}
	exp.Elements = []ast.Expression{firstExp}
	if !p.expectPeekIs(token.COMMA) {
		return nil
	}
	// Skip the comma
	p.nextToken()

	for {
		// get into the next exp
		value := p.parseExpression(LOWEST)

		exp.Elements = append(exp.Elements, value)

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeekIs(token.COMMA) {
			return nil
		}
		p.nextToken()
		if p.peekTokenIs(token.RBRACE) && p.curTokenIs(token.RBRACE) {
			// This seems very edge casey but we want to break if the end of a set is an object that ends with a }
			break
		}
		if p.curTokenIs(token.RBRACE) {
			break
		}
	}
	return exp
}

// parseMapLiteral parses a map and returns an ast expression node
func (p *Parser) parseMapOrSetLiteral() ast.Expression {
	firstTok := p.curToken
	exp := &ast.MapLiteral{Token: firstTok}
	exp.Pairs = make(map[ast.Expression]ast.Expression)
	exp.PairsIndex = make(map[int]ast.Expression)

	i := 0
	for !p.peekTokenIs(token.RBRACE) {
		// get into the map
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if p.peekTokenIs(token.COMMA) {
			return p.parseSetLiteral(firstTok, key)
		}
		if p.peekTokenIs(token.FOR) {
			return p.parseSetComprehension(firstTok, key)
		}
		if !p.expectPeekIs(token.COLON) {
			return nil
		}
		// get into the next exp
		p.nextToken()
		value := p.parseExpression(LOWEST)
		if p.peekTokenIs(token.FOR) {
			return p.parseMapComprehension(firstTok, key, value)
		}

		exp.Pairs[key] = value
		exp.PairsIndex[i] = key
		i++

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeekIs(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeekIs(token.RBRACE) {
		return nil
	}

	return exp
}

// parseIndexExpression parses the index expression and returns an ast expression node
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	indxExp := &ast.IndexExpression{Token: p.curToken, Left: left}
	// skip over the [
	p.nextToken()
	indxExp.Index = p.parseExpression(LOWEST)

	if !p.expectPeekIs(token.RBRACKET) {
		return nil
	}

	return indxExp
}

// parseMemberAccessExpression parses a dot token to use as an index expression
func (p *Parser) parseMemberAccessExpression(left ast.Expression) ast.Expression {
	dotTok := p.curToken
	// first item needs to be a identifier
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		// create a string literal to use as a lookup for member access
		indx := &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
		// Token doesnt matter for this one but we need the rest
		indxExp := &ast.IndexExpression{Token: dotTok, Left: left, Index: indx}
		return indxExp
	} else if p.peekTokenIs(token.INT) {
		p.nextToken()
		i, err := strconv.Atoi(p.curToken.Literal)
		if err != nil {
			return nil
		}
		indx := &ast.IntegerLiteral{Token: p.curToken, Value: int64(i)}
		indxExp := &ast.IndexExpression{Token: dotTok, Left: left, Index: indx}
		return indxExp
	} else {
		p.peekError(token.INT)
		return nil
	}
}

// parseForExpression parses a for expression and returns the for expressions
// ast node
func (p *Parser) parseForExpression() ast.Expression {
	exp := &ast.ForExpression{
		Token: p.curToken,
	}
	// current token is for, expect next to be lparen
	if !p.expectPeekIs(token.LPAREN) {
		return nil
	}
	// move to the expression in the condition
	p.nextToken()

	exp.Condition = p.parseExpression(LOWEST)

	if !p.expectPeekIs(token.RPAREN) {
		return nil
	}
	if !p.expectPeekIs(token.LBRACE) {
		return nil
	}
	exp.Consequence = p.parseBlockStatement()
	return exp
}

// parseAssignmentExpression will return a parsed assignment as an Expression ast node
func (p *Parser) parseAssignmentExpression(exp ast.Expression) ast.Expression {
	switch node := exp.(type) {
	case *ast.Identifier, *ast.IndexExpression:
	default:
		errorLine := lexer.GetErrorLineMessage(p.curToken)
		msg := fmt.Sprintf("expected identifier or index expression on left but got %T %#v\n%s", node, exp, errorLine)
		p.errors = append(p.errors, msg)
		return nil
	}

	ae := &ast.AssignmentExpression{Token: p.curToken, Left: exp}

	p.nextToken()

	ae.Value = p.parseExpression(LOWEST)

	return ae
}

func (p *Parser) parseMatchExpression() ast.Expression {
	me := &ast.MatchExpression{Token: p.curToken,
		Conditions:   []ast.Expression{},
		Consequences: []*ast.BlockStatement{},
	}
	if !p.peekTokenIs(token.LBRACE) {
		// Skip over the `match` statement to the value to bind
		p.nextToken()
		me.OptionalValue = p.parseExpression(LOWEST)
	}
	if !p.expectPeekIs(token.LBRACE) {
		return nil
	}
	p.nextToken()
	for {
		me.Conditions = append(me.Conditions, p.parseExpression(LOWEST))
		if !p.expectPeekIs(token.RARROW) {
			return nil
		}
		p.nextToken()

		me.Consequences = append(me.Consequences, p.parseBlockStatement())
		if !p.expectPeekIs(token.COMMA) {
			return nil
		}
		p.nextToken()
		if p.curTokenIs(token.RBRACE) {
			break
		}
	}

	// Skip over the } on the way out of parsing the
	p.nextToken()

	return me
}

func (p *Parser) parseTryCatchBlock() *ast.TryCatchStatement {
	t := p.curToken
	p.nextToken()
	tryBlock := p.parseBlockStatement()
	var catchIdent *ast.Identifier
	var catchBlock *ast.BlockStatement
	var finallyBlock *ast.BlockStatement
	if p.peekTokenIs(token.CATCH) {
		p.nextToken()
		if !p.expectPeekIs(token.LPAREN) {
			return nil
		}
		if !p.expectPeekIs(token.IDENT) {
			return nil
		}
		catchIdent = p.parseIdentifier().(*ast.Identifier)
		if !p.expectPeekIs(token.RPAREN) {
			return nil
		}
		p.nextToken()
		catchBlock = p.parseBlockStatement()
	} else {
		// Cant use 'expectPeekIs' because it consumes the token (which is done below)
		if !p.peekTokenIs(token.FINALLY) {
			return nil
		}
	}
	if p.peekTokenIs(token.FINALLY) {
		p.nextToken() // Skip }
		p.nextToken() // Skip finally
		finallyBlock = p.parseBlockStatement()
	}
	return &ast.TryCatchStatement{
		Token:           t,
		TryBlock:        tryBlock,
		CatchIdentifier: catchIdent,
		CatchBlock:      catchBlock,
		FinallyBlock:    finallyBlock,
	}
}

// Helper functions

// parseExpressionList takes an end token and returns the slice
// of expressions that make up the list
func (p *Parser) parseExpressionList(end token.Type) ([]ast.Expression, map[string]ast.Expression) {
	list := []ast.Expression{}
	defaultArgs := make(map[string]ast.Expression)

	if p.peekTokenIs(end) {
		p.nextToken()
		return list, defaultArgs
	}

	p.nextToken()
	val := p.parseExpression(LOWEST)
	assignmentExpression, ok := val.(*ast.AssignmentExpression)
	if ok {
		identString := assignmentExpression.Left.String()
		defaultArgs[identString] = assignmentExpression.Value
	} else {
		if p.peekTokenIs(token.FOR) {
			return p.parseListComprehension(val), nil
		}
		list = append(list, val)
	}

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		val := p.parseExpression(LOWEST)
		assignmentExpression, ok := val.(*ast.AssignmentExpression)
		if ok {
			identString := assignmentExpression.Left.String()
			defaultArgs[identString] = assignmentExpression.Value
		} else {
			list = append(list, val)
		}
	}

	if !p.expectPeekIs(end) {
		return nil, nil
	}

	return list, defaultArgs
}

func (p *Parser) parseListComprehension(valueToBind ast.Expression) []ast.Expression {
	// Skip over the for
	p.nextToken()
	// current token is for, expect next to be lparen
	if !p.expectPeekIs(token.LPAREN) {
		return nil
	}
	// move to the expression in the condition
	p.nextToken()

	expCond := p.parseExpression(LOWEST)
	if !p.expectPeekIs(token.RPAREN) {
		return nil
	}

	var ifCond ast.Expression
	if p.peekTokenIs(token.IF) {
		// skip RPAREN
		p.nextToken()
		// skip IF
		p.nextToken()
		ifCond = p.parseExpression(LOWEST)
	}

	var program string
	if ifCond != nil {
		program = fmt.Sprintf("var __internal__ = []; for %s { if %s { var __result__ = %s; __internal__ = append(__internal__, __result__); } };", expCond, ifCond, valueToBind.String())
	} else {
		program = fmt.Sprintf("var __internal__ = []; for %s { var __result__ = %s; __internal__ = append(__internal__, __result__); };", expCond, valueToBind.String())
	}
	if !p.expectPeekIs(token.RBRACKET) {
		return nil
	}
	return []ast.Expression{&ast.ListCompLiteral{NonEvaluatedProgram: program}}
}

func (p *Parser) parseMapComprehension(tok token.Token, key, value ast.Expression) ast.Expression {
	// Skip over the for
	p.nextToken()
	// current token is for, expect next to be lparen
	if !p.expectPeekIs(token.LPAREN) {
		return nil
	}
	// move to the expression in the condition
	p.nextToken()

	expCond := p.parseExpression(LOWEST)
	if !p.expectPeekIs(token.RPAREN) {
		return nil
	}

	var ifCond ast.Expression
	if p.peekTokenIs(token.IF) {
		// skip RPAREN
		p.nextToken()
		// skip IF
		p.nextToken()
		ifCond = p.parseExpression(LOWEST)
	}

	var program string
	if ifCond != nil {
		program = fmt.Sprintf("var __internal__ = {}; for %s { if %s { __internal__[%s] = %s } };", expCond, ifCond, key.String(), value.String())
	} else {
		program = fmt.Sprintf("var __internal__ = {}; for %s { __internal__[%s] = %s  };", expCond, key.String(), value.String())
	}
	if !p.expectPeekIs(token.RBRACE) {
		return nil
	}

	return &ast.MapCompLiteral{Token: tok, NonEvaluatedProgram: program}
}

func (p *Parser) parseSetComprehension(tok token.Token, value ast.Expression) ast.Expression {
	// Skip over the for
	p.nextToken()
	// current token is for, expect next to be lparen
	if !p.expectPeekIs(token.LPAREN) {
		return nil
	}
	// move to the expression in the condition
	p.nextToken()

	expCond := p.parseExpression(LOWEST)
	if !p.expectPeekIs(token.RPAREN) {
		return nil
	}

	var ifCond ast.Expression
	if p.peekTokenIs(token.IF) {
		// skip RPAREN
		p.nextToken()
		// skip IF
		p.nextToken()
		ifCond = p.parseExpression(LOWEST)
	}

	var program string
	if ifCond != nil {
		program = fmt.Sprintf("var __internal__ = []; for %s { if %s { __internal__ = append(__internal__, %s) } }; __internal__ = set(__internal__);", expCond, ifCond, value.String())
	} else {
		program = fmt.Sprintf("var __internal__ = []; for %s { __internal__ = append(__internal__, %s) }; __internal__ = set(__internal__);", expCond, value.String())
	}
	if !p.expectPeekIs(token.RBRACE) {
		return nil
	}

	return &ast.SetCompLiteral{Token: tok, NonEvaluatedProgram: program}
}

// stringLexer is used to parse string interpolation values
type stringLexer struct {
	input        string
	position     int  // current pos. in input (points to current char)
	readPosition int  // current reading pos. in input (after current char)
	ch           byte // current char under examination
}

// newStringLexer takes an input and returns a *stringLexer
func newStringLexer(input string) *stringLexer {
	sl := &stringLexer{input: input}
	sl.readStringChar()
	return sl
}

// readStringChar reads the next char in the input and advances
// the read position
func (sl *stringLexer) readStringChar() {
	if sl.readPosition >= len(sl.input) {
		sl.ch = 0
	} else {
		sl.ch = sl.input[sl.readPosition]
	}
	sl.position = sl.readPosition
	sl.readPosition++
}

// peekChar checks the next readPosition char and returns the byte
// without advancing the position
func (sl *stringLexer) peekChar() byte {
	if sl.readPosition >= len(sl.input) {
		return 0
	}
	return sl.input[sl.readPosition]
}

// parseStringInterpolationValues helps with parsing string interpolation values
// it makes a new lexer for itself just to quickly parse the internal string
// interpolation expressions
func (p *Parser) parseStringInterpolationValues(value string) ([]ast.Expression, []string) {
	interps := []ast.Expression{}
	origStrings := []string{}

	sl := newStringLexer(value)
	for sl.ch != 0 {
		toLex := &strings.Builder{}
		if sl.ch == '#' && sl.peekChar() == '{' {
			sl.readStringChar()
			sl.readStringChar()
			for {
				if sl.ch == '}' {
					break
				}
				toLex.WriteByte(sl.ch)
				sl.readStringChar()
			}
		}
		if sl.ch == '}' {
			l := lexer.New(toLex.String(), "<internal: StringInterpolation>")
			lcpy := *l
			var tok token.Token
			for {
				tok = lcpy.NextToken()
				if tok.Type == token.ILLEGAL {
					errorLine := lexer.GetErrorLineMessage(tok)
					msg := fmt.Sprintf("%s token encountered. got=%q\n%s", tok.Type, tok.Literal, errorLine)
					p.errors = append(p.errors, msg)
				}
				if tok.Type == token.EOF {
					break
				}
			}
			origStrings = append(origStrings, fmt.Sprintf("#{%s}", toLex.String()))
			parseString := New(l)
			interps = append(interps, parseString.parseExpression(LOWEST))
		} else if sl.peekChar() == 0 {
			break
		}
		sl.readStringChar()
	}
	return interps, origStrings
}

// curTokenIs will check if the given token type matches the
// parsers current token's type
func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

// peekTokenIs will check if the given token type matches the
// parsers current token's type
func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

// expectPeekIs will check if the given token type matches the
// next token and if so will advance the tokens and return true
func (p *Parser) expectPeekIs(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	// create a peek error
	p.peekError(t)
	return false
}

func (p *Parser) peekTokenIsAssignmentToken() bool {
	if p.peekTokenIs(token.ASSIGN) ||
		p.peekTokenIs(token.PLUSEQ) ||
		p.peekTokenIs(token.MINUSEQ) ||
		p.peekTokenIs(token.DIVEQ) ||
		p.peekTokenIs(token.FDIVEQ) ||
		p.peekTokenIs(token.MULEQ) ||
		p.peekTokenIs(token.POWEQ) ||
		p.peekTokenIs(token.ANDEQ) ||
		p.peekTokenIs(token.OREQ) ||
		p.peekTokenIs(token.BINNOTEQ) ||
		p.peekTokenIs(token.LSHIFTEQ) ||
		p.peekTokenIs(token.RSHIFTEQ) ||
		p.peekTokenIs(token.PERCENTEQ) ||
		p.peekTokenIs(token.XOREQ) {
		return true
	}
	return false
}

// registerPrefix associates a token with a prefix parsing function
func (p *Parser) registerPrefix(tokenType token.Type, fun prefixParseFun) {
	p.prefixParseFuns[tokenType] = fun
}

// registerInfix associates a token with an infix parsing function
func (p *Parser) registerInfix(tokenType token.Type, fun infixParseFun) {
	p.infixParseFuns[tokenType] = fun
}

// Not using so commenting out
// // registerPostfix associates a token with a postfix parsing function
// func (p *Parser) registerPostfix(tokenType token.Type, fun postfixParseFun) {
// 	p.postfixParseFuns[tokenType] = fun
// }

// peekPrecedence is a helper function to return the precedence if it exists
// on the peek token
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence is a helper function to return the precedence if it exists
// on the current token
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
