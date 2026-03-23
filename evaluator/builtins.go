package evaluator

import (
	"blue/object"
	"fmt"
)

type helpStrArgs struct {
	explanation string
	signature   string
	errors      string
	example     string
}

func (hsa helpStrArgs) String() string {
	return fmt.Sprintf("%s\n    Signature:  %s\n    Error(s):   %s\n    Example(s): %s\n", hsa.explanation, hsa.signature, hsa.errors, hsa.example)
}

type BuiltinMapType struct {
	*object.ConcurrentMap[string, *object.Builtin]
}

func NewBuiltinObjMap(input map[string]*object.Builtin) BuiltinMapType {
	return BuiltinMapType{&object.ConcurrentMap[string, *object.Builtin]{
		Kv: input,
	}}
}

type BuiltinMapTypeInternal map[string]*object.Builtin

func NewBuiltinMap(builtins object.NewBuiltinSliceType) BuiltinMapType {
	m := NewBuiltinObjMap(BuiltinMapTypeInternal{})
	for _, o := range builtins {
		m.Put(o.Name, o.Builtin)
	}
	return m
}

var builtins = NewBuiltinMap(object.Builtins)

func GetBuiltins(e *Evaluator) BuiltinMapType {
	b := builtins
	b.Put("to_num", createToNumBuiltin(e))
	b.Put("_sort", createSortBuiltin(e))
	b.Put("_sorted", createSortedBuiltin(e))
	b.Put("all", createAllBuiltin(e))
	b.Put("any", createAnyBuiltin(e))
	b.Put("map", createMapBuiltin(e))
	b.Put("filter", createFilterBuiltin(e))
	b.Put("load", createLoadBuiltin(e))
	return b
}
