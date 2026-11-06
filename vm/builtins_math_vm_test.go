package vm

import (
	"blue/object"
	"testing"
)

func mathTestInt(v int64) object.Object     { return &object.Integer{Value: v} }
func mathTestFloat(v float64) object.Object { return &object.Float{Value: v} }

func mathTestBuiltin(t *testing.T, machine *VM, name string) func(args ...object.Object) object.Object {
	t.Helper()
	return GetStdBuiltinWithVm("math", name, machine)
}

func TestMathRandStateIsPerVM(t *testing.T) {
	a := &VM{}
	b := &VM{}
	if a.mathRand != nil || b.mathRand != nil {
		t.Fatal("math rand should not be initialized before first use")
	}
	seedA := mathTestBuiltin(t, a, "_seed")
	seedB := mathTestBuiltin(t, b, "_seed")
	randA := mathTestBuiltin(t, a, "_rand")
	randB := mathTestBuiltin(t, b, "_rand")

	seedA(mathTestInt(42))
	seedB(mathTestInt(42))
	if a.mathRand == nil || b.mathRand == nil {
		t.Fatal("math rand should be initialized after seeding")
	}
	for i := range 5 {
		av := randA().(*object.Float).Value
		bv := randB().(*object.Float).Value
		if av != bv {
			t.Fatalf("draw %d: vm a = %v, vm b = %v, want identical seeded sequences", i, av, bv)
		}
	}
}

func TestMathRandomBuiltins(t *testing.T) {
	machine := &VM{}
	seed := mathTestBuiltin(t, machine, "_seed")
	randFn := mathTestBuiltin(t, machine, "_rand")
	gauss := mathTestBuiltin(t, machine, "_gauss")
	wc := mathTestBuiltin(t, machine, "_weighted_choice")

	if _, ok := seed(mathTestInt(42)).(*object.Boolean); !ok {
		t.Fatal("_seed should return a BOOLEAN")
	}
	for range 2 {
		res := randFn()
		r, ok := res.(*object.Float)
		if !ok {
			t.Fatalf("_rand returned %T, want *Float", res)
		}
		if r.Value < 0 || r.Value >= 1 {
			t.Fatalf("_rand value %v out of range [0,1)", r.Value)
		}
	}
	if _, ok := gauss(mathTestFloat(0), mathTestFloat(1)).(*object.Float); !ok {
		t.Fatal("_gauss should return a FLOAT")
	}

	items := &object.List{Elements: []object.Object{mathTestInt(10), mathTestInt(20), mathTestInt(30)}}
	weights := &object.List{Elements: []object.Object{mathTestFloat(1), mathTestFloat(1), mathTestFloat(1)}}
	if _, ok := wc(items, weights).(*object.Integer); !ok {
		t.Fatal("_weighted_choice should return an item")
	}
	if _, ok := wc(items, &object.List{Elements: []object.Object{mathTestInt(1)}}).(*object.Error); !ok {
		t.Fatal("_weighted_choice with mismatched lengths should error")
	}
	zeroWeights := &object.List{Elements: []object.Object{mathTestFloat(0), mathTestFloat(0), mathTestFloat(0)}}
	if _, ok := wc(items, zeroWeights).(*object.Error); !ok {
		t.Fatal("_weighted_choice with zero weight sum should error")
	}

	if _, ok := randFn(mathTestInt(1)).(*object.Error); !ok {
		t.Fatal("_rand with arguments should error")
	}
	if _, ok := seed().(*object.Error); !ok {
		t.Fatal("_seed with no arguments should error")
	}
	if _, ok := seed(mathTestFloat(1)).(*object.Error); !ok {
		t.Fatal("_seed with a non-integer should error")
	}
	if _, ok := gauss(mathTestFloat(0)).(*object.Error); !ok {
		t.Fatal("_gauss with one argument should error")
	}
	if _, ok := wc(items).(*object.Error); !ok {
		t.Fatal("_weighted_choice with one argument should error")
	}
}
