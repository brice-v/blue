package compiler

import (
	"blue/ast"
	"bytes"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	FreeScope    SymbolScope = "FREE"
	BuiltinScope SymbolScope = "BUILTIN"

	SpecialFunctionScope SymbolScope = "SPECIAL"
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

	HelpStr string
}

func (s Symbol) Equal(other Symbol) bool {
	return s.Name == other.Name && s.Scope == other.Scope && s.Index == other.Index && s.Immutable == other.Immutable && s.BuiltinModuleIndex == other.BuiltinModuleIndex
}

type SymbolTable struct {
	store                     map[string]Symbol
	specialStore              map[SpecialScopeKey]Symbol
	specialStoreParamIndexMap map[string][]int
	numDefinitions            int
	specialDefinitions        int
	FreeSymbols               []Symbol

	BlockNestLevel int

	Outer *SymbolTable
}

func (st *SymbolTable) NumLocals() int {
	return st.numDefinitions + st.specialDefinitions
}

type SpecialScopeKey struct {
	ScopeIndex int
	ParamIndex int
	Name       string
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	ss := make(map[SpecialScopeKey]Symbol)
	ssim := make(map[string][]int)
	free := []Symbol{}
	return &SymbolTable{store: s, specialStore: ss, specialStoreParamIndexMap: ssim, FreeSymbols: free, BlockNestLevel: -1}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (s *SymbolTable) Define(name string, isImmutable bool) Symbol {
	return s.defineActual(name, isImmutable, nil, nil, "")
}

func (s *SymbolTable) DefineFun(name string, isImmutable bool, parameters []*ast.Identifier, parameterExpressions []ast.Expression, helpStr string) Symbol {
	return s.defineActual(name, isImmutable, parameters, parameterExpressions, helpStr)
}

func (s *SymbolTable) defineActual(name string, isImmutable bool, parameters []*ast.Identifier, parameterExpressions []ast.Expression, helpStr string) Symbol {
	if helpStr != "" {
		helpStr = s.getHelpInPublicFunctionHelpStore(name, helpStr)
	}
	symbol := Symbol{Name: name, Index: s.numDefinitions, Immutable: isImmutable, Parameters: parameters, ParameterExpressions: parameterExpressions, HelpStr: helpStr}
	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}
	if s.BlockNestLevel != -1 {
		newName := fmt.Sprintf("%d:%s", s.BlockNestLevel, name)
		s.store[newName] = symbol
	} else {
		s.store[name] = symbol
	}
	s.numDefinitions++
	return symbol
}

func (s *SymbolTable) defineSpecial(name string, scopeIndex, paramIndex, listIndex int) Symbol {
	// Using BuiltinModuleIndex for the parameter index
	symbol := Symbol{Name: name, Index: listIndex, BuiltinModuleIndex: paramIndex, Immutable: true, Scope: SpecialFunctionScope}
	s.specialStore[SpecialScopeKey{Name: name, ScopeIndex: scopeIndex, ParamIndex: paramIndex}] = symbol
	s.specialStoreParamIndexMap[name] = append(s.specialStoreParamIndexMap[name], paramIndex)
	s.specialDefinitions++
	return symbol
}

func (s *SymbolTable) DefineBuiltin(index int, name string, builtinModuleIndex int, helpStr string) Symbol {
	symbol := Symbol{Name: name, Index: index, Scope: BuiltinScope, BuiltinModuleIndex: builtinModuleIndex, HelpStr: helpStr}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) Remove(name string) {
	delete(s.store, name)
}

func (s *SymbolTable) ResolveSpecial(name string, scopeIndex int) (Symbol, bool, bool) {
	indexMap, ok := s.specialStoreParamIndexMap[name]
	if ok {
		if len(indexMap) > 1 {
			return emptySym, true, true
		}
		for _, index := range indexMap {
			if symbol, ok := s.specialStore[SpecialScopeKey{Name: name, ScopeIndex: scopeIndex, ParamIndex: index}]; ok {
				return symbol, ok, false
			}
		}
	}
	return emptySym, false, false
}

func (s *SymbolTable) resolveFromCurrentBlockNestLevel(name string) (Symbol, bool) {
	for i := s.BlockNestLevel; i >= 0; i-- {
		newName := fmt.Sprintf("%d:%s", i, name)
		if obj, ok := s.store[newName]; ok {
			return obj, ok
		}
	}
	obj, ok := s.store[name]
	return obj, ok
}

