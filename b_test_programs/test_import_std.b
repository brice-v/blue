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
        return false;
    }
    return true;
}

return true;