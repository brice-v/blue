## `gg` is the module that contains functions needed to
## run the games graphics library (Raylib)


val __init_window = _init_window;
val __clear_background = _clear_background;
val color = _color_map();
val begin_drawing = _begin_drawing;
val end_drawing = _end_drawing;
val set_target_fps = _set_target_fps;
val window_should_close = _window_should_close;
val close_window = _close_window;
val __draw_text = _draw_text;

fun init_window(width=800, height=600, title="gg - example app") {
    ## `init_window` will initialize the raylib window
    ##
    ## init_window(width: int=800, height: int=600, title: str="gg - example app") -> null
    __init_window(width, height, title)
}

fun clear_background(clear_color=color.white) {
    ## `clear_background` will use the clear color to clear the background with that color
    ##
    ## clear_background(clear_color: GO_OBJ[rl.Color]=color.white) -> null
    __clear_background(clear_color)
}

fun draw_text(text, pos_x=0, pos_y=0, font_size=20, text_color=color.black) {
    ## `draw_text` will draw the given text to the raylib window with the given parameters
    ##
    ## pos_x the x position of the text
    ## pos_y the y position of the text
    ## font_size the size of the font
    ## text_color the color of the text
    ##
    ## draw_text(text: str, pos_x: int=0, pos_y: int=0, font_size: int=20, text_color: GO_OBJ[rl.Color]=color.black) -> null
    __draw_text(text, pos_x, pos_y, font_size, text_color)
}