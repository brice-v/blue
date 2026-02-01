var cond = 5 in 1..10;
println(cond)
if (not cond) {
    assert(false);
}

var cond_ = 5 in 1 ..< 5;
assert(cond_ == false);