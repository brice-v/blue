package object

import (
	"blue/consts"
	"bytes"
	"fmt"
	"slices"
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
	return &Environment{store: xsync.NewMapOf[string, *ObjectRef]()}
}

// NewEnvironment returns a new environment
func NewEnvironment(coreEnv *Environment) *Environment {
	return &Environment{
		store:   xsync.NewMapOf[string, *ObjectRef](),
		coreEnv: coreEnv,
	}
}

type ObjectRef struct {
	Ref         Object
	IsImmutable bool
	HelpStr     string
}

// Environment is a map of strings to `Object`s
type Environment struct {
	store *xsync.MapOf[string, *ObjectRef]

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
	e.store.Range(func(key string, value *ObjectRef) bool {
		newEnv.store.Store(key, value)
		return true
	})
	outer := e.outer
	for outer != nil {
		outer.store.Range(func(key string, value *ObjectRef) bool {
			newEnv.store.Store(key, value)
			return true
		})
		outer = outer.outer
	}
	return newEnv
}

// func (e *Environment) String() string {
// 	var out bytes.Buffer
// 	out.WriteString("Environment{\n\tstore:\n")
// 	e.store.Range(func(key string, value ObjectRef) bool {
// 		out.WriteString(fmt.Sprintf("\t\t%q -> %s\n", key, value.Ref.Type()))
// 		return true
// 	})
// 	out.WriteString("}")
// 	return out.String()
// }

// Get returns the object from the environment store
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.GetRef(name)
	if !ok {
		return nil, ok
	}
	return obj.Ref, ok
}

func (e *Environment) GetRef(name string) (*ObjectRef, bool) {
	obj, ok := e.store.Load(name)
	if !ok && e.coreEnv != nil {
		obj, ok = e.coreEnv.GetRef(name)
		if !ok && e.outer != nil {
			obj, ok = e.outer.GetRef(name)
		}
	}
	if obj == nil {
		return nil, ok
	}
	return obj, ok
}

func (e *Environment) SetAllPublicOnEnv(newEnv *Environment) {
	e.store.Range(func(key string, value *ObjectRef) bool {
		if !strings.HasPrefix(key, "_") {
			newEnv.SetObj(key, value.Ref, value.IsImmutable)
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

// Set puts a new object into the environment
func (e *Environment) Set(name string, val Object) Object {
	return e.SetObj(name, val, false)
}

func (e *Environment) SetObj(name string, val Object, isImmutable bool) Object {
	helpStr := e.getHelpInPublicFunctionHelpStore(name, val)
	objRef := &ObjectRef{Ref: val, HelpStr: helpStr, IsImmutable: isImmutable}
	e.store.Store(name, objRef)
	return val
}

func (e *Environment) SetFunStatementAndHelp(name string, val *Function) Object {
	helpStr := e.getHelpInPublicFunctionHelpStore(name, val)
	e.store.Store(name, &ObjectRef{Ref: val, HelpStr: helpStr})
	return val
}

func (e *Environment) getHelpInPublicFunctionHelpStore(name string, val Object) string {
	// We do store nil values so those can be skipped entirely for pfhs
	var help string
	if val != nil && val.Type() == FUNCTION_OBJ && !strings.HasPrefix(name, "_") {
		ogHelp := val.Help()
		if !strings.HasPrefix(ogHelp, "core:ignore") {
			if strings.HasPrefix(ogHelp, "core:") {
				help = e.getFunctionHelpString(ogHelp, "core:")
			} else if strings.HasPrefix(ogHelp, "std:") {
				help = e.getFunctionHelpString(ogHelp, "std:")
			} else {
				help = ogHelp
			}
			val.(*Function).HelpStr = help
		}
	}
	return help
}

// IsImmutable checks if the give identifier name is in the immutable map
func (e *Environment) IsImmutable(name string) bool {
	obj, ok := e.GetRef(name)
	if !ok {
		return ok
	}
	return obj.IsImmutable
}

// RemoveIdentifier removes a key from the environment
// this is used in for loops to remove temporary variables
func (e *Environment) RemoveIdentifier(name string) {
	e.store.Delete(name)
}

func (e *Environment) getLengthOfLargestStringAndOrderedKeys() (int, []string) {
	lengthOfLargestString := 0
	keys := []string{}
	e.store.Range(func(key string, value *ObjectRef) bool {
		if value != nil && value.HelpStr != "" {
			keys = append(keys, key)
			l := utf8.RuneCountInString(key)
			if l > lengthOfLargestString {
				lengthOfLargestString = l
			}
		}
		return true
	})
	lengthOfLargestString++
	slices.Sort(keys)
	return lengthOfLargestString, keys
}

func (e *Environment) GetOrderedPublicFunctionHelpString() string {
	var out bytes.Buffer
	lengthOfLargestString, orderedKeys := e.getLengthOfLargestStringAndOrderedKeys()
	for _, k := range orderedKeys {
		value, _ := e.store.Load(k)
		v := value.HelpStr
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
	return out.String()
}
