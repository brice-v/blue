package parser

import (
	"blue/ast"
	"blue/consts"
	"blue/lexer"
	"blue/token"
	"bytes"
	"fmt"
	"io"
	"math"
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
	token.ANDANDEQ:    COMPOUND_ASSIGNMENT,
	token.OROREQ:      COMPOUND_ASSIGNMENT,
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

type parserError struct {
	Message        string
	FileLineColumn string
	PointerPos     string
	SourceLine     string
	LineNumber     int
	Hints          []string
}

func parseErrorString(errStr string, lineNumber int) parserError {
	lines := strings.SplitN(errStr, "\n", 3)
	err := parserError{}
	if len(lines) >= 1 {
		err.Message = lines[0]
	}
	if len(lines) >= 2 {
		posToSplit := strings.Index(lines[1], " ")
		err.FileLineColumn = lines[1][:posToSplit]
		err.SourceLine = lines[1][posToSplit+1:]
		if len(lines) >= 3 {
			pointerPos := strings.Index(lines[2], "^")
			if pointerPos != -1 {
				err.PointerPos = lines[2][posToSplit+1 : pointerPos+1]
			} else {
				err.PointerPos = "^"
			}
		}
	}
	err.LineNumber = lineNumber + 1

	return err
}

func (p *Parser) JoinedErrors() string {
	var out bytes.Buffer
	for _, err := range p.errors {
		fmt.Fprintf(&out, "|%#+v|", err)
	}
	return out.String()
}

// hintPattern is a pattern-to-hints mapping for error suggestions.
// Patterns are matched using strings.Contains against the error message.
type hintPattern struct {
	Pattern string
	Hints   []string
}

// parserHints is the ordered list of hint patterns. Earlier entries
// take priority when multiple patterns match the same message.
var parserHints = []hintPattern{
	{
		"expected = got for",
		[]string{"Did you mean to use '=' for assignment? e.g. val x = 42"},
	},
	{
		"expected = got if",
		[]string{"Did you mean to use '=' for assignment? e.g. val x = 42"},
	},
	{
		"expected = got while",
		[]string{"Did you mean to use '=' for assignment? e.g. val x = 42"},
	},
	{
		"unexpected }",
		[]string{"Unmatched closing brace — check for a missing '{' earlier"},
	},
	{
		"expected : got }",
		[]string{"Expected ':' after the key in a map or struct literal"},
	},
	{
		"expected : got for",
		[]string{"Expected ':' after the key in a map or struct literal"},
	},
	{
		"expected : got in",
		[]string{"Expected ':' after the key in a map or struct literal"},
	},
	{
		"expected : got an",
		[]string{"Expected ':' after the key in a map or struct literal"},
	},
	{
		"expected : got a",
		[]string{"Expected ':' after the key in a map or struct literal"},
	},
	{
		"expected ) got }",
		[]string{"Missing closing parenthesis"},
	},
	{
		"expected ) got for",
		[]string{"Missing closing parenthesis"},
	},
	{
		"expected ) got while",
		[]string{"Missing closing parenthesis"},
	},
	{
		"expected , got }",
		[]string{"Missing comma between elements"},
	},
	{
		"expected ; here got }",
		[]string{"Missing semicolon — did you forget to separate the loop parts?"},
	},
	{
		"expected ; here got for",
		[]string{"Missing semicolon — did you forget to separate the loop parts?"},
	},
	{
		"expected { got }",
		[]string{"Missing opening brace — check for a missing '{' earlier"},
	},
	{
		"invalid type for destructuring",
		[]string{"Destructuring expects identifiers or string keys, e.g. {key: value}"},
	},
	{
		"struct literal keys must be unique",
		[]string{"Each key in a struct literal must be unique"},
	},
	{
		"unexpected while",
		[]string{"blue does not have a 'while' keyword — use 'for' instead"},
	},
	{
		"unexpected do",
		[]string{"blue does not have a 'do' keyword — use 'for' instead"},
	},
}

// lookupHints checks the error message against known patterns and
// returns matching hints. Returns nil if no hints apply.
func lookupHints(message string) []string {
	for _, hp := range parserHints {
		if strings.Contains(message, hp.Pattern) {
			return hp.Hints
		}
	}
	return nil
}

func (p *Parser) ErrorMessages() []string {
	result := make([]string, len(p.errors))
	for i, err := range p.errors {
		result[i] = err.Message
	}
	return result
}

func (p *Parser) PrintParserErrors(out io.Writer) {
	// First pass: count duplicates keyed by (Message, FileLineColumn)
	// and track the index of the last occurrence of each key.
	dedupMap := make(map[string]int)       // key -> count
	lastOccurrence := make(map[string]int) // key -> index of last occurrence
	for i, err := range p.errors {
		key := err.Message + "|" + err.FileLineColumn
		dedupMap[key]++
		lastOccurrence[key] = i
	}

	for i, err := range p.errors {
		key := err.Message + "|" + err.FileLineColumn
		count := dedupMap[key]
		isLast := i == lastOccurrence[key]

		consts.ErrorPrinter("%s%s\n", consts.PARSER_ERROR_PREFIX, err.Message)
		if err.FileLineColumn != "" {
			fmt.Fprintf(out, "   %s\n", err.FileLineColumn)
		}
		if err.SourceLine != "" {
			fmt.Fprintf(out, "    %d │ %s\n", err.LineNumber, err.SourceLine)
			// Compute the line number prefix width so the pointer line
			// aligns with the source content (not the line number)
			lineNumWidth := len(fmt.Sprintf("%d", err.LineNumber))
			prefix := "    " + strings.Repeat(" ", lineNumWidth) + " │ "
			fmt.Fprintf(out, "%s%s", prefix, err.PointerPos)
			if err.Message != "" {
				fmt.Fprintf(out, " %s", err.Message)
			}
			fmt.Fprintln(out)
		}

		if len(err.Hints) > 0 {
			for _, hint := range err.Hints {
				fmt.Fprintf(out, "  [HINT] %s\n", hint)
			}
		}

		if count > 1 && isLast {
			fmt.Fprintf(out, "  [%d more similar error(s) omitted]\n", count-1)
		}

		fmt.Fprintln(out)
	}
}

// Parser is the struct containing information relevant to parsing
type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []parserError

	prefixParseFuns map[token.Type]prefixParseFun
	infixParseFuns  map[token.Type]infixParseFun
	// postfixParseFuns map[token.Type]postfixParseFun

	// StopAfterFirstError causes the parser to stop immediately after
	// the first error is encountered, preventing cascade errors.
	StopAfterFirstError bool
	// stopParsing is set internally when StopAfterFirstError triggers.
	stopParsing bool
}

