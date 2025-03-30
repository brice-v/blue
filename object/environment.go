package object

import (
	"blue/consts"
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gookit/color"
	"github.com/puzpuzpuz/xsync/v3"
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
	pfhs := &OrderedMap2[string, string]{
		store: make(map[string]string),
		Keys:  []string{},
	}
	return &Environment{store: xsync.NewMapOf[string, Object](), immutableStore: xsync.NewMapOf[string, struct{}](), publicFunctionHelpStore: pfhs}
}

// NewEnvironment returns a new environment
func NewEnvironment(coreEnv *Environment) *Environment {
	pfhs := &OrderedMap2[string, string]{
		store: make(map[string]string),
		Keys:  []string{},
	}
	return &Environment{store: xsync.NewMapOf[string, Object](), immutableStore: xsync.NewMapOf[string, struct{}](), publicFunctionHelpStore: pfhs, coreEnv: coreEnv}
}

// Environment is a map of strings to `Object`s
type Environment struct {
	store          *xsync.MapOf[string, Object]
	immutableStore *xsync.MapOf[string, struct{}]

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
	e.store.Range(func(key string, value Object) bool {
		newEnv.store.Store(key, value)
		return true
	})
	e.immutableStore.Range(func(key string, value struct{}) bool {
		newEnv.immutableStore.Store(key, value)
		return true
	})
	outer := e.outer
	for outer != nil {
		outer.store.Range(func(key string, value Object) bool {
			newEnv.store.Store(key, value)
			return true
		})
		outer.immutableStore.Range(func(key string, value struct{}) bool {
			newEnv.immutableStore.Store(key, value)
			return true
		})
		outer = outer.outer
	}
	for _, k := range e.publicFunctionHelpStore.Keys {
		v, _ := e.publicFunctionHelpStore.Get(k)
		newEnv.publicFunctionHelpStore.Set(k, v)
	}
	return newEnv
}

// Rewrite Func with sync.Map implementation
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
	obj, ok := e.store.Load(name)
	if !ok && e.coreEnv != nil {
		obj, ok = e.coreEnv.Get(name)
		if !ok && e.outer != nil {
			obj, ok = e.outer.Get(name)
		}
	}
	if obj == nil {
		return nil, ok
	}
	return obj, ok
}

func (e *Environment) SetAllPublicOnEnv(newEnv *Environment) {
	e.store.Range(func(key string, value Object) bool {
		if !strings.HasPrefix(key, "_") {
			newEnv.store.Store(key, value)
		}
		return true
	})
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
	e.store.Store(name, val)
	e.setHelpInPublicFunctionHelpStore(name, val)
	return val
}

func (e *Environment) SetFunStatementAndHelp(name string, val *Function) Object {
	e.store.Store(name, val)
	e.setHelpInPublicFunctionHelpStore(name, val)
	return val
}

func (e *Environment) setHelpInPublicFunctionHelpStore(name string, val Object) {
	// We do store nil values so those can be skipped entirely for pfhs
	if val != nil && val.Type() == FUNCTION_OBJ && !strings.HasPrefix(name, "_") {
		ogHelp := val.Help()
		if !strings.HasPrefix(ogHelp, "core:ignore") {
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
}

// ImmutableSet puts the name of the identifier in a map and sets it as true
func (e *Environment) ImmutableSet(name string) {
	e.immutableStore.Store(name, struct{}{})
}

// IsImmutable checks if the give identifier name is in the immutable map
func (e *Environment) IsImmutable(name string) bool {
	_, ok := e.immutableStore.Load(name)
	if !ok && e.outer != nil {
		ok = e.outer.IsImmutable(name)
		return ok
	}
	return ok
}

// RemoveIdentifier removes a key from the environment
// this is used in for loops to remove temporary variables
func (e *Environment) RemoveIdentifier(name string) {
	e.store.Delete(name)
	e.immutableStore.Delete(name)
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
