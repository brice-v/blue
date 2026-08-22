package object

import (
	"math"
	"strings"
	"testing"
)

func mathBuiltinFn(t *testing.T, name string) BuiltinFunction {
	t.Helper()
	for _, b := range MathBuiltins {
		if b.Name == name {
			if b.Fun == nil {
				t.Fatalf("math builtin %q has nil Fun", name)
			}
			return b.Fun
		}
	}
	t.Fatalf("math builtin %q not found", name)
	return nil
}

func fl(v float64) Object { return &Float{Value: v} }
func in(v int64) Object   { return &Integer{Value: v} }

func TestMathRegistry(t *testing.T) {
	seen := make(map[string]bool)
	for _, b := range MathBuiltins {
		if b.Name == "" || b.HelpStr == "" {
			t.Fatalf("math builtin %q missing Name or HelpStr", b.Name)
		}
		if seen[b.Name] {
			t.Errorf("duplicate math builtin name %q", b.Name)
		}
		seen[b.Name] = true
		if b.Fun == nil {
			t.Errorf("math builtin %q has nil Fun", b.Name)
		}
	}
}

func TestMathUnaryFloatBuiltins(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		goFn func(float64) float64
	}{
		{"_acos", 0.5, math.Acos},
		{"_acosh", 1.5, math.Acosh},
		{"_asin", 0.5, math.Asin},
		{"_asinh", 0.5, math.Asinh},
		{"_atan", 0.5, math.Atan},
		{"_atanh", 0.5, math.Atanh},
		{"_cbrt", 27, math.Cbrt},
		{"_ceil", 1.2, math.Ceil},
		{"_cos", 1.2, math.Cos},
		{"_cosh", 1.2, math.Cosh},
		{"_erf", 0.7, math.Erf},
		{"_erfc", 0.7, math.Erfc},
		{"_erfcinv", 0.5, math.Erfcinv},
		{"_erfinv", 0.5, math.Erfinv},
		{"_floor", 1.8, math.Floor},
		{"_gamma", 5, math.Gamma},
		{"_j0", 1.2, math.J0},
		{"_j1", 1.2, math.J1},
		{"_log", 10, math.Log},
		{"_log10", 120, math.Log10},
		{"_log1p", 0.2, math.Log1p},
		{"_log2", 8, math.Log2},
		{"_logb", 0.2, math.Logb},
		{"_round", 3.5, math.Round},
		{"_round_to_even", 3.5, math.RoundToEven},
		{"_sin", 0.5, math.Sin},
		{"_sinh", 0.5, math.Sinh},
		{"_tan", 0.5, math.Tan},
		{"_tanh", 0.5, math.Tanh},
		{"_trunc", 2.5, math.Trunc},
		{"_y0", 2, math.Y0},
		{"_y1", 2, math.Y1},
	}
	for _, tt := range tests {
		res := mathBuiltinFn(t, tt.name)(fl(tt.in))
		f, ok := res.(*Float)
		if !ok {
			t.Errorf("%s(%v) returned %T, want *Float", tt.name, tt.in, res)
			continue
		}
		want := tt.goFn(tt.in)
		if f.Value != want {
			t.Errorf("%s(%v) = %v, want %v", tt.name, tt.in, f.Value, want)
		}
	}
	spotChecks := []struct {
		name string
		in   float64
		want float64
	}{
		{"_cbrt", 8, 2},
		{"_floor", 1.9, 1},
		{"_ceil", 1.2, 2},
		{"_trunc", 2.9, 2},
		{"_log2", 8, 3},
		{"_log10", 100, 2},
		{"_round", 3.5, 4},
		{"_round_to_even", 4.5, 4},
	}
	for _, sc := range spotChecks {
		res := mathBuiltinFn(t, sc.name)(fl(sc.in)).(*Float)
		if res.Value != sc.want {
			t.Errorf("%s(%v) = %v, want %v", sc.name, sc.in, res.Value, sc.want)
		}
	}
}

