package ast

import (
	"bytes"
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
		out.WriteByte('\n')
	}

	return out.String()
}
