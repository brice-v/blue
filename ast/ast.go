package ast

import (
	"blue/token"
	"bytes"
	"fmt"
	"math/big"
	"strings"
)

// Node asserts that to be a node their must be a method
// that returns the token name as a string
type Node interface {
	// TokenLiteral is used for debugging and testing
	TokenLiteral() string
	// String
	String() string // String will allow the printing of ast nodes
}

// Statement is a node with a statementNode method
type Statement interface {
	Node
	statementNode()
}

// Expression is a node with an expressionNode method
type Expression interface {
	Node
	expressionNode()
}

// Program defines a struct that contains a slice of statement nodes
// any valid program is a slice of statements
type Program struct {
	Statements []Statement
}

// TokenLiteral makes the Program struct a Node and becomes
// the root node of the ast
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// String will return the entire program ast as a string
func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// Identifier is the node for the ident token
type Identifier struct {
	Token token.Token // Token == token.IDENT
	Value string      // Value is the actual identifier string
}

// expressionNode makes identifers expressions
func (i *Identifier) expressionNode() {}

// TokenLiteral returns IDENT
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// String returns the string value of the identifier
func (i *Identifier) String() string { return i.Value }

// VarStatement is the node for var statements
type VarStatement struct {
	Token           token.Token // Token == token.VAR
	Name            *Identifier // Name is the identifier that Value is being binded to
	Value           Expression  // Value is the expression node that is being assinged to
	AssignmentToken token.Token // AssignmentToken is the token used for assignment
}

// statementNode makes var a statement
func (vars *VarStatement) statementNode() {}

// TokenLiteral returns VAR
func (vars *VarStatement) TokenLiteral() string { return vars.Token.Literal }

// String returns the VarStatement node as a string
func (vars *VarStatement) String() string {
	var out bytes.Buffer

	out.WriteString(vars.TokenLiteral() + " ")
	out.WriteString(vars.Name.String())
	out.WriteString(" ")
	out.WriteString(vars.AssignmentToken.Literal)
	out.WriteString(" ")

	if vars.Value != nil {
		out.WriteString(vars.Value.String())
	}

	out.WriteString(";\n")

	return out.String()
}

// ValStatement is the node for val statements
type ValStatement struct {
	Token token.Token // Token == token.VAL
	Name  *Identifier // Name is the identifier that Value is being binded to
	Value Expression  // Value is the expression node that is being assinged to
}

// statementNode makes val a statement
func (vals *ValStatement) statementNode() {}

// TokenLiteral returns VAL
func (vals *ValStatement) TokenLiteral() string { return vals.Token.Literal }

