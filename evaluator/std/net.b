val __connect = _connect;
val __listen = _listen;
val net_accept = _accept;
val listener_close = _listener_close;
val conn_close = _conn_close;
val tcp_read = _tcp_read;
val tcp_write = _tcp_write;

fun connect(transport="tcp", addr="localhost", port="18650") {
    __connect(transport, addr, port)
}

fun listen(transport="tcp", addr="localhost", port="18650") {
    __listen(transport, addr, port)
}

# listen just gives you a listener that you can accept new connections with (for tcp)
# for UDP we need to use a different listener which allows us to read directly from it

# connect should work the same for either, but we can read/write from tcp conn, not udp?