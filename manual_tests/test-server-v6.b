import http


fun main() {
    val s = http.new_server(network="tcp6")

    s.handle("/", fun() {
        return "<h1> Hello World</h1>"
    })
    s.serve("[::1]:3001");
}

main();