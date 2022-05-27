package evaluator

import (
	"blue/object"
	"os"
	"strings"
)

var builtinobjs = map[string]*object.BuiltinObj{
	"ENV": {
		Obj: populateENVObj(),
	},
	"ARGV": {
		Obj: populateARGVObj(),
	},
}

func populateENVObj() *object.Map {
	m := object.Map{
		Pairs: make(map[object.HashKey]object.MapPair),
	}

	for _, e := range os.Environ() {
		es := strings.Split(e, "=")
		e1, e2 := es[0], es[1]
		key := &object.Stringo{Value: e1}
		hashKey := object.HashObject(key)
		hk := object.HashKey{
			Type:  object.STRING_OBJ,
			Value: hashKey,
		}
		m.Pairs[hk] = object.MapPair{
			Key:   key,
			Value: &object.Stringo{Value: e2},
		}
	}
	return &m
}

func populateARGVObj() *object.List {
	l := &object.List{
		Elements: make([]object.Object, 0),
	}
	for _, e := range os.Args {
		value := &object.Stringo{Value: e}
		l.Elements = append(l.Elements, value)
	}
	return l
}
