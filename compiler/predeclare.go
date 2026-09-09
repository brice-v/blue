package compiler

import (
	"blue/ast"
)

// predeclareDeclarations reserves a slot for every name that a statement in the
// given list binds at the current block nest level. It runs before any of those
// statements get compiled so that code appearing before a declaration resolves to
// the same slot the declaration ends up binding to, which is what allows
// functions and variables to be defined after they are used.
//
// Only the statements making up the list itself are looked at: entering a
// function starts a whole new scope (and therefore its own list), and anything
// declared inside a nested block belongs to that block instead.
func (c *Compiler) predeclareDeclarations(statements []ast.Statement) {
	for _, statement := range statements {
		switch node := statement.(type) {
		case *ast.FunctionStatement:
			if node.Name == nil {
				continue
			}
			c.predeclareName(node.Name.Value, true)
		case *ast.VarStatement:
			c.predeclareBoundNames(node, false)
		case *ast.ValStatement:
			c.predeclareBoundNames(node, true)
		}
	}
}

func (c *Compiler) predeclareBoundNames(node ast.VarValStatement, immutable bool) {
	for _, name := range node.VVNames() {
		c.predeclareName(name.Value, immutable)
	}
	for _, name := range node.VVKeyValueNames() {
		c.predeclareName(name.Value, immutable)
	}
}

func (c *Compiler) predeclareName(name string, immutable bool) {
	c.symbolTable.Predeclare(c.getName(name), immutable)
}
