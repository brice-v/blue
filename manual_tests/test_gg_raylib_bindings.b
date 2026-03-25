import gg

fun main() {
    val win = gg.new_window(width=800, height=640, title="Example Raylib Bindings");
    gg.set_target_fps(60);
    for [k,v] in gg.monitor {
        println("k = #{k} => #{v}")
        try {
            println("k = #{k} => #{v()}")
        } catch (e) {
            println("k = #{k} => #{v(gg.monitor.get_current())}")
        }
    }
    println("HERE? win #{win}");
    println("win.should_close() = #{win.should_close()}")
    for (!win.should_close()) {
        gg.begin_drawing()

        gg.clear_background(gg.color.white)
		gg.draw_text("Congrats! You created your first window!", pos_x=190, pos_y=200, font_size=20, gg.color.light_gray)

        gg.end_drawing()
    }

    win.close();
}

main();