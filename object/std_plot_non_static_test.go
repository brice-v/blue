//go:build !static

package object

import (
	"bytes"
	"encoding/binary"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"strings"
	"testing"
)

func newTestFigure() *NumFigure { return &NumFigure{Plot: plot.New()} }

// pngDimensions reads width and height from a PNG IHDR chunk.
func pngDimensions(t *testing.T, b []byte) (int, int) {
	t.Helper()
	if len(b) < 24 {
		t.Fatalf("png too short: %d bytes", len(b))
	}
	w := int(binary.BigEndian.Uint32(b[16:20]))
	h := int(binary.BigEndian.Uint32(b[20:24]))
	return w, h
}

func isPNG(b []byte) bool {
	return len(b) >= 8 && bytes.Equal(b[:8], []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})
}

func TestPlotRenderPNG(t *testing.T) {
	f := newTestFigure()
	line, err := lineFromSlices([]float64{1, 2, 3}, []float64{3, 2, 1})
	if err != nil {
		t.Fatalf("lineFromSlices: %v", err)
	}
	f.Plot.Add(line)

	data, errObj := renderFigure(f, "png", nil, nil)
	if errObj != nil {
		t.Fatalf("renderFigure: %s", errObj.Inspect())
	}
	if !isPNG(data) {
		t.Fatal("render did not produce a PNG")
	}
	w, h := pngDimensions(t, data)
	if w == 0 || h == 0 {
		t.Fatalf("png dimensions are %dx%d", w, h)
	}
}

func TestPlotRenderSVGHasText(t *testing.T) {
	// This is the font test: text must actually be laid out, which proves the
	// embedded fonts are working and no external font files are needed.
	f := newTestFigure()
	f.Plot.Title.Text = "probe title"
	f.Plot.X.Label.Text = "xlabel"
	f.Plot.Y.Label.Text = "ylabel"
	line, err := lineFromSlices([]float64{0, 1}, []float64{0, 1})
	if err != nil {
		t.Fatalf("lineFromSlices: %v", err)
	}
	f.Plot.Add(line)

	data, errObj := renderFigure(f, "svg", nil, nil)
	if errObj != nil {
		t.Fatalf("renderFigure: %s", errObj.Inspect())
	}
	s := string(data)
	if !strings.Contains(s, "<svg") {
		t.Fatal("render did not produce SVG")
	}
	if !strings.Contains(s, "probe title") {
		t.Fatal("the title text is missing from the SVG, so text rendering failed")
	}
	if !strings.Contains(s, "xlabel") || !strings.Contains(s, "ylabel") {
		t.Fatal("axis label text is missing from the SVG")
	}
}

func TestPlotRenderFormats(t *testing.T) {
	f := newTestFigure()
	line, err := lineFromSlices([]float64{0, 1}, []float64{0, 1})
	if err != nil {
		t.Fatalf("lineFromSlices: %v", err)
	}
	f.Plot.Add(line)
	for _, format := range []string{"png", "svg", "pdf", "eps", "jpg", "tif", "tex"} {
		data, errObj := renderFigure(f, format, nil, nil)
		if errObj != nil {
			t.Fatalf("render %s: %s", format, errObj.Inspect())
		}
		if len(data) == 0 {
			t.Fatalf("render %s produced no bytes", format)
		}
	}
}

func TestPlotRenderRejectsUnknownFormat(t *testing.T) {
	f := newTestFigure()
	if _, errObj := renderFigure(f, "not-a-format", nil, nil); errObj == nil {
		t.Fatal("render should reject an unknown format")
	}
}

func TestPlotDimensions(t *testing.T) {
	f := newTestFigure()
	data, errObj := renderFigure(f, "png", &Float{Value: 4}, &Float{Value: 3})
	if errObj != nil {
		t.Fatalf("renderFigure: %s", errObj.Inspect())
	}
	// 4 by 3 inches at 96 DPI is 384 by 288 pixels.
	w, h := pngDimensions(t, data)
	if w != 384 || h != 288 {
		t.Fatalf("dimensions %dx%d, want 384x288", w, h)
	}
}

