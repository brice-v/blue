package object

import (
	"testing"
)

func TestCloneListCycle(t *testing.T) {
	l := &List{}
	l.Elements = []Object{l}

	cloned, ok := l.Clone().(*List)
	if !ok {
		t.Fatalf("Clone() = %T, want *List", l.Clone())
	}
	if len(cloned.Elements) != 1 {
		t.Fatalf("cloned list has %d elements, want 1", len(cloned.Elements))
	}
	if cloned.Elements[0] != cloned {
		t.Fatal("cloned list does not point to itself")
	}
	if l.Elements[0] != l {
		t.Fatal("original list was mutated by Clone")
	}
}

func TestCloneMapCycle(t *testing.T) {
	m := &Map{Pairs: NewPairsMap()}
	key := &Stringo{Value: "self"}
	m.Set(key, m)

	cloned, ok := m.Clone().(*Map)
	if !ok {
		t.Fatalf("Clone() = %T, want *Map", m.Clone())
	}
	pair, found := cloned.Get(&Stringo{Value: "self"})
	if !found {
		t.Fatal("cloned map is missing the self key")
	}
	if pair.Value != cloned {
		t.Fatal("cloned map does not point to itself")
	}
	pair, found = m.Get(&Stringo{Value: "self"})
	if !found || pair.Value != m {
		t.Fatal("original map was mutated by Clone")
	}
}

func TestCloneSlicePreservesSharedIdentity(t *testing.T) {
	shared := &List{Elements: []Object{NewInteger(1)}}
	src := []Object{shared, shared}
	cloned := CloneSlice(src)

	first, ok := cloned[0].(*List)
	if !ok {
		t.Fatalf("cloned[0] = %T, want *List", cloned[0])
	}
	second, ok := cloned[1].(*List)
	if !ok {
		t.Fatalf("cloned[1] = %T, want *List", cloned[1])
	}
	if first != second {
		t.Fatal("shared element was cloned twice, identity was not preserved")
	}
	if first == shared {
		t.Fatal("shared element was not cloned")
	}
}

func TestCloneGlobalsCrossSlotCycle(t *testing.T) {
	list := &List{}
	m := &Map{Pairs: NewPairsMap()}
	m.Set(&Stringo{Value: "list"}, list)
	list.Elements = []Object{m}

	globals := []Object{list, m, NewInteger(7)}
	cloned := CloneGlobals(globals, 0)

	clonedList, ok := cloned[0].(*List)
	if !ok {
		t.Fatalf("cloned[0] = %T, want *List", cloned[0])
	}
	clonedMap, ok := cloned[1].(*Map)
	if !ok {
		t.Fatalf("cloned[1] = %T, want *Map", cloned[1])
	}
	if clonedList == list || clonedMap == m {
		t.Fatal("globals were not cloned")
	}
	if len(clonedList.Elements) != 1 || clonedList.Elements[0] != clonedMap {
		t.Fatal("cross-slot list -> map cycle was not preserved")
	}
	pair, found := clonedMap.Get(&Stringo{Value: "list"})
	if !found || pair.Value != clonedList {
		t.Fatal("cross-slot map -> list cycle was not preserved")
	}
}

func TestCloneGlobalsAboveHighWater(t *testing.T) {
	globals := make([]Object, 8)
	globals[0] = NewInteger(1)
	globals[5] = &List{Elements: []Object{NewInteger(2)}}

	cloned := CloneGlobals(globals, 1)
	if len(cloned) < 6 {
		t.Fatalf("cloned globals length = %d, want at least 6", len(cloned))
	}
	if cloned[5] == nil {
		t.Fatal("global above high water was dropped")
	}
	if cloned[5] == globals[5] {
		t.Fatal("global above high water was not cloned")
	}
}

func TestInspectCycles(t *testing.T) {
	list := &List{}
	list.Elements = []Object{list}
	if got := list.Inspect(); got != "[[...]]" {
		t.Errorf("list cycle Inspect() = %q, want %q", got, "[[...]]")
	}

	m := &Map{Pairs: NewPairsMap()}
	m.Set(&Stringo{Value: "self"}, m)
	if got := m.Inspect(); got != "{self: {...:...}}" {
		t.Errorf("map cycle Inspect() = %q, want %q", got, "{self: {...:...}}")
	}

	bs := &BlueStruct{Fields: []string{"self"}, Values: []Object{nil}}
	bs.Values[0] = bs
	if got := bs.Inspect(); got != "@{self: @{...}}" {
		t.Errorf("struct cycle Inspect() = %q, want %q", got, "@{self: @{...}}")
	}
}

