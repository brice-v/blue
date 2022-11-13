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

# form_values shouldn't be handled in a get request
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


# THIS IS WORKING! WOO
fun ws_handler(ws) {
    for (true) {
        val x = ws.recv();
        #println("x = #{x}");
        ws.send(x);
    }
}


http.handle("/", get_handler, method="GET");
http.handle_ws("/ws", ws_handler);
http.handle("/abc/:a/:b", get_handler_for_multiple_things, method="GET");
http.handle("/hello/:name", hello_handler, method="GET");
http.handle("/json", json_handler, method="GET");
http.handle("/post/:a/:b", post_handler, method="POST");
http.handle_monitor("/monitor");


fun run() {
    http.serve();
}
true;