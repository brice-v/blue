package vm

import (
	"blue/object"
	"math"

	"gonum.org/v1/gonum/integrate/quad"
	"gonum.org/v1/gonum/optimize"
)

// numVmBuiltin resolves the num builtins that call back into blue. They live in
// the vm package because only the vm can invoke a blue closure, following the
// same pattern as the ui module's handlers.
func numVmBuiltin(name string, vm *VM) *object.Builtin {
	switch name {
	case "_num_quad":
		return createNumQuadBuiltin(vm)
	case "_num_minimize":
		return createNumMinimizeBuiltin(vm)
	default:
		return nil
	}
}

// blueFnCaller adapts a blue callable to a Go func(float64...) float64.
type blueFnCaller struct {
	vm   *VM
	fn   object.Object
	name string
}

// call invokes the blue function with the given float64 arguments. A call error
// or a non-numeric result becomes NaN, so a bad closure degrades the numeric
// result rather than corrupting the VM.
func (c *blueFnCaller) call(xs ...float64) float64 {
	args := make([]object.Object, len(xs))
	for i, v := range xs {
		args[i] = &object.Float{Value: v}
	}
	res := c.vm.applyFunctionFastWithMultipleArgs(c.fn, args)
	if res == nil || res.Type() == object.ERROR_OBJ {
		return math.NaN()
	}
	switch v := res.(type) {
	case *object.Float:
		return v.Value
	case *object.Integer:
		return float64(v.Value)
	default:
		return math.NaN()
	}
}

// blueFnArg validates a blue callable argument.
func blueFnArg(name string, vm *VM, pos int, o object.Object) (*blueFnCaller, object.Object) {
	switch o.Type() {
	case object.FUNCTION_OBJ, object.CLOSURE, object.BUILTIN_OBJ, object.COMPILED_FUNCTION_OBJ:
	default:
		return nil, newPositionalTypeError(name, pos, "FUNCTION", o.Type())
	}
	return &blueFnCaller{vm: vm, fn: o, name: name}, nil
}

// createNumQuadBuiltin builds `num.quad(f, min, max, n)`: Gauss-Legendre
// quadrature of a blue function over [min, max].
func createNumQuadBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "_num_quad",
		Fun: func(args ...object.Object) object.Object {
			if len(args) != 4 {
				return newInvalidArgCountError("quad", len(args), 4, "")
			}
			caller, errObj := blueFnArg("quad", vm, 1, args[0])
			if errObj != nil {
				return errObj
			}
			minV, ok := objectFloat64(args[1])
			if !ok {
				return newPositionalTypeError("quad", 2, object.FLOAT_OBJ, args[1].Type())
			}
			maxV, ok := objectFloat64(args[2])
			if !ok {
				return newPositionalTypeError("quad", 3, object.FLOAT_OBJ, args[2].Type())
			}
			n, ok := args[3].(*object.Integer)
			if !ok {
				return newPositionalTypeError("quad", 4, object.INTEGER_OBJ, args[3].Type())
			}
			if n.Value <= 0 {
				return newError("`quad` error: n must be positive, got %d", n.Value)
			}
			if maxV < minV {
				return newError("`quad` error: max must be at least min, got %g and %g", minV, maxV)
			}
			v := quad.Fixed(func(x float64) float64 { return caller.call(x) }, minV, maxV, int(n.Value), quad.Legendre{}, 1)
			return &object.Float{Value: v}
		},
		HelpStr: "quad(f: fun, min: float, max: float, n: int=100) -> float",
	}
}