func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.resolveFromCurrentBlockNestLevel(name)
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

func (s *SymbolTable) LookupInCurrentBlockLevel(name string) (Symbol, bool) {
	if s.BlockNestLevel == -1 {
		sym, ok := s.store[name]
		return sym, ok
	}
	prefixedName := fmt.Sprintf("%d:%s", s.BlockNestLevel, name)
	sym, ok := s.store[prefixedName]
	return sym, ok
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
		fmt.Fprintf(&sb, "%s %#+v\n", k, v)
	}
	outer := s.Outer
	if outer != nil {
		for outer != nil {
			for k, v := range outer.store {
				fmt.Fprintf(&sb, "%s %#+v\n", k, v)
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

// Help String related

func (s *SymbolTable) getLengthOfLargestStringAndOrderedKeys(modName string) (int, []string) {
	lengthOfLargestString := 0
	keys := []string{}
	prefix := modName + "."
	for key, value := range s.store {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		key = strings.TrimPrefix(key, prefix)
		if value.HelpStr != "" {
			keys = append(keys, key)
			l := utf8.RuneCountInString(key)
			if l > lengthOfLargestString {
				lengthOfLargestString = l
			}
		}
	}
	lengthOfLargestString++
	slices.Sort(keys)
	return lengthOfLargestString, keys
}

func (s *SymbolTable) GetOrderedPublicFunctionHelpString(modName string) string {
	var out bytes.Buffer
	lengthOfLargestString, orderedKeys := s.getLengthOfLargestStringAndOrderedKeys(modName)
	prefix := modName + "."
	for _, k := range orderedKeys {
		keyToUse := k
		if modName != "" {
			keyToUse = prefix + k
		}
		value := s.store[keyToUse]
		v := value.HelpStr
		// if v == "" {
		// 	continue
		// }
		vSplit := strings.Split(v, "\ntype(")[0]
		// remove the trailing \n
		vSplit = vSplit[:len(vSplit)-1]
		vSplitFurther := strings.Split(vSplit, "\n")
		for i, partStr := range vSplitFurther {
			if i == 0 {
				initialPadLen := lengthOfLargestString - utf8.RuneCountInString(k)
				initialPad := strings.Repeat(" ", initialPadLen)
				fmt.Fprintf(&out, "\n%s%s| %s", k, initialPad, partStr)
				continue
			}
			pad := strings.Repeat(" ", lengthOfLargestString+2)
			nl := "\n"
			if i == len(vSplitFurther)-1 {
				nl = ""
			}
			prefixNl := ""
			if i == 1 {
				prefixNl = "\n"
			} else {
				prefixNl = ""
			}
			fmt.Fprintf(&out, "%s%s %s%s", prefixNl, pad, partStr, nl)
		}
	}
	return out.String()
}

func (s *SymbolTable) getFunctionHelpString(origHelp, prefix string) string {
	parts := strings.Split(origHelp, "\n")
	thingsToGet := strings.Split(strings.Split(parts[0], prefix)[1], ",")
	var out bytes.Buffer
	l := len(thingsToGet)
	for j, v := range thingsToGet {
		if v == "this" {
			indexForTypeFun := 0
			for i, e := range parts {
				if strings.HasPrefix(e, "type(") {
					indexForTypeFun = i - 1
					break
				}
			}
			newHelp := strings.Join(parts[1:indexForTypeFun][:], "\n")
			out.WriteString(newHelp)
		} else {
			vToUse := v
			if strings.HasPrefix(v, "__") {
				vToUse = vToUse[1:]
			}
			if val, ok := s.store[vToUse]; ok {
				out.WriteString(val.HelpStr)
			}
		}
		if l > 1 && j != j-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func (s *SymbolTable) getHelpInPublicFunctionHelpStore(name, ogHelp string) string {
	var help string
	if !strings.HasPrefix(name, "_") {
		if !strings.HasPrefix(ogHelp, "core:ignore") {
			if strings.HasPrefix(ogHelp, "core:") {
				help = s.getFunctionHelpString(ogHelp, "core:")
			} else if strings.HasPrefix(ogHelp, "std:") {
				help = s.getFunctionHelpString(ogHelp, "std:")
			} else {
				help = ogHelp
			}
		}
	}
	return help
}

// End of HelpStr related
