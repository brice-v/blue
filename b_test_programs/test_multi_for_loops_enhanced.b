import time

# testing break
var i = 0;
for (true) {
    println("inside loop, i=#{i}");
    i += 1;
    if (i == 10) {
        println("before break");
        break;
        println("UNREACHABLE!!!");
        time.sleep(1000);
    }
}

println("after break, i=#{i} (should be 10)");

if (i != 10) {
    assert(false);
} else {
    assert(true);
}

# Testing continue

i = 0;
for (true) {
    println("inside loop, i=#{i}");
    i += 1;
    if (i == 1) {
        println("continue...");
        continue;
    } else {
        break;
    }
}

println("after break, i=#{i} (should be 2)");

if (i != 2) {
    assert(false);
} else {
    assert(true);
}

# The above all works so thats good


# Testing multi for loops

i = 0;


# Tests pass with this weird setup but we still cant get it to the print after the for loop
for (true) {
    println("before loop, i=#{i}");
    for (x in 1..10) {
        if (i > 30) {
            println("break inside for loop (inside for loop)");
            #time.sleep(1000);
            break;
            error("UNREACHABLE");
        }
        i += 1;
        println("inside loop, i=#{i}");
    }
    println("after loop, i=#{i}");
    if (i > 30) {
        i += 100;
        println("i > 30, breaking");
        break;
        error("UNREACHABLE");
    }
}
println("LAST LINE HERE #{i}");

if (i != 131) {
    assert(false);
} else {
    assert(true);
}

assert(true);