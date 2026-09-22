//go:build !static

package object

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette/brewer"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

// NumFigure is the blue-facing handle for a plot. gonum's *plot.Plot already
// owns the series and the legend, so the wrapper only needs to hold it.
type NumFigure struct {
	Plot *plot.Plot
}

// numFigureArg accepts a figure handle.
func numFigureArg(name string, pos int, o Object) (*NumFigure, Object) {
	g, ok := o.(*GoObj[*NumFigure])
	if !ok || g.Value == nil {
		return nil, newPositionalTypeError(name, pos, "FIGURE", o.Type())
	}
	return g.Value, nil
}

// numPointSeriesArg reads a series argument as parallel x and y slices. It
// accepts:
//
//   - a list of [x, y] pairs
//   - an [n, 2] ml tensor or num matrix of points
//   - a flat list or 1d tensor, which becomes y against its index
//
// This is what makes `plot.line(f, tensor)` and `plot.line(f, xs, ys)` both
// work, so a training loop can plot a loss tensor with no ceremony.
func numPointSeriesArg(name string, pos int, o Object) (xs, ys []float64, errObj Object) {
	switch v := o.(type) {
	case *List:
		return listToPoints(name, pos, v)
	case *Tensor, *GoObj[*NumMatrix]:
		// A rank 2 tensor of width 2 is a point list; anything else is y values.
		var shape []int
		switch t := v.(type) {
		case *Tensor:
			shape = t.T.Shape()
		default:
			r, c := v.(*GoObj[*NumMatrix]).Value.Dims()
			shape = []int{r, c}
		}
		if len(shape) == 2 && shape[1] == 2 {
			flat, errObj := numSeriesArg(name, pos, o)
			if errObj != nil {
				return nil, nil, errObj
			}
			xs = make([]float64, shape[0])
			ys = make([]float64, shape[0])
			for i := 0; i < shape[0]; i++ {
				xs[i] = flat[i*2]
				ys[i] = flat[i*2+1]
			}
			return xs, ys, nil
		}
		ys, errObj = numSeriesArg(name, pos, o)
		if errObj != nil {
			return nil, nil, errObj
		}
		xs = make([]float64, len(ys))
		for i := range xs {
			xs[i] = float64(i)
		}
		return xs, ys, nil
	default:
		return nil, nil, newPositionalTypeError(name, pos, "LIST, TENSOR, or MATRIX", o.Type())
	}
}

// listToPoints reads a list as either [x, y] pairs or a flat list of y values.
func listToPoints(name string, pos int, l *List) (xs, ys []float64, errObj Object) {
	if len(l.Elements) == 0 {
		return nil, nil, newError("`%s` error: empty series", name)
	}
	// A list whose first element is itself a 2 element list is a point list.
	if inner, ok := l.Elements[0].(*List); ok {
		if len(inner.Elements) != 2 {
			return nil, nil, newError("`%s` error: a point needs exactly 2 values, found %d", name, len(inner.Elements))
		}
		xs = make([]float64, len(l.Elements))
		ys = make([]float64, len(l.Elements))
		for i, e := range l.Elements {
			pair, ok := e.(*List)
			if !ok || len(pair.Elements) != 2 {
				return nil, nil, newError("`%s` error: every point needs exactly 2 values", name)
			}
			x, ok := objectToFloat64(pair.Elements[0])
			if !ok {
				return nil, nil, newPositionalTypeError(name, pos, "LIST of [x, y] pairs", pair.Elements[0].Type())
			}
			y, ok := objectToFloat64(pair.Elements[1])
			if !ok {
				return nil, nil, newPositionalTypeError(name, pos, "LIST of [x, y] pairs", pair.Elements[1].Type())
			}
			xs[i] = x
			ys[i] = y
		}
		return xs, ys, nil
	}
	// Otherwise it is a flat list of y values indexed by position.
	ys = make([]float64, len(l.Elements))
	for i, e := range l.Elements {
		v, ok := objectToFloat64(e)
		if !ok {
			return nil, nil, newPositionalTypeError(name, pos, "LIST of numbers", e.Type())
		}
		ys[i] = v
	}
	xs = make([]float64, len(ys))
	for i := range xs {
		xs[i] = float64(i)
	}
	return xs, ys, nil
}

// xysFrom converts parallel slices into the plotter.XYs gonum expects.
func xysFrom(xs, ys []float64) (plotter.XYs, error) {
	if len(xs) != len(ys) {
		return nil, fmt.Errorf("x has %d values, y has %d", len(xs), len(ys))
	}
	out := make(plotter.XYs, len(xs))
	for i := range xs {
		out[i].X = xs[i]
		out[i].Y = ys[i]
	}
	return out, nil
}