func TestMathBinaryFloatBuiltins(t *testing.T) {
	tests := []struct {
		name string
		x, y float64
		goFn func(float64, float64) float64
	}{
		{"_atan2", 0.4, 0.8, math.Atan2},
		{"_copysign", 3, -1, math.Copysign},
		{"_dim", 3.4, 1.2, math.Dim},
		{"_hypot", 3, 4, math.Hypot},
		{"_next_after", 1, 2, math.Nextafter},
		{"_remainder", 98.2, 38.3, math.Remainder},
	}
	for _, tt := range tests {
		res := mathBuiltinFn(t, tt.name)(fl(tt.x), fl(tt.y))
		f, ok := res.(*Float)
		if !ok {
			t.Errorf("%s(%v,%v) returned %T, want *Float", tt.name, tt.x, tt.y, res)
			continue
		}
		want := tt.goFn(tt.x, tt.y)
		if f.Value != want {
			t.Errorf("%s(%v,%v) = %v, want %v", tt.name, tt.x, tt.y, f.Value, want)
		}
	}
	if got := mathBuiltinFn(t, "_hypot")(fl(3), fl(4)); got.Inspect() != "5.0" {
		t.Errorf("_hypot(3,4) = %s, want 5.0", got.Inspect())
	}
}

func TestMathMixedArgBuiltins(t *testing.T) {
	if got := mathBuiltinFn(t, "_fma")(fl(2), fl(3), fl(4)); got.Inspect() != "10.0" {
		t.Errorf("_fma(2,3,4) = %s, want 10.0", got.Inspect())
	}
	jnRes := mathBuiltinFn(t, "_jn")(fl(1.2), in(2))
	if f := jnRes.(*Float); f.Value != math.Jn(2, 1.2) {
		t.Errorf("_jn(1.2,2) = %v, want %v", f.Value, math.Jn(2, 1.2))
	}
	ynRes := mathBuiltinFn(t, "_yn")(fl(3), in(2))
	if f := ynRes.(*Float); f.Value != math.Yn(2, 3) {
		t.Errorf("_yn(3,2) = %v, want %v", f.Value, math.Yn(2, 3))
	}
	if got := mathBuiltinFn(t, "_ldexp")(fl(0.75), in(2)); got.Inspect() != "3.0" {
		t.Errorf("_ldexp(0.75,2) = %s, want 3.0", got.Inspect())
	}
	if got := mathBuiltinFn(t, "_ilogb")(fl(203)); got.Inspect() != "7" {
		t.Errorf("_ilogb(203) = %s, want 7", got.Inspect())
	}
	if _, ok := mathBuiltinFn(t, "_ilogb")(fl(203)).(*Integer); !ok {
		t.Error("_ilogb should return an INTEGER")
	}
	posInf := mathBuiltinFn(t, "_inf")(in(1))
	negInf := mathBuiltinFn(t, "_inf")(in(-1))
	if pf := posInf.(*Float); !math.IsInf(pf.Value, 1) {
		t.Errorf("_inf(1) = %v, want +Inf", pf.Value)
	}
	if nf := negInf.(*Float); !math.IsInf(nf.Value, -1) {
		t.Errorf("_inf(-1) = %v, want -Inf", nf.Value)
	}
	if got := mathBuiltinFn(t, "_is_inf")(posInf, in(0)); got.Inspect() != "true" {
		t.Errorf("_is_inf(+Inf, 0) = %s, want true", got.Inspect())
	}
	if got := mathBuiltinFn(t, "_is_inf")(fl(1.5), in(0)); got.Inspect() != "false" {
		t.Errorf("_is_inf(1.5, 0) = %s, want false", got.Inspect())
	}
	nanObj := mathBuiltinFn(t, "_NaN")()
	if f := nanObj.(*Float); !math.IsNaN(f.Value) {
		t.Errorf("_NaN() = %v, want NaN", f.Value)
	}
	if got := mathBuiltinFn(t, "_is_NaN")(nanObj); got.Inspect() != "true" {
		t.Errorf("_is_NaN(NaN) = %s, want true", got.Inspect())
	}
	if got := mathBuiltinFn(t, "_is_NaN")(fl(1)); got.Inspect() != "false" {
		t.Errorf("_is_NaN(1) = %s, want false", got.Inspect())
	}
	if got := mathBuiltinFn(t, "_signbit")(fl(-3)); got.Inspect() != "true" {
		t.Errorf("_signbit(-3) = %s, want true", got.Inspect())
	}
	if got := mathBuiltinFn(t, "_signbit")(fl(3)); got.Inspect() != "false" {
		t.Errorf("_signbit(3) = %s, want false", got.Inspect())
	}
	r1 := mathBuiltinFn(t, "_rand")()
	r2 := mathBuiltinFn(t, "_rand")()
	for i, r := range []*Float{r1.(*Float), r2.(*Float)} {
		if r.Value < 0 || r.Value >= 1 {
			t.Errorf("_rand()[%d] out of range [0,1): %v", i, r.Value)
		}
	}
	if r1.Inspect() == r2.Inspect() {
		t.Error("_rand returned identical values twice")
	}
}

