package ast

import (
	"blue/token"
	"bytes"
	"fmt"
	"strings"
)

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