// helper functions at bottom
type (
	prefixParseFun func() ast.Expression
	infixParseFun  func(ast.Expression) ast.Expression
)

// New takes a lexer and returns a Parser object
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []parserError{}}

	p.prefixParseFuns = make(map[token.Type]prefixParseFun)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.HEX, p.parseHexLiteral)
	p.registerPrefix(token.OCTAL, p.parseOctalLiteral)
	p.registerPrefix(token.BINARY, p.parseBinaryLiteral)
	p.registerPrefix(token.UINT, p.parseUIntegerLiteral)
	p.registerPrefix(token.BIGINT, p.parseBigIntegerLiteral)
	p.registerPrefix(token.BIGFLOAT, p.parseBigFloatLiteral)
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
	p.registerPrefix(token.ATLBRACE, p.parseStructLiteral)
	p.registerPrefix(token.MATCH, p.parseMatchExpression)
	p.registerPrefix(token.NULL_KW, p.parseNullKeyword)
	p.registerPrefix(token.EVAL, p.parseEvalExpression)
	p.registerPrefix(token.SPAWN, p.parseSpawnExpression)
	p.registerPrefix(token.DEFER, p.parseDeferExpression)
	p.registerPrefix(token.SELF, p.parseSelfExpression)
	p.registerPrefix(token.LSHIFT, p.parsePrefixExpression)
	p.registerPrefix(token.REGEX, p.parseRegexLiteral)
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
	p.registerInfix(token.ANDANDEQ, p.parseAssignmentExpression)
	p.registerInfix(token.OROREQ, p.parseAssignmentExpression)
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

// NewWithStopAfterFirst creates a Parser that stops immediately after
// the first error is encountered, preventing cascade errors.
func NewWithStopAfterFirst(l *lexer.Lexer) *Parser {
	p := New(l)
	p.StopAfterFirstError = true
	return p
}

// Errors returns a list of all the parser errors
func (p *Parser) HasErrors() bool {
	return len(p.errors) > 0
}

// error is a unified error method that appends a formatted error
// message to the parser's error list. It uses UserFriendlyName() for
// token types so error messages display human-readable names like '}'
// instead of raw token constants like RBRACE.
//
// The errorLine parameter determines which token's context to include:
//   - "peek" uses p.peekToken (for "expected next token" messages)
//   - "cur" uses p.curToken (for "got" messages)
//   - "tok" uses the provided token
//
// If StopAfterFirstError is true, this also sets stopParsing to signal
// that parsing should halt after the current call returns.
//
// Hints are automatically attached based on the error message content.
func (p *Parser) error(msg string, tokenContext token.Token) {
	errorLine := lexer.GetErrorLineMessage(tokenContext)
	fullMsg := msg + "\n" + errorLine
	pe := parseErrorString(fullMsg, tokenContext.LineNumber)
	// Attach contextual hints based on the error message
	pe.Hints = lookupHints(msg)
	p.errors = append(p.errors, pe)
	if p.StopAfterFirstError {
		p.stopParsing = true
	}
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
			p.error(fmt.Sprintf("%s token encountered. got=%q", tok.Type, tok.Literal), tok)
		}
		if tok.Type == token.EOF {
			break
		}
	}

	for !p.curTokenIs(token.EOF) {
		if p.stopParsing {
			break
		}
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
			if p.curTokenIs(token.SEMICOLON) {
				p.nextToken()
			}
		}
	}

	return program
}

