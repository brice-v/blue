# TODO/Note: When we match, we HAVE to return otherwise a method will get overridden?

fun send(obj, value) {
    return match obj {
        {t: "pid", v: _} => {
            _send(obj.v, value)
        },
        {t: "ws", v: _} => {
            import http
            http.ws_send(obj.v, value)
        },
        _ => {
            error("obj `#{obj}` is invalid type")
        },
    };
}

fun recv(obj) {
    return match obj {
        {t: "pid", v: _} => {
            _recv(obj.v)
        },
        {t: "ws", v: _} => {
            import http
            http.ws_recv(obj.v)
        },
        _ => {
            error("obj `#{obj}` is invalid type")
        },
    };
}

fun read(obj, as_bytes=false) {
    return match obj {
        {t: _, v: _} => {
            if ("net" in obj.t) {
                import net
                net.net_read(obj.v, obj.t)
            }
        },
        _ => {
            _read(obj, as_bytes)
        },
    };
}

fun write(obj, value) {
    return match obj {
        {t: _, v: _} => {
            if ("net" in obj.t) {
                import net
                net.net_write(obj.v, obj.t, value)
            }
        },
        _ => {
            _write(obj, value)
        },
    };
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
        _ => {
            error("`find_all` unsupported method '#{method}'")
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
        _ => {
            error("`find_one` unsupported method '#{method}'")
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

fun ping(obj) {
    return match obj {
        {t: "db", v: _} => {
            import db
            db.db_ping(obj.v)
        },
        _ => {
            error("obj `#{obj}` is invalid type")
        },
    };
}

fun close(obj) {
    return match obj {
        {t: "db", v: _} => {
            import db
            db.db_close(obj.v)
        },
        {t: "net/tcp", v: _} => {
            import net
            net.listener_close(obj.v)
        },
        {t: "net", v: _} => {
            import net
            net.conn_close(obj.v)
        },
        _ => {
            error("obj `#{obj}` is invalid type")
        },
    };
}

fun execute(db_obj, exec_query, exec_args=[]) {
    return match db_obj {
        {t: "db", v: _} => {
            import db
            db.db_exec(db_obj.v, exec_query, exec_args)
        },
        _ => {
            error("db_obj `#{db_obj}` is invalid type")
        },
    };
}

fun query(db_obj, query_s, query_args=[], named_cols=false) {
    return match db_obj {
        {t: "db", v: _} => {
            import db
            db.db_query(db_obj.v, query_s, query_args, named_cols)
        },
        _ => {
            error("db_obj `#{db_obj}` is invalid type")
        },
    };
}

fun accept(obj) {
    return match obj {
        {t: "net/tcp", v: _} => {
            import net
            net.net_accept(obj.v)
        },
        _ => {
            error("obj `#{obj}` is invalid type")
        },
    };
}
