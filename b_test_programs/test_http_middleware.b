import http

http.helmet({'x_frame_options': 'DENY', 'hsts_max_age': 60});
http.compress(1);
http.limiter(10, 30);
val key = http.generate_cookie_key(16);
http.encrypt_cookie(key, ['session']);

val s = http.new_server('tcp4');
s.helmet();
s.compress();
s.limiter();
s.encrypt_cookie(http.generate_cookie_key());

assert(true);