// parseStatement will parse any potential statement nodes and
// return a statement node otherwise nil
func (p *Parser) parseStatement() ast.Statement {
	if p.stopParsing {
		return nil
	}
	switch p.curToken.Type {
	case token.VAR:
		return p.parseVarStatement()
	case token.VAL:
		return p.parseValStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.FROM:
		return p.parseFromStatement()
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

func (p *Parser) parseDestructorIdents(isMapDestructor, isListDestructor bool) ([]*ast.Identifier, map[ast.Expression]*ast.Identifier, bool) {
	names := []*ast.Identifier{}
	kvNames := make(map[ast.Expression]*ast.Identifier)
	if isMapDestructor {
		for !p.curTokenIs(token.RBRACE) {
			p.nextToken()
			exp := p.parseExpression(LOWEST)
			ident, isExpIdent := exp.(*ast.Identifier)
			if !p.peekTokenIs(token.COLON) && isExpIdent {
				p.nextToken()
				names = append(names, ident)
			} else if p.peekTokenIs(token.COMMA) {
				continue
			} else if p.peekTokenIs(token.COLON) {
				keyExp := p.parseExpression(LOWEST)
				// Check KeyExp is string or ident
				_, ok1 := keyExp.(*ast.Identifier)
				_, ok2 := keyExp.(*ast.StringLiteral)
				if !ok1 && !ok2 {
					p.error(fmt.Sprintf("invalid type for destructuring, expected identifier or string got %T instead", keyExp), p.peekToken)
					return nil, nil, true
				}
				// Skip key exp
				p.nextToken()
				// skip colon
				p.nextToken()
				valueExp := p.parseExpression(LOWEST)
				// skip valueExp
				p.nextToken()
				ident, isIdent := valueExp.(*ast.Identifier)
				if isIdent {
					kvNames[keyExp] = ident
				} else {
					p.error(fmt.Sprintf("invalid type for destructuring, expected identifier or string got %T instead", keyExp), p.peekToken)
					return nil, nil, true
				}
			}
			if p.peekTokenIs(token.RBRACE) {
				break
			}
		}
	} else if isListDestructor {
		p.nextToken()
		for !p.curTokenIs(token.RBRACKET) {
			names = append(names, p.parseIdentifier().(*ast.Identifier))
			if p.peekTokenIs(token.RBRACKET) {
				p.nextToken()
				continue
			}
			if !p.expectPeekIs(token.COMMA) {
				return nil, nil, true
			}
			p.nextToken()
		}
	} else {
		p.error("destructuring needs to be used with [ or { as first token", p.peekToken)
		return nil, nil, true
	}
	return names, kvNames, false
}

// parseVarStatement will try to parse a var statement and if
// successful will return the constructed ast node otherwise nil
func (p *Parser) parseVarStatement() *ast.VarStatement {
	// initialize the var statement with the current var token
	stmt := &ast.VarStatement{Token: p.curToken}

	if p.peekTokenIs(token.LBRACE) || p.peekTokenIs(token.LBRACKET) {
		// This will be supported for parsing destructuring
		p.nextToken()
		stmt.IsListDestructor = p.curTokenIs(token.LBRACKET)
		stmt.IsMapDestructor = p.curTokenIs(token.LBRACE)
		stmtIdents, stmtKVIdents, isErr := p.parseDestructorIdents(stmt.IsMapDestructor, stmt.IsListDestructor)
		if isErr {
			return nil
		}
		stmt.Names = stmtIdents
		stmt.KeyValueNames = stmtKVIdents
	} else if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		stmt.Names = []*ast.Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
	}

	if stmt.IsListDestructor || stmt.IsMapDestructor {
		if !p.expectPeekIs(token.ASSIGN) {
			return nil
		}
	} else {
		if !p.peekTokenIsAssignmentToken() {
			p.error(fmt.Sprintf("expected '%s' got %s instead", token.ASSIGN.UserFriendlyName(), p.peekToken.Type.UserFriendlyName()), p.peekToken)
			return nil
		}
		if p.peekTokenIsAssignmentToken() {
			// nextToken used to be in the peekTokenIsAssignmentToken method which meant it would assign this to nothing
			p.nextToken()
			stmt.AssignmentToken = p.curToken
		}
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

	if p.peekTokenIs(token.LBRACE) || p.peekTokenIs(token.LBRACKET) {
		// This will be supported for parsing destructuring
		p.nextToken()
		stmt.IsListDestructor = p.curTokenIs(token.LBRACKET)
		stmt.IsMapDestructor = p.curTokenIs(token.LBRACE)
		stmtIdents, stmtKVIdents, isErr := p.parseDestructorIdents(stmt.IsMapDestructor, stmt.IsListDestructor)
		if isErr {
			return nil
		}
		stmt.Names = stmtIdents
		stmt.KeyValueNames = stmtKVIdents
	} else if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		stmt.Names = []*ast.Identifier{{Token: p.curToken, Value: p.curToken.Literal}}
	}

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
		p.error(fmt.Sprintf("unexpected %s", p.curToken.Literal), p.curToken)
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
	tokenLiteral := strings.ReplaceAll(p.curToken.Literal, "_", "")
	value, err := strconv.ParseInt(tokenLiteral, 0, 64)
	if err != nil {
		bigInt := new(big.Int)
		bigValue, ok := bigInt.SetString(tokenLiteral, 10)
		if ok {
			bigLit := &ast.BigIntegerLiteral{Token: p.curToken}
			bigLit.Value = bigValue
			return bigLit
		}
		p.error(fmt.Sprintf("could not parse %q as %s", p.curToken.Literal, token.INT.UserFriendlyName()), p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

// ExactFloat64 uses decimal package to convert both numbers simultaneously to see if they
// are equal when rounded by default
func ExactFloat64(s string) (float64, bool, *decimal.Decimal) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false, nil
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return f, true, nil
	}
	fromPotentiallyRoundedFloat := decimal.NewFromFloat(f)
	fromString, err := decimal.NewFromString(s)
	if err != nil {
		return 0, false, nil
	}

	if fromPotentiallyRoundedFloat.Equal(fromString) {
		return f, true, nil
	}
	return 0, false, &fromString
}

// parseFloatLiteral will return the float literal ast node
func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}
	tokenLiteral := strings.ReplaceAll(p.curToken.Literal, "_", "")
	exactF, ok, maybeDecimal := ExactFloat64(tokenLiteral)
	if ok {
		lit.Value = exactF
		return lit
	}
	if maybeDecimal != nil {
		bigLit := &ast.BigFloatLiteral{Token: p.curToken}
		bigLit.Value = *maybeDecimal
		return bigLit
	}
	p.error(fmt.Sprintf("could not parse %q as %s", p.curToken.Literal, token.FLOAT.UserFriendlyName()), p.curToken)
	return nil
}

