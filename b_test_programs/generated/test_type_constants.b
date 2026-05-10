# Test Type constants

val tInt = Type.INT
val tString = Type.STRING
val tList = Type.LIST
val tMap = Type.MAP
val tSet = Type.SET
val tBool = Type.BOOL
val tNull = Type.NULL
val tFunction = Type.FUN
val tClosure = Type.CLOSURE
val tBuiltin = Type.BUILTIN
val tRegex = Type.REGEX
val tBytes = Type.BYTES
val tBigint = Type.BIGINT
val tBigfloat = Type.BIGFLOAT
val tUinteger = Type.UINT
val tProcess = Type.PROCESS
val tError = Type.ERR

# Verify types match expected strings
assert(tInt == "INTEGER")
assert(tString == "STRING")
assert(tList == "LIST")
assert(tMap == "MAP")
assert(tSet == "SET")
assert(tBool == "BOOLEAN")
assert(tNull == "NULL")
assert(tFunction == "FUNCTION")
assert(tBuiltin == "BUILTIN")
assert(tRegex == "REGEX")
assert(tBytes == "BYTES")
assert(tBigint == "BIG_INTEGER")
assert(tBigfloat == "BIG_FLOAT")
assert(tUinteger == "UINTEGER")
assert(tProcess == "PROCESS")
assert(tError == "ERROR")

# Verify actual type() calls match constants
assert(type(42) == Type.INT)
assert(type("hello") == Type.STRING)
assert(type([1, 2, 3]) == Type.LIST)
assert(type({a: 1}) == Type.MAP)
assert(type({1, 2, 3}) == Type.SET)
assert(type(true) == Type.BOOL)
assert(type(false) == Type.BOOL)
assert(type(null) == Type.NULL)
assert(type(1234n) == Type.BIGINT)
assert(type(3.14n) == Type.BIGFLOAT)
assert(type(0x1F) == Type.UINT)
assert(type(r/abc/) == Type.REGEX)

# Function type (could be FUNCTION or CLOSURE depending on context)
val fn = fun() { 1 }
assert(type(fn) == Type.FUN || type(fn) == Type.CLOSURE)

# Closure type
fun makeClosure() {
    var x = 1
    return fun() { x }
}
val closure = makeClosure()
assert(type(closure) == Type.CLOSURE)

# Builtin function type
assert(type(len) == Type.BUILTIN)
assert(type(println) == Type.BUILTIN)
assert(type(print) == Type.BUILTIN)

# Error type
try {
    error("test error")
} catch (e) {
    assert(type(e) == Type.ERROR)
}

# Process type
val myPid = self()
assert(type(myPid) == Type.PROCESS)

# Bytes type
val b = "hello".to_bytes()
assert(type(b) == Type.BYTES)

# Float type
assert(type(3.14) == "FLOAT")

# Integer type
assert(type(0) == Type.INT)
assert(type(-1) == Type.INT)

# Test that Type constants are consistent
assert(Type.INT == "INTEGER")
assert(Type.STRING == "STRING")
assert(Type.LIST == "LIST")
assert(Type.MAP == "MAP")
assert(Type.SET == "SET")

# Test with list comprehension result
val compList = [x for x in 1..5]
assert(type(compList) == Type.LIST)

# Test with set comprehension result
val compSet = {x for x in 1..5}
assert(type(compSet) == Type.SET)

# Test with map comprehension result
val compMap = {x: x*2 for x in 1..5}
assert(type(compMap) == Type.MAP)

# Test with if expression result
val ifResult = if (true) { 1 } else { 2 }
assert(type(ifResult) == Type.INT)

# Test with match expression result
val matchResult = match (5) {
    5 => { "five" },
    _ => { "other" },
}
assert(type(matchResult) == Type.STRING)

# Test with range
val range = 1..5
assert(type(range) == "LIST")

# Test with raw string
val raw = """hello"""
assert(type(raw) == Type.STRING)

# Test with backtick string
val backtick = `echo test`
assert(type(backtick) == "STRING")
