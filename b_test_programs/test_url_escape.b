import http

val path = "file:///C:/test/test/test/Something here to encode?with=var"
val expected_path = "file:///C:/test/test/test/Something%20here%20to%20encode?with=var";

assert(http.url_encode(path) == expected_path);