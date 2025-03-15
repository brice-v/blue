package object

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

type Scope struct {
	store          map[string]Object
	immutableStore map[string]struct{}
	lock           sync.RWMutex
}

func (s *Scope) Set(name string, val Object) {
	s.lock.Lock()
	s.store[name] = val
	s.lock.Unlock()
}

func (s *Scope) ImmutableSet(name string) {
	s.lock.Lock()
	s.immutableStore[name] = struct{}{}
	s.lock.Unlock()
}

func (s *Scope) Get(name string) (Object, bool) {
	s.lock.RLock()
	val, ok := s.store[name]
	s.lock.RUnlock()
	return val, ok
}

func (s *Scope) Clone() *Scope {
	s.lock.Lock()
	scope := &Scope{
		store:          copyMap(s.store),
		immutableStore: copyMap(s.immutableStore),
		lock:           sync.RWMutex{},
	}
	s.lock.Unlock()
	return scope
}

func NewScope() *Scope {
	return &Scope{
		store:          make(map[string]Object),
		immutableStore: make(map[string]struct{}),
		lock:           sync.RWMutex{},
	}
}

type Environment2 struct {
	scopes       []*Scope
	framePointer int
	coreEnv      *Environment2

	// publicFunctionHelpStore *OrderedMap2[string, string]
}

func NewEnvironment2() *Environment2 {
	return &Environment2{
		scopes:       []*Scope{NewScope()},
		framePointer: 0,
		coreEnv:      nil,
		// publicFunctionHelpStore: &OrderedMap2[string, string]{
		// 	store: make(map[string]string),
		// 	Keys:  []string{},
		// },
	}
}

func NewEnvironment2WithSize(size int) *Environment2 {
	e := &Environment2{
		scopes:       make([]*Scope, 0, size),
		framePointer: 0,
		// publicFunctionHelpStore: &OrderedMap2[string, string]{
		// 	store: make(map[string]string),
		// 	Keys:  []string{},
		// },
	}
	// TODO: Need go1.22 for range size
	// for i := range size {
	for i := 0; i < size; i += 1 {
		e.scopes = append(e.scopes, NewScope())
	}
	return e
}

func (e *Environment2) PushScope() {
	e.scopes = append(e.scopes, NewScope())
	e.framePointer++
}

func (e *Environment2) PushExistingScope(scope *Scope) {
	e.scopes = append(e.scopes, scope)
	e.framePointer++
}

func (e *Environment2) PopScope() {
	e.scopes = e.scopes[:e.framePointer]
	e.framePointer--
}

func (e *Environment2) Set(name string, val Object) {
	scope := e.scopes[e.framePointer]
	scope.lock.Lock()
	scope.store[name] = val
	scope.lock.Unlock()
	// e.setPublicFunctionHelpString(name, val)
}

func (e *Environment2) ImmutableSet(name string) {
	scope := e.scopes[e.framePointer]
	scope.lock.Lock()
	scope.immutableStore[name] = struct{}{}
	scope.lock.Unlock()
}

func (e *Environment2) IsImmutable(name string) bool {
	for _, scope := range e.scopes {
		scope.lock.RLock()
		defer scope.lock.RUnlock()
		if _, ok := scope.immutableStore[name]; ok {
			return ok
		}
	}
	return false
}

func (e *Environment2) Get(name string) (Object, bool) {
	for i := e.framePointer; i >= 0; i -= 1 {
		scope := e.scopes[i]
		scope.lock.RLock()
		val, ok := scope.store[name]
		if ok {
			scope.lock.RUnlock()
			return val, ok
		}
		scope.lock.RUnlock()
	}
	if e.coreEnv != nil {
		return e.coreEnv.Get(name)
	}
	return nil, false
}

func (e *Environment2) GetAll() map[string]Object {
	m := make(map[string]Object)
	for _, scope := range e.scopes {
		scope.lock.RLock()
		for k, v := range scope.store {
			m[k] = v
		}
		scope.lock.RUnlock()
	}
	return m
}

func (e *Environment2) RemoveIdentifier(name string) {
	for i := e.framePointer; i >= 0; i -= 1 {
		scope := e.scopes[i]
		delete(scope.store, name)
		delete(scope.immutableStore, name)
	}
}

