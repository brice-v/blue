# When Looping

# when caught with try-catch
# This may be the only one we have to fix


for (i in ['a','b']) {
    try {
        to_num(i);
    } catch (e) {}
}

assert(false);

# Others?