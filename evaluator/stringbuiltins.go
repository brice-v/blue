package evaluator

import (
	"blue/object"
	"regexp"
	"strings"

	"github.com/huandu/xstrings"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
				return newInvalidArgCountError("startswith", len(args), 2, "")
			}
			arg0, ok := args[0].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("startswith", 1, object.STRING_OBJ, args[0].Type())
			}
			arg1, ok := args[1].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("startswith", 2, object.STRING_OBJ, args[1].Type())
			}
			return nativeToBooleanObject(strings.HasPrefix(arg0.Value, arg1.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`startswith` returns a BOOLEAN if the given STRING starts with the prefix STRING",
			signature:   "startswith(arg: str, prefix: str) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "startswith('Hello', 'Hell') => true",
		}.String(),
	},
	"endswith": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("endswith", len(args), 2, "")
			}
			arg0, ok := args[0].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("endswith", 1, object.STRING_OBJ, args[0].Type())
			}
			arg1, ok := args[1].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("endswith", 2, object.STRING_OBJ, args[1].Type())
			}
			return nativeToBooleanObject(strings.HasSuffix(arg0.Value, arg1.Value))
		},
		HelpStr: helpStrArgs{
			explanation: "`endswith` returns a BOOLEAN if the given STRING ends with the suffix STRING",
			signature:   "endswith(arg: str, suffix: str) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "endswith('Hello', 'o') => true",
		}.String(),
	},
	"split": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("split", len(args), 1, "or 2")
			}
			if len(args) == 1 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("split", 1, object.STRING_OBJ, args[0].Type())
				}
				strList := strings.Split(arg0.Value, " ")
				return &object.List{Elements: createStringList(strList)}
			}
			if len(args) == 2 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("split", 1, object.STRING_OBJ, args[0].Type())
				}
				arg1, ok := args[1].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("split", 2, object.STRING_OBJ, args[1].Type())
				}
				strList := strings.Split(arg0.Value, arg1.Value)
				return &object.List{Elements: createStringList(strList)}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`split` returns a LIST of STRINGs based on a STRING separator",
			signature:   "split(arg: str, sep: str) -> list[str]",
			errors:      "InvalidArgCount,PositionalType",
			example:     "split('Hello', '') => ['H','e','l','l','o']",
		}.String(),
	},
	"_replace": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("replace", len(args), 3, "")
			}
			arg0, ok := args[0].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("replace", 1, object.STRING_OBJ, args[0].Type())
			}
			arg1, ok := args[1].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("replace", 2, object.STRING_OBJ, args[1].Type())
			}
			arg2, ok := args[2].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("replace", 3, object.STRING_OBJ, args[2].Type())
			}
			replacedString := strings.ReplaceAll(arg0.Value, arg1.Value, arg2.Value)
			return &object.Stringo{Value: replacedString}
		},
		HelpStr: helpStrArgs{
			explanation: "`replace` will return a STRING with all occurrences of the given replacer STRING replaced by the next given STRING",
			signature:   "replace(arg: str, replacer: str, replaced: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "replace('Hello', 'l', 'X') => 'HeXXo'",
		}.String(),
	},
	"_replace_regex": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("replace_regex", len(args), 3, "")
			}
			arg0, ok := args[0].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("replace_regex", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ && args[1].Type() != object.REGEX_OBJ {
				return newPositionalTypeError("replace_regex", 2, object.STRING_OBJ+" or REGEX", args[1].Type())
			}
			var re *regexp.Regexp
			if args[1].Type() == object.STRING_OBJ {
				re1, err := regexp.Compile(args[1].(*object.Stringo).Value)
				if err != nil {
					return newError("`replace_regex` error: %s", err.Error())
				}
				re = re1
			} else {
				re = args[1].(*object.Regex).Value
			}
			arg2, ok := args[2].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("replace_regex", 3, object.STRING_OBJ, args[2].Type())
			}
			return &object.Stringo{Value: string(re.ReplaceAll([]byte(arg0.Value), []byte(arg2.Value)))}
		},
		HelpStr: helpStrArgs{
			explanation: "`replace_regex` will return a STRING with all occurrences of the given replacer REGEX STRING replaced by the next given STRING",
			signature:   "replace_regex(arg: str, replacer: str, replaced: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "replace_regex('Hello', 'l', 'X') => 'HeXXo'",
		}.String(),
	},
	"strip": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("strip", len(args), 1, "or 2")
			}
			if len(args) == 1 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("strip", 1, object.STRING_OBJ, args[0].Type())
				}
				str := strings.TrimSpace(arg0.Value)
				return &object.Stringo{Value: str}
			}
			if len(args) == 2 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("strip", 1, object.STRING_OBJ, args[0].Type())
				}
				arg1, ok := args[1].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("strip", 2, object.STRING_OBJ, args[1].Type())
				}
				str := strings.Trim(arg0.Value, arg1.Value)
				return &object.Stringo{Value: str}
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`strip` will return a STRING with the given cutset STRING removed",
			signature:   "strip(arg: str, cutset: str='') -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "strip(' Hello ') => 'Hello'",
		}.String(),
	},
	"lstrip": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("lstrip", len(args), 1, "or 2")
			}
			if len(args) == 1 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("lstrip", 1, object.STRING_OBJ, args[0].Type())
				}
				str := strings.TrimLeft(arg0.Value, " ")
				return &object.Stringo{Value: str}
			}
			if len(args) == 2 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("lstrip", 1, object.STRING_OBJ, args[0].Type())
				}
				arg1, ok := args[1].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("lstrip", 2, object.STRING_OBJ, args[1].Type())
				}
				str := strings.TrimLeft(arg0.Value, arg1.Value)
				return &object.Stringo{Value: str}
			}
			return NULL
		},
	},
	"rstrip": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 && len(args) != 2 {
				return newInvalidArgCountError("rstrip", len(args), 1, "or 2")
			}
			if len(args) == 1 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("rstrip", 1, object.STRING_OBJ, args[0].Type())
				}
				str := strings.TrimRight(arg0.Value, " ")
				return &object.Stringo{Value: str}
			}
			if len(args) == 2 {
				arg0, ok := args[0].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("rstrip", 1, object.STRING_OBJ, args[0].Type())
				}
				arg1, ok := args[1].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("rstrip", 2, object.STRING_OBJ, args[1].Type())
				}
				str := strings.TrimRight(arg0.Value, arg1.Value)
				return &object.Stringo{Value: str}
			}
			return NULL
		},
	},
	"to_json": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_json", len(args), 1, "")
			}
			if isError(args[0]) {
				return args[0]
			}
			return blueObjToJsonObject(args[0])
		},
		HelpStr: helpStrArgs{
			explanation: "`to_json` will return the JSON STRING of the given MAP",
			signature:   "to_json(arg: map[str:any]) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "to_json({'x': 123}) => '{\"x\":123}'",
		}.String(),
	},
	"to_upper": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_upper", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("to_upper", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return &object.Stringo{Value: strings.ToUpper(s)}
		},
		HelpStr: helpStrArgs{
			explanation: "`to_upper` returns the uppercase version of the given STRING",
			signature:   "to_upper(arg: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "to_upper('Hello') => 'HELLO'",
		}.String(),
	},
	"to_lower": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_lower", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("to_lower", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return &object.Stringo{Value: strings.ToLower(s)}
		},
		HelpStr: helpStrArgs{
			explanation: "`to_lower` returns the lowercase version of the given STRING",
			signature:   "to_lower(arg: str) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "to_lower('Hello') => 'hello'",
		}.String(),
	},
	"join": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("join", len(args), 2, "")
			}
			if args[0].Type() != object.LIST_OBJ {
				return newPositionalTypeError("join", 1, object.LIST_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("join", 2, object.STRING_OBJ, args[1].Type())
			}
			joiner := args[1].(*object.Stringo).Value
			elements := args[0].(*object.List).Elements
			strsToJoin := make([]string, len(elements))
			for i, e := range elements {
				if e.Type() != object.STRING_OBJ {
					return newError("found a type that was not a STRING in `join`. found=%s", e.Type())
				}
				strsToJoin[i] = e.(*object.Stringo).Value
			}
			return &object.Stringo{Value: strings.Join(strsToJoin, joiner)}
		},
		HelpStr: helpStrArgs{
			explanation: "`join` returns a STRING joined by the given joiner STRING for a list of STRINGs",
			signature:   "join(arg: list[str], joiner: str) -> str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "join(['H','e','l','l','o'], ' ') => 'H e l l o'",
		}.String(),
	},
	"_substr": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("substr", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("substr", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("substr", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("substr", 3, object.INTEGER_OBJ, args[2].Type())
			}
			s := args[0].(*object.Stringo).Value
			start := args[1].(*object.Integer).Value
			end := args[2].(*object.Integer).Value
			if end == -1 {
				end = int64(len(s))
			}
			return &object.Stringo{Value: s[start:end]}
		},
		HelpStr: helpStrArgs{
			explanation: "`_substr` returns the STRING from start INTEGER to end INTEGER",
			signature:   "_substr(arg: str, start: int=0, end: int=-1) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "_substr('Hello', 1, 3) => 'el'",
		}.String(),
	},
	"index_of": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("index_of", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("index_of", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.STRING_OBJ {
				return newPositionalTypeError("index_of", 2, object.STRING_OBJ, args[1].Type())
			}
			s := args[0].(*object.Stringo).Value
			subs := args[1].(*object.Stringo).Value
			return &object.Integer{Value: int64(strings.Index(s, subs))}
		},
		HelpStr: helpStrArgs{
			explanation: "`index_of` returns the INTEGER index of the given STRING substring",
			signature:   "index_of(arg: str, substr: str) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "index_of('Hello', 'ell') => 1",
		}.String(),
	},
	"_center": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("center", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("center", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("center", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newPositionalTypeError("center", 3, object.STRING_OBJ, args[2].Type())
			}
			s := args[0].(*object.Stringo).Value
			length := int(args[1].(*object.Integer).Value)
			pad := args[2].(*object.Stringo).Value
			return &object.Stringo{Value: xstrings.Center(s, length, pad)}
		},
	},
	"_ljust": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("ljust", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("ljust", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("ljust", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newPositionalTypeError("ljust", 3, object.STRING_OBJ, args[2].Type())
			}
			s := args[0].(*object.Stringo).Value
			length := int(args[1].(*object.Integer).Value)
			pad := args[2].(*object.Stringo).Value
			return &object.Stringo{Value: xstrings.LeftJustify(s, length, pad)}
		},
	},
	"_rjust": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 3 {
				return newInvalidArgCountError("rjust", len(args), 3, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("rjust", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != object.INTEGER_OBJ {
				return newPositionalTypeError("rjust", 2, object.INTEGER_OBJ, args[1].Type())
			}
			if args[2].Type() != object.STRING_OBJ {
				return newPositionalTypeError("rjust", 3, object.STRING_OBJ, args[2].Type())
			}
			s := args[0].(*object.Stringo).Value
			length := int(args[1].(*object.Integer).Value)
			pad := args[2].(*object.Stringo).Value
			return &object.Stringo{Value: xstrings.RightJustify(s, length, pad)}
		},
	},
	"to_title": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_title", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("to_title", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			caser := cases.Title(language.Und)
			return &object.Stringo{Value: caser.String(s)}
		},
	},
	"to_kebab": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_kebab", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("to_kebab", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return &object.Stringo{Value: xstrings.ToKebabCase(s)}
		},
	},
	"to_camel": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_camel", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("to_camel", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return &object.Stringo{Value: xstrings.ToCamelCase(s)}
		},
	},
	"to_snake": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newInvalidArgCountError("to_snake", len(args), 1, "")
			}
			if args[0].Type() != object.STRING_OBJ {
				return newPositionalTypeError("to_snake", 1, object.STRING_OBJ, args[0].Type())
			}
			s := args[0].(*object.Stringo).Value
			return &object.Stringo{Value: xstrings.ToSnakeCase(s)}
		},
	},
	"matches": {
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newInvalidArgCountError("matches", len(args), 2, "")
			}
			if args[0].Type() != object.STRING_OBJ && args[0].Type() != object.REGEX_OBJ {
				return newPositionalTypeError("matches", 1, object.STRING_OBJ+" or REGEX", args[0].Type())
			}
			if args[0].Type() == object.REGEX_OBJ {
				if args[1].Type() != object.STRING_OBJ {
					return newPositionalTypeError("matches", 2, object.STRING_OBJ, args[1].Type())
				}
				re := args[0].(*object.Regex).Value
				str := args[1].(*object.Stringo).Value
				return nativeToBooleanObject(re.MatchString(str))
			}
			// TODO: Support inverted arg as well? Like regex on left and string on right
			arg0, ok := args[0].(*object.Stringo)
			if !ok {
				return newPositionalTypeError("matches", 1, object.STRING_OBJ, args[0].Type())
			}
			if args[1].Type() == object.STRING_OBJ {
				arg1, ok := args[1].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("matches", 2, object.STRING_OBJ, args[1].Type())
				}
				re, err := regexp.Compile(arg1.Value)
				if err != nil {
					return newError("`matches` error: %s", err.Error())
				}
				return nativeToBooleanObject(re.MatchString(arg0.Value))
			}
			if args[1].Type() != object.REGEX_OBJ {
				return newPositionalTypeError("matches", 2, object.REGEX_OBJ, args[1].Type())
			}
			re := args[1].(*object.Regex).Value
			return nativeToBooleanObject(re.MatchString(arg0.Value))
		},
	},
})
