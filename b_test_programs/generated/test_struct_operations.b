# Test struct literal operations

# Basic struct creation
val point = @{x: 10, y: 20}
assert(point.x == 10)
assert(point.y == 20)

# Struct with string fields
val person = @{name: "Alice", age: 30, city: "NYC"}
assert(person.name == "Alice")
assert(person.age == 30)
assert(person.city == "NYC")

# Struct with mixed types
val mixed = @{int: 1, float: 2.5, str: "hello", bool: true, list: [1, 2], map: {a: 1}}
assert(mixed.int == 1)
assert(mixed.float == 2.5)
assert(mixed.str == "hello")
assert(mixed.bool == true)
assert(mixed.list == [1, 2])
assert(mixed.map == {a: 1})

# Struct field assignment
var mutableStruct = @{x: 1, y: 2}
mutableStruct.x = 10
mutableStruct.y = 20
assert(mutableStruct.x == 10)
assert(mutableStruct.y == 20)

# Struct in function
fun createPoint(x, y) {
    return @{x: x, y: y}
}

val p = createPoint(5, 10)
assert(p.x == 5)
assert(p.y == 10)

# Struct returned from function
fun createPerson(name, age) {
    return @{name: name, age: age}
}

val alice = createPerson("Alice", 30)
val bob = createPerson("Bob", 25)
assert(alice.name == "Alice")
assert(bob.name == "Bob")

# Struct as function argument
fun describe(obj) {
    return "name=#{obj.name}, age=#{obj.age}"
}

val result = describe(alice)
assert(result == "name=Alice, age=30")

# Struct in list
val people = [
    @{name: "Alice", age: 30},
    @{name: "Bob", age: 25},
    @{name: "Charlie", age: 35},
]
assert(len(people) == 3)
assert(people[0].name == "Alice")
assert(people[1].age == 25)

# Struct in map
val data = {
    user1: @{name: "Alice", age: 30},
    user2: @{name: "Bob", age: 25},
}
assert(data["user1"].name == "Alice")
assert(data["user2"].age == 25)

# Struct with nested struct
val nested = @{
    outer: @{
        inner: @{value: 42}
    }
}
assert(nested.outer.inner.value == 42)

# Struct with list of structs
val employees = [
    @{name: "Alice", dept: "Engineering", salary: 100000},
    @{name: "Bob", dept: "Marketing", salary: 80000},
    @{name: "Charlie", dept: "Engineering", salary: 95000},
]

# Filter employees by department
val engineers = [emp for emp in employees if emp.dept == "Engineering"]
assert(len(engineers) == 2)
assert(engineers[0].name == "Alice")
assert(engineers[1].name == "Charlie")

# Struct with null field
val nullable = @{name: "test", optional: null}
assert(nullable.name == "test")
assert(nullable.optional == null)

# Struct equality
val s1 = @{a: 1, b: 2}
val s2 = @{a: 1, b: 2}
val s3 = @{a: 1, b: 3}
assert(s1 == s2)
assert(s1 != s3)

# Struct with boolean fields
val flags = @{enabled: true, disabled: false}
assert(flags.enabled == true)
assert(flags.disabled == false)

# Struct with empty list/map
val emptyFields = @{items: [], data: {}}
assert(emptyFields.items == [])
assert(emptyFields.data == {})

# Struct field access with variable
val fieldName = "name"
val obj = @{name: "test"}
# Note: obj[fieldName] is map-style access, not struct-style
assert(obj["name"] == "test")

# Struct with number field names
val numbered = @{1: "one", 2: "two"}
assert(numbered[1] == "one")
assert(numbered[2] == "two")

# Struct in comprehension
val squares = [@{x: i, y: i * i} for i in 1..5]
assert(len(squares) == 5)
assert(squares[0].x == 1)
assert(squares[0].y == 1)
assert(squares[4].x == 5)
assert(squares[4].y == 25)

# Struct modification
var modifiable = @{x: 1, y: 2}
modifiable.x = 100
modifiable.y = 200
assert(modifiable.x == 100)
assert(modifiable.y == 200)

# Struct with arithmetic
val p1 = @{x: 10, y: 20}
val p2 = @{x: 5, y: 15}
val sum = @{x: p1.x + p2.x, y: p1.y + p2.y}
assert(sum.x == 15)
assert(sum.y == 35)

# Struct in match
val item = @{type: "point", x: 1, y: 2}
val result = match (item) {
    {type: "point", x: _, y: _} => { "point at #{item.x},#{item.y}" },
    _ => { "unknown" }
}
assert(result == "point at 1,2")

# Struct with default values
fun createConfig(host = "localhost", port = 8080) {
    return @{host: host, port: port}
}

val config1 = createConfig()
assert(config1.host == "localhost")
assert(config1.port == 8080)

val config2 = createConfig("example.com", 3000)
assert(config2.host == "example.com")
assert(config2.port == 3000)

# Struct immutability when declared with val
try {
    val immutableStruct = @{x: 1}
    immutableStruct.x = 2
    assert(false, "should have errored")
} catch (e) {
    assert(e.contains("immutable") || e.contains("already defined"))
}

assert(true)
