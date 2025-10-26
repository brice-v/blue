package object

import (
	"blue/consts"
	"os"
	"strings"
)

type BuiltinObjMapType map[string]*BuiltinObj

var Builtinobjs = BuiltinObjMapType{
	"ENV": {
		Obj: PopulateENVObj(),
	},
	"ARGV": {
		Obj: populateARGVObj(),
	},
	"STDIN": {
		Obj: &Stringo{Value: os.Stdin.Name()},
	},
	"STDERR": {
		Obj: &Stringo{Value: os.Stderr.Name()},
	},
	"STDOUT": {
		Obj: &Stringo{Value: os.Stdout.Name()},
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
		Obj: &Stringo{Value: consts.VERSION},
	},
}

func PopulateENVObj() *Map {
	m := Map{
		Pairs: NewPairsMap(),
	}

	for _, e := range os.Environ() {
		es := strings.Split(e, "=")
		e1, e2 := es[0], es[1]
		key := &Stringo{Value: e1}
		hashKey := HashObject(key)
		hk := HashKey{
			Type:  STRING_OBJ,
			Value: hashKey,
		}
		m.Pairs.Set(hk, MapPair{
			Key:   key,
			Value: &Stringo{Value: e2},
		})
	}
	return &m
}

func populateARGVObj() *List {
	l := &List{
		Elements: make([]Object, len(os.Args)),
	}
	for i, e := range os.Args {
		value := &Stringo{Value: e}
		l.Elements[i] = value
	}
	return l
}
