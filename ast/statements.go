package ast

import (
	"blue/token"
	"bytes"
	"fmt"
	"strings"
)

// VarStatement is the node for var statements
type VarStatement struct {
	Token            token.Token // Token == token.VAR
	KeyValueNames    map[Expression]*Identifier
	Names            []*Identifier // Names are the identifiers that Value is being binded to (or that we are desctructuring)
	Value            Expression    // Value is the expression node that is being assinged to
	IsMapDestructor  bool
	IsListDestructor bool

	AssignmentToken token.Token // AssignmentToken is the token used for assignment (destructuring only supports =)
}

// statementNode makes var a statement
func (vars *VarStatement) statementNode() {}

// TokenLiteral returns VAR
func (vars *VarStatement) TokenLiteral() string    { return vars.Token.Literal }
func (vars *VarStatement) TokenToken() token.Token { return vars.Token }

// String returns the VarStatement node as a string
func (vars *VarStatement) String() string {
	var out bytes.Buffer

	out.WriteString(vars.TokenLiteral())
	out.WriteByte(' ')
	if vars.IsListDestructor {
		out.WriteByte('[')
	} else if vars.IsMapDestructor {
		out.WriteByte('{')
	}
	for i, name := range vars.Names {
		out.WriteString(name.String())
		if i != len(vars.Names)-1 {
			out.WriteString(", ")
		}
	}
	if vars.IsMapDestructor {
		if len(vars.KeyValueNames) > 0 {
			out.WriteString(", ")
		}
		i := 0
		for k, v := range vars.KeyValueNames {
			out.WriteString(k.String())
			out.WriteString(": ")
			out.WriteString(v.String())
			if i != len(vars.KeyValueNames)-1 {
				out.WriteString(", ")
			}
			i++
		}
	}
	if vars.IsListDestructor {
		out.WriteByte(']')
	} else if vars.IsMapDestructor {
		out.WriteByte('}')
	}
	out.WriteByte(' ')
	out.WriteString(vars.AssignmentToken.Literal)
	out.WriteByte(' ')

	if vars.Value != nil {
		out.WriteString(vars.Value.String())
	}

	return out.String()
}

// ValStatement is the node for val statements
type ValStatement struct {
	Token         token.Token // Token == token.VAL
	KeyValueNames map[Expression]*Identifier
	Names         []*Identifier // Names are the identifiers that Value is being binded to (or that we are desctructuring)
	Value         Expression    // Value is the expression node that is being assinged to

	IsMapDestructor  bool
	IsListDestructor bool
}

// statementNode makes val a statement
func (vals *ValStatement) statementNode() {}

// TokenLiteral returns VAL
func (vals *ValStatement) TokenLiteral() string    { return vals.Token.Literal }
func (vals *ValStatement) TokenToken() token.Token { return vals.Token }

// String returns the ValStatement node as a string
func (vals *ValStatement) String() string {
	var out bytes.Buffer

	out.WriteString(vals.TokenLiteral())
	out.WriteByte(' ')
	if vals.IsListDestructor {
		out.WriteByte('[')
	} else if vals.IsMapDestructor {
		out.WriteByte('{')
	}
	for i, name := range vals.Names {
		out.WriteString(name.String())
		if i != len(vals.Names)-1 {
			out.WriteString(", ")
		}
	}
	if vals.IsMapDestructor {
		if len(vals.KeyValueNames) > 0 {
			out.WriteString(", ")
		}
		i := 0
		for k, v := range vals.KeyValueNames {
			out.WriteString(k.String())
			out.WriteString(": ")
			out.WriteString(v.String())
			if i != len(vals.KeyValueNames)-1 {
				out.WriteString(", ")
			}
			i++
		}
	}
	if vals.IsListDestructor {
		out.WriteByte(']')
	} else if vals.IsMapDestructor {
		out.WriteByte('}')
	}
	out.WriteString(" = ")

	if vals.Value != nil {
		out.WriteString(vals.Value.String())
	}

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
func (fs *FunctionStatement) TokenLiteral() string    { return fs.Token.Literal }
func (fs *FunctionStatement) TokenToken() token.Token { return fs.Token }

// String returns a stringified version of the function statement ast node
func (fs *FunctionStatement) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fun ")
	out.WriteString(fs.Name.String())
	out.WriteByte('(')
	out.WriteString(strings.Join(params, ", "))
	out.WriteByte(')')
	out.WriteString(" { ")
	out.WriteString(fs.Body.String())
	out.WriteString(" } ")

	return out.String()
}