// String returns the ValStatement node as a string
func (vals *ValStatement) String() string {
	var out bytes.Buffer

	out.WriteString(vals.TokenLiteral() + " ")
	out.WriteString(vals.Name.String())
	out.WriteString(" = ")

	if vals.Value != nil {
		out.WriteString(vals.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// FunctionStatement is the function definition that is used at the source leve
// this is what allows fun hello() to assign the identifier `hello` to the function
// literal
type FunctionStatement struct {
	Token                token.Token     // Token == token.FUNCTION
	Name                 *Identifier     // Name is the identifier to assign the function literal to
	Body                 *BlockStatement // Body is a block statement containing the work to be done
	Parameters           []*Identifier
	ParameterExpressions []Expression // ParameterExpressions defines the expression to perform for identifier if
	// if it is not nil the value will be used as the default parameter
}

// statementNode satisfies the statement interface
func (fs *FunctionStatement) statementNode() {}

// TokenLiteral is the function statements token literal ie. `fun`
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }

// String returns a stringified version of the function statement ast node
func (fs *FunctionStatement) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fun ")
	out.WriteString(fs.Name.String() + "(")
	out.WriteString(strings.Join(params, ", ") + ")")
	out.WriteString(" {\n\t")
	out.WriteString(fs.Body.String())
	out.WriteString("\n}\n")

	return out.String()
}

// ReturnStatement is the node for return statements
type ReturnStatement struct {
	Token       token.Token // Token == token.RETURN
	ReturnValue Expression  // ReturnValue is an expression node that returns a value
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns RETURN
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// String returns the ReturnStatement node as a string
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// ExpressionStatement is the node for expression statements
type ExpressionStatement struct {
	Token      token.Token // Token is the first token of the expression
	Expression Expression  // Expression is the expression node that evaluates to something
}

// statementNode satisfys the statement interface and allows it to be added to the program
func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns the first token of the expression
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// String will return the string version of the expression statement
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BigIntegerLiteral is the big int literal ast node
type BigIntegerLiteral struct {
	Token token.Token // token == token.INT
	Value *big.Int    // Value stores the big integer value
}

// expressionNode satisfies the Expression interface
func (bil *BigIntegerLiteral) expressionNode() {}

// TokenLiteral returns the string value of the big int
func (bil *BigIntegerLiteral) TokenLiteral() string { return bil.Token.Literal }

// String returns the string value of the big int
func (bil *BigIntegerLiteral) String() string { return bil.Token.Literal }

// IntegerLiteral is the integer literal expression
type IntegerLiteral struct {
	Token token.Token // Token == token.INT
	Value int64       // Value stores the integer as an int64
}

// expressionNode satisfies the Expression interface
func (il *IntegerLiteral) expressionNode() {}

// TokenLiteral returns the string value of the int
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// String returns the string value of the int
func (il *IntegerLiteral) String() string { return il.Token.Literal }

// FloatLiteral is the float literal expression
type FloatLiteral struct {
	Token token.Token // Token == token.FLOAT
	Value float64     // Value stores the float as an float64
}

// expressionNode satisfies the Expression interface
func (fl *FloatLiteral) expressionNode() {}

// TokenLiteral returns the string value of the float
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }

// String returns the string value of the float
func (fl *FloatLiteral) String() string { return fl.Token.Literal }

// HexLiteral is the hex literal expression
type HexLiteral struct {
	Token token.Token // Token == token.HEX
	Value uint64      // Value stores the hex as an uint64
}

// expressionNode satisfies the Expression interface
func (hl *HexLiteral) expressionNode() {}

// TokenLiteral returns the string value of the hex number
func (hl *HexLiteral) TokenLiteral() string { return hl.Token.Literal }

// String returns the string value of the hex number
func (hl *HexLiteral) String() string { return hl.Token.Literal }

// OctalLiteral is the octal literal expression
type OctalLiteral struct {
	Token token.Token // Token == token.OCTAL
	Value uint64      // Value stores the octal as an uint64
}

func (ol *OctalLiteral) expressionNode() {}

// TokenLiteral returns the string value of the ocatal number
func (ol *OctalLiteral) TokenLiteral() string { return ol.Token.Literal }

// String returns the string value of the octal number
func (ol *OctalLiteral) String() string { return ol.Token.Literal }

// BinaryLiteral is the binary literal expression
type BinaryLiteral struct {
	Token token.Token // Token == token.BINARY
	Value uint64      // Value stores the binary as an uint64
}

// expressionNode satisfies the Expression interface
func (bl *BinaryLiteral) expressionNode() {}

// TokenLiteral returns the string value of the binary number
func (bl *BinaryLiteral) TokenLiteral() string { return bl.Token.Literal }

// String returns the string value of the binary number
func (bl *BinaryLiteral) String() string { return bl.Token.Literal }

// PrefixExpression is the prefix expression ast node
type PrefixExpression struct {
	Token    token.Token // Token is the prefix token, ! -
	Operator string      // Operator is the string rep. of the operation
	Right    Expression  // Right is the right expression to evaluate after
}

// expressionNode satisfies the Expression interface
func (pe *PrefixExpression) expressionNode() {}

// TokenLiteral returns the prefix expressions token
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// String returns the string representation of the prefix expression ast node
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// InfixExpression is the infix expression ast node
type InfixExpression struct {
	Token    token.Token // Token is the infix token
	Operator string      // Operator is the string rep. of the operation
	Left     Expression  // Left is the left expression of the infix operator
	Right    Expression  // Right is the right expression to evaluate after
}

// expressionNode satisfies the Expression interface
func (oe *InfixExpression) expressionNode() {}

// TokenLiteral returns the infix expressions token
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }

// String returns the string representation of the infix expression ast node
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" ")
	out.WriteString(oe.Operator)
	out.WriteString(" ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")

	return out.String()
}

// Null is the null ast node
type Null struct {
	Token token.Token
}

func (n *Null) expressionNode() {}

// TokenLiteral returns the string token literal
func (n *Null) TokenLiteral() string {
	return n.Token.Literal
}

func (n *Null) String() string { return "null" }

// Boolean is the boolean literal ast node
type Boolean struct {
	Token token.Token
	Value bool
}

// expressionNode satisfies the Expression interface
func (b *Boolean) expressionNode() {}

// TokenLiteral returns true or false
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }

// String returns true or false as a string
func (b *Boolean) String() string { return b.Token.Literal }

