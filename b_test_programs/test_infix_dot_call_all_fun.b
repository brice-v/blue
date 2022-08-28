fun hello() {
    "HELLO"
}

if (hello() != "HELLO") {
    return false;
}

fun hello_a(a) {
    "HELLO #{a}"
}

if ("A".hello_a() != "HELLO A") {
    return false;
}

fun hello_b(b, c) {
    "HELLO #{b} #{c}"
}

if ("3".hello_b("C") != "HELLO 3 C") {
    return false;
}

try {
    (1 + 2).hello_b("3")
    return false;
} catch (e) {}

return true;