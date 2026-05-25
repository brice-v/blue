var MyObj = {
    x: 2,
    y: 3,
    test_abc: |x| => { x + 3 },
};

assert(MyObj.test_abc(MyObj.x) == 5);