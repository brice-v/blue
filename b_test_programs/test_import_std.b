import http

val url = "https://catfact.ninja/fact";
val res = http.get(url);

println("url='#{url}', res=`#{res}`");

if (res == "") {
    return false;
}

try {
    http.get(1);
    return false;
} catch (e) {
    println("#{e}");
    if (e != "EvaluatorError: argument to `get` must be string. got INTEGER") {
        println("#{e}");
    }
    true;
}

try {
    println(http._get(url));
    println("This should be unreachable");
    return false;
} catch (e) {
    println("Hit exception as expected");
    true;
}

return true;