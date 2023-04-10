val Type = {
    BOOL: 'BOOLEAN',
    INT: 'INTEGER',
    UINT: 'UINTEGER',
    FLOAT: 'FLOAT',
    BIGINT: 'BIG_INTEGER',
    BIGFLOAT: 'BIG_FLOAT',
    STRING: 'STRING',
    SET: 'SET',
    MAP: 'MAP',
    LIST: 'LIST',
    FUN: 'FUNCTION',
    BUILTIN: 'BUILTIN',
    MODULE: 'MODULE_OBJ',
};

fun send(obj, value) {
    ##core:ignore
    match obj {
        {t: "pid", v: _} => {
            _send(obj.v, value)
        },
        {t: "ws", v: _} => {
            import http
            http.ws_send(obj.v, value)
        },
        {t: "ws/client", v: _} => {
            import http
            http.ws_client_send(obj.v, value)
        },
        _ => {
            error("obj `#{obj}` is invalid type. got=`#{obj}` (#{type(obj)})")
        },
    }
}

fun recv(obj) {
    ##core:ignore
    match obj {
        {t: "pid", v: _} => {
            _recv(obj.v)
        },
        {t: "ws", v: _} => {
            import http
            http.ws_recv(obj.v)
        },
        {t: "ws/client", v: _} => {
            import http
            http.ws_client_recv(obj.v)
        },
        {t: "sub", v: _} => {
            _pubsub_sub_listen(obj.v)
        },
        _ => {
            error("obj `#{obj}` is invalid type. got=`#{obj}` (#{type(obj)})")
        },
    }
}

fun read(obj, as_bytes=false) {
    ##core:ignore
    match obj {
        {t: _, v: _} => {
            if ("net" in obj.t) {
                import net
                net.net_read(obj.v, obj.t)
            }
        },
        _ => {
            _read(obj, as_bytes)
        },
    }
}

fun write(obj, value) {
    ##core:ignore
    match obj {
        {t: _, v: _} => {
            if ("net" in obj.t) {
                import net
                net.net_write(obj.v, obj.t, value)
            }
        },
        _ => {
            _write(obj, value)
        },
    }
}

fun map(list, f) {
    ##core:ignore
    var __internal__ = [];
    for (e in list) {
        __internal__ << f(e);
    }
    return __internal__;
}

fun filter(list, f) {
    ##core:ignore
    var __internal__ = [];
    for (e in list) {
        if (f(e)) {
            __internal__ << e;
        }
    }
    return __internal__;
}

fun reduce(list, f, acc=null) {
    ##core:ignore
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
    ##core:ignore
    import search
    match method {
        "regex" => {
            search.by_regex(str_to_search, query, false)
        },
        "xpath" => {
            search.by_xpath(str_to_search, query, false)
        },
        _ => {
            error("`find_all` unsupported method '#{method}'")
        },
    }
}

fun find_one(str_to_search, query, method="regex") {
    ##core:ignore
    import search
    match method {
        "regex" => {
            search.by_regex(str_to_search, query, true)
        },
        "xpath" => {
            search.by_xpath(str_to_search, query, true)
        },
        _ => {
            error("`find_one` unsupported method '#{method}'")
        },
    }
}

fun from_json(json_str) {
    ##core:ignore
    if (not is_valid_json(json_str)) {
        return error("from_json error: invalid json_str #{json_str}, e=#{e}");
    }
    try {
        return eval(json_str);
    } catch (e) {
        return error("from_json error: invalid json_str #{json_str}, e=#{e}");
    }
}

fun ping(obj) {
    ##core:ignore
    match obj {
        {t: "db", v: _} => {
            import db
            db.db_ping(obj.v)
        },
        _ => {
            error("obj `#{obj}` is invalid type. got=`#{obj}` (#{type(obj)})")
        },
    }
}

fun close(obj) {
    ##core:ignore
    match obj {
        {t: "db", v: _} => {
            import db
            db.db_close(obj.v)
        },
        {t: _, v: _} => {
            if ("net" in obj.t) {
                import net
                net.net_close(obj.v, obj.t)
            }
        },
        _ => {
            error("obj `#{obj}` is invalid type. got=`#{obj}` (#{type(obj)})")
        },
    }
}

fun execute(db_obj, exec_query, exec_args=[]) {
    ##core:ignore
    match db_obj {
        {t: "db", v: _} => {
            import db
            db.db_exec(db_obj.v, exec_query, exec_args)
        },
        _ => {
            error("db_obj `#{db_obj}` is invalid type")
        },
    }
}

fun query(db_obj, query_s, query_args=[], named_cols=false) {
    ##core:ignore
    match db_obj {
        {t: "db", v: _} => {
            import db
            db.db_query(db_obj.v, query_s, query_args, named_cols)
        },
        _ => {
            error("db_obj `#{db_obj}` is invalid type")
        },
    }
}

fun accept(obj) {
    ##core:ignore
    match obj {
        {t: "net/tcp", v: _} => {
            import net
            net.net_accept(obj.v)
        },
        _ => {
            error("obj `#{obj}` is invalid type. got=`#{obj}` (#{type(obj)})")
        },
    }
}

fun get_text(obj) {
    ##core:ignore
    match obj {
        {t: "ui/entry", v: _} => {
            import ui
            ui.entry_get_text(obj.v)
        },
        _ => {
            error("obj `#{obj}` is invalid type. got=`#{obj}` (#{type(obj)})")
        },
    }
}

fun substr(s, start=0, end=-1) {
    ##core:ignore
    _substr(s, start, end)
}

fun go_metrics(flat=false) {
    ##core:ignore
    # flat implies that we just want each metric path as its own key
    var __metrics = _go_metrics();
    var __metrics_split_nl = __metrics.split("\n");
    __metrics_split_nl = [x for (x in __metrics_split_nl) if (x != '')];

    var __total_metrics = {};
    if (flat) {
        for (metric in __metrics_split_nl) {
            val metric_path_and_value = metric.split(": ");
            val path = metric_path_and_value[0];
            val metric_value = to_num(metric_path_and_value[1]);
            __total_metrics[path] = metric_value;
        }
        return __total_metrics;
    }

    val __set_value_from_list_of_keys_in_map = fun(m, l, value) {
        # m is our map
        # l is our list of string keys (the last key there is a key to a value)
        # value is what were trying to set
        
        # this is our starting point
        var current_map = m;
        var i = 0;
        for (true) {
            var starting_key = l[i];
            if (starting_key in current_map) {
                current_map = current_map[starting_key];
            } else {
                if (i == len(l) - 1) {
                    current_map[starting_key] = value;
                    break;
                }
                current_map[starting_key] = {};
                current_map = current_map[starting_key];
            }
            i += 1;
        }
    }

    for (metric in __metrics_split_nl) {
        val metric_path_and_value = metric.split(": ");
        val path = metric_path_and_value[0];
        val metric_value = to_num(metric_path_and_value[1]);
        var metric_path_list = path.split("/");
        metric_path_list = [x for (x in metric_path_list) if (x != '')];
        __set_value_from_list_of_keys_in_map(__total_metrics, metric_path_list, metric_value);
    }

    return __total_metrics;
}

val KV = {
    put: _kv_put,
    get: _kv_get,
    delete: _kv_delete,
}

val pubsub = {
    subscribe: _subscribe,
    publish: _publish,
    broadcast: _broadcast,
    get_subscriber_count: _get_subscriber_count,
};

