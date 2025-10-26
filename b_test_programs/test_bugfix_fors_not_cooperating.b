var l1 = [1,2,3, 'a'];
var l2 = [{}, null, ''];
var l3 = [6,7,8,8,9,'c'];
var lol = [l1,l2,l3];
for (l in lol) {
    println("l = #{l}");
    for (var j = 0; j < 3; j += 1) {
        println("j = #{j}, minL = #{3}");
        println("l[j] = #{l[j]}")
    }
    try {
        l[0];
    } catch (e) {
        assert(e != "identifier not found: l");
    }
}
assert(true);