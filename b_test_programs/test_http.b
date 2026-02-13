#VM IGNORE
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

fun() {
    var { status, body } = http.post(URL, expected_post_data, full_resp=true);
    assert(status == 200);
    assert(body == '{}');
}()

val get_resp = http.get(URL);
println("get_resp = #{get_resp}");
if (get_resp != expected_post_data) {
    return false;
}

fun() {
    var { status, body } = http.get(URL, full_resp=true);
    assert(status == 200);
    assert(body == expected_post_data);
}()

### Cant test with this URL (patch not supported)
val expected_patch_data = '{"Hello":5555}';
val patch_resp = http.patch(URL, expected_patch_data);
println("patch_resp = #{patch_resp}");
if (patch_resp != '{}') {
    return false;
}

val get_resp1 = http.get(URL);
println("get_resp1 = #{get_resp1}");
if (get_resp1 != expected_patch_data) {
    return false;
}
###

val expected_put_data = '{"abc":123456}';
val put_resp = http.put(URL, expected_put_data);
println("put_resp = #{put_resp}");
if (put_resp != '{}') {
    return false;
}

fun() {
    var { status, body } = http.put(URL, expected_put_data, full_resp=true);
    assert(status == 200);
    assert(body == '{}');
}()

val get_resp1 = http.get(URL);
println("get_resp1 = #{get_resp1}");
if (get_resp1 != expected_put_data) {
    return false;
}

fun() {
    var { status, body } = http.get(URL, full_resp=true);
    assert(status == 200);
    assert(body == expected_put_data);
}()

val delete_resp = http.delete(URL, full_resp=true);
println("delete_resp = #{delete_resp.body}");
assert(delete_resp.status == 200);
if (delete_resp.body != '{}') {
    return false
}

val get_resp2 = http.get(URL);
println("get_resp2 = #{get_resp2}");
if (get_resp2 != '{"message":"Not Found"}') {
    return false;
}

true;