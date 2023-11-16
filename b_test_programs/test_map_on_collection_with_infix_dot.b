var result = "1,2,3,4,5".split(",").map(int).filter(|x| => x > 3);
assert(result == [4,5]);