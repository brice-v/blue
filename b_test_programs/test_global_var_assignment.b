val input = """addx 15
addx -11
addx 6
addx -3
addx 5
addx -1
addx -8
addx 13
addx 4
noop
addx -1
addx 5
addx -1
addx 5
addx -1
addx 5
addx -1
addx 5
addx -1
addx -35
addx 1
addx 24
addx -19
addx 1
addx 16
addx -11
noop
noop
addx 21
addx -15
noop
noop
addx -3
addx 9
addx 1
addx -3
addx 8
addx 1
addx 5
noop
noop
noop
noop
noop
addx -36
noop
addx 1
addx 7
noop
noop
noop
addx 2
addx 6
noop
noop
noop
noop
noop
addx 1
noop
noop
addx 7
addx 1
noop
addx -13
addx 13
addx 7
noop
addx 1
addx -33
noop
noop
noop
addx 2
noop
noop
noop
addx 8
noop
addx -1
addx 2
addx 1
noop
addx 17
addx -9
addx 1
addx 1
addx -3
addx 11
noop
noop
addx 1
noop
addx 1
noop
noop
addx -13
addx -19
addx 1
addx 3
addx 26
addx -30
addx 12
addx -1
addx 3
addx 1
noop
noop
noop
addx -9
addx 18
addx 1
addx 2
noop
noop
addx 9
noop
noop
noop
addx -1
addx 2
addx -37
addx 1
addx 3
noop
addx 15
addx -21
addx 22
addx -6
addx 1
noop
addx 2
addx 1
noop
addx -10
noop
noop
addx 20
addx 1
addx 2
addx 2
addx -6
addx -11
noop
noop
noop""";

val lines = input.split("\n");


fun decode_and_execute(op) {
    println("(before) pc = #{pc}, op = #{op}, x = #{x}");
    match op {
        "noop" => {
            pc += 1;
        },
        _ => {
            var addx_in_op = "addx" in op;
            println("addx_in_op = #{addx_in_op}, op = `#{op}`");
            assert("addx" in op);
            var value_to_add = to_num(op.split(" ")[1]);
            x += value_to_add;
            pc += 2;
        },
    };
    println("(after) pc = #{pc}, op = #{op}, x = #{x}");
}

var pc = 0;
var x = 1;
fun part1(lines) {
    for (line in lines) {
        decode_and_execute(line);
    }
}

part1(lines);
println("pc = #{pc}, x = #{x}")

assert(pc == 0);
assert(x == 1);
# Note: This is the case because when we call the function wiht the global vars
# we can only enclose on these vars once. Which means that they will not update
# as the function runs