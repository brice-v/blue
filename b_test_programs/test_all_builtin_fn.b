var input = """LR

11A = (11B, XXX)
11B = (XXX, 11Z)
11Z = (11B, XXX)
22A = (22B, XXX)
22B = (22C, 22C)
22C = (22Z, 22Z)
22Z = (22B, 22B)
XXX = (XXX, XXX)""".replace("\r", "");

fun parse_input2(split_lines) {
    var this = {};
    var order = split_lines[0];
    var cur_index = 0;
    var cur = if (order[0] == 'L') {
        0
    } else {
        1
    };
    var m = {};
    var starting_keys = [];
    var ending_keys = [];
    for var i = 2; i < split_lines.len(); i += 1 {
        val line = split_lines[i];
        val key = line.split(" = ")[0];
        if key[2] == 'A' {
            starting_keys << key;
        } else if key[2] == 'Z' {
            ending_keys << key;
        }
        val values = line.split(" = ")[1].replace("[\\(\\),]","", is_regex=true).split(" ");
        m[key] = values;
    }
    this['order'] = order;
    this['cur_index'] = cur_index;
    this['m'] = m;
    this['cur'] = cur;
    this.next = fun() {
        if this.cur_index+1 == this.order.len() {
            this.cur_index = 0;
        } else {
            this.cur_index += 1;
        }
        this.cur = if (this.order[this.cur_index] == 'L') {
            0
        } else {
            1
        };
    }
    this['starting_keys'] = starting_keys;
    this['ending_keys'] = ending_keys;
    return this;
}

fun part2(_input) {
    val split_lines = _input.split("\n");
    var sum = 0;
    var game = parse_input2(split_lines);
    var next_keys = [];
    for key in game.starting_keys {
        next_keys << game.m[key][game.cur];
    }
    #println("next_keys = #{next_keys}")
    sum += 1;
    for !next_keys.all(|e| => e[2] == 'Z') {
        game.next();
        for var j = 0; j < next_keys.len(); j += 1 {
            val key = next_keys[j];
            next_keys[j] = game.m[key][game.cur];
        }
        #println("next_keys = #{next_keys}")
        sum += 1;
        if next_keys.all(|e| => e[2] == 'Z') {
            break;
        }
    }
    println("Part2 = #{sum}");
    return sum;
}

assert(part2(input) == 6);
