val fun1 = |x, y| => {
    x + y
};

assert(fun1(1,2) == 3);

val fun2 = |x| => x + 1;

assert(fun2(1) == 2);