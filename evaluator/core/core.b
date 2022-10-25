fun read(fname, as_bytes=false) {
    _read(fname, as_bytes)
}

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

fun find_all(str_to_search, query, method="regex") {
    import search
    return match method {
        "regex" => {
            search.by_regex(str_to_search, query, false)
        },
        "xpath" => {
            search.by_xpath(str_to_search, query, false)
        },
    };
}

fun find_one(str_to_search, query, method="regex") {
    import search
    return match method {
        "regex" => {
            search.by_regex(str_to_search, query, true)
        },
        "xpath" => {
            search.by_xpath(str_to_search, query, true)
        },
    };
}

fun json_to_map(json_str) {
    try {
        return eval(json_str);
    } catch (e) {
        return error("json_to_map error: invalid json_str #{json_str}, e=#{e}");
    }
}

# DB Object Methods
# TODO: If these methods need to be used for more than one obj we can expand the match scenarios
fun ping(db_obj) {
    import db
    return match db_obj {
        {t: "DB", v: _} => {
            db.ping_(db_obj.v)
        },
        _ => {
            error("db_obj `#{db_obj}` is invalid type")
        },
    };
}

fun close(db_obj) {
    import db
    return match db_obj {
        {t: "DB", v: _} => {
            db.close_(db_obj.v)
        },
        _ => {
            error("db_obj `#{db_obj}` is invalid type")
        },
    };
}

fun execute(db_obj, exec_query, exec_args=[]) {
    import db
    return match db_obj {
        {t: "DB", v: _} => {
            db.exec_(db_obj.v, exec_query, exec_args)
        },
        _ => {
            error("db_obj `#{db_obj}` is invalid type")
        },
    };
}

fun query(db_obj, query_s, query_args=[], named_cols=false) {
    import db
    return match db_obj {
        {t: "DB", v: _} => {
            db.query_(db_obj.v, query_s, query_args, named_cols)
        },
        _ => {
            error("db_obj `#{db_obj}` is invalid type")
        },
    };
}