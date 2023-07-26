import gg

val player_speed = 3.0;
var player_dst = gg.Rectangle(200, 200, 100, 100);


player_dst.x -= player_speed;
println("player_dst.x = #{player_dst.x}");
assert(player_dst.x == 197.0);

val x_coord = player_dst.x-(player_dst.width/2);
println("x_coord = #{x_coord}");
assert(x_coord == 147.0, "x_coord should be 147.0, got=#{x_coord}");
