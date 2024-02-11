## `http` is the module that deals with most http related
## functions. When imported, it stores references to the
## http server when it is created

val __download = _download;
val __serve = _serve;
val __static = _static;
val __handle = _handle;
val __handle_use = _handle_use;
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
        __server = __new_server('tcp4');
    }
    return __server;
}();
val __shutdown_server = _shutdown_server;
val __md_to_html = _md_to_html;
val __sanitize_and_minify = _sanitize_and_minify;

val __inspect = _inspect;

val url_encode = _url_encode;
val url_escape = _url_escape;
val url_unescape = _url_unescape;

# This doesn't really belong here but is most likely useful if creating servers
val open_browser = _open_browser;

fun get(url, full_resp=false) {
    ## `get` is just a call to `fetch` with the given url
    ## it will return the body returned as a string, or null
    ##
    ## get(url: str, full_resp: bool=false) -> str|null
    fetch(url, full_resp=full_resp)
}

fun download(url, filename="") {
    ##std:this,__download
    ## `download` will download the given url to a file passed in
    ##
    ## This function will work best with actual files hosted on the
    ## internet
    ##
    ## download(url: str, filename: str="") -> null
    __download(url, filename)
}

fun post(url, post_body, mime_type="application/json", full_resp=false) {
    ## `post` will send a POST request to the given url with the body as a string
    ## post_body should be a string in the format of the mime_type.
    ##
    ## post(url: str, post_body: str, mime_type: str="application/json", full_resp: bool=false)
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
    fetch(url, options, full_resp=full_resp)
}

fun put(url, put_body, mime_type="application/json", full_resp=false)  {
    ## `put` will send a PUT request to the given url with the body as a string
    ## put_body should be a string in the format of the mime_type.
    ##
    ## put(url: str, put_body: str, mime_type: str="application/json", full_resp: bool=false)
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
    fetch(url, options, full_resp=full_resp)
}

fun patch(url, patch_body, mime_type="application/json", full_resp=false)  {
    ## `patch` will send a PATCH request to the given url with the body as a string
    ## patch_body should be a string in the format of the mime_type.
    ##
    ## patch(url: str, patch_body: str, mime_type: str="application/json", full_resp: bool=false)
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
    fetch(url, options, full_resp=full_resp)
}

fun delete(url, full_resp=false) {
    ## `delete` will send a DELETE request to the given url
    ##
    ## delete(url: str, full_resp: bool=false)
    val options = {
        method: 'DELETE',
        headers: {},
        body: null,
    };
    fetch(url, options, full_resp=full_resp)
}

fun serve(addr_port="localhost:3001", use_embedded_lib_web=true) {
    ##std:this,__serve
    ## `serve` will start up the _server in http on the given
    ## address and port.
    ##
    ## address and port should be a single string with a colon
    ## delimeter between the too. see below signature.
    ##
    ## other http module functions will operate on this
    ## open server object.
    ##
    ## use_embedded_lib_web is set to true by default and allows the use
    ## of an embedded copy of twind, preact, and water(-dark/light).css
    ##    Example for twind: <script src="twind.js"></script> in <head> tag
    ##    Example for preact: import { html, render } from './preact.js' (in mjs files/modules)
    ##    Example for water: <link rel="stylesheet" type="text/css" href="/water-dark.css">
    ##                       or water-light.css or water.css which uses system theme
    ##
    ## serve(addr_port: str='localhost:3001', use_embedded_lib_web: bool=true) -> null
    __serve(_server, addr_port, use_embedded_lib_web)
}

