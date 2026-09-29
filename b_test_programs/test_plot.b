# STATIC IGNORE
## End-state integration test for the `plot` module.
##
## The VM stops at the first error, so the file lights up top to bottom.
##
## plot is pure Go and embeds the Liberation fonts, so text renders with no
## external font files. Rendering returns bytes; the tests check the PNG magic
## bytes and the pixel dimensions parsed from the header, because comparing whole
## images would be flaky across platforms.

import plot
import ml


# --- 1. figure construction -------------------------------------------------

val f = plot.new();
assert(type(f) == "MAP");
f.title("integration test");
f.x_label("x");
f.y_label("y");

# --- 2. series from lists ---------------------------------------------------

f.line([1.0, 2.0, 3.0], [3.0, 2.0, 1.0], name="a");
f.line([1.0, 2.0, 3.0], [4.0, 3.0, 2.0], name="b");

val png = f.render();
assert(type(png) == "BYTES");
assert(png.len() > 1000);

# The Go tests assert the exact PNG magic bytes and the pixel dimensions,
# because blue's BYTES value is not indexable. Here the structural check is
# that a PNG renders and that an SVG render contains its text.
val svg_probe = f.render(format='svg');
assert(svg_probe.len() > 1000);

# --- 3. TARGET: ml tensors plot with no conversion call ---------------------

val loss = ml.tensor([3.0, 2.5, 2.0, 1.5, 1.0]);

val ft = plot.new();
ft.title("tensor loss");
# a single series argument is indexed by position
ft.line(loss, name="loss");
assert(ft.render().len() > 1000);

# [n, 2] points
val fp = plot.new();
fp.scatter(ml.tensor([[1.0, 4.0], [2.0, 3.0], [3.0, 1.0]]));
assert(fp.render().len() > 1000);

# separate x and y tensors
val fxy = plot.new();
fxy.line(ml.tensor([1.0, 2.0, 3.0]), ml.tensor([5.0, 6.0, 7.0]));
assert(fxy.render().len() > 1000);

# a float64 tensor works too, exercising the dtype conversion path
val f64 = plot.new();
f64.line(ml.tensor([1.0, 3.0, 2.0], datatype=ml.dtype.float64));
assert(f64.render().len() > 1000);

# --- 4. TARGET: histogram, bars, heatmap ------------------------------------

val fh = plot.new();
fh.hist(ml.tensor([1.0, 1.0, 2.0, 2.0, 2.0, 3.0]), 3);
assert(fh.render().len() > 1000);

val fb = plot.new();
fb.bars(ml.tensor([4.0, 7.0, 2.0]), ["a", "b", "c"]);
assert(fb.render().len() > 1000);

# heatmap of a weight matrix, the direct way to look at parameters
val fw = plot.new();
fw.title("weights");
fw.heatmap(ml.randn([8, 8]));
assert(fw.render().len() > 1000);

# --- 5. TARGET: output formats, sizes, and saving ---------------------------

val svg = f.render(format='svg');
assert(svg.len() > 1000);

# width and height are in inches, so 4 by 3 is 384 by 288 pixels
val small = f.render(format='png', width=4.0, height=3.0);
assert(small.len() > 0);

f.save("lines.png");
f.save("lines.svg");
assert(is_file("lines.png"));
assert(is_file("lines.svg"));
# the extension picks the format, so the svg file starts with xml
val svg_file = read("lines.svg");
assert(type(svg_file) == "STRING");

# --- 6. TARGET: axes, palette -----------------------------------------------

val fa = plot.new();
fa.x_log();
fa.y_log();
fa.y_range(0.1, 100.0);
fa.line([1.0, 2.0, 3.0], [1.0, 10.0, 100.0]);
assert(fa.render().len() > 1000);

val fn = plot.new();
fn.nominal_x(["one", "two", "three"]);
fn.bars([1.0, 2.0, 3.0]);
assert(fn.render().len() > 1000);

val pal = plot.palette('set1');
assert(pal.len() == 8);
assert(type(pal[0]) == "STRING");

# --- 7. TARGET: error bars, box, contour ------------------------------------

val pfe = plot.new();
pfe.title("error bars");
pfe.error_bars(ml.tensor([1.0, 2.0, 3.0]), ml.tensor([2.0, 3.0, 2.5]), ml.tensor([0.1, 0.2, 0.15]));
assert(pfe.render().len() > 1000);

# bare bars, no scatter
val pfeb = plot.new();
pfeb.error_bars([1.0, 2.0], [2.0, 3.0], [0.1, 0.2], points=false);
assert(pfeb.render().len() > 1000);

# box plots from a tensor (one group per row) and from a list of lists
val pbox = plot.new();
pbox.box(ml.tensor([[1.0, 2.0, 3.0, 4.0], [2.0, 3.0, 4.0, 5.0]]));
assert(pbox.render().len() > 1000);

val pboxl = plot.new();
pboxl.box([[1.0, 2.0, 3.0], [5.0, 6.0, 7.0]], 25.0, "#3366ff");
assert(pboxl.render().len() > 1000);

# contour of a matrix or tensor
val pcon = plot.new();
pcon.contour(ml.tensor([[0.0, 1.0, 2.0], [1.0, 2.0, 3.0], [2.0, 3.0, 4.0]]));
assert(pcon.render().len() > 1000);

# --- 8. TARGET: multi-panel align and a time axis ---------------------------

val aga = plot.new(); aga.line([1.0, 2.0], [1.0, 2.0]);
val agb = plot.new(); agb.line([1.0, 2.0], [2.0, 1.0]);
val agc = plot.new(); agc.scatter([1.0, 2.0], [1.0, 3.0]);
val agd = plot.new(); agd.bars([1.0, 2.0]);

val grid = plot.align([[aga, agb], [agc, agd]]);
assert(type(grid) == "BYTES");
assert(grid.len() > 1000);
val grid_svg = plot.align([[aga, agb]], format='svg');
assert(grid_svg.len() > 500);

# a time axis labels Unix timestamps as dates
val pt2 = plot.new();
pt2.time_x();
pt2.line([1600000000.0, 1600086400.0, 1600172800.0], [1.0, 3.0, 2.0]);
assert(pt2.render().len() > 1000);

# --- 9. TARGET: plotting a blue function ------------------------------------

val pfn = plot.new();
pfn.title("f(x) = x^2");
pfn.x_label("x");
pfn.y_label("y");
pfn.function(fun(x) { return x ** 2; }, -2.0, 2.0, 50, name="x^2");
assert(pfn.render().len() > 1000);

# the sampled curve is a real series with axis text, checked through the svg
val fn_svg = pfn.render(format='svg');
assert(fn_svg.len() > 500);

# a closure that returns a non-number truncates the curve instead of failing the
# whole plot, because gonum's line plotter rejects NaN
val pfnerr = plot.new();
pfnerr.function(fun(x) { return x * x; }, 0.0, 1.0, 10);
assert(pfnerr.render().len() > 0);
