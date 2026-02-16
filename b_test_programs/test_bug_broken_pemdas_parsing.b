# Stub
fun __rectangle(x, y, width, height) {
    return null;
}

fun Rectangle(x=0.0, y=0.0, width=0.0, height=0.0) {
    ##std:this,__rectangle
    ## `Rectangle` is an object constructor that represents the GO_OBJ for rectangles
    ##
    ## the object can be modified using the x,y,width,height variables and the GO_OBJ
    ## can be retrieved with this.obj()
    ##
    ## Rectangle(x: float=0.0, y: float=0.0, width: float=0.0, height: float=0.0) ->
    ## {'x':float,'y':float,'width':float,'height':float,'obj':|this|=>GO_OBJ[rl.Rectangle]}
    var this = {
        'x': x,
        'y': y,
        'width': width,
        'height': height,
    };
    this.obj = fun() {
        return __rectangle(float(this['x']), float(this['y']), float(this['width']), float(this['height']));
    }
    return this;
}

val player_speed = 3.0;
var player_dst = Rectangle(200, 200, 100, 100);


player_dst.x -= player_speed;
println("player_dst.x = #{player_dst.x}");
assert(player_dst.x == 197.0);

val x_coord = player_dst.x-(player_dst.width/2);
println("x_coord = #{x_coord}");
assert(x_coord == 147.0, "x_coord should be 147.0, got=#{x_coord}");
