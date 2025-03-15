fun abc() {
    var this = {};
    this.x = 1;
    this.next = fun() {
        this.x += 1;
    }
    return this;
}

var y = abc();
y.next();
assert(y.x == 2);