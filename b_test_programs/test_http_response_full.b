import http

val URL = "https://www.google.com/";
val resp = http.fetch(URL, full_resp=true)
for ([k, v] in resp) {
    if (k == 'body') {
        continue;
    }
    println("#{k}=#{v}");
}
assert(resp.body != null);
assert(len(resp.body) > 10);
assert(resp.status == 200);
assert(len(resp.headers) != 0);
assert(len(resp.cookies) != 0);
assert(resp.proto == 'HTTP/2.0');
assert(resp.uncompressed);
assert(resp.request == {method: 'GET', proto: 'HTTP/1.1', url: URL});