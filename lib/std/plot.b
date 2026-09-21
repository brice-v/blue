## `plot` is the module that draws figures.
##
## It wraps gonum/plot, which is pure Go: no CGO, so it works in static and wasm
## builds. Rendering returns bytes, so `plot_save` writes a file and `plot_render`
## hands the bytes back for the `ui` or `wasm` modules to display.
##
## Every series function accepts a list, an ml tensor, or a num matrix, so a
## training loop plots a loss tensor with no conversion call. Pass two arguments
## for `(x, y)` or one for a series indexed by position, and either a list of
## `[x, y]` pairs or an `[n, 2]` tensor for explicit points.
##
## Supported formats are png, svg, pdf, eps, jpg, and tex. Text renders with the
## Liberation fonts embedded in the binary, so no external font files are needed.

val __plot_new = _plot_new;
val __plot_title = _plot_title;
val __plot_x_label = _plot_x_label;
val __plot_y_label = _plot_y_label;
val __plot_x_log = _plot_x_log;
val __plot_y_log = _plot_y_log;
val __plot_x_range = _plot_x_range;
val __plot_y_range = _plot_y_range;
val __plot_nominal_x = _plot_nominal_x;
val __plot_line = _plot_line;
val __plot_scatter = _plot_scatter;
val __plot_bars = _plot_bars;
val __plot_hist = _plot_hist;
val __plot_heatmap = _plot_heatmap;
val __plot_render = _plot_render;
val __plot_save = _plot_save;
val __plot_palette = _plot_palette;
val __plot_error_bars = _plot_error_bars;
val __plot_box = _plot_box;
val __plot_contour = _plot_contour;
val __plot_align = _plot_align;
val __plot_time_x = _plot_time_x;
val __plot_function = _plot_function;

fun new() {
    ##std:this,__plot_new
    ## `new` returns a new empty figure.
    ##
    ## new() -> figure
    __plot_new()
}

fun title(f, text) {
    ##std:this,__plot_title
    ## `title` sets the figure title.
    ##
    ## title(f: figure, text: str) -> null
    __plot_title(f, text)
}

fun x_label(f, text) {
    ##std:this,__plot_x_label
    ## `x_label` sets the x axis label.
    ##
    ## x_label(f: figure, text: str) -> null
    __plot_x_label(f, text)
}

fun y_label(f, text) {
    ##std:this,__plot_y_label
    ## `y_label` sets the y axis label.
    ##
    ## y_label(f: figure, text: str) -> null
    __plot_y_label(f, text)
}

fun x_log(f) {
    ##std:this,__plot_x_log
    ## `x_log` switches the x axis to a log scale.
    ##
    ## x_log(f: figure) -> null
    __plot_x_log(f)
}

fun y_log(f) {
    ##std:this,__plot_y_log
    ## `y_log` switches the y axis to a log scale.
    ##
    ## y_log(f: figure) -> null
    __plot_y_log(f)
}

fun x_range(f, min_val, max_val) {
    ##std:this,__plot_x_range
    ## `x_range` pins the x axis range.
    ##
    ## x_range(f: figure, min: float, max: float) -> null
    __plot_x_range(f, min_val, max_val)
}

fun y_range(f, min_val, max_val) {
    ##std:this,__plot_y_range
    ## `y_range` pins the y axis range.
    ##
    ## y_range(f: figure, min: float, max: float) -> null
    __plot_y_range(f, min_val, max_val)
}

fun nominal_x(f, names) {
    ##std:this,__plot_nominal_x
    ## `nominal_x` labels the x axis with names instead of numbers.
    ##
    ## nominal_x(f: figure, names: list[str]) -> null
    __plot_nominal_x(f, names)
}

fun line(f, xs, ys=null, name=null, color=null) {
    ##std:this,__plot_line
    ## `line` adds a line series. With one data argument, the series is indexed
    ## by position. It accepts a list, an ml tensor, or a num matrix.
    ##
    ## line(f: figure, xs: list|tensor, ys: list|tensor|null=null, name: str|null=null, color: str|null=null) -> null
    __plot_line(f, xs, ys, name, color)
}

fun scatter(f, xs, ys=null, name=null, color=null) {
    ##std:this,__plot_scatter
    ## `scatter` adds a scatter series, with the same argument forms as `line`.
    ##
    ## scatter(f: figure, xs: list|tensor, ys: list|tensor|null=null, name: str|null=null, color: str|null=null) -> null
    __plot_scatter(f, xs, ys, name, color)
}

