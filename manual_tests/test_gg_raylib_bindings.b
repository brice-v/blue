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
    for [k,v] in gg.mouse {
        println("k = #{k} => #{v}")
        if is_callable(v) && !k.startswith("set") {
            println("Found callable k = #{k}")
            try {
                println("v() = #{v()}")
            } catch (e) {
                println("v(0) = #{v(0)}")
            }
        }
    }
    println("gg.cursor.is_hidden() = #{gg.cursor.is_hidden()}")
    println("gg.cursor.is_on_screen() = #{gg.cursor.is_on_screen()}")
    val text = "Test Text for gg";
    gg.clipboard.set_text(text);
    assert(gg.clipboard.get_text() == text);
    println("HERE? win #{win}");
    println("win.should_close() = #{win.should_close()}")
    for (!win.should_close()) {
        gg.begin_drawing()

        gg.clear_background(gg.color.white)
		gg.draw_text("Congrats! You created your first window!", pos_x=190, pos_y=200, font_size=20, gg.color.light_gray)
        gg.draw.line(0, 10, 100, 120, gg.color.red)
        gg.draw.line(gg.Vector(10, 50).obj(), gg.Vector(110, 130).obj(), gg.color.blue)
        gg.draw.line(gg.Vector(10, 75).obj(), gg.Vector(110, 130).obj(), 2.0, gg.color.green)
        gg.draw.line_strip([gg.Vector(10, 100).obj(), gg.Vector(10, 110).obj()], gg.color.purple)
        gg.draw.line_bezier(gg.Vector(10, 120).obj(), gg.Vector(10, 130).obj(), 4.0, gg.color.maroon)
        gg.draw.line_bezier(gg.Vector(20, 140).obj(), gg.Vector(20, 150).obj(), gg.Vector(30, 40).obj(), 4.0, gg.color.lime)
        gg.draw.line_bezier(gg.Vector(50, 160).obj(), gg.Vector(50, 180).obj(), gg.Vector(40, 40).obj(), gg.Vector(50,50).obj(), 2.0, gg.color.sky_blue)

        gg.draw.circle(gg.Vector(200,200).obj(), 10.0, gg.color.dark_brown)
        gg.draw.circle(210, 210, 10.0, gg.color.magenta)
        gg.draw.circle(220, 220, 10.0, gg.color.red, gg.color.blue)
        gg.draw.circle_sector(gg.Vector(230,230).obj(), 10.0, 0.0, 90.0, 16, gg.color.beige, with_lines=false)
        gg.draw.circle_sector(gg.Vector(240,240).obj(), 10.0, 0.0, 90.0, 16, gg.color.black)
        gg.draw.circle_sector(gg.Vector(250,250).obj(), 10.0, 0.0, 90.0, 16, gg.color.violet, with_lines=true)
        gg.draw.circle_sector(gg.Vector(260,260).obj(), 10.0, 0.0, 90.0, 16, gg.color.pink, true)
        gg.draw.circle_sector(gg.Vector(270,270).obj(), 10.0, 0.0, 90.0, 16, gg.color.yellow, false)
        gg.draw.circle_lines(280, 280, 12.0, gg.color.orange)

        gg.draw.rectangle_gradient(gg.Rectangle(200,100,10,30).obj(), gg.color.red, gg.color.blue, gg.color.green, gg.color.yellow)
        gg.draw.rectangle_gradient(100,100,10,30, gg.color.red, gg.color.blue, is_vertical=false)
        gg.draw.rectangle_gradient(10,200,10,30, gg.color.red, gg.color.blue, false)
        gg.draw.rectangle_gradient(10,300,10,30, gg.color.red, gg.color.blue, true)
        gg.draw.rectangle_gradient(10,500,10,30, gg.color.red, gg.color.blue, is_vertical=true)
        gg.draw.rectangle_lines(10, 600, 10, 30, gg.color.gray)
        gg.draw.rectangle_lines(gg.Rectangle(100, 600, 10, 30).obj(), 2.0, gg.color.gray)
        gg.draw.rectangle_rounded(gg.Rectangle(500, 40, 10, 30).obj(), 2.0, 10, gg.color.dark_gray)
        gg.draw.rectangle_rounded_lines(gg.Rectangle(500, 40, 10, 30).obj(), 2.0, 5.0, 1.0, gg.color.dark_gray)
        gg.draw.triangle(gg.Vector(600/4*3, 10).obj(), gg.Vector(600/4*3-60, 150).obj(), gg.Vector(600/4*3+60, 150).obj(), gg.color.red)
        gg.draw.triangle(gg.Vector(600/4*3, 20).obj(), gg.Vector(600/4*3-60, 150).obj(), gg.Vector(600/4*3+60, 150).obj(), gg.color.pink, with_lines=true)
        gg.draw.triangle(gg.Vector(600/4*3, 30).obj(), gg.Vector(600/4*3-60, 150).obj(), gg.Vector(600/4*3+60, 150).obj(), gg.color.black, true)
        gg.draw.triangle(gg.Vector(600/4*3, 40).obj(), gg.Vector(600/4*3-60, 150).obj(), gg.Vector(600/4*3+60, 150).obj(), gg.color.dark_blue, false)
        gg.draw.triangle(gg.Vector(600/4*3, 50).obj(), gg.Vector(600/4*3-60, 150).obj(), gg.Vector(600/4*3+60, 150).obj(), gg.color.dark_brown, with_lines=false)
        gg.draw.triangle_fan([gg.Vector(110, 600).obj(),gg.Vector(120, 600).obj(),gg.Vector(130, 400).obj(),gg.Vector(140, 200).obj(),gg.Vector(150, 700).obj()],gg.color.sky_blue)
        gg.draw.triangle_strip([gg.Vector(90, 500).obj(),gg.Vector(120, 500).obj(),gg.Vector(130, 300).obj(),gg.Vector(140, 100).obj(),gg.Vector(150, 800).obj()],gg.color.green)
        gg.draw.poly(gg.Vector(600,600).obj(), 5, 100.0, 0.0, 2.0, gg.color.black)
        gg.draw.poly(gg.Vector(400,400).obj(), 12, 50.0, 0.0, gg.color.magenta, with_lines=true)
        gg.draw.poly(gg.Vector(300,400).obj(), 8, 50.0, 0.0, gg.color.blue, with_lines=false)
        gg.draw.poly(gg.Vector(200,400).obj(), 10, 50.0, 0.0, 2.0, gg.color.red, true)
        gg.draw.poly(gg.Vector(100,400).obj(), 6, 50.0, 0.0, 3.0, gg.color.green, false)

        gg.end_drawing()
    }

    win.close();
}

main();