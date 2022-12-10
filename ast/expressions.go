package ast

import (
	"blue/token"
	"bytes"
	"strings"
)

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
func (ie *InfixExpression) expressionNode() {}

// TokenLiteral returns the infix expressions token
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

// String returns the string representation of the infix expression ast node
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" ")
	out.WriteString(ie.Operator)
	out.WriteString(" ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

// IfExpression is the if expression ast node
type IfExpression struct {
	Token        token.Token       // Token == IF
	Conditions   []Expression      // Conditions is a list of expressions for if statements
	Consequences []*BlockStatement // Consequences is a list of block statemenets that evaluate if true
	Alternative  *BlockStatement   // Alternative is a block of statements that evaluate if false
}

// expressionNode satisfies the Expression Interface
func (ie *IfExpression) expressionNode() {}

// TokenLiteral returns the string IF token
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

// String returns the string representation of the if expression
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	for i := 0; i < len(ie.Conditions); i++ {
		if i == 0 {
			out.WriteString("if (")
		} else {
			out.WriteString("else if (")
		}
		out.WriteString(ie.Conditions[i].String())
		out.WriteString(") {")
		out.WriteString(ie.Consequences[i].ExpressionString())
		out.WriteString(" } ")
	}

	if ie.Alternative != nil {
		out.WriteString("else { ")
		out.WriteString(ie.Alternative.ExpressionString())
		out.WriteString(" }")
	}
	return out.String()
}

// MatchExpression is the match expression ast node
type MatchExpression struct {
	Token         token.Token       // Token == MATCH
	OptionalValue Expression        // OptionalValue is the value that could be used to check against the conditions
	Conditions    []Expression      // Condition is an expression to determine whether to run the Consequence
	Consequences  []*BlockStatement // Consequence is a block statement to run if the condition in the same position is true
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
	out.WriteString(" { ")
	for i, e := range me.Conditions {
		out.WriteString(e.String())
		out.WriteString(" => { ")
		out.WriteString(me.Consequences[i].ExpressionString())
		out.WriteString(" }, ")
	}
	out.WriteString(" } ")

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
	out.WriteByte('(')
	out.WriteString(strings.Join(args, ", "))
	out.WriteByte(')')

	return out.String()
}

// IndexExpression is the ast node of an index call expression
type IndexExpression struct {
	Token token.Token // Token is [ or .
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

	isDotCall := ie.Token.Literal == "."

	out.WriteByte('(')
	out.WriteString(ie.Left.String())
	if isDotCall {
		out.WriteByte('.')
	} else {
		out.WriteByte('[')
	}
	if isDotCall {
		out.WriteString(strings.ReplaceAll(ie.Index.String(), "'", ""))
	} else {
		out.WriteString(ie.Index.String())
		out.WriteByte(']')
	}
	out.WriteByte(')')

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
	out.WriteString(") { ")
	out.WriteString(fe.Consequence.ExpressionString())
	out.WriteString(" } ")
	return out.String()
}

// AssignmentExpression is the type that supports rebinding variables
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

	out.WriteString(ae.Left.String())
	out.WriteByte(' ')
	out.WriteString(ae.TokenLiteral())
	out.WriteByte(' ')
	out.WriteString(ae.Value.String())

	return out.String()
}

// EvalExpression is the eval ast node
type EvalExpression struct {
	Token     token.Token // token == eval
	StrToEval Expression  // StrToEval is the Expression (that should be a string) to eval in the current env context
}

// expressionNode satisfies the expression interface
func (ee *EvalExpression) expressionNode() {}

// TokenLiteral returns the for token
func (ee *EvalExpression) TokenLiteral() string { return ee.Token.Literal }

// String returns the string representation of the for expression ast node
func (ee *EvalExpression) String() string {
	var out bytes.Buffer

	out.WriteString("eval(\"")
	out.WriteString(ee.StrToEval.String())
	out.WriteString("\")")
	return out.String()
}

// SpawnExpression is the spaws ast node
type SpawnExpression struct {
	Token     token.Token  // token == spawn
	Arguments []Expression // Arguments is the list of expression to be passed as arguments
}

// expressionNode satisfies the expression interface
func (se *SpawnExpression) expressionNode() {}

// TokenLiteral returns the for token
func (se *SpawnExpression) TokenLiteral() string { return se.Token.Literal }

// String returns the string representation of the for expression ast node
func (se *SpawnExpression) String() string {
	var out bytes.Buffer

	out.WriteString("spawn(")
	for i, a := range se.Arguments {
		out.WriteString(a.String())
		if i < len(se.Arguments)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteByte(')')
	return out.String()
}

type SelfExpression struct {
	Token token.Token // token == spawn
}

// expressionNode satisfies the expression interface
func (se *SelfExpression) expressionNode() {}

// TokenLiteral returns the for token
func (se *SelfExpression) TokenLiteral() string { return se.Token.Literal }

// String returns the string representation of the for expression ast node
func (se *SelfExpression) String() string {
	return "self()"
}