func mapGetString(t *testing.T, m *Map, key string) Object {
	t.Helper()
	hk := HashKey{Type: STRING_OBJ, Value: HashObject(&Stringo{Value: key})}
	pair, ok := m.Pairs.Get(hk)
	if !ok {
		t.Fatalf("key %q missing from map %s", key, m.Inspect())
	}
	return pair.Value
}

func TestMathMultiReturnBuiltins(t *testing.T) {
	fr := mathBuiltinFn(t, "_frexp")(fl(3)).(*Map)
	if fr.Pairs.Len() != 2 {
		t.Fatalf("_frexp returned %d pairs, want 2", fr.Pairs.Len())
	}
	if frac := mapGetString(t, fr, "frac").(*Float); frac.Value != 0.75 {
		t.Errorf("_frexp(3) frac = %v, want 0.75", frac.Value)
	}
	if exp := mapGetString(t, fr, "exp").(*Integer); exp.Value != 2 {
		t.Errorf("_frexp(3) exp = %d, want 2", exp.Value)
	}
	md := mathBuiltinFn(t, "_modf")(fl(10.1)).(*Map)
	wantI, wantFrac := math.Modf(10.1)
	if i := mapGetString(t, md, "i").(*Integer); i.Value != int64(wantI) {
		t.Errorf("_modf(10.1) i = %d, want %d", i.Value, int64(wantI))
	}
	if frac := mapGetString(t, md, "frac").(*Float); frac.Value != wantFrac {
		t.Errorf("_modf(10.1) frac = %v, want %v", frac.Value, wantFrac)
	}
	lg := mathBuiltinFn(t, "_lgamma")(fl(2.3)).(*Map)
	wantLg, wantSign := math.Lgamma(2.3)
	if lgv := mapGetString(t, lg, "lgamma").(*Float); lgv.Value != wantLg {
		t.Errorf("_lgamma(2.3) lgamma = %v, want %v", lgv.Value, wantLg)
	}
	if sg := mapGetString(t, lg, "sign").(*Integer); sg.Value != int64(wantSign) {
		t.Errorf("_lgamma(2.3) sign = %d, want %d", sg.Value, wantSign)
	}
	sc := mathBuiltinFn(t, "_sincos")(fl(0.5)).(*Map)
	wantSin, wantCos := math.Sincos(0.5)
	if sv := mapGetString(t, sc, "sin").(*Float); sv.Value != wantSin {
		t.Errorf("_sincos(0.5) sin = %v, want %v", sv.Value, wantSin)
	}
	if cv := mapGetString(t, sc, "cos").(*Float); cv.Value != wantCos {
		t.Errorf("_sincos(0.5) cos = %v, want %v", cv.Value, wantCos)
	}
}

func TestGcdLcmBuiltins(t *testing.T) {
	tests := []builtinTestCase{
		{name: "basic", args: []Object{in(12), in(18)}, want: "6"},
		{name: "coprime", args: []Object{in(9), in(28)}, want: "1"},
		{name: "zero", args: []Object{in(0), in(5)}, want: "5"},
		{name: "first arg not int", args: []Object{fl(12), in(18)}, err: "PositionalTypeError"},
		{name: "second arg not int", args: []Object{in(12), fl(18)}, err: "PositionalTypeError"},
		{name: "one arg", args: []Object{in(12)}, err: "InvalidArgCountError"},
	}
	runBuiltinTestsFor(t, MathBuiltins, "_gcd", tests)

	lcmTests := []builtinTestCase{
		{name: "two ints", args: []Object{in(4), in(6)}, want: "12"},
		{name: "three ints", args: []Object{in(4), in(6), in(8)}, want: "24"},
		{name: "list form", args: []Object{&List{Elements: intObjs(3, 4, 5)}}, want: "60"},
		{name: "two-element list", args: []Object{&List{Elements: intObjs(4, 6)}}, want: "12"},
		{name: "list too short", args: []Object{&List{Elements: intObjs(4)}}, err: "`lcm` error: list must be at least 2 elements long"},
		{name: "list bad element", args: []Object{&List{Elements: []Object{in(4), fl(6)}}}, err: "`lcm` error: all elements in list need to be INTEGER"},
		{name: "single int", args: []Object{in(4)}, err: "InvalidArgCountError"},
		{name: "non-list non-int", args: []Object{fl(4), fl(6)}, err: "PositionalTypeError"},
	}
	runBuiltinTestsFor(t, MathBuiltins, "_lcm", lcmTests)

	gcdFn := func(a, b int64) int64 { return gcd(a, b) }
	if gcdFn(48, 36) != 12 {
		t.Errorf("gcd(48,36) = %d, want 12", gcd(48, 36))
	}
	if lcm(2, 3, 4, 5) != 60 {
		t.Errorf("lcm(2,3,4,5) = %d, want 60", lcm(2, 3, 4, 5))
	}
}

