val __connect = _connect;
val __listen = _listen;
val __inspect = _inspect;
val net_accept = _accept;
val net_close = _net_close;
val net_read = _net_read;
val net_write = _net_write;

fun connect(transport="tcp", addr="localhost", port="18650") {
    __connect(transport, addr, port)
}

fun listen(transport="tcp", addr="localhost", port="18650") {
    __listen(transport, addr, port)
}

fun inspect(obj) {
    return match obj {
        {t: _, v: _} => {
            __inspect(obj.v, obj.t)
        },
        _ => {
            error("`inspect` expects object")
        },
    };
}