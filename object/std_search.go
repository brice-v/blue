package object

import (
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
)

var SearchBuiltins = NewBuiltinSliceType{
	{Name: "_by_xpath", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("by_xpath", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("by_xpath", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ {
				return newPositionalTypeError("by_xpath", 2, STRING_OBJ, args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("by_xpath", 3, BOOLEAN_OBJ, args[2].Type())
			}
			strToSearch := args[0].(*Stringo).Value
			if strToSearch == "" {
				return newError("`by_xpath` error: str_to_search argument is empty")
			}
			strQuery := args[1].(*Stringo).Value
			if strQuery == "" {
				return newError("`by_xpath` error: query argument is empty")
			}
			shouldFindOne := args[2].(*Boolean).Value
			doc, err := htmlquery.Parse(strings.NewReader(strToSearch))
			if err != nil {
				return newError("`by_xpath` failed to parse document as html: error %s", err.Error())
			}
			if !shouldFindOne {
				listToReturn := &List{Elements: []Object{}}
				for _, e := range htmlquery.Find(doc, strQuery) {
					result := htmlquery.OutputHTML(e, true)
					listToReturn.Elements = append(listToReturn.Elements, &Stringo{Value: result})
				}
				return listToReturn
			} else {
				e := htmlquery.FindOne(doc, strQuery)
				result := htmlquery.OutputHTML(e, true)
				return &Stringo{Value: result}
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`by_xpath` finds the string based on an xpath query from the given html",
			signature:   "by_xpath(str_to_search: str, str_query: str, should_find_one: bool) -> list[str]|str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "by_xpath('<html><div id='abc'>123</div></html>', '//*[@id='abc']', true) => '<div id='abc'>123</div>'",
		}.String(),
	}},
	{Name: "_by_regex", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("by_regex", len(args), 3, "")
			}
			if args[0].Type() != STRING_OBJ {
				return newPositionalTypeError("by_regex", 1, STRING_OBJ, args[0].Type())
			}
			if args[1].Type() != STRING_OBJ && args[1].Type() != REGEX_OBJ {
				return newPositionalTypeError("by_regex", 2, STRING_OBJ+" or REGEX", args[1].Type())
			}
			if args[2].Type() != BOOLEAN_OBJ {
				return newPositionalTypeError("by_regex", 3, BOOLEAN_OBJ, args[2].Type())
			}
			strToSearch := args[0].(*Stringo).Value

			var re *regexp.Regexp
			if args[1].Type() == STRING_OBJ {
				strQuery := args[1].(*Stringo).Value
				re1, err := regexp.Compile(strQuery)
				if err != nil {
					return newError("`by_regex` error: failed to compile regexp %q", strQuery)
				}
				re = re1
			} else {
				re = args[1].(*Regex).Value
			}
			shouldFindOne := args[2].(*Boolean).Value

			if !shouldFindOne {
				listToReturn := &List{Elements: []Object{}}
				results := re.FindAllString(strToSearch, -1)
				for _, str := range results {
					listToReturn.Elements = append(listToReturn.Elements, &Stringo{Value: str})
				}
				return listToReturn
			} else {
				result := re.FindString(strToSearch)
				return &Stringo{Value: result}
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`by_regex` finds the string given a regex or string to search with",
			signature:   "by_regex(str_to_search: str, query: str|regex, should_find_one: bool) -> list[str]|str",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "by_regex('abc', r/abc/, true) => 'abc'",
		}.String(),
	}},
}
