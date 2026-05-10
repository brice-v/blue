# Test string methods

val s = "Hello World"

# len()
assert(s.len() == 11)
assert("".len() == 0)
assert("a".len() == 1)

# to_lower()
assert(s.to_lower() == "hello world")
assert("HELLO".to_lower() == "hello")

# to_upper()
assert(s.to_upper() == "HELLO WORLD")
assert("hello".to_upper() == "HELLO")

# to_title()
assert("hello world".to_title() == "Hello World")
assert("hELLO wORLD".to_title() == "Hello World")

# to_camel()
assert("hello_world".to_camel() == "helloWorld")
assert("hello_world_test".to_camel() == "helloWorldTest")

# to_snake()
assert("helloWorld".to_snake() == "hello_world")
assert("helloWorldTest".to_snake() == "hello_world_test")

# to_kebab()
assert("helloWorld".to_kebab() == "hello-world")
assert("helloWorldTest".to_kebab() == "hello-world-test")

# center()
assert("hi".center(6) == "  hi  ")
assert("hi".center(6, "_") == "__hi__")
assert("hello".center(3) == "hello")  # narrower than string

# ljust()
assert("hi".ljust(5) == "hi   ")
assert("hi".ljust(5, "-") == "hi---")

# rjust()
assert("hi".rjust(5) == "   hi")
assert("hi".rjust(5, "-") == "---hi")

# strip()
assert("  hello  ".strip() == "hello")
assert("\thello\t".strip() == "hello")
assert("\nhello\n".strip() == "hello")

# lstrip()
assert("  hello  ".lstrip() == "hello  ")
assert("\thello\t".lstrip() == "hello\t")

# rstrip()
assert("  hello  ".rstrip() == "  hello")
assert("\thello\t".rstrip() == "\thello")

# startswith()
assert("hello world".startswith("hello"))

# endswith()
assert("hello world".endswith("world"))

# index_of()
assert("hello world".index_of("world") == 6)
assert("hello world".index_of("o") == 4)
assert("hello world".index_of("xyz") == -1)

# replace()
assert("hello world".replace("world", "blue") == "hello blue")
assert("aaa".replace("a", "b") == "bbb")
assert("hello".replace("xyz", "abc") == "hello")

# split()
val parts = "a,b,c".split(",")
assert(parts == ["a", "b", "c"])

val words = "hello world".split(" ")
assert(words == ["hello", "world"])

val noSplit = "hello".split(",")
assert(noSplit == ["hello"])

# reverse()
assert("hello".reverse() == "olleh")
assert("abcde".reverse() == "edcba")
assert("".reverse() == "")

# to_bytes()
val bytes = "hello".to_bytes()
assert(type(bytes) == "BYTES")

# String concatenation
assert("hello" + " " + "world" == "hello world")
assert("a" + "b" + "c" == "abc")

# String repetition
assert("ab" * 3 == "ababab")
assert(3 * "x" == "xxx")
assert("hello" * 0 == "")

# String comparison
assert("abc" == "abc")
assert("abc" != "def")
### Not yet supported
assert("abc" < "abd")
assert("abc" > "abc" == false)
assert("abc" <= "abc")
assert("abc" >= "abc")
###

# Empty string operations
assert("".len() == 0)
assert("".strip() == "")
assert("".to_upper() == "")
assert("".to_lower() == "")
assert("".reverse() == "")
assert("".split(",") == [""])

# Unicode string
val unicode = "héllo"
assert(unicode.len() == 5)
assert(unicode.to_upper() == "HÉLLO")

# String interpolation test
val name = "blue"
val version = 1
val msg = "Hello #{name} v#{version}"
assert(msg == "Hello blue v1")