fun handle(pattern, fn, method="GET") {
    ##std:this,__handle
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

fun handle_use(pattern="", fn) {
    ##std:this,__handle_use
    ## `handle_use` takes an optional pattern, and a function, and method
    ## and attaches itself to the _server http object
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
    __handle_use(_server, pattern, fn, "")
}

fun handle_ws(pattern, fn) {
    ##std:this,__handle_ws
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
    ##std:this,__static
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
    ##std:this,__handle_monitor
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
    ##std:this,__sanitize_and_minify
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
    ##std:this,__ws_send
    ## `ws_send` will send on the websocket connection within an http server websocket handler
    ## note: this function should mostly be called from the core 'send' function
    ##
    ## ws_send(ws_id: uint, value: any) -> null
    __ws_send(ws_id, value)
}

fun ws_recv(ws_id) {
    ##std:this,__ws_recv
    ## `ws_recv` will receive on the websocket connection within an http server websocket handler
    ## note: this function should mostly be called from the core 'recv' function
    ##
    ## ws_recv(ws_id: uint) -> any
    __ws_recv(ws_id)
}

fun ws_client_send(ws_id, value) {
    ##std:this,__ws_client_send
    ## `ws_client_send` will send on a websocket connection initalized via 'new_ws'
    ## note: this function should mostly be called from the core 'send' function
    ##
    ## ws_client_send(ws_id: uint, value: any) -> null
    __ws_client_send(ws_id, value)
}

fun ws_client_recv(ws_id) {
    ##std:this,__ws_client_recv
    ## `ws_client_recv` will receive on a websocket connection initalized via 'new_ws'
    ## note: this function should mostly be called from the core 'recv' function
    ##
    ## ws_client_recv(ws_id: uint) -> null
    __ws_client_recv(ws_id)
}

fun new_ws(path) {
    ##std:this,__new_ws
    ## `new_ws` will initalize a websocket client to be used to 'send' and 'recv' on
    ##
    ## path should be in the normal websocket format
    ## ex: "ws://localhost:3001/ws"
    ##
    ## new_ws(path: str) -> {t: 'ws/client', v: uint}
    __new_ws(path)
}

fun shutdown_server() {
    ##std:this,__shutdown_server
    ## `shutdown_server` will catch interupts to shutdown the http server
    ##
    ## shutdown_server() -> null
    __shutdown_server()
}

fun md_to_html(content) {
    ##std:this,__md_to_html
    ## `md_to_html` will transform the markdown content passed in to valid html
    ##
    ## md_to_html(content: str) -> str
    __md_to_html(content)
}

fun inspect(obj) {
    ##std:this,__inspect
    ## `inspect` prints out the details of the http object
    ##
    ## inspect(obj: {t: 'ws': v: uint}) -> str
    return match obj {
        {t: _, v: _} => {
            __inspect(obj.v, obj.t)
        },
        _ => {
            error("`inspect` expects object")
        },
    };
}

fun status(code) {
    ## `status` will return a status code for any http request if returned in an http handler
    ##
    ## status(code: int) -> {'t':'http/status','code':int}
    if (type(code) != Type.INT) {
        return error("http status code must be #{Type.INT}, got=#{type(code)}");
    }
    {'t':'http/status', 'code':code}
}

fun redirect(location, code=302) {
    ## `redirect` will redirect the http handler to a new location the code defaults to 302
    ##
    ## redirect(location: str, code: int) -> {'t':'http/redirect','location':str,'code':int}
    if (type(code) != Type.INT) {
        return error("http status code must be #{Type.INT}, got=#{type(code)}");
    }
    if (type(location) != Type.STRING) {
        return error("http redirect location must be #{Type.STRING}, got=#{type(location)}");
    }
    {'t':'http/redirect', 'location':location, 'code':code}
}

fun next() {
    ## `next` will send the http handler to the next handler available
    ##
    ## next() -> {'t':'http/next'}
    {'t': 'http/next'}
}

fun send_file(path) {
    ## `send_file` will send the file on the http handler with the proper content-type
    ## set based on file extension
    ##
    ## send_file(path: str) -> {'t':'http/send_file','path': str}
    assert(type(path) == Type.STRING);
    {'t': 'http/send_file','path':path}
}


fun new_server(network="tcp4") {
    ##std:this,__new_server
    ## `new_server` will return a core http server object that can be used
    ## to call http functions against
    ##
    ## network should be 'tcp4' or 'tcp6'
    ##
    ## new_server(network: str='tcp4') -> HTTP_SERVER_OBJ
    var this = {};
    this._s = __new_server(network);
    this.serve = fun(addr_port="localhost:3001", use_embedded_lib_web=true) {
        __serve(this._s, addr_port, use_embedded_lib_web)
    };
    this.handle = fun(pattern, fn, method="GET") {
        __handle(this._s, pattern, fn, method)
    };
    this.handle_use = fun(pattern="", fn) {
        __handle_use(this._s, pattern, fn, "")
    };
    this.handle_ws = fun(pattern, fn) {
        __handle_ws(this._s, pattern, fn)
    };
    this.static = fun(prefix="/", dir_path=".", browse=false) {
        __static(this._s, prefix, dir_path, browse)
    };
    this.handle_monitor = fun(path, should_show=false) {
        __handle_monitor(this._s, path, should_show)
    };
    return this;
}