// numColorArg parses an optional color. Colors come through as a hex string like
// "#3366ff", which keeps the argument surface small.
func numColorArg(o Object) (hexColor, bool) {
	s, ok := o.(*Stringo)
	if !ok {
		return "", false
	}
	return hexColor(s.Value), true
}

// hexColor is a color given as "#rrggbb".
type hexColor string

// vgColor converts a hex string into the color type gonum wants.
func vgColor(h hexColor) (col colorRGB, ok bool) {
	s := string(h)
	if len(s) != 7 || s[0] != '#' {
		return colorRGB{}, false
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return colorRGB{}, false
	}
	return colorRGB{R: r, G: g, B: b, A: 255}, true
}

// colorRGB mirrors color.RGBA without importing image/color at each call site.
type colorRGB struct{ R, G, B, A uint8 }

// rgb satisfies the color.Color interface.
func (c colorRGB) RGBA() (r, g, b, a uint32) {
	return uint32(c.R) * 257, uint32(c.G) * 257, uint32(c.B) * 257, uint32(c.A) * 257
}

// figureDimensions parses width and height in points, defaulting to a sensible
// size so callers can omit them.
func figureDimensions(w, h Object) (vg.Length, vg.Length, Object) {
	width, height := 6*vg.Inch, 4*vg.Inch
	if w != nil {
		v, ok := objectToFloat64(w)
		if !ok {
			return 0, 0, newPositionalTypeError("render", 2, "number", w.Type())
		}
		width = vg.Length(v) * vg.Inch
	}
	if h != nil {
		v, ok := objectToFloat64(h)
		if !ok {
			return 0, 0, newPositionalTypeError("render", 3, "number", h.Type())
		}
		height = vg.Length(v) * vg.Inch
	}
	return width, height, nil
}

