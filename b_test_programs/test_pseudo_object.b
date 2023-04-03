fun Person(name, age) {
    var this = {
        'name': name,
        'age': age,
    };

    this.to_s = fun() {
        "Person{name='#{this.name}', age=#{this.age}}"
    }

    this.inc_age = fun() {
        this.age += 1;
    }

    return this;
}

var me = Person("Person1", 20);
println(me.to_s());
assert(me.to_s() == "Person{name='Person1', age=20}");
me.inc_age()
println(me.to_s());
assert(me.to_s() == "Person{name='Person1', age=21}");