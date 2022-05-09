var cond = 5 in 1..10;
println(cond)
if (not cond) {
    false
}

var cond_ = 5 in 1 ..< 5;
if (cond_ != false) {
    false
}

true