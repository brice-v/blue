#val get = _get;
val __fetch = _fetch;
#val __post = _post;
val __serve = _serve;
val __static = _static;
val __handle = _handle;
val __handle_ws = _handle_ws;
val ws_send = _ws_send;
val ws_recv = _ws_recv;
# _server is an id that corresponds to the gofiber server object
val _server = _new_server();

# Default headers seem to be host, user-agent, accept-encoding (not case sensitive for these check pictures)
# deno also used accept: */* (not sure what that is)
fun fetch(resource, options=null) {
    if (options == null) {
        options = {
            method: 'GET',
            headers: {},
            body: null,
        };
    } else {
        val t = options.type();
        if (t != 'MAP') {
            return error("`fetch` error:  options must be MAP. got=#{t}");
        }
        if (options.method == null) {
            options.method = 'GET';
        }
        if (options.headers == null) {
            options.headers = {};
        } else {
            # TODO: Need to support options.headers.type();
            val ht = type(options.headers);
            if (ht != 'MAP') {
                return error("`fetch` error:  options.headers must be MAP. got=#{ht}");
            }
        }
        if (options.body == null and options.method != 'GET') {
            return error("`fetch` options.body must not be null when method is not 'GET'");
        }
    }
    __fetch(resource, options.method, options.headers, options.body)
}

fun get(url) {
    fetch(url)
}

fun post(url, post_body, mime_type="application/json") {
    val options = {
        method: 'POST',
        body: post_body,
        headers: {
            'content-type': mime_type
        }
    };
    fetch(url, options)
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