func copyMap[T any](m map[string]T) map[string]T {
	newMap := make(map[string]T, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func (e *Environment2) Clone() *Environment2 {
	newEnv := NewEnvironment2WithSize(len(e.scopes))
	for i, s := range e.scopes {
		s.lock.RLock()
		newEnv.scopes[i].store = copyMap(s.store)
		newEnv.scopes[i].immutableStore = copyMap(s.immutableStore)
		s.lock.RUnlock()
	}
	// for _, k := range e.publicFunctionHelpStore.Keys {
	// 	v, _ := e.publicFunctionHelpStore.Get(k)
	// 	newEnv.publicFunctionHelpStore.Set(k, v)
	// }
	return newEnv
}

func (e *Environment2) ClonePriorScopeIntoNewScope() *Scope {
	newScope := NewScope()
	priorScope := e.scopes[len(e.scopes)-1]
	newScope.store = copyMap(priorScope.store)
	newScope.immutableStore = copyMap(priorScope.immutableStore)
	return newScope
}

func (e *Environment2) CloneAllIntoNewScope() *Scope {
	newScope := NewScope()
	for _, s := range e.scopes {
		for k, v := range s.store {
			newScope.store[k] = v
		}
		for k, v := range s.immutableStore {
			newScope.immutableStore[k] = v
		}
	}
	return newScope
}

func (e *Environment2) SetCore(coreEnv *Environment2) {
	e.coreEnv = coreEnv
}

// String is for debugging, do not call in normal use
func (e *Environment2) String() string {
	var out bytes.Buffer
	out.WriteString("Environment{\n")
	for i := e.framePointer; i >= 0; i -= 1 {
		scope := e.scopes[i]
		out.WriteString(fmt.Sprintf("\tscope %d: \n", i))
		for k, v := range scope.store {
			out.WriteString(fmt.Sprintf("\t\tname: %s, value: %s\n", k, strings.ReplaceAll(v.Inspect(), "\n", "")))
		}
	}
	out.WriteByte('}')
	return out.String()
}

// Functions for Public Function Help Strings

// func (e *Environment2) GetPublicFunctionHelpString() string {
// 	var out bytes.Buffer
// 	lengthOfLargestString := 0
// 	for _, k := range e.publicFunctionHelpStore.Keys {
// 		l := utf8.RuneCountInString(k)
// 		if l > lengthOfLargestString {
// 			lengthOfLargestString = l
// 		}
// 	}
// 	lengthOfLargestString++
// 	for _, k := range e.publicFunctionHelpStore.Keys {
// 		if v, ok := e.publicFunctionHelpStore.Get(k); ok {
// 			vSplit := strings.Split(v, "\ntype(")[0]
// 			// remove the trailing \n
// 			vSplit = vSplit[:len(vSplit)-1]
// 			vSplitFurther := strings.Split(vSplit, "\n")
// 			for i, partStr := range vSplitFurther {
// 				if i == 0 {
// 					initialPadLen := lengthOfLargestString - utf8.RuneCountInString(k)
// 					initialPad := strings.Repeat(" ", initialPadLen)
// 					consts.DisableColorIfNoColorEnvVarSet()
// 					green := color.FgGreen.Render
// 					bold := color.Bold.Render
// 					out.WriteString(fmt.Sprintf("\n%s%s| %s", bold(green(k)), initialPad, partStr))
// 				} else {
// 					pad := strings.Repeat(" ", lengthOfLargestString+2)
// 					nl := "\n"
// 					if i == len(vSplitFurther)-1 {
// 						nl = ""
// 					}
// 					prefixNl := ""
// 					if i == 1 {
// 						prefixNl = "\n"
// 					} else {
// 						prefixNl = ""
// 					}
// 					out.WriteString(fmt.Sprintf("%s%s %s%s", prefixNl, pad, partStr, nl))
// 				}
// 			}
// 		}
// 	}
// 	return out.String()
// }

// func (e *Environment2) setPublicFunctionHelpString(name string, val Object) {
// 	if val != nil && (val.Type() == "FUNCTION" && !strings.HasPrefix(name, "_")) {
// 		coreIgnored := strings.HasPrefix(val.Help(), "core:ignore")
// 		if !coreIgnored {
// 			ogHelp := val.Help()
// 			if strings.HasPrefix(ogHelp, "core:") {
// 				coreHelpStr := e.getFunctionHelpString(ogHelp, "core:")
// 				e.setFunctionHelpStr(name, coreHelpStr)
// 			} else if strings.HasPrefix(ogHelp, "std:") {
// 				stdHelpStr := e.getFunctionHelpString(ogHelp, "std:")
// 				e.setFunctionHelpStr(name, stdHelpStr)
// 			} else {
// 				e.publicFunctionHelpStore.Set(name, ogHelp)
// 			}
// 		}
// 	}
// }

// func (e *Environment2) setFunctionHelpStr(name, newHelpStr string) {
// 	if val, ok := e.Get(name); ok {
// 		if fun, ok := val.(*Function); ok {
// 			fun.HelpStr = newHelpStr
// 			e.Set(name, fun)
// 		}
// 	}
// }

// func (e *Environment2) getFunctionHelpString(origHelp, prefix string) string {
// 	parts := strings.Split(origHelp, "\n")
// 	thingsToGet := strings.Split(strings.Split(parts[0], prefix)[1], ",")
// 	var out bytes.Buffer
// 	l := len(thingsToGet)
// 	for j, v := range thingsToGet {
// 		if v == "this" {
// 			indexForTypeFun := 0
// 			for i, e := range parts {
// 				if strings.HasPrefix(e, "type(") {
// 					indexForTypeFun = i - 1
// 					break
// 				}
// 			}
// 			newHelp := strings.Join(parts[1:indexForTypeFun][:], "\n")
// 			out.WriteString(newHelp)
// 		} else {
// 			if val, ok := e.Get(v); ok {
// 				out.WriteString(val.Help())
// 			}
// 		}
// 		if l > 1 && j != j-1 {
// 			out.WriteByte('\n')
// 		}
// 	}
// 	return out.String()
// }
