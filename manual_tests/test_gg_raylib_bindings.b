import gg

fun main() {
    gg.init_window(width=800, height=640, title="Example Raylib Bindings");
    gg.set_target_fps(60);

    for (!gg.window_should_close()) {
        gg.begin_drawing()

        gg.clear_background(gg.color.white)
		gg.draw_text("Congrats! You created your first window!", pos_x=190, pos_y=200, font_size=20, gg.color.light_gray)

        gg.end_drawing()
    }

    gg.close_window();
}

main();