// ReturnStatement is the node for return statements
type ReturnStatement struct {
	Token       token.Token // Token == token.RETURN
	ReturnValue Expression  // ReturnValue is an expression node that returns a value
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns RETURN
func (rs *ReturnStatement) TokenLiteral() string    { return rs.Token.Literal }
func (rs *ReturnStatement) TokenToken() token.Token { return rs.Token }

// String returns the ReturnStatement node as a string
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral())
	out.WriteByte(' ')

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	return out.String()
}

// TryCatchStatement is the try-catch block
type TryCatchStatement struct {
	Token           token.Token     // Token == token.TRY
	TryBlock        *BlockStatement // TryBlock is a block statement containing the work to be done that may error out
	CatchIdentifier *Identifier     // CatchIdentifier is the ident to assign the error too if it occurs in the TryBlock
	CatchBlock      *BlockStatement // CatchBlock is a block statement containing the work to be done in event of an error in the TryBlock
	FinallyBlock    *BlockStatement // FinallyBlock is a block statement containing the work to be done at the end of the try-catch
}

// statementNode satisfies the statement interface
func (tcs *TryCatchStatement) statementNode() {}

// TokenLiteral is the try-catch statements token literal ie. `try`
func (tcs *TryCatchStatement) TokenLiteral() string    { return tcs.Token.Literal }
func (tcs *TryCatchStatement) TokenToken() token.Token { return tcs.Token }

