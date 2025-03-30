import time
val start = time.now();
var x = @{abc: 0}

for var j = 0; j < 10_000_000; j += 1 {
    x.abc += j;
}

println(x);
println(time.now() - start);