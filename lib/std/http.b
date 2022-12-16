val __fetch = _fetch;
val __download = _download;
val __serve = _serve;
val __static = _static;
val __handle = _handle;
val __handle_ws = _handle_ws;
val __handle_monitor = _handle_monitor;
val ws_send = _ws_send;
val ws_recv = _ws_recv;
# used as a websocket client
val new_ws = _new_ws;
val ws_client_send = _ws_client_send;
val ws_client_recv = _ws_client_recv;
# _server is an id that corresponds to the gofiber server object
var __server = null;
val __new_server = _new_server;
val _server = fun() {
    if (__server == null) {
        __server = __new_server();
    }
    return __server;
}();
val shutdown_server = _shutdown_server;
val md_to_html = _md_to_html;
val __sanitize_and_minify = _sanitize_and_minify;

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
    }
    __fetch(resource, options.method, options.headers, options.body)
}

fun get(url) {
    fetch(url)
}

fun download(url, filename="") {
    __download(url, filename)
}

fun post(url, post_body, mime_type="application/json") {
    if (mime_type.to_lower().startswith('content-type')) {
        return error("`post` error: mime_type should not start with 'content-type'");
    }
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

fun handle_monitor(path, should_show=true) {
    __handle_monitor(_server, path, should_show)
}

fun sanitize_and_minify(content, should_sanitize=true, should_minify=true) {
    __sanitize_and_minify(content, should_sanitize, should_minify)
}