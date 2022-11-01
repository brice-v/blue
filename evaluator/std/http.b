val get = _get;
val __post = _post;
val __serve = _serve;
val __static = _static;
val __handle = _handle;
val __handle_ws = _handle_ws;
val ws_send = _ws_send;
val ws_recv = _ws_recv;
# _server is an id that corresponds to the gofiber server object
val _server = _new_server();

fun post(url, body, mime_type="application/json") {
    __post(url, mime_type, body)
}

fun serve(addr_port="localhost:3001") {
    __serve(_server, addr_port)
}

fun handle(pattern, fn, method="GET") {
    __handle(_server, pattern, fn, method)
}

fun handle_ws(pattern, fn) {
    __handle_ws(_server, pattern, fn)
}

fun static(prefix="/", dir_path=".", browse=false) {
    __static(_server, prefix, dir_path, browse)
}