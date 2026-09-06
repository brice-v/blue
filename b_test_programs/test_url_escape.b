import http

val url_path = "file:///C:/test/test/test/Something here to encode?with=var"
val expected_path = "file:///C:/test/test/test/Something%20here%20to%20encode?with=var";

assert(http.url_encode(url_path) == expected_path);

# Test escape and unescape

val escaped = http.url_escape(url_path);
assert(escaped == "file%3A%2F%2F%2FC%3A%2Ftest%2Ftest%2Ftest%2FSomething+here+to+encode%3Fwith%3Dvar");
assert(http.url_unescape(escaped) == url_path);