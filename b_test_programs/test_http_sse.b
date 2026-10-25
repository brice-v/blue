import http

fun sse_handler(sse) {
    http.sse_send(sse, "hello");
}

http.handle_sse("/events", sse_handler);

val server = http.new_server();
server.handle_sse("/events", sse_handler);

assert(true);
