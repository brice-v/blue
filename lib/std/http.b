## `http` is the module that deals with most http related
## functions. When imported, it stores references to the
## http server when it is created


val __fetch = _fetch;
val __download = _download;
val __serve = _serve;
val __static = _static;
val __handle = _handle;
val __handle_ws = _handle_ws;
val __handle_monitor = _handle_monitor;
val __ws_send = _ws_send;
val __ws_recv = _ws_recv;
# used as a websocket client
val __new_ws = _new_ws;
val __ws_client_send = _ws_client_send;
val __ws_client_recv = _ws_client_recv;
# _server is an id that corresponds to the gofiber server object
var __server = null;
val __new_server = _new_server;
val _server = fun() {
    if (__server == null) {
        __server = __new_server();
    }
    return __server;
}();
val __shutdown_server = _shutdown_server;
val __md_to_html = _md_to_html;
val __sanitize_and_minify = _sanitize_and_minify;

# Default headers seem to be host, user-agent, accept-encoding (not case sensitive for these check pictures)
# deno also used accept: */* (not sure what that is)
fun fetch(resource, options=null) {
    ## `fetch` allows the user to send GET, POST, PUT, DELETE
    ## http methods to a various resource
    ##
    ## there are other specific methods that populate these
    ## options appropriately. user-agent in header is always
    ## set to one specific to blue.
    ##
    ## example option to send get request:
    ##                 {method: 'GET', headers: {}, body: null}
    ##
    ## example option to send post request:
    ## {method: 'POST', body: str, headers: {'content-type': mime_type}}
    ##
    ## fetch(resource: str, options: map[any:str]=null) -> any
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
            val ht = type(options.headers);
            if (ht != 'MAP') {
                return error("`fetch` error:  options.headers must be MAP. got=#{ht}");
            }
        }
    }
    __fetch(resource, options.method, options.headers, options.body)
}

fun get(url) {
    ## `get` is just a call to `fetch` with the given url
    ## it will return the body returned as a string, or null
    ##
    ## get(url: str) -> string|null
    fetch(url)
}

fun download(url, filename="") {
    ## `download` will download the given url to a file passed in
    ##
    ## This function will work best with actual files hosted on the
    ## internet
    ##
    ## download(url: str, filename: str="") -> null
    __download(url, filename)
}

