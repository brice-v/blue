val get = _get;
val post_ = _post;
val serve_ = _serve;
val static_ = _static;
val handle_ = _handle;
# _server is an id that corresponds to the gofiber server object
val _server = _new_server();

fun post(url, body, mime_type="application/json") {
    post_(url, mime_type, body)
}

fun serve(addr="localhost:3001") {
    serve_(_server, addr)
}

fun handle(pattern, fn, method="GET") {
    handle_(_server, pattern, fn, method)
}

fun static(prefix="/", dir_path=".", browse=false) {
    static_(_server, prefix, dir_path, browse)
}