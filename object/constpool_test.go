package object

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestValidateReservedPrefix(t *testing.T) {
	t.Run("valid pool", func(t *testing.T) {
		pool := NewObjectConstants()
		pool = append(pool, NewInteger(1))
		if err := ValidateReservedPrefix(pool); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("pool too small", func(t *testing.T) {
		if err := ValidateReservedPrefix(NewObjectConstants()[:2]); err == nil {
			t.Error("expected error for undersized pool")
		} else if !strings.Contains(err.Error(), "constant pool too small") {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("empty pool", func(t *testing.T) {
		if err := ValidateReservedPrefix(nil); err == nil {
			t.Error("expected error for empty pool")
		}
	})
	t.Run("wrong object in slot", func(t *testing.T) {
		pool := NewObjectConstants()
		pool[0] = NewInteger(1)
		err := ValidateReservedPrefix(pool)
		if err == nil {
			t.Fatal("expected error for wrong reserved object")
		}
		if !strings.Contains(err.Error(), "reserved constant slot 0") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestConstantPoolRoundTrip(t *testing.T) {
	m := makeTestMap([]Object{NewString("k")}, []Object{&Float{Value: 1.25}})
	set := makeTestSet(NewString("s1"), NewString("s2"))
	bs := &BlueStruct{Fields: []string{"f"}, Values: []Object{NewString("v")}}

	original := NewObjectConstants()
	original = append(original,
		NewInteger(42),
		NewString("hello"),
		makeTestCompiledFunction(),
		&List{Elements: []Object{NewInteger(1), m}},
		set,
		bs,
		NewGoObj([]string{"a", "b"}),
	)

	data, err := EncodeConstantPool(original)
	if err != nil {
		t.Fatalf("EncodeConstantPool failed: %v", err)
	}

	decoded, err := DecodeConstantPool(data)
	if err != nil {
		t.Fatalf("DecodeConstantPool failed: %v", err)
	}

	if len(decoded) != len(original) {
		t.Fatalf("len(decoded) = %d, want %d", len(decoded), len(original))
	}

	for i := range OBJECT_CONSTANTS {
		if decoded[i] != OBJECT_CONSTANTS[i] {
			t.Errorf("reserved slot %d identity broken: got %T", i, decoded[i])
		}
	}

	for i := len(OBJECT_CONSTANTS); i < len(original); i++ {
		want, have := original[i], decoded[i]
		if want.Type() != have.Type() {
			t.Errorf("constant %d: type = %s, want %s", i, have.Type(), want.Type())
			continue
		}
		if cf, isCF := want.(*CompiledFunction); isCF {
			haveCF := have.(*CompiledFunction)
			if string(haveCF.Instructions) != string(cf.Instructions) || haveCF.DisplayString != cf.DisplayString {
				t.Errorf("constant %d (compiled fn) mismatched", i)
			}
			continue
		}
		if gf, isGF := want.(*GoObj[[]string]); isGF {
			haveGF := have.(*GoObj[[]string])
			if len(haveGF.Value) != len(gf.Value) {
				t.Errorf("constant %d (struct fields) mismatched: %#v", i, haveGF.Value)
			}
			continue
		}
		if want.Inspect() != have.Inspect() {
			t.Errorf("constant %d: Inspect() = %q, want %q", i, have.Inspect(), want.Inspect())
		}
	}
}

func TestEncodeConstantPoolErrors(t *testing.T) {
	t.Run("invalid prefix rejected", func(t *testing.T) {
		bad := []Object{NewInteger(9)}
		if _, err := EncodeConstantPool(bad); err == nil {
			t.Error("expected prefix validation error")
		}
	})
	t.Run("unserializable constant reports index", func(t *testing.T) {
		pool := NewObjectConstants()
		pool = append(pool, NewInteger(1), &Closure{Fun: makeTestCompiledFunction()})
		_, err := EncodeConstantPool(pool)
		if err == nil {
			t.Fatal("expected encode failure for closure constant")
		}
		wantIdx := len(OBJECT_CONSTANTS) + 1
		if !strings.Contains(err.Error(), fmt.Sprintf("constant %d", wantIdx)) {
			t.Errorf("error %q should mention constant index %d", err, wantIdx)
		}
	})
}

func TestDecodeConstantPoolGarbage(t *testing.T) {
	if _, err := DecodeConstantPool([]byte("definitely not cbor")); err == nil {
		t.Error("expected decode error for garbage input")
	} else if !strings.Contains(err.Error(), "failed to decode constant pool") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDecodeConstantPoolBadConstant(t *testing.T) {
	badWrapper := ObjectWrapper{Type: iType(200)}
	data, err := cbor.Marshal([]ObjectWrapper{badWrapper})
	if err != nil {
		t.Fatal(err)
	}
	_, err = DecodeConstantPool(data)
	if err == nil {
		t.Fatal("expected error decoding bad constant")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("failed to decode constant %d", len(OBJECT_CONSTANTS))) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFindUnserializableConstant(t *testing.T) {
	t.Run("clean pool", func(t *testing.T) {
		pool := NewObjectConstants()
		pool = append(pool, NewInteger(1), NewString("x"))
		idx, err := FindUnserializableConstant(pool)
		if idx != -1 || err != nil {
			t.Errorf("got (%d, %v), want (-1, nil)", idx, err)
		}
	})
	t.Run("closure found at correct index", func(t *testing.T) {
		pool := NewObjectConstants()
		pool = append(pool, NewInteger(1), &Closure{}, NewString("after"))
		idx, err := FindUnserializableConstant(pool)
		wantIdx := len(OBJECT_CONSTANTS) + 1
		if idx != wantIdx {
			t.Fatalf("idx = %d, want %d", idx, wantIdx)
		}
		if err == nil {
			t.Error("expected error describing why")
		}
	})
	t.Run("reserved-only pool", func(t *testing.T) {
		idx, err := FindUnserializableConstant(NewObjectConstants())
		if idx != -1 || err != nil {
			t.Errorf("got (%d, %v), want (-1, nil)", idx, err)
		}
	})
}

func TestDebugDumpConstants(t *testing.T) {
	pool := NewObjectConstants()
	pool = append(pool, nil, NewInteger(7))
	dump := DebugDumpConstants(pool)
	lines := strings.Split(strings.TrimRight(dump, "\n"), "\n")
	if len(lines) != len(pool) {
		t.Fatalf("dump has %d lines, want %d:\n%s", len(lines), len(pool), dump)
	}
	if lines[len(OBJECT_CONSTANTS)] != fmt.Sprintf("%d: <nil>", len(OBJECT_CONSTANTS)) {
		t.Errorf("nil entry not rendered as <nil>:\n%s", dump)
	}
	if !strings.Contains(dump, "7") {
		t.Errorf("integer inspect missing from dump:\n%s", dump)
	}
}