fun post(url, post_body, mime_type="application/json") {
    ## `post` will send a POST request to the given url with the body as a string
    ## post_body should be a string in the format of the mime_type.
    ##
    ## post(url: str, post_body: str, mime_type: str="application/json")
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

fun put(url, put_body, mime_type="application/json")  {
    ## `put` will send a PUT request to the given url with the body as a string
    ## put_body should be a string in the format of the mime_type.
    ##
    ## put(url: str, put_body: str, mime_type: str="application/json")
    if (mime_type.to_lower().startswith('content-type')) {
        return error("`put` error: mime_type should not start with 'content-type'");
    }
    val options = {
        method: 'PUT',
        body: put_body,
        headers: {
            'content-type': mime_type
        }
    };
    fetch(url, options)
}

fun patch(url, patch_body, mime_type="application/json")  {
    ## `patch` will send a PATCH request to the given url with the body as a string
    ## patch_body should be a string in the format of the mime_type.
    ##
    ## patch(url: str, patch_body: str, mime_type: str="application/json")
    if (mime_type.to_lower().startswith('content-type')) {
        return error("`patch` error: mime_type should not start with 'content-type'");
    }
    val options = {
        method: 'PATCH',
        body: patch_body,
        headers: {
            'content-type': mime_type
        }
    };
    fetch(url, options)
}

fun delete(url) {
    ## `delete` will send a DELETE request to the given url
    ##
    ## delete(url: str)
    val options = {
        method: 'DELETE',
        headers: {},
    };
    fetch(url, options)
}

fun serve(addr_port="localhost:3001") {
    ## `serve` will start up the _server in http on the given
    ## address and port.
    ##
    ## address and port should be a single string with a colon
    ## delimeter between the too. see below signature.
    ##
    ## other http module functions will operate on this
    ## open server object.
    ##
    ## serve(addr_port: str='localhost:3001') -> null
    __serve(_server, addr_port)
}

fun handle(pattern, fn, method="GET") {
    ## `handle` takes a pattern, function, and method
    ## and attaches itself to the _server http object
    ## 
    ## handler functions should return a string or bytes or null
    ## they can take parameters which correspond to the pattern
    ## string.
    ##
    ## example: with a pattern string of '/hello/:a/:b'
    ## the handler function should have a signature such as
    ## fun(a, b) {} which allows the params to be used
    ##
    ## example: with a 'POST' method the handler function
    ## should have a signature such as fun(post_values=['a', 'b']) {}
    ## where 'a' and 'b' are values received in the POST request
    ##
    ## query_params operates in a similar fashion to the post_values
    ## accepting a list of strings for the query_params passed to the
    ## request
    ##
    ## headers is also reserved in the function signature
    ## to allow the user to retrieve the headers of the request
    ## passed in
    ##
    ## handle(pattern: str, fn: fun, method: str='GET') -> null
    __handle(_server, pattern, fn, method)
}

fun handle_ws(pattern, fn) {
    ## `handle_ws` allows the user to receive and send on a
    ## websocket connection. It attaches itself to the _server
    ## http object (so serve must be called first)
    ##
    ## ws is reserved in the handler function signature
    ## to recv() and send() on the websocket connection
    ##
    ## messages will almost all be of type 'TextMessage' which
    ## is just strings
    ##
    ## example: a simple echo websocket server looks like so 
    ## handler_ws("/ws", fun(ws) { for (true) { val x = ws.recv(); ws.send(x); } })
    ##
    ## handle_ws(pattern: str, fn: fun) -> null
    __handle_ws(_server, pattern, fn)
}

fun static(prefix="/", dir_path=".", browse=false) {
    ## `static` declares a static directory to be used by the http
    ## server that is attached to the _server http object
    ##
    ## prefix is how the files can be retrieved once they are
    ## being served.
    ##
    ## dir_path is the directory of files to be served statically
    ##
    ## browse is a boolean to let the user decide if this directory
    ## should be browsable by the end user, usually for ftp style sites
    ##
    ## static(prefix: str='/', dir_path: str='.', browse: bool=false) -> null
    __static(_server, prefix, dir_path, browse)
}

fun handle_monitor(path, should_show=true) {
    ## `handle_monitor` serves the fiber monitor to the user
    ## at the specified path on the _http server object
    ##
    ## path is the path the user will request to see the monitor
    ##
    ## should_show is a boolean to determine whether the monitor should
    ## display, mostly used if it should be hidden from end users
    ##
    ## handle_monitor(path: str, should_show: bool=false) -> null
    __handle_monitor(_server, path, should_show)
}

fun sanitize_and_minify(content, should_sanitize=true, should_minify=true) {
    ## `sanitize_and_minify` is used to minify and/or sanitize the string
    ## content passed in
    ##
    ## this function can be used for any html content and does not need to
    ## be called with any http.serve() function before
    ##
    ## sanitize_and_minify(content: str, should_sanitize: bool=true, should_minify: bool=true) -> str
    __sanitize_and_minify(content, should_sanitize, should_minify)
}

fun ws_send(ws_id, value) {
    ## `ws_send` will send on the websocket connection within an http server websocket handler
    ## note: this function should mostly be called from the core 'send' function
    ##
    ## ws_send(ws_id: uint, value: any) -> null
    __ws_send(ws_id, value)
}

fun ws_recv(ws_id) {
    ## `ws_recv` will receive on the websocket connection within an http server websocket handler
    ## note: this function should mostly be called from the core 'recv' function
    ##
    ## ws_recv(ws_id: uint) -> any
    __ws_recv(ws_id)
}

fun ws_client_send(ws_id, value) {
    ## `ws_client_send` will send on a websocket connection initalized via 'new_ws'
    ## note: this function should mostly be called from the core 'send' function
    ##
    ## ws_client_send(ws_id: uint, value: any) -> null
    __ws_client_send(ws_id, value)
}

fun ws_client_recv(ws_id) {
    ## `ws_client_recv` will receive on a websocket connection initalized via 'new_ws'
    ## note: this function should mostly be called from the core 'recv' function
    ##
    ## ws_client_recv(ws_id: uint) -> null
    __ws_client_recv(ws_id)
}

fun new_ws(path) {
    ## `new_ws` will initalize a websocket client to be used to 'send' and 'recv' on
    ##
    ## path should be in the normal websocket format
    ## ex: "ws://localhost:3001/ws"
    ##
    ## new_ws(path: str) -> {t: 'ws/client', v: uint}
    __new_ws(path)
}

fun shutdown_server() {
    ## `shutdown_server` will catch interupts to shutdown the http server
    ##
    ## shutdown_server() -> null
    __shutdown_server()
}

fun md_to_html(content) {
    ## `md_to_html` will transform the markdown content passed in to valid html
    ##
    ## md_to_html(content: str) -> str
    __md_to_html(content)
}