import http

val some_map = {
    a: 1,
    b: 2,
    c: ['d', 'e', 'f'],
};

fun get_handler() {
    "<h1>Hello World!</h1>"
}

fun hello_handler(name) {
    "<b>Hello #{name}!</b>"
}

# form_values shouldnt be handled in a get request
# Should we use default params to figure out what we need to pass in though? - we can set them to nil
# This would be specialty logic that ignores default params for function handlers
# Example: http://localhost:3001/abc/AAAA/BBBB?name=Brice&email=brice%40example.com
fun get_handler_for_multiple_things(a, b, query_params=["name", "email"], headers) {
    "url params: a=#{a}, b=#{b} | query_params: name=#{name}, email=#{email}, headers=#{headers}"
}

fun json_handler(name) {
    some_map.to_json()
}

# to test below use the following
# curl -X POST -H "Content-Type: application/json" --data "{\"name\":\"john\",\"pass\":\"doe\"}" localhost:3001/post/abc/213
# curl -X POST -H "Content-Type: application/json" --data "{\"name\":\"john\",\"pass\": [\"doe\", 2]}" localhost:3001/post/abc/213
# curl -X POST -H "Content-Type: application/xml" --data "<login><name>john</name><pass>doe</pass></login>" localhost:3001/post/abc/213
# curl -X POST -H "Content-Type: application/x-www-form-urlencoded" --data "name=john&pass=doe" localhost:3001/post/abc/213
# curl -X POST -F name=john -F pass=doe http://localhost:3001/post/abc/213
# curl -X POST "http://localhost:3001/post/abc/213/?name=john&pass=doe"
fun post_handler(a, b, post_values=["name", "pass"]) {
    val thing = {
        'name': name,
        'pass': pass,
        'a': a,
        'b': b,
    };
    #println("post_handler a = #{a}, b = #{b}");
    #println("post_handler thing = #{thing}");
    thing.to_json()
}


fun all_handler(a, b, post_values=['name', 'pass'], put_values=['name', 'pass'], patch_values=['name', 'pass'], request, cookies) {
    #println("all_handler request = #{request}");
    val thing = {
        'name': name,
        'pass': pass,
        'a': a,
        'b': b,
        'request': request,
        'cookies': cookies,
    };
    #println("all_handler a = #{a}, b = #{b}");
    #println("all_handler thing = #{thing}");
    thing.to_json()
}

fun ctx_handler(ctx) {
    ctx.set_cookie({'name': 'SOME_NAME', 'value': 'SOME_VALUE'})
    try {
        val localThing = ctx.get_local('my-local');
        println("localThing = #{localThing}");
        return to_json(localThing);
    } catch (e) {
        println(e);
    }
    return null;
}


fun redirect_handler() {
    return http.redirect('/healthcheck')
}

fun status_handler() {
    # teapot code
    return http.status(418);
}

fun status_handler2() {
    return http.status(999);
}


# THIS IS WORKING! WOO
fun ws_handler(ws) {
    for (true) {
        val x = ws.recv();
        #println("x = #{x}");
        ws.send(x);
    }
}

fun return_special_json_handler(t, request) {
    # t is the type of thing we want to test out
    match t {
        "int" => {
            return 123;
        },
        "float" => {
            return 1.234;
        },
        "map" => {
            return {'hello':123};
        },
        "list" => {
            return [1,2,3];
        },
        "null" => {
            if (request.method == 'GET') {
                return null;
            } else {
                # null is a special case for all non-GET requests where it just returns statusOk (200)
                return to_json(null);
            }
        },
        "bool" => {
            return true;
        },
    };
}


http.handle_use(fun(ctx) {
    ctx.set_local("my-local", [1,2,3]);
    return http.next()
})

http.handle("/", get_handler, method="GET");
http.handle_ws("/ws", ws_handler);
http.handle("/abc/:a/:b", get_handler_for_multiple_things, method="GET");
http.handle("/hello/:name", hello_handler, method="GET");
http.handle("/json", json_handler, method="GET");
http.handle("/post/:a/:b", post_handler, method="POST");

http.handle("/all/:a/:b", all_handler);
http.handle("/all/:a/:b", all_handler, method="POST");
http.handle("/all/:a/:b", all_handler, method="PUT");
http.handle("/all/:a/:b", all_handler, method="PATCH");
http.handle("/all/:a/:b", all_handler, method="DELETE");

http.handle("/return-json/:t", return_special_json_handler)
http.handle("/return-json/:t", return_special_json_handler, method="POST")

http.handle('/redirect', redirect_handler);
http.handle('/healthcheck', status_handler);
http.handle('/status2', status_handler2);

http.handle('/ctx-handler', ctx_handler);

http.handle_monitor("/monitor");


fun run() {
    http.serve();
}
true;