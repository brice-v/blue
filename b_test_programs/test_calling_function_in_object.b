var x = {
    somevar: 0,
    inc: fun() {
        x.somevar += 1;
    },
}

var y = {
    abc: 1,
    inc: fun() {
        y.abc += 4;
    }
}


x.inc()
println(x.somevar)

y.inc()
println(y.abc)
return true;