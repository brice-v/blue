import http

val url = "https://catfact.ninja/fact";
val res = http.get(url);

println("url='#{url}', res=`#{res}`");

assert(res != "");

try {
    http.get(1);
    assert(false);
} catch (e) {
    println("#{e}");
    if (e != "argument to `get` must be string. got INTEGER") {
        println("#{e}");
    }
    assert(true);
}

try {
    println(http._get(url));
    println("This should be unreachable");
    assert(false);
} catch (e) {
    println("Hit exception as expected");
    assert(true);
}

assert(true);