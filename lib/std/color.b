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


# no styling
val normal = _normal();

# colors
val red = _red();
val cyan = _cyan();
val gray =  _gray();
val blue = _blue();
val black = _black();
val green = _green();
val white = _white();
val yellow = _yellow();
val magenta = _magenta();

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