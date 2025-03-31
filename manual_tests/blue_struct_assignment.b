import time
val start = time.now();
var x = @{abc: 0}

for var j = 0; j < 10_000_000; j += 1 {
    x.abc += j;
}

println(x);
assert(x.abc == 49999995000000);
println("blue_struct_assignment took: #{time.now() - start}ms");