package object

import (
	"math"
	mr "math/rand"
)

// greatest common divisor (GCD) via Euclidean algorithm
func gcd(a, b int64) int64 {
	for b != 0 {
		t := b
		b = a % b
		a = t
	}
	return a
}

// find Least Common Multiple (LCM) via GCD
func lcm(a, b int64, integers ...int64) int64 {
	result := a * b / gcd(a, b)
	for i := range integers {
		result = lcm(result, integers[i])
	}
	return result
}

var MathBuiltins = NewBuiltinSliceType{
	{Name: "_rand", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("rand", len(args), 0, "")
			}
			return &Float{Value: mr.Float64()}
		},
		HelpStr: helpStrArgs{
			explanation: "`rand` returns a FLOAT a pseudo-random number in the half-open interval [0.0,1.0)",
			signature:   "rand() -> float",
			errors:      "InvalidArgCount",
			example:     "rand() => 0.125215",
		}.String(),
	}},
	{Name: "_NaN", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 0 {
				return newInvalidArgCountError("NaN", len(args), 0, "")
			}
			return &Float{Value: math.NaN()}
		},
		HelpStr: helpStrArgs{
			explanation: "`NaN` is the representation of NaN",
			signature:   "NaN() -> NaN",
			errors:      "InvalidArgCount",
			example:     "NaN() => NaN",
		}.String(),
	}},
	{Name: "_acos", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("acos", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("acos", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Acos(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`acos` returns the arccosine, in radians, of x",
			signature:   "acos(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "acos(0.5) => 1.047198",
		}.String(),
	}},
	{Name: "_acosh", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("acosh", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("acosh", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Acosh(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`acosh` returns the inverse hyperbolic cosine of x",
			signature:   "acosh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "acosh(1.04) => 0.281908",
		}.String(),
	}},
	{Name: "_asin", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("asin", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("asin", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Asin(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`asin` returns the arcsine, in radians, of x",
			signature:   "asin(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "asin(0.4) => 0.411517",
		}.String(),
	}},
	{Name: "_asinh", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("asinh", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("asinh", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Asinh(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`asinh` returns the inverse hyperbolic sine of x",
			signature:   "asinh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "asinh(0.4) => 0.390035",
		}.String(),
	}},
	{Name: "_atan", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("atan", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("atan", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Atan(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`atan` returns the arctangent, in radians, of x",
			signature:   "atan(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "atan(0.4) => 0.380506",
		}.String(),
	}},
	{Name: "_atan2", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("atan2", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("atan2", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("atan2", 2, FLOAT_OBJ, args[1].Type())
			}
			x := args[0].(*Float).Value
			y := args[1].(*Float).Value
			return &Float{Value: math.Atan2(x, y)}
		},
		HelpStr: helpStrArgs{
			explanation: "`atan2` returns the arc tangent of y/x, using the signs of the two to determine the quadrant of the return value",
			signature:   "atan2(x: float, y: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "atan2(0.4,0.4) => 0.785398",
		}.String(),
	}},
	{Name: "_atanh", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("atanh", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("atanh", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Atanh(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`atanh` returns the inverse hyperbolic tangent of x",
			signature:   "atanh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "atanh(0.4) => 0.423649",
		}.String(),
	}},
	{Name: "_cbrt", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("cbrt", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("cbrt", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Cbrt(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`cbrt` returns the cube root of x",
			signature:   "cbrt(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "cbrt(8.0) => 2.0",
		}.String(),
	}},
	{Name: "_ceil", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("ceil", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("ceil", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Ceil(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`ceil` returns the least integer value greater than or equal to x",
			signature:   "ceil(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "ceil(1.2) => 2.0",
		}.String(),
	}},
	{Name: "_copysign", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("copysign", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("copysign", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("copysign", 2, FLOAT_OBJ, args[1].Type())
			}
			f := args[0].(*Float).Value
			sign := args[1].(*Float).Value
			return &Float{Value: math.Copysign(f, sign)}
		},
		HelpStr: helpStrArgs{
			explanation: "`copysign` returns a value with the magnitude of f and the sign of sign",
			signature:   "copysign(f: float, sign: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "copysign(1.2, -2.8) => -1.2",
		}.String(),
	}},
	{Name: "_cos", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("cos", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("cos", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Cos(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`cos` returns the cosine of the radian argument x",
			signature:   "cos(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "cos(1.20) => 0.362358",
		}.String(),
	}},
	{Name: "_cosh", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("cosh", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("cosh", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Cosh(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`cosh` returns the hyperbolic cosine of x",
			signature:   "cosh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "cosh(1.2) => 1.810656",
		}.String(),
	}},
	{Name: "_dim", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("dim", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("dim", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("dim", 2, FLOAT_OBJ, args[1].Type())
			}
			x := args[0].(*Float).Value
			y := args[1].(*Float).Value
			return &Float{Value: math.Dim(x, y)}
		},
		HelpStr: helpStrArgs{
			explanation: "`dim` returns the maximum of x-y or 0",
			signature:   "dim(x: float, y: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "dim(3.4, 1.2) => 2.2",
		}.String(),
	}},
	{Name: "_erf", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("erf", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("erf", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Erf(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`erf` returns the error function of x",
			signature:   "erf(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "erf(1.234567) => 0.919179",
		}.String(),
	}},
	{Name: "_erfc", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("erfc", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("erfc", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Erfc(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`erfc` returns the complementary error function of x",
			signature:   "erfc(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "erfc(1.234567) => 0.080821",
		}.String(),
	}},
	{Name: "_erfcinv", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("erfcinv", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("erfcinv", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Erfcinv(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`erfcinv` returns the inverse of erfc(x)",
			signature:   "erfcinv(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "erfcinv(1.234567) => -0.210968",
		}.String(),
	}},
	{Name: "_erfinv", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("erfinv", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("erfinv", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Erfinv(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`erfinv` returns the inverse error function of x",
			signature:   "erfinv(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "erfinv(0.234567) => 0.210968",
		}.String(),
	}},
	{Name: "_fma", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 3 {
				return newInvalidArgCountError("fma", len(args), 3, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("fma", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("fma", 2, FLOAT_OBJ, args[1].Type())
			}
			if args[2].Type() != FLOAT_OBJ {
				return newPositionalTypeError("fma", 3, FLOAT_OBJ, args[2].Type())
			}
			x := args[0].(*Float).Value
			y := args[1].(*Float).Value
			z := args[2].(*Float).Value
			return &Float{Value: math.FMA(x, y, z)}
		},
		HelpStr: helpStrArgs{
			explanation: "`fma` returns x * y + z, computed with only one rounding. fma returns the fused multiply-add of x, y, and z",
			signature:   "fma(x: float, y: float, z: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "fma(2.0, 3.0, 4.0) => 10.0",
		}.String(),
	}},
	{Name: "_floor", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("floor", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("floor", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Floor(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`floor` returns the greatest integer value less than or equal to x",
			signature:   "floor(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "floor(1.2) => 1.0",
		}.String(),
	}},
	{Name: "_frexp", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("frexp", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("frexp", 1, FLOAT_OBJ, args[0].Type())
			}
			frac, exp := math.Frexp(args[0].(*Float).Value)
			mapObj := NewOrderedMap[string, Object]()
			mapObj.Set("frac", &Float{Value: frac})
			mapObj.Set("exp", &Integer{Value: int64(exp)})
			return CreateMapObjectForGoMap(*mapObj)
		},
		HelpStr: helpStrArgs{
			explanation: "`frexp` breaks f into a normalized fraction and an integral power of two. it returns frac and exp satisfying f == frac x 2**exp, with the absolute value of frac in the interval [1/2, 1)",
			signature:   "frexp(x: float) -> {frac: float, exp: int}",
			errors:      "InvalidArgCount,PositionalType",
			example:     "frexp(3.0) => {frac: 0.750000, exp: 2}",
		}.String(),
	}},
	{Name: "_gamma", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("gamma", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("gamma", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Gamma(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`gamma` returns the Gamma function of x",
			signature:   "gamma(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "gamma(2.0) => 1.0",
		}.String(),
	}},
	{Name: "_gcd", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("gcd", len(args), 2, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("gcd", 1, INTEGER_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("gcd", 2, INTEGER_OBJ, args[1].Type())
			}
			a, b := args[0].(*Integer).Value, args[1].(*Integer).Value
			return &Integer{Value: gcd(a, b)}
		},
		HelpStr: helpStrArgs{
			explanation: "`gcd` returns the greatest common divisor (GCD) via Euclidean algorithm",
			signature:   "gcd(a: int, b: int) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "gcd(10,20) => 10",
		}.String(),
	}},
	{Name: "_hypot", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("hypot", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("hypot", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("hypot", 2, FLOAT_OBJ, args[1].Type())
			}
			p := args[0].(*Float).Value
			q := args[1].(*Float).Value
			return &Float{Value: math.Hypot(p, q)}
		},
		HelpStr: helpStrArgs{
			explanation: "`hypot` returns sqrt(p*p + q*q), taking care to avoid unnecessary overflow and underflow",
			signature:   "hypot(p: float, q: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "hypot(3.0,4.0) => 5.0",
		}.String(),
	}},
	{Name: "_ilogb", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("ilogb", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("ilogb", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Integer{Value: int64(math.Ilogb(x))}
		},
		HelpStr: helpStrArgs{
			explanation: "`ilogb` returns the binary exponent of x as an INTEGER",
			signature:   "ilogb(x: float) -> int",
			errors:      "InvalidArgCount,PositionalType",
			example:     "ilogb(203.0) => 7",
		}.String(),
	}},
	{Name: "_inf", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("inf", len(args), 1, "")
			}
			if args[0].Type() != INTEGER_OBJ {
				return newPositionalTypeError("inf", 1, INTEGER_OBJ, args[0].Type())
			}
			sign := args[0].(*Integer).Value
			return &Float{Value: math.Inf(int(sign))}
		},
		HelpStr: helpStrArgs{
			explanation: "`inf` returns positive infinity if sign >= 0, negative infinity if sign < 0",
			signature:   "inf(sign: int) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "inf(1) => +Inf",
		}.String(),
	}},
	{Name: "_is_inf", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("is_inf", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("is_inf", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("is_inf", 2, INTEGER_OBJ, args[1].Type())
			}
			f := args[0].(*Float).Value
			sign := int(args[1].(*Integer).Value)
			return nativeToBooleanObject(math.IsInf(f, sign))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_inf` reports whether f is an infinity, according to sign. if sign > 0 { f == +Inf } else if sign < 0 { f == -Inf } else if sign == 0 { f == +Inf || f == -Inf}",
			signature:   "is_inf(x: float, sign: int) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_inf(inf(1), 0) => true",
		}.String(),
	}},
	{Name: "_is_NaN", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("is_NaN", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("is_NaN", 1, FLOAT_OBJ, args[0].Type())
			}
			f := args[0].(*Float).Value
			return nativeToBooleanObject(math.IsNaN(f))
		},
		HelpStr: helpStrArgs{
			explanation: "`is_NaN` reports whether f is not-a-number value",
			signature:   "is_NaN(x: float) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "is_NaN(NaN) => true",
		}.String(),
	}},
	{Name: "_j0", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("j0", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("j0", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.J0(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`j0` returns the order-zero Bessel function of the first kind",
			signature:   "j0(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "j0(1.2) => 0.671133",
		}.String(),
	}},
	{Name: "_j1", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("j1", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("j1", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.J1(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`j1` returns the order-one Bessel function of the first kind",
			signature:   "j1(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "j1(1.2) => 0.498289",
		}.String(),
	}},
	{Name: "_jn", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("jn", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("jn", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("jn", 2, INTEGER_OBJ, args[1].Type())
			}
			n := int(args[1].(*Integer).Value)
			x := args[0].(*Float).Value
			return &Float{Value: math.Jn(n, x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`jn` returns the order-n Bessel function of the first kind",
			signature:   "jn(x: float, n: int) -> float",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "jn(1.2, 3) => 0.032874",
		}.String(),
	}},
	{Name: "_lcm", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) < 1 {
				return newInvalidArgCountError("lcm", len(args), 1, "as a list, or 2 or more")
			}
			if args[0].Type() == LIST_OBJ {
				l := args[0].(*List)
				ints := make([]int64, len(l.Elements))
				if len(l.Elements) < 2 {
					return newError("`lcm` error: list must be at least 2 elements long")
				}
				for i, e := range l.Elements {
					if e.Type() != INTEGER_OBJ {
						return newError("`lcm` error: all elements in list need to be INTEGER. got=%s", e.Type())
					}
					ints[i] = e.(*Integer).Value
				}
				if len(ints) > 2 {
					return &Integer{Value: lcm(ints[0], ints[1], ints[2:]...)}
				}
				return &Integer{Value: lcm(ints[0], ints[1])}
			}
			if len(args) < 2 {
				return newInvalidArgCountError("lcm", len(args), 2, "or more")
			}
			if len(args) == 2 {
				if args[0].Type() != INTEGER_OBJ {
					return newPositionalTypeError("lcm", 1, INTEGER_OBJ, args[0].Type())
				}
				if args[1].Type() != INTEGER_OBJ {
					return newPositionalTypeError("lcm", 2, INTEGER_OBJ, args[1].Type())
				}
				return &Integer{Value: lcm(args[0].(*Integer).Value, args[1].(*Integer).Value)}
			} else {
				ints := make([]int64, len(args))
				for i, e := range args {
					if e.Type() != INTEGER_OBJ {
						return newPositionalTypeError("lcm", i+1, INTEGER_OBJ, e.Type())
					}
					ints[i] = e.(*Integer).Value
				}
				return &Integer{Value: lcm(ints[0], ints[1], ints[2:]...)}
			}
		},
		HelpStr: helpStrArgs{
			explanation: "`lcm` finds the Least Common Multiple (LCM) via GCD",
			signature:   "lcm(a: int, b: int, args: int) -> int || lcm(arg: list[int]) -> int",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "lcm(1,2,3,4) => 12",
		}.String(),
	}},
	{Name: "_ldexp", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("ldexp", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("ldexp", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("ldexp", 2, INTEGER_OBJ, args[1].Type())
			}
			frac := args[0].(*Float).Value
			exp := int(args[1].(*Integer).Value)
			return &Float{Value: math.Ldexp(frac, exp)}
		},
		HelpStr: helpStrArgs{
			explanation: "`ldexp` is the inverse of frexp, returns frac x 2**exp.",
			signature:   "ldexp(frac: float, exp: int) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "ldexp(0.75, 2) => 3.0",
		}.String(),
	}},
	{Name: "_lgamma", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("lgamma", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("lgamma", 1, FLOAT_OBJ, args[0].Type())
			}
			lgamma, sign := math.Lgamma(args[0].(*Float).Value)
			mapObj := NewOrderedMap[string, Object]()
			mapObj.Set("lgamma", &Float{Value: lgamma})
			mapObj.Set("sign", &Integer{Value: int64(sign)})
			return CreateMapObjectForGoMap(*mapObj)
		},
		HelpStr: helpStrArgs{
			explanation: "`lgamma` returns the natural logarithm and sign (-1 or +1) of gamma(x)",
			signature:   "lgamma(x: float) -> {lgamma: float, sign: int}",
			errors:      "InvalidArgCount,PositionalType",
			example:     "lgamma(2.3) => {lgamma: 0.154189, sign: 1}",
		}.String(),
	}},
	{Name: "_log", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("log", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("log", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Log(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`log` returns the natural logarithm of x",
			signature:   "log(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "log(120.0) => 4.787492",
		}.String(),
	}},
	{Name: "_log10", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("log10", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("log10", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Log10(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`log10` returns the decimal logarithm of x",
			signature:   "log10(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "log10(120.0) => 2.079181",
		}.String(),
	}},
	{Name: "_log1p", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("log1p", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("log1p", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Log1p(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`log1p` returns the natural logarithm of 1 plus its argument x. it is more accurate than log(1 + x) when x is near zero",
			signature:   "log1p(x: float) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "log1p(0.2) => 0.182322",
		}.String(),
	}},
	{Name: "_log2", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("log2", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("log2", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Log2(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`log2` returns the binary logarithm of x",
			signature:   "log2(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "log2(0.2) => -2.321928",
		}.String(),
	}},
	{Name: "_logb", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("logb", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("logb", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Logb(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`logb` returns the binary exponent of x",
			signature:   "logb(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "logb(0.2) => -3.0",
		}.String(),
	}},
	{Name: "_modf", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("modf", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("modf", 1, FLOAT_OBJ, args[0].Type())
			}
			i, frac := math.Modf(args[0].(*Float).Value)
			mapObj := NewOrderedMap[string, Object]()
			mapObj.Set("i", &Integer{Value: int64(i)})
			mapObj.Set("frac", &Float{Value: frac})
			return CreateMapObjectForGoMap(*mapObj)
		},
		HelpStr: helpStrArgs{
			explanation: "`modf` returns INTEGER and fractional FLOAT numbers that sum to f. both values have the same sign as f",
			signature:   "modf(x: float) -> {i: int, frac: float}",
			errors:      "InvalidArgCount,PositionalType",
			example:     "modf(10.1) => {i: 10, frac: 0.1}",
		}.String(),
	}},
	{Name: "_next_after", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("next_after", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("next_after", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("next_after", 2, FLOAT_OBJ, args[1].Type())
			}
			x := args[0].(*Float).Value
			y := args[1].(*Float).Value
			return &Float{Value: math.Nextafter(x, y)}
		},
		HelpStr: helpStrArgs{
			explanation: "`next_after` returns the next representable FLOAT value after x towards y",
			signature:   "next_after(x: float, y: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "next_after(3.1, 5.0) => 3.1",
		}.String(),
	}},
	{Name: "_remainder", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("remainder", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("remainder", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != FLOAT_OBJ {
				return newPositionalTypeError("remainder", 2, FLOAT_OBJ, args[1].Type())
			}
			x := args[0].(*Float).Value
			y := args[1].(*Float).Value
			return &Float{Value: math.Remainder(x, y)}
		},
		HelpStr: helpStrArgs{
			explanation: "`remainder` returns the FLOAT remainder of x/y",
			signature:   "remainder(x: float, y: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "remainder(98.2,38.3) => -16.7",
		}.String(),
	}},
	{Name: "_round", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("round", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("round", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Round(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`round` returns the nearest integer as a float, rounding half away from zero",
			signature:   "round(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "round(3.5) => 4.0",
		}.String(),
	}},
	{Name: "_round_to_even", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("round_to_even", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("round_to_even", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.RoundToEven(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`round_to_even` returns the nearest integer as a float, rounding ties to even",
			signature:   "round_to_even(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "round_to_even(3.2) => 3.0",
		}.String(),
	}},
	{Name: "_signbit", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("signbit", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("signbit", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return nativeToBooleanObject(math.Signbit(x))
		},
		HelpStr: helpStrArgs{
			explanation: "`signbit` reports whether x is negative or negative zero",
			signature:   "signbit(x: float) -> bool",
			errors:      "InvalidArgCount,PositionalType",
			example:     "signbit(-3.0) => true",
		}.String(),
	}},
	{Name: "_sin", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("sin", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("sin", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Sin(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`sin` returns the sine of the radian argument x",
			signature:   "sin(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "sin(0.5) => 0.479426",
		}.String(),
	}},
	{Name: "_sincos", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("sincos", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("sincos", 1, FLOAT_OBJ, args[0].Type())
			}
			sin, cos := math.Sincos(args[0].(*Float).Value)
			mapObj := NewOrderedMap[string, Object]()
			mapObj.Set("sin", &Float{Value: sin})
			mapObj.Set("cos", &Float{Value: cos})
			return CreateMapObjectForGoMap(*mapObj)
		},
		HelpStr: helpStrArgs{
			explanation: "`sincos` returns sin(x), cos(x)",
			signature:   "sincos(x: float) -> {sin: float, cos: float}",
			errors:      "InvalidArgCount,PositionalType",
			example:     "sincos(0.5) => {sin: 0.479426, cos: 0.877583}",
		}.String(),
	}},
	{Name: "_sinh", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("sinh", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("sinh", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Sinh(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`sinh` returns the hyperbolic sine of x",
			signature:   "sinh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "sinh(0.5) => 0.521095",
		}.String(),
	}},
	{Name: "_tan", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("tan", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("tan", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Tan(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`tan` returns the tangent of the radian argument x",
			signature:   "tan(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "tan(0.5) => 0.546302",
		}.String(),
	}},
	{Name: "_tanh", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("tanh", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("tanh", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Tanh(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`tanh` returns the hyperbolic tangent of x",
			signature:   "tanh(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "tanh(0.5) => 0.462117",
		}.String(),
	}},
	{Name: "_trunc", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("trunc", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("trunc", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Trunc(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`trunc` returns the integer value of x as a FLOAT",
			signature:   "trunc(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "trunc(2.5) => 2.0",
		}.String(),
	}},
	{Name: "_y0", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("y0", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("y0", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Y0(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`y0` returns the order-zero Bessel function of the second kind",
			signature:   "y0(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "y0(2.0) => 0.510376",
		}.String(),
	}},
	{Name: "_y1", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 1 {
				return newInvalidArgCountError("y1", len(args), 1, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("y1", 1, FLOAT_OBJ, args[0].Type())
			}
			x := args[0].(*Float).Value
			return &Float{Value: math.Y1(x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`y1` returns the order-one Bessel function of the second kind",
			signature:   "y1(x: float) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "y1(2.0) => -0.107032",
		}.String(),
	}},
	{Name: "_yn", Builtin: &Builtin{
		Fun: func(args ...Object) Object {
			if len(args) != 2 {
				return newInvalidArgCountError("yn", len(args), 2, "")
			}
			if args[0].Type() != FLOAT_OBJ {
				return newPositionalTypeError("yn", 1, FLOAT_OBJ, args[0].Type())
			}
			if args[1].Type() != INTEGER_OBJ {
				return newPositionalTypeError("yn", 2, INTEGER_OBJ, args[1].Type())
			}
			n := int(args[1].(*Integer).Value)
			x := args[0].(*Float).Value
			return &Float{Value: math.Yn(n, x)}
		},
		HelpStr: helpStrArgs{
			explanation: "`yn` returns the order-n Bessel function of the second kind",
			signature:   "yn(x: float, n: int) -> float",
			errors:      "InvalidArgCount,PositionalType",
			example:     "yn(3.0, 5) => -1.905946",
		}.String(),
	}},
}
