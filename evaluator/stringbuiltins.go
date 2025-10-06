package evaluator

import (
	"blue/object"
)

var stringbuiltins = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"startswith":     object.GetBuiltinByName("startswith"),
	"endswith":       object.GetBuiltinByName("endswith"),
	"split":          object.GetBuiltinByName("split"),
	"_replace":       object.GetBuiltinByName("_replace"),
	"_replace_regex": object.GetBuiltinByName("_replace_regex"),
	"strip":          object.GetBuiltinByName("strip"),
	"lstrip":         object.GetBuiltinByName("lstrip"),
	"rstrip":         object.GetBuiltinByName("rstrip"),
	"to_json":        object.GetBuiltinByName("to_json"),
	"to_upper":       object.GetBuiltinByName("to_upper"),
	"to_lower":       object.GetBuiltinByName("to_lower"),
	"join":           object.GetBuiltinByName("join"),
	"_substr":        object.GetBuiltinByName("_substr"),
	"index_of":       object.GetBuiltinByName("index_of"),
	"_center":        object.GetBuiltinByName("_center"),
	"_ljust":         object.GetBuiltinByName("_ljust"),
	"_rjust":         object.GetBuiltinByName("_rjust"),
	"to_title":       object.GetBuiltinByName("to_title"),
	"to_kebab":       object.GetBuiltinByName("to_kebab"),
	"to_camel":       object.GetBuiltinByName("to_camel"),
	"to_snake":       object.GetBuiltinByName("to_snake"),
	"matches":        object.GetBuiltinByName("matches"),
})
