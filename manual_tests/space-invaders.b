import gg

val SCREEN_WIDTH = 800;
val SCREEN_HEIGHT = 640;
val rec_speed = 5;
val bullet_speed = 10;
val bullet_width = 10;
val bullet_height = 20;

var G = {
    rec: gg.Rectangle(width=70, height=50, x=(SCREEN_WIDTH/2)-(35), y=SCREEN_HEIGHT-50-5),
    bullet: null,
    enemy_direction: 'right',
    enemy_top: 1,
    enemy_speed: 1,
    lost: false,
    enemies: [[],[],[]],
    won: false,
};

val ENEMY_HEIGHT = 40;
val ENEMY_WIDTH = 70;


val COLOR_MAP = {
    0: gg.color.blue,
    1: gg.color.dark_green,
    2: gg.color.violet,
    3: gg.color.brown,
    4: gg.color.sky_blue,
};

fun reset() {
    G.enemies = [[],[],[]];
    for var j = 0; j < 3; j += 1 {
        for (var i = 0; i < 5; i += 1) {
            var enemy_x = (ENEMY_WIDTH*i)+(bullet_width*i)+((SCREEN_WIDTH/4));
            var enemy_y = j*ENEMY_HEIGHT+j+(j*bullet_height);
            G.enemies[j] << gg.Rectangle(width=ENEMY_WIDTH, height=ENEMY_HEIGHT, y=enemy_y, x=enemy_x);
        }
    }
    G.rec = gg.Rectangle(width=70, height=50, x=(SCREEN_WIDTH/2)-(35), y=SCREEN_HEIGHT-50-5);
    G.bullet = null;
    G.enemy_direction = 'right';
    G.enemy_speed = 1;
    G.enemy_top = 1;
    G.lost = false;
    G.won = false;
}


fun g_input() {
    if (G.lost || G.won) && gg.is_key_down(gg.Key.Down) {
        reset();
        return null;
    }
    if gg.is_key_down(gg.Key.Left) {
        G.rec.x -= rec_speed;
        val rec_left_width = (1/2) * G.rec.width;
        if G.rec.x < rec_left_width {
            G.rec.x = rec_left_width
        }
    } else if gg.is_key_down(gg.Key.Right) {
        G.rec.x += rec_speed;
        val rec_right_width = SCREEN_WIDTH - G.rec.width;
        if G.rec.x > rec_right_width {
            G.rec.x = rec_right_width
        }
    } 
    if gg.is_key_down(gg.Key.Up) {
        # Fire Projectile
        if G.bullet == null {
            val bullet_x = G.rec.x+(0.5 * G.rec.width);
            val bullet_y = SCREEN_HEIGHT-G.rec.height;
            G.bullet = gg.Rectangle(width=bullet_width, height=bullet_height, x=bullet_x, y=bullet_y)
        }
    }
}

fun update_enemy_speed() {
    if G.enemy_speed > 8 {
        return null;
    }
    G.enemy_speed += 0.5;
}

fun check_bullet_hit() {
    if !G.bullet {
        return null;
    }
    for ([i, row] in G.enemies) {
        for ([j, enemy] in row) {
            if enemy {
                if enemy.check_collision(G.bullet) {
                    G.enemies[i][j] = null;
                    G.bullet = null;
                    update_enemy_speed();
                    return null;
                }
            }
        }
    }
}

fun fire_bullet() {
    if !G.bullet {
        return null;
    }
    G.bullet.y -= bullet_speed;
    if G.bullet.y < 0 {
        G.bullet = null;
    }
}

fun update_enemy_heights_and_speeds() {
    G.enemy_top += 1;
    update_enemy_speed();
    for var j = 0; j < 3; j += 1 {
        for (var i = 0; i < 5; i += 1) {
            var enemy = G.enemies[j][i];
            if !enemy {
                continue;
            }
            enemy.y += ENEMY_HEIGHT;
        }
    }
}

fun move_enemies() {
    for ([i, row] in G.enemies) {
        for ([j, enemy] in row) {
            if enemy {
                if G.enemy_direction == 'right' {
                    enemy.x += G.enemy_speed;
                } else {
                    enemy.x -= G.enemy_speed;
                }
                if enemy.x > SCREEN_WIDTH-enemy.width {
                    enemy.x -= G.enemy_speed;
                    G.enemy_direction = 'left';
                    update_enemy_heights_and_speeds();
                } else if enemy.x < 0 {
                    enemy.x += G.enemy_speed;
                    G.enemy_direction = 'right';
                    update_enemy_heights_and_speeds();
                }
            }
        }
    }
}

fun check_lost() {
    for ([i, row] in G.enemies) {
        for ([j, enemy] in row) {
            if enemy {
                if enemy.y > SCREEN_HEIGHT-G.rec.height {
                    G.lost = true;
                    return null;
                }
            }
        }
    }
}

fun check_won() {
    for ([i, row] in G.enemies) {
        for ([j, enemy] in row) {
            if enemy {
                G.won = false;
                return null;
            }
        }
    }
    G.won = true;
}

fun g_update() {
    if !G.lost && !G.won {
        fire_bullet();
        check_bullet_hit();
        move_enemies();
    }
    check_won();
    check_lost();
}

fun g_render() {
    gg.begin_drawing()

    gg.clear_background(gg.color.black)
    if G.lost {
        var posx = (SCREEN_WIDTH/3);
        var posy = (SCREEN_HEIGHT/2);
        gg.draw_text("YOU LOST", pos_x=posx, pos_y=posy, text_color=gg.color.white);
        gg.draw_text("press down to continue...", pos_x=posx, pos_y=posy+20);
    } else if G.won {
        var posx = (SCREEN_WIDTH/3);
        var posy = (SCREEN_HEIGHT/2);
        gg.draw_text("YOU WIN!", pos_x=posx, pos_y=posy, text_color=gg.color.white);
        gg.draw_text("press down to continue...", pos_x=posx, pos_y=posy+20);
    } else {
        G.rec.draw()
        if G.bullet {
            G.bullet.draw();
        }
        for (row in G.enemies) {
            for ([i,enemy] in row) {
                if enemy {
                    enemy.draw(COLOR_MAP[i]);
                }
            }
        }
    }

    gg.end_drawing()
}

fun main() {
    gg.init_window(width=SCREEN_WIDTH, height=SCREEN_HEIGHT, title="Space Invaders");
    gg.set_target_fps(60);

    reset();

    for (!gg.window_should_close()) {
        g_input();
        g_update();
        g_render();
    }

    gg.close_window();
}

main();