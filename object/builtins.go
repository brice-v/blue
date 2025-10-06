package object

import (
	"fmt"
	"unicode/utf8"

	"github.com/gookit/color"
	"github.com/huandu/go-clone"
	"github.com/huandu/xstrings"
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

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{Name: "println",
		Builtin: &Builtin{
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
	},
	{Name: "print",
		Builtin: &Builtin{
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
	},
	{Name: "help",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("help", len(args), 1, "")
				}
				return &Stringo{Value: args[0].Help()}
			},
			HelpStr: helpStrArgs{
				explanation: "`help` returns the help STRING for a given OBJECT",
				signature:   "help(arg: any) -> str",
				errors:      "None",
				example:     "help(print) => `prints help string for print builtin function`",
			}.String(),
		},
	},
	{Name: "new",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("new", len(args), 1, "")
				}
				if !IsCollectionType(args[0].Type()) {
					return newPositionalTypeError("new", 1, "MAP or LIST or SET", args[0].Type())
				}
				if args[0].Type() == MAP_OBJ {
					m := args[0].(*Map)
					return clone.Clone(m).(*Map)
				} else if args[0].Type() == LIST_OBJ {
					l := args[0].(*List)
					return clone.Clone(l).(*List)
				} else {
					s := args[0].(*Set)
					return clone.Clone(s).(*Set)
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`new` returns a cloned MAP object from the given arg",
				signature:   "new(arg: map|list|set) -> map|list|set",
				errors:      "InvalidArgCount,PositionalType",
				example:     "new({'x': 1}) => {'x': 1}",
			}.String(),
		},
	},
	{Name: "keys",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("keys", len(args), 1, "")
				}
				if args[0].Type() != MAP_OBJ {
					return newPositionalTypeError("keys", 1, MAP_OBJ, args[0].Type())
				}
				returnList := &List{
					Elements: []Object{},
				}
				m := args[0].(*Map)
				for _, k := range m.Pairs.Keys {
					mp, _ := m.Pairs.Get(k)
					returnList.Elements = append(returnList.Elements, mp.Key)
				}
				return returnList
			},
			HelpStr: helpStrArgs{
				explanation: "`keys` returns a LIST of key objects from given arg",
				signature:   "keys(arg: map) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "keys({'a': 1, 'B': 2}) => ['a', 'B']",
			}.String(),
		},
	},
	{Name: "values",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("values", len(args), 1, "")
				}
				if args[0].Type() != MAP_OBJ {
					return newPositionalTypeError("values", 1, MAP_OBJ, args[0].Type())
				}
				returnList := &List{
					Elements: []Object{},
				}
				m := args[0].(*Map)
				for _, k := range m.Pairs.Keys {
					mp, _ := m.Pairs.Get(k)
					returnList.Elements = append(returnList.Elements, mp.Value)
				}
				return returnList
			},
			HelpStr: helpStrArgs{
				explanation: "`values` returns a LIST of value OBJECTs from given arg",
				signature:   "value(arg: map) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "keys({'a': 1, 'B': 2}) => [1, 2]",
			}.String(),
		},
	},
	{Name: "del",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 2 {
					return newInvalidArgCountError("del", len(args), 2, "")
				}
				if !IsCollectionType(args[0].Type()) {
					return newPositionalTypeError("del", 1, "MAP or LIST or SET", args[0].Type())
				}
				if args[0].Type() == MAP_OBJ {
					m := args[0].(*Map)
					hk := HashKey{
						Type:  args[1].Type(),
						Value: HashObject(args[1]),
					}
					m.Pairs.Delete(hk)
				} else if args[1].Type() == LIST_OBJ {
					if args[1].Type() != INTEGER_OBJ {
						return newPositionalTypeError("del", 2, LIST_OBJ, args[1].Type())
					}
					l := args[0].(*List)
					index := args[1].(*Integer).Value
					l.Elements = append(l.Elements[:index], l.Elements[index+1:]...)
				} else {
					hk := HashKey{
						Type:  args[1].Type(),
						Value: HashObject(args[1]),
					}
					s := args[0].(*Set)
					s.Elements.Delete(hk.Value)
				}
				return NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`del` deletes a key from a MAP or index from a LIST",
				signature:   "del(m: map|list, key: any|int) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "del({'a': 1, 'B': 2}, 'a') => null - side effect: {'B': 2}",
			}.String(),
		},
	},
	{Name: "len",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("len", len(args), 1, "")
				}

				switch arg := args[0].(type) {
				case *Stringo:
					return &Integer{Value: int64(utf8.RuneCountInString(arg.Value))}
				case *List:
					return &Integer{Value: int64(len(arg.Elements))}
				case *Map:
					return &Integer{Value: int64(arg.Pairs.Len())}
				case *Set:
					return &Integer{Value: int64(arg.Elements.Len())}
				case *Bytes:
					return &Integer{Value: int64(len(arg.Value))}
				default:
					return newPositionalTypeError("len", 1, "STRING, LIST, MAP, SET, or BYTES", args[0].Type())
				}
			},
			HelpStr: helpStrArgs{
				explanation: "`len` returns the INTEGER length of the given STRING, LIST, MAP, or SET",
				signature:   "len(arg: str|list|map|set) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "len([1,2,3]) => 3",
			}.String(),
		},
	},
	{Name: "append",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("append", len(args), 2, " or more")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("append", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				length := len(l.Elements)
				args = args[1:]
				argsLength := len(args)
				newElements := make([]Object, length+argsLength)
				copy(newElements, l.Elements)
				copy(newElements[length:length+argsLength], args)
				return &List{Elements: newElements}
			},
			HelpStr: helpStrArgs{
				explanation: "`append` returns the LIST of elements with given arg OBJECT at the end",
				signature:   "append(arg0: list, args...: any) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "append([1,2,3], 1) => [1,2,3,4]",
			}.String(),
		},
	},
	{Name: "prepend",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("prepend", len(args), 2, "")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("prepend", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				length := len(l.Elements)
				args = args[1:]
				argsLength := len(args)
				newElements := make([]Object, length+argsLength)
				copy(newElements, args)
				copy(newElements[argsLength:argsLength+length], l.Elements)
				return &List{Elements: newElements}
			},
			HelpStr: helpStrArgs{
				explanation: "`prepend` returns the LIST of elements with given arg OBJECT at the front",
				signature:   "prepend(arg0: list, args...: any) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "prepend([1,2,3], 4) => [4,1,2,3]",
			}.String(),
		},
	},
	{Name: "push",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("push", len(args), 2, " or more")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("push", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				args = args[1:]
				l.Elements = append(l.Elements, args...)
				return &Integer{Value: int64(len(l.Elements))}
			},
			HelpStr: helpStrArgs{
				explanation: "`push` puts the given args at the end of the LIST and mutates it. The value returned is the length after pushing",
				signature:   "push(arg0: list[any], args...: any) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "push([1,2,3], 1) => 4",
			}.String(),
			Mutates: true,
		},
	},
	{Name: "pop",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("pop", len(args), 1, "")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("pop", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				if len(l.Elements) == 0 {
					return NULL
				}
				elem := l.Elements[len(l.Elements)-1]
				l.Elements = l.Elements[:len(l.Elements)-1]
				return elem
			},
			HelpStr: helpStrArgs{
				explanation: "`pop` returns the last element of the LIST and mutates it",
				signature:   "pop(arg0: list[any]) -> any",
				errors:      "InvalidArgCount,PositionalType",
				example:     "pop([1,2,3]) => 3",
			}.String(),
			Mutates: true,
		},
	},
	{Name: "unshift",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("unshift", len(args), 2, " or more")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("unshift", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				length := len(l.Elements)
				args = args[1:]
				argsLength := len(args)
				elems := make([]Object, argsLength+length)
				copy(elems, args)
				copy(elems[argsLength:argsLength+length], l.Elements)
				l.Elements = elems
				return &Integer{Value: int64(len(l.Elements))}
			},
			HelpStr: helpStrArgs{
				explanation: "`unshift` prepends the LIST with the given arguments and mutates it. The new length is returned",
				signature:   "unshift(arg0: list[any], args...: any) -> int",
				errors:      "InvalidArgCount,PositionalType",
				example:     "unshift([1,2,3], 1) => 4",
			}.String(),
			Mutates: true,
		},
	},
	{Name: "shift",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("shift", len(args), 1, "")
				}
				if args[0].Type() != LIST_OBJ {
					return newPositionalTypeError("shift", 1, LIST_OBJ, args[0].Type())
				}
				l := args[0].(*List)
				if len(l.Elements) == 0 {
					return NULL
				}
				elem := l.Elements[0]
				if len(l.Elements) == 1 {
					l.Elements = []Object{}
				} else {
					l.Elements = l.Elements[1:]
				}
				return elem
			},
			HelpStr: helpStrArgs{
				explanation: "`shift` returns the first element of the LIST and mutates it",
				signature:   "shift(arg0: list[any]) -> any",
				errors:      "InvalidArgCount,PositionalType",
				example:     "shift([1,2,3]) => 1",
			}.String(),
			Mutates: true,
		},
	},
	{Name: "concat",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) < 2 {
					return newInvalidArgCountError("concat", len(args), 2, " or more")
				}
				allElems := []Object{}
				for i, e := range args {
					if e.Type() != LIST_OBJ {
						return newPositionalTypeError("concat", i+1, LIST_OBJ, e.Type())
					}
					l := e.(*List)
					allElems = append(allElems, l.Elements...)
				}
				return &List{Elements: allElems}
			},
			HelpStr: helpStrArgs{
				explanation: "`concat` merges 2 or more LISTs together and returns the result",
				signature:   "concat(arg0: list[any], args...: list[any]) -> list[any]",
				errors:      "InvalidArgCount,PositionalType",
				example:     "concat([1,2,3], [1]) => [1,2,3,1]",
			}.String(),
		},
	},
	{Name: "reverse",
		Builtin: &Builtin{
			Fun: func(args ...Object) Object {
				if len(args) != 1 {
					return newInvalidArgCountError("reverse", len(args), 1, "")
				}
				if args[0].Type() != LIST_OBJ && args[0].Type() != STRING_OBJ {
					return newPositionalTypeError("reverse", 1, LIST_OBJ+" or STRING", args[0].Type())
				}
				if args[0].Type() == STRING_OBJ {
					s := args[0].(*Stringo).Value
					return &Stringo{Value: xstrings.Reverse(s)}
				}
				l := args[0].(*List)
				length := len(l.Elements)
				newl := make([]Object, length)
				for i := 0; i < len(l.Elements); i++ {
					newl[length-i-1] = l.Elements[i]
				}
				return &List{Elements: newl}
			},
			HelpStr: helpStrArgs{
				explanation: "`reverse` reverse a string or list",
				signature:   "reverse(arg: list[any]|str) -> list[any]|str",
				errors:      "InvalidArgCount,PositionalType",
				example:     "reverse([1,2,3]) => [3,2,1]",
			}.String(),
		},
	},
}

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

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}
