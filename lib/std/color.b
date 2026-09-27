## color will allow the user to print to the console with colors
##
## The first argument to print or println should be a style created
## in this module
##
## Color styles reset at the end of the print
##
## Colors Available:
## red, cyan, gray, blue, black, green, white, yellow, and magenta
##
## Styling Available:
## bold, italic, underlined
##
## All of these are available in the color module as integer constants
##
## Additionally, these color constants can be used anywhere else
## throughout blue that accepts a color.

val _cm = _color_map();

# no styling
val normal = _normal();

# colors
val beige = _cm.beige;
val black = _cm.black;
val blank = _cm.blank;
val blue = _cm.blue;
val brown = _cm.brown;
val cyan = _cm.cyan;
val dark_blue = _cm.dark_blue;
val dark_brown = _cm.dark_brown;
val dark_gray = _cm.dark_gray;
val dark_green = _cm.dark_green;
val dark_grey = _cm.dark_grey;
val dark_purple = _cm.dark_purple;
val gold = _cm.gold;
val gray = _cm.gray;
val green = _cm.green;
val grey = _cm.grey;
val light_gray = _cm.light_gray;
val light_grey = _cm.light_grey;
val lime = _cm.lime;
val magenta = _cm.magenta;
val maroon = _cm.maroon;
val orange = _cm.orange;
val pink = _cm.pink;
val purple = _cm.purple;
val ray_white = _cm.ray_white;
val red = _cm.red;
val sky_blue = _cm.sky_blue;
val violet = _cm.violet;
val white = _cm.white;
val yellow = _cm.yellow;
val new = _new;
println("_cm.new_color = #{_cm.new_color}")

# styles
val bold = _bold();
val italic = _italic();
val underlined = _underlined();

var __style = _style;

fun style(text=normal, fg_color=normal, bg_color=normal) {
    ##std:this,__style
    ## `style` takes a text style, foreground color, and background color
    ## to create a style object of shape {t: 'color', v: _}
    ##
    ## style(text: int=normal, fg_color: int=normal, bg_color: int=normal) -> {t: 'color', v: uint}
    __style(text, fg_color, bg_color)
}