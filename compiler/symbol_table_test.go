package compiler

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
	}
	global := NewSymbolTable()
	a := global.Define("a", false, false)
	if a != expected["a"] {
		t.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}
	b := global.Define("b", false, false)
	if b != expected["b"] {
		t.Errorf("expected b=%+v, got=%+v", expected["b"], b)
	}
	firstLocal := NewEnclosedSymbolTable(global)
	c := firstLocal.Define("c", false, false)
	if c != expected["c"] {
		t.Errorf("expected c=%+v, got=%+v", expected["c"], c)
	}
	d := firstLocal.Define("d", false, false)
	if d != expected["d"] {
		t.Errorf("expected d=%+v, got=%+v", expected["d"], d)
	}
	secondLocal := NewEnclosedSymbolTable(firstLocal)
	e := secondLocal.Define("e", false, false)
	if e != expected["e"] {
		t.Errorf("expected e=%+v, got=%+v", expected["e"], e)
	}
	f := secondLocal.Define("f", false, false)
	if f != expected["f"] {
		t.Errorf("expected f=%+v, got=%+v", expected["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a", false, false)
	global.Define("b", false, false)
	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
	}
	for _, sym := range expected {
		result, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a", false, false)
	global.Define("b", false, false)
	local := NewEnclosedSymbolTable(global)
	local.Define("c", false, false)
	local.Define("d", false, false)
	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}
	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
		}
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a", false, false)
	global.Define("b", false, false)
	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c", false, false)
	firstLocal.Define("d", false, false)
	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e", false, false)
	secondLocal.Define("f", false, false)
	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}
	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
			}
		}
	}
}

func TestResolveFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a", false, false)
	global.Define("b", false, false)
	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c", false, false)
	firstLocal.Define("d", false, false)
	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e", false, false)
	secondLocal.Define("f", false, false)
	tests := []struct {
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]Symbol{},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: FreeScope, Index: 0},
				{Name: "d", Scope: FreeScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
			[]Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
	}
	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
			}
		}
		if len(tt.table.FreeSymbols) != len(tt.expectedFreeSymbols) {
			t.Errorf("wrong number of free symbols. got=%d, want=%d", len(tt.table.FreeSymbols), len(tt.expectedFreeSymbols))
			continue
		}
		for i, sym := range tt.expectedFreeSymbols {
			result := tt.table.FreeSymbols[i]
			if result != sym {
				t.Errorf("wrong free symbol. got=%+v, want=%+v", result, sym)
			}
		}
	}
}

func TestResolveUnresolvableFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a", false, false)
	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c", false, false)
	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e", false, false)
	secondLocal.Define("f", false, false)
	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "c", Scope: FreeScope, Index: 0},
		{Name: "e", Scope: LocalScope, Index: 0},
		{Name: "f", Scope: LocalScope, Index: 1},
	}
	for _, sym := range expected {
		result, ok := secondLocal.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
		}
	}
	expectedUnresolvable := []string{
		"b",
		"d",
	}
	for _, name := range expectedUnresolvable {
		_, ok := secondLocal.Resolve(name)
		if ok {
			t.Errorf("name %s resolved, but was expected not to", name)
		}
	}
}

func TestDefineResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)
	expected := []Symbol{
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "c", Scope: BuiltinScope, Index: 1},
		{Name: "e", Scope: BuiltinScope, Index: 2},
		{Name: "f", Scope: BuiltinScope, Index: 3},
	}
	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}
	for _, table := range []*SymbolTable{global, firstLocal, secondLocal} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
			}
		}
	}
}
