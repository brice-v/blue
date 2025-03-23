package ast

import (
	"blue/token"
	"bytes"
	"math/big"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

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

// BigFloatLiteral is the big float literal ast node
type BigFloatLiteral struct {
	Token token.Token     // token == token.FLOAT
	Value decimal.Decimal // Value stores the big float value
}

// expressionNode satisfies the Expression interface
func (bfl *BigFloatLiteral) expressionNode() {}

// TokenLiteral returns the string value of the big int
func (bfl *BigFloatLiteral) TokenLiteral() string { return bfl.Token.Literal }

// String returns the string value of the big float
func (bfl *BigFloatLiteral) String() string { return bfl.Token.Literal }

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

// UIntegerLiteral is the binary literal expression
type UIntegerLiteral struct {
	Token token.Token // Token == token.BINARY
	Value uint64      // Value stores the binary as an uint64
}

// expressionNode satisfies the Expression interface
func (ul *UIntegerLiteral) expressionNode() {}

// TokenLiteral returns the string value of the binary number
func (ul *UIntegerLiteral) TokenLiteral() string { return ul.Token.Literal }

// String returns the string value of the binary number
func (ul *UIntegerLiteral) String() string { return ul.Token.Literal }

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
	out.WriteString(" ) { ")
	out.WriteString(fl.Body.String())
	out.WriteString(" } ")

	return out.String()
}

// ExecStringLiteral is the contents of a string within backticks `
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
func (sl *StringLiteral) String() string {
	// The reason we differentiate is so that the quotes are properly escaped when parsed internally
	if sl.Token.Type == token.STRING_DOUBLE_QUOTE {
		return `"` + sl.Value + `"`
	} else {
		return `'` + sl.Value + `'`
	}
}

// StringWithoutQuotes returns the string value without quotes
func (sl *StringLiteral) StringWithoutQuotes() string { return sl.Value }

type RegexLiteral struct {
	Token token.Token
	Value string
}

func (rl *RegexLiteral) expressionNode()      {}
func (rl *RegexLiteral) TokenLiteral() string { return rl.Token.Literal }
func (rl *RegexLiteral) String() string {
	// replace literal backslash as escaped to make it easier to see
	return "r/" + strings.ReplaceAll(rl.Value, "\\", "\\\\") + "/"
}

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
	Token      token.Token               // Token == {
	Pairs      map[Expression]Expression // Pairs is a map of expressions to expressions
	PairsIndex map[int]Expression        // Insertion Index -> Key Expression
}

// expressionNode satisfies the expression interface
func (ml *MapLiteral) expressionNode() {}

// TokenLiteral returns the { token as a string
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }

// String returns the string representation of the map literal ast node
func (ml *MapLiteral) String() string {
	var out bytes.Buffer

	indices := []int{}
	for k := range ml.PairsIndex {
		indices = append(indices, k)
	}
	sort.Ints(indices)
	pairs := []string{}
	for _, i := range indices {
		k := ml.PairsIndex[i]
		v := ml.Pairs[k]
		pairs = append(pairs, k.String()+": "+v.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

// MapCompLiteral is the struct representing a map comprehension
type MapCompLiteral struct {
	Token               token.Token // Doesnt really have a token
	NonEvaluatedProgram string      // The program we will evaluate in evaluator
}

// expressionNode satisfies the expression interface
func (mcl *MapCompLiteral) expressionNode() {}

// TokenLiteral returns something but mcl currently doesnt really support it
func (mcl *MapCompLiteral) TokenLiteral() string {
	return mcl.Token.Literal
}

// String returns the program to execute
func (mcl *MapCompLiteral) String() string {
	return mcl.NonEvaluatedProgram
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

// SetCompLiteral is the struct representing a set comprehension
type SetCompLiteral struct {
	Token               token.Token // Doesnt really have a token
	NonEvaluatedProgram string      // The program we will evaluate in evaluator
}

// expressionNode satisfies the expression interface
func (scl *SetCompLiteral) expressionNode() {}

// TokenLiteral returns something but scl currently doesnt really support it
func (scl *SetCompLiteral) TokenLiteral() string {
	return scl.Token.Literal
}

// String returns the program to execute
func (scl *SetCompLiteral) String() string {
	return scl.NonEvaluatedProgram
}

type StructLiteral struct {
	Token       token.Token                // Token == @{
	Fields      map[*Identifier]Expression // Fields is a map of identifiers to expressions
	FieldsIndex map[int]*Identifier        // Insertion Index -> identifier
}

// expressionNode satisfies the expression interface
func (sl *StructLiteral) expressionNode() {}

// TokenLiteral returns the { token as a string
func (sl *StructLiteral) TokenLiteral() string { return sl.Token.Literal }

// String returns the string representation of the map literal ast node
func (sl *StructLiteral) String() string {
	var out bytes.Buffer

	indices := []int{}
	for k := range sl.FieldsIndex {
		indices = append(indices, k)
	}
	sort.Ints(indices)
	pairs := []string{}
	for _, i := range indices {
		k := sl.FieldsIndex[i]
		v := sl.Fields[k]
		pairs = append(pairs, k.String()+": "+v.String())
	}

	out.WriteString("@{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