func TestPlotFormatFromPath(t *testing.T) {
	cases := map[string]string{
		"a.png": "png", "a.svg": "svg", "a.PDF": "pdf", "a.eps": "eps",
		"a.jpg": "jpg", "a.jpeg": "jpg", "a.tif": "tif", "a.tex": "tex",
		"a.unknown": "png", "noext": "png",
	}
	for path, want := range cases {
		if got := formatFromPath(path); got != want {
			t.Fatalf("formatFromPath(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestPlotPointSeriesFromList(t *testing.T) {
	l := &List{Elements: []Object{
		&List{Elements: []Object{&Float{Value: 1}, &Float{Value: 2}}},
		&List{Elements: []Object{&Float{Value: 3}, &Float{Value: 4}}},
	}}
	xs, ys, errObj := numPointSeriesArg("test", 1, l)
	if errObj != nil {
		t.Fatalf("numPointSeriesArg: %s", errObj.Inspect())
	}
	if len(xs) != 2 || xs[0] != 1 || xs[1] != 3 || ys[0] != 2 || ys[1] != 4 {
		t.Fatalf("points = %v %v", xs, ys)
	}
}

func TestPlotPointSeriesFromFlatList(t *testing.T) {
	l := &List{Elements: []Object{&Float{Value: 5}, &Float{Value: 6}, &Float{Value: 7}}}
	xs, ys, errObj := numPointSeriesArg("test", 1, l)
	if errObj != nil {
		t.Fatalf("numPointSeriesArg: %s", errObj.Inspect())
	}
	// A flat list becomes y against its index.
	if len(xs) != 3 || xs[0] != 0 || xs[2] != 2 {
		t.Fatalf("xs = %v, want the index", xs)
	}
	if ys[0] != 5 || ys[2] != 7 {
		t.Fatalf("ys = %v", ys)
	}
}

func TestPlotRejectsRaggedPoints(t *testing.T) {
	l := &List{Elements: []Object{
		&List{Elements: []Object{&Float{Value: 1}, &Float{Value: 2}}},
		&List{Elements: []Object{&Float{Value: 3}}},
	}}
	if _, _, errObj := numPointSeriesArg("test", 1, l); errObj == nil {
		t.Fatal("a ragged point list should be rejected")
	}
}

func TestPlotPalette(t *testing.T) {
	pal, ok := brewerPalette("set1")
	if !ok {
		t.Fatal("set1 palette not found")
	}
	if len(pal) == 0 {
		t.Fatal("set1 palette is empty")
	}
	if _, ok := brewerPalette("nope"); ok {
		t.Fatal("an unknown palette should not resolve")
	}
}

// lineFromSlices builds a plotter line, so tests do not repeat the conversion.
func lineFromSlices(xs, ys []float64) (*plotter.Line, error) {
	pts, err := xysFrom(xs, ys)
	if err != nil {
		return nil, err
	}
	return plotter.NewLine(pts)
}

func TestPlotErrorBars(t *testing.T) {
	f := newTestFigure()
	f.Plot.Add(&plotter.YErrorBars{})
	// The real check is that rendering succeeds with bars and points present.
	pts, err := xysFrom([]float64{1, 2, 3}, []float64{2, 3, 2.5})
	if err != nil {
		t.Fatalf("xysFrom: %v", err)
	}
	sc, err := plotter.NewScatter(pts)
	if err != nil {
		t.Fatalf("NewScatter: %v", err)
	}
	f.Plot.Add(sc)
	bars := &plotter.YErrorBars{XYs: pts}
	bars.YErrors = make(plotter.YErrors, len(pts))
	for i := range pts {
		bars.YErrors[i].Low = -0.1
		bars.YErrors[i].High = 0.1
	}
	f.Plot.Add(bars)
	data, errObj := renderFigure(f, "png", nil, nil)
	if errObj != nil {
		t.Fatalf("renderFigure: %s", errObj.Inspect())
	}
	if !isPNG(data) {
		t.Fatal("error bar figure did not render to a PNG")
	}
}

func TestPlotBoxGroupsFromTensor(t *testing.T) {
	tensor := newTensorForTest(t, []float32{1, 2, 3, 4, 5, 6}, []int{2, 3})
	groups, errObj := numBoxGroups("test", &Tensor{T: tensor})
	if errObj != nil {
		t.Fatalf("numBoxGroups: %s", errObj.Inspect())
	}
	if len(groups) != 2 {
		t.Fatalf("a 2x3 tensor gave %d groups, want 2", len(groups))
	}
	if len(groups[0]) != 3 || groups[0][0] != 1 || groups[1][2] != 6 {
		t.Fatalf("groups split wrong: %v", groups)
	}
}

func TestPlotBoxGroupsFromListOfLists(t *testing.T) {
	l := &List{Elements: []Object{
		&List{Elements: []Object{&Float{Value: 1}, &Float{Value: 2}}},
		&List{Elements: []Object{&Float{Value: 3}, &Float{Value: 4}}},
	}}
	groups, errObj := numBoxGroups("test", l)
	if errObj != nil {
		t.Fatalf("numBoxGroups: %s", errObj.Inspect())
	}
	if len(groups) != 2 || groups[0][0] != 1 || groups[1][1] != 4 {
		t.Fatalf("groups = %v", groups)
	}
}

func TestPlotBoxGroupsFlatList(t *testing.T) {
	l := &List{Elements: []Object{&Float{Value: 1}, &Float{Value: 2}, &Float{Value: 3}}}
	groups, errObj := numBoxGroups("test", l)
	if errObj != nil {
		t.Fatalf("numBoxGroups: %s", errObj.Inspect())
	}
	if len(groups) != 1 || len(groups[0]) != 3 {
		t.Fatalf("a flat list should be one group, got %v", groups)
	}
}

func TestPlotContourLevels(t *testing.T) {
	m := testDense(t, [][]float64{{0, 1, 2}, {1, 2, 3}, {2, 3, 4}})
	levels := contourLevels(m)
	if len(levels) != 5 {
		t.Fatalf("got %d levels, want 5", len(levels))
	}
	// Levels must be strictly inside the data range and increasing.
	for i, v := range levels {
		if v <= 0 || v >= 4 {
			t.Fatalf("level %d is %v, outside (0, 4)", i, v)
		}
		if i > 0 && v <= levels[i-1] {
			t.Fatalf("levels are not increasing: %v", levels)
		}
	}
}

func TestPlotContourLevelsFlatMatrix(t *testing.T) {
	// A constant matrix has no range, so there is nothing to contour.
	m := testDense(t, [][]float64{{1, 1}, {1, 1}})
	if levels := contourLevels(m); levels != nil {
		t.Fatalf("a constant matrix should give no levels, got %v", levels)
	}
}

func TestPlotAlignRenders(t *testing.T) {
	f := newTestFigure()
	line, err := lineFromSlices([]float64{1, 2}, []float64{1, 2})
	if err != nil {
		t.Fatalf("lineFromSlices: %v", err)
	}
	f.Plot.Add(line)
	// Call the builtin directly, since the layout argument is a list of lists.
	rows := &List{Elements: []Object{
		&List{Elements: []Object{&GoObj[*NumFigure]{Value: f}}},
	}}
	builtin := findPlotBuiltin(t, "_plot_align")
	got := builtin.Fun(rows)
	bs, ok := got.(*Bytes)
	if !ok {
		t.Fatalf("align did not return bytes: %s", got.Inspect())
	}
	if !isPNG(bs.Value) {
		t.Fatal("aligned grid is not a PNG")
	}
}

func TestPlotAlignRejectsBadLayout(t *testing.T) {
	builtin := findPlotBuiltin(t, "_plot_align")
	if got := builtin.Fun(&Stringo{Value: "nope"}); got.Type() != ERROR_OBJ {
		t.Fatal("align should reject a non-list layout")
	}
	empty := &List{}
	if got := builtin.Fun(empty); got.Type() != ERROR_OBJ {
		t.Fatal("align should reject an empty layout")
	}
}

func TestPlotTimeAxisTicks(t *testing.T) {
	f := newTestFigure()
	builtin := findPlotBuiltin(t, "_plot_time_x")
	if got := builtin.Fun(&GoObj[*NumFigure]{Value: f}, &Stringo{Value: "2006-01-02"}); got.Type() == ERROR_OBJ {
		t.Fatalf("plot_time_x failed: %s", got.Inspect())
	}
	if f.Plot.X.Tick.Marker == nil {
		t.Fatal("time axis marker was not set")
	}
	// The marker must produce RFC-style labels for a Unix timestamp range.
	ticks := f.Plot.X.Tick.Marker.Ticks(1600000000, 1600172800)
	if len(ticks) == 0 {
		t.Fatal("time axis produced no ticks")
	}
	if !strings.Contains(ticks[0].Label, "-") {
		t.Fatalf("time tick label %q does not look like a date", ticks[0].Label)
	}
}

func TestPlotSampledFunctionAddsSeries(t *testing.T) {
	f := newTestFigure()
	xs := []float64{-1, 0, 1}
	ys := []float64{1, 0, 1}
	if obj := PlotAddSampledFunction(f, xs, ys, "squared"); obj != nil {
		t.Fatalf("PlotAddSampledFunction: %s", obj.Inspect())
	}
	data, errObj := renderFigure(f, "svg", nil, nil)
	if errObj != nil {
		t.Fatalf("renderFigure: %s", errObj.Inspect())
	}
	if !strings.Contains(string(data), "squared") {
		t.Fatal("the sampled function's legend entry is missing from the SVG")
	}
}

func TestPlotSampledFunctionRejectsMismatch(t *testing.T) {
	f := newTestFigure()
	if obj := PlotAddSampledFunction(f, []float64{1, 2}, []float64{1}, ""); obj == nil {
		t.Fatal("a length mismatch should be reported")
	}
}

// findPlotBuiltin looks up a plot builtin by name, failing when it is missing.
func findPlotBuiltin(t *testing.T, name string) *Builtin {
	t.Helper()
	for _, b := range PlotBuiltins {
		if b.Name == name {
			return b
		}
	}
	t.Fatalf("plot builtin %q not found", name)
	return nil
}
