package evaluator

import (
	"blue/consts"
	"blue/object"
	"os"
	"strings"
)

type BuiltinObjMapType map[string]*object.BuiltinObj

var builtinobjs = BuiltinObjMapType{
	"ENV": {
		Obj: populateENVObj(),
	},
	"ARGV": {
		Obj: populateARGVObj(),
	},
	"STDIN": {
		Obj: &object.Stringo{Value: os.Stdin.Name()},
	},
	"STDERR": {
		Obj: &object.Stringo{Value: os.Stderr.Name()},
	},
	"STDOUT": {
		Obj: &object.Stringo{Value: os.Stdout.Name()},
	},
	"FSTDIN": {
		Obj: NewGoObj(os.Stdin),
	},
	"FSTDERR": {
		Obj: NewGoObj(os.Stderr),
	},
	"FSTDOUT": {
		Obj: NewGoObj(os.Stderr),
	},
	"VERSION": {
		Obj: &object.Stringo{Value: consts.VERSION},
	},
}

func populateENVObj() *object.Map {
	m := object.Map{
		Pairs: object.NewPairsMap(),
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
		m.Pairs.Set(hk, object.MapPair{
			Key:   key,
			Value: &object.Stringo{Value: e2},
		})
	}
	return &m
}

func populateARGVObj() *object.List {
	l := &object.List{
		Elements: make([]object.Object, len(os.Args)),
	}
	for i, e := range os.Args {
		value := &object.Stringo{Value: e}
		l.Elements[i] = value
	}
	return l
}