// IfExpression is the if expression ast node
type IfExpression struct {
	Token       token.Token     // Token == IF
	Condition   Expression      // Condition is an expression for if statements
	Consequence *BlockStatement // Consequence is a block of statemenets that evaluate if true
	Alternative *BlockStatement // Alternative is a block of statements that evaluate if false
}

// expressionNode satisfies the Expression Interface
func (ie *IfExpression) expressionNode() {}

// TokenLiteral returns the string IF token
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

// String returns the string representation of the if expression
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" {\n\t")
	out.WriteString(ie.Consequence.String())
	out.WriteString("\n}")
	if ie.Alternative == nil {
		out.WriteString("\n")
	}

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
		out.WriteString("\n")
	}
	return out.String()
}

// MatchExpression is the match expression ast node
type MatchExpression struct {
	Token         token.Token       // Token == MATCH
	OptionalValue Expression        // OptionalValue is the value that could be used to check against the conditions
	Condition     []Expression      // Condition is an expression to determine whether to run the Consequence
	Consequence   []*BlockStatement // Consequence is a block statement to run if the condition in the same position is true
}

// expressionNode satisfies the expression interface
func (me *MatchExpression) expressionNode() {}

// TokenLiteral returns the match literal token
func (me *MatchExpression) TokenLiteral() string { return me.Token.Literal }

// String returns the stringified version of the match statment
func (me *MatchExpression) String() string {
	var out bytes.Buffer

	out.WriteString("match ")
	out.WriteString(me.OptionalValue.String())
	out.WriteString(" {\n")
	for i, e := range me.Condition {
		out.WriteString("\t")
		out.WriteString(e.String())
		out.WriteString(" => {")
		out.WriteString(me.Consequence[i].String())
		out.WriteString("},\n")
	}
	out.WriteString("}\n")

	return out.String()
}

// BlockStatement is the ast node for block statements
type BlockStatement struct {
	Token      token.Token // Token == {
	Statements []Statement // Statements is the list of statements in the block
}

// statementNode satisifes the statement interface
func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns the { token
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// String returns the string representation of the block statement
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// FunctionLiteral is the functional literal ast node
type FunctionLiteral struct {
	Token                token.Token // Token == FUNCTION
	Parameters           []*Identifier
	ParameterExpressions []Expression // ParameterExpressions defines the expression to perform for identifier if
	// if it is not nil the value will be used as the default parameter
	Body *BlockStatement
}

// expressionNode satisfies the expression interface
func (fl *FunctionLiteral) expressionNode() {}

// TokenLiteral returns the FUNCTION token
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }

// String returns the string representation of the function literal
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(" ) {\n\t")
	out.WriteString(fl.Body.String())
	out.WriteString("\n}\n")

	return out.String()
}

// CallExpression is the ast node for call expression
type CallExpression struct {
	Token     token.Token  // Token == (
	Function  Expression   // Function is the expression being called
	Arguments []Expression // Arguments is the list of expression to be passed as arguments

	DefaultArguments map[string]Expression // DefaultArguments is the map of the identifer as a string to the expression to be used as the value
}

// expressionNode satisfies the expression interface
func (ce *CallExpression) expressionNode() {}

// TokenLiteral returns the ( token
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

// String returns the string representation of the call expression
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	// TODO: Put a \n here to make the ast print nicer.  This makes tests fail though
	out.WriteString(")")

	return out.String()
}

// ExecStringLiteral is the contents of a string within backticks ``
type ExecStringLiteral struct {
	Token token.Token
	Value string
}

// expressionNode satisfies the expression interface
func (esl *ExecStringLiteral) expressionNode() {}

// TokenLiteral returns the backtick token
func (esl *ExecStringLiteral) TokenLiteral() string { return esl.Token.Literal }

// String returns the string representation of the exec string cmd
func (esl *ExecStringLiteral) String() string { return "`" + esl.Value + "`" }

// StringLiteral represents a string ast node
type StringLiteral struct {
	Token               token.Token  // Token == "
	Value               string       // Value is the full string (with interpolation not removed)
	InterpolationValues []Expression // InterpolationValues is the expressions that need to be evaluated and put back into the string

	OriginalInterpolationString []string // OriginalInterpolationString is a slice of strings to use to replace with interpolation
}

// expressionNode satisfies the expression interface
func (sl *StringLiteral) expressionNode() {}

// TokenLiteral returns the " token
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// String returns the string representation of the string literal ast node
func (sl *StringLiteral) String() string { return `"` + sl.Value + `"` }

// StringWithoutQuotes returns the string value without quotes
func (sl *StringLiteral) StringWithoutQuotes() string { return sl.Value }

