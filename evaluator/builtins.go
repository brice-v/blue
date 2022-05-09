package evaluator

import (
	"blue/object"
	"fmt"
)

var builtins = map[string]*object.Builtin{
	"len": &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Stringo:
				return &object.Integer{Value: int64(len([]rune(arg.Value)))}
			case *object.List:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.Map:
				return &object.Integer{Value: int64(len(arg.Pairs))}
			default:
				// TODO: add in support for other items that will be supported with len
				// ie. lists, maps, sets, etc.
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"first": &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument to `first` must be LIST, got %s", args[0].Type())
			}
			l := args[0].(*object.List)
			if len(l.Elements) > 0 {
				return l.Elements[0]
			}
			return NULL
		},
	},
	"last": &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument to `first` must be LIST, got %s", args[0].Type())
			}
			l := args[0].(*object.List)
			last := len(l.Elements)
			if last != 0 {
				return l.Elements[last-1]
			}
			return NULL
		},
	},
	"rest": &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument to `rest` must be LIST, got %s", args[0].Type())
			}
			l := args[0].(*object.List)
			length := len(l.Elements)
			if length > 0 {
				// NOTE: This is an efficient way of copying slices/lists so use elsewhere
				newElements := make([]object.Object, length-1, length-1)
				copy(newElements, l.Elements[1:length])
				return &object.List{Elements: newElements}
			}
			return NULL
		},
	},
	"append": &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != object.LIST_OBJ {
				return newError("argument to `append` must be LIST, got %s", args[0].Type())
			}
			l := args[0].(*object.List)
			length := len(l.Elements)
			// NOTE: This is an efficient way of appending but probably could just append onto the list
			newElements := make([]object.Object, length+1, length+1)
			copy(newElements, l.Elements)
			newElements[length] = args[1]
			return &object.List{Elements: newElements}
		},
	},
	"println": &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
	"print": &object.Builtin{
		Fun: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			return NULL
		},
	},
}
