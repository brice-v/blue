var MyObj = {
    x: 2,
    y: 3,
    test: |x| => { x + 3 },
};

if (MyObj.test(MyObj.x) == 5) {
    return true;
} else {
    return false;
}