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
// func NewEnclosedEnvironment2(outer *Environment2) *Environment2 {
// 	env := NewEnvironment(outer.coreEnv)
// 	env.outer = outer
// 	return env
// }

// NewEnvironment returns a new environment
func NewEnvironmentWithoutCore2() *Environment2 {
	return &Environment2{frames: []*Frame{NewFrame()}}
}

// NewEnvironment returns a new environment
func NewEnvironment2(coreEnv *Environment2) *Environment2 {
	e := NewEnvironmentWithoutCore2()
	e.SetCore(coreEnv)
	return e
}

type Frame struct {
	*xsync.MapOf[string, *ObjectRef]
}

func NewFrame() *Frame {
	return &Frame{xsync.NewMapOf[string, *ObjectRef]()}
}

func (f *Frame) Set(name string, value Object) {
	f.Store(name, &ObjectRef{Ref: value, isImmutable: false})
}

// Environment is a map of strings to `Object`s
type Environment2 struct {
	frames []*Frame

	coreEnv *Environment2
	fp      int
}

func (e *Environment2) SetCore(coreEnv *Environment2) {
	e.coreEnv = coreEnv
}

// Clone will do a deep copy of the environment object and return a new environment
// it will not write to the outer env, but it will read from it
func (e *Environment2) Clone() *Environment2 {
	newEnv := NewEnvironment2(e.coreEnv)
	for i := e.fp; i >= 0; i-- {
		e.frames[i].Range(func(key string, value *ObjectRef) bool {
			newEnv.frames[newEnv.fp].Store(key, value)
			return true
		})
	}
	return newEnv
}

// Clone will do a deep copy of the environment object and return a new environment
// it will not write to the outer env, but it will read from it
func (e *Environment2) CloneToFrame() *Frame {
	f := NewFrame()
	for i := e.fp; i >= 0; i-- {
		e.frames[i].Range(func(key string, value *ObjectRef) bool {
			f.Store(key, value)
			return true
		})
	}
	// TODO: Clone Core Env?
	return f
}

func (e *Environment2) GetLatestFrame() *Frame {
	return e.frames[e.fp]
}

func (e *Environment2) PushFrame(f *Frame) {
	e.frames = append(e.frames, f)
	e.fp++
}

func (e *Environment2) PopFrame() {
	e.fp--
	e.frames = e.frames[:e.fp+1]
}

// func (e *Environment2) String() string {
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
func (e *Environment2) Get(name string) (Object, bool) {
	obj, ok := e.GetRef(name)
	if !ok {
		return nil, ok
	}
	return obj.Ref, ok
}

func (e *Environment2) GetRef(name string) (*ObjectRef, bool) {
	for i := e.fp; i >= 0; i-- {
		if obj, ok := e.frames[i].Load(name); ok {
			return obj, ok
		}
	}
	if e.coreEnv != nil {
		for i := e.coreEnv.fp; i >= 0; i-- {
			if obj, ok := e.coreEnv.frames[i].Load(name); ok {
				return obj, ok
			}
		}
	}
	return nil, false
}

func (e *Environment2) SetAllPublicOnEnv(newEnv *Environment2) {
	e.frames[e.fp].Range(func(key string, value *ObjectRef) bool {
		if !strings.HasPrefix(key, "_") {
			newEnv.SetObj(key, value.Ref, value.isImmutable)
		}
		return true
	})
}

func (e *Environment2) getFunctionHelpString(origHelp, prefix string) string {
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
func (e *Environment2) Set(name string, val Object) Object {
	return e.SetObj(name, val, false)
}

func (e *Environment2) SetObj(name string, val Object, isImmutable bool) Object {
	helpStr := e.getHelpInPublicFunctionHelpStore(name, val)
	e.frames[e.fp].Store(name, &ObjectRef{Ref: val, isImmutable: isImmutable, HelpStr: helpStr})
	return val
}

func (e *Environment2) SetImmutable(name string, val Object) Object {
	return e.SetObj(name, val, true)
}

func (e *Environment2) SetFunStatementAndHelp(name string, val *Function) Object {
	helpStr := e.getHelpInPublicFunctionHelpStore(name, val)
	e.frames[e.fp].Store(name, &ObjectRef{Ref: val, isImmutable: false, HelpStr: helpStr})
	return val
}

func (e *Environment2) getHelpInPublicFunctionHelpStore(name string, val Object) string {
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
func (e *Environment2) IsImmutable(name string) bool {
	obj, ok := e.GetRef(name)
	if !ok {
		return ok
	}
	return obj.isImmutable
}

// RemoveIdentifier removes a key from the environment
// this is used in for loops to remove temporary variables
func (e *Environment2) RemoveIdentifier(name string) {
	e.frames[e.fp].Delete(name)
}

func (e *Environment2) getLengthOfLargestStringAndOrderedKeys() (int, []string) {
	lengthOfLargestString := 0
	keys := []string{}
	e.frames[e.fp].Range(func(key string, value *ObjectRef) bool {
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

func (e *Environment2) GetOrderedPublicFunctionHelpString() string {
	var out bytes.Buffer
	lengthOfLargestString, orderedKeys := e.getLengthOfLargestStringAndOrderedKeys()
	for _, k := range orderedKeys {
		value, _ := e.frames[e.fp].Load(k)
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
