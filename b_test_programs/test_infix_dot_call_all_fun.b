fun hello() {
    "HELLO"
}

assert(hello() == "HELLO");

fun hello_a(a) {
    "HELLO #{a}"
}

assert("A".hello_a() == "HELLO A")

fun hello_b(b, c) {
    "HELLO #{b} #{c}"
}

assert("3".hello_b("C") == "HELLO 3 C")

try {
    assert((1 + 2).hello_b("3") == "HELLO 3 3")
} catch (e) {
    assert(false, "UNREACHABLE");
}