// createNumMinimizeBuiltin builds `num.minimize(f, x0)`.
func createNumMinimizeBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "_num_minimize",
		Fun: func(args ...object.Object) object.Object {
			if len(args) < 2 || len(args) > 4 {
				return newInvalidArgCountError("minimize", len(args), 2, "to 4")
			}
			caller, errObj := blueFnArg("minimize", vm, 1, args[0])
			if errObj != nil {
				return errObj
			}
			x0, ok := objectFloat64Slice(args[1])
			if !ok {
				return newPositionalTypeError("minimize", 2, "LIST of numbers", args[1].Type())
			}
			if len(x0) == 0 {
				return newError("`minimize` error: no starting point")
			}
			methodName := ""
			if len(args) >= 3 {
				s, ok := args[2].(*object.Stringo)
				if !ok {
					return newPositionalTypeError("minimize", 3, object.STRING_OBJ, args[2].Type())
				}
				methodName = s.Value
			}
			var maxIter int
			if len(args) == 4 {
				n, ok := args[3].(*object.Integer)
				if !ok {
					return newPositionalTypeError("minimize", 4, object.INTEGER_OBJ, args[3].Type())
				}
				maxIter = int(n.Value)
			}

			var meth optimize.Method
			needsGrad := false
			switch methodName {
			case "", "nelder-mead", "neldermead":
				meth = &optimize.NelderMead{}
			case "bfgs":
				meth = &optimize.BFGS{}
				needsGrad = true
			case "cg", "conjugate-gradient":
				meth = &optimize.CG{}
				needsGrad = true
			default:
				return newError("`minimize` error: unknown method %q, want 'nelder-mead', 'bfgs', or 'cg'", methodName)
			}

			p := optimize.Problem{
				Func: func(x []float64) float64 { return caller.call(x...) },
			}
			if needsGrad {
				p.Grad = func(grad, x []float64) { numericGradient(caller, grad, x) }
			}
			settings := &optimize.Settings{}
			if maxIter > 0 {
				settings.MajorIterations = maxIter
			}
			res, err := optimize.Minimize(p, x0, settings, meth)
			if err != nil {
				return newError("`minimize` error: %s", err.Error())
			}
			xs := make([]object.Object, len(res.X))
			for i, v := range res.X {
				xs[i] = &object.Float{Value: v}
			}
			out := object.NewOrderedMap[string, object.Object]()
			out.Set("x", &object.List{Elements: xs})
			out.Set("value", &object.Float{Value: res.F})
			out.Set("iterations", object.NewInteger(int64(res.MajorIterations)))
			return object.CreateMapObjectForGoMap(*out)
		},
		HelpStr: "minimize(f: fun, x0: list, method: str='nelder-mead', max_iter: int=0) -> map",
	}
}

// numericGradient fills grad with a central difference of the objective, for the
// methods that want a gradient but only received a plain blue function.
func numericGradient(c *blueFnCaller, grad, x []float64) {
	const h = 1e-6
	pts := make([]float64, len(x))
	copy(pts, x)
	for i := range x {
		orig := pts[i]
		pts[i] = orig + h
		fp := c.call(pts...)
		pts[i] = orig - h
		fm := c.call(pts...)
		pts[i] = orig
		grad[i] = (fp - fm) / (2 * h)
	}
}

// objectFloat64 reads a blue number as float64.
func objectFloat64(o object.Object) (float64, bool) {
	switch v := o.(type) {
	case *object.Float:
		return v.Value, true
	case *object.Integer:
		return float64(v.Value), true
	case *object.UInteger:
		return float64(v.Value), true
	case *object.Boolean:
		if v.Value {
			return 1, true
		}
		return 0, true
	default:
		return 0, false
	}
}

// objectFloat64Slice reads a blue list of numbers as []float64. A num list, an
// ml tensor, or a nested matrix is not accepted here: optimize needs a flat
// starting point.
func objectFloat64Slice(o object.Object) ([]float64, bool) {
	l, ok := o.(*object.List)
	if !ok {
		return nil, false
	}
	out := make([]float64, len(l.Elements))
	for i, e := range l.Elements {
		v, ok := objectFloat64(e)
		if !ok {
			return nil, false
		}
		out[i] = v
	}
	return out, true
}
