package object

import (
	"time"

	"github.com/golang-module/carbon/v2"
)

var TimeBuiltins = NewBuiltinSliceType{
	{Name: "_sleep", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("sleep", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("sleep", 1, INTEGER_OBJ, args[0].Type())
			}
			i := args[0].(*Integer).Value
			if i < 0 {
				return newError("INTEGER argument to `sleep` must be > 0, got=%d", i)
			}
			time.Sleep(time.Duration(i) * time.Millisecond)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`sleep` will sleep and block for the given INTEGER by milliseconds",
			signature:   "sleep(i: int) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "sleep(100) => null",
		}.String(),
	}},
	{Name: "_now", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("now", len(args), 0, "")
			}
			return &Integer{Value: time.Now().UnixMilli()}
		},
		HelpStr: helpStrArgs{
			explanation: "`now` returns the current unix timestamp in milliseconds as an INTEGER",
			signature:   "now() -> int",
			errors:      "InvalidArgCount",
			example:     "now(100) => 1703479130205",
		}.String(),
	}},
	{Name: "_parse", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("parse", len(args), 1, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("parse", 1, STRING_OBJ, args[0].Type())
			}
			s := args[0].(*Stringo).Value
			return &Integer{Value: carbon.Parse(s).StdTime().UnixMilli()}
		},
		HelpStr: helpStrArgs{
			explanation: "`parse` returns the parsed timestamp a unix timestamp in milliseconds as an INTEGER",
			signature:   "parse(s: str) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "parse('now') => 1703479130205",
		}.String(),
	}},
	{Name: "_to_str", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("to_str", len(args), 2, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("to_str", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ && args[1].Type() != NULL_OBJ {
				return newPositionalTypeError("to_str", 2, "STRING or NULL", args[1].Type())
			}
			i := args[0].(*Integer).Value
			tm := time.UnixMilli(i)
			if args[1].Type() == STRING_OBJ {
				tz := args[1].(*Stringo).Value
				return &Stringo{Value: carbon.CreateFromStdTime(tm).ToDateTimeMilliString(tz)}
			} else {
				return &Stringo{Value: carbon.CreateFromStdTime(tm).ToDateTimeMilliString()}
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`to_str` returns the string fomratted version of a unix timestamp value",
			signature:   "to_str(i: int) -> str",
			errors:      "InvalidArgCount,PositionalType",
			example:     "to_str(1703479130205) => '2023-12-24 23:42:28.144'",
		}.String(),
	}},
}
