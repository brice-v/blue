package evaluator

import (
	"blue/object"
)

var stringbuiltins = NewBuiltinObjMap(BuiltinMapTypeInternal{
	"startswith":     object.GetBuiltinByName(object.BuiltinBaseType, "startswith"),
	"endswith":       object.GetBuiltinByName(object.BuiltinBaseType, "endswith"),
	"split":          object.GetBuiltinByName(object.BuiltinBaseType, "split"),
	"_replace":       object.GetBuiltinByName(object.BuiltinBaseType, "_replace"),
	"_replace_regex": object.GetBuiltinByName(object.BuiltinBaseType, "_replace_regex"),
	"strip":          object.GetBuiltinByName(object.BuiltinBaseType, "strip"),
	"lstrip":         object.GetBuiltinByName(object.BuiltinBaseType, "lstrip"),
	"rstrip":         object.GetBuiltinByName(object.BuiltinBaseType, "rstrip"),
	"to_json":        object.GetBuiltinByName(object.BuiltinBaseType, "to_json"),
	"to_upper":       object.GetBuiltinByName(object.BuiltinBaseType, "to_upper"),
	"to_lower":       object.GetBuiltinByName(object.BuiltinBaseType, "to_lower"),
	"join":           object.GetBuiltinByName(object.BuiltinBaseType, "join"),
	"_substr":        object.GetBuiltinByName(object.BuiltinBaseType, "_substr"),
	"index_of":       object.GetBuiltinByName(object.BuiltinBaseType, "index_of"),
	"_center":        object.GetBuiltinByName(object.BuiltinBaseType, "_center"),
	"_ljust":         object.GetBuiltinByName(object.BuiltinBaseType, "_ljust"),
	"_rjust":         object.GetBuiltinByName(object.BuiltinBaseType, "_rjust"),
	"to_title":       object.GetBuiltinByName(object.BuiltinBaseType, "to_title"),
	"to_kebab":       object.GetBuiltinByName(object.BuiltinBaseType, "to_kebab"),
	"to_camel":       object.GetBuiltinByName(object.BuiltinBaseType, "to_camel"),
	"to_snake":       object.GetBuiltinByName(object.BuiltinBaseType, "to_snake"),
	"matches":        object.GetBuiltinByName(object.BuiltinBaseType, "matches"),
})