func TestHashCycles(t *testing.T) {
	list := &List{}
	list.Elements = []Object{list}
	first := HashObject(list)
	second := HashObject(list)
	if first != second {
		t.Fatalf("cyclic list hash is not deterministic: %d != %d", first, second)
	}

	m := &Map{Pairs: NewPairsMap()}
	m.Set(&Stringo{Value: "self"}, m)
	first = HashObject(m)
	second = HashObject(m)
	if first != second {
		t.Fatalf("cyclic map hash is not deterministic: %d != %d", first, second)
	}
}

func TestFastEqualCycles(t *testing.T) {
	a := &List{}
	a.Elements = []Object{a}
	b := &List{}
	b.Elements = []Object{b}
	if !FastEqual(a, b) {
		t.Fatal("two isomorphic list cycles should be equal")
	}

	different := &List{Elements: []Object{NewInteger(1)}}
	if FastEqual(a, different) {
		t.Fatal("a cycle and a non-cycle with different contents should not be equal")
	}
}

func TestToJSONCycleReturnsError(t *testing.T) {
	list := &List{}
	list.Elements = []Object{list}
	got := BlueObjToJsonObject(list)
	if _, ok := got.(*Error); !ok {
		t.Fatalf("BlueObjToJsonObject(cycle) = %T, want *Error", got)
	}
}

func TestBuiltinInspectUsesName(t *testing.T) {
	builtin := &Builtin{Name: "println"}
	if got, want := builtin.Inspect(), "builtin_println"; got != want {
		t.Errorf("Builtin.Inspect() = %q, want %q", got, want)
	}
	if got, want := newInspector().inspect(builtin), "builtin_println"; got != want {
		t.Errorf("Inspector.inspect(Builtin) = %q, want %q", got, want)
	}

	builtinObj := &BuiltinObj{Obj: NewInteger(1)}
	if got, want := builtinObj.Inspect(), "builtin object"; got != want {
		t.Errorf("BuiltinObj.Inspect() = %q, want %q", got, want)
	}
}

func TestBuiltinsAreHashable(t *testing.T) {
	static := GetBuiltinByName(BuiltinCryptoType, "_sha")
	if static == nil {
		t.Fatal("static _sha builtin not found")
	}
	if static.Module != "crypto" {
		t.Fatalf("static _sha module = %q, want %q", static.Module, "crypto")
	}
	if !IsHashable(static) {
		t.Fatal("Builtin should be hashable")
	}
	first := HashObject(static)
	second := HashObject(static)
	if first != second {
		t.Fatalf("Builtin hash is not deterministic: %d != %d", first, second)
	}

	resolved := &Builtin{Name: "_sha", Module: "crypto"}
	if HashObject(static) != HashObject(resolved) {
		t.Fatal("builtins with the same module and name should hash the same")
	}
	if !FastEqual(static, resolved) {
		t.Fatal("builtins with the same module and name should be equal")
	}

	differentModule := &Builtin{Name: "_sha", Module: "other"}
	if FastEqual(static, differentModule) {
		t.Fatal("builtins with different modules should not be equal")
	}

	anonymous := &Builtin{Name: "println"}
	otherAnonymous := &Builtin{Name: "println"}
	if FastEqual(anonymous, otherAnonymous) {
		t.Fatal("dynamic builtins without module identity should remain pointer-distinct")
	}

	builtinObj := Builtinobjs["ENV"]
	if builtinObj.Module != "builtinobj" || builtinObj.Name != "ENV" {
		t.Fatalf("ENV identity = %q/%q, want builtinobj/ENV", builtinObj.Module, builtinObj.Name)
	}
	if !IsHashable(builtinObj) {
		t.Fatal("BuiltinObj should be hashable")
	}
	_ = HashObject(builtinObj)
}