func TestMathErrorPaths(t *testing.T) {
	for _, b := range MathBuiltins {
		switch b.Name {
		case "_rand", "_NaN":
			if _, ok := b.Fun(in(1)).(*Error); !ok {
				t.Errorf("%s() with extra args should error", b.Name)
			}
			continue
		}
		res := b.Fun()
		if _, ok := res.(*Error); !ok {
			t.Errorf("%s() with no args should error, got %T %s", b.Name, res, res.Inspect())
		}
		if b.Name == "_inf" {
			continue
		}
		res = b.Fun(in(1))
		if _, ok := res.(*Error); !ok {
			t.Errorf("%s(INTEGER) should error, got %T %s", b.Name, res, res.Inspect())
		}
	}
	for _, name := range []string{"_atan2", "_copysign", "_dim", "_hypot", "_next_after", "_remainder"} {
		res := mathBuiltinFn(t, name)(fl(1), in(2))
		if _, ok := res.(*Error); !ok {
			t.Errorf("%s(float, INTEGER) should error, got %s", name, res.Inspect())
		}
	}
	for _, name := range []string{"_jn", "_yn", "_ldexp", "_is_inf"} {
		res := mathBuiltinFn(t, name)(fl(1), fl(2))
		if _, ok := res.(*Error); !ok {
			t.Errorf("%s(float, float-as-int slot) should error, got %s", name, res.Inspect())
		}
	}
	res := mathBuiltinFn(t, "_fma")(fl(1), fl(2), in(3))
	if _, ok := res.(*Error); !ok {
		t.Errorf("_fma with INTEGER third arg should error, got %s", res.Inspect())
	}
	unarySpot := []struct {
		name string
		arg  Object
	}{
		{"_acos", in(1)},
		{"_cbrt", &Stringo{Value: "x"}},
		{"_log", TRUE},
		{"_is_NaN", in(1)},
		{"_signbit", &Stringo{Value: "-3"}},
		{"_inf", fl(1)},
		{"_ilogb", in(1)},
	}
	for _, tc := range unarySpot {
		res := mathBuiltinFn(t, tc.name)(tc.arg)
		if _, ok := res.(*Error); !ok {
			t.Errorf("%s(bad type) should error, got %s", tc.name, res.Inspect())
		}
	}
	binarySpot := []struct {
		name string
		args []Object
	}{
		{"_atan2", []Object{fl(1), in(1)}},
		{"_copysign", []Object{in(1), fl(1)}},
		{"_dim", []Object{fl(1), &Stringo{Value: "a"}}},
		{"_hypot", []Object{fl(1)}},
		{"_remainder", []Object{}},
		{"_fma", []Object{fl(1), fl(2)}},
		{"_jn", []Object{fl(1), fl(2)}},
		{"_yn", []Object{fl(1), fl(2)}},
		{"_ldexp", []Object{fl(1), fl(2)}},
		{"_is_inf", []Object{fl(1), fl(1)}},
		{"_next_after", []Object{fl(1), in(2)}},
	}
	for _, tc := range binarySpot {
		res := mathBuiltinFn(t, tc.name)(tc.args...)
		errObj, ok := res.(*Error)
		if !ok {
			t.Errorf("%s(bad args) should error, got %s", tc.name, res.Inspect())
			continue
		}
		if !strings.Contains(errObj.Message, "Error") {
			t.Errorf("%s(bad args) unexpected message: %s", tc.name, errObj.Message)
		}
	}
}
