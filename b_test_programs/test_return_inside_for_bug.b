import math

val input = """[[1],[2,3,4]]
[[1],4]""";
var lines = input.split("\n");

fun get_pairs_of_lists(lines) {
    var pairs = [];
    var index = 1;
    var i = 0;
    for (true) {
        if (i > len(lines) or (i+1) > len(lines)) {
            break;
        }
        if (lines[i] == '') {
            i += 1;
        } else {
            var pair1 = eval(lines[i]);
            var pair2 = eval(lines[i+1]);
            pairs << {'pair1': pair1, 'pair2': pair2, 'index': index};
            index += 1;
            i += 2;
        }
    }
    return pairs;
}

fun are_lists_in_order(leftl, rightl) {
    var maxL = math.max(len(leftl), len(rightl));
    var result = null;
    for var i = 0; i < maxL; i += 1 {
        var lval = leftl[i];
        var rval = rightl[i];
        println("lval = #{lval} (#{lval.type()}), rval = #{rval} (#{rval.type()})");
        if lval.type() == Type.INT && rval.type() == Type.INT {
            if lval < rval {
                result = true;
                break;
            } else if rval > lval {
                result = false;
                break;
            } else {
                continue;
            }
        } else if lval.type() == Type.INT && rval.type() == Type.LIST {
            # convert left to list then compare that to current list
            var newlval = [lval];
            result = are_lists_in_order(newlval, rval);
            break;
        } else if lval.type() == Type.LIST && rval.type() == Type.INT {
            # convert right to list then compare that to current list
            var newrval = [rval];
            result = are_lists_in_order(lval, newrval);
            break;
        } else if lval.type() == Type.LIST && rval.type() == Type.LIST {
            result = are_lists_in_order(lval, rval);
            break;
        }
    }
    return result;
}

fun are_pairs_in_order(left_pair, right_pair) {     
    val lpairLen = len(left_pair);
    val rpairLen = len(right_pair);
    var maxL = math.max(lpairLen, rpairLen);
    for var i = 0; i < maxL; i += 1 {
        var result = null;
        if i == 0 {
            result = are_lists_in_order(left_pair, right_pair);
        } else {
            result = are_lists_in_order([left_pair[i]], [right_pair[i]]);
            println("IN HERE result = #{result}")
        }
        println("result = #{result}");
        if result == null {
            continue;
        }
        println("should be returning result = #{result}")
        return result;
    }
}

fun part1(lines) {
    val pairs_of_lists = get_pairs_of_lists(lines);
    var indexes = [];
    for (pair in pairs_of_lists) {
        println("pair = #{pair}")
        var pairs_in_order = are_pairs_in_order(pair.pair1, pair.pair2);
        println("pairs_in_order = #{pairs_in_order}");
        assert(pairs_in_order);
    }
}


part1(lines);
assert(true);