// PlotBuiltins is the plot module surface: figures, series, and rendering.
//
// Every series function accepts a list, an ml tensor, or a num matrix, so a
// training loop can plot a loss tensor directly. Rendering returns bytes, which
// blue writes with the existing `write` builtin or displays via ui/wasm.
var PlotBuiltins = []*Builtin{
	{
		Name: "_plot_new",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("plot_new", 0, args); err != nil {
				return err
			}
			return &GoObj[*NumFigure]{Value: &NumFigure{Plot: plot.New()}}
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_new` returns a new empty figure",
			signature:   "plot_new() -> figure",
			errors:      "InvalidArgCount",
			example:     "plot_new() => figure",
		}.String(),
	},
	{
		Name: "_plot_title",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("plot_title", 2, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_title", 1, args[0])
			if errObj != nil {
				return errObj
			}
			s, ok := args[1].(*Stringo)
			if !ok {
				return newPositionalTypeError("plot_title", 2, STRING_OBJ, args[1].Type())
			}
			f.Plot.Title.Text = s.Value
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_title` sets the figure title",
			signature:   "plot_title(f: figure, text: str) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "plot_title(f, 'loss') => null",
		}.String(),
	},
	{
		Name: "_plot_x_label",
		Fun:  plotAxisLabel("plot_x_label", func(f *NumFigure, s string) { f.Plot.X.Label.Text = s }),
		HelpStr: helpStrArgs{
			explanation: "`plot_x_label` sets the x axis label",
			signature:   "plot_x_label(f: figure, text: str) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "plot_x_label(f, 'step') => null",
		}.String(),
	},
	{
		Name: "_plot_y_label",
		Fun:  plotAxisLabel("plot_y_label", func(f *NumFigure, s string) { f.Plot.Y.Label.Text = s }),
		HelpStr: helpStrArgs{
			explanation: "`plot_y_label` sets the y axis label",
			signature:   "plot_y_label(f: figure, text: str) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "plot_y_label(f, 'loss') => null",
		}.String(),
	},
	{
		Name: "_plot_x_log",
		Fun:  plotFlag("plot_x_log", func(f *NumFigure) { f.Plot.X.Scale = plot.LogScale{} }),
		HelpStr: helpStrArgs{
			explanation: "`plot_x_log` switches the x axis to a log scale",
			signature:   "plot_x_log(f: figure) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "plot_x_log(f) => null",
		}.String(),
	},
	{
		Name: "_plot_y_log",
		Fun:  plotFlag("plot_y_log", func(f *NumFigure) { f.Plot.Y.Scale = plot.LogScale{} }),
		HelpStr: helpStrArgs{
			explanation: "`plot_y_log` switches the y axis to a log scale",
			signature:   "plot_y_log(f: figure) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "plot_y_log(f) => null",
		}.String(),
	},
	{
		Name: "_plot_x_range",
		Fun:  plotRange("plot_x_range", func(f *NumFigure, lo, hi float64) { f.Plot.X.Min, f.Plot.X.Max = lo, hi }),
		HelpStr: helpStrArgs{
			explanation: "`plot_x_range` pins the x axis range",
			signature:   "plot_x_range(f: figure, min: float, max: float) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "plot_x_range(f, 0.0, 100.0) => null",
		}.String(),
	},
	{
		Name: "_plot_y_range",
		Fun:  plotRange("plot_y_range", func(f *NumFigure, lo, hi float64) { f.Plot.Y.Min, f.Plot.Y.Max = lo, hi }),
		HelpStr: helpStrArgs{
			explanation: "`plot_y_range` pins the y axis range",
			signature:   "plot_y_range(f: figure, min: float, max: float) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "plot_y_range(f, 0.0, 1.0) => null",
		}.String(),
	},
	{
		Name: "_plot_nominal_x",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("plot_nominal_x", 2, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_nominal_x", 1, args[0])
			if errObj != nil {
				return errObj
			}
			l, ok := args[1].(*List)
			if !ok {
				return newPositionalTypeError("plot_nominal_x", 2, LIST_OBJ, args[1].Type())
			}
			names := make([]string, len(l.Elements))
			for i, e := range l.Elements {
				s, ok := e.(*Stringo)
				if !ok {
					return newPositionalTypeError("plot_nominal_x", 2, "LIST of strings", e.Type())
				}
				names[i] = s.Value
			}
			f.Plot.NominalX(names...)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_nominal_x` labels the x axis with names instead of numbers",
			signature:   "plot_nominal_x(f: figure, names: list[str]) -> null",
			errors:      "InvalidArgCount,PositionalType",
			example:     "plot_nominal_x(f, ['a', 'b']) => null",
		}.String(),
	},
	// --- series -------------------------------------------------------------
	{
		Name: "_plot_line",
		Fun:  plotXYSeries("plot_line", "line"),
		HelpStr: helpStrArgs{
			explanation: "`plot_line` adds a line series; it accepts a tensor, a list of y values, or a list of [x, y] pairs",
			signature:   "plot_line(f: figure, xs: list|tensor, ys: list|tensor, name: str|null=null, color: str|null=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_line(f, loss_tensor, name='train') => null",
		}.String(),
	},
	{
		Name: "_plot_scatter",
		Fun:  plotXYSeries("plot_scatter", "scatter"),
		HelpStr: helpStrArgs{
			explanation: "`plot_scatter` adds a scatter series",
			signature:   "plot_scatter(f: figure, xs: list|tensor, ys: list|tensor, name: str|null=null, color: str|null=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_scatter(f, xs, ys) => null",
		}.String(),
	},
	{
		Name: "_plot_bars",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("plot_bars", []int{2, 3, 4}, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_bars", 1, args[0])
			if errObj != nil {
				return errObj
			}
			values, errObj := numSeriesArg("plot_bars", 2, args[1])
			if errObj != nil {
				return errObj
			}
			labels := make([]string, len(values))
			for i := range labels {
				labels[i] = fmt.Sprintf("%d", i)
			}
			if len(args) >= 3 {
				if l, ok := args[2].(*List); ok {
					for i, e := range l.Elements {
						if i >= len(labels) {
							break
						}
						if s, ok := e.(*Stringo); ok {
							labels[i] = s.Value
						}
					}
				} else if _, isNull := args[2].(*Null); !isNull {
					return newPositionalTypeError("plot_bars", 3, "LIST or NULL", args[2].Type())
				}
			}
			vs := make(plotter.Values, len(values))
			copy(vs, values)
			bars, err := plotter.NewBarChart(vs, vg.Points(20))
			if err != nil {
				return newError("`plot_bars` error: %s", err.Error())
			}
			if len(args) == 4 {
				if col, ok := numColorArg(args[3]); ok {
					if c, ok := vgColor(col); ok {
						bars.Color = color.RGBA(c)
					}
				}
			}
			f.Plot.Add(bars)
			f.Plot.NominalX(labels...)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_bars` adds a bar chart",
			signature:   "plot_bars(f: figure, values: list|tensor, labels: list[str]|null=null, color: str|null=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_bars(f, [1.0, 2.0], ['a', 'b']) => null",
		}.String(),
	},
	{
		Name: "_plot_hist",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("plot_hist", []int{2, 3, 4}, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_hist", 1, args[0])
			if errObj != nil {
				return errObj
			}
			values, errObj := numSeriesArg("plot_hist", 2, args[1])
			if errObj != nil {
				return errObj
			}
			bins := 10
			if len(args) >= 3 {
				if _, isNull := args[2].(*Null); !isNull {
					n, ok := args[2].(*Integer)
					if !ok {
						return newPositionalTypeError("plot_hist", 3, "INTEGER or NULL", args[2].Type())
					}
					bins = int(n.Value)
				}
			}
			if len(values) == 0 {
				return newError("`plot_hist` error: empty input")
			}
			vs := make(plotter.Values, len(values))
			copy(vs, values)
			hist, err := plotter.NewHist(vs, bins)
			if err != nil {
				return newError("`plot_hist` error: %s", err.Error())
			}
			if len(args) == 4 {
				if col, ok := numColorArg(args[3]); ok {
					if c, ok := vgColor(col); ok {
						hist.FillColor = color.RGBA(c)
					}
				}
			}
			f.Plot.Add(hist)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_hist` adds a histogram",
			signature:   "plot_hist(f: figure, values: list|tensor, bins: int|null=null, color: str|null=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_hist(f, error_tensor, 20) => null",
		}.String(),
	},
	{
		Name: "_plot_heatmap",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("plot_heatmap", 2, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_heatmap", 1, args[0])
			if errObj != nil {
				return errObj
			}
			m, errObj := numMatrixArg("plot_heatmap", 2, args[1])
			if errObj != nil {
				return errObj
			}
			r, c := m.Dims()
			if r == 0 || c == 0 {
				return newError("`plot_heatmap` error: empty matrix")
			}
			grid := &numGrid{m: m}
			hm := plotter.NewHeatMap(grid, heapPalette)
			f.Plot.Add(hm)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_heatmap` adds a heatmap of a matrix, which is a direct way to visualise model weights; an ml tensor works as the input",
			signature:   "plot_heatmap(f: figure, m: matrix|tensor|list) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_heatmap(f, weight_tensor) => null",
		}.String(),
	},
	// --- output -------------------------------------------------------------
	{
		Name: "_plot_render",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("plot_render", []int{1, 2, 3, 4}, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_render", 1, args[0])
			if errObj != nil {
				return errObj
			}
			format := "png"
			if len(args) >= 2 {
				s, ok := args[1].(*Stringo)
				if !ok {
					return newPositionalTypeError("plot_render", 2, STRING_OBJ, args[1].Type())
				}
				format = s.Value
			}
			var wArg, hArg Object
			if len(args) >= 3 {
				wArg = args[2]
			}
			if len(args) == 4 {
				hArg = args[3]
			}
			data, errObj := renderFigure(f, format, wArg, hArg)
			if errObj != nil {
				return errObj
			}
			return &Bytes{Value: data}
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_render` renders a figure to bytes; formats are png, svg, pdf, eps, jpg, and tex",
			signature:   "plot_render(f: figure, format: str='png', width: float=6.0, height: float=4.0) -> bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_render(f) => bytes",
		}.String(),
	},
	{
		Name: "_plot_save",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("plot_save", []int{2, 3, 4}, args); err != nil {
				return err
			}
			path, ok := args[1].(*Stringo)
			if !ok {
				return newPositionalTypeError("plot_save", 2, STRING_OBJ, args[1].Type())
			}
			f, errObj := numFigureArg("plot_save", 1, args[0])
			if errObj != nil {
				return errObj
			}
			// The format follows the file extension, so plot_save(f, "x.svg")
			// writes SVG. An explicit format argument overrides it.
			format := formatFromPath(path.Value)
			var wArg, hArg Object
			if len(args) >= 3 {
				if s, ok := args[2].(*Stringo); ok {
					format = s.Value
				} else {
					wArg = args[2]
				}
			}
			if len(args) == 4 {
				hArg = args[3]
			}
			data, errObj := renderFigure(f, format, wArg, hArg)
			if errObj != nil {
				return errObj
			}
			return writeBytesToPath(path.Value, data)
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_save` renders a figure and writes it to a path; the format comes from the file extension",
			signature:   "plot_save(f: figure, path: str, width: float=6.0, height: float=4.0) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_save(f, 'loss.png') => null",
		}.String(),
	},
	{
		Name: "_plot_palette",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("plot_palette", 1, args); err != nil {
				return err
			}
			s, ok := args[0].(*Stringo)
			if !ok {
				return newPositionalTypeError("plot_palette", 1, STRING_OBJ, args[0].Type())
			}
			pal, ok := brewerPalette(s.Value)
			if !ok {
				return newError("`plot_palette` error: unknown palette %q", s.Value)
			}
			elems := make([]Object, 0, len(pal))
			for _, c := range pal {
				r, g, b, _ := c.RGBA()
				elems = append(elems, &Stringo{Value: fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)})
			}
			return &List{Elements: elems}
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_palette` returns a named color palette as a list of hex strings",
			signature:   "plot_palette(name: str) -> list[str]",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_palette('dark') => ['#1b9e77', ...]",
		}.String(),
	},
	{
		Name: "_plot_error_bars",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("plot_error_bars", []int{4, 5}, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_error_bars", 1, args[0])
			if errObj != nil {
				return errObj
			}
			xs, errObj := numSeriesArg("plot_error_bars", 2, args[1])
			if errObj != nil {
				return errObj
			}
			ys, errObj := numSeriesArg("plot_error_bars", 3, args[2])
			if errObj != nil {
				return errObj
			}
			errs, errObj := numSeriesArg("plot_error_bars", 4, args[3])
			if errObj != nil {
				return errObj
			}
			pts, err := xysFrom(xs, ys)
			if err != nil {
				return newError("`plot_error_bars` error: %s", err.Error())
			}
			if len(errs) != len(pts) {
				return newError("`plot_error_bars` error: %d errors for %d points", len(errs), len(pts))
			}
			// gonum's YErrorBars draws only the caps, so a scatter is drawn under
			// them for the points themselves. Points are on by default; pass
			// false to draw bare bars.
			drawPoints := true
			if len(args) == 5 {
				if b, ok := args[4].(*Boolean); ok {
					drawPoints = b.Value
				}
			}
			if drawPoints {
				sc, err := plotter.NewScatter(pts)
				if err != nil {
					return newError("`plot_error_bars` error: %s", err.Error())
				}
				f.Plot.Add(sc)
			}
			bars := &plotter.YErrorBars{XYs: pts, CapWidth: vg.Points(5)}
			bars.YErrors = make(plotter.YErrors, len(pts))
			for i := range pts {
				bars.YErrors[i].Low = -errs[i]
				bars.YErrors[i].High = errs[i]
			}
			f.Plot.Add(bars)
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_error_bars` adds symmetric vertical error bars at each point, with a scatter on top by default",
			signature:   "plot_error_bars(f: figure, x: list|tensor, y: list|tensor, yerr: list|tensor, points: bool=true) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_error_bars(f, [1.0, 2.0], [3.0, 4.0], [0.1, 0.2]) => null",
		}.String(),
	},
	{
		Name: "_plot_box",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("plot_box", []int{2, 3, 4}, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_box", 1, args[0])
			if errObj != nil {
				return errObj
			}
			// A list of lists draws one box per group; anything flat draws one box.
			groups, errObj := numBoxGroups("plot_box", args[1])
			if errObj != nil {
				return errObj
			}
			width := vg.Points(20)
			if len(args) >= 3 {
				w, errObj := numFloatArg("plot_box", 3, args[2])
				if errObj != nil {
					return errObj
				}
				width = vg.Points(w)
			}
			var fill color.Color
			if len(args) == 4 {
				if col, ok := numColorArg(args[3]); ok {
					if c, ok := vgColor(col); ok {
						fill = color.RGBA(c)
					}
				}
			}
			for i, g := range groups {
				vs := make(plotter.Values, len(g))
				copy(vs, g)
				box, err := plotter.NewBoxPlot(width, float64(i), vs)
				if err != nil {
					return newError("`plot_box` error: %s", err.Error())
				}
				if fill != nil {
					box.FillColor = fill
				}
				f.Plot.Add(box)
			}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_box` adds a box plot per group; a list of lists gives several groups, a flat list one, and a tensor works too",
			signature:   "plot_box(f: figure, groups: list|list[list]|tensor, width: float=20.0, color: str|null=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_box(f, [[1.0, 2.0, 3.0], [2.0, 3.0, 4.0]]) => null",
		}.String(),
	},
	{
		Name: "_plot_contour",
		Fun: func(args ...Object) Object {
			if err := checkArgCount("plot_contour", 2, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_contour", 1, args[0])
			if errObj != nil {
				return errObj
			}
			m, errObj := numMatrixArg("plot_contour", 2, args[1])
			if errObj != nil {
				return errObj
			}
			levels := contourLevels(m)
			if levels == nil {
				return newError("`plot_contour` error: matrix has no range to contour")
			}
			f.Plot.Add(plotter.NewContour(&numGrid{m: m}, levels, heapPalette))
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_contour` draws contour lines of a matrix at automatically chosen levels; an ml tensor works as the input",
			signature:   "plot_contour(f: figure, m: matrix|tensor|list) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_contour(f, loss_surface_tensor) => null",
		}.String(),
	},
	{
		Name: "_plot_align",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("plot_align", []int{1, 2, 3, 4}, args); err != nil {
				return err
			}
			// The layout is a list of rows, each a list of figures:
			// [[f1, f2], [f3, f4]].
			rowsObj, ok := args[0].(*List)
			if !ok {
				return newPositionalTypeError("plot_align", 1, "LIST of rows", args[0].Type())
			}
			if len(rowsObj.Elements) == 0 {
				return newError("`plot_align` error: no rows")
			}
			rows := make([][]*plot.Plot, len(rowsObj.Elements))
			ncols := 0
			for i, r := range rowsObj.Elements {
				rowList, ok := r.(*List)
				if !ok {
					return newPositionalTypeError("plot_align", 1, "LIST of rows", r.Type())
				}
				rows[i] = make([]*plot.Plot, len(rowList.Elements))
				if len(rowList.Elements) > ncols {
					ncols = len(rowList.Elements)
				}
				for j, e := range rowList.Elements {
					fg, errObj := numFigureArg("plot_align", 1, e)
					if errObj != nil {
						return errObj
					}
					rows[i][j] = fg.Plot
				}
			}
			if ncols == 0 {
				return newError("`plot_align` error: no columns")
			}
			w, h, errObj := figureDimensions(argOrNil(args, 1), argOrNil(args, 2))
			if errObj != nil {
				return errObj
			}
			format := "png"
			if len(args) == 4 {
				s, ok := args[3].(*Stringo)
				if !ok {
					return newPositionalTypeError("plot_align", 4, STRING_OBJ, args[3].Type())
				}
				format = s.Value
			}
			c, err := draw.NewFormattedCanvas(w, h, format)
			if err != nil {
				return newError("`plot_align` error: %s", err.Error())
			}
			// Align wants a draw.Canvas, so the writer canvas is wrapped the same
			// way gonum's own WriterTo does it.
			canvases := plot.Align(rows, draw.Tiles{Rows: len(rows), Cols: ncols}, draw.New(c))
			for i := range rows {
				for j := range rows[i] {
					if rows[i][j] != nil {
						rows[i][j].Draw(canvases[i][j])
					}
				}
			}
			var buf bytes.Buffer
			if _, err := c.WriteTo(&buf); err != nil {
				return newError("`plot_align` error: %s", err.Error())
			}
			return &Bytes{Value: buf.Bytes()}
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_align` draws figures in a grid with aligned data areas and returns the rendered bytes",
			signature:   "plot_align(rows: list[list[figure]], width: float=6.0, height: float=4.0, format: str='png') -> bytes",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_align([[f1, f2], [f3, f4]]) => bytes",
		}.String(),
	},
	{
		Name: "_plot_time_x",
		Fun: func(args ...Object) Object {
			if err := checkArgsCount("plot_time_x", []int{1, 2}, args); err != nil {
				return err
			}
			f, errObj := numFigureArg("plot_time_x", 1, args[0])
			if errObj != nil {
				return errObj
			}
			format := time.RFC3339
			if len(args) == 2 {
				s, ok := args[1].(*Stringo)
				if !ok {
					return newPositionalTypeError("plot_time_x", 2, STRING_OBJ, args[1].Type())
				}
				format = s.Value
			}
			f.Plot.X.Tick.Marker = plot.TimeTicks{Format: format, Time: plot.UTCUnixTime}
			return NULL
		},
		HelpStr: helpStrArgs{
			explanation: "`plot_time_x` formats the x axis as Unix timestamps, so seconds-since-epoch data gets date labels",
			signature:   "plot_time_x(f: figure, format: str='RFC3339') -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_time_x(f) => null",
		}.String(),
	},
	{
		// vm bound: plotting a blue function needs the vm to call the closure.
		Name: "_plot_function",
		Fun:  nil,
		HelpStr: helpStrArgs{
			explanation: "`plot_function` plots a blue function over an interval, sampling it at evenly spaced points",
			signature:   "plot_function(f: figure, fn: fun, min: float, max: float, samples: int=200, name: str|null=null) -> null",
			errors:      "InvalidArgCount,PositionalType,CustomError",
			example:     "plot_function(f, fun(x) { return x * x; }, -1.0, 1.0) => null",
		}.String(),
	},
}

// plotAxisLabel builds a setter that takes a figure and a string.
func plotAxisLabel(name string, set func(*NumFigure, string)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 2, args); err != nil {
			return err
		}
		f, errObj := numFigureArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		s, ok := args[1].(*Stringo)
		if !ok {
			return newPositionalTypeError(name, 2, STRING_OBJ, args[1].Type())
		}
		set(f, s.Value)
		return NULL
	}
}

// plotFlag builds a figure-only action.
func plotFlag(name string, act func(*NumFigure)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 1, args); err != nil {
			return err
		}
		f, errObj := numFigureArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		act(f)
		return NULL
	}
}

// plotRange builds a figure plus two numbers action.
func plotRange(name string, act func(*NumFigure, float64, float64)) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgCount(name, 3, args); err != nil {
			return err
		}
		f, errObj := numFigureArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		lo, errObj := numFloatArg(name, 2, args[1])
		if errObj != nil {
			return errObj
		}
		hi, errObj := numFloatArg(name, 3, args[2])
		if errObj != nil {
			return errObj
		}
		act(f, lo, hi)
		return NULL
	}
}

// plotXYSeries builds a line or scatter adder. It accepts either
// (figure, points) or (figure, xs, ys), plus optional name and color.
func plotXYSeries(name, kind string) func(...Object) Object {
	return func(args ...Object) Object {
		if err := checkArgsCount(name, []int{2, 3, 4, 5}, args); err != nil {
			return err
		}
		f, errObj := numFigureArg(name, 1, args[0])
		if errObj != nil {
			return errObj
		}
		var xs, ys []float64
		rest := 2
		if len(args) >= 3 {
			if _, ok := args[2].(*List); ok {
				// (f, ys, labels/name) is ambiguous, so a 3 argument call means
				// a single series plus an optional name. Two data arguments are
				// only read when the third is a series-like value.
				_ = ok
			}
		}
		// Decide between the one-series and two-series forms by looking at
		// whether the third argument is data (list/tensor/matrix) or a name.
		if len(args) >= 3 && isSeriesLike(args[2]) {
			xs, errObj = numSeriesArg(name, 2, args[1])
			if errObj != nil {
				return errObj
			}
			ys, errObj = numSeriesArg(name, 3, args[2])
			if errObj != nil {
				return errObj
			}
			rest = 3
		} else {
			xs, ys, errObj = numPointSeriesArg(name, 2, args[1])
			if errObj != nil {
				return errObj
			}
		}
		seriesName := ""
		var col color.Color
		for i := rest; i < len(args); i++ {
			if _, isNull := args[i].(*Null); isNull {
				continue
			}
			if s, ok := args[i].(*Stringo); ok {
				if c, ok := vgColor(hexColor(s.Value)); ok {
					col = color.RGBA(c)
					continue
				}
				if seriesName == "" {
					seriesName = s.Value
					continue
				}
			}
		}
		pts, err := xysFrom(xs, ys)
		if err != nil {
			return newError("`%s` error: %s", name, err.Error())
		}
		switch kind {
		case "line":
			line, err := plotter.NewLine(pts)
			if err != nil {
				return newError("`%s` error: %s", name, err.Error())
			}
			if col != nil {
				line.Color = col
			}
			f.Plot.Add(line)
			if seriesName != "" {
				f.Plot.Legend.Add(seriesName, line)
			}
		case "scatter":
			sc, err := plotter.NewScatter(pts)
			if err != nil {
				return newError("`%s` error: %s", name, err.Error())
			}
			if col != nil {
				if c, ok := col.(color.RGBA); ok {
					sc.Color = c
				}
			}
			f.Plot.Add(sc)
			if seriesName != "" {
				f.Plot.Legend.Add(seriesName, sc)
			}
		}
		return NULL
	}
}

// isSeriesLike reports whether an object is data for a series rather than a name.
func isSeriesLike(o Object) bool {
	switch o.(type) {
	case *List, *Tensor, *GoObj[*NumMatrix]:
		return true
	default:
		return false
	}
}

// numGrid adapts a *mat.Dense to the plotter.GridXYZ interface, so a matrix or
// an ml tensor can be drawn as a heatmap.
type numGrid struct {
	m interface {
		Dims() (int, int)
		At(int, int) float64
	}
}

func (g *numGrid) Dims() (c, r int) {
	r, c = g.m.Dims()
	return c, r
}

func (g *numGrid) Z(c, r int) float64 { return g.m.At(r, c) }
func (g *numGrid) X(c int) float64    { return float64(c) }
func (g *numGrid) Y(r int) float64    { return float64(r) }

// heapPalette is the default heatmap palette. A sequential ColorBrewer ramp
// reads well for weight matrices, and it satisfies palette.Palette, which
// plotter.NewHeatMap requires.
var heapPalette = largestSequential(brewer.Blues)

// largestSequential returns the largest palette in a brewer sequential set, so a
// caller does not have to know the available sizes.
func largestSequential(m map[int]brewer.NonDivergingPalette) *brewer.NonDivergingPalette {
	best := -1
	for k := range m {
		if k > best {
			best = k
		}
	}
	p := m[best]
	return &p
}

// brewerPalette looks up a ColorBrewer palette by name and returns its colors.
func brewerPalette(name string) ([]color.Color, bool) {
	// The ColorBrewer palettes are keyed by how many colors they hold. Pick a
	// size close to a typical series count, falling back to the largest
	// available. The qualitative palettes are the ones that make sense for
	// distinct series, so those are what this exposes.
	pick := func(m map[int]brewer.NonDivergingPalette) ([]color.Color, bool) {
		for _, want := range []int{8, 7, 9, 6, 5, 4, 3, 10, 11, 12} {
			if p, ok := m[want]; ok {
				return p.Color, true
			}
		}
		for _, p := range m {
			return p.Color, true
		}
		return nil, false
	}
	switch name {
	case "dark", "dark2", "Dark2":
		return pick(brewer.Dark2)
	case "set1", "Set1":
		return pick(brewer.Set1)
	case "set2", "Set2":
		return pick(brewer.Set2)
	case "set3", "Set3":
		return pick(brewer.Set3)
	case "paired", "Paired":
		return pick(brewer.Paired)
	case "accent", "Accent":
		return pick(brewer.Accent)
	case "pastel1", "Pastel1":
		return pick(brewer.Pastel1)
	case "pastel2", "Pastel2":
		return pick(brewer.Pastel2)
	default:
		return nil, false
	}
}

// renderFigure renders a figure to bytes. plot_render and plot_save both call
// this, so there is exactly one rendering path.
func renderFigure(f *NumFigure, format string, wArg, hArg Object) ([]byte, Object) {
	w, h, errObj := figureDimensions(wArg, hArg)
	if errObj != nil {
		return nil, errObj
	}
	wt, err := f.Plot.WriterTo(w, h, format)
	if err != nil {
		return nil, newError("render error: %s", err.Error())
	}
	var buf bytes.Buffer
	if _, err := wt.WriteTo(&buf); err != nil {
		return nil, newError("render error: %s", err.Error())
	}
	return buf.Bytes(), nil
}

// writeBytesToPath writes rendered bytes to a file and returns NULL on success.
func writeBytesToPath(path string, data []byte) Object {
	fname := filepath.FromSlash(path)
	if err := os.WriteFile(fname, data, 0644); err != nil {
		return newError("`plot_save` error writing `%s`: %s", path, err.Error())
	}
	return NULL
}

// formatFromPath picks a plot format from a file extension, defaulting to png.
// gonum's Save does this internally; plot_render returns bytes, so plot_save has
// to do it here.
func formatFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "png"
	case ".svg":
		return "svg"
	case ".pdf":
		return "pdf"
	case ".eps":
		return "eps"
	case ".jpg", ".jpeg":
		return "jpg"
	case ".tif", ".tiff":
		return "tif"
	case ".tex":
		return "tex"
	default:
		return "png"
	}
}

// argOrNil returns args[i] or nil when the index is past the end.
func argOrNil(args []Object, i int) Object {
	if i < len(args) {
		return args[i]
	}
	return nil
}

// numBoxGroups reads a box plot argument as one or more groups of samples. A flat
// list or a 1d tensor is one group; a list of lists, or a 2d matrix, gives one
// group per row.
func numBoxGroups(name string, o Object) ([][]float64, Object) {
	if l, ok := o.(*List); ok && len(l.Elements) > 0 {
		if _, isList := l.Elements[0].(*List); isList {
			out := make([][]float64, len(l.Elements))
			for i, e := range l.Elements {
				g, errObj := numSeriesArg(name, 2, e)
				if errObj != nil {
					return nil, errObj
				}
				out[i] = g
			}
			return out, nil
		}
	}
	switch v := o.(type) {
	case *Tensor:
		shape := v.T.Shape()
		if len(shape) == 2 {
			flat, errObj := numSeriesArg(name, 2, o)
			if errObj != nil {
				return nil, errObj
			}
			out := make([][]float64, shape[0])
			for i := 0; i < shape[0]; i++ {
				out[i] = flat[i*shape[1] : (i+1)*shape[1]]
			}
			return out, nil
		}
	case *GoObj[*NumMatrix]:
		r, c := v.Value.Dims()
		flat, errObj := numSeriesArg(name, 2, o)
		if errObj != nil {
			return nil, errObj
		}
		out := make([][]float64, r)
		for i := range r {
			out[i] = flat[i*c : (i+1)*c]
		}
		return out, nil
	}
	g, errObj := numSeriesArg(name, 2, o)
	if errObj != nil {
		return nil, errObj
	}
	return [][]float64{g}, nil
}

// contourLevels picks evenly spaced levels across a matrix's value range. It
// returns nil when the matrix has no range to contour, which the caller reports.
func contourLevels(m interface {
	Dims() (int, int)
	At(int, int) float64
}) []float64 {
	r, c := m.Dims()
	if r == 0 || c == 0 {
		return nil
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for i := range r {
		for j := range c {
			v := m.At(i, j)
			if math.IsNaN(v) {
				continue
			}
			if v < lo {
				lo = v
			}
			if v > hi {
				hi = v
			}
		}
	}
	if math.IsInf(lo, 1) || math.IsInf(hi, -1) || hi <= lo {
		return nil
	}
	const n = 5
	out := make([]float64, n)
	for i := range out {
		out[i] = lo + (hi-lo)*float64(i+1)/float64(n+1)
	}
	return out
}

// PlotAddSampledFunction adds a line series built from already sampled points.
// The vm samples a blue function and calls this, because only the vm can invoke a
// closure while only this package knows how to build a gonum plotter.
func PlotAddSampledFunction(f *NumFigure, xs, ys []float64, name string) Object {
	if f == nil {
		return newError("`function` error: nil figure")
	}
	pts, err := xysFrom(xs, ys)
	if err != nil {
		return newError("`function` error: %s", err.Error())
	}
	line, err := plotter.NewLine(pts)
	if err != nil {
		return newError("`function` error: %s", err.Error())
	}
	f.Plot.Add(line)
	if name != "" {
		f.Plot.Legend.Add(name, line)
	}
	return nil
}