// ListLiteral is the list literal ast node representation
type ListLiteral struct {
	Token    token.Token  // Token == [ (LBRACE)
	Elements []Expression // Elements in a list are expressions
}

// expressionNode satisfies the
func (ll *ListLiteral) expressionNode() {}

// TokenLiteral returns the [ token
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal }

// String returns the string representation of the list literal ast node
func (ll *ListLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range ll.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// ListCompLiteral is the struct representing a list comprehension
type ListCompLiteral struct {
	Token               token.Token // Doesnt really have a token
	NonEvaluatedProgram string      // The program we will evaluate in evaluator
}

// expressionNode satisfies the expression interface
func (lcl *ListCompLiteral) expressionNode() {}

// String returns the program to execute
func (lcl *ListCompLiteral) String() string {
	return lcl.NonEvaluatedProgram
}

// TokenLiteral returns something but lcl currently doesnt really support it
func (lcl *ListCompLiteral) TokenLiteral() string {
	return lcl.Token.Literal
}

// MapLiteral is the representation of the map literal ast node
type MapLiteral struct {
	Token token.Token               // Token == {
	Pairs map[Expression]Expression // Pairs is a map of expressions to expressions
}

// expressionNode satisfies the expression interface
func (ml *MapLiteral) expressionNode() {}

// TokenLiteral returns the { token as a string
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }

// String returns the string representation of the map literal ast node
func (ml *MapLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for k, v := range ml.Pairs {
		pairs = append(pairs, k.String()+": "+v.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

// IndexExpression is the ast node of an index call expression
type IndexExpression struct {
	Token token.Token // Token == [
	Left  Expression
	Index Expression
}

// expressionNode satisfies the expression interface
func (ie *IndexExpression) expressionNode() {}

// TokenLiteral returns the [ token
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }

// String returns a string representation of an index call expression
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

// ForExpression is the for loop ast node
type ForExpression struct {
	Token       token.Token     // token == for
	Condition   Expression      // Condition is the condition to test whether the loop should continue
	Consequence *BlockStatement // Consequence contains a block of statements that happen if the condition is true
}

// expressionNode satisfies the expression interface
func (fe *ForExpression) expressionNode() {}

// TokenLiteral returns the for token
func (fe *ForExpression) TokenLiteral() string { return fe.Token.Literal }

// String returns the string representation of the for expression ast node
func (fe *ForExpression) String() string {
	var out bytes.Buffer

	out.WriteString("for (")
	out.WriteString(fe.Condition.String())
	out.WriteString(") {\n\t")
	out.WriteString(fe.Consequence.String())
	out.WriteString("\n}\n")
	return out.String()
}

// AssignmentExpression is the type that supports rebinding variables
// TODO: This should only be allowed on mutable fields/values - need to figure this out
type AssignmentExpression struct {
	Token token.Token // Token is the assignment token being used
	Left  Expression  // Left is an expression to get assigned to
	Value Expression  // Value is an expression being used to assign
}

// expressionNode satisfies the expression interface
func (ae *AssignmentExpression) expressionNode() {}

// TokenLiteral prints the literal value of the token associated with this node
func (ae *AssignmentExpression) TokenLiteral() string { return ae.Token.Literal }

// String returns a stringified version of the AST for debugging
func (ae *AssignmentExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ae.Left.String() + " ")
	out.WriteString(ae.TokenLiteral())
	out.WriteString(" " + ae.Value.String())

	return out.String()
}

// SetLiteral is the set literal struct ast node
type SetLiteral struct {
	Token    token.Token  // Token == {
	Elements []Expression // Elements is a slice of expressions (to be mapped to a map[Object]bool where checks can be made at evaluation)
}

// expressionNode satisfies the expression interface
func (set *SetLiteral) expressionNode() {}

// TokenLiteral prints the set token literal
func (set *SetLiteral) TokenLiteral() string { return set.Token.Literal }

// String returns a stringified version of the AST for debugging
func (set *SetLiteral) String() string {
	var out bytes.Buffer

	elems := []string{}
	for _, e := range set.Elements {
		elems = append(elems, e.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(elems, ", "))
	out.WriteString("}")

	return out.String()
}

// ImportStatement is the representation of the map literal ast node
type ImportStatement struct {
	Token token.Token // Token == import
	Path  *Identifier // Path is the import's path which refers to a file
}

// statementNode satisfies the statement interface
func (is *ImportStatement) statementNode() {}

// TokenLiteral returns the import token as a string
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }

// String returns the string representation of the map literal ast node
func (is *ImportStatement) String() string {
	return fmt.Sprintf("%s %s", is.Token.Literal, is.Path)
}
