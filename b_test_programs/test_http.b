import http

val URL = "https://jsonendpoint.com/blue/blue-http-test";

val data = "Hello";

val test_data = {data: 1234};
val test_data_as_json = test_data.to_json();
println("test_data_as_json = #{test_data_as_json}");
val expected_post_data = '{"Hello":1234}';
println("expected_post_data = #{expected_post_data}");

if (test_data_as_json != expected_post_data) {
    return false;
}

val post_resp = http.post(URL, expected_post_data);
println("post_resp = #{post_resp}");
if (post_resp != '{}') {
    return false;
}

val get_resp = http.get(URL);
println("get_resp = #{get_resp}");
if (get_resp != expected_post_data) {
    return false;
}

true;