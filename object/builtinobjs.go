package object

import (
	"blue/consts"
	"os"
	"strings"
)

type BuiltinObjMapType map[string]*BuiltinObj

const BuiltinobjsModuleIndex = 255

const EnvBuiltinobjsListIndex = 0

var BuiltinobjsList = []struct {
	Name    string
	Builtin *BuiltinObj
}{
	{
		Name:    "ENV",
		Builtin: &BuiltinObj{Obj: PopulateENVObj()},
	},
	{
		Name:    "ARGV",
		Builtin: &BuiltinObj{Obj: populateARGVObj()},
	},
	{
		Name:    "STDIN",
		Builtin: &BuiltinObj{Obj: &Stringo{Value: os.Stdin.Name()}},
	},
	{
		Name:    "STDERR",
		Builtin: &BuiltinObj{Obj: &Stringo{Value: os.Stderr.Name()}},
	},
	{
		Name:    "STDOUT",
		Builtin: &BuiltinObj{Obj: &Stringo{Value: os.Stdout.Name()}},
	},
	{
		Name:    "FSTDIN",
		Builtin: &BuiltinObj{Obj: NewGoObj(os.Stdin)},
	},
	{
		Name:    "FSTDERR",
		Builtin: &BuiltinObj{Obj: NewGoObj(os.Stderr)},
	},
	{
		Name:    "FSTDOUT",
		Builtin: &BuiltinObj{Obj: NewGoObj(os.Stderr)},
	},
	{
		Name:    "VERSION",
		Builtin: &BuiltinObj{Obj: &Stringo{Value: consts.VERSION}},
	},
}

func getBuiltinobjByName(name string) *BuiltinObj {
	for _, bo := range BuiltinobjsList {
		if bo.Name == name {
			return bo.Builtin
		}
	}
	panic("Unhandled builtinobj lookup for name " + name)
}

var Builtinobjs = BuiltinObjMapType{
	"ENV":     getBuiltinobjByName("ENV"),
	"ARGV":    getBuiltinobjByName("ARGV"),
	"STDIN":   getBuiltinobjByName("STDIN"),
	"STDERR":  getBuiltinobjByName("STDERR"),
	"STDOUT":  getBuiltinobjByName("STDOUT"),
	"FSTDIN":  getBuiltinobjByName("FSTDIN"),
	"FSTDERR": getBuiltinobjByName("FSTDERR"),
	"FSTDOUT": getBuiltinobjByName("FSTDOUT"),
	"VERSION": getBuiltinobjByName("VERSION"),
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
