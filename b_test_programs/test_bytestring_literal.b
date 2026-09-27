# Test the b"..." and b'...' byte string literals.

# Both quote styles build the same bytes, and they are not strings.
val dq = b"hello";
val sq = b'hello';
assert(dq == sq);
assert(len(dq) == 5);
assert(dq != "hello");
println("Both quote styles: PASS");

# Empty literals are legal and hold no bytes.
assert(len(b"") == 0);
assert(b"" == b'');
println("Empty literals: PASS");

# Escape sequences are decoded, and a quote of the other kind is plain data.
assert(b"a\nb".len() == 3);
assert(b"tab\there".len() == 8);
assert(b"\x10AAA".len() == 4);
assert(b"\x10AAA" != b"AAA");
assert(b"quote\"inside".len() == 12);
assert(b'quote"inside'.len() == 12);
assert(b'\'' == b"'");
println("Escapes: PASS");

# A byte string is a value, so it passes through calls and collections.
fun width(x) {
    return len(x);
}
assert(width(b"abcd") == 4);
assert([b"a", b"b"].len() == 2);

val conf = {"key": b"value"};
assert(conf["key"] == b"value");

val pair = [b"left", b'right'];
assert(pair[0] == b"left");
assert(pair[1] == b'right');
println("Values: PASS");

# A b is only a prefix when a quote follows it, so names keep working.
val b = 7;
val b1 = 8;
val bb = 9;
assert(b + b1 + bb == 24);
assert(len(b"AB") == 2);
assert(b == 7);
println("Bare b is a name: PASS");
