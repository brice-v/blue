package evaluator

import (
	"blue/object"
	"bytes"
	"strings"
)

func createStringList(input []string) []object.Object {
	list := make([]object.Object, len(input))
	for i, v := range input {
		list[i] = &object.Stringo{Value: v}
	}
	return list
}

var stringbuiltins = NewBuiltinObjMap(BuiltinMapTypeInternal{
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
	"to_json": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`to_json` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.MAP_OBJ {
				return newError("argument 1 to `to_json` should be MAP. got=%s", args[0].Type())
			}
			mObj := args[0].(*object.Map)
			// https://www.w3schools.com/Js/js_json_objects.asp
			// Keys must be strings, and values must be a valid JSON data type:
			// string
			// number
			// object
			// array
			// boolean
			// null
			ok, err := checkMapObjPairsForValidJsonKeysAndValues(mObj.Pairs)
			if !ok {
				return newError("`to_json` error validating MAP object. %s", err.Error())
			}
			var buf bytes.Buffer
			jsonString := generateJsonStringFromValidMapObjPairs(buf, mObj.Pairs)
			return &object.Stringo{Value: jsonString.String()}
		},
	},
	"trim": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("`trim` expects 1 argument. got=%d", len(args))
			}
			if args[0].Type() != object.STRING_OBJ {
				return newError("argument 1 to `trim` should be STRING. got=%s", args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return &object.Stringo{Value: strings.TrimSpace(s)}
		},
	},

	// TODO: join (list of strings)
	// TODO: We can probably create a solid regex object to use in the string methods
	// "test": {
	// 	Fun: func(args ...object.Object) object.Object {
	// 		if len(args) != 2 {
	// 			return newError("wrong number of arguments to `test`. got=%d want 2", len(args))
	// 		}
	// 		arg0, ok := args[0].(*object.Stringo)
	// 		if !ok {
	// 			return newError("first argument to `test` must be string. got=%T", args[0])
	// 		}
	// 		arg1, ok := args[1].(*object.Stringo)
	// 		if !ok {
	// 			return newError("second argument to `test` must be string. got=%T", args[1])
	// 		}
	// 		re, err := regexp.Compile(arg1.Value)
	// 		if err != nil {
	// 			return newError("second argument to `test` must be regex. got error: %s", err.Error())
	// 		}
	// 		if re.MatchString(arg0.Value) {
	// 			return TRUE
	// 		}
	// 		return FALSE
	// 	},
	// },
})
