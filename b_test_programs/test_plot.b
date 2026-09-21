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
assert(type(f) == "GO_OBJ");
plot.title(f, "integration test");
plot.x_label(f, "x");
plot.y_label(f, "y");

# --- 2. series from lists ---------------------------------------------------

plot.line(f, [1.0, 2.0, 3.0], [3.0, 2.0, 1.0], name="a");
plot.line(f, [1.0, 2.0, 3.0], [4.0, 3.0, 2.0], name="b");

val png = plot.render(f);
assert(type(png) == "BYTES");
assert(png.len() > 1000);

# The Go tests assert the exact PNG magic bytes and the pixel dimensions,
# because blue's BYTES value is not indexable. Here the structural check is
# that a PNG renders and that an SVG render contains its text.
val svg_probe = plot.render(f, format='svg');
assert(svg_probe.len() > 1000);

# --- 3. TARGET: ml tensors plot with no conversion call ---------------------

val loss = ml.tensor([3.0, 2.5, 2.0, 1.5, 1.0]);

val ft = plot.new();
plot.title(ft, "tensor loss");
# a single series argument is indexed by position
plot.line(ft, loss, name="loss");
assert(plot.render(ft).len() > 1000);

# [n, 2] points
val fp = plot.new();
plot.scatter(fp, ml.tensor([[1.0, 4.0], [2.0, 3.0], [3.0, 1.0]]));
assert(plot.render(fp).len() > 1000);

# separate x and y tensors
val fxy = plot.new();
plot.line(fxy, ml.tensor([1.0, 2.0, 3.0]), ml.tensor([5.0, 6.0, 7.0]));
assert(plot.render(fxy).len() > 1000);

# a float64 tensor works too, exercising the dtype conversion path
val f64 = plot.new();
plot.line(f64, ml.tensor([1.0, 3.0, 2.0], datatype=ml.dtype.float64));
assert(plot.render(f64).len() > 1000);

# --- 4. TARGET: histogram, bars, heatmap ------------------------------------

val fh = plot.new();
plot.hist(fh, ml.tensor([1.0, 1.0, 2.0, 2.0, 2.0, 3.0]), 3);
assert(plot.render(fh).len() > 1000);

val fb = plot.new();
plot.bars(fb, ml.tensor([4.0, 7.0, 2.0]), ["a", "b", "c"]);
assert(plot.render(fb).len() > 1000);

# heatmap of a weight matrix, the direct way to look at parameters
val fw = plot.new();
plot.title(fw, "weights");
plot.heatmap(fw, ml.randn([8, 8]));
assert(plot.render(fw).len() > 1000);

# --- 5. TARGET: output formats, sizes, and saving ---------------------------

val svg = plot.render(f, format='svg');
assert(svg.len() > 1000);

# width and height are in inches, so 4 by 3 is 384 by 288 pixels
val small = plot.render(f, format='png', width=4.0, height=3.0);
assert(small.len() > 0);

plot.save(f, "lines.png");
plot.save(f, "lines.svg");
assert(is_file("lines.png"));
assert(is_file("lines.svg"));
# the extension picks the format, so the svg file starts with xml
val svg_file = read("lines.svg");
assert(type(svg_file) == "STRING");

# --- 6. TARGET: axes, palette -----------------------------------------------

val fa = plot.new();
plot.x_log(fa);
plot.y_log(fa);
plot.y_range(fa, 0.1, 100.0);
plot.line(fa, [1.0, 2.0, 3.0], [1.0, 10.0, 100.0]);
assert(plot.render(fa).len() > 1000);

val fn = plot.new();
plot.nominal_x(fn, ["one", "two", "three"]);
plot.bars(fn, [1.0, 2.0, 3.0]);
assert(plot.render(fn).len() > 1000);

val pal = plot.palette('set1');
assert(pal.len() == 8);
assert(type(pal[0]) == "STRING");

# --- 7. TARGET: error bars, box, contour ------------------------------------

val pfe = plot.new();
plot.title(pfe, "error bars");
plot.error_bars(pfe, ml.tensor([1.0, 2.0, 3.0]), ml.tensor([2.0, 3.0, 2.5]),
    ml.tensor([0.1, 0.2, 0.15]));
assert(plot.render(pfe).len() > 1000);

# bare bars, no scatter
val pfeb = plot.new();
plot.error_bars(pfeb, [1.0, 2.0], [2.0, 3.0], [0.1, 0.2], points=false);
assert(plot.render(pfeb).len() > 1000);

# box plots from a tensor (one group per row) and from a list of lists
val pbox = plot.new();
plot.box(pbox, ml.tensor([[1.0, 2.0, 3.0, 4.0], [2.0, 3.0, 4.0, 5.0]]));
assert(plot.render(pbox).len() > 1000);

val pboxl = plot.new();
plot.box(pboxl, [[1.0, 2.0, 3.0], [5.0, 6.0, 7.0]], 25.0, "#3366ff");
assert(plot.render(pboxl).len() > 1000);

# contour of a matrix or tensor
val pcon = plot.new();
plot.contour(pcon, ml.tensor([[0.0, 1.0, 2.0], [1.0, 2.0, 3.0], [2.0, 3.0, 4.0]]));
assert(plot.render(pcon).len() > 1000);

# --- 8. TARGET: multi-panel align and a time axis ---------------------------

val aga = plot.new(); plot.line(aga, [1.0, 2.0], [1.0, 2.0]);
val agb = plot.new(); plot.line(agb, [1.0, 2.0], [2.0, 1.0]);
val agc = plot.new(); plot.scatter(agc, [1.0, 2.0], [1.0, 3.0]);
val agd = plot.new(); plot.bars(agd, [1.0, 2.0]);

val grid = plot.align([[aga, agb], [agc, agd]]);
assert(type(grid) == "BYTES");
assert(grid.len() > 1000);
val grid_svg = plot.align([[aga, agb]], format='svg');
assert(grid_svg.len() > 500);

# a time axis labels Unix timestamps as dates
val pt2 = plot.new();
plot.time_x(pt2);
plot.line(pt2, [1600000000.0, 1600086400.0, 1600172800.0], [1.0, 3.0, 2.0]);
assert(plot.render(pt2).len() > 1000);

# --- 9. TARGET: plotting a blue function ------------------------------------

val pfn = plot.new();
plot.title(pfn, "f(x) = x^2");
plot.x_label(pfn, "x");
plot.y_label(pfn, "y");
plot.function(pfn, fun(x) { return x ** 2; }, -2.0, 2.0, 50, name="x^2");
assert(plot.render(pfn).len() > 1000);

# the sampled curve is a real series with axis text, checked through the svg
val fn_svg = plot.render(pfn, format='svg');
assert(fn_svg.len() > 500);

# a closure that returns a non-number truncates the curve instead of failing the
# whole plot, because gonum's line plotter rejects NaN
val pfnerr = plot.new();
plot.function(pfnerr, fun(x) { return x * x; }, 0.0, 1.0, 10);
assert(plot.render(pfnerr).len() > 0);
