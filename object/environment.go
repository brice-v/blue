package object

import (
	"blue/consts"
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gookit/color"
)

// NewEnclosedEnvironment supports adding an environment to an exisiting
// environment.  This allows closures and proper binding within functions
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment(outer.coreEnv)
	env.outer = outer
	return env
}

// NewEnvironment returns a new environment
func NewEnvironmentWithoutCore() *Environment {
	s := &ConcurrentMap[string, Object]{
		kv: make(map[string]Object),
	}
	is := &ConcurrentMap[string, struct{}]{
		kv: make(map[string]struct{}),
	}
	pfhs := &OrderedMap2[string, string]{
		store: make(map[string]string),
		Keys:  []string{},
	}
	return &Environment{store: s, immutableStore: is, publicFunctionHelpStore: pfhs}
}

// NewEnvironment returns a new environment
func NewEnvironment(coreEnv *Environment) *Environment {
	s := &ConcurrentMap[string, Object]{
		kv: make(map[string]Object),
	}
	is := &ConcurrentMap[string, struct{}]{
		kv: make(map[string]struct{}),
	}
	pfhs := &OrderedMap2[string, string]{
		store: make(map[string]string),
		Keys:  []string{},
	}
	return &Environment{store: s, immutableStore: is, publicFunctionHelpStore: pfhs, coreEnv: coreEnv}
}

// Environment is a map of strings to `Object`s
type Environment struct {
	store          *ConcurrentMap[string, Object]
	immutableStore *ConcurrentMap[string, struct{}]

	publicFunctionHelpStore *OrderedMap2[string, string]

	outer   *Environment
	coreEnv *Environment
}

func (e *Environment) SetCore(coreEnv *Environment) {
	e.coreEnv = coreEnv
}

// Clone will do a deep copy of the environment object and return a new environment
// it will not write to the outer env, but it will read from it
func (e *Environment) Clone() *Environment {
	newEnv := NewEnvironment(e.coreEnv)
	for k, v := range e.store.kv {
		newEnv.store.Put(k, v)
	}
	for k, v := range e.immutableStore.kv {
		newEnv.immutableStore.Put(k, v)
	}
	outer := e.outer
	for outer != nil {
		for k, v := range outer.store.kv {
			newEnv.store.Put(k, v)
		}
		for k, v := range outer.immutableStore.kv {
			newEnv.immutableStore.Put(k, v)
		}
		outer = outer.outer
	}
	for _, k := range e.publicFunctionHelpStore.Keys {
		v, _ := e.publicFunctionHelpStore.Get(k)
		newEnv.publicFunctionHelpStore.Set(k, v)
	}
	return newEnv
}

// func (e *Environment) String() string {
// 	var out bytes.Buffer
// 	out.WriteString("Environment{\n\tstore:\n")
// 	for s, elem := range e.store.kv {
// 		out.WriteString(fmt.Sprintf("\t\t%q -> %s\n", s, elem.Type()))
// 	}
// 	out.WriteString("\n\timmuatebleStore:\n")
// 	for s := range e.immutableStore.kv {
// 		out.WriteString(fmt.Sprintf("\t\t%qs\n", s))
// 	}
// 	out.WriteString("}")
// 	return out.String()
// }

// Get returns the object from the environment store
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store.Get(name)
	if !ok && e.coreEnv != nil {
		obj, ok = e.coreEnv.Get(name)
		if !ok && e.outer != nil {
			obj, ok = e.outer.Get(name)
		}
	}
	return obj, ok
}

func (e *Environment) GetAll() map[string]Object {
	// TODO: do we need to get all from outer as well?
	return e.store.GetAll()
}

func (e *Environment) getFunctionHelpString(origHelp, prefix string) string {
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
			if val, ok := e.Get(v); ok {
				out.WriteString(val.Help())
			}
		}
		if l > 1 && j != j-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func (e *Environment) SetFunctionHelpStr(name, newHelpStr string) {
	if val, ok := e.Get(name); ok {
		if fun, ok := val.(*Function); ok {
			fun.HelpStr = newHelpStr
			e.Set(name, fun)
		}
	}
}

// Set puts a new object into the environment
func (e *Environment) Set(name string, val Object) Object {
	e.store.Put(name, val)
	// We do store nil values so those can be skipped entirely for pfhs
	if val != nil && (val.Type() == "FUNCTION" && !strings.HasPrefix(name, "_")) {
		coreIgnored := strings.HasPrefix(val.Help(), "core:ignore")
		if !coreIgnored {
			ogHelp := val.Help()
			if strings.HasPrefix(ogHelp, "core:") {
				coreHelpStr := e.getFunctionHelpString(ogHelp, "core:")
				e.SetFunctionHelpStr(name, coreHelpStr)
			} else if strings.HasPrefix(ogHelp, "std:") {
				stdHelpStr := e.getFunctionHelpString(ogHelp, "std:")
				e.SetFunctionHelpStr(name, stdHelpStr)
			} else {
				e.publicFunctionHelpStore.Set(name, ogHelp)
			}
		}
	}
	return val
}

// ImmutableSet puts the name of the identifier in a map and sets it as true
func (e *Environment) ImmutableSet(name string) {
	e.immutableStore.Put(name, struct{}{})
}

// IsImmutable checks if the give identifier name is in the immutable map
func (e *Environment) IsImmutable(name string) bool {
	_, ok := e.immutableStore.Get(name)
	if !ok && e.outer != nil {
		ok = e.outer.IsImmutable(name)
		return ok
	}
	return ok
}

// RemoveIdentifier removes a key from the environment
// this is used in for loops to remove temporary variables
func (e *Environment) RemoveIdentifier(name string) {
	e.store.Remove(name)
	e.immutableStore.Remove(name)
	e.publicFunctionHelpStore.Delete(name)
}

func (e *Environment) GetPublicFunctionHelpString() string {
	var out bytes.Buffer
	lengthOfLargestString := 0
	for _, k := range e.publicFunctionHelpStore.Keys {
		l := utf8.RuneCountInString(k)
		if l > lengthOfLargestString {
			lengthOfLargestString = l
		}
	}
	lengthOfLargestString++
	for _, k := range e.publicFunctionHelpStore.Keys {
		if v, ok := e.publicFunctionHelpStore.Get(k); ok {
			vSplit := strings.Split(v, "\ntype(")[0]
			// remove the trailing \n
			vSplit = vSplit[:len(vSplit)-1]
			vSplitFurther := strings.Split(vSplit, "\n")
			for i, partStr := range vSplitFurther {
				if i == 0 {
					initialPadLen := lengthOfLargestString - utf8.RuneCountInString(k)
					initialPad := strings.Repeat(" ", initialPadLen)
					consts.DisableColorIfNoColorEnvVarSet()
					green := color.FgGreen.Render
					bold := color.Bold.Render
					out.WriteString(fmt.Sprintf("\n%s%s| %s", bold(green(k)), initialPad, partStr))
				} else {
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
					out.WriteString(fmt.Sprintf("%s%s %s%s", prefixNl, pad, partStr, nl))
				}
			}
		}
	}
	return out.String()
}
