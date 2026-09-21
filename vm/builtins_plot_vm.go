//go:build !static

package vm

import (
	"blue/object"
	"math"
)

// plotVmBuiltin resolves the plot builtin that calls back into blue. It is
// non-static only, because a static build omits the plot module entirely.
func plotVmBuiltin(name string, vm *VM) *object.Builtin {
	if name == "_plot_function" {
		return createPlotFunctionBuiltin(vm)
	}
	return nil
}

// createPlotFunctionBuiltin builds `plot.function(f, fn, min, max, samples)`.
// It samples the blue function at evenly spaced points and adds a line series, so
// a curve can be plotted without materialising the samples first.
func createPlotFunctionBuiltin(vm *VM) *object.Builtin {
	return &object.Builtin{
		Name: "_plot_function",
		Fun: func(args ...object.Object) object.Object {
			if len(args) < 4 || len(args) > 6 {
				return newInvalidArgCountError("function", len(args), 4, "to 6")
			}
			figure, ok := args[0].(*object.GoObj[*object.NumFigure])
			if !ok || figure.Value == nil {
				return newPositionalTypeError("function", 1, "FIGURE", args[0].Type())
			}
			caller, errObj := blueFnArg("function", vm, 2, args[1])
			if errObj != nil {
				return errObj
			}
			minV, ok := objectFloat64(args[2])
			if !ok {
				return newPositionalTypeError("function", 3, object.FLOAT_OBJ, args[2].Type())
			}
			maxV, ok := objectFloat64(args[3])
			if !ok {
				return newPositionalTypeError("function", 4, object.FLOAT_OBJ, args[3].Type())
			}
			if maxV <= minV {
				return newError("`function` error: max must be greater than min, got %g and %g", minV, maxV)
			}
			samples := 200
			if len(args) >= 5 {
				if _, isNull := args[4].(*object.Null); !isNull {
					n, ok := args[4].(*object.Integer)
					if !ok {
						return newPositionalTypeError("function", 5, "INTEGER or NULL", args[4].Type())
					}
					samples = int(n.Value)
				}
			}
			if samples < 2 {
				return newError("`function` error: samples must be at least 2, got %d", samples)
			}
			name := ""
			if len(args) == 6 {
				if _, isNull := args[5].(*object.Null); !isNull {
					s, ok := args[5].(*object.Stringo)
					if !ok {
						return newPositionalTypeError("function", 6, "STRING or NULL", args[5].Type())
					}
					name = s.Value
				}
			}
			// Sample the blue function here, where the vm can call it, then hand
			// the points to the plot layer to add as a line series.
			//
			// gonum's line plotter rejects NaN, so a point where the closure
			// failed (or produced a non-number) truncates the curve at the last
			// finite sample instead of failing the whole plot.
			xs := make([]float64, 0, samples)
			ys := make([]float64, 0, samples)
			step := (maxV - minV) / float64(samples-1)
			for i := 0; i < samples; i++ {
				x := minV + step*float64(i)
				y := caller.call(x)
				if math.IsNaN(y) || math.IsInf(y, 0) {
					break
				}
				xs = append(xs, x)
				ys = append(ys, y)
			}
			if len(xs) < 2 {
				// Fewer than two finite points cannot make a line, so a constant
				// value is not plottable. Report it rather than silently drawing
				// nothing.
				return newError("`function` error: the function produced fewer than 2 finite samples")
			}
			if obj := object.PlotAddSampledFunction(figure.Value, xs, ys, name); obj != nil {
				return obj
			}
			return object.NULL
		},
		HelpStr: "function(f: figure, fn: fun, min: float, max: float, samples: int=200, name: str|null=null) -> null",
	}
}
