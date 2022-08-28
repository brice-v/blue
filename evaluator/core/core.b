fun map(list, f) {
    var __internal__ = [];
    for (e in list) {
        __internal__ = __internal__.append(f(e));
    }
    return __internal__;
}

fun filter(list, f) {
    var __internal__ = [];
    for (e in list) {
        if (f(e)) {
            __internal__ = __internal__.append(e);
        }
    }
    return __internal__;
}

fun reduce(list, f, acc=null) {
    ###
    if (acc == null) {
        if (list.len() == 0) {
            return [];
        }
        acc = list[0];
    }
    ###
    #println("acc=#{acc} before loop");
    for (e in list) {
        #println("e=#{e}, acc=#{acc}");
        acc = f(acc,e)
    }
    return acc;
}