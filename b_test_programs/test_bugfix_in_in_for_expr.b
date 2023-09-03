val x = ['1','2','3','4','5','6'];
val y = ['6','5','4','3','2','1'];

for (a in x) {
    println("a = #{a}, type a = #{type(a)}");
    println("y = #{y}")
    assert(a in y);
}

assert(true);