// parseHexLiteral will return the Hex literal ast node
func (p *Parser) parseHexLiteral() ast.Expression {
	lit := &ast.HexLiteral{Token: p.curToken}
	tokenLiteral := strings.ReplaceAll(p.curToken.Literal, "_", "")
	tokenLiteral = strings.ReplaceAll(tokenLiteral, "0x", "")
	value, err := strconv.ParseUint(tokenLiteral, 16, 64)
	if err != nil {
		p.error(fmt.Sprintf("could not parse %q as %s", p.curToken.Literal, token.HEX.UserFriendlyName()), p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

// parseOctalLiteral will return the Octal literal ast node
func (p *Parser) parseOctalLiteral() ast.Expression {
	lit := &ast.OctalLiteral{Token: p.curToken}
	tokenLiteral := strings.ReplaceAll(p.curToken.Literal, "_", "")
	tokenLiteral = strings.ReplaceAll(tokenLiteral, "0o", "")
	value, err := strconv.ParseUint(tokenLiteral, 8, 64)
	if err != nil {
		p.error(fmt.Sprintf("could not parse %q as %s", p.curToken.Literal, token.OCTAL.UserFriendlyName()), p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

// parseBinaryLiteral will return the Binary literal ast node
func (p *Parser) parseBinaryLiteral() ast.Expression {
	lit := &ast.BinaryLiteral{Token: p.curToken}
	tokenLiteral := strings.ReplaceAll(p.curToken.Literal, "_", "")
	tokenLiteral = strings.ReplaceAll(tokenLiteral, "0b", "")
	value, err := strconv.ParseUint(tokenLiteral, 2, 64)
	if err != nil {
		p.error(fmt.Sprintf("could not parse %q as %s", p.curToken.Literal, token.BINARY.UserFriendlyName()), p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

// parseUIntegerLiteral will return the UInteger literal ast node
func (p *Parser) parseUIntegerLiteral() ast.Expression {
	lit := &ast.UIntegerLiteral{Token: p.curToken}
	tokenLiteral := strings.ReplaceAll(p.curToken.Literal, "_", "")
	tokenLiteral = strings.ReplaceAll(tokenLiteral, "0u", "")
	value, err := strconv.ParseUint(tokenLiteral, 10, 64)
	if err != nil {
		p.error(fmt.Sprintf("could not parse %q as %s", p.curToken.Literal, token.UINT.UserFriendlyName()), p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

// parseBigIntegerLiteral will return the BigInteger literal ast node
func (p *Parser) parseBigIntegerLiteral() ast.Expression {
	lit := &ast.BigIntegerLiteral{Token: p.curToken}
	tokenLiteral := strings.ReplaceAll(p.curToken.Literal, "_", "")
	tokenLiteral = strings.ReplaceAll(tokenLiteral, "n", "")
	bi := new(big.Int)
	value, ok := bi.SetString(tokenLiteral, 10)
	if !ok {
		p.error(fmt.Sprintf("could not parse %q as %s", p.curToken.Literal, token.BIGINT.UserFriendlyName()), p.curToken)
		return nil
	}
	lit.Value = value
	return lit
}

// parseBigFloatLiteral will return the BigFloat literal ast node
func (p *Parser) parseBigFloatLiteral() ast.Expression {
	lit := &ast.BigFloatLiteral{Token: p.curToken}
	tokenLiteral := strings.ReplaceAll(p.curToken.Literal, "_", "")
	tokenLiteral = strings.ReplaceAll(tokenLiteral, "n", "")
	value, err := decimal.NewFromString(tokenLiteral)
	if err != nil {
		p.error(fmt.Sprintf("could not parse %q as %s", p.curToken.Literal, token.BIGFLOAT.UserFriendlyName()), p.curToken)
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
		p.error(fmt.Sprintf("expected %s got %s instead", token.RPAREN.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
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

func (p *Parser) parseDeferExpression() ast.Expression {
	se := &ast.DeferExpression{
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

func (p *Parser) parseFromStatement() ast.Statement {
	stmt := &ast.ImportStatement{
		Token: p.curToken,
		Alias: nil,

		IdentsToImport: []*ast.Identifier{},
		ImportAll:      false,
	}

	// return nil if the next token is not an import token
	if !p.expectPeekIs(token.IMPORT_PATH) {
		return nil
	}

	// set the statement name to the identifier with the current token being
	// token.IDENT and the value being the actual string of the identifier
	stmt.Path = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeekIs(token.IMPORT) {
		return nil
	}
	// skip over import
	p.nextToken()

	if p.curTokenIs(token.STAR) {
		// p.nextToken()
		// For some reason we dont need to call next token here?
		stmt.ImportAll = true
		return stmt
	}
	savedPoint := p.peekToken
	// Because we want to use {} we now do this via parseMapOrSet and hope its a set of idents
	list := p.parseMapOrSetLiteral()
	if _, ok := list.(*ast.SetLiteral); !ok {
		p.error(fmt.Sprintf("expected {brackets} to import multiple identifiers got %T instead", list), savedPoint)
		return nil
	}
	for _, e := range list.(*ast.SetLiteral).Elements {
		v, ok := e.(*ast.Identifier)
		if !ok {
			p.error(fmt.Sprintf("expected import elements to be identifiers got %s instead", v), savedPoint)
			return nil
		}
		stmt.IdentsToImport = append(stmt.IdentsToImport, v)
	}

	return stmt
}

func (p *Parser) parseImportStatement() ast.Statement {
	stmt := &ast.ImportStatement{
		Token: p.curToken,
		Alias: nil,

		IdentsToImport: []*ast.Identifier{},
		ImportAll:      false,
	}

	// return nil if the next token is not an import token
	if !p.expectPeekIs(token.IMPORT_PATH) {
		return nil
	}

	// set the statement name to the identifier with the current token being
	// token.IDENT and the value being the actual string of the identifier
	stmt.Path = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if p.peekTokenIs(token.AS) {
		p.nextToken()
		if !p.expectPeekIs(token.IDENT) {
			return nil
		}
		stmt.Alias = &ast.Identifier{Value: p.curToken.Literal}
		// TODO: p.nextToken()?
	}
	// p.nextToken()
	return stmt
}

func (p *Parser) parseBreakStatement() ast.Statement {
	bks := &ast.BreakStatement{
		Token: p.curToken,
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return bks
}

func (p *Parser) parseContinueStatement() ast.Statement {
	cs := &ast.ContinueStatement{
		Token: p.curToken,
	}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
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
		if p.curTokenIs(token.IF) {
			p.nextToken()
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
		if val == nil {
			p.error(fmt.Sprintf("expected %s or identifier got <nil> instead", token.IDENT.UserFriendlyName()), p.curToken)
		} else {
			p.error(fmt.Sprintf("expected %s or identifier got %q instead", token.IDENT.UserFriendlyName(), val.String()), p.curToken)
		}
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
			if val == nil {
				p.error(fmt.Sprintf("expected %s or identifier got <nil> instead", token.IDENT.UserFriendlyName()), p.curToken)
			} else {
				p.error(fmt.Sprintf("expected %s or identifier got %s instead", token.IDENT.UserFriendlyName(), val.String()), p.curToken)
			}
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

	// If the next token is lbrace, parse the block statement as a body and return
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		lit.Body = p.parseBlockStatement()
		return lit
	}
	if !p.curTokenIs(token.RARROW) {
		p.error(fmt.Sprintf("expected %s got %s instead", token.RARROW.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil
	}
	p.nextToken()
	// Otherwise only parse the next statement
	lit.Body = &ast.BlockStatement{Statements: []ast.Statement{}}
	stmt := p.parseStatement()
	lit.Body.Statements = append(lit.Body.Statements, stmt)
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

// parseRegexLiteral will parse the regex literal and return its ast node
func (p *Parser) parseRegexLiteral() ast.Expression {
	return &ast.RegexLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
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
	if !p.peekTokenIs(token.COMMA) && !p.peekTokenIs(token.RBRACE) {
		p.error(fmt.Sprintf("expected %s or %s got %s instead", token.COMMA.UserFriendlyName(), token.RBRACE.UserFriendlyName(), p.peekToken.Type.UserFriendlyName()), p.peekToken)
		return nil
	}
	if p.peekTokenIs(token.COMMA) {
		// Skip curToken and comma token to get to expression to evaluate
		// If its an RBRACE we want to evaluate the single element here so no
		// need to skip ahead
		p.nextToken()
		p.nextToken()
	} else if p.peekTokenIs(token.RBRACE) {
		// This is the case where first expression is valid and next element is not there
		// so 1 element sets
		p.nextToken()
		return exp
	}

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

// parseMapOrSetLiteral parses a map and returns an ast expression node
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

		if p.peekTokenIs(token.COMMA) || p.peekTokenIs(token.RBRACE) {
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

func (p *Parser) parseStructLiteral() ast.Expression {
	firstTok := p.curToken
	exp := &ast.StructLiteral{Token: firstTok}
	exp.Fields = []string{}
	exp.Values = []ast.Expression{}

	keys := make(map[string]struct{})
	for !p.peekTokenIs(token.RBRACE) {
		// get into the map
		p.nextToken()
		key := p.parseIdentifier().(*ast.Identifier)
		if !p.expectIdentIsUnique(key, keys) {
			return nil
		}

		if !p.expectPeekIs(token.COLON) {
			return nil
		}
		// get into the next exp
		p.nextToken()
		value := p.parseExpression(LOWEST)

		exp.Fields = append(exp.Fields, key.Value)
		exp.Values = append(exp.Values, value)

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
		p.error(fmt.Sprintf("expected %s got %s instead", token.INT.UserFriendlyName(), p.peekToken.Type.UserFriendlyName()), p.peekToken)
		return nil
	}
}

// parseForStatement parses a for expression and returns the for expressions ast node
func (p *Parser) parseForStatement() ast.Statement {
	exp := &ast.ForStatement{
		Token:   p.curToken,
		UsesVar: false,
	}
	// current token is for, expect next to be lparen
	if p.curTokenIs(token.FOR) {
		p.nextToken()
	}
	shouldExpectRPAREN := p.curTokenIs(token.LPAREN)

	if p.peekTokenIs(token.VAR) || (!shouldExpectRPAREN && p.curTokenIs(token.VAR)) {
		exp.UsesVar = true
		if shouldExpectRPAREN {
			p.nextToken()
		}
		exp.Initializer = p.parseVarStatement()
		if !p.curTokenIs(token.SEMICOLON) {
			p.error(fmt.Sprintf("expected %s here got %s instead", token.SEMICOLON.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
			return nil
		}
		p.nextToken()
		exp.Condition = p.parseExpression(LOWEST)
		if !p.expectPeekIs(token.SEMICOLON) {
			return nil
		}
		if !p.curTokenIs(token.SEMICOLON) {
			p.error(fmt.Sprintf("expected %s here got %s instead", token.SEMICOLON.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
			return nil
		}
		p.nextToken()
		exp.PostExp = p.parseExpression(LOWEST)
		if shouldExpectRPAREN {
			p.nextToken()
		}
	} else {
		exp.Condition = p.parseExpression(LOWEST)
	}

	if shouldExpectRPAREN && !p.curTokenIs(token.RPAREN) {
		p.error(fmt.Sprintf("expected %s got %s instead", token.RPAREN.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil
	}
	if !p.expectPeekIs(token.LBRACE) {
		return nil
	}
	exp.Consequence = p.parseBlockStatement()
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return exp
}

// parseAssignmentExpression will return a parsed assignment as an Expression ast node
func (p *Parser) parseAssignmentExpression(exp ast.Expression) ast.Expression {
	switch node := exp.(type) {
	case *ast.Identifier, *ast.IndexExpression:
	default:
		p.error(fmt.Sprintf("expected identifier or index expression on left got %T instead", node), p.curToken)
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
	return me
}

func (p *Parser) parseTryCatchBlock() *ast.TryCatchStatement {
	t := p.curToken
	p.nextToken()
	tryBlock := p.parseBlockStatement()
	var catchIdent *ast.Identifier
	var catchBlock *ast.BlockStatement
	var finallyBlock *ast.BlockStatement
	if !p.peekTokenIs(token.CATCH) && !p.peekTokenIs(token.FINALLY) {
		p.error(fmt.Sprintf("expected %s or %s got %s instead", token.CATCH.UserFriendlyName(), token.FINALLY.UserFriendlyName(), p.peekToken.Type.UserFriendlyName()), p.peekToken)
		return nil
	}
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

	skipEndPeek := false
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		if p.curTokenIs(token.RBRACKET) && end == token.RBRACKET {
			// To allow trailing comma in list literal
			skipEndPeek = true
			break
		}
		val := p.parseExpression(LOWEST)
		assignmentExpression, ok := val.(*ast.AssignmentExpression)
		if ok {
			identString := assignmentExpression.Left.String()
			defaultArgs[identString] = assignmentExpression.Value
		} else {
			list = append(list, val)
		}
	}

	if !skipEndPeek && !p.expectPeekIs(end) {
		return nil, nil
	} else if skipEndPeek && !p.curTokenIs(end) {
		p.error(fmt.Sprintf("expected %s got %s instead", token.RPAREN.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil, nil
	}

	return list, defaultArgs
}

func (p *Parser) parseListComprehension(valueToBind ast.Expression) []ast.Expression {
	// Skip over the for
	p.nextToken()
	// current token is for, expect next to be lparen
	shouldExpectRPAREN := false
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		shouldExpectRPAREN = true
	}
	if p.curTokenIs(token.FOR) {
		p.nextToken()
	}

	skipLparen := p.curTokenIs(token.LPAREN) && p.peekTokenIs(token.VAR)
	parseNewFor := p.curTokenIs(token.VAR) || skipLparen
	condStr := ""
	if parseNewFor {
		if skipLparen {
			p.nextToken()
		}
		varStmt := p.parseVarStatement()
		if !p.curTokenIs(token.SEMICOLON) {
			p.error(fmt.Sprintf("expected %s here got %s instead", token.SEMICOLON.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
			return nil
		}
		p.nextToken()
		cond := p.parseExpression(LOWEST)
		if !p.expectPeekIs(token.SEMICOLON) {
			return nil
		}
		if !p.curTokenIs(token.SEMICOLON) {
			p.error(fmt.Sprintf("expected %s here got %s instead", token.SEMICOLON.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
			return nil
		}
		p.nextToken()
		postExp := p.parseExpression(LOWEST)
		condStr = fmt.Sprintf("(%s; %s; %s)", varStmt, cond, postExp)
		if skipLparen {
			p.nextToken()
		}
	} else {
		expCond := p.parseExpression(LOWEST)
		condStr = expCond.String()
	}

	if shouldExpectRPAREN && !p.curTokenIs(token.RPAREN) {
		p.error(fmt.Sprintf("expected %s got %s instead", token.RPAREN.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil
	} else if !shouldExpectRPAREN && p.peekTokenIs(token.RPAREN) {
		p.error(fmt.Sprintf("expected %s here got %s instead", token.RPAREN.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil
	}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	var ifCond ast.Expression
	if p.peekTokenIs(token.IF) {
		// skip expression ending/RPAREN
		p.nextToken()
		// skip IF
		p.nextToken()
		ifCond = p.parseExpression(LOWEST)
	}

	var program string
	if ifCond != nil {
		program = fmt.Sprintf("var __internal__ = []; for %s { if %s { var __result__ = %s; __internal__ << __result__; } };", condStr, ifCond, valueToBind.String())
	} else {
		program = fmt.Sprintf("var __internal__ = []; for %s { var __result__ = %s; __internal__ << __result__; };", condStr, valueToBind.String())
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
	shouldExpectRPAREN := false
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		shouldExpectRPAREN = true
	}
	if p.curTokenIs(token.FOR) {
		p.nextToken()
	}

	skipLparen := p.curTokenIs(token.LPAREN) && p.peekTokenIs(token.VAR)
	parseNewFor := p.curTokenIs(token.VAR) || skipLparen
	condStr := ""
	if parseNewFor {
		if skipLparen {
			p.nextToken()
		}
		varStmt := p.parseVarStatement()
		if !p.curTokenIs(token.SEMICOLON) {
			p.error(fmt.Sprintf("expected %s here got %s instead", token.SEMICOLON.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
			return nil
		}
		p.nextToken()
		cond := p.parseExpression(LOWEST)
		if !p.expectPeekIs(token.SEMICOLON) {
			return nil
		}
		if !p.curTokenIs(token.SEMICOLON) {
			p.error(fmt.Sprintf("expected %s here got %s instead", token.SEMICOLON.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
			return nil
		}
		p.nextToken()
		postExp := p.parseExpression(LOWEST)
		condStr = fmt.Sprintf("(%s; %s; %s)", varStmt, cond, postExp)
		if skipLparen {
			p.nextToken()
		}
	} else {
		expCond := p.parseExpression(LOWEST)
		condStr = expCond.String()
	}

	if shouldExpectRPAREN && !p.curTokenIs(token.RPAREN) {
		p.error(fmt.Sprintf("expected %s got %s instead", token.RPAREN.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil
	} else if !shouldExpectRPAREN && p.peekTokenIs(token.RPAREN) {
		p.error(fmt.Sprintf("expected %s here got %s instead", token.RPAREN.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil
	}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	var ifCond ast.Expression
	if p.peekTokenIs(token.IF) {
		// skip RPAREN/ending exp
		p.nextToken()
		// skip IF
		p.nextToken()
		ifCond = p.parseExpression(LOWEST)
	}

	var program string
	if ifCond != nil {
		program = fmt.Sprintf("var __internal__ = {}; for %s { if %s { __internal__[%s] = %s } };", condStr, ifCond, key.String(), value.String())
	} else {
		program = fmt.Sprintf("var __internal__ = {}; for %s { __internal__[%s] = %s  };", condStr, key.String(), value.String())
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
	shouldExpectRPAREN := false
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		shouldExpectRPAREN = true
	}
	if p.curTokenIs(token.FOR) {
		p.nextToken()
	}

	skipLparen := p.curTokenIs(token.LPAREN) && p.peekTokenIs(token.VAR)
	parseNewFor := p.curTokenIs(token.VAR) || skipLparen
	condStr := ""
	if parseNewFor {
		if skipLparen {
			p.nextToken()
		}
		varStmt := p.parseVarStatement()
		if !p.curTokenIs(token.SEMICOLON) {
			p.error(fmt.Sprintf("expected %s here got %s instead", token.SEMICOLON.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
			return nil
		}
		p.nextToken()
		cond := p.parseExpression(LOWEST)
		if !p.expectPeekIs(token.SEMICOLON) {
			return nil
		}
		if !p.curTokenIs(token.SEMICOLON) {
			p.error(fmt.Sprintf("expected %s here got %s instead", token.SEMICOLON.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
			return nil
		}
		p.nextToken()
		postExp := p.parseExpression(LOWEST)
		condStr = fmt.Sprintf("(%s; %s; %s)", varStmt, cond, postExp)
		if skipLparen {
			p.nextToken()
		}
	} else {
		expCond := p.parseExpression(LOWEST)
		condStr = expCond.String()
	}

	if shouldExpectRPAREN && !p.curTokenIs(token.RPAREN) {
		p.error(fmt.Sprintf("expected %s got %s instead", token.RPAREN.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil
	} else if !shouldExpectRPAREN && p.peekTokenIs(token.RPAREN) {
		p.error(fmt.Sprintf("expected %s here got %s instead", token.RPAREN.UserFriendlyName(), p.curToken.Type.UserFriendlyName()), p.curToken)
		return nil
	}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
	}

	var ifCond ast.Expression
	if p.peekTokenIs(token.IF) {
		// skip RPAREN/ending exp
		p.nextToken()
		// skip IF
		p.nextToken()
		ifCond = p.parseExpression(LOWEST)
	}

	var program string
	if ifCond != nil {
		program = fmt.Sprintf("var __internal__ = []; for %s { if %s { __internal__ << %s } }; __internal__ = set(__internal__);", condStr, ifCond, value.String())
	} else {
		program = fmt.Sprintf("var __internal__ = []; for %s { __internal__ << %s }; __internal__ = set(__internal__);", condStr, value.String())
	}
	if !p.expectPeekIs(token.RBRACE) {
		return nil
	}

	return &ast.SetCompLiteral{Token: tok, NonEvaluatedProgram: program}
}

// stringLexer is used to parse string interpolation values
type stringLexer struct {
	input        string
	runeInput    []rune
	position     int  // current pos. in input (points to current char)
	readPosition int  // current reading pos. in input (after current char)
	ch           rune // current char under examination
}

// newStringLexer takes an input and returns a *stringLexer
func newStringLexer(input string) *stringLexer {
	sl := &stringLexer{input: input, runeInput: []rune(input)}
	sl.readStringChar()
	return sl
}

// readStringChar reads the next char in the input and advances
// the read position
func (sl *stringLexer) readStringChar() {
	if sl.readPosition >= len(sl.runeInput) {
		sl.ch = 0
	} else {
		sl.ch = sl.runeInput[sl.readPosition]
	}
	sl.position = sl.readPosition
	sl.readPosition++
}

// peekChar checks the next readPosition char and returns the byte
// without advancing the position
func (sl *stringLexer) peekChar() rune {
	if sl.readPosition >= len(sl.runeInput) {
		return 0
	}
	return sl.runeInput[sl.readPosition]
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
				toLex.WriteRune(sl.ch)
				sl.readStringChar()
			}
		}
		if sl.ch == '}' {
			l := lexer.New(toLex.String(), "<internal:StringInterpolation>")
			lcpy := *l
			var tok token.Token
			for {
				tok = lcpy.NextToken()
				if tok.Type == token.ILLEGAL {
					p.error(fmt.Sprintf("%s token encountered. got=%q", tok.Type, tok.Literal), tok)
				}
				if tok.Type == token.EOF {
					break
				}
			}
			origStrings = append(origStrings, fmt.Sprintf("#{%s}", toLex.String()))
			parseString := New(l)
			parsedExp := parseString.parseExpression(LOWEST)
			if parsedExp == nil {
				// If the Interpolation is empty #{} we want to replace with empty string for the evaluator to use
				interps = append(interps, &ast.StringLiteral{Value: ""})
			} else {
				interps = append(interps, parsedExp)
			}
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
	// create a peek error using the unified error method
	p.error(fmt.Sprintf("expected %s got %s instead", t.UserFriendlyName(), p.peekToken.Type.UserFriendlyName()), p.peekToken)
	return false
}

func (p *Parser) expectIdentIsUnique(key *ast.Identifier, keys map[string]struct{}) bool {
	if _, ok := keys[key.Value]; ok {
		p.error(fmt.Sprintf("struct literal keys must be unique, current identifier %s", key.Value), key.Token)
		return false
	} else {
		keys[key.Value] = struct{}{}
	}
	return true
}

func (p *Parser) peekTokenIsAssignmentToken() bool {
	return p.peekTokenIs(token.ASSIGN) ||
		p.peekTokenIs(token.PLUSEQ) ||
		p.peekTokenIs(token.MINUSEQ) ||
		p.peekTokenIs(token.DIVEQ) ||
		p.peekTokenIs(token.FDIVEQ) ||
		p.peekTokenIs(token.MULEQ) ||
		p.peekTokenIs(token.POWEQ) ||
		p.peekTokenIs(token.ANDEQ) ||
		p.peekTokenIs(token.OREQ) ||
		p.peekTokenIs(token.ANDANDEQ) ||
		p.peekTokenIs(token.OROREQ) ||
		p.peekTokenIs(token.BINNOTEQ) ||
		p.peekTokenIs(token.LSHIFTEQ) ||
		p.peekTokenIs(token.RSHIFTEQ) ||
		p.peekTokenIs(token.PERCENTEQ) ||
		p.peekTokenIs(token.XOREQ)
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
