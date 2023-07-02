import http
import time

import test_http_server

fun main() {
    val server_pid = spawn(test_http_server.run, []);
    println("server_pid = #{server_pid}");
    # This makes sure the server is up
    for (true) {
        time.sleep(1000);
        try {
            http.get("http://localhost:3001/");
            break;
        } catch (ignored) {
            continue;
        }
    }
    val resp = http.get("http://localhost:3001/abc/AAAA/BBBB?name=Brice&email=brice%40example.com")
    println("resp = `#{resp}`");
    val expected_resp = "<p>url params: a=AAAA, b=BBBB | query_params: name=Brice, email=brice@example.com, headers="
    assert(resp.startswith(expected_resp), "Response for get_handler_for_multiple_things did not return expected result");
    assert("Host: localhost:3001" in resp, "Response for get_handler_for_multiple_things did not return expected result");
    assert("User-Agent: blue/v" in resp, "Response for get_handler_for_multiple_things did not return expected result");
    assert("Accept-Encoding: gzip" in resp, "Response for get_handler_for_multiple_things did not return expected result");

    val resp1 = http.post("http://localhost:3001/post/abc/213/?name=john&pass=doe", null, mime_type="application/x-www-form-urlencoded");
    println("resp1 = `#{resp1}`");
    val expected_resp1 = '{"name":"john","pass":"doe","a":"abc","b":"213"}';
    assert(resp1 == expected_resp1, "Response for post_handler with url form data did not return expected");

    val json_data1 = '{"name":"john","pass":"doe"}';
    val resp2 = http.post("http://localhost:3001/post/abc/213", json_data1);
    println("resp2 = `#{resp2}`");
    val expected_resp2 = '{"name":"john","pass":"doe","a":"abc","b":"213"}';
    assert(resp2 == expected_resp2, "Response for post_handler with json data did not return expected");

    val json_data2 = '{"name":"john","pass":["doe",2]}';
    val resp3 = http.post("http://localhost:3001/post/json/213", json_data2);
    println("resp3 = `#{resp3}`");
    val expected_resp3 = '{"name":"john","pass":["doe",2.000000],"a":"json","b":"213"}';
    assert(resp3 == expected_resp3, "Response for post_handler with json data (with list) did not return expected");

    val xml_data1 = "<login><name>john</name><pass>doe</pass></login>";
    val resp4 = http.post("http://localhost:3001/post/xml/data1", xml_data1, mime_type="application/xml");
    println("resp4 = `#{resp4}`");
    val expected_resp4 = '{"name":"john","pass":"doe","a":"xml","b":"data1"}';
    assert(resp4 == expected_resp4, "Response for post_handler, with xml data did not return expected");

    val form_data1 = "name=john&pass=doe";
    val resp5 = http.post("http://localhost:3001/post/FORM/data", form_data1, mime_type="application/x-www-form-urlencoded");
    println("resp5 = `#{resp5}`");
    val expected_resp5 = '{"name":"john","pass":"doe","a":"FORM","b":"data"}';
    assert(resp5 == expected_resp5, "Response for post_handler, with form encoded as body url encoded, did not return expected");

    val post_put_patch_delete_url = "http://localhost:3001/all/aaa/bbb";
    val _post_data = '{"name":"HELLO","pass":"WORLD"}';
    val _put_data = '{"put new":"HELLO","keys":"WORLD"}';
    val _patch_data = '{"new":"HELLO PATCH","keys":"WORLD PATCH"}';
    val post_put_patch_delete_expected_resp = '{"name":"john","pass":"doe","a":"abc","b":"213"}';
    val __resp = http.post(post_put_patch_delete_url, _post_data).from_json();
    println("POST __resp = #{__resp}");
    val expected_resp__0 = {"name":"HELLO","pass":"WORLD","a":"aaa","b":"bbb","request":{"method":"POST","proto":"http","uri":"http://localhost:3001/all/aaa/bbb","scheme":"http","host":"localhost:3001","request_uri":"/all/aaa/bbb","hash":"","headers":{"Accept-Encoding":"gzip","Content-Length":"31","Content-Type":"application/json","Host":"localhost:3001","User-Agent":"blue/v#{VERSION}"},"ip":"127.0.0.1","is_from_local":true,"is_secure":false},"cookies":""};
    println("POST expected_resp__0 = #{expected_resp__0}");
    assert(__resp == expected_resp__0);
    val __resp1 = http.put(post_put_patch_delete_url, _put_data).from_json();
    println("PUT __resp1 = #{__resp1}");
    val expected_resp__1 = {"name":"","pass":"","a":"aaa","b":"bbb","request":{"method":"PUT","proto":"http","uri":"http://localhost:3001/all/aaa/bbb","scheme":"http","host":"localhost:3001","request_uri":"/all/aaa/bbb","hash":"","headers":{"Accept-Encoding":"gzip","Content-Length":"34","Content-Type":"application/json","Host":"localhost:3001","User-Agent":"blue/v#{VERSION}"},"ip":"127.0.0.1","is_from_local":true,"is_secure":false},"cookies":""};
    println("PUT expected_resp__1 = #{expected_resp__1}");
    assert(__resp1 == expected_resp__1);
    val __resp2 = http.patch(post_put_patch_delete_url, _patch_data).from_json();
    println("PATCH __resp2 = #{__resp2}");
    val expected_resp__2 = {"name":"","pass":"","a":"aaa","b":"bbb","request":{"method":"PATCH","proto":"http","uri":"http://localhost:3001/all/aaa/bbb","scheme":"http","host":"localhost:3001","request_uri":"/all/aaa/bbb","hash":"","headers":{"Accept-Encoding":"gzip","Content-Length":"42","Content-Type":"application/json","Host":"localhost:3001","User-Agent":"blue/v#{VERSION}"},"ip":"127.0.0.1","is_from_local":true,"is_secure":false},"cookies":""};
    println("PATCH expected_resp__2 = #{expected_resp__2}");
    assert(__resp2 == expected_resp__2);
    val __resp3 = http.delete(post_put_patch_delete_url).from_json();
    println("DELETE __resp3 = #{__resp3}");
    val expected_resp__3 = {"name":"","pass":"","a":"aaa","b":"bbb","request":{"method":"DELETE","proto":"http","uri":"http://localhost:3001/all/aaa/bbb","scheme":"http","host":"localhost:3001","request_uri":"/all/aaa/bbb","hash":"","headers":{"Accept-Encoding":"gzip","Content-Length":"0","Host":"localhost:3001","User-Agent":"blue/v#{VERSION}"},"ip":"127.0.0.1","is_from_local":true,"is_secure":false},"cookies":""};
    println("DELETE expected_resp__3 = #{expected_resp__3}");
    assert(__resp3 == expected_resp__3);
    val __resp4 = http.get(post_put_patch_delete_url).from_json();
    println("GET __resp4 = #{__resp4}");
    val expected_resp__4 = {"name":"","pass":"","a":"aaa","b":"bbb","request":{"method":"GET","proto":"http","uri":"http://localhost:3001/all/aaa/bbb","scheme":"http","host":"localhost:3001","request_uri":"/all/aaa/bbb","hash":"","headers":{"Accept-Encoding":"gzip","Host":"localhost:3001","User-Agent":"blue/v#{VERSION}"},"ip":"127.0.0.1","is_from_local":true,"is_secure":false},"cookies":""};
    println("GET expected_resp__4 = #{expected_resp__4}");
    assert(__resp4 == expected_resp__4);

    val redirect_handler_resp = http.get("http://localhost:3001/redirect");
    println("redirect_handler_resp (#{type(redirect_handler_resp)}) = #{redirect_handler_resp}");
    assert(redirect_handler_resp == "I'm a teapot");
    val status_handler_resp = http.get("http://localhost:3001/healthcheck");
    println("status_handler_resp (#{type(status_handler_resp)}) = #{status_handler_resp}");
    assert(status_handler_resp == "I'm a teapot");
    val status_handler2_resp = fetch("http://localhost:3001/status2");
    println("status_handler2_resp (#{type(status_handler2_resp)}) = #{status_handler2_resp}");
    assert(status_handler2_resp.status == 999);

    val resp6 = http.get("http://localhost:3001/json");
    println("resp6 = `#{resp6}`");
    val expected_resp6 = '{"a":1,"b":2,"c":["d","e","f"]}';
    assert(resp6 == expected_resp6, "Response for get handler with json resp did not return expected");

    val resp7 = http.get("http://localhost:3001/hello/Someone");
    println("resp7 = `#{resp7}`");
    val expected_resp7 = '<p><b>Hello Someone!</b></p>';
    assert(resp7 == expected_resp7, "Response for get handler with param did not return expected");



    val get_test_special_json_case_int = from_json(http.get("http://localhost:3001/return-json/int"));
    println("get_test_special_json_case_int = #{get_test_special_json_case_int}, type(get_test_special_json_case_int) = #{type(get_test_special_json_case_int)}")
    assert(get_test_special_json_case_int == 123)
    val get_test_special_json_case_float = from_json(http.get("http://localhost:3001/return-json/float"));
    println("get_test_special_json_case_float = #{get_test_special_json_case_float}, type(get_test_special_json_case_float) = #{type(get_test_special_json_case_float)}")
    assert(get_test_special_json_case_float == 1.234)
    val get_test_special_json_case_list = from_json(http.get("http://localhost:3001/return-json/list"));
    println("get_test_special_json_case_list = #{get_test_special_json_case_list}, type(get_test_special_json_case_list) = #{type(get_test_special_json_case_list)}")
    assert(get_test_special_json_case_list == [1,2,3])
    val get_test_special_json_case_map = from_json(http.get("http://localhost:3001/return-json/map"));
    println("get_test_special_json_case_map = #{get_test_special_json_case_map}, type(get_test_special_json_case_map) = #{type(get_test_special_json_case_map)}")
    assert(get_test_special_json_case_map == {'hello':123})
    val get_test_special_json_case_null = from_json(http.get("http://localhost:3001/return-json/null"));
    println("get_test_special_json_case_null = #{get_test_special_json_case_null}, type(get_test_special_json_case_null) = #{type(get_test_special_json_case_null)}")
    assert(get_test_special_json_case_null == null)
    val get_test_special_json_case_bool = from_json(http.get("http://localhost:3001/return-json/bool"));
    println("get_test_special_json_case_bool = #{get_test_special_json_case_bool}, type(get_test_special_json_case_bool) = #{type(get_test_special_json_case_bool)}")
    assert(get_test_special_json_case_bool == true)

    val post_test_special_json_case_int = from_json(http.post("http://localhost:3001/return-json/int", ""));
    println("post_test_special_json_case_int = #{post_test_special_json_case_int}, type(post_test_special_json_case_int) = #{type(post_test_special_json_case_int)}")
    assert(post_test_special_json_case_int == 123)
    val post_test_special_json_case_float = from_json(http.post("http://localhost:3001/return-json/float", ""));
    println("post_test_special_json_case_float = #{post_test_special_json_case_float}, type(post_test_special_json_case_float) = #{type(post_test_special_json_case_float)}")
    assert(post_test_special_json_case_float == 1.234)
    val post_test_special_json_case_list = from_json(http.post("http://localhost:3001/return-json/list", ""));
    println("post_test_special_json_case_list = #{post_test_special_json_case_list}, type(post_test_special_json_case_list) = #{type(post_test_special_json_case_list)}")
    assert(post_test_special_json_case_list == [1,2,3])
    val post_test_special_json_case_map = from_json(http.post("http://localhost:3001/return-json/map", ""));
    println("post_test_special_json_case_map = #{post_test_special_json_case_map}, type(post_test_special_json_case_map) = #{type(post_test_special_json_case_map)}")
    assert(post_test_special_json_case_map == {'hello':123})
    val post_test_special_json_case_null = from_json(http.post("http://localhost:3001/return-json/null", ""));
    println("post_test_special_json_case_null = #{post_test_special_json_case_null}, type(post_test_special_json_case_null) = #{type(post_test_special_json_case_null)}")
    assert(post_test_special_json_case_null == null)
    val post_test_special_json_case_bool = from_json(http.post("http://localhost:3001/return-json/bool", ""));
    println("post_test_special_json_case_bool = #{post_test_special_json_case_bool}, type(post_test_special_json_case_bool) = #{type(post_test_special_json_case_bool)}")
    assert(post_test_special_json_case_bool == true)

    val ws = http.new_ws("ws://localhost:3001/ws");
    #for (true) {
        var x = "Sending from Client!";
        ws.send(x);
        val y = ws.recv();
        println("Received on Client, y = #{y}");
        assert(x == y, "Websocket Handler did not echo the sent value from the client");
    #}
    http.shutdown_server();
}

main();
true;