package evaluator

import (
	"blue/object"
	"strings"
)

func createStringList(input []string) []object.Object {
	list := make([]object.Object, len(input))
	for i, v := range input {
		list[i] = &object.Stringo{Value: v}
	}
	return list
}

var stringbuiltins = map[string]*object.Builtin{
	"startswith": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arg0, ok := args[0].(*object.Stringo)
			if !ok {
				return newError("first argument to `startwith` is not a string, got %s", args[0].Type())
			}
			arg1, ok := args[1].(*object.Stringo)
			if !ok {
				return newError("second argument to `startwith` is not a string, got %s", args[1].Type())
			}
			if strings.HasPrefix(arg0.Value, arg1.Value) {
				return TRUE
			}
			return FALSE
		},
	},
	"endswith": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arg0, ok := args[0].(*object.Stringo)
			if !ok {
				return newError("first argument to `startwith` is not a string, got %s", args[0].Type())
			}
			arg1, ok := args[1].(*object.Stringo)
			if !ok {
				return newError("second argument to `startwith` is not a string, got %s", args[1].Type())
			}
			if strings.HasSuffix(arg0.Value, arg1.Value) {
				return TRUE
			}
			return FALSE
		},
	},
	"split": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 and an optional separator", len(args))
			}
			if len(args) == 1 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newError("first argument to `split` is not a string, got %s", args[0].Type())
				}
				strList := strings.Split(arg0.Value, " ")
				return &object.List{Elements: createStringList(strList)}
			}
			if len(args) > 3 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arg0, ok := args[0].(*object.Stringo)
			if !ok {
				return newError("first argument to `split` is not a string, got %s", args[0].Type())
			}
			arg1, ok := args[1].(*object.Stringo)
			if !ok {
				return newError("second argument to `split` is not a string, got %s", args[1].Type())
			}
			strList := strings.Split(arg0.Value, arg1.Value)
			return &object.List{Elements: createStringList(strList)}
		},
	},
	"replace": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			arg0, ok := args[0].(*object.Stringo)
			if !ok {
				return newError("first argument to `replace` is not a string, got %s", args[0].Type())
			}
			arg1, ok := args[1].(*object.Stringo)
			if !ok {
				return newError("second argument to `replace` is not a string, got %s", args[1].Type())
			}
			arg2, ok := args[2].(*object.Stringo)
			if !ok {
				return newError("third argument to `replace` is not a string, got %s", args[2].Type())
			}
			replacedString := strings.ReplaceAll(arg0.Value, arg1.Value, arg2.Value)
			return &object.Stringo{Value: replacedString}
		},
	},
	"strip": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1 and an optional character to strip", len(args))
			}
			if len(args) == 1 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newError("first argument to `strip` is not a string, got %s", args[0].Type())
				}
				str := strings.TrimSpace(arg0.Value)
				return &object.Stringo{Value: str}
			}
			if len(args) == 2 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newError("second argument to `strip` is not a string, got %s", args[0].Type())
				}
				arg1, ok := args[1].(*object.Stringo)
				if !ok {
					return newError("third argument to `strip` is not a string, got %s", args[1].Type())
				}
				str := strings.Trim(arg0.Value, arg1.Value)
				return &object.Stringo{Value: str}
			}
			return newError("wrong number of arguments. got=%d want 2", len(args))
		},
	},
}
