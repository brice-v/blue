var x = true;
var b = {name: "Dont matter", age: 1094_0024};
var c = {age: 140, name: "Brice", another: [1,2,3,4,5]};
var mma = "Mixed Martial Arts";

var myList = [x,b,c, mma];

fun matchItem(item) {
    return match item {
        {name: _, age: _ } => { "we got a name=#{item.name} and an age=#{item.age}" },
        x => { "This would also match because a map/obj is truthy: #{x}" },
        {name: "Brice", age:_, another: [1,2,3,4,5]} => { 
            "Could put some complex logic in here like this: The object is obj{name: #{item.name}, age: #{item.age}, another: #{item.another}}"
        },
        _ => { null },
    };
}

for (item in myList) {
    println("This item = #{item} Should match = #{matchItem(item)}");
}

var HelloWorld = "Hello World!";
var resultToCompare = match HelloWorld {
    "Hello" => { "This would return hello" },
    _ => { null },
};

if (resultToCompare != null) {
    return false;
}

return true;