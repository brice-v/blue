val input = """1000
2000
3000

4000

5000
6000

7000
8000
9000

10000

""";

val lines = input.split("\n");

for (line in lines) {
    if (line == '') {
        continue;
    }
    println(line);
}

assert(true);