// String returns a stringified version of the try-catch statement ast node
func (tcs *TryCatchStatement) String() string {
	var out bytes.Buffer

	out.WriteString("try { ")
	out.WriteString(tcs.TryBlock.ExpressionString())
	out.WriteString(" } catch (")
	out.WriteString(tcs.CatchIdentifier.Value)
	out.WriteString(") { ")
	out.WriteString(tcs.CatchBlock.ExpressionString())
	out.WriteString(" } ")
	if tcs.FinallyBlock != nil {
		out.WriteString(" finally { ")
		out.WriteString(tcs.FinallyBlock.ExpressionString())
		out.WriteString(" } ")
	}

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
func (es *ExpressionStatement) TokenLiteral() string    { return es.Token.Literal }
func (es *ExpressionStatement) TokenToken() token.Token { return es.Token }

// String will return the string version of the expression statement
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement is the ast node for block statements
type BlockStatement struct {
	Token         token.Token // Token == {
	Statements    []Statement // Statements is the list of statements in the block
	HelpStrTokens []string
}

// statementNode satisifes the statement interface
func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns the { token
func (bs *BlockStatement) TokenLiteral() string    { return bs.Token.Literal }
func (bs *BlockStatement) TokenToken() token.Token { return bs.Token }

// String returns the string representation of the block statement
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for i, s := range bs.Statements {
		out.WriteString(s.String())
		if i != len(bs.Statements)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

// ExpressionString returns the BlockStatement string without any newlines to be used by other AST Nodes
func (bs *BlockStatement) ExpressionString() string {
	var out bytes.Buffer

	for i, s := range bs.Statements {
		out.WriteString(s.String())
		// We use the semicolon here to make it clear how its being broken up (all other areas should not be using the ;)
		out.WriteByte(';')
		if i+1 != len(bs.Statements) {
			out.WriteString(", ")
		}
	}
	return out.String()
}

// ImportStatement is the representation of the map literal ast node
type ImportStatement struct {
	Token          token.Token   // Token == import
	Path           *Identifier   // Path is the import's path which refers to a file
	IdentsToImport []*Identifier // IdentsToImport refers to all the idents that should be imported to the current module
	Alias          *Identifier   // Alias is what an import statement should be aliased to in the current module
	ImportAll      bool          // ImportAll is a boolean to determine if a * was used in a from import statement
}

// statementNode satisfies the statement interface
func (is *ImportStatement) statementNode() {}

// TokenLiteral returns the import token as a string
func (is *ImportStatement) TokenLiteral() string    { return is.Token.Literal }
func (is *ImportStatement) TokenToken() token.Token { return is.Token }

// String returns the string representation of the map literal ast node
func (is *ImportStatement) String() string {
	if len(is.IdentsToImport) == 0 {
		if is.ImportAll {
			return fmt.Sprintf("from %s import *", is.Path)
		}
		if is.Alias == nil {
			return fmt.Sprintf("%s %s", is.Token.Literal, is.Path)
		}
		return fmt.Sprintf("%s %s as %s", is.Token.Literal, is.Path, is.Alias.Value)
	} else {
		toStrs := []string{}
		for _, e := range is.IdentsToImport {
			toStrs = append(toStrs, e.Value)
		}
		return fmt.Sprintf("from %s import [%s]", is.Path, strings.Join(toStrs, ", "))
	}
}

type BreakStatement struct {
	Token token.Token // Token == break
}

// statementNode satisfies the statement interface
func (bks *BreakStatement) statementNode() {}

// TokenLiteral returns the break token as a string
func (bks *BreakStatement) TokenLiteral() string    { return bks.Token.Literal }
func (bks *BreakStatement) TokenToken() token.Token { return bks.Token }

// String returns the string representation of the break literal ast node
func (bks *BreakStatement) String() string {
	return bks.TokenLiteral()
}

type ContinueStatement struct {
	Token token.Token // Token == continue
}

// statementNode satisfies the statement interface
func (cs *ContinueStatement) statementNode() {}

// TokenLiteral returns the continue token as a string
func (cs *ContinueStatement) TokenLiteral() string    { return cs.Token.Literal }
func (cs *ContinueStatement) TokenToken() token.Token { return cs.Token }

// String returns the string representation of the continue literal ast node
func (cs *ContinueStatement) String() string {
	return cs.TokenLiteral()
}

// ForStatement is the for loop ast node
type ForStatement struct {
	Token       token.Token     // token == for
	Condition   Expression      // Condition is the condition to test whether the loop should continue
	Consequence *BlockStatement // Consequence contains a block of statements that happen if the condition is true
	// for (var i = 0; i < 10; i += 1)
	UsesVar     bool          // UsesVar if the for expression condition starts with 'var'
	Initializer *VarStatement // initializer to be used for the for expression (var x = 0)
	PostExp     Expression    // PostExp expression to run after the loop
}

// statementNode satisfies the statement interface
func (fe *ForStatement) statementNode() {}

// TokenLiteral returns the for token
func (fe *ForStatement) TokenLiteral() string    { return fe.Token.Literal }
func (fe *ForStatement) TokenToken() token.Token { return fe.Token }

// String returns the string representation of the for expression ast node
func (fe *ForStatement) String() string {
	var out bytes.Buffer

	out.WriteString("for (")
	if !fe.UsesVar {
		out.WriteString(fe.Condition.String())
	} else {
		out.WriteString(fe.Initializer.String())
		out.WriteString("; ")
		out.WriteString(fe.Condition.String())
		out.WriteString("; ")
		out.WriteString(fe.PostExp.String())
	}
	out.WriteString(") { ")
	out.WriteString(fe.Consequence.ExpressionString())
	out.WriteString(" } ")
	return out.String()
}
