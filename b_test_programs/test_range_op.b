var cond = 5 in 1..10;
println(cond)
if (not cond) {
    return false;
}

var cond_ = 5 in 1 ..< 5;
if (cond_ != false) {
    return false;
}

return true;