fun bars(f, values, labels=null, color=null) {
    ##std:this,__plot_bars
    ## `bars` adds a bar chart, with optional x labels.
    ##
    ## bars(f: figure, values: list|tensor, labels: list[str]|null=null, color: str|null=null) -> null
    __plot_bars(f, values, labels, color)
}

fun hist(f, values, bins=null, color=null) {
    ##std:this,__plot_hist
    ## `hist` adds a histogram of the values.
    ##
    ## hist(f: figure, values: list|tensor, bins: int|null=null, color: str|null=null) -> null
    __plot_hist(f, values, bins, color)
}

fun heatmap(f, m) {
    ##std:this,__plot_heatmap
    ## `heatmap` draws a matrix as a heatmap, which is a direct way to look at
    ## model weights or activations. An ml tensor works as the input.
    ##
    ## heatmap(f: figure, m: matrix|tensor|list) -> null
    __plot_heatmap(f, m)
}

fun render(f, format='png', width=6.0, height=4.0) {
    ##std:this,__plot_render
    ## `render` renders the figure to bytes. Formats are png, svg, pdf, eps, jpg,
    ## and tex. Width and height are in inches.
    ##
    ## render(f: figure, format: str='png', width: float=6.0, height: float=4.0) -> bytes
    __plot_render(f, format, width, height)
}

fun save(f, path, width=6.0, height=4.0) {
    ##std:this,__plot_save
    ## `save` renders the figure and writes it to a path; the format comes from
    ## the file extension.
    ##
    ## save(f: figure, path: str, width: float=6.0, height: float=4.0) -> null
    __plot_save(f, path, width, height)
}

fun palette(name) {
    ##std:this,__plot_palette
    ## `palette` returns a named ColorBrewer palette as a list of hex strings.
    ## Names include dark, set1, set2, set3, paired, accent, pastel1, pastel2.
    ##
    ## palette(name: str) -> list[str]
    __plot_palette(name)
}

fun error_bars(f, x, y, yerr, points=true) {
    ##std:this,__plot_error_bars
    ## `error_bars` adds symmetric vertical error bars at each point, with a
    ## scatter on top by default. Pass `points=false` for bare bars.
    ##
    ## error_bars(f: figure, x: list|tensor, y: list|tensor, yerr: list|tensor, points: bool=true) -> null
    __plot_error_bars(f, x, y, yerr, points)
}

fun box(f, groups, width=20.0, color=null) {
    ##std:this,__plot_box
    ## `box` adds a box plot per group. A list of lists gives several groups and a
    ## flat list gives one; a tensor works too.
    ##
    ## box(f: figure, groups: list|list[list]|tensor, width: float=20.0, color: str|null=null) -> null
    __plot_box(f, groups, width, color)
}

fun contour(f, m) {
    ##std:this,__plot_contour
    ## `contour` draws contour lines of a matrix at automatically chosen levels.
    ## An ml tensor works as the input, which is useful for a loss surface.
    ##
    ## contour(f: figure, m: matrix|tensor|list) -> null
    __plot_contour(f, m)
}

fun align(rows, width=6.0, height=4.0, format='png') {
    ##std:this,__plot_align
    ## `align` draws figures in a grid with aligned data areas and returns the
    ## bytes. The layout is a list of rows: [[f1, f2], [f3, f4]].
    ##
    ## align(rows: list[list[figure]], width: float=6.0, height: float=4.0, format: str='png') -> bytes
    __plot_align(rows, width, height, format)
}

fun time_x(f, format='2006-01-02 15:04') {
    ##std:this,__plot_time_x
    ## `time_x` formats the x axis as Unix timestamps, so seconds-since-epoch data
    ## gets date labels. The format is a Go time layout.
    ##
    ## time_x(f: figure, format: str='2006-01-02 15:04') -> null
    __plot_time_x(f, format)
}

fun function(f, fn, min_val, max_val, samples=200, name=null) {
    ##std:this,__plot_function
    ## `function` plots a blue function over [min, max], sampling it at evenly
    ## spaced points. The coordinates arrive at the function as a single scalar.
    ##
    ## function(f: figure, fn: fun, min: float, max: float, samples: int=200, name: str|null=null) -> null
    __plot_function(f, fn, min_val, max_val, samples, name)
}
