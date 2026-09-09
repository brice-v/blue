# Forward declarations: functions and variables can be used before they are
# defined, as long as the use happens inside a function that runs afterwards.

fun main() {
    assert(greet() == "hello")
    assert(twice(21) == 42)
    assert(scaled(5) == 15)
    assert(describe() == "value is 9")
    assert(is_even(10))
    true;
}

# defined below main but called by it
fun greet() { "hello" }
fun twice(n) { n * 2 }

# read by scaled() and describe() before their declaration shows up
fun scaled(n) { n * scale }
var scale = 3

fun describe() { "value is #{value}" }
var value = 9

# mutual recursion through helpers that are defined after the functions using them
fun is_even(n) {
    if (n == 0) { return true }
    odd_helper(n - 1)
}
fun odd_helper(n) { is_odd(n) }
fun is_odd(n) {
    if (n == 0) { return false }
    even_helper(n - 1)
}
fun even_helper(n) { is_even(n) }

main();

# they all keep working with whatever the name holds at call time
assert(greet() == "hello")
assert(twice(3) == 6)
assert(scaled(2) == 6)
assert(describe() == "value is 9")
scale = 10
assert(scaled(2) == 20)
value = 40
assert(describe() == "value is 40")
assert(is_even(8))
assert(!is_even(7))
assert(is_odd(7))
assert(!is_odd(4));
