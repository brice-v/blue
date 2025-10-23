package compiler

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
	symbol := Symbol{Name: name, Index: s.numDefinitions, Immutable: isImmutable}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	if blockNestLevel != -1 {
		if len(s.BlockSymbols) <= blockNestLevel {
			s.BlockSymbols = append(s.BlockSymbols, []Symbol{})
		}
		s.BlockSymbols[blockNestLevel] = append(s.BlockSymbols[blockNestLevel], symbol)
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	s.store[name] = symbol
	return symbol
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
