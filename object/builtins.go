package object

import (
	"fmt"

	"github.com/gookit/color"
)

type BuiltinMapType struct {
	*ConcurrentMap[string, *Builtin]
}

func NewBuiltinObjMap(input map[string]*Builtin) BuiltinMapType {
	return BuiltinMapType{&ConcurrentMap[string, *Builtin]{
		Kv: input,
	}}
}

type BuiltinMapTypeInternal map[string]*Builtin

var Builtins = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"println": {
		Fun: func(args ...Object) Object {
			useColorPrinter := false
			var style color.Style
			for i, arg := range args {
				if i == 0 {
					t, s, ok := getBasicObjectForGoObj[color.Style](arg)
					if ok && t == "color" {
						// Use color printer
						useColorPrinter = true
						style = s
						continue
					} else {
						useColorPrinter = false
					}
				}
				if useColorPrinter {
					style.Println(arg.Inspect())
				} else {
					fmt.Println(arg.Inspect())
				}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`println` returns NULL and prints each of args on a new line",
			signature:   "println(args... : any...) -> null",
			errors:      "None",
			example:     "println(1,2,3) => null\n    STDOUT: 1\\n2\\n3\\n",
		}.String(),
	},
	"print": {
		Fun: func(args ...Object) Object {
			useColorPrinter := false
			var style color.Style
			for i, arg := range args {
				if i == 0 {
					t, s, ok := getBasicObjectForGoObj[color.Style](arg)
					if ok && t == "color" {
						// Use color printer
						useColorPrinter = true
						style = s
						continue
					} else {
						useColorPrinter = false
					}
				}
				if useColorPrinter {
					style.Print(arg.Inspect())
				} else {
					fmt.Print(arg.Inspect())
				}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`print` returns NULL and prints each of args",
			signature:   "print(args... : any...) -> null",
			errors:      "None",
			example:     "print(1,2,3) => null\n    STDOUT: 123",
		}.String(),
	},
})

func getBasicObjectForGoObj[T any](arg Object) (string, T, bool) {
	var zero T
	if arg == nil {
		return "", zero, false
	}
	if arg.Type() != MAP_OBJ {
		return "", zero, false
	}
	objPairs := arg.(*Map).Pairs
	if objPairs.Len() != 2 {
		return "", zero, false
	}
	// Get the 't' value
	hk1 := objPairs.Keys[0]
	mp1, ok := objPairs.Get(hk1)
	if !ok {
		return "", zero, false
	}
	if mp1.Key.Type() != STRING_OBJ {
		return "", zero, false
	}
	if mp1.Value.Type() != STRING_OBJ {
		return "", zero, false
	}
	if mp1.Key.(*Stringo).Value != "t" {
		return "", zero, false
	}
	t := mp1.Value.(*Stringo).Value
	// Get the 'v' value
	hk2 := objPairs.Keys[1]
	mp2, ok := objPairs.Get(hk2)
	if !ok {
		return "", zero, false
	}
	if mp2.Key.Type() != STRING_OBJ {
		return "", zero, false
	}
	if mp2.Value.Type() != GO_OBJ {
		return "", zero, false
	}
	if mp2.Key.(*Stringo).Value != "v" {
		return "", zero, false
	}
	v := mp2.Value.(*GoObj[T]).Value
	return t, v, true
}

type helpStrArgs struct {
	explanation string
	signature   string
	errors      string
	example     string
}

func (hsa helpStrArgs) String() string {
	return fmt.Sprintf("%s\n    Signature:  %s\n    Error(s):   %s\n    Example(s): %s\n", hsa.explanation, hsa.signature, hsa.errors, hsa.example)
}
