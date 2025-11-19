package compiler

import (
	"blue/ast"
	"fmt"
	"strings"
)

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	FreeScope    SymbolScope = "FREE"
	BuiltinScope SymbolScope = "BUILTIN"
)

type Symbol struct {
	Name      string
	Scope     SymbolScope
	Index     int
	Immutable bool

	// Only used for builtins at the moment
	BuiltinModuleIndex int

	// For Functions (to allow consistent calling with default args)
	Parameters           []*ast.Identifier
	ParameterExpressions []ast.Expression
}

func (s Symbol) Equal(other Symbol) bool {
	return s.Name == other.Name && s.Scope == other.Scope && s.Index == other.Index && s.Immutable == other.Immutable && s.BuiltinModuleIndex == other.BuiltinModuleIndex
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
	FreeSymbols    []Symbol
	BlockSymbols   [][]Symbol

	Outer *SymbolTable
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	block := [][]Symbol{}
	return &SymbolTable{store: s, FreeSymbols: free, BlockSymbols: block}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (s *SymbolTable) Define(name string, isImmutable bool, blockNestLevel int) Symbol {
	return s.defineActual(name, isImmutable, blockNestLevel, nil, nil)
}

func (s *SymbolTable) DefineFun(name string, isImmutable bool, blockNestLevel int, parameters []*ast.Identifier, parameterExpressions []ast.Expression) Symbol {
	return s.defineActual(name, isImmutable, blockNestLevel, parameters, parameterExpressions)
}

func (s *SymbolTable) defineActual(name string, isImmutable bool, blockNestLevel int, parameters []*ast.Identifier, parameterExpressions []ast.Expression) Symbol {
	symbol := Symbol{Name: name, Index: s.numDefinitions, Immutable: isImmutable, Parameters: parameters, ParameterExpressions: parameterExpressions}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	if blockNestLevel != -1 {
		numToAppend := blockNestLevel - len(s.BlockSymbols)
		for i := 0; i <= numToAppend; i++ {
			s.BlockSymbols = append(s.BlockSymbols, []Symbol{})
		}
		s.BlockSymbols[blockNestLevel] = append(s.BlockSymbols[blockNestLevel], symbol)
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(index int, name string, builtinModuleIndex int) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope, BuiltinModuleIndex: builtinModuleIndex}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) RemoveBuiltin(name string) {
	delete(s.store, name)
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		obj, ok := s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}
		if obj.Scope == GlobalScope || obj.Scope == BuiltinScope {
			return obj, ok
		}
		free := s.defineFree(obj)
		return free, true
	}
	return obj, ok
}

func (s *SymbolTable) defineFree(original Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, original)
	symbol := Symbol{Name: original.Name, Index: len(s.FreeSymbols) - 1}
	symbol.Scope = FreeScope
	s.store[original.Name] = symbol
	return symbol
}

func (s *SymbolTable) String() string {
	var sb strings.Builder
	for k, v := range s.store {
		sb.WriteString(fmt.Sprintf("%s %#+v\n", k, v))
	}
	outer := s.Outer
	if outer != nil {
		for outer != nil {
			for k, v := range outer.store {
				sb.WriteString(fmt.Sprintf("%s %#+v\n", k, v))
			}
			outer = outer.Outer
		}
	}
	return sb.String()
}

func (s *SymbolTable) UpdateName(ogName, newName string) error {
	orig, exists := s.store[ogName]
	if !exists {
		return fmt.Errorf("cant find name %s in symbol table", ogName)
	}
	_, newExistsAlready := s.store[newName]
	if newExistsAlready {
		return fmt.Errorf("name %s already exists in symbol table", newName)
	}
	delete(s.store, ogName)
	s.store[newName] = orig
	return nil
}
