val x = ['1'];

for (i in x) {
    println("i = `#{i}`");
}

var z = {'a': 123};

for (i in z) {
    println("i = `#{i}`");
}

for ([j, k] in z) {
    println("j = `#{j}`, k = `#{k}`");
}


if ({}) {
    error("{} should not be truthy");
}

if ([]) {
    error("[] should not be truthy");
}

val abc123 = [yyy for (yyy in x) if (yyy != '1')];
println(abc123);

if ([yyy for (yyy in x) if (yyy != '1')]) {
    error("abc123 should not be truthy");